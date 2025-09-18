package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/interfaces"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"blockChainBrowser/server/internal/utils"

	"github.com/sirupsen/logrus"
)

// TransactionStatusScheduler 交易状态调度器
type TransactionStatusScheduler struct {
	userTxRepo  repository.UserTransactionRepository
	btcUtxoRepo repository.BTCUTXORepository
	logger      *logrus.Logger
	rpcManager  *utils.RPCClientManager
	baseCfgSvc  BaseConfigService
	wsHandler   interfaces.WebSocketBroadcaster // WebSocket广播接口
}

// NewTransactionStatusScheduler 创建交易状态调度器
func NewTransactionStatusScheduler(userTxRepo repository.UserTransactionRepository, btcUtxoRepo repository.BTCUTXORepository) *TransactionStatusScheduler {
	return &TransactionStatusScheduler{
		userTxRepo:  userTxRepo,
		btcUtxoRepo: btcUtxoRepo,
		logger:      logrus.New(),
		rpcManager:  utils.NewRPCClientManager(),
		baseCfgSvc:  NewBaseConfigService(repository.NewBaseConfigRepository()),
	}
}

// SetWebSocketHandler 设置WebSocket处理器
func (s *TransactionStatusScheduler) SetWebSocketHandler(handler interfaces.WebSocketBroadcaster) {
	s.wsHandler = handler
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
		confirmations := uint(txStatus.Confirmations)
		tx.Confirmations = &confirmations
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

			if tx.Chain == "btc" && tx.Status == "confirmed" {
				// 更新utxo表数据，把未花费的输出改为已经花费的输出
				fmt.Printf("更新utxo表数据，把未花费的输出改为已经花费的输出: tx:%s blockHeight:%d \n", *tx.TxHash, *tx.BlockHeight)
				s.markBTCInputsAsSpent(ctx, tx)
			}
			// 广播状态更新事件
			s.broadcastTransactionStatusUpdate(tx)
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

// broadcastTransactionStatusUpdate 广播交易状态更新
func (s *TransactionStatusScheduler) broadcastTransactionStatusUpdate(tx *models.UserTransaction) {
	if s.wsHandler == nil {
		return
	}

	// 构建状态更新数据
	statusUpdateData := map[string]interface{}{
		"id":            tx.ID,
		"status":        tx.Status,
		"tx_hash":       tx.TxHash,
		"block_height":  tx.BlockHeight,
		"confirmations": tx.Confirmations,
		"error_msg":     tx.ErrorMsg,
		"updated_at":    tx.UpdatedAt,
	}

	s.logger.Infof("广播交易状态更新: ID=%d, Status=%s, Data=%+v", tx.ID, tx.Status, statusUpdateData)

	// 通过接口调用WebSocket广播方法
	s.wsHandler.BroadcastTransactionStatusUpdate(tx.Chain, statusUpdateData)
}

// markBTCInputsAsSpent 标记BTC交易的输入为已花费
func (s *TransactionStatusScheduler) markBTCInputsAsSpent(ctx context.Context, tx *models.UserTransaction) {
	// 检查是否有BTC交易输入数据
	if tx.BTCTxInJSON == nil || *tx.BTCTxInJSON == "" {
		s.logger.Warnf("BTC交易 %s 没有输入数据，跳过UTXO标记", *tx.TxHash)
		return
	}

	// 解析BTC交易输入数据
	var btcTxIns []dto.BTCTxIn
	if err := json.Unmarshal([]byte(*tx.BTCTxInJSON), &btcTxIns); err != nil {
		s.logger.Errorf("解析BTC交易输入数据失败: %v", err)
		return
	}

	// 标记每个输入为已花费
	for vinIndex, txIn := range btcTxIns {
		if txIn.TxID == "" {
			continue // 跳过空的输入
		}

		// 调用MarkSpent方法标记UTXO为已花费
		// 参数：chain, prevTxID, voutIndex, spentTxID, vinIndex, spentHeight
		err := s.btcUtxoRepo.MarkSpent(
			ctx,
			tx.Chain,         // chain
			txIn.TxID,        // prevTxID (被花费的UTXO的交易ID)
			txIn.Vout,        // voutIndex (被花费的UTXO的输出索引)
			*tx.TxHash,       // spentTxID (当前交易ID)
			uint32(vinIndex), // vinIndex (当前交易中的输入索引)
			*tx.BlockHeight,  // spentHeight (花费发生时的区块高度)
		)

		if err != nil {
			s.logger.Errorf("标记UTXO为已花费失败: prevTxID=%s, voutIndex=%d, spentTxID=%s, vinIndex=%d, error=%v",
				txIn.TxID, txIn.Vout, *tx.TxHash, vinIndex, err)
		} else {
			s.logger.Infof("成功标记UTXO为已花费: prevTxID=%s, voutIndex=%d, spentTxID=%s, vinIndex=%d",
				txIn.TxID, txIn.Vout, *tx.TxHash, vinIndex)
		}
	}
}

// Stop 停止调度器
func (s *TransactionStatusScheduler) Stop() {
	if s.rpcManager != nil {
		s.rpcManager.Close()
	}
	s.logger.Info("交易状态调度器已关闭")
}
