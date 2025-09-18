package failover

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
)

// SOLNodeStatus 节点状态
type SOLNodeStatus int

const (
	SOLNodeHealthy  SOLNodeStatus = iota // 健康
	SOLNodeOverheat                      // 过热（如429等限流）
	SOLNodeDamaged                       // 故障（如5xx/解析失败/超时）
)

// SOLNodeState 节点状态结构
type SOLNodeState struct {
	status        SOLNodeStatus
	firstCallTime time.Time
	lastCallTime  time.Time
	restUntil     time.Time
	totalRequests int64
}

// SOLRequestTask 请求任务
type SOLRequestTask struct {
	ID        string
	Operation string
	Callback  func(*rpc.Client) (interface{}, error)
}

// SOLFailoverManager Solana 故障转移调度器（基于 RPC Client）
// localClient 可为空；externalClients 长度可以为0或多个
// 注意：该管理器是轻量即时执行的，不维护后台协程
type SOLFailoverManager struct {
	localClient     *rpc.Client
	externalClients []*rpc.Client
	nodeStates      []*SOLNodeState
	currentIndex    int64
}

// NewSOLFailoverManager 创建管理器
func NewSOLFailoverManager(localClient *rpc.Client, externalClients []*rpc.Client) *SOLFailoverManager {
	states := make([]*SOLNodeState, len(externalClients))
	for i := range states {
		states[i] = &SOLNodeState{status: SOLNodeHealthy}
	}
	return &SOLFailoverManager{
		localClient:     localClient,
		externalClients: externalClients,
		nodeStates:      states,
	}
}

// Execute 执行带故障转移的请求
func (m *SOLFailoverManager) Execute(operation string, cb func(*rpc.Client) (interface{}, error)) (interface{}, error) {
	start := time.Now()

	// 尝试本地节点
	if m.localClient != nil {
		if data, err := cb(m.localClient); err == nil {
			return data, nil
		}
	}

	// 轮询外部节点，直到成功或全部失败
	if len(m.externalClients) == 0 {
		return nil, fmt.Errorf("failed to %s: no nodes configured", operation)
	}

	startIndex := int(atomic.AddInt64(&m.currentIndex, 1)) % len(m.externalClients)
	for i := 0; i < len(m.externalClients); i++ {
		idx := (startIndex + i) % len(m.externalClients)
		client := m.externalClients[idx]
		node := m.nodeStates[idx]

		// 检查休息
		now := time.Now()
		if node.status != SOLNodeHealthy && now.Before(node.restUntil) {
			continue
		}
		if node.status != SOLNodeHealthy && now.After(node.restUntil) {
			m.resetNode(idx)
		}

		if node.firstCallTime.IsZero() {
			node.firstCallTime = now
		}
		node.lastCallTime = now
		atomic.AddInt64(&node.totalRequests, 1)

		data, err := cb(client)
		if err == nil {
			return data, nil
		}

		// 根据错误设置休息策略
		if m.isRateLimit(err) {
			node.status = SOLNodeOverheat
			used := node.lastCallTime.Sub(node.firstCallTime)
			rest := time.Second - used
			if rest < 0 {
				rest = time.Millisecond * 10
			}
			node.restUntil = now.Add(rest)
		} else {
			node.status = SOLNodeDamaged
			node.restUntil = now.Add(10 * time.Second)
		}
	}

	return nil, fmt.Errorf("failed to %s: all nodes failed after %v", operation, time.Since(start))
}

// 便捷方法族
func (m *SOLFailoverManager) CallWithFailover(operation string, cb func(*rpc.Client) error) error {
	_, err := m.Execute(operation, func(client *rpc.Client) (interface{}, error) {
		return nil, cb(client)
	})
	return err
}

func (m *SOLFailoverManager) CallWithFailoverUint64(operation string, cb func(*rpc.Client) (uint64, error)) (uint64, error) {
	res, err := m.Execute(operation, func(client *rpc.Client) (interface{}, error) {
		return cb(client)
	})
	if err != nil {
		return 0, err
	}
	if v, ok := res.(uint64); ok {
		return v, nil
	}
	return 0, fmt.Errorf("unexpected result type")
}

func (m *SOLFailoverManager) CallWithFailoverSlot(operation string, cb func(*rpc.Client) (uint64, error)) (uint64, error) {
	res, err := m.Execute(operation, func(client *rpc.Client) (interface{}, error) {
		return cb(client)
	})
	if err != nil {
		return 0, err
	}
	if v, ok := res.(uint64); ok {
		return v, nil
	}
	return 0, fmt.Errorf("unexpected result type")
}

func (m *SOLFailoverManager) CallWithFailoverBlock(operation string, cb func(*rpc.Client) (*rpc.GetBlockResult, error)) (*rpc.GetBlockResult, error) {
	res, err := m.Execute(operation, func(client *rpc.Client) (interface{}, error) {
		return cb(client)
	})
	if err != nil {
		return nil, err
	}
	if v, ok := res.(*rpc.GetBlockResult); ok {
		return v, nil
	}
	return nil, fmt.Errorf("unexpected result type")
}

func (m *SOLFailoverManager) CallWithFailoverString(operation string, cb func(*rpc.Client) (string, error)) (string, error) {
	res, err := m.Execute(operation, func(client *rpc.Client) (interface{}, error) {
		return cb(client)
	})
	if err != nil {
		return "", err
	}
	if v, ok := res.(string); ok {
		return v, nil
	}
	return "", fmt.Errorf("unexpected result type")
}

func (m *SOLFailoverManager) CallWithFailoverMap(operation string, cb func(*rpc.Client) (map[string]interface{}, error)) (map[string]interface{}, error) {
	res, err := m.Execute(operation, func(client *rpc.Client) (interface{}, error) {
		return cb(client)
	})
	if err != nil {
		return nil, err
	}
	if v, ok := res.(map[string]interface{}); ok {
		return v, nil
	}
	return nil, fmt.Errorf("unexpected result type")
}

func (m *SOLFailoverManager) CallWithFailoverMaps(operation string, cb func(*rpc.Client) ([]map[string]interface{}, error)) ([]map[string]interface{}, error) {
	res, err := m.Execute(operation, func(client *rpc.Client) (interface{}, error) {
		return cb(client)
	})
	if err != nil {
		return nil, err
	}
	if v, ok := res.([]map[string]interface{}); ok {
		return v, nil
	}
	return nil, fmt.Errorf("unexpected result type")
}

// 私有工具
func (m *SOLFailoverManager) resetNode(idx int) {
	n := m.nodeStates[idx]
	n.status = SOLNodeHealthy
	n.firstCallTime = time.Time{}
	n.lastCallTime = time.Time{}
	n.restUntil = time.Time{}
}

func (m *SOLFailoverManager) isRateLimit(err error) bool {
	if err == nil {
		return false
	}
	es := err.Error()
	return strings.Contains(es, "429") ||
		strings.Contains(es, "rate limit") ||
		strings.Contains(es, "too many requests") ||
		strings.Contains(es, "request limit")
}

// GetStats 获取统计信息
func (m *SOLFailoverManager) GetStats() map[string]interface{} {
	healthyNodes := 0
	for _, node := range m.nodeStates {
		if node.status == SOLNodeHealthy {
			healthyNodes++
		}
	}

	return map[string]interface{}{
		"healthy_nodes": healthyNodes,
		"total_nodes":   len(m.nodeStates),
		"local_client":  m.localClient != nil,
	}
}
