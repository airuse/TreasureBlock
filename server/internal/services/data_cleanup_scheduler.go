package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"blockChainBrowser/server/config"
	"blockChainBrowser/server/internal/models"
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
		MaxBlocks:        20000,
		CleanupThreshold: 50000,
		BatchSize:        10000,
		Interval:         60,
	}

	btcConfig := &DataCleanupConfig{
		Chain:            "btc",
		MaxBlocks:        20000,
		CleanupThreshold: 50000,
		BatchSize:        10000,
		Interval:         120,
	}

	bscConfig := &DataCleanupConfig{
		Chain:            "bsc",
		MaxBlocks:        20000,
		CleanupThreshold: 50000,
		BatchSize:        10000,
		Interval:         120,
	}
	solConfig := &DataCleanupConfig{
		Chain:            "sol",
		MaxBlocks:        20000,
		CleanupThreshold: 50000,
		BatchSize:        10000,
		Interval:         120,
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
	if config.AppConfig.DataCleanup.BSC != nil {
		bscConfig.MaxBlocks = config.AppConfig.DataCleanup.BSC.MaxBlocks
		bscConfig.CleanupThreshold = config.AppConfig.DataCleanup.BSC.CleanupThreshold
		bscConfig.BatchSize = config.AppConfig.DataCleanup.BSC.BatchSize
		bscConfig.Interval = config.AppConfig.DataCleanup.BSC.Interval
	}
	if config.AppConfig.DataCleanup.SOL != nil {
		solConfig.MaxBlocks = config.AppConfig.DataCleanup.SOL.MaxBlocks
		solConfig.CleanupThreshold = config.AppConfig.DataCleanup.SOL.CleanupThreshold
		solConfig.BatchSize = config.AppConfig.DataCleanup.SOL.BatchSize
		solConfig.Interval = config.AppConfig.DataCleanup.SOL.Interval
	}
	s.SetConfig("eth", ethConfig)
	s.SetConfig("btc", btcConfig)
	s.SetConfig("bsc", bscConfig)
	s.SetConfig("sol", solConfig)
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
	startTime := time.Now()
	// s.logger.Infof("开始清理 %s 链数据", chain)

	// 1. 检查是否需要清理
	shouldCleanup, err := s.shouldCleanup(ctx, chain, config)
	if err != nil {
		return fmt.Errorf("检查清理条件失败: %w", err)
	}

	if !shouldCleanup {
		// s.logger.Infof("%s 链数据量未达到清理阈值，跳过清理", chain)
		return nil
	}

	// 2. 获取清理基准高度
	cleanupHeight, err := s.getCleanupHeight(ctx, chain, config)
	if err != nil {
		return fmt.Errorf("获取清理基准高度失败: %w", err)
	}

	// s.logger.Infof("%s 链清理基准高度: %d", chain, cleanupHeight)

	// 3. 获取受保护的地址列表
	protectedAddresses, err := s.getProtectedAddresses(ctx)
	if err != nil {
		return fmt.Errorf("获取受保护地址失败: %w", err)
	}

	// s.logger.Infof("受保护地址数量: %d", len(protectedAddresses))

	// 4. 根据链类型执行不同的清理逻辑
	if chain == "sol" {
		// SOL 链使用专门的清理逻辑
		if err := s.executeSolCleanup(ctx, chain, cleanupHeight, protectedAddresses, config); err != nil {
			return fmt.Errorf("执行 SOL 链清理失败: %w", err)
		}
	} else {
		// 传统链（BTC、ETH、BSC）使用原有清理逻辑
		if err := s.executeTraditionalCleanup(ctx, chain, cleanupHeight, protectedAddresses, config); err != nil {
			return fmt.Errorf("执行传统链清理失败: %w", err)
		}
	}

	s.logger.Infof(`%s 🧹 链数据清理完成，耗时: %s 链清理基准高度: %d 受保护地址数量: %d\n`,
		chain, time.Since(startTime), cleanupHeight, len(protectedAddresses))
	return nil
}

// shouldCleanup 检查是否需要清理
func (s *DataCleanupScheduler) shouldCleanup(ctx context.Context, chain string, config *DataCleanupConfig) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Table("blocks").
		Where("chain = ? and deleted_at is null", chain).
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
		Where("chain = ? and is_verified = ? and deleted_at is null", chain, 1).
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

// getProtectedHeights 获取需要保护的高度列表（基于受保护地址）
func (s *DataCleanupScheduler) getProtectedHeights(tx *gorm.DB, chain string, cleanupHeight uint64) ([]uint64, error) {
	// 1. 获取受保护的地址
	addresses, err := s.getProtectedAddresses(tx.Statement.Context)
	if err != nil {
		return nil, err
	}

	// 2. 基于受保护地址找到相关的高度
	heights, err := s.getProtectedHeightsByAddresses(tx, chain, addresses, cleanupHeight)
	if err != nil {
		return nil, err
	}

	return heights, nil
}

// getProtectedHeightsByAddresses 基于受保护地址获取高度（优化版本）
func (s *DataCleanupScheduler) getProtectedHeightsByAddresses(tx *gorm.DB, chain string, addresses []string, cleanupHeight uint64) ([]uint64, error) {
	if len(addresses) == 0 {
		return []uint64{}, nil
	}

	heightMap := make(map[uint64]bool)
	batchSize := 1000 // 分批处理，避免 IN 查询过长

	// 分批处理地址列表
	for i := 0; i < len(addresses); i += batchSize {
		end := i + batchSize
		if end > len(addresses) {
			end = len(addresses)
		}
		batchAddresses := addresses[i:end]

		// 1. 从 transaction 表获取受保护地址相关的高度
		if err := s.getHeightsFromTransactions(tx, chain, cleanupHeight, batchAddresses, heightMap); err != nil {
			return nil, err
		}

		// 2. 从 contract_parse_result 表获取受保护地址相关的高度
		if err := s.getHeightsFromContractParseResults(tx, chain, cleanupHeight, batchAddresses, heightMap); err != nil {
			return nil, err
		}
	}

	// 转换为切片并返回
	var uniqueHeights []uint64
	for h := range heightMap {
		uniqueHeights = append(uniqueHeights, h)
	}

	return uniqueHeights, nil
}

// getHeightsFromTransactions 从 transaction 表获取高度
func (s *DataCleanupScheduler) getHeightsFromTransactions(tx *gorm.DB, chain string, cleanupHeight uint64, addresses []string, heightMap map[uint64]bool) error {
	var heights []uint64
	err := tx.Table("transaction").
		Select("DISTINCT height").
		Where("chain = ? AND height < ? AND (address_from IN ? OR address_to IN ?)", chain, cleanupHeight, addresses, addresses).
		Pluck("height", &heights).Error
	if err != nil {
		return err
	}

	for _, h := range heights {
		heightMap[h] = true
	}
	return nil
}

// getHeightsFromContractParseResults 从 contract_parse_results 表获取高度
func (s *DataCleanupScheduler) getHeightsFromContractParseResults(tx *gorm.DB, chain string, cleanupHeight uint64, addresses []string, heightMap map[uint64]bool) error {
	var heights []uint64
	err := tx.Table("contract_parse_result").
		Select("DISTINCT block_number").
		Where("chain = ? AND block_number < ? AND (from_address IN ? OR to_address IN ?)", chain, cleanupHeight, addresses, addresses).
		Pluck("block_number", &heights).Error
	if err != nil {
		return err
	}

	for _, h := range heights {
		heightMap[h] = true
	}
	return nil
}

// cleanupWithProtection 基于受保护的高度和地址进行关联清理
func (s *DataCleanupScheduler) cleanupWithProtection(tx *gorm.DB, chain string, cleanupHeight uint64, protectedHeights []uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	// 1. 清理 transaction 表
	if err := s.cleanupTransactionsWithProtection(tx, chain, cleanupHeight, protectedHeights, protectedAddresses, config); err != nil {
		return fmt.Errorf("清理 transaction 失败: %w", err)
	}

	// 2. 清理 transaction_receipts 表
	if err := s.cleanupTransactionReceiptsWithProtection(tx, chain, cleanupHeight, protectedHeights, protectedAddresses, config); err != nil {
		return fmt.Errorf("清理 transaction_receipts 失败: %w", err)
	}

	// 3. 清理 contract_parse_results 表
	if err := s.cleanupContractParseResultsWithProtection(tx, chain, cleanupHeight, protectedHeights, protectedAddresses, config); err != nil {
		return fmt.Errorf("清理 contract_parse_result 失败: %w", err)
	}

	return nil
}

// executeTraditionalCleanup 执行传统链数据清理（BTC、ETH、BSC）
func (s *DataCleanupScheduler) executeTraditionalCleanup(ctx context.Context, chain string, cleanupHeight uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	// 开始事务
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 获取需要保护的高度列表（基于受保护地址）
	protectedHeights, err := s.getProtectedHeights(tx, chain, cleanupHeight)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("获取受保护高度失败: %w", err)
	}

	s.logger.Infof("发现 %d 个受保护的高度需要保留", len(protectedHeights))

	// 2. 基于受保护的高度进行关联清理
	if err := s.cleanupWithProtection(tx, chain, cleanupHeight, protectedHeights, protectedAddresses, config); err != nil {
		tx.Rollback()
		return fmt.Errorf("执行关联清理失败: %w", err)
	}

	// 3. 清理 blocks 表（最后清理，因为其他表可能依赖它）
	result := tx.Where("chain = ? AND height < ?", chain, cleanupHeight).Delete(&models.Block{})
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("清理 blocks 失败: %w", result.Error)
	}

	// 4. 清理blocks表验证失败的记录
	result = tx.Where("chain = ? AND is_verified = ?", chain, 2).Delete(&models.Block{})
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("清理 blocks 验证失败记录失败: %w", result.Error)
	}

	// 提交事务
	return tx.Commit().Error
}

// cleanupTransactionsWithProtection 基于受保护高度和地址清理 transaction 表
func (s *DataCleanupScheduler) cleanupTransactionsWithProtection(tx *gorm.DB, chain string, cleanupHeight uint64, protectedHeights []uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	// 构建保护条件
	conditions := []string{"chain = ?", "height < ?"}
	args := []interface{}{chain, cleanupHeight}

	// 如果有受保护的高度，排除这些高度
	if len(protectedHeights) > 0 {
		conditions = append(conditions, "height NOT IN ?")
		args = append(args, protectedHeights)
	}

	whereClause := ""
	for i, condition := range conditions {
		if i > 0 {
			whereClause += " AND "
		}
		whereClause += condition
	}

	return s.batchDelete(tx, "transaction", whereClause, args, config.BatchSize)
}

// cleanupTransactionReceiptsWithProtection 基于受保护高度和地址清理 transaction_receipts 表
func (s *DataCleanupScheduler) cleanupTransactionReceiptsWithProtection(tx *gorm.DB, chain string, cleanupHeight uint64, protectedHeights []uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	// 构建保护条件
	conditions := []string{"chain = ?", "block_number < ?"}
	args := []interface{}{chain, cleanupHeight}

	// 如果有受保护的高度，排除这些高度
	if len(protectedHeights) > 0 {
		conditions = append(conditions, "block_number NOT IN ?")
		args = append(args, protectedHeights)
	}

	whereClause := ""
	for i, condition := range conditions {
		if i > 0 {
			whereClause += " AND "
		}
		whereClause += condition
	}

	return s.batchDelete(tx, "transaction_receipts", whereClause, args, config.BatchSize)
}

// cleanupContractParseResultsWithProtection 基于受保护高度和地址清理 contract_parse_result 表
func (s *DataCleanupScheduler) cleanupContractParseResultsWithProtection(tx *gorm.DB, chain string, cleanupHeight uint64, protectedHeights []uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	// 构建保护条件
	conditions := []string{"chain = ?", "block_number < ?"}
	args := []interface{}{chain, cleanupHeight}

	// 如果有受保护的高度，排除这些高度
	if len(protectedHeights) > 0 {
		conditions = append(conditions, "block_number NOT IN ?")
		args = append(args, protectedHeights)
	}

	whereClause := ""
	for i, condition := range conditions {
		if i > 0 {
			whereClause += " AND "
		}
		whereClause += condition
	}

	return s.batchDelete(tx, "contract_parse_result", whereClause, args, config.BatchSize)
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

// executeSolCleanup 执行 SOL 链数据清理
func (s *DataCleanupScheduler) executeSolCleanup(ctx context.Context, chain string, cleanupHeight uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	// 开始事务
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 获取需要保护的高度列表（基于受保护地址）
	protectedHeights, err := s.getSolProtectedHeights(tx, chain, cleanupHeight)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("获取 SOL 受保护高度失败: %w", err)
	}

	s.logger.Infof("发现 %d 个受保护的高度需要保留", len(protectedHeights))

	// 2. 基于受保护的高度进行关联清理
	if err := s.cleanupSolWithProtection(tx, chain, cleanupHeight, protectedHeights, protectedAddresses, config); err != nil {
		tx.Rollback()
		return fmt.Errorf("执行 SOL 关联清理失败: %w", err)
	}

	// 3. 清理 blocks 表（最后清理，因为其他表可能依赖它）
	result := tx.Where("chain = ? AND height < ?", chain, cleanupHeight).Delete(&models.Block{})
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("清理 blocks 失败: %w", result.Error)
	}

	// 4. 清理blocks表验证失败的记录
	result = tx.Where("chain = ? AND is_verified = ?", chain, 2).Delete(&models.Block{})
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("清理 blocks 验证失败记录失败: %w", result.Error)
	}

	// 提交事务
	return tx.Commit().Error
}

// getSolProtectedHeights 获取 SOL 链需要保护的高度列表（基于受保护地址）
func (s *DataCleanupScheduler) getSolProtectedHeights(tx *gorm.DB, chain string, cleanupHeight uint64) ([]uint64, error) {
	// 1. 获取受保护的地址
	addresses, err := s.getProtectedAddresses(tx.Statement.Context)
	if err != nil {
		return nil, err
	}

	// 2. 基于受保护地址找到相关的高度
	heights, err := s.getSolProtectedHeightsByAddresses(tx, chain, addresses, cleanupHeight)
	if err != nil {
		return nil, err
	}

	return heights, nil
}

// getSolProtectedHeightsByAddresses 基于受保护地址获取 SOL 链高度
func (s *DataCleanupScheduler) getSolProtectedHeightsByAddresses(tx *gorm.DB, chain string, addresses []string, cleanupHeight uint64) ([]uint64, error) {
	if len(addresses) == 0 {
		return []uint64{}, nil
	}

	heightMap := make(map[uint64]bool)
	batchSize := 1000 // 分批处理，避免 IN 查询过长

	// 分批处理地址列表
	for i := 0; i < len(addresses); i += batchSize {
		end := i + batchSize
		if end > len(addresses) {
			end = len(addresses)
		}
		batchAddresses := addresses[i:end]

		// 1. 从 sol_tx_detail 表获取受保护地址相关的高度
		if err := s.getHeightsFromSolTxDetail(tx, chain, cleanupHeight, batchAddresses, heightMap); err != nil {
			return nil, err
		}

		// 2. 从 sol_instruction 表获取受保护地址相关的高度
		if err := s.getHeightsFromSolInstruction(tx, chain, cleanupHeight, batchAddresses, heightMap); err != nil {
			return nil, err
		}

		// 3. 从 sol_event 表获取受保护地址相关的高度
		if err := s.getHeightsFromSolEvent(tx, chain, cleanupHeight, batchAddresses, heightMap); err != nil {
			return nil, err
		}

		// 4. 从 sol_parsed_extra 表获取受保护地址相关的高度
		if err := s.getHeightsFromSolParsedExtra(tx, chain, cleanupHeight, batchAddresses, heightMap); err != nil {
			return nil, err
		}
	}

	// 转换为切片并返回
	var uniqueHeights []uint64
	for h := range heightMap {
		uniqueHeights = append(uniqueHeights, h)
	}

	return uniqueHeights, nil
}

// getHeightsFromSolTxDetail 从 sol_tx_detail 表获取高度
func (s *DataCleanupScheduler) getHeightsFromSolTxDetail(tx *gorm.DB, chain string, cleanupHeight uint64, addresses []string, heightMap map[uint64]bool) error {
	var heights []uint64
	err := tx.Table("sol_tx_detail").
		Select("DISTINCT slot").
		Where("slot < ?", cleanupHeight).
		Pluck("slot", &heights).Error
	if err != nil {
		return err
	}

	for _, h := range heights {
		heightMap[h] = true
	}
	return nil
}

// getHeightsFromSolInstruction 从 sol_instruction 表获取高度
func (s *DataCleanupScheduler) getHeightsFromSolInstruction(tx *gorm.DB, chain string, cleanupHeight uint64, addresses []string, heightMap map[uint64]bool) error {
	var heights []uint64
	err := tx.Table("sol_instruction").
		Select("DISTINCT slot").
		Where("slot < ?", cleanupHeight).
		Pluck("slot", &heights).Error
	if err != nil {
		return err
	}

	for _, h := range heights {
		heightMap[h] = true
	}
	return nil
}

// getHeightsFromSolEvent 从 sol_event 表获取高度
func (s *DataCleanupScheduler) getHeightsFromSolEvent(tx *gorm.DB, chain string, cleanupHeight uint64, addresses []string, heightMap map[uint64]bool) error {
	var heights []uint64
	err := tx.Table("sol_event").
		Select("DISTINCT slot").
		Where("slot < ? AND (from_address IN ? OR to_address IN ?)", cleanupHeight, addresses, addresses).
		Pluck("slot", &heights).Error
	if err != nil {
		return err
	}

	for _, h := range heights {
		heightMap[h] = true
	}
	return nil
}

// getHeightsFromSolParsedExtra 从 sol_parsed_extra 表获取高度
func (s *DataCleanupScheduler) getHeightsFromSolParsedExtra(tx *gorm.DB, chain string, cleanupHeight uint64, addresses []string, heightMap map[uint64]bool) error {
	var heights []uint64
	err := tx.Table("sol_parsed_extra").
		Select("DISTINCT slot").
		Where("slot < ?", cleanupHeight).
		Pluck("slot", &heights).Error
	if err != nil {
		return err
	}

	for _, h := range heights {
		heightMap[h] = true
	}
	return nil
}

// cleanupSolWithProtection 基于受保护的高度和地址进行 SOL 链关联清理
func (s *DataCleanupScheduler) cleanupSolWithProtection(tx *gorm.DB, chain string, cleanupHeight uint64, protectedHeights []uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	// 1. 清理 sol_instruction 表
	if err := s.cleanupSolInstructionWithProtection(tx, chain, cleanupHeight, protectedHeights, protectedAddresses, config); err != nil {
		return fmt.Errorf("清理 sol_instruction 失败: %w", err)
	}

	// 2. 清理 sol_parsed_extra 表
	if err := s.cleanupSolParsedExtraWithProtection(tx, chain, cleanupHeight, protectedHeights, protectedAddresses, config); err != nil {
		return fmt.Errorf("清理 sol_parsed_extra 失败: %w", err)
	}

	// 3. 清理 sol_tx_detail 表
	if err := s.cleanupSolTxDetailWithProtection(tx, chain, cleanupHeight, protectedHeights, protectedAddresses, config); err != nil {
		return fmt.Errorf("清理 sol_tx_detail 失败: %w", err)
	}

	// 4. 清理 sol_event 表
	if err := s.cleanupSolEventWithProtection(tx, chain, cleanupHeight, protectedHeights, protectedAddresses, config); err != nil {
		return fmt.Errorf("清理 sol_event 失败: %w", err)
	}

	return nil
}

// cleanupSolInstructionWithProtection 基于受保护高度清理 sol_instruction 表
func (s *DataCleanupScheduler) cleanupSolInstructionWithProtection(tx *gorm.DB, chain string, cleanupHeight uint64, protectedHeights []uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	// 构建保护条件
	conditions := []string{"slot < ?"}
	args := []interface{}{cleanupHeight}

	// 如果有受保护的高度，排除这些高度
	if len(protectedHeights) > 0 {
		conditions = append(conditions, "slot NOT IN ?")
		args = append(args, protectedHeights)
	}

	whereClause := ""
	for i, condition := range conditions {
		if i > 0 {
			whereClause += " AND "
		}
		whereClause += condition
	}

	return s.batchDelete(tx, "sol_instruction", whereClause, args, config.BatchSize)
}

// cleanupSolParsedExtraWithProtection 基于受保护高度清理 sol_parsed_extra 表
func (s *DataCleanupScheduler) cleanupSolParsedExtraWithProtection(tx *gorm.DB, chain string, cleanupHeight uint64, protectedHeights []uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	// 构建保护条件
	conditions := []string{"slot < ?"}
	args := []interface{}{cleanupHeight}

	// 如果有受保护的高度，排除这些高度
	if len(protectedHeights) > 0 {
		conditions = append(conditions, "slot NOT IN ?")
		args = append(args, protectedHeights)
	}

	whereClause := ""
	for i, condition := range conditions {
		if i > 0 {
			whereClause += " AND "
		}
		whereClause += condition
	}

	return s.batchDelete(tx, "sol_parsed_extra", whereClause, args, config.BatchSize)
}

// cleanupSolTxDetailWithProtection 基于受保护高度清理 sol_tx_detail 表
func (s *DataCleanupScheduler) cleanupSolTxDetailWithProtection(tx *gorm.DB, chain string, cleanupHeight uint64, protectedHeights []uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	// 构建保护条件
	conditions := []string{"slot < ?"}
	args := []interface{}{cleanupHeight}

	// 如果有受保护的高度，排除这些高度
	if len(protectedHeights) > 0 {
		conditions = append(conditions, "slot NOT IN ?")
		args = append(args, protectedHeights)
	}

	whereClause := ""
	for i, condition := range conditions {
		if i > 0 {
			whereClause += " AND "
		}
		whereClause += condition
	}

	return s.batchDelete(tx, "sol_tx_detail", whereClause, args, config.BatchSize)
}

// cleanupSolEventWithProtection 基于受保护高度清理 sol_event 表
func (s *DataCleanupScheduler) cleanupSolEventWithProtection(tx *gorm.DB, chain string, cleanupHeight uint64, protectedHeights []uint64, protectedAddresses []string, config *DataCleanupConfig) error {
	// 构建保护条件
	conditions := []string{"slot < ?"}
	args := []interface{}{cleanupHeight}

	// 如果有受保护的高度，排除这些高度
	if len(protectedHeights) > 0 {
		conditions = append(conditions, "slot NOT IN ?")
		args = append(args, protectedHeights)
	}

	whereClause := ""
	for i, condition := range conditions {
		if i > 0 {
			whereClause += " AND "
		}
		whereClause += condition
	}

	return s.batchDelete(tx, "sol_event", whereClause, args, config.BatchSize)
}
