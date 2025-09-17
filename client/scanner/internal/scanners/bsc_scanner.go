package scanners

import (
	"blockChainBrowser/client/scanner/config"
	"blockChainBrowser/client/scanner/internal/models"
)

// BSCScanner BSC扫块器 - 继承EVM扫块器
type BSCScanner struct {
	*EVMScanner
}

// NewBSCScanner 创建新的BSC扫块器
func NewBSCScanner(cfg *config.ChainConfig) *BSCScanner {
	evmScanner := NewEVMSanner(cfg, "bsc")
	return &BSCScanner{
		EVMScanner: evmScanner,
	}
}

// GetLatestBlockHeight 获取最新区块高度
func (bs *BSCScanner) GetLatestBlockHeight() (uint64, error) {
	return bs.EVMScanner.GetLatestBlockHeight()
}

// GetBlockByHeight 根据高度获取区块
func (bs *BSCScanner) GetBlockByHeight(height uint64) (*models.Block, error) {
	return bs.EVMScanner.GetBlockByHeight(height)
}

// ValidateBlock 验证区块
func (bs *BSCScanner) ValidateBlock(block *models.Block) error {
	return bs.EVMScanner.ValidateBlock(block)
}

// GetBlockTransactionsFromBlock 从区块获取交易信息
func (bs *BSCScanner) GetBlockTransactionsFromBlock(block *models.Block) ([]map[string]interface{}, error) {
	return bs.EVMScanner.GetBlockTransactionsFromBlock(block)
}

// CalculateBlockStats 计算区块统计信息
func (bs *BSCScanner) CalculateBlockStats(block *models.Block, transactions []map[string]interface{}) {
	// BSC使用与ETH相同的统计计算逻辑
	bs.EVMScanner.CalculateBlockStats(block, transactions)
}
