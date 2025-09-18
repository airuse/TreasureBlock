package scanners

import (
	"context"
	"fmt"
	"time"

	"blockChainBrowser/client/scanner/config"
	"blockChainBrowser/client/scanner/internal/failover"
	"blockChainBrowser/client/scanner/internal/models"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
)

// SolanaScanner 使用官方 solana-go SDK 的 Solana 扫块器
type SolanaScanner struct {
	config          *config.ChainConfig
	failoverManager *failover.SOLFailoverManager
}

// NewSolanaScanner 创建新的 Solana 扫块器（使用官方 SDK）
func NewSolanaScanner(cfg *config.ChainConfig) *SolanaScanner {
	// 初始化主 RPC 客户端
	var localClient *rpc.Client
	if cfg.RPCURL != "" {
		localClient = rpc.New(cfg.RPCURL)
		logrus.Infof("Initialized Solana main RPC client: %s", cfg.RPCURL)
	}

	// 初始化多个外部API客户端
	externalClients := make([]*rpc.Client, 0)
	if len(cfg.ExplorerAPIURLs) > 0 {
		for _, apiURL := range cfg.ExplorerAPIURLs {
			if client := rpc.New(apiURL); client != nil {
				externalClients = append(externalClients, client)
				logrus.Infof("Initialized Solana external RPC client: %s", apiURL)
			} else {
				logrus.Warnf("Failed to initialize Solana RPC client: %s", apiURL)
			}
		}
	}

	// 创建故障转移管理器
	failoverManager := failover.NewSOLFailoverManager(localClient, externalClients)

	scanner := &SolanaScanner{
		config:          cfg,
		failoverManager: failoverManager,
	}

	logrus.Infof("Initialized Solana scanner with failover manager (%d external clients)", len(externalClients))
	return scanner
}

// GetLatestBlockHeight 获取最新区块高度
func (s *SolanaScanner) GetLatestBlockHeight() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slot, err := s.failoverManager.CallWithFailoverSlot("GetLatestBlockHeight", func(client *rpc.Client) (uint64, error) {
		slot, err := client.GetSlot(ctx, rpc.CommitmentFinalized)
		if err != nil {
			return 0, err
		}
		return uint64(slot), nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get latest slot: %w", err)
	}

	return uint64(slot), nil
}

// GetBlockByHeight 根据高度获取区块
func (s *SolanaScanner) GetBlockByHeight(height uint64) (*models.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	block, err := s.failoverManager.CallWithFailoverBlock("GetBlockByHeight", func(client *rpc.Client) (*rpc.GetBlockResult, error) {
		return client.GetBlockWithOpts(
			ctx,
			height,
			&rpc.GetBlockOpts{
				Encoding:                       solana.EncodingBase64,
				TransactionDetails:             rpc.TransactionDetailsFull,
				Rewards:                        &[]bool{true}[0],
				MaxSupportedTransactionVersion: &[]uint64{0}[0],
			},
		)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	return s.convertToGenericBlock(block, height), nil
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

// GetBlockTransactionsFromBlock 从区块获取交易
func (s *SolanaScanner) GetBlockTransactionsFromBlock(block *models.Block) ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 重新获取区块以获取完整交易数据
	solanaBlock, err := s.failoverManager.CallWithFailoverBlock("GetBlockTransactionsFromBlock", func(client *rpc.Client) (*rpc.GetBlockResult, error) {
		return client.GetBlockWithOpts(
			ctx,
			block.Height,
			&rpc.GetBlockOpts{
				Encoding:                       solana.EncodingBase64,
				TransactionDetails:             rpc.TransactionDetailsFull,
				Rewards:                        &[]bool{true}[0],
				MaxSupportedTransactionVersion: &[]uint64{0}[0],
			},
		)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get solana block: %w", err)
	}

	// 转换交易格式
	transactions := make([]map[string]interface{}, 0, len(solanaBlock.Transactions))
	for i, tx := range solanaBlock.Transactions {
		genericTx := s.convertToGenericTransaction(&tx, block.Height, i)
		transactions = append(transactions, genericTx)
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

// convertToGenericBlock 转换为通用区块格式
func (s *SolanaScanner) convertToGenericBlock(solanaBlock *rpc.GetBlockResult, height uint64) *models.Block {
	block := &models.Block{
		Height:           height,
		Hash:             solanaBlock.Blockhash.String(),
		PreviousHash:     solanaBlock.PreviousBlockhash.String(),
		Timestamp:        time.Now(), // 使用当前时间作为默认值
		TransactionCount: len(solanaBlock.Transactions),
		Size:             uint64(len(solanaBlock.Transactions) * 200), // 估算大小
		Miner:            "",                                          // Solana没有矿工概念
		Difficulty:       0,                                           // Solana不使用工作量证明
		Nonce:            0,
		Chain:            s.config.Name,
		ChainID:          s.config.ChainID,
	}

	// 设置时间戳
	if solanaBlock.BlockTime != nil {
		block.Timestamp = time.Unix(int64(*solanaBlock.BlockTime), 0)
	}

	// 计算总费用
	var totalFees float64
	for _, tx := range solanaBlock.Transactions {
		if tx.Meta != nil {
			totalFees += float64(tx.Meta.Fee)
		}
	}
	block.Fee = totalFees

	// 设置矿工（使用第一个奖励接收者）
	if len(solanaBlock.Rewards) > 0 {
		block.Miner = solanaBlock.Rewards[0].Pubkey.String()
	}

	return block
}

// convertToGenericTransaction 转换为通用交易格式
func (s *SolanaScanner) convertToGenericTransaction(solanaTx *rpc.TransactionWithMeta, blockHeight uint64, index int) map[string]interface{} {
	tx := map[string]interface{}{
		"hash":         "",
		"block_height": blockHeight,
		"block_hash":   "mock_block_hash", // 需要从上下文获取
		"from":         "",
		"to":           "",
		"value":        "0",
		"fee":          uint64(0),
		// 提供双写键，兼容上层取值（camelCase 与 snake_case）
		"gasUsed":          0,
		"gasPrice":         "0",
		"gasLimit":         0,
		"nonce":            0,
		"status":           "success",
		"timestamp":        time.Now().Unix(),
		"chain_id":         s.config.ChainID,
		"chain_name":       s.config.Name,
		"transaction_type": "sol_transfer",
		"raw_data":         solanaTx,
	}

	// 解析交易数据
	parsedTx, err := solanaTx.GetParsedTransaction()
	if err != nil {
		// 如果解析失败，使用基本信息
		tx["hash"] = "unknown"
		return tx
	}

	// 设置交易哈希
	if len(parsedTx.Signatures) > 0 {
		tx["hash"] = parsedTx.Signatures[0].String()
	}

	// 设置费用
	if solanaTx.Meta != nil {
		tx["fee"] = uint64(solanaTx.Meta.Fee)
	}

	// 设置发送方和接收方
	if len(parsedTx.Message.AccountKeys) >= 2 {
		tx["from"] = parsedTx.Message.AccountKeys[0].String()
		tx["to"] = parsedTx.Message.AccountKeys[1].String()
	}

	// 计算转账金额
	if solanaTx.Meta != nil && len(solanaTx.Meta.PreBalances) >= 2 && len(solanaTx.Meta.PostBalances) >= 2 {
		preBalance := solanaTx.Meta.PreBalances[0]
		postBalance := solanaTx.Meta.PostBalances[0]
		if preBalance > postBalance {
			amount := preBalance - postBalance - uint64(solanaTx.Meta.Fee)
			tx["value"] = fmt.Sprintf("%d", amount)
		}
	}

	// 设置交易状态
	if solanaTx.Meta != nil && solanaTx.Meta.Err != nil {
		tx["status"] = "failed"
	}

	return tx
}

// GetAccountInfo 获取账户信息（额外功能）
func (s *SolanaScanner) GetAccountInfo(ctx context.Context, address string) (*rpc.GetAccountInfoResult, error) {
	pubkey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	_, err = s.failoverManager.CallWithFailoverMap("GetAccountInfo", func(client *rpc.Client) (map[string]interface{}, error) {
		_, err := client.GetAccountInfo(ctx, pubkey)
		if err != nil {
			return nil, err
		}
		// 这里需要将 result 转换为 map，但为了简化，我们直接返回错误
		return nil, fmt.Errorf("not implemented")
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}

	// 临时返回 nil，实际使用时需要实现转换
	return nil, fmt.Errorf("not implemented")
}

// GetBalance 获取账户余额（额外功能）
func (s *SolanaScanner) GetBalance(ctx context.Context, address string) (uint64, error) {
	pubkey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return 0, fmt.Errorf("invalid public key: %w", err)
	}

	balance, err := s.failoverManager.CallWithFailoverUint64("GetBalance", func(client *rpc.Client) (uint64, error) {
		result, err := client.GetBalance(ctx, pubkey, rpc.CommitmentFinalized)
		if err != nil {
			return 0, err
		}
		return result.Value, nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

// GetTransaction 获取交易详情（额外功能）
func (s *SolanaScanner) GetTransaction(ctx context.Context, signature string) (*rpc.GetTransactionResult, error) {
	sig, err := solana.SignatureFromBase58(signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	_, err = s.failoverManager.CallWithFailoverMap("GetTransaction", func(client *rpc.Client) (map[string]interface{}, error) {
		_, err := client.GetTransaction(
			ctx,
			sig,
			&rpc.GetTransactionOpts{
				Encoding: solana.EncodingBase64,
			},
		)
		if err != nil {
			return nil, err
		}
		// 这里需要将 result 转换为 map，但为了简化，我们直接返回错误
		return nil, fmt.Errorf("not implemented")
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// 临时返回 nil，实际使用时需要实现转换
	return nil, fmt.Errorf("not implemented")
}
