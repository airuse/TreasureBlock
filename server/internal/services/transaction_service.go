package services

import (
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"context"
	"fmt"
	"strings"
)

type TransactionService interface {
	GetTransactionByHash(ctx context.Context, hash string) (*models.Transaction, error)
	GetTransactionsByAddress(ctx context.Context, address string, page, pageSize int) ([]*models.Transaction, int64, error)
	GetTransactionsByBlockHash(ctx context.Context, blockHash string) ([]*models.Transaction, error)
	GetTransactionsByBlockHeight(ctx context.Context, blockHeight uint64, page, pageSize int, chain string) ([]*models.Transaction, int64, error)
	ListTransactions(ctx context.Context, page, pageSize int, chain string) ([]*models.Transaction, int64, error)
	GetLatestTransactions(ctx context.Context, chain string, limit int) ([]*models.Transaction, error)
	CreateTransaction(ctx context.Context, tx *models.Transaction) error
	UpdateTransaction(ctx context.Context, tx *models.Transaction) error
	DeleteTransaction(ctx context.Context, hash string) error
}

type transactionService struct {
	txRepo         repository.TransactionRepository
	coinConfigRepo repository.CoinConfigRepository
}

func NewTransactionService(txRepo repository.TransactionRepository, coinConfigRepo repository.CoinConfigRepository) TransactionService {
	return &transactionService{
		txRepo:         txRepo,
		coinConfigRepo: coinConfigRepo,
	}
}

// GetTransactionByHash 根据哈希获取交易
func (s *transactionService) GetTransactionByHash(ctx context.Context, hash string) (*models.Transaction, error) {
	if hash == "" {
		return nil, fmt.Errorf("transaction hash cannot be empty")
	}

	tx, err := s.txRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by hash: %w", err)
	}

	return tx, nil
}

// GetTransactionsByAddress 根据地址获取交易列表
func (s *transactionService) GetTransactionsByAddress(ctx context.Context, address string, page, pageSize int) ([]*models.Transaction, int64, error) {
	if address == "" {
		return nil, 0, fmt.Errorf("address cannot be empty")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	txs, total, err := s.txRepo.GetByAddress(ctx, address, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions by address: %w", err)
	}

	return txs, total, nil
}

// GetTransactionsByBlockHash 根据区块哈希获取交易列表
func (s *transactionService) GetTransactionsByBlockHash(ctx context.Context, blockHash string) ([]*models.Transaction, error) {
	if blockHash == "" {
		return nil, fmt.Errorf("block hash cannot be empty")
	}

	txs, err := s.txRepo.GetByBlockHash(ctx, blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by block hash: %w", err)
	}

	return txs, nil
}

// GetTransactionsByBlockHeight 根据区块高度获取交易列表
func (s *transactionService) GetTransactionsByBlockHeight(ctx context.Context, blockHeight uint64, page, pageSize int, chain string) ([]*models.Transaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	txs, total, err := s.txRepo.GetByBlockHeight(ctx, blockHeight, offset, pageSize, chain)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions by block height: %w", err)
	}

	// 获取代币地址列表，用于判断交易是否为代币交易
	tokenMap, err := s.getTokenAddresses(ctx, chain)
	if err != nil {
		// 如果获取代币地址失败，记录错误但不影响交易列表返回
		// fmt.Printf("Warning: Failed to get token addresses for chain %s: %v\n", chain, err)
	}

	// 为每个交易添加代币标识
	for _, tx := range txs {
		tx.IsToken = s.isTokenTransaction(tx, tokenMap)
	}

	return txs, total, nil
}

// ListTransactions 分页查询交易列表
func (s *transactionService) ListTransactions(ctx context.Context, page, pageSize int, chain string) ([]*models.Transaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	txs, total, err := s.txRepo.List(ctx, offset, pageSize, chain)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	// 获取代币地址列表，用于判断交易是否为代币交易
	tokenMap, err := s.getTokenAddresses(ctx, chain)
	if err != nil {
		// 如果获取代币地址失败，记录错误但不影响交易列表返回
		// fmt.Printf("Warning: Failed to get token addresses for chain %s: %v\n", chain, err)
	}

	// 为每个交易添加代币标识
	for _, tx := range txs {
		tx.IsToken = s.isTokenTransaction(tx, tokenMap)
	}

	return txs, total, nil
}

// GetLatestTransactions 获取最新区块的前几条交易
func (s *transactionService) GetLatestTransactions(ctx context.Context, chain string, limit int) ([]*models.Transaction, error) {
	if limit <= 0 {
		limit = 5 // 默认返回5条
	}
	if limit > 20 {
		limit = 20 // 最大限制20条
	}

	// 获取最新区块高度
	latestBlock, err := s.txRepo.GetLatestBlockHeight(ctx, chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block height: %w", err)
	}

	// 获取最新区块的前几条交易
	txs, err := s.txRepo.GetLatestTransactionsByBlockIndex(ctx, chain, latestBlock, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest transactions: %w", err)
	}

	// 为每个交易添加代币标识
	tokenMap, err := s.getTokenAddresses(ctx, chain)
	if err != nil {
		// 如果获取代币配置失败，继续执行，只是不设置代币信息
		tokenMap = make(map[string]*models.CoinConfig)
	}

	for _, tx := range txs {
		tx.IsToken = s.isTokenTransaction(tx, tokenMap)
	}

	return txs, nil
}

// CreateTransaction 创建交易
func (s *transactionService) CreateTransaction(ctx context.Context, tx *models.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	if tx.TxID == "" {
		return fmt.Errorf("transaction hash cannot be empty")
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// UpdateTransaction 更新交易
func (s *transactionService) UpdateTransaction(ctx context.Context, tx *models.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	if err := s.txRepo.Update(ctx, tx); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

// DeleteTransaction 删除交易
func (s *transactionService) DeleteTransaction(ctx context.Context, hash string) error {
	if hash == "" {
		return fmt.Errorf("transaction hash cannot be empty")
	}

	if err := s.txRepo.Delete(ctx, hash); err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	return nil
}

// getTokenAddresses 获取指定链的代币配置信息
func (s *transactionService) getTokenAddresses(ctx context.Context, chain string) (map[string]*models.CoinConfig, error) {
	coinConfigs, err := s.coinConfigRepo.GetByChain(ctx, chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get coin configs for chain %s: %w", chain, err)
	}

	// 返回地址到代币配置的映射
	tokenMap := make(map[string]*models.CoinConfig)
	for _, config := range coinConfigs {
		if config.Status == 1 && config.ContractAddr != "" { // 只获取启用的代币
			address := strings.ToLower(config.ContractAddr)
			tokenMap[address] = config // 保存完整的代币配置
		}
	}

	return tokenMap, nil
}

// isTokenTransaction 判断交易是否为代币交易，并设置代币名称和合约地址
func (s *transactionService) isTokenTransaction(tx *models.Transaction, tokenMap map[string]*models.CoinConfig) bool {
	if len(tokenMap) == 0 {
		return false
	}

	// 检查交易的to地址是否在代币地址列表中
	txToAddress := strings.ToLower(tx.AddressTo)
	if config, exists := tokenMap[txToAddress]; exists {
		tx.IsToken = true
		tx.TokenName = config.Name                // 设置代币名称
		tx.TokenSymbol = config.Symbol            // 设置代币符号
		tx.TokenDecimals = uint8(config.Decimals) // 设置代币精度
		tx.TokenDescription = config.Description  // 设置代币描述
		tx.TokenWebsite = config.WebsiteURL       // 设置代币官网
		tx.TokenExplorer = config.ExplorerURL     // 设置代币浏览器链接
		tx.TokenLogo = config.LogoURL             // 设置代币Logo
		tx.TokenMarketCapRank = func() *int {
			rank := int(config.MarketCapRank)
			return &rank
		}() // 设置市值排名
		tx.TokenIsStablecoin = config.IsStablecoin // 设置是否为稳定币
		tx.TokenIsVerified = config.IsVerified     // 设置是否已验证
		tx.ContractAddr = txToAddress              // 设置合约地址
		return true
	}

	return false
}
