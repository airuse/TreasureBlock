package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"blockChainBrowser/server/config"
	"blockChainBrowser/server/internal/repository"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// DataCleanupConfig 数据清理配置
type DataCleanupConfig struct {
	Chain            string `json:"chain"`             // 链名称 (eth, btc)
	MaxBlocks        int64  `json:"max_blocks"`        // 最大保留区块数
	CleanupThreshold int64  `json:"cleanup_threshold"` // 清理阈值（当超过此数量时开始清理）
	BatchSize        int    `json:"batch_size"`        // 批量删除大小
	Interval         int    `json:"interval"`          // 清理间隔（分钟）
}

// DataCleanupScheduler 数据清理调度器
type DataCleanupScheduler struct {
	db                *gorm.DB
	userAddressRepo   repository.UserAddressRepository
	blockRepo         repository.BlockRepository
	txRepo            repository.TransactionRepository
	receiptRepo       repository.TransactionReceiptRepository
	contractParseRepo repository.ContractParseResultRepository
	logger            *logrus.Logger
	configs           map[string]*DataCleanupConfig
	stopChan          chan struct{}
	wg                sync.WaitGroup
}

// NewDataCleanupScheduler 创建数据清理调度器
func NewDataCleanupScheduler(
	db *gorm.DB,
	userAddressRepo repository.UserAddressRepository,
	blockRepo repository.BlockRepository,
	txRepo repository.TransactionRepository,
	receiptRepo repository.TransactionReceiptRepository,
	contractParseRepo repository.ContractParseResultRepository,
) *DataCleanupScheduler {
	return &DataCleanupScheduler{
		db:                db,
		userAddressRepo:   userAddressRepo,
		blockRepo:         blockRepo,
		txRepo:            txRepo,
		receiptRepo:       receiptRepo,
		contractParseRepo: contractParseRepo,
		logger:            logrus.New(),
		configs:           make(map[string]*DataCleanupConfig),
		stopChan:          make(chan struct{}),
	}
}

// SetConfig 设置清理配置
func (s *DataCleanupScheduler) SetConfig(chain string, config *DataCleanupConfig) {
	s.configs[chain] = config
	s.logger.Infof("设置 %s 链清理配置: 最大保留=%d, 清理阈值=%d, 批量大小=%d, 间隔=%d分钟",
		chain, config.MaxBlocks, config.CleanupThreshold, config.BatchSize, config.Interval)
}

// Start 启动调度器
func (s *DataCleanupScheduler) Start(ctx context.Context) {
	s.logger.Info("启动数据清理调度器")

	// 设置默认配置
	s.setDefaultConfigs()

	// 为每个链启动独立的清理协程
	for chain, config := range s.configs {
		s.wg.Add(1)
		go s.runCleanupForChain(ctx, chain, config)
	}
}

// Stop 停止调度器
func (s *DataCleanupScheduler) Stop() {
	s.logger.Info("停止数据清理调度器")
	close(s.stopChan)
	s.wg.Wait()
}

// setDefaultConfigs 设置默认配置
func (s *DataCleanupScheduler) setDefaultConfigs() {
	// 从配置文件读取清理配置，如果没有配置则使用默认值
	ethConfig := &DataCleanupConfig{
		Chain:            "eth",
		MaxBlocks:        50000,
		CleanupThreshold: 50000,
		BatchSize:        1000,
		Interval:         60, // 1小时
	}

	btcConfig := &DataCleanupConfig{
		Chain:            "btc",
		MaxBlocks:        5000,
		CleanupThreshold: 5000,
		BatchSize:        500,
		Interval:         120, // 2小时
	}

	// 如果配置文件中有清理配置，则使用配置文件的值
	if config.AppConfig.DataCleanup.ETH != nil {
		ethConfig.MaxBlocks = config.AppConfig.DataCleanup.ETH.MaxBlocks
		ethConfig.CleanupThreshold = config.AppConfig.DataCleanup.ETH.CleanupThreshold
		ethConfig.BatchSize = config.AppConfig.DataCleanup.ETH.BatchSize
		ethConfig.Interval = config.AppConfig.DataCleanup.ETH.Interval
	}
	if config.AppConfig.DataCleanup.BTC != nil {
		btcConfig.MaxBlocks = config.AppConfig.DataCleanup.BTC.MaxBlocks
		btcConfig.CleanupThreshold = config.AppConfig.DataCleanup.BTC.CleanupThreshold
		btcConfig.BatchSize = config.AppConfig.DataCleanup.BTC.BatchSize
		btcConfig.Interval = config.AppConfig.DataCleanup.BTC.Interval
	}

	s.SetConfig("eth", ethConfig)
	s.SetConfig("btc", btcConfig)
}

// runCleanupForChain 为指定链运行清理任务
func (s *DataCleanupScheduler) runCleanupForChain(ctx context.Context, chain string, config *DataCleanupConfig) {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Duration(config.Interval) * time.Minute)
	defer ticker.Stop()

	s.logger.Infof("启动 %s 链数据清理任务，间隔: %d分钟", chain, config.Interval)

	for {
		select {
		case <-ctx.Done():
			s.logger.Infof("%s 链清理任务因上下文取消而停止", chain)
			return
		case <-s.stopChan:
			s.logger.Infof("%s 链清理任务因停止信号而停止", chain)
			return
		case <-ticker.C:
			if err := s.CleanupChainData(ctx, chain, config); err != nil {
				s.logger.Errorf("%s 链数据清理失败: %v", chain, err)
			}
		}
	}
}

// CleanupChainData 清理指定链的数据（公开方法）
func (s *DataCleanupScheduler) CleanupChainData(ctx context.Context, chain string, config *DataCleanupConfig) error {
	s.logger.Infof("开始清理 %s 链数据", chain)

	// 1. 检查是否需要清理
	shouldCleanup, err := s.shouldCleanup(ctx, chain, config)
	if err != nil {
		return fmt.Errorf("检查清理条件失败: %w", err)
	}

	if !shouldCleanup {
		s.logger.Infof("%s 链数据量未达到清理阈值，跳过清理", chain)
		return nil
	}

	// 2. 获取清理基准高度
	cleanupHeight, err := s.getCleanupHeight(ctx, chain, config)
	if err != nil {
		return fmt.Errorf("获取清理基准高度失败: %w", err)
	}

	s.logger.Infof("%s 链清理基准高度: %d", chain, cleanupHeight)

	// 3. 获取受保护的地址列表
	protectedAddresses, err := s.getProtectedAddresses(ctx)
	if err != nil {
		return fmt.Errorf("获取受保护地址失败: %w", err)
	}

	s.logger.Infof("受保护地址数量: %d", len(protectedAddresses))

	// 4. 执行清理
	if err := s.executeCleanup(ctx, chain, cleanupHeight, protectedAddresses, config); err != nil {
		return fmt.Errorf("执行清理失败: %w", err)
	}

	s.logger.Infof("%s 链数据清理完成", chain)
	return nil
}

// shouldCleanup 检查是否需要清理
func (s *DataCleanupScheduler) shouldCleanup(ctx context.Context, chain string, config *DataCleanupConfig) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Table("blocks").
		Where("chain = ?", chain).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > config.CleanupThreshold, nil
}

// getCleanupHeight 获取清理基准高度
func (s *DataCleanupScheduler) getCleanupHeight(ctx context.Context, chain string, config *DataCleanupConfig) (uint64, error) {
	var blocks []struct {
		Height uint64 `gorm:"column:height"`
	}

	// 按高度降序排列，取第 MaxBlocks 个区块的高度
	err := s.db.WithContext(ctx).
		Table("blocks").
		Select("height").
		Where("chain = ?", chain).
		Order("height DESC").
		Limit(int(config.MaxBlocks)).
		Offset(int(config.MaxBlocks) - 1).
		Scan(&blocks).Error

	if err != nil {
		return 0, err
	}

	if len(blocks) == 0 {
		return 0, fmt.Errorf("未找到足够的区块数据")
	}

	return blocks[0].Height, nil
}

// getProtectedAddresses 获取受保护的地址列表
func (s *DataCleanupScheduler) getProtectedAddresses(ctx context.Context) ([]string, error) {
	var addresses []string
	err := s.db.WithContext(ctx).
		Table("user_addresses").
		Select("address").
		Pluck("address", &addresses).Error

	return addresses, err
}

// executeCleanup 执行数据清理
func (s *DataCleanupScheduler) executeCleanup(ctx context.Context, chain string, cleanupHeight uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	// 开始事务
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 清理 transactions 表
	if err := s.cleanupTransactions(tx, chain, cleanupHeight, protectedAddresses, config); err != nil {
		tx.Rollback()
		return fmt.Errorf("清理 transactions 失败: %w", err)
	}

	// 2. 清理 transaction_receipts 表
	if err := s.cleanupTransactionReceipts(tx, chain, cleanupHeight, protectedAddresses, config); err != nil {
		tx.Rollback()
		return fmt.Errorf("清理 transaction_receipts 失败: %w", err)
	}

	// 3. 清理 contract_parse_results 表
	if err := s.cleanupContractParseResults(tx, chain, cleanupHeight, protectedAddresses, config); err != nil {
		tx.Rollback()
		return fmt.Errorf("清理 contract_parse_results 失败: %w", err)
	}

	// 4. 清理 blocks 表
	if err := s.cleanupBlocks(tx, chain, cleanupHeight, config); err != nil {
		tx.Rollback()
		return fmt.Errorf("清理 blocks 失败: %w", err)
	}

	// 提交事务
	return tx.Commit().Error
}

// cleanupTransactions 清理 transactions 表
func (s *DataCleanupScheduler) cleanupTransactions(tx *gorm.DB, chain string, cleanupHeight uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	if len(protectedAddresses) == 0 {
		// 没有受保护地址，直接删除
		return s.batchDelete(tx, "transactions", "chain = ? AND height < ?", []interface{}{chain, cleanupHeight}, config.BatchSize)
	}

	// 有受保护地址，需要排除相关交易
	query := `
		DELETE FROM transactions 
		WHERE chain = ? AND height < ? 
		AND id NOT IN (
			SELECT DISTINCT t.id FROM transactions t
			WHERE t.chain = ? AND t.height < ?
			AND (t.from_address IN ? OR t.to_address IN ?)
		)
	`

	return s.batchDeleteWithQuery(tx, query, []interface{}{chain, cleanupHeight, chain, cleanupHeight, protectedAddresses, protectedAddresses}, config.BatchSize)
}

// cleanupTransactionReceipts 清理 transaction_receipts 表
func (s *DataCleanupScheduler) cleanupTransactionReceipts(tx *gorm.DB, chain string, cleanupHeight uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	if len(protectedAddresses) == 0 {
		// 没有受保护地址，直接删除
		return s.batchDelete(tx, "transaction_receipts", "chain = ? AND block_number < ?", []interface{}{chain, cleanupHeight}, config.BatchSize)
	}

	// 有受保护地址，需要排除相关收据
	query := `
		DELETE FROM transaction_receipts 
		WHERE chain = ? AND block_number < ?
		AND tx_hash NOT IN (
			SELECT DISTINCT tr.tx_hash FROM transaction_receipts tr
			JOIN transactions t ON tr.tx_hash = t.tx_id
			WHERE tr.chain = ? AND tr.block_number < ?
			AND (t.from_address IN ? OR t.to_address IN ?)
		)
	`

	return s.batchDeleteWithQuery(tx, query, []interface{}{chain, cleanupHeight, chain, cleanupHeight, protectedAddresses, protectedAddresses}, config.BatchSize)
}

// cleanupContractParseResults 清理 contract_parse_results 表
func (s *DataCleanupScheduler) cleanupContractParseResults(tx *gorm.DB, chain string, cleanupHeight uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	if len(protectedAddresses) == 0 {
		// 没有受保护地址，直接删除
		return s.batchDelete(tx, "contract_parse_results", "chain = ? AND block_number < ?", []interface{}{chain, cleanupHeight}, config.BatchSize)
	}

	// 有受保护地址，需要排除相关解析结果
	query := `
		DELETE FROM contract_parse_results 
		WHERE chain = ? AND block_number < ?
		AND tx_hash NOT IN (
			SELECT DISTINCT cpr.tx_hash FROM contract_parse_results cpr
			JOIN transactions t ON cpr.tx_hash = t.tx_id
			WHERE cpr.chain = ? AND cpr.block_number < ?
			AND (t.from_address IN ? OR t.to_address IN ?)
		)
	`

	return s.batchDeleteWithQuery(tx, query, []interface{}{chain, cleanupHeight, chain, cleanupHeight, protectedAddresses, protectedAddresses}, config.BatchSize)
}

// cleanupBlocks 清理 blocks 表
func (s *DataCleanupScheduler) cleanupBlocks(tx *gorm.DB, chain string, cleanupHeight uint64, config *DataCleanupConfig) error {
	// blocks 表不需要考虑受保护地址，直接删除
	return s.batchDelete(tx, "blocks", "chain = ? AND height < ?", []interface{}{chain, cleanupHeight}, config.BatchSize)
}

// batchDelete 批量删除数据
func (s *DataCleanupScheduler) batchDelete(tx *gorm.DB, table, whereClause string, args []interface{}, batchSize int) error {
	for {
		result := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s LIMIT %d", table, whereClause, batchSize), args...)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			break
		}

		s.logger.Debugf("删除了 %d 条 %s 记录", result.RowsAffected, table)

		// 短暂延迟，避免长时间锁表
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// batchDeleteWithQuery 使用自定义查询批量删除数据
func (s *DataCleanupScheduler) batchDeleteWithQuery(tx *gorm.DB, query string, args []interface{}, batchSize int) error {
	// 添加 LIMIT 子句
	queryWithLimit := query + fmt.Sprintf(" LIMIT %d", batchSize)

	for {
		result := tx.Exec(queryWithLimit, args...)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			break
		}

		s.logger.Debugf("删除了 %d 条记录", result.RowsAffected)

		// 短暂延迟，避免长时间锁表
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}
