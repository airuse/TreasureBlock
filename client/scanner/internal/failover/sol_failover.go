package failover

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/sirupsen/logrus"
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

// SOLFailoverManager Solana 故障转移调度器（基于 blocto/solana-go-sdk）
type SOLFailoverManager struct {
	mainClient      *client.Client
	failoverClients []*client.Client
	failoverStates  []*SOLNodeState // 只存储故障转移节点的状态
	currentIndex    int64
}

// NewSOLFailoverManager 创建管理器
func NewSOLFailoverManager(mainClient *client.Client, failoverClients []*client.Client) *SOLFailoverManager {
	// 只为故障转移节点创建状态
	failoverStates := make([]*SOLNodeState, len(failoverClients))
	for i := range failoverStates {
		failoverStates[i] = &SOLNodeState{status: SOLNodeHealthy}
	}

	return &SOLFailoverManager{
		mainClient:      mainClient,
		failoverClients: failoverClients,
		failoverStates:  failoverStates,
	}
}

// Execute 执行带故障转移的请求
func (m *SOLFailoverManager) Execute(ctx context.Context, operation string, cb func(*client.Client) (interface{}, error)) (interface{}, error) {
	start := time.Now()

	// 首先尝试主客户端
	if m.mainClient != nil {
		if data, err := cb(m.mainClient); err == nil {
			logrus.Debugf("[sol] Main client %s succeeded", operation)
			return data, nil
		} else {
			logrus.Warnf("[sol] Main client %s failed: %v", operation, err)
		}
	}

	// 主客户端失败后，轮询故障转移客户端
	if len(m.failoverClients) == 0 {
		return nil, fmt.Errorf("failed to %s: no failover clients configured", operation)
	}

	startIndex := int(atomic.AddInt64(&m.currentIndex, 1)) % len(m.failoverClients)
	for i := 0; i < len(m.failoverClients); i++ {
		idx := (startIndex + i) % len(m.failoverClients)
		client := m.failoverClients[idx]
		node := m.failoverStates[idx] // 直接使用故障转移节点状态

		// 检查休息状态
		now := time.Now()
		if node.status != SOLNodeHealthy && now.Before(node.restUntil) {
			logrus.Debugf("[sol] Failover client %d is resting until %v", idx, node.restUntil)
			continue
		}
		if node.status != SOLNodeHealthy && now.After(node.restUntil) {
			m.resetFailoverNode(idx)
		}

		if node.firstCallTime.IsZero() {
			node.firstCallTime = now
		}
		node.lastCallTime = now
		atomic.AddInt64(&node.totalRequests, 1)

		data, err := cb(client)
		if err == nil {
			logrus.Infof("[sol] Failover client %d %s succeeded", idx, operation)
			return data, nil
		}

		// 根据错误设置休息策略
		if m.isRateLimit(err) {
			logrus.Warnf("[sol] Failover client %d %s rate limited", idx, operation)
			node.status = SOLNodeOverheat
			used := node.lastCallTime.Sub(node.firstCallTime)
			rest := time.Second - used
			if rest < 0 {
				rest = time.Millisecond * 10000
			}
			node.restUntil = now.Add(rest)
		} else {
			logrus.Warnf("[sol] Failover client %d %s failed: %v", idx, operation, err)
			node.status = SOLNodeDamaged
			node.restUntil = now.Add(10 * time.Second)
		}
	}

	return nil, fmt.Errorf("failed to %s: all clients failed after %v", operation, time.Since(start))
}

// GetSlot 获取最新slot
func (m *SOLFailoverManager) GetSlot(ctx context.Context) (uint64, error) {
	res, err := m.Execute(ctx, "GetSlot", func(c *client.Client) (interface{}, error) {
		return c.GetSlot(ctx)
	})
	if err != nil {
		return 0, err
	}
	if slot, ok := res.(uint64); ok {
		return slot, nil
	}
	return 0, fmt.Errorf("unexpected result type for GetSlot")
}

// GetBlock 获取区块
func (m *SOLFailoverManager) GetBlock(ctx context.Context, slot uint64) (*client.Block, error) {
	res, err := m.Execute(ctx, "GetBlock", func(c *client.Client) (interface{}, error) {
		// 使用自定义的RPC配置来支持jsonParsed编码和版本化交易
		return c.RpcClient.GetBlockWithConfig(ctx, slot, rpc.GetBlockConfig{
			Encoding:                       rpc.GetBlockConfigEncodingJsonParsed,
			Commitment:                     rpc.CommitmentFinalized,
			MaxSupportedTransactionVersion: &[]uint8{0}[0], // 支持版本化交易
		})
	})
	if err != nil {
		return nil, err
	}

	// 调试输出

	// 手动转换RPC响应为Block结构
	if response, ok := res.(rpc.JsonRpcResponse[*rpc.GetBlock]); ok && response.Result != nil {
		// 由于jsonParsed格式与SDK的Block结构不完全兼容，
		// 我们需要创建一个简化的Block结构来适配
		var blockTime *time.Time
		if response.Result.BlockTime != nil {
			t := time.Unix(*response.Result.BlockTime, 0)
			blockTime = &t
		}

		block := &client.Block{
			Blockhash:         response.Result.Blockhash,
			BlockTime:         blockTime,
			BlockHeight:       response.Result.BlockHeight,
			PreviousBlockhash: response.Result.PreviousBlockhash,
			ParentSlot:        response.Result.ParentSlot,
			Signatures:        response.Result.Signatures,
			Rewards:           convertRewards(response.Result.Rewards),
		}

		// 转换交易数据
		transactions := make([]client.BlockTransaction, 0, len(response.Result.Transactions))
		for _, rpcTx := range response.Result.Transactions {
			// 创建简化的交易结构
			tx := client.BlockTransaction{
				Meta: convertTransactionMeta(rpcTx.Meta),
			}

			// 从jsonParsed数据中提取账户密钥
			if txData, ok := rpcTx.Transaction.(map[string]interface{}); ok {
				if message, ok := txData["message"].(map[string]interface{}); ok {
					if accountKeys, ok := message["accountKeys"].([]interface{}); ok {
						keys := make([]common.PublicKey, 0, len(accountKeys))
						for _, keyData := range accountKeys {
							if keyMap, ok := keyData.(map[string]interface{}); ok {
								if pubkey, ok := keyMap["pubkey"].(string); ok {
									keys = append(keys, common.PublicKeyFromString(pubkey))
								}
							}
						}
						tx.AccountKeys = keys
					}
				}
			}

			transactions = append(transactions, tx)
		}
		block.Transactions = transactions

		return block, nil
	}
	return nil, fmt.Errorf("unexpected result type for GetBlock")
}

// GetBlockRaw 获取原始区块数据（用于jsonParsed格式的深度解析）
func (m *SOLFailoverManager) GetBlockRaw(ctx context.Context, slot uint64) (map[string]interface{}, error) {
	res, err := m.Execute(ctx, "GetBlockRaw", func(c *client.Client) (interface{}, error) {
		// 直接调用RPC客户端获取原始响应
		return c.RpcClient.GetBlockWithConfig(ctx, slot, rpc.GetBlockConfig{
			Encoding:                       rpc.GetBlockConfigEncodingJsonParsed,
			Commitment:                     rpc.CommitmentFinalized,
			MaxSupportedTransactionVersion: &[]uint8{0}[0], // 支持版本化交易
		})
	})
	if err != nil {
		return nil, err
	}

	// 将响应转换为map以便进一步处理
	if response, ok := res.(rpc.JsonRpcResponse[*rpc.GetBlock]); ok && response.Result != nil {
		// 将RPC响应转换为map
		jsonData, err := json.Marshal(response.Result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal block data: %v", err)
		}

		var blockMap map[string]interface{}
		if err := json.Unmarshal(jsonData, &blockMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal block data: %v", err)
		}

		return blockMap, nil
	}
	return nil, fmt.Errorf("unexpected result type for GetBlockRaw")
}

// convertRewards 转换奖励数据
func convertRewards(rewards []rpc.Reward) []client.Reward {
	if len(rewards) == 0 {
		return nil
	}

	result := make([]client.Reward, 0, len(rewards))
	for _, r := range rewards {
		result = append(result, client.Reward{
			Pubkey:     common.PublicKeyFromString(r.Pubkey),
			Lamports:   r.Lamports,
			RewardType: r.RewardType,
			Commission: r.Commission,
		})
	}
	return result
}

// convertTransactionMeta 转换交易元数据（简化版本）
func convertTransactionMeta(meta *rpc.TransactionMeta) *client.TransactionMeta {
	if meta == nil {
		return nil
	}

	// 简化转换，只转换基本字段
	return &client.TransactionMeta{
		Err:                  meta.Err,
		Fee:                  meta.Fee,
		PreBalances:          meta.PreBalances,
		PostBalances:         meta.PostBalances,
		LogMessages:          meta.LogMessages,
		ComputeUnitsConsumed: meta.ComputeUnitsConsumed,
	}
}

// 私有工具方法
func (m *SOLFailoverManager) resetFailoverNode(idx int) {
	if idx >= len(m.failoverStates) {
		return
	}
	n := m.failoverStates[idx]
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
		strings.Contains(es, "request limit") ||
		strings.Contains(es, "throttled")
}

// GetStats 获取统计信息
func (m *SOLFailoverManager) GetStats() map[string]interface{} {
	healthyFailoverNodes := 0
	for _, node := range m.failoverStates {
		if node.status == SOLNodeHealthy {
			healthyFailoverNodes++
		}
	}

	return map[string]interface{}{
		"main_client":            m.mainClient != nil,
		"failover_clients":       len(m.failoverClients),
		"healthy_failover_nodes": healthyFailoverNodes,
		"total_failover_nodes":   len(m.failoverStates),
	}
}
