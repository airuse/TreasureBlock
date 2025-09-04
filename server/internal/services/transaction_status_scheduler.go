package services

import (
	"context"
	"strconv"
	"strings"
	"time"

	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"blockChainBrowser/server/internal/utils"

	"github.com/sirupsen/logrus"
)

// TransactionStatusScheduler 交易状态调度器
type TransactionStatusScheduler struct {
	userTxRepo repository.UserTransactionRepository
	logger     *logrus.Logger
	rpcManager *utils.RPCClientManager
	baseCfgSvc BaseConfigService
}

// NewTransactionStatusScheduler 创建交易状态调度器
func NewTransactionStatusScheduler() *TransactionStatusScheduler {
	return &TransactionStatusScheduler{
		userTxRepo: repository.NewUserTransactionRepository(),
		logger:     logrus.New(),
		rpcManager: utils.NewRPCClientManager(),
		baseCfgSvc: NewBaseConfigService(repository.NewBaseConfigRepository()),
	}
}

// Start 启动调度器
func (s *TransactionStatusScheduler) Start(ctx context.Context) {
	s.logger.Info("交易状态调度器已启动")

	// 每30秒检查一次需要更新的交易
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("交易状态调度器已停止")
			return
		case <-ticker.C:
			s.updatePendingTransactions(ctx)
		}
	}
}

// updatePendingTransactions 更新待确认的交易状态
func (s *TransactionStatusScheduler) updatePendingTransactions(ctx context.Context) {
	// 获取所有在途和已打包状态的交易
	statuses := []string{"in_progress", "packed"}

	for _, status := range statuses {
		// 这里需要添加一个方法来获取指定状态的交易列表
		// 暂时使用现有的方法，实际应该添加一个专门的查询方法
		transactions, err := s.userTxRepo.GetByStatus(ctx, 0, status) // userID=0表示查询所有用户
		if err != nil {
			s.logger.Errorf("获取%s状态交易失败: %v", status, err)
			continue
		}

		for _, tx := range transactions {
			s.updateTransactionStatus(ctx, tx)
		}
	}
}

// updateTransactionStatus 更新单个交易状态
func (s *TransactionStatusScheduler) updateTransactionStatus(ctx context.Context, tx *models.UserTransaction) {
	// 检查是否有交易哈希
	if tx.TxHash == nil || *tx.TxHash == "" {
		return
	}

	// 查询交易状态
	txStatus, err := s.rpcManager.GetTransactionStatus(ctx, tx.Chain, *tx.TxHash)
	if err != nil {
		s.logger.Errorf("查询交易状态失败: ID=%d, TxHash=%s, Error=%v", tx.ID, *tx.TxHash, err)
		return
	}

	// 根据查询结果更新状态
	oldStatus := tx.Status
	needUpdate := false

	switch txStatus.Status {
	case "pending":
		if tx.Status != "in_progress" {
			tx.Status = "in_progress"
			needUpdate = true
		}
	case "confirmed":
		// 根据安全块高阈值决定是 packed 还是 confirmed
		threshold := s.getSafeConfirmations(ctx, tx.Chain)
		if txStatus.BlockHeight > 0 {
			tx.BlockHeight = &txStatus.BlockHeight
		}
		if txStatus.Confirmations > 0 {
			confirmations := uint(txStatus.Confirmations)
			tx.Confirmations = &confirmations
		}
		desired := "confirmed"
		if threshold > 0 && txStatus.Confirmations < threshold {
			desired = "packed"
		}
		if tx.Status != desired {
			tx.Status = desired
			needUpdate = true
		}
	case "failed":
		if tx.Status != "failed" {
			tx.Status = "failed"
			errorMsg := "交易在区块链上失败"
			tx.ErrorMsg = &errorMsg
			needUpdate = true
		}
	}

	// 如果状态有变化，保存更新
	if needUpdate {
		if err := s.userTxRepo.Update(ctx, tx); err != nil {
			s.logger.Errorf("更新交易状态失败: ID=%d, Error=%v", tx.ID, err)
		} else {
			s.logger.Infof("交易状态已更新: ID=%d, 从 %s 到 %s", tx.ID, oldStatus, tx.Status)
		}
	}
}

// getSafeConfirmations 从 base_config 读取确认阈值（group=scan, config_type=1, key=confirmations_<chain>）
func (s *TransactionStatusScheduler) getSafeConfirmations(ctx context.Context, chain string) uint64 {
	key := "confirmations_" + strings.ToLower(chain)
	cfg, err := s.baseCfgSvc.GetByConfigKey(ctx, key, 1, "scan")
	if err != nil || cfg == nil || cfg.ConfigValue == "" {
		return 0
	}
	v, err := strconv.ParseUint(cfg.ConfigValue, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// Stop 停止调度器
func (s *TransactionStatusScheduler) Stop() {
	if s.rpcManager != nil {
		s.rpcManager.Close()
	}
	s.logger.Info("交易状态调度器已关闭")
}
