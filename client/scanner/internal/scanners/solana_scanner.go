package scanners

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"blockChainBrowser/client/scanner/config"
	"blockChainBrowser/client/scanner/internal/failover"
	"blockChainBrowser/client/scanner/internal/models"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/sirupsen/logrus"
)

// SolanaScanner 基于 jsonParsed 数据的简化 Solana 扫块器
type SolanaScanner struct {
	config          *config.ChainConfig
	rpcURL          string
	failoverManager *failover.SOLFailoverManager
}

// NewSolanaScanner 创建新的 Solana 扫块器
func NewSolanaScanner(cfg *config.ChainConfig) *SolanaScanner {
	// 初始化主客户端
	var mainClient *client.Client
	if cfg.RPCURL != "" {
		mainClient = client.NewClient(cfg.RPCURL)
		logrus.Infof("Initialized Solana main RPC client: %s", cfg.RPCURL)
	}

	// 初始化故障转移客户端
	var foClients []*client.Client
	if len(cfg.ExplorerAPIURLs) > 0 {
		for _, u := range cfg.ExplorerAPIURLs {
			foClients = append(foClients, client.NewClient(u))
			logrus.Infof("Initialized Solana failover RPC client: %s", u)
		}
	}

	fm := failover.NewSOLFailoverManager(mainClient, foClients)

	return &SolanaScanner{
		config:          cfg,
		rpcURL:          cfg.RPCURL,
		failoverManager: fm,
	}
}

// GetLatestBlockHeight 获取最新区块高度
func (s *SolanaScanner) GetLatestBlockHeight() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slot, err := s.failoverManager.GetSlot(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest slot: %w", err)
	}
	return slot, nil
}

// GetBlockByHeight 根据高度获取区块
func (s *SolanaScanner) GetBlockByHeight(height uint64) (*models.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	blockData, err := s.failoverManager.GetBlockRaw(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %d: %w", height, err)
	}
	return s.parseBlock(blockData, height), nil
}

// ValidateBlock 验证区块
func (s *SolanaScanner) ValidateBlock(block *models.Block) error {
	if block == nil {
		return fmt.Errorf("block is nil")
	}
	if block.Height == 0 {
		return fmt.Errorf("invalid block height: %d", block.Height)
	}
	if block.Hash == "" {
		return fmt.Errorf("block hash is empty")
	}
	if block.Timestamp.IsZero() {
		return fmt.Errorf("invalid block timestamp")
	}
	return nil
}

// GetBlockTransactionsFromBlock 从区块获取交易
func (s *SolanaScanner) GetBlockTransactionsFromBlock(block *models.Block) ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	blockData, err := s.failoverManager.GetBlockRaw(ctx, block.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block transactions for height %d: %w", block.Height, err)
	}

	// 解析所有交易
	allTransactions := s.parseTransactions(blockData, block.Height)

	// 过滤掉投票交易
	filteredTransactions := s.filterVoteTransactions(allTransactions)

	logrus.Debugf("[sol] Filtered transactions: %d -> %d (removed %d vote transactions)",
		len(allTransactions), len(filteredTransactions), len(allTransactions)-len(filteredTransactions))
	// fmt.Println("已经过滤掉投票交易的数量是", len(allTransactions)-len(filteredTransactions))
	// fmt.Println("投票交易占用的比例", (float64(len(filteredTransactions))/float64(len(allTransactions)))*100, "%")
	return filteredTransactions, nil
}

// CalculateBlockStats 计算区块统计信息
func (s *SolanaScanner) CalculateBlockStats(block *models.Block, transactions []map[string]interface{}) {
	block.TransactionCount = len(transactions)

	var totalFees float64
	var totalAmount float64

	for _, tx := range transactions {
		// 费用统计
		if fee, ok := tx["fee"].(uint64); ok {
			totalFees += float64(fee)
		}

		// 转账金额统计
		if value, ok := tx["value"].(string); ok && value != "0" {
			if amount, err := strconv.ParseFloat(value, 64); err == nil {
				totalAmount += amount
			}
		}
	}

	block.Fee = totalFees
	block.TotalAmount = totalAmount
	block.Size = uint64(len(transactions) * 200) // 估算大小
}

// makeRPCCall 执行 RPC 调用
// 移除直接HTTP调用，统一通过 failover manager 调用 RPC

// parseBlock 解析区块数据
func (s *SolanaScanner) parseBlock(blockData map[string]interface{}, height uint64) *models.Block {
	block := &models.Block{
		Height:           height,
		Hash:             getStringValue(blockData, "blockhash"),
		PreviousHash:     getStringValue(blockData, "previousBlockhash"),
		Timestamp:        time.Now(),
		TransactionCount: 0,
		Size:             0,
		Miner:            "",
		Difficulty:       0,
		Nonce:            0,
		Chain:            s.config.Name,
		ChainID:          s.config.ChainID,
	}

	// 设置时间戳
	if blockTime, ok := blockData["blockTime"].(float64); ok {
		block.Timestamp = time.Unix(int64(blockTime), 0)
	}

	// 设置交易数量
	if transactions, ok := blockData["transactions"].([]interface{}); ok {
		block.TransactionCount = len(transactions)
	}

	// 设置矿工（从奖励中获取）
	if rewards, ok := blockData["rewards"].([]interface{}); ok && len(rewards) > 0 {
		if reward, ok := rewards[0].(map[string]interface{}); ok {
			block.Miner = getStringValue(reward, "pubkey")
		}
	}

	return block
}

// parseTransactions 解析交易数据
func (s *SolanaScanner) parseTransactions(blockData map[string]interface{}, blockHeight uint64) []map[string]interface{} {
	var transactions []map[string]interface{}

	txsData, ok := blockData["transactions"].([]interface{})
	if !ok {
		return transactions
	}

	for i, txData := range txsData {
		if txMap, ok := txData.(map[string]interface{}); ok {
			tx := s.parseTransaction(txMap, blockHeight, i)
			transactions = append(transactions, tx)
		}
	}

	return transactions
}

// parseTransaction 解析单个交易
func (s *SolanaScanner) parseTransaction(txData map[string]interface{}, blockHeight uint64, index int) map[string]interface{} {
	tx := map[string]interface{}{
		"hash":             "",
		"block_height":     blockHeight,
		"block_hash":       "",
		"from":             "",
		"to":               "",
		"value":            "0",
		"fee":              uint64(0),
		"block_index":      index,
		"gas_used":         0,
		"gas_price":        "0",
		"gas_limit":        0,
		"nonce":            0,
		"status":           "success",
		"timestamp":        time.Now().Unix(),
		"chain_id":         s.config.ChainID,
		"chain_name":       s.config.Name,
		"transaction_type": "sol",
	}

	// 解析交易基本信息
	if transaction, ok := txData["transaction"].(map[string]interface{}); ok {
		// 获取签名（作为交易哈希）
		if signatures, ok := transaction["signatures"].([]interface{}); ok && len(signatures) > 0 {
			if sig, ok := signatures[0].(string); ok {
				tx["hash"] = sig
			}
		}

		// 解析消息
		if message, ok := transaction["message"].(map[string]interface{}); ok {
			s.parseMessage(message, tx)
		}
	}

	// 解析元数据
	if meta, ok := txData["meta"].(map[string]interface{}); ok {
		s.parseMeta(meta, tx)
	}

	// 版本信息
	if version, ok := txData["version"]; ok {
		tx["version"] = version
	} else {
		tx["version"] = "legacy"
	}

	return tx
}

// parseMessage 解析交易消息
func (s *SolanaScanner) parseMessage(message map[string]interface{}, tx map[string]interface{}) {
	// 获取账户密钥
	if accountKeys, ok := message["accountKeys"].([]interface{}); ok && len(accountKeys) > 0 {
		tx["account_keys"] = accountKeys

		// 设置 from（第一个签名账户）
		if firstAccount, ok := accountKeys[0].(map[string]interface{}); ok {
			if pubkey, ok := firstAccount["pubkey"].(string); ok {
				tx["from"] = pubkey
			}
		}

		// 解析指令
		s.parseInstructions(message, accountKeys, tx)
	}

	// 最近区块哈希
	if recentBlockhash, ok := message["recentBlockhash"].(string); ok {
		tx["recent_blockhash"] = recentBlockhash
	}

	// 地址表查找
	if addressTableLookups, ok := message["addressTableLookups"]; ok {
		tx["address_table_lookups"] = addressTableLookups
	}
}

// parseInstructions 解析指令
func (s *SolanaScanner) parseInstructions(message map[string]interface{}, accountKeys []interface{}, tx map[string]interface{}) {
	instructions, ok := message["instructions"].([]interface{})
	if !ok {
		return
	}

	var solInstructions []map[string]interface{}
	var events []map[string]interface{}

	for i, instData := range instructions {
		if inst, ok := instData.(map[string]interface{}); ok {
			instruction := s.parseInstruction(inst, accountKeys, i, false)
			solInstructions = append(solInstructions, instruction)

			// 解析特定类型的指令以生成事件
			if event := s.parseInstructionToEvent(inst, accountKeys, tx, i); event != nil {
				events = append(events, event)
			}
		}
	}

	if len(solInstructions) > 0 {
		tx["sol_instructions"] = solInstructions
	}
	if len(events) > 0 {
		tx["sol_events"] = events
	}
}

// parseInstruction 解析单个指令
func (s *SolanaScanner) parseInstruction(inst map[string]interface{}, accountKeys []interface{}, index int, isInner bool) map[string]interface{} {
	instruction := map[string]interface{}{
		"index":      index,
		"is_inner":   isInner,
		"accounts":   []string{},
		"data":       "",
		"program_id": "",
	}

	// 程序 ID
	if programId, ok := inst["programId"].(string); ok {
		instruction["program_id"] = programId
	}

	// 账户
	if accounts, ok := inst["accounts"].([]interface{}); ok {
		var accountAddrs []string
		for _, acc := range accounts {
			if accStr, ok := acc.(string); ok {
				accountAddrs = append(accountAddrs, accStr)
			}
		}
		instruction["accounts"] = accountAddrs
	}

	// 数据
	if data, ok := inst["data"].(string); ok {
		instruction["data"] = data
	}

	// 解析的指令信息
	if parsed, ok := inst["parsed"].(map[string]interface{}); ok {
		instruction["parsed"] = parsed
		if instType, ok := parsed["type"].(string); ok {
			instruction["type"] = instType
		}
	}

	return instruction
}

// parseInstructionToEvent 将指令解析为事件
func (s *SolanaScanner) parseInstructionToEvent(inst map[string]interface{}, accountKeys []interface{}, tx map[string]interface{}, index int) map[string]interface{} {
	programId, _ := inst["programId"].(string)

	// 解析 parsed 指令
	if parsed, ok := inst["parsed"].(map[string]interface{}); ok {
		if instType, ok := parsed["type"].(string); ok {
			if info, ok := parsed["info"].(map[string]interface{}); ok {
				switch instType {
				case "transfer":
					return s.parseTransferEvent(programId, info, tx, index)
				case "transferChecked":
					return s.parseTransferCheckedEvent(programId, info, tx, index)
				}
			}
		}
	}

	return nil
}

// parseTransferEvent 解析转账事件
func (s *SolanaScanner) parseTransferEvent(programId string, info map[string]interface{}, tx map[string]interface{}, index int) map[string]interface{} {
	if programId == "11111111111111111111111111111111" {
		// 系统转账
		source, _ := info["source"].(string)
		destination, _ := info["destination"].(string)
		lamports, _ := info["lamports"].(float64)

		if source != "" && destination != "" {
			// 设置交易的 to 和 value
			if tx["to"] == "" {
				tx["to"] = destination
			}
			if tx["value"] == "0" {
				tx["value"] = fmt.Sprintf("%.0f", lamports)
			}

			return map[string]interface{}{
				"type":        "system_transfer",
				"from":        source,
				"to":          destination,
				"lamports":    fmt.Sprintf("%.0f", lamports),
				"program_id":  programId,
				"block_index": index,
				"is_inner":    false,
			}
		}
	}

	return nil
}

// parseTransferCheckedEvent 解析带检查的转账事件
func (s *SolanaScanner) parseTransferCheckedEvent(programId string, info map[string]interface{}, tx map[string]interface{}, index int) map[string]interface{} {
	if tokenAmount, ok := info["tokenAmount"].(map[string]interface{}); ok {
		amount, _ := tokenAmount["amount"].(string)
		decimals, _ := tokenAmount["decimals"].(float64)
		source, _ := info["source"].(string)
		destination, _ := info["destination"].(string)
		mint, _ := info["mint"].(string)

		return map[string]interface{}{
			"type":         "spl_transfer",
			"from_account": source,
			"to_account":   destination,
			"amount":       amount,
			"mint":         mint,
			"program_id":   programId,
			"block_index":  index,
			"is_inner":     false,
			"decimals":     int(decimals),
		}
	}

	return nil
}

// parseMeta 解析元数据
func (s *SolanaScanner) parseMeta(meta map[string]interface{}, tx map[string]interface{}) {
	// 费用
	if fee, ok := meta["fee"].(float64); ok {
		tx["fee"] = uint64(fee)
		tx["gas_used"] = uint64(fee)
	}

	// 计算单元
	if computeUnits, ok := meta["computeUnitsConsumed"].(float64); ok {
		tx["compute_units"] = uint64(computeUnits)
	}

	// 余额变化
	if preBalances, ok := meta["preBalances"]; ok {
		tx["pre_balances"] = preBalances
	}
	if postBalances, ok := meta["postBalances"]; ok {
		tx["post_balances"] = postBalances
	}

	// 代币余额变化
	if preTokenBalances, ok := meta["preTokenBalances"]; ok {
		tx["pre_token_balances"] = preTokenBalances
	}
	if postTokenBalances, ok := meta["postTokenBalances"]; ok {
		tx["post_token_balances"] = postTokenBalances
	}

	// 日志
	if logs, ok := meta["logMessages"]; ok {
		tx["logs"] = logs
	}

	// 内部指令
	if innerInstructions, ok := meta["innerInstructions"]; ok {
		tx["inner_instructions"] = innerInstructions
	}

	// 加载的地址
	if loadedAddresses, ok := meta["loadedAddresses"]; ok {
		tx["loaded_addresses"] = loadedAddresses
	}

	// 错误状态
	if err, ok := meta["err"]; ok && err != nil {
		tx["status"] = "failed"
	}

	// 奖励
	if rewards, ok := meta["rewards"]; ok {
		tx["rewards"] = rewards
	}
}

// ProcessSolanaArtifacts 处理 Solana 特定的工件
func (s *SolanaScanner) ProcessSolanaArtifacts(transactions []map[string]interface{}, block interface{}, blockID uint64) {
	// 转换 block 参数
	var blockHeight uint64
	var blockHash string
	if b, ok := block.(map[string]interface{}); ok {
		if h, exists := b["height"]; exists {
			if height, ok := h.(uint64); ok {
				blockHeight = height
			}
		}
		if h, exists := b["hash"]; exists {
			if hash, ok := h.(string); ok {
				blockHash = hash
			}
		}
	}

	// 检查是否启用批量上传
	if s.config.Scan.BatchUpload {
		s.uploadSolTxDetailsBatch(transactions, blockHeight, blockHash, blockID)
	} else {
		// 单笔上传模式
		for _, tx := range transactions {
			if txID, ok := tx["hash"].(string); ok && txID != "" {
				s.uploadSolTxDetail(tx, blockHeight, blockHash, blockID)
			}
		}
	}
}

// uploadSolTxDetail 上传单笔 Solana 交易明细
func (s *SolanaScanner) uploadSolTxDetail(tx map[string]interface{}, blockHeight uint64, blockHash string, blockID uint64) {
	// 构建交易明细请求
	detail := map[string]interface{}{
		"tx_id":               tx["hash"],
		"slot":                blockHeight,
		"block_id":            blockID,
		"blockhash":           blockHash,
		"recent_blockhash":    getStringFromMap(tx, "recent_blockhash"),
		"version":             getStringFromMap(tx, "version"),
		"fee":                 getUint64FromMap(tx, "fee"),
		"compute_units":       getUint64FromMap(tx, "compute_units"),
		"status":              getStringFromMap(tx, "status"),
		"account_keys":        toJSONString(tx["account_keys"]),
		"pre_balances":        toJSONString(tx["pre_balances"]),
		"post_balances":       toJSONString(tx["post_balances"]),
		"pre_token_balances":  toJSONString(tx["pre_token_balances"]),
		"post_token_balances": toJSONString(tx["post_token_balances"]),
		"logs":                toJSONString(tx["logs"]),
		"instructions":        toJSONString(tx["sol_instructions"]),
		"inner_instructions":  toJSONString(tx["inner_instructions"]),
		"loaded_addresses":    toJSONString(tx["loaded_addresses"]),
		"rewards":             toJSONString(tx["rewards"]),
		"events":              toJSONString(tx["sol_events"]),
		"raw_transaction":     toJSONString(tx),
		"raw_meta":            "{}",
	}

	// 构建事件数据
	var events []map[string]interface{}
	if solEvents, ok := tx["sol_events"].([]map[string]interface{}); ok {
		for i, event := range solEvents {
			eventData := map[string]interface{}{
				"tx_id":        tx["hash"],
				"block_id":     blockID,
				"slot":         blockHeight,
				"event_index":  i,
				"event_type":   getStringFromMap(event, "type"),
				"program_id":   getStringFromMap(event, "program_id"),
				"from_address": getStringFromMap(event, "from"),
				"to_address":   getStringFromMap(event, "to"),
				"amount":       getStringFromMap(event, "lamports"),
				"mint":         getStringFromMap(event, "mint"),
				"decimals":     getIntFromMap(event, "decimals"),
				"is_inner":     getBoolFromMap(event, "is_inner"),
				"asset_type":   "NATIVE",
				"extra_data":   toJSONString(event),
			}

			// 根据事件类型调整资产类型
			if eventType := getStringFromMap(event, "type"); eventType == "spl_transfer" {
				eventData["asset_type"] = "SPL"
				eventData["amount"] = getStringFromMap(event, "amount")
			}

			events = append(events, eventData)
		}
	}

	// 构建指令数据
	var instructions []map[string]interface{}
	if solInstructions, ok := tx["sol_instructions"].([]map[string]interface{}); ok {
		for i, inst := range solInstructions {
			instData := map[string]interface{}{
				"tx_id":             tx["hash"],
				"block_id":          blockID,
				"slot":              blockHeight,
				"instruction_index": i,
				"program_id":        getStringFromMap(inst, "program_id"),
				"accounts":          toJSONString(inst["accounts"]),
				"data":              getStringFromMap(inst, "data"),
				"parsed_data":       toJSONString(inst["parsed"]),
				"instruction_type":  getStringFromMap(inst, "type"),
				"is_inner":          getBoolFromMap(inst, "is_inner"),
				"stack_height":      1,
			}
			instructions = append(instructions, instData)
		}
	}

	// 构建完整请求并上传到服务器
	requestBody := map[string]interface{}{
		"detail":       detail,
		"events":       events,
		"instructions": instructions,
	}

	// 获取扫描器API实例并上传
	api := config.GetScannerAPI()
	if api == nil {
		logrus.Warnf("[sol] Scanner API not available, skipping upload for tx %s", tx["hash"])
		return
	}

	if err := api.UploadSolTxDetail(requestBody); err != nil {
		logrus.Errorf("[sol] Failed to upload Sol tx detail for %s: %v", tx["hash"], err)
		return
	}

	logrus.Debugf("[sol] Successfully uploaded Sol tx detail for %s", tx["hash"])
}

// uploadSolTxDetailsBatch 批量上传 Solana 交易明细
func (s *SolanaScanner) uploadSolTxDetailsBatch(transactions []map[string]interface{}, blockHeight uint64, blockHash string, blockID uint64) {
	if len(transactions) == 0 {
		return
	}

	// 构建批量请求
	var batchRequests []map[string]interface{}

	for _, tx := range transactions {
		txID, ok := tx["hash"].(string)
		if !ok || txID == "" {
			continue
		}

		// 构建交易明细
		detail := map[string]interface{}{
			"tx_id":               txID,
			"slot":                blockHeight,
			"block_id":            blockID,
			"blockhash":           blockHash,
			"recent_blockhash":    getStringFromMap(tx, "recent_blockhash"),
			"version":             getStringFromMap(tx, "version"),
			"fee":                 getUint64FromMap(tx, "fee"),
			"compute_units":       getUint64FromMap(tx, "compute_units"),
			"status":              getStringFromMap(tx, "status"),
			"account_keys":        toJSONString(tx["account_keys"]),
			"pre_balances":        toJSONString(tx["pre_balances"]),
			"post_balances":       toJSONString(tx["post_balances"]),
			"pre_token_balances":  toJSONString(tx["pre_token_balances"]),
			"post_token_balances": toJSONString(tx["post_token_balances"]),
			"logs":                toJSONString(tx["logs"]),
			"instructions":        toJSONString(tx["sol_instructions"]),
			"inner_instructions":  toJSONString(tx["inner_instructions"]),
			"loaded_addresses":    toJSONString(tx["loaded_addresses"]),
			"rewards":             toJSONString(tx["rewards"]),
			"events":              toJSONString(tx["sol_events"]),
			"raw_transaction":     toJSONString(tx),
			"raw_meta":            "{}",
		}

		// 构建事件数据
		var events []map[string]interface{}
		if solEvents, ok := tx["sol_events"].([]map[string]interface{}); ok {
			for i, event := range solEvents {
				eventData := map[string]interface{}{
					"tx_id":        txID,
					"block_id":     blockID,
					"slot":         blockHeight,
					"event_index":  i,
					"event_type":   getStringFromMap(event, "type"),
					"program_id":   getStringFromMap(event, "program_id"),
					"from_address": getStringFromMap(event, "from"),
					"to_address":   getStringFromMap(event, "to"),
					"amount":       getStringFromMap(event, "lamports"),
					"mint":         getStringFromMap(event, "mint"),
					"decimals":     getIntFromMap(event, "decimals"),
					"is_inner":     getBoolFromMap(event, "is_inner"),
					"asset_type":   "NATIVE",
					"extra_data":   toJSONString(event),
				}

				// 根据事件类型调整资产类型
				if eventType := getStringFromMap(event, "type"); eventType == "spl_transfer" {
					eventData["asset_type"] = "SPL"
					eventData["amount"] = getStringFromMap(event, "amount")
				}

				events = append(events, eventData)
			}
		}

		// 构建指令数据
		var instructions []map[string]interface{}
		if solInstructions, ok := tx["sol_instructions"].([]map[string]interface{}); ok {
			for i, inst := range solInstructions {
				instData := map[string]interface{}{
					"tx_id":             txID,
					"block_id":          blockID,
					"slot":              blockHeight,
					"instruction_index": i,
					"program_id":        getStringFromMap(inst, "program_id"),
					"accounts":          toJSONString(inst["accounts"]),
					"data":              getStringFromMap(inst, "data"),
					"parsed_data":       toJSONString(inst["parsed"]),
					"instruction_type":  getStringFromMap(inst, "type"),
					"is_inner":          getBoolFromMap(inst, "is_inner"),
					"stack_height":      1,
				}
				instructions = append(instructions, instData)
			}
		}

		// 构建单个交易的完整请求
		txRequest := map[string]interface{}{
			"detail":       detail,
			"events":       events,
			"instructions": instructions,
		}

		batchRequests = append(batchRequests, txRequest)
	}

	if len(batchRequests) == 0 {
		return
	}

	// 构建批量请求体
	batchRequestBody := map[string]interface{}{
		"transactions": batchRequests,
	}

	// 获取扫描器API实例并批量上传
	api := config.GetScannerAPI()
	if api == nil {
		logrus.Warnf("[sol] Scanner API not available, skipping batch upload for %d transactions", len(batchRequests))
		return
	}

	// 调用批量上传API
	if err := api.UploadSolTxDetailBatch(batchRequestBody); err != nil {
		logrus.Errorf("[sol] Failed to batch upload Sol tx details (%d transactions): %v", len(batchRequests), err)

		// 批量上传失败时，回退到单笔上传模式
		logrus.Infof("[sol] Falling back to individual upload mode for %d transactions", len(transactions))
		for _, tx := range transactions {
			if txID, ok := tx["hash"].(string); ok && txID != "" {
				s.uploadSolTxDetail(tx, blockHeight, blockHash, blockID)
			}
		}
		return
	}

	logrus.Infof("[sol] Successfully batch uploaded %d Sol tx details for block %d", len(batchRequests), blockHeight)
}

// getStringValue 安全获取字符串值
func getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}

// getStringFromMap 从map中安全获取字符串值
func getStringFromMap(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}

// getUint64FromMap 从map中安全获取uint64值
func getUint64FromMap(data map[string]interface{}, key string) uint64 {
	if value, ok := data[key].(uint64); ok {
		return value
	}
	if value, ok := data[key].(float64); ok {
		return uint64(value)
	}
	return 0
}

// getIntFromMap 从map中安全获取int值
func getIntFromMap(data map[string]interface{}, key string) int {
	if value, ok := data[key].(int); ok {
		return value
	}
	if value, ok := data[key].(float64); ok {
		return int(value)
	}
	return 0
}

// getBoolFromMap 从map中安全获取bool值
func getBoolFromMap(data map[string]interface{}, key string) bool {
	if value, ok := data[key].(bool); ok {
		return value
	}
	return false
}

// toJSONString 将数据转换为JSON字符串
func toJSONString(data interface{}) string {
	if data == nil {
		return "[]"
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "[]"
	}

	return string(jsonBytes)
}

// filterVoteTransactions 过滤掉投票交易
func (s *SolanaScanner) filterVoteTransactions(transactions []map[string]interface{}) []map[string]interface{} {
	// Solana 投票程序 ID
	voteProgramID := "Vote111111111111111111111111111111111111111"

	var filtered []map[string]interface{}

	for _, tx := range transactions {
		// 检查交易是否包含投票程序
		if s.isVoteTransaction(tx, voteProgramID) {
			continue // 跳过投票交易
		}
		filtered = append(filtered, tx)
	}

	return filtered
}

// isVoteTransaction 判断是否为投票交易
func (s *SolanaScanner) isVoteTransaction(tx map[string]interface{}, voteProgramID string) bool {
	// 检查 sol_instructions 字段
	if solInstructions, ok := tx["sol_instructions"].([]map[string]interface{}); ok {
		for _, instruction := range solInstructions {
			if programID, ok := instruction["program_id"].(string); ok {
				if programID == voteProgramID {
					return true
				}
			}
		}
	}

	// 检查 account_keys 字段中的程序 ID
	if accountKeys, ok := tx["account_keys"].([]interface{}); ok {
		for _, key := range accountKeys {
			if keyStr, ok := key.(string); ok && keyStr == voteProgramID {
				return true
			}
		}
	}

	// 检查指令中的程序 ID
	if instructions, ok := tx["instructions"].([]map[string]interface{}); ok {
		for _, instruction := range instructions {
			if programID, ok := instruction["program_id"].(string); ok {
				if programID == voteProgramID {
					return true
				}
			}
		}
	}

	return false
}
