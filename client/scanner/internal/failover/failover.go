package failover

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// NodeStatus 节点状态
type NodeStatus int

const (
	NodeHealthy  NodeStatus = iota // 健康：可以正常使用
	NodeOverheat                   // 过热：429错误，需要休息计算时间
	NodeDamaged                    // 故障：其他错误，休息3分钟
)

// SimpleNodeState 简化的节点状态
type SimpleNodeState struct {
	status        NodeStatus
	firstCallTime time.Time // 第一次调用时间
	lastCallTime  time.Time // 最后一次调用时间
	restUntil     time.Time // 休息到什么时候
	totalRequests int64     // 总请求数（用于轮询）
}

// RequestTask 请求任务
type RequestTask struct {
	ID         string
	Operation  string
	Callback   func(*ethclient.Client) (interface{}, error)
	ResultChan chan *RequestResult
	CreatedAt  time.Time
	RetryCount int
	ShouldSkip bool   // 是否应该跳过
	SkipReason string // 跳过原因
}

// RequestResult 请求结果
type RequestResult struct {
	Data     interface{}
	Error    error
	NodeID   int
	Duration time.Duration
}

// 删除整个PersistentQueue结构体，不再需要队列

// FailoverManager 极简智能调度器
type FailoverManager struct {
	localClient *ethclient.Client
	clients     []*ethclient.Client
	nodeStates  []*SimpleNodeState

	// 简单轮询
	currentIndex int64

	// 统计
	totalTasks     int64
	completedTasks int64
	skippedTasks   int64

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	mutex  sync.RWMutex
}

// NewFailoverManager 创建极简调度器
func NewFailoverManager(localClient *ethclient.Client, externalClients []*ethclient.Client) *FailoverManager {
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化节点状态
	nodeStates := make([]*SimpleNodeState, len(externalClients))
	for i := range nodeStates {
		nodeStates[i] = &SimpleNodeState{
			status:        NodeHealthy,
			totalRequests: 0,
		}
	}

	fm := &FailoverManager{
		localClient: localClient,
		clients:     externalClients,
		nodeStates:  nodeStates,
		ctx:         ctx,
		cancel:      cancel,
	}

	return fm
}

// ExecuteRequest 执行请求
func (fm *FailoverManager) ExecuteRequest(operation string, callback func(*ethclient.Client) (interface{}, error)) (interface{}, error) {
	task := &RequestTask{
		ID:         fmt.Sprintf("%s_%d", operation, time.Now().UnixNano()),
		Operation:  operation,
		Callback:   callback,
		ResultChan: make(chan *RequestResult, 1),
		CreatedAt:  time.Now(),
	}

	atomic.AddInt64(&fm.totalTasks, 1)

	// 立即尝试处理
	go fm.processTask(task)

	// 等待结果
	select {
	case result := <-task.ResultChan:
		return result.Data, result.Error
	case <-time.After(time.Minute * 3):
		fmt.Printf("[Simple Scheduler] Task %s timeout, but continuing in background\n", task.ID)
		return nil, fmt.Errorf("processing in background")
	}
}

// processTask 处理任务
func (fm *FailoverManager) processTask(task *RequestTask) {
	startTime := time.Now()

	// 尝试本地节点
	if fm.localClient != nil {
		if result, err := task.Callback(fm.localClient); err == nil {
			fm.completeTask(task, &RequestResult{
				Data:     result,
				NodeID:   -1,
				Duration: time.Since(startTime),
			})
			return
		}
	}

	// 选择外部节点，无限轮询，直到返回结果
	for {
		nodeID := fm.selectNode()
		if nodeID == -1 {
			// 所有节点都在休息，等待一下，避免频繁轮询
			time.Sleep(time.Millisecond * 10)
			continue
		}

		client := fm.clients[nodeID]
		node := fm.nodeStates[nodeID]

		// 记录调用时间
		now := time.Now()
		if node.firstCallTime.IsZero() {
			node.firstCallTime = now
		}
		node.lastCallTime = now
		atomic.AddInt64(&node.totalRequests, 1)

		// 执行请求
		result, err := task.Callback(client)
		duration := time.Since(startTime)

		if err == nil {
			// 成功！不重置状态，保持时间窗口用于下次计算
			fm.completeTask(task, &RequestResult{
				Data:     result,
				NodeID:   nodeID,
				Duration: duration,
			})
			return
		}

		// 失败，处理错误
		if fm.isUnrecoverableError(err) {
			// 不可恢复错误，标记为跳过
			task.ShouldSkip = true
			task.SkipReason = err.Error()
			fmt.Printf("[Simple Scheduler] Task %s marked as SKIP due to unrecoverable error: %v\n", task.ID, err)
			// 直接返回错误，不重试
			fm.completeTask(task, &RequestResult{
				Data:     nil,
				Error:    err,
				NodeID:   nodeID,
				Duration: duration,
			})
			return
		}

		fm.handleNodeError(nodeID, err)
		// 继续循环，尝试下一个节点
	}
}

// selectNode 选择可用节点（简单轮询）
func (fm *FailoverManager) selectNode() int {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	now := time.Now()
	startIndex := int(atomic.AddInt64(&fm.currentIndex, 1)) % len(fm.clients)

	// 轮询查找可用节点
	for i := 0; i < len(fm.clients); i++ {
		index := (startIndex + i) % len(fm.clients)
		node := fm.nodeStates[index]

		// 检查是否还在休息
		if node.status != NodeHealthy && now.Before(node.restUntil) {
			continue
		}

		// 如果休息时间到了，恢复健康
		if node.status != NodeHealthy && now.After(node.restUntil) {
			// fmt.Printf("[Simple Scheduler] Node %d RECOVERED from %s\n", index, fm.getStatusString(node.status))
			fm.resetNodeState(index)
		}

		return index
	}

	return -1 // 没有可用节点
}

// handleNodeError 处理节点错误
func (fm *FailoverManager) handleNodeError(nodeID int, err error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	node := fm.nodeStates[nodeID]
	now := time.Now()

	if fm.is429Error(err) {
		// 429错误 - 过热
		node.status = NodeOverheat

		// 计算精确休息时间：1秒 - (最后一次 - 第一次时间)
		usedTime := node.lastCallTime.Sub(node.firstCallTime)
		restTime := time.Second - usedTime
		if restTime < 0 {
			restTime = time.Millisecond * 10 // 最少休息50ms
		}

		node.restUntil = now.Add(restTime)
		// fmt.Printf("[Simple Scheduler] Node %d OVERHEAT, rest for %v (used: %v)\n", nodeID, restTime, usedTime)

	} else {
		// 其他错误 - 故障
		node.status = NodeDamaged
		node.restUntil = now.Add(time.Second * 30) // 固定5秒
		fmt.Printf("[Simple Scheduler] Node %d DAMAGED, rest for 10 Second (error: %v)\n", nodeID, err)
	}
}

// resetNodeState 重置节点状态
func (fm *FailoverManager) resetNodeState(nodeID int) {
	node := fm.nodeStates[nodeID]
	node.status = NodeHealthy
	node.firstCallTime = time.Time{}
	node.lastCallTime = time.Time{}
	node.restUntil = time.Time{}
}

// is429Error 检查是否是限流错误
func (fm *FailoverManager) is429Error(err error) bool {
	if err == nil {
		return false
	}
	errorStr := err.Error()

	// 必须是429错误
	if !strings.Contains(errorStr, "429") {
		return false
	}

	// 必须包含code -32007
	if !strings.Contains(errorStr, "-32007") {
		return false
	}

	// 可选：检查是否包含限流相关消息（提高准确性）
	if !strings.Contains(errorStr, "request limit") {
		return false
	}

	return true
}

// isUnrecoverableError 检查是否是不可恢复的错误
func (fm *FailoverManager) isUnrecoverableError(err error) bool {
	if err == nil {
		return false
	}
	errorStr := err.Error()

	// 交易类型不支持 - 这是业务逻辑错误，重试无意义
	return strings.Contains(errorStr, "transaction type not supported") ||
		strings.Contains(errorStr, "unsupported transaction type") ||
		strings.Contains(errorStr, "invalid transaction type")
}

// completeTask 完成任务
func (fm *FailoverManager) completeTask(task *RequestTask, result *RequestResult) {
	// 如果是跳过的任务，增加跳过计数
	if task.ShouldSkip {
		atomic.AddInt64(&fm.skippedTasks, 1)
	}

	atomic.AddInt64(&fm.completedTasks, 1)

	// 发送结果
	select {
	case task.ResultChan <- result:
	default:
	}
}

// GetStats 获取统计信息
func (fm *FailoverManager) GetStats() map[string]interface{} {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	total := atomic.LoadInt64(&fm.totalTasks)
	completed := atomic.LoadInt64(&fm.completedTasks)

	skipped := atomic.LoadInt64(&fm.skippedTasks)

	healthyNodes := 0
	for _, node := range fm.nodeStates {
		if node.status == NodeHealthy {
			healthyNodes++
		}
	}

	return map[string]interface{}{
		"total_tasks":     total,
		"completed_tasks": completed,
		"skipped_tasks":   skipped,
		"healthy_nodes":   healthyNodes,
		"total_nodes":     len(fm.nodeStates),
	}
}

// 兼容方法
func (fm *FailoverManager) CallWithFailoverUint64(operation string, clientCall func(*ethclient.Client) (uint64, error)) (uint64, error) {
	result, err := fm.ExecuteRequest(operation, func(client *ethclient.Client) (interface{}, error) {
		return clientCall(client)
	})
	if err != nil {
		return 0, err
	}
	if val, ok := result.(uint64); ok {
		return val, nil
	}
	return 0, fmt.Errorf("unexpected result type")
}

func (fm *FailoverManager) CallWithFailoverBlock(operation string, clientCall func(*ethclient.Client) (*types.Block, error)) (*types.Block, error) {
	result, err := fm.ExecuteRequest(operation, func(client *ethclient.Client) (interface{}, error) {
		return clientCall(client)
	})
	if err != nil {
		return nil, err
	}
	if val, ok := result.(*types.Block); ok {
		return val, nil
	}
	return nil, fmt.Errorf("unexpected result type")
}

func (fm *FailoverManager) CallWithFailoverReceipt(operation string, clientCall func(*ethclient.Client) (*types.Receipt, error)) (*types.Receipt, error) {
	result, err := fm.ExecuteRequest(operation, func(client *ethclient.Client) (interface{}, error) {
		return clientCall(client)
	})
	if err != nil {
		return nil, err
	}
	if val, ok := result.(*types.Receipt); ok {
		return val, nil
	}
	return nil, fmt.Errorf("unexpected result type")
}

// CallWithFailoverReceipts 支持获取多个交易回执的故障转移调用方法
func (fm *FailoverManager) CallWithFailoverReceipts(operation string, clientCall func(*ethclient.Client) ([]*types.Receipt, error)) ([]*types.Receipt, error) {
	result, err := fm.ExecuteRequest(operation, func(client *ethclient.Client) (interface{}, error) {
		return clientCall(client)
	})
	if err != nil {
		return nil, err
	}
	if val, ok := result.([]*types.Receipt); ok {
		return val, nil
	}
	return nil, fmt.Errorf("unexpected result type")
}

func (fm *FailoverManager) CallWithFailoverTransactions(operation string, clientCall func(*ethclient.Client) ([]map[string]interface{}, error)) ([]map[string]interface{}, error) {
	result, err := fm.ExecuteRequest(operation, func(client *ethclient.Client) (interface{}, error) {
		return clientCall(client)
	})
	if err != nil {
		return nil, err
	}
	if val, ok := result.([]map[string]interface{}); ok {
		return val, nil
	}
	return nil, fmt.Errorf("unexpected result type")
}

func (fm *FailoverManager) CallWithFailoverBytes(operation string, clientCall func(*ethclient.Client) ([]byte, error)) ([]byte, error) {
	result, err := fm.ExecuteRequest(operation, func(client *ethclient.Client) (interface{}, error) {
		return clientCall(client)
	})
	if err != nil {
		return nil, err
	}
	if val, ok := result.([]byte); ok {
		return val, nil
	}
	return nil, fmt.Errorf("unexpected result type")
}

func (fm *FailoverManager) CallWithFailoverNetworkID(operation string, clientCall func(*ethclient.Client) (*big.Int, error)) (*big.Int, error) {
	result, err := fm.ExecuteRequest(operation, func(client *ethclient.Client) (interface{}, error) {
		return clientCall(client)
	})
	if err != nil {
		return nil, err
	}
	if val, ok := result.(*big.Int); ok {
		return val, nil
	}
	return nil, fmt.Errorf("unexpected result type")
}

func (fm *FailoverManager) CallWithFailover(operation string, clientCall func(*ethclient.Client) error) error {
	_, err := fm.ExecuteRequest(operation, func(client *ethclient.Client) (interface{}, error) {
		return nil, clientCall(client)
	})
	return err
}

func (fm *FailoverManager) CallWithFailoverRawBlock(operation string, clientCall func(*ethclient.Client) (*types.Block, error)) (*types.Block, error) {
	result, err := fm.ExecuteRequest(operation, func(client *ethclient.Client) (interface{}, error) {
		return clientCall(client)
	})
	if err != nil {
		return nil, err
	}
	if val, ok := result.(*types.Block); ok {
		return val, nil
	}
	return nil, fmt.Errorf("unexpected result type")
}

func (fm *FailoverManager) GetLocalClient() *ethclient.Client {
	return fm.localClient
}

func (fm *FailoverManager) GetExternalClients() []*ethclient.Client {
	return fm.clients
}

func (fm *FailoverManager) Shutdown() {
	if fm.cancel != nil {
		fm.cancel()
	}
}
