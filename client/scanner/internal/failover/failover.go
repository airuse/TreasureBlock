package failover

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ExternalNodeStatus 外部节点状态
type ExternalNodeStatus struct {
	isHealthy        bool
	consecutiveFails int
	lastFailTime     time.Time
	lastSuccessTime  time.Time
	requestCount     int64     // 请求计数
	lastResetTime    time.Time // 上次重置计数的时间
}

// FailoverManager 智能负载均衡管理器
type FailoverManager struct {
	localClient     *ethclient.Client
	externalClients []*ethclient.Client
	nodeStatus      []*ExternalNodeStatus // 外部节点状态
	currentIndex    int                   // 当前轮询索引（仅用于外部节点）
	mutex           sync.RWMutex
}

// NewFailoverManager 创建智能负载均衡管理器
func NewFailoverManager(localClient *ethclient.Client, externalClients []*ethclient.Client) *FailoverManager {
	// 初始化外部节点状态
	nodeStatus := make([]*ExternalNodeStatus, len(externalClients))
	for i := range nodeStatus {
		nodeStatus[i] = &ExternalNodeStatus{
			isHealthy:        true,
			consecutiveFails: 0,
			lastFailTime:     time.Time{},
			lastSuccessTime:  time.Now(),
			requestCount:     0,
			lastResetTime:    time.Now(),
		}
	}

	return &FailoverManager{
		localClient:     localClient,
		externalClients: externalClients,
		nodeStatus:      nodeStatus,
		currentIndex:    0,
	}
}

// getNextHealthyExternalNode 获取下一个健康的外部节点
func (fm *FailoverManager) getNextHealthyExternalNode() (*ethclient.Client, int) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if len(fm.externalClients) == 0 {
		return nil, -1
	}

	// 轮询查找健康且未达到限流的节点
	for attempts := 0; attempts < len(fm.externalClients); attempts++ {
		index := fm.currentIndex % len(fm.externalClients)

		// 检查节点是否健康
		if fm.isNodeHealthy(index) && fm.canMakeRequest(index) {
			fm.currentIndex = (fm.currentIndex + 1) % len(fm.externalClients)
			return fm.externalClients[index], index
		}

		fm.currentIndex = (fm.currentIndex + 1) % len(fm.externalClients)
	}

	// 如果没有找到健康节点，返回当前节点（让它尝试并更新状态）
	index := fm.currentIndex % len(fm.externalClients)
	return fm.externalClients[index], index
}

// isNodeHealthy 检查节点是否健康
func (fm *FailoverManager) isNodeHealthy(index int) bool {
	if index >= len(fm.nodeStatus) {
		return false
	}

	status := fm.nodeStatus[index]
	now := time.Now()

	// 如果连续失败超过3次，且最后一次失败在5分钟内，认为不健康
	if status.consecutiveFails >= 3 && now.Sub(status.lastFailTime) < 5*time.Minute {
		return false
	}

	// 如果最后一次成功在10分钟前，尝试重新检测
	if now.Sub(status.lastSuccessTime) > 10*time.Minute {
		// 重置状态，给它一次机会
		status.consecutiveFails = 0
		status.isHealthy = true
	}

	return status.isHealthy
}

// canMakeRequest 检查是否可以向节点发起请求（限流检测）
func (fm *FailoverManager) canMakeRequest(index int) bool {
	if index >= len(fm.nodeStatus) {
		return false
	}

	status := fm.nodeStatus[index]
	now := time.Now()

	// 每分钟重置计数
	if now.Sub(status.lastResetTime) >= time.Minute {
		status.requestCount = 0
		status.lastResetTime = now
	}

	// QuickNode 免费版限制：15 requests/second = 900 requests/minute
	// 为了安全，我们限制为 50 requests/minute
	return status.requestCount < 50
}

// recordSuccess 记录节点成功
func (fm *FailoverManager) recordSuccess(index int) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if index >= 0 && index < len(fm.nodeStatus) {
		status := fm.nodeStatus[index]
		status.isHealthy = true
		status.consecutiveFails = 0
		status.lastSuccessTime = time.Now()
		status.requestCount++
		// 移除成功日志，保持安静
	}
}

// recordFailure 记录节点失败
func (fm *FailoverManager) recordFailure(index int, err error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if index >= 0 && index < len(fm.nodeStatus) {
		status := fm.nodeStatus[index]
		status.consecutiveFails++
		status.lastFailTime = time.Now()
		status.requestCount++

		// 连续失败3次后标记为不健康
		if status.consecutiveFails >= 3 {
			status.isHealthy = false
			fmt.Printf("[Failover] External node %d marked unhealthy (consecutive fails: %d)\n", index, status.consecutiveFails)
		}

		fmt.Printf("[Failover] External node %d failed (fails: %d): %v\n", index, status.consecutiveFails, err)
	}
}

// CallWithFailoverUint64 智能负载均衡调用方法（返回uint64）
func (fm *FailoverManager) CallWithFailoverUint64(operation string, clientCall func(*ethclient.Client) (uint64, error)) (uint64, error) {
	// 首先尝试本地节点（本地节点快且稳定）
	if fm.localClient != nil {
		result, err := clientCall(fm.localClient)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[Failover] Local node failed for %s: %v, using external node rotation\n", operation, err)
	}

	// 使用智能外部节点选择
	client, index := fm.getNextHealthyExternalNode()
	if client != nil {
		result, err := clientCall(client)
		if err == nil {
			fm.recordSuccess(index)
			return result, nil
		}
		fm.recordFailure(index, err)
	}

	return 0, fmt.Errorf("failed to %s: no healthy external nodes available", operation)
}

// CallWithFailoverBlock 智能负载均衡调用方法（返回*types.Block）
func (fm *FailoverManager) CallWithFailoverBlock(operation string, clientCall func(*ethclient.Client) (*types.Block, error)) (*types.Block, error) {
	// 首先尝试本地节点
	if fm.localClient != nil {
		result, err := clientCall(fm.localClient)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[Failover] Local node failed for %s: %v, using external node rotation\n", operation, err)
	}

	// 使用智能外部节点选择
	client, index := fm.getNextHealthyExternalNode()
	if client != nil {
		result, err := clientCall(client)
		if err == nil {
			fm.recordSuccess(index)
			return result, nil
		}
		fm.recordFailure(index, err)
	}

	return nil, fmt.Errorf("failed to %s: no healthy external nodes available", operation)
}

// CallWithFailoverReceipt 智能负载均衡调用方法（返回*types.Receipt）
func (fm *FailoverManager) CallWithFailoverReceipt(operation string, clientCall func(*ethclient.Client) (*types.Receipt, error)) (*types.Receipt, error) {
	// 首先尝试本地节点
	if fm.localClient != nil {
		result, err := clientCall(fm.localClient)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[Failover] Local node failed for %s: %v, using external node rotation\n", operation, err)
	}

	// 使用智能外部节点选择
	client, index := fm.getNextHealthyExternalNode()
	if client != nil {
		result, err := clientCall(client)
		if err == nil {
			fm.recordSuccess(index)
			return result, nil
		}
		fm.recordFailure(index, err)
	}

	return nil, fmt.Errorf("failed to %s: no healthy external nodes available", operation)
}

// CallWithFailoverTransactions 通用的故障转移调用方法（返回[]map[string]interface{}）
func (fm *FailoverManager) CallWithFailoverTransactions(operation string, clientCall func(*ethclient.Client) ([]map[string]interface{}, error)) ([]map[string]interface{}, error) {
	// 首先尝试本地节点
	if fm.localClient != nil {
		result, err := clientCall(fm.localClient)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[Failover] Local node failed for %s: %v, trying external APIs\n", operation, err)
	}

	// 如果本地节点失败或不存在，尝试外部节点
	if len(fm.externalClients) > 0 {
		// 使用智能外部节点选择
		client, index := fm.getNextHealthyExternalNode()
		if client != nil {
			result, err := clientCall(client)
			if err == nil {
				fm.recordSuccess(index)
				return result, nil
			}
			fm.recordFailure(index, err)
		}
	}

	return nil, fmt.Errorf("failed to %s: no healthy external nodes available", operation)
}

// CallWithFailoverBytes 字节数组类型的故障转移调用方法
func (fm *FailoverManager) CallWithFailoverBytes(operation string, clientCall func(*ethclient.Client) ([]byte, error)) ([]byte, error) {
	// 首先尝试本地节点
	if fm.localClient != nil {
		if result, err := clientCall(fm.localClient); err == nil {
			return result, nil
		}
		fmt.Printf("[Failover] Local node failed for %s, trying external APIs\n", operation)
	}

	// 如果本地节点失败或不存在，尝试外部节点
	if len(fm.externalClients) > 0 {
		// 使用智能外部节点选择
		client, index := fm.getNextHealthyExternalNode()
		if client != nil {
			result, err := clientCall(client)
			if err == nil {
				fm.recordSuccess(index)
				return result, nil
			}
			fm.recordFailure(index, err)
		}
	}

	return nil, fmt.Errorf("failed to %s: no healthy external nodes available", operation)
}

// CallWithFailoverRawBlock 通用的故障转移调用方法（返回*types.Block）
func (fm *FailoverManager) CallWithFailoverRawBlock(operation string, clientCall func(*ethclient.Client) (*types.Block, error)) (*types.Block, error) {
	// 首先尝试本地节点
	if fm.localClient != nil {
		result, err := clientCall(fm.localClient)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[Failover] Local node failed for %s: %v, trying external APIs\n", operation, err)
	}

	// 如果本地节点失败或不存在，尝试外部节点
	if len(fm.externalClients) > 0 {
		// 使用智能外部节点选择
		client, index := fm.getNextHealthyExternalNode()
		if client != nil {
			result, err := clientCall(client)
			if err == nil {
				fm.recordSuccess(index)
				return result, nil
			}
			fm.recordFailure(index, err)
		}
	}

	return nil, fmt.Errorf("failed to %s: no healthy external nodes available", operation)
}

// CallWithFailover 智能负载均衡调用方法（无返回值）
func (fm *FailoverManager) CallWithFailover(operation string, clientCall func(*ethclient.Client) error) error {
	// 首先尝试本地节点
	if fm.localClient != nil {
		err := clientCall(fm.localClient)
		if err == nil {
			return nil
		}
		fmt.Printf("[Failover] Local node failed for %s: %v, using external node rotation\n", operation, err)
	}

	// 使用智能外部节点选择
	client, index := fm.getNextHealthyExternalNode()
	if client != nil {
		err := clientCall(client)
		if err == nil {
			fm.recordSuccess(index)
			return nil
		}
		fm.recordFailure(index, err)
	}

	return fmt.Errorf("failed to %s: no healthy external nodes available", operation)
}

// GetLocalClient 获取本地客户端
func (fm *FailoverManager) GetLocalClient() *ethclient.Client {
	return fm.localClient
}

// GetExternalClients 获取外部客户端
func (fm *FailoverManager) GetExternalClients() []*ethclient.Client {
	return fm.externalClients
}
