package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
)

// ContractParseService 合约解析服务接口
type ContractParseService interface {
	// ParseBlockTransactions 解析指定区块的所有交易合约日志
	// 这是唯一的入口方法，同步执行
	ParseBlockTransactions(ctx context.Context, blockID uint64) error

	// GetByTxHash 根据交易哈希获取解析结果
	GetByTxHash(ctx context.Context, hash string) ([]*models.ContractParseResult, error)
}

type contractParseService struct {
	repo             repository.ContractParseResultRepository
	receiptRepo      repository.TransactionReceiptRepository
	txRepo           repository.TransactionRepository
	blockRepo        repository.BlockRepository
	parserConfigRepo repository.ParserConfigRepository
	coinConfigRepo   repository.CoinConfigRepository
	userAddressRepo  repository.UserAddressRepository
}

func NewContractParseService(
	repo repository.ContractParseResultRepository,
	receiptRepo repository.TransactionReceiptRepository,
	txRepo repository.TransactionRepository,
	blockRepo repository.BlockRepository,
	parserConfigRepo repository.ParserConfigRepository,
	coinConfigRepo repository.CoinConfigRepository,
	userAddressRepo repository.UserAddressRepository,
) ContractParseService {
	return &contractParseService{
		repo:             repo,
		receiptRepo:      receiptRepo,
		txRepo:           txRepo,
		blockRepo:        blockRepo,
		parserConfigRepo: parserConfigRepo,
		coinConfigRepo:   coinConfigRepo,
		userAddressRepo:  userAddressRepo,
	}
}

// ParseBlockTransactions 解析指定区块的所有交易合约日志
func (s *contractParseService) ParseBlockTransactions(ctx context.Context, blockID uint64) error {
	// 1. 获取区块信息
	block, err := s.blockRepo.GetByID(ctx, blockID)
	if err != nil {
		return fmt.Errorf("failed to get block %d: %w", blockID, err)
	}
	if block == nil {
		return fmt.Errorf("block %d not found", blockID)
	}

	// 2. 获取该区块的所有交易
	transactions, err := s.txRepo.GetByBlockID(ctx, blockID)
	if err != nil {
		return fmt.Errorf("failed to get transactions for block %d: %w", blockID, err)
	}

	if len(transactions) == 0 {
		return nil // 没有交易，直接返回
	}

	// 3. 获取交易收据
	receiptMap := make(map[string]*models.TransactionReceipt)

	receipts, err := s.receiptRepo.GetByBlockID(ctx, blockID)
	if err != nil {
		return fmt.Errorf("failed to get transaction receipts by block id: %w", err)
	}

	for _, receipt := range receipts {
		receiptMap[receipt.TxHash] = receipt
	}

	// 4. 获取所有解析器配置
	allParserConfigs, err := s.parserConfigRepo.GetAllParserConfigs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get parser configs: %w", err)
	}
	// 按合约地址分组
	configsByContract := make(map[string][]*models.ParserConfig)
	for _, config := range allParserConfigs {
		if config != nil && config.ContractAddress != "" {
			configsByContract[config.ContractAddress] = append(configsByContract[config.ContractAddress], config)
		}
	}

	// 5. 获取所有币种配置
	coinConfigs, err := s.coinConfigRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get coin configs: %w", err)
	}
	// 6. 按交易索引顺序处理所有交易
	var allResults []*models.ContractParseResult

	for _, tx := range transactions {
		receipt := receiptMap[tx.TxID]
		if receipt == nil || receipt.Status != 1 { // 只处理成功的交易
			continue
		}

		// 获取对应的解析器配置
		configs := configsByContract[tx.AddressTo]
		if len(configs) == 0 {
			// 尝试使用通用配置
			configs = configsByContract["*"]
		}

		if len(configs) == 0 {
			continue // 没有相关的解析器配置
		}
		// 解析交易收据中的日志
		results, err := s.parseTransactionLogs(ctx, tx, receipt, configs, coinConfigs)
		if err != nil {
			// 记录错误但继续处理其他交易
			fmt.Printf("Failed to parse transaction %s: %v\n", tx.TxID, err)
			continue
		}

		allResults = append(allResults, results...)
	}

	// 7. 批量保存解析结果
	if len(allResults) > 0 {
		if err := s.repo.CreateBatch(ctx, allResults); err != nil {
			return fmt.Errorf("failed to save parse results for block %d: %w", blockID, err)
		}
	}

	return nil
}

// parseTransactionLogs 解析单个交易的日志
func (s *contractParseService) parseTransactionLogs(
	ctx context.Context,
	tx *models.Transaction,
	receipt *models.TransactionReceipt,
	parserConfigs []*models.ParserConfig,
	coinConfigs []*models.CoinConfig,
) ([]*models.ContractParseResult, error) {
	if receipt.LogsData == "" {
		return nil, nil
	}

	// 解析日志数据
	var logs []map[string]interface{}
	if err := json.Unmarshal([]byte(receipt.LogsData), &logs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal logs: %w", err)
	}

	// 计算原始日志hash用于幂等性检查
	rawHash := sha256.Sum256([]byte(receipt.LogsData))
	rawHashHex := "0x" + hex.EncodeToString(rawHash[:])

	// 建立事件签名到配置的映射
	sigToConfig := make(map[string]*models.ParserConfig)
	for _, config := range parserConfigs {
		if config != nil && config.EventSignature != "" {
			sig := strings.ToLower(config.EventSignature)
			sigToConfig[sig] = config
		}
	}

	var results []*models.ContractParseResult

	// 处理每个日志条目
	for logIdx, log := range logs {
		topics, _ := log["topics"].([]interface{})
		data, _ := log["data"].(string)

		if len(topics) == 0 {
			continue
		}

		topic0, _ := topics[0].(string)
		eventSig := strings.ToLower(topic0)

		config := sigToConfig[eventSig]
		if config == nil {
			continue
		}

		// 检查幂等性
		if s.isAlreadyProcessed(ctx, receipt.TxHash, uint(logIdx)) {
			continue
		}

		// 提取日志数据
		fromAddr := s.extractAddress(config.LogsParserRules.ExtractFromAddress, topics, data)
		toAddr := s.extractAddress(config.LogsParserRules.ExtractToAddress, topics, data)
		ownerAddr := s.extractAddress(config.LogsParserRules.ExtractOwnerAddress, topics, data)
		spenderAddr := s.extractAddress(config.LogsParserRules.ExtractSpenderAddress, topics, data)
		amountWei := s.extractAmount(config.LogsParserRules.ExtractAmount, data)

		// 获取币种配置
		coinConfig, err := s.getCoinConfigByContract(coinConfigs, tx.AddressTo)
		if err != nil {
			continue // 跳过没有币种配置的合约
		}
		fmt.Printf("提取日志数据结果： fromAddr: %s, toAddr: %s, ownerAddr: %s, spenderAddr: %s, amountWei: %s\n", fromAddr, toAddr, ownerAddr, spenderAddr, amountWei)

		// 更新用户地址余额
		if err := s.updateUserBalances(ctx, config.EventName, tx, fromAddr, toAddr, ownerAddr, spenderAddr, amountWei, coinConfig, uint(logIdx)); err != nil {
			fmt.Printf("Failed to update balances for tx %s log %d: %v\n", tx.TxID, logIdx, err)
		}

		// 创建解析结果
		logJSON, _ := json.Marshal(log)
		result := &models.ContractParseResult{
			TxHash:          tx.TxID,
			ContractAddress: tx.AddressTo,
			Chain:           receipt.Chain,
			BlockNumber:     receipt.BlockNumber,
			LogIndex:        uint(logIdx),
			EventSignature:  eventSig,
			EventName:       config.EventName,
			FromAddress:     fromAddr,
			ToAddress:       toAddr,
			OwnerAddress:    ownerAddr,
			SpenderAddress:  spenderAddr,
			AmountWei:       amountWei,
			TokenDecimals:   uint16(tx.TokenDecimals),
			TokenSymbol:     coinConfig.Symbol,
			RawLogsHash:     rawHashHex,
			ParsedJSON:      string(logJSON),
		}

		results = append(results, result)
	}

	return results, nil
}

// updateUserBalances 更新用户地址余额
func (s *contractParseService) updateUserBalances(
	ctx context.Context,
	eventName string,
	tx *models.Transaction,
	fromAddr, toAddr, ownerAddr, spenderAddr, amountWei string,
	coinConfig *models.CoinConfig,
	logIdx uint,
) error {
	switch eventName {
	case "Transfer":
		return s.handleTransferEvent(ctx, tx, fromAddr, toAddr, amountWei, coinConfig, logIdx)
	case "TransferFrom":
		return s.handleTransferFromEvent(ctx, tx, fromAddr, toAddr, amountWei, coinConfig, logIdx)
	case "Approve":
		return s.handleApproveEvent(ctx, tx, ownerAddr, spenderAddr, amountWei, coinConfig, logIdx)
	}
	return nil
}

// handleTransferEvent 处理Transfer事件
func (s *contractParseService) handleTransferEvent(
	ctx context.Context,
	tx *models.Transaction,
	fromAddr, toAddr, amountWei string,
	coinConfig *models.CoinConfig,
	logIdx uint,
) error {
	amount, ok := new(big.Int).SetString(amountWei, 10)
	if !ok || amount.Sign() <= 0 {
		return nil
	}

	// 更新发送方余额（减少）
	if fromAddr != "" {
		fromUserAddr, err := s.userAddressRepo.GetByContractAddress(fromAddr, coinConfig.ID)
		if err == nil && fromUserAddr != nil {
			if s.shouldUpdateBalance(fromUserAddr, tx.Height) {
				if err := s.updateContractBalance(fromUserAddr, amount, false); err != nil {
					fmt.Printf("Failed to update from balance: %v\n", err)
				} else {
					s.userAddressRepo.Update(fromUserAddr)
				}
			}
		}
	}

	// 更新接收方余额（增加）
	if toAddr != "" {
		toUserAddr, err := s.userAddressRepo.GetByContractAddress(toAddr, coinConfig.ID)
		if err == nil && toUserAddr != nil {
			if s.shouldUpdateBalance(toUserAddr, tx.Height) {
				if err := s.updateContractBalance(toUserAddr, amount, true); err != nil {
					fmt.Printf("Failed to update to balance: %v\n", err)
				} else {
					s.userAddressRepo.Update(toUserAddr)
				}
			}
		}
	}

	return nil
}

// handleTransferFromEvent 处理TransferFrom事件
func (s *contractParseService) handleTransferFromEvent(
	ctx context.Context,
	tx *models.Transaction,
	fromAddr, toAddr, amountWei string,
	coinConfig *models.CoinConfig,
	logIdx uint,
) error {
	amount, ok := new(big.Int).SetString(amountWei, 10)
	if !ok || amount.Sign() <= 0 {
		return nil
	}

	// 更新发送方授权余额（减少）
	if fromAddr != "" {
		fromUserAddr, err := s.userAddressRepo.GetByContractApproveAddress(fromAddr, coinConfig.ID)
		if err == nil && fromUserAddr != nil {
			if s.shouldUpdateBalance(fromUserAddr, tx.Height) {
				if err := s.updateContractBalance(fromUserAddr, amount, false); err != nil {
					fmt.Printf("Failed to update from approve balance: %v\n", err)
				} else {
					s.userAddressRepo.Update(fromUserAddr)
				}
			}
		}
	}

	// 更新接收方授权余额（增加）
	if toAddr != "" {
		toUserAddr, err := s.userAddressRepo.GetByContractApproveAddress(toAddr, coinConfig.ID)
		if err == nil && toUserAddr != nil {
			if s.shouldUpdateBalance(toUserAddr, tx.Height) {
				if err := s.updateContractBalance(toUserAddr, amount, true); err != nil {
					fmt.Printf("Failed to update to approve balance: %v\n", err)
				} else {
					s.userAddressRepo.Update(toUserAddr)
				}
			}
		}
	}

	return nil
}

// handleApproveEvent 处理Approve事件
func (s *contractParseService) handleApproveEvent(
	ctx context.Context,
	tx *models.Transaction,
	ownerAddr, spenderAddr, amountWei string,
	coinConfig *models.CoinConfig,
	logIdx uint,
) error {
	if ownerAddr == "" || spenderAddr == "" {
		return nil
	}

	ownerUserAddr, err := s.userAddressRepo.GetByContractAddress(ownerAddr, coinConfig.ID)
	if err != nil || ownerUserAddr == nil {
		return nil
	}

	if !s.shouldUpdateBalance(ownerUserAddr, tx.Height) {
		return nil
	}

	// 更新授权地址
	if ownerUserAddr.AuthorizedAddresses == nil {
		ownerUserAddr.AuthorizedAddresses = make(models.AuthorizedAddressesJSON)
	}

	ownerUserAddr.AuthorizedAddresses[spenderAddr] = models.AuthorizedInfo{
		Allowance: amountWei,
	}

	return s.userAddressRepo.Update(ownerUserAddr)
}

// handleBalanceOfEvent 处理BalanceOf事件
func (s *contractParseService) handleBalanceOfEvent(
	ctx context.Context,
	tx *models.Transaction,
	ownerAddr, amountWei string,
	coinConfig *models.CoinConfig,
	logIdx uint,
) error {
	if ownerAddr == "" || amountWei == "0" {
		return nil
	}

	ownerUserAddr, err := s.userAddressRepo.GetByContractAddress(ownerAddr, coinConfig.ID)
	if err != nil || ownerUserAddr == nil {
		return nil
	}

	if !s.shouldUpdateBalance(ownerUserAddr, tx.Height) {
		return nil
	}

	// 设置绝对余额
	ownerUserAddr.ContractBalance = &amountWei

	return s.userAddressRepo.Update(ownerUserAddr)
}

// shouldUpdateBalance 检查是否应该更新余额（基于区块高度）
func (s *contractParseService) shouldUpdateBalance(userAddr *models.UserAddress, currentHeight uint64) bool {
	return userAddr.BalanceHeight == 0 || userAddr.BalanceHeight <= currentHeight
}

// updateContractBalance 更新合约余额
func (s *contractParseService) updateContractBalance(userAddr *models.UserAddress, amount *big.Int, isAdd bool) error {
	currentBalance := new(big.Int)
	if userAddr.ContractBalance != nil {
		var ok bool
		currentBalance, ok = currentBalance.SetString(*userAddr.ContractBalance, 10)
		if !ok {
			currentBalance.SetInt64(0)
		}
	}

	if isAdd {
		currentBalance.Add(currentBalance, amount)
	} else {
		currentBalance.Sub(currentBalance, amount)
		if currentBalance.Sign() < 0 {
			currentBalance.SetInt64(0)
		}
	}

	balanceStr := currentBalance.String()
	userAddr.ContractBalance = &balanceStr
	return nil
}

// isAlreadyProcessed 检查是否已经处理过该日志
func (s *contractParseService) isAlreadyProcessed(ctx context.Context, txHash string, logIdx uint) bool {
	existing, err := s.repo.GetByTxHashAndLogIndex(ctx, txHash, logIdx)
	return err == nil && existing != nil
}

// extractAddress 根据规则提取地址
func (s *contractParseService) extractAddress(rule string, topics []interface{}, data string) string {
	if rule == "" || len(topics) == 0 {
		return ""
	}

	ruleLower := strings.ToLower(rule)
	if strings.HasPrefix(ruleLower, "topics[") {
		idx := s.extractIndexFromBracket(ruleLower)
		if idx > 0 && idx < len(topics) {
			if topic, ok := topics[idx].(string); ok && len(topic) >= 66 {
				return "0x" + strings.ToLower(topic[len(topic)-40:])
			}
		}
	} else if strings.HasPrefix(ruleLower, "data[") {
		idx := s.extractIndexFromBracket(ruleLower)
		slot := s.getDataSlot(data, idx)
		if slot != "" && len(slot) == 64 {
			return "0x" + strings.ToLower(slot[24:])
		}
	}

	return ""
}

// extractAmount 根据规则提取金额
func (s *contractParseService) extractAmount(rule string, data string) string {
	if rule == "" || data == "" {
		return "0"
	}

	ruleLower := strings.ToLower(rule)
	if strings.HasPrefix(ruleLower, "data[") {
		idx := s.extractIndexFromBracket(ruleLower)
		slot := s.getDataSlot(data, idx)
		if slot != "" {
			return s.hexToDecimal(slot)
		}
	}

	return "0"
}

// extractIndexFromBracket 从括号中提取索引
func (s *contractParseService) extractIndexFromBracket(expr string) int {
	start := strings.Index(expr, "[")
	end := strings.Index(expr, "]")
	if start >= 0 && end > start+1 {
		var idx int
		fmt.Sscanf(expr[start+1:end], "%d", &idx)
		return idx
	}
	return -1
}

// getDataSlot 获取data中指定索引的32字节数据槽
func (s *contractParseService) getDataSlot(data string, idx int) string {
	if data == "" || idx < 0 {
		return ""
	}

	dataHex := strings.TrimPrefix(strings.ToLower(data), "0x")
	start := idx * 64
	end := start + 64

	if start < 0 || end > len(dataHex) {
		return ""
	}

	return dataHex[start:end]
}

// hexToDecimal 将十六进制字符串转换为十进制字符串
func (s *contractParseService) hexToDecimal(hexStr string) string {
	if hexStr == "" {
		return "0"
	}

	// 移除0x前缀并去掉前导0
	hex := strings.TrimPrefix(strings.ToLower(hexStr), "0x")
	for len(hex) > 0 && hex[0] == '0' {
		hex = hex[1:]
	}

	if hex == "" {
		return "0"
	}

	num := new(big.Int)
	if _, ok := num.SetString(hex, 16); !ok {
		return "0"
	}

	return num.String()
}

// getCoinConfigByContract 根据合约地址获取币种配置
func (s *contractParseService) getCoinConfigByContract(coinConfigs []*models.CoinConfig, contractAddr string) (*models.CoinConfig, error) {
	for _, config := range coinConfigs {
		if strings.ToLower(config.ContractAddr) == strings.ToLower(contractAddr) {
			return config, nil
		}
	}
	return nil, fmt.Errorf("coin config not found for contract address: %s", contractAddr)
}

// GetByTxHash 根据交易哈希获取解析结果
func (s *contractParseService) GetByTxHash(ctx context.Context, hash string) ([]*models.ContractParseResult, error) {
	return s.repo.GetByTxHash(ctx, hash)
}
