package scanners

import (
	"context"
	"fmt"
	"time"

	"blockChainBrowser/client/scanner/config"
	"blockChainBrowser/client/scanner/internal/failover"
	"blockChainBrowser/client/scanner/internal/models"
	"blockChainBrowser/client/scanner/pkg"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/mr-tron/base58"
	"github.com/sirupsen/logrus"
)

// SolanaScanner 使用 blocto/solana-go-sdk 的 Solana 扫块器
type SolanaScanner struct {
	config          *config.ChainConfig
	failoverManager *failover.SOLFailoverManager
}

// NewSolanaScanner 创建新的 Solana 扫块器（使用 blocto/solana-go-sdk）
func NewSolanaScanner(cfg *config.ChainConfig) *SolanaScanner {
	// 初始化主 RPC 客户端
	var mainClient *client.Client
	if cfg.RPCURL != "" {
		mainClient = client.NewClient(cfg.RPCURL)
		logrus.Infof("Initialized Solana main RPC client: %s", cfg.RPCURL)
	}

	// 初始化多个外部API客户端作为故障转移
	failoverClients := make([]*client.Client, 0)
	if len(cfg.ExplorerAPIURLs) > 0 {
		for _, apiURL := range cfg.ExplorerAPIURLs {
			failoverClient := client.NewClient(apiURL)
			failoverClients = append(failoverClients, failoverClient)
			logrus.Infof("Initialized Solana failover RPC client: %s", apiURL)
		}
	}

	// 创建故障转移管理器
	failoverManager := failover.NewSOLFailoverManager(mainClient, failoverClients)

	scanner := &SolanaScanner{
		config:          cfg,
		failoverManager: failoverManager,
	}

	logrus.Infof("Initialized Solana scanner with %d failover clients", len(failoverClients))
	return scanner
}

// GetLatestBlockHeight 获取最新区块高度
func (s *SolanaScanner) GetLatestBlockHeight() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.failoverManager.GetSlot(ctx)
}

// GetBlockByHeight 根据高度获取区块
func (s *SolanaScanner) GetBlockByHeight(height uint64) (*models.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	block, err := s.failoverManager.GetBlock(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %d: %v", height, err)
	}

	logrus.Infof("[sol] GetBlock(height=%d) using blocto/solana-go-sdk succeeded", height)
	return s.convertToGenericBlockSDK(block, height), nil
}

// ValidateBlock 验证区块
func (s *SolanaScanner) ValidateBlock(block *models.Block) error {
	// 基本验证
	if block == nil {
		return fmt.Errorf("block is nil")
	}
	if block.Height == 0 {
		return fmt.Errorf("invalid block height: %d", block.Height)
	}
	if block.Hash == "" {
		return fmt.Errorf("block hash is empty")
	}

	// 验证时间戳
	if block.Timestamp.IsZero() {
		return fmt.Errorf("invalid block timestamp: %v", block.Timestamp)
	}

	return nil
}

// convertToGenericBlockSDK 转换新SDK的区块格式为通用区块格式
func (s *SolanaScanner) convertToGenericBlockSDK(block *client.Block, height uint64) *models.Block {
	genericBlock := &models.Block{
		Height:           height,
		Hash:             block.Blockhash,
		PreviousHash:     block.PreviousBlockhash,
		Timestamp:        time.Now(), // 默认值
		TransactionCount: len(block.Transactions),
		Size:             uint64(len(block.Transactions) * 200), // 估算大小
		Miner:            "",                                    // Solana没有矿工概念
		Difficulty:       0,                                     // Solana不使用工作量证明
		Nonce:            0,
		Chain:            s.config.Name,
		ChainID:          s.config.ChainID,
	}

	// 设置时间戳
	if block.BlockTime != nil {
		genericBlock.Timestamp = *block.BlockTime
	}

	// 计算总费用
	var totalFees float64
	for _, tx := range block.Transactions {
		if tx.Meta != nil {
			totalFees += float64(tx.Meta.Fee)
		}
	}
	genericBlock.Fee = totalFees

	// 设置矿工（使用第一个奖励接收者）
	if len(block.Rewards) > 0 {
		genericBlock.Miner = block.Rewards[0].Pubkey.String()
	}

	return genericBlock
}

// convertToGenericTransactionSDK 转换新SDK的交易格式为通用交易格式
func (s *SolanaScanner) convertToGenericTransactionSDK(blockTx *client.BlockTransaction, blockHeight uint64, index int) map[string]interface{} {
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
		"gasUsed":          0,
		"gas_price":        "0",
		"gasPrice":         "0",
		"gas_limit":        0,
		"gasLimit":         0,
		"nonce":            0,
		"status":           "success",
		"timestamp":        time.Now().Unix(),
		"chain_id":         s.config.ChainID,
		"chain_name":       s.config.Name,
		"transaction_type": "sol",
	}

	// 设置交易哈希
	if len(blockTx.Transaction.Signatures) > 0 {
		// 交易哈希是64字节的字节数组，需要转换为Base58编码的字符串
		tx["hash"] = base58.Encode(blockTx.Transaction.Signatures[0][:])
	}

	// 设置费用信息
	if blockTx.Meta != nil {
		fee := uint64(blockTx.Meta.Fee)
		tx["fee"] = fee
		tx["gas_used"] = fee
		tx["gasUsed"] = fee
	}

	// 设置from/to地址
	if len(blockTx.AccountKeys) > 0 {
		tx["from"] = blockTx.AccountKeys[0].String()
		if len(blockTx.AccountKeys) > 1 {
			tx["to"] = blockTx.AccountKeys[1].String()
		}
	}

	// 设置交易状态
	if blockTx.Meta != nil && blockTx.Meta.Err != nil {
		tx["status"] = "failed"
	}

	// 使用SDK数据提取指令信息（保持向后兼容）
	s.extractInstructionsFromSDK(blockTx, tx, blockHeight, index)

	// 减少日志输出，只在调试模式下记录
	logrus.Debugf("[sol] converted transaction %s (height=%d, idx=%d)", tx["hash"], blockHeight, index)

	return tx
}

// extractTransactionsFromJsonParsed 从jsonParsed格式的区块数据中提取交易
func (s *SolanaScanner) extractTransactionsFromJsonParsed(blockData map[string]interface{}, blockHeight uint64) ([]map[string]interface{}, error) {
	transactions := make([]map[string]interface{}, 0)

	// 获取交易数组
	txns, ok := blockData["transactions"].([]interface{})
	if !ok {
		return transactions, nil
	}

	for i, txData := range txns {
		txMap, ok := txData.(map[string]interface{})
		if !ok {
			continue
		}

		// 转换单个交易
		tx := s.convertJsonParsedTransaction(txMap, blockHeight, i)
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// convertJsonParsedTransaction 转换jsonParsed格式的交易
func (s *SolanaScanner) convertJsonParsedTransaction(txData map[string]interface{}, blockHeight uint64, index int) map[string]interface{} {
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
		"gasUsed":          0,
		"gas_price":        "0",
		"gasPrice":         "0",
		"gas_limit":        0,
		"gasLimit":         0,
		"nonce":            0,
		"status":           "success",
		"timestamp":        time.Now().Unix(),
		"chain_id":         s.config.ChainID,
		"chain_name":       s.config.Name,
		"transaction_type": "sol",
	}

	// 提取交易哈希
	if signatures, ok := txData["signatures"].([]interface{}); ok && len(signatures) > 0 {
		if sig, ok := signatures[0].(string); ok {
			tx["hash"] = sig
		}
	}

	// 提取费用信息
	if meta, ok := txData["meta"].(map[string]interface{}); ok {
		if fee, ok := meta["fee"].(float64); ok {
			tx["fee"] = uint64(fee)
			tx["gas_used"] = uint64(fee)
			tx["gasUsed"] = uint64(fee)
		}

		// 设置交易状态
		if err, ok := meta["err"]; ok && err != nil {
			tx["status"] = "failed"
		}
	}

	// 提取账户信息
	if transaction, ok := txData["transaction"].(map[string]interface{}); ok {
		if message, ok := transaction["message"].(map[string]interface{}); ok {
			if accountKeys, ok := message["accountKeys"].([]interface{}); ok && len(accountKeys) > 0 {
				// 设置from/to地址
				if fromKey, ok := accountKeys[0].(map[string]interface{}); ok {
					if pubkey, ok := fromKey["pubkey"].(string); ok {
						tx["from"] = pubkey
					}
				}
				if len(accountKeys) > 1 {
					if toKey, ok := accountKeys[1].(map[string]interface{}); ok {
						if pubkey, ok := toKey["pubkey"].(string); ok {
							tx["to"] = pubkey
						}
					}
				}

				// 提取指令信息
				s.extractInstructionsFromJsonParsed(message, tx, blockHeight, index)
			}
		}
	}

	return tx
}

// extractInstructionsFromJsonParsed 从jsonParsed格式的交易中提取指令信息
func (s *SolanaScanner) extractInstructionsFromJsonParsed(message map[string]interface{}, tx map[string]interface{}, blockHeight uint64, index int) {
	simpleInstructions := make([]map[string]interface{}, 0)
	decodedEvents := make([]map[string]interface{}, 0)

	// 获取指令数组
	instructions, ok := message["instructions"].([]interface{})
	if !ok {
		return
	}

	// 获取账户密钥
	accountKeys, ok := message["accountKeys"].([]interface{})
	if !ok {
		return
	}

	for i, instData := range instructions {
		inst, ok := instData.(map[string]interface{})
		if !ok {
			continue
		}

		// 提取程序ID
		var programID string
		if pid, ok := inst["programId"].(string); ok {
			programID = pid
		}

		// 提取账户
		accAddrs := make([]string, 0)
		if accounts, ok := inst["accounts"].([]interface{}); ok {
			for _, accIdx := range accounts {
				if idx, ok := accIdx.(float64); ok {
					idxInt := int(idx)
					if idxInt < len(accountKeys) {
						if accKey, ok := accountKeys[idxInt].(map[string]interface{}); ok {
							if pubkey, ok := accKey["pubkey"].(string); ok {
								accAddrs = append(accAddrs, pubkey)
							}
						}
					}
				}
			}
		}

		// 提取数据
		var dataB58 string
		if data, ok := inst["data"].(string); ok {
			dataB58 = data
		}

		// 创建简单指令
		simpleInst := map[string]interface{}{
			"program_id": programID,
			"accounts":   accAddrs,
			"data_b58":   dataB58,
			"is_inner":   false,
		}

		// 如果是指令有parsed字段，提取解析后的信息
		if parsed, ok := inst["parsed"].(map[string]interface{}); ok {
			simpleInst["parsed"] = parsed

			// 提取指令类型
			if instType, ok := parsed["type"].(string); ok {
				simpleInst["type"] = instType

				// 处理特定类型的指令
				if info, ok := parsed["info"].(map[string]interface{}); ok {
					switch instType {
					case "transfer":
						// 系统转账指令
						if source, ok := info["source"].(string); ok {
							if destination, ok := info["destination"].(string); ok {
								if lamports, ok := info["lamports"].(float64); ok {
									decodedEvents = append(decodedEvents, map[string]interface{}{
										"type":        "system_transfer",
										"from":        source,
										"to":          destination,
										"lamports":    fmt.Sprintf("%.0f", lamports),
										"program_id":  programID,
										"block_index": index,
										"is_inner":    false,
									})
								}
							}
						}
					case "write":
						// BPF写入指令
						if account, ok := info["account"].(string); ok {
							if authority, ok := info["authority"].(string); ok {
								decodedEvents = append(decodedEvents, map[string]interface{}{
									"type":        "bpf_write",
									"account":     account,
									"authority":   authority,
									"program_id":  programID,
									"block_index": index,
									"is_inner":    false,
								})
							}
						}
					}
				}
			}
		}

		simpleInstructions = append(simpleInstructions, simpleInst)

		logrus.Debugf("[sol] jsonParsed instruction[%d]: programID=%s, type=%s, accounts=%d",
			i, programID, simpleInst["type"], len(accAddrs))
	}

	// 设置结果
	if len(simpleInstructions) > 0 {
		tx["sol_instructions"] = simpleInstructions
		logrus.Debugf("[sol] extracted %d instructions from jsonParsed", len(simpleInstructions))
	}
	if len(decodedEvents) > 0 {
		tx["sol_events"] = decodedEvents
		logrus.Debugf("[sol] decoded %d events from jsonParsed", len(decodedEvents))
	}
}

// extractInstructionsFromSDK 从新SDK的交易中提取指令信息
func (s *SolanaScanner) extractInstructionsFromSDK(blockTx *client.BlockTransaction, tx map[string]interface{}, blockHeight uint64, index int) {
	simpleInstructions := make([]map[string]interface{}, 0)
	decodedEvents := make([]map[string]interface{}, 0)

	// 处理主要指令
	for i, instruction := range blockTx.Transaction.Message.Instructions {
		var programID string
		if instruction.ProgramIDIndex < len(blockTx.AccountKeys) {
			programID = blockTx.AccountKeys[instruction.ProgramIDIndex].String()
		}

		accAddrs := make([]string, 0, len(instruction.Accounts))
		for _, accIdx := range instruction.Accounts {
			if int(accIdx) < len(blockTx.AccountKeys) {
				accAddrs = append(accAddrs, blockTx.AccountKeys[accIdx].String())
			}
		}

		var dataB58 string
		if len(instruction.Data) > 0 {
			dataB58 = base58.Encode(instruction.Data)
		}

		simpleInstructions = append(simpleInstructions, map[string]interface{}{
			"program_id": programID,
			"accounts":   accAddrs,
			"data_b58":   dataB58,
			"is_inner":   false,
		})

		// 解析系统转账指令
		if programID == "11111111111111111111111111111111" && len(instruction.Data) >= 12 && len(accAddrs) >= 2 {
			// 检查是否为转账指令（tag=2）
			tag := uint32(instruction.Data[0]) | uint32(instruction.Data[1])<<8 | uint32(instruction.Data[2])<<16 | uint32(instruction.Data[3])<<24
			if tag == 2 {
				amt := uint64(instruction.Data[4]) | uint64(instruction.Data[5])<<8 | uint64(instruction.Data[6])<<16 | uint64(instruction.Data[7])<<24 |
					uint64(instruction.Data[8])<<32 | uint64(instruction.Data[9])<<40 | uint64(instruction.Data[10])<<48 | uint64(instruction.Data[11])<<56
				decodedEvents = append(decodedEvents, map[string]interface{}{
					"type":        "system_transfer",
					"from":        accAddrs[0],
					"to":          accAddrs[1],
					"lamports":    fmt.Sprintf("%d", amt),
					"program_id":  programID,
					"block_index": index,
					"is_inner":    false,
				})
			}
		}

		logrus.Debugf("[sol] blocto/solana-go-sdk instruction[%d]: programID=%s, data_len=%d, accounts=%d", i, programID, len(instruction.Data), len(accAddrs))
	}

	// 设置结果
	if len(simpleInstructions) > 0 {
		tx["sol_instructions"] = simpleInstructions
		logrus.Debugf("[sol] extracted %d instructions", len(simpleInstructions))
	}
	if len(decodedEvents) > 0 {
		tx["sol_events"] = decodedEvents
		logrus.Debugf("[sol] decoded %d events", len(decodedEvents))
	}
}

// GetBlockTransactionsFromBlock 从区块获取交易
func (s *SolanaScanner) GetBlockTransactionsFromBlock(block *models.Block) ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 使用故障转移管理器获取原始jsonParsed数据
	blockData, err := s.failoverManager.GetBlockRaw(ctx, block.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block transactions for height %d: %v", block.Height, err)
	}

	logrus.Infof("[sol] GetBlockRaw(txns,height=%d) using jsonParsed succeeded", block.Height)

	// 从jsonParsed数据中提取交易
	transactions, err := s.extractTransactionsFromJsonParsed(blockData, block.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to extract transactions from jsonParsed data: %v", err)
	}

	return transactions, nil
}

// CalculateBlockStats 计算区块统计信息
func (s *SolanaScanner) CalculateBlockStats(block *models.Block, transactions []map[string]interface{}) {
	// 计算交易数量
	block.TransactionCount = len(transactions)

	// 计算总费用
	var totalFees float64
	for _, tx := range transactions {
		if fee, ok := tx["fee"].(uint64); ok {
			totalFees += float64(fee)
		}
	}
	block.Fee = totalFees

	// 计算区块大小（估算）
	block.Size = uint64(len(transactions) * 200) // 估算每个交易200字节
}

// convertToGenericBlock 转换为通用区块格式 (已弃用，使用convertToGenericBlockSDK)
func (s *SolanaScanner) convertToGenericBlock(solanaBlock interface{}, height uint64) *models.Block {
	// 旧方法，已弃用
	return &models.Block{
		Height:    height,
		Hash:      "deprecated",
		Timestamp: time.Now(),
		Chain:     s.config.Name,
		ChainID:   s.config.ChainID,
	}
}

// convertToGenericTransaction 转换为通用交易格式 (已弃用，使用convertToGenericTransactionSDK)
func (s *SolanaScanner) convertToGenericTransaction(solanaTx interface{}, blockHeight uint64, index int) map[string]interface{} {
	// 旧方法，已弃用
	return map[string]interface{}{
		"hash":             "deprecated",
		"block_height":     blockHeight,
		"block_index":      index,
		"status":           "deprecated",
		"timestamp":        time.Now().Unix(),
		"chain_id":         s.config.ChainID,
		"chain_name":       s.config.Name,
		"transaction_type": "sol",
	}
}

// tryExtractInstructionsFromJSON 尝试从JSON模式的交易中提取instructions (已弃用)
func (s *SolanaScanner) tryExtractInstructionsFromJSON(solanaTx interface{}, tx map[string]interface{}, blockHeight uint64, index int) bool {
	// 旧方法，已弃用
	return false
}

// GetAccountInfo 获取账户信息（已弃用）
func (s *SolanaScanner) GetAccountInfo(ctx context.Context, address string) (interface{}, error) {
	// 旧方法，已弃用
	return nil, fmt.Errorf("deprecated method")
}

// GetBalance 获取账户余额（已弃用）
func (s *SolanaScanner) GetBalance(ctx context.Context, address string) (uint64, error) {
	// 旧方法，已弃用
	return 0, fmt.Errorf("deprecated method")
}

// GetTransaction 获取交易详情（已弃用）
func (s *SolanaScanner) GetTransaction(ctx context.Context, signature string) (interface{}, error) {
	// 旧方法，已弃用
	return nil, fmt.Errorf("deprecated method")
}

// parseBase64Transaction 解析Base64模式下的交易（已弃用）
func (s *SolanaScanner) parseBase64Transaction(solanaTx interface{}, blockHeight uint64, index int, tx map[string]interface{}) map[string]interface{} {
	// 旧方法，已弃用
	return tx
}

// ProcessSolanaArtifacts 处理 Solana 特定的工件（转账事件和交易明细）
func (s *SolanaScanner) ProcessSolanaArtifacts(transactions []map[string]interface{}, block interface{}) {
	api := s.getScannerAPI()
	if api == nil {
		return
	}

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

	// 事件收集
	events := make([]map[string]interface{}, 0)
	for _, tx := range transactions {
		// 仅处理成功交易
		if st, ok := tx["status"].(string); ok && st == "failed" {
			continue
		}

		// 优先使用解码出的 sol_events 作为标准事件
		if raw, ok := tx["sol_events"]; ok && raw != nil {
			switch v := raw.(type) {
			case []map[string]interface{}:
				s.buildSolTransferEvents(&events, v, tx, blockHeight)
			case []interface{}:
				arr := make([]map[string]interface{}, 0, len(v))
				for _, it := range v {
					if m, ok2 := it.(map[string]interface{}); ok2 {
						arr = append(arr, m)
					}
				}
				s.buildSolTransferEvents(&events, arr, tx, blockHeight)
			}
		} else {
			// 尝试从 meta_raw 中解析事件
			if metaRaw, ok := tx["meta_raw"]; ok && metaRaw != nil {
				if metaEvents := s.extractEventsFromMeta(metaRaw, tx, blockHeight); len(metaEvents) > 0 {
					events = append(events, metaEvents...)
				}
			}
		}

		// 回退：from/to/value 作为原生SOL近似事件
		from, _ := tx["from"].(string)
		to, _ := tx["to"].(string)
		amount := "0"
		if v, ok := tx["value"].(string); ok {
			amount = v
		}
		if from != "" && to != "" && amount != "0" {
			events = append(events, map[string]interface{}{
				"chain":            "sol",
				"tx_id":            tx["hash"],
				"height":           blockHeight,
				"block_index":      tx["block_index"],
				"event_index":      0,
				"program_id":       "11111111111111111111111111111111",
				"asset_type":       "NATIVE",
				"mint_or_contract": "SOL",
				"from_address":     from,
				"to_address":       to,
				"amount_raw":       amount,
				"decimals":         9,
				"amount_ui":        amount,
				"is_inner":         false,
				"status":           1,
			})
		}

		// 上传交易明细
		s.uploadSolTransactionDetail(tx, blockHeight, blockHash, api)
	}

	// 批量上传转账事件
	if len(events) > 0 {
		if err := api.UploadTransferEventsBatch(events); err != nil {
			logrus.Debugf("[sol] upload transfer events failed: %v", err)
		}
	}
}

// buildSolTransferEvents 将已解析的 sol_events 转为统一 TransferEvent 结构
func (s *SolanaScanner) buildSolTransferEvents(events *[]map[string]interface{}, decoded []map[string]interface{}, tx map[string]interface{}, blockHeight uint64) {
	for _, e := range decoded {
		typ, _ := e["type"].(string)
		switch typ {
		case "system_transfer":
			fromAddr, _ := e["from"].(string)
			toAddr, _ := e["to"].(string)
			lamports, _ := e["lamports"].(string)
			programID, _ := e["program_id"].(string)
			*events = append(*events, map[string]interface{}{
				"chain":            "sol",
				"tx_id":            tx["hash"],
				"height":           blockHeight,
				"block_index":      tx["block_index"],
				"event_index":      0,
				"program_id":       programID,
				"asset_type":       "NATIVE",
				"mint_or_contract": "SOL",
				"from_address":     fromAddr,
				"to_address":       toAddr,
				"amount_raw":       lamports,
				"decimals":         9,
				"amount_ui":        lamports,
				"is_inner":         false,
				"status":           1,
			})
		case "spl_transfer":
			fromAcc, _ := e["from_account"].(string)
			toAcc, _ := e["to_account"].(string)
			amount, _ := e["amount"].(string)
			programID, _ := e["program_id"].(string)
			*events = append(*events, map[string]interface{}{
				"chain":            "sol",
				"tx_id":            tx["hash"],
				"height":           blockHeight,
				"block_index":      tx["block_index"],
				"event_index":      0,
				"program_id":       programID,
				"asset_type":       "SPL",
				"mint_or_contract": "",
				"from_address":     fromAcc,
				"to_address":       toAcc,
				"amount_raw":       amount,
				"decimals":         0,
				"amount_ui":        amount,
				"is_inner":         false,
				"status":           1,
			})
		case "balance_change_inferred":
			lamports, _ := e["lamports"].(string)
			programID, _ := e["program_id"].(string)
			*events = append(*events, map[string]interface{}{
				"chain":            "sol",
				"tx_id":            tx["hash"],
				"height":           blockHeight,
				"block_index":      tx["block_index"],
				"event_index":      0,
				"program_id":       programID,
				"asset_type":       "NATIVE",
				"mint_or_contract": "SOL",
				"from_address":     "",
				"to_address":       "",
				"amount_raw":       lamports,
				"decimals":         9,
				"amount_ui":        lamports,
				"is_inner":         false,
				"status":           1,
			})
		case "fee_transaction":
			lamports, _ := e["lamports"].(string)
			programID, _ := e["program_id"].(string)
			*events = append(*events, map[string]interface{}{
				"chain":            "sol",
				"tx_id":            tx["hash"],
				"height":           blockHeight,
				"block_index":      tx["block_index"],
				"event_index":      0,
				"program_id":       programID,
				"asset_type":       "NATIVE",
				"mint_or_contract": "SOL",
				"from_address":     tx["from"],
				"to_address":       "",
				"amount_raw":       lamports,
				"decimals":         9,
				"amount_ui":        lamports,
				"is_inner":         false,
				"status":           1,
			})
		case "system_transfer_log", "spl_transfer_log", "generic_transfer_log":
			programID, _ := e["program_id"].(string)
			assetType, _ := e["asset_type"].(string)
			amount, _ := e["amount"].(string)
			if amount == "" {
				amount = "0"
			}
			*events = append(*events, map[string]interface{}{
				"chain":            "sol",
				"tx_id":            tx["hash"],
				"height":           blockHeight,
				"block_index":      tx["block_index"],
				"event_index":      0,
				"program_id":       programID,
				"asset_type":       assetType,
				"mint_or_contract": e["mint_or_contract"],
				"from_address":     "",
				"to_address":       "",
				"amount_raw":       amount,
				"decimals":         9,
				"amount_ui":        amount,
				"is_inner":         false,
				"status":           1,
			})
		}
	}
}

// extractEventsFromMeta 从 Meta 信息中提取转账事件
func (s *SolanaScanner) extractEventsFromMeta(metaRaw interface{}, tx map[string]interface{}, blockHeight uint64) []map[string]interface{} {
	events := make([]map[string]interface{}, 0)

	if metaRaw == nil {
		return events
	}

	// 简化实现，已弃用旧的解析逻辑
	logrus.Debugf("[sol] extractEventsFromMeta: deprecated method, no events extracted")

	return events
}

// convertSolEventToTransferEvent 将 Solana 事件转换为统一的转账事件格式
func (s *SolanaScanner) convertSolEventToTransferEvent(solEvent map[string]interface{}, tx map[string]interface{}, blockHeight uint64) map[string]interface{} {
	eventType, _ := solEvent["type"].(string)

	switch eventType {
	case "balance_change_inferred":
		lamports, _ := solEvent["lamports"].(string)
		programID, _ := solEvent["program_id"].(string)

		return map[string]interface{}{
			"chain":            "sol",
			"tx_id":            tx["hash"],
			"height":           blockHeight,
			"block_index":      tx["block_index"],
			"event_index":      0,
			"program_id":       programID,
			"asset_type":       "NATIVE",
			"mint_or_contract": "SOL",
			"from_address":     tx["from"],
			"to_address":       tx["to"],
			"amount_raw":       lamports,
			"decimals":         9,
			"amount_ui":        lamports,
			"is_inner":         false,
			"status":           1,
		}

	case "fee_transaction":
		lamports, _ := solEvent["lamports"].(string)
		programID, _ := solEvent["program_id"].(string)

		return map[string]interface{}{
			"chain":            "sol",
			"tx_id":            tx["hash"],
			"height":           blockHeight,
			"block_index":      tx["block_index"],
			"event_index":      0,
			"program_id":       programID,
			"asset_type":       "NATIVE",
			"mint_or_contract": "SOL",
			"from_address":     tx["from"],
			"to_address":       "",
			"amount_raw":       lamports,
			"decimals":         9,
			"amount_ui":        lamports,
			"is_inner":         false,
			"status":           1,
		}

	case "system_transfer_log", "spl_transfer_log", "generic_transfer_log":
		amount, _ := solEvent["amount"].(string)
		if amount == "" {
			amount = "0"
		}
		return map[string]interface{}{
			"chain":            "sol",
			"tx_id":            tx["hash"],
			"height":           blockHeight,
			"block_index":      tx["block_index"],
			"event_index":      0,
			"program_id":       solEvent["program_id"],
			"asset_type":       solEvent["asset_type"],
			"mint_or_contract": solEvent["mint_or_contract"],
			"from_address":     tx["from"],
			"to_address":       tx["to"],
			"amount_raw":       amount,
			"decimals":         9,
			"amount_ui":        amount,
			"is_inner":         solEvent["is_inner"],
			"status":           1,
		}
	}

	return nil
}

// uploadSolTransactionDetail 上传单笔Sol交易明细
func (s *SolanaScanner) uploadSolTransactionDetail(tx map[string]interface{}, blockHeight uint64, blockHash string, api ScannerAPIInterface) {
	detail := map[string]interface{}{
		"tx_id":            tx["hash"],
		"slot":             blockHeight,
		"blockhash":        blockHash,
		"recent_blockhash": "",
		"version":          "legacy",
		"fee":              tx["fee"],
		"compute_units":    0,
		"cost_units":       0,
		"status_json":      "{}",
	}

	// 添加可选字段
	if v, ok := tx["account_keys"]; ok {
		detail["account_keys"] = s.mustJSON(v)
	}
	if v, ok := tx["pre_balances"]; ok {
		detail["pre_balances"] = s.mustJSON(v)
	}
	if v, ok := tx["post_balances"]; ok {
		detail["post_balances"] = s.mustJSON(v)
	}
	if v, ok := tx["meta_logs"]; ok {
		detail["logs_json"] = s.mustJSON(v)
	}
	if v, ok := tx["sol_instructions"]; ok {
		detail["instructions"] = v
	}
	if v, ok := tx["sol_events"]; ok {
		detail["decoded_events"] = v
	}

	// 附带简化指令
	var insPayload []map[string]interface{}
	if arr, ok := tx["sol_instructions"].([]map[string]interface{}); ok {
		insPayload = arr
	}

	if err := api.UploadSolTxDetail(detail, insPayload); err != nil {
		logrus.Debugf("[sol] upload detail failed for %v: %v", tx["hash"], err)
	}
}

// getScannerAPI 获取扫描器API实例
func (s *SolanaScanner) getScannerAPI() ScannerAPIInterface {
	api := config.GetScannerAPI()
	if api == nil {
		return nil
	}
	return &ScannerAPIAdapter{api: api}
}

// mustJSON 将值转为字符串
func (s *SolanaScanner) mustJSON(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

// ScannerAPIInterface 定义扫描器API接口，用于解耦
type ScannerAPIInterface interface {
	UploadTransferEventsBatch(events []map[string]interface{}) error
	UploadSolTxDetail(detail map[string]interface{}, instructions []map[string]interface{}) error
}

// ScannerAPIAdapter 实现接口适配器
type ScannerAPIAdapter struct {
	api *pkg.ScannerAPI
}

func (a *ScannerAPIAdapter) UploadTransferEventsBatch(events []map[string]interface{}) error {
	return a.api.UploadTransferEventsBatch(events)
}

func (a *ScannerAPIAdapter) UploadSolTxDetail(detail map[string]interface{}, instructions []map[string]interface{}) error {
	return a.api.UploadSolTxDetail(detail, instructions)
}
