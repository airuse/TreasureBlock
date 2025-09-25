package failover

import (
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"
)

// BTCNodeStatus 节点状态
type BTCNodeStatus int

const (
	BTCNodeHealthy  BTCNodeStatus = iota // 健康
	BTCNodeOverheat                      // 过热（如429等限流）
	BTCNodeDamaged                       // 故障（如5xx/解析失败/超时）
)

// BTCNodeState 节点状态结构
type BTCNodeState struct {
	status        BTCNodeStatus
	firstCallTime time.Time
	lastCallTime  time.Time
	restUntil     time.Time
	totalRequests int64
}

// BTCRequestTask 请求任务
type BTCRequestTask struct {
	ID        string
	Operation string
	Callback  func(baseURL string) (interface{}, error)
}

// BTCFailoverManager 比特币故障转移调度器（基于HTTP BaseURL）
// localURL 可为空；externalURLs 长度可以为0或多个
// 注意：该管理器是轻量即时执行的，不维护后台协程
type BTCFailoverManager struct {
	localURL      string
	externalURLs  []string
	nodeStates    []*BTCNodeState
	currentIndex  int64
	lastGoodIndex int64 // -1 表示未知
}

// NewBTCFailoverManager 创建管理器
func NewBTCFailoverManager(localURL string, externalURLs []string) *BTCFailoverManager {
	states := make([]*BTCNodeState, len(externalURLs))
	for i := range states {
		states[i] = &BTCNodeState{status: BTCNodeHealthy}
	}
	m := &BTCFailoverManager{
		localURL:      localURL,
		externalURLs:  externalURLs,
		nodeStates:    states,
		lastGoodIndex: -1,
	}
	rand.Seed(time.Now().UnixNano())
	if len(externalURLs) > 0 {
		atomic.StoreInt64(&m.currentIndex, rand.Int63n(int64(len(externalURLs))))
	}
	return m
}

// Execute 执行带故障转移的请求
func (m *BTCFailoverManager) Execute(operation string, cb func(baseURL string) (interface{}, error)) (interface{}, error) {
	start := time.Now()

	// 尝试本地节点
	if m.localURL != "" {
		if data, err := cb(m.localURL); err == nil {
			return data, nil
		}
	}

	// 轮询外部节点，直到成功或全部失败
	if len(m.externalURLs) == 0 {
		return nil, fmt.Errorf("failed to %s: no nodes configured", operation)
	}

	startIndex := int(atomic.AddInt64(&m.currentIndex, 1)) % len(m.externalURLs)

	// 优先尝试上次成功节点
	if lg := int(atomic.LoadInt64(&m.lastGoodIndex)); lg >= 0 && lg < len(m.externalURLs) {
		n := m.nodeStates[lg]
		if !(n.status != BTCNodeHealthy && time.Now().Before(n.restUntil)) {
			if n.status != BTCNodeHealthy && time.Now().After(n.restUntil) {
				m.resetNode(lg)
			}
			if data, err := cb(m.externalURLs[lg]); err == nil {
				return data, nil
			}
			// 失败则继续常规轮询
		}
	}
	for i := 0; i < len(m.externalURLs); i++ {
		idx := (startIndex + i) % len(m.externalURLs)
		url := m.externalURLs[idx]
		node := m.nodeStates[idx]

		// 检查休息
		now := time.Now()
		if node.status != BTCNodeHealthy && now.Before(node.restUntil) {
			continue
		}
		if node.status != BTCNodeHealthy && now.After(node.restUntil) {
			m.resetNode(idx)
		}

		if node.firstCallTime.IsZero() {
			node.firstCallTime = now
		}
		node.lastCallTime = now
		atomic.AddInt64(&node.totalRequests, 1)

		data, err := cb(url)
		if err == nil {
			atomic.StoreInt64(&m.lastGoodIndex, int64(idx))
			return data, nil
		}

		// 根据错误设置休息策略
		if m.isRateLimit(err) {
			node.status = BTCNodeOverheat
			used := node.lastCallTime.Sub(node.firstCallTime)
			rest := time.Second - used
			if rest < 0 {
				rest = time.Millisecond * 10
			}
			node.restUntil = now.Add(rest)
		} else {
			node.status = BTCNodeDamaged
			node.restUntil = now.Add(60 * time.Second)
		}
	}

	return nil, fmt.Errorf("failed to %s: all nodes failed after %v", operation, time.Since(start))
}

// 便捷方法族
func (m *BTCFailoverManager) CallWithFailover(operation string, cb func(baseURL string) error) error {
	_, err := m.Execute(operation, func(baseURL string) (interface{}, error) {
		return nil, cb(baseURL)
	})
	return err
}

func (m *BTCFailoverManager) CallWithFailoverUint64(operation string, cb func(baseURL string) (uint64, error)) (uint64, error) {
	res, err := m.Execute(operation, func(baseURL string) (interface{}, error) {
		return cb(baseURL)
	})
	if err != nil {
		return 0, err
	}
	if v, ok := res.(uint64); ok {
		return v, nil
	}
	return 0, fmt.Errorf("unexpected result type")
}

func (m *BTCFailoverManager) CallWithFailoverString(operation string, cb func(baseURL string) (string, error)) (string, error) {
	res, err := m.Execute(operation, func(baseURL string) (interface{}, error) {
		return cb(baseURL)
	})
	if err != nil {
		return "", err
	}
	if v, ok := res.(string); ok {
		return v, nil
	}
	return "", fmt.Errorf("unexpected result type")
}

func (m *BTCFailoverManager) CallWithFailoverMap(operation string, cb func(baseURL string) (map[string]interface{}, error)) (map[string]interface{}, error) {
	res, err := m.Execute(operation, func(baseURL string) (interface{}, error) {
		return cb(baseURL)
	})
	if err != nil {
		return nil, err
	}
	if v, ok := res.(map[string]interface{}); ok {
		return v, nil
	}
	return nil, fmt.Errorf("unexpected result type")
}

func (m *BTCFailoverManager) CallWithFailoverMaps(operation string, cb func(baseURL string) ([]map[string]interface{}, error)) ([]map[string]interface{}, error) {
	res, err := m.Execute(operation, func(baseURL string) (interface{}, error) {
		return cb(baseURL)
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
func (m *BTCFailoverManager) resetNode(idx int) {
	n := m.nodeStates[idx]
	n.status = BTCNodeHealthy
	n.firstCallTime = time.Time{}
	n.lastCallTime = time.Time{}
	n.restUntil = time.Time{}
}

func (m *BTCFailoverManager) isRateLimit(err error) bool {
	if err == nil {
		return false
	}
	es := err.Error()
	return strings.Contains(es, "429") || strings.Contains(es, "rate limit")
}
