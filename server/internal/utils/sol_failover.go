package utils

import (
	"blockChainBrowser/server/config"
	"context"
	"fmt"
	"sync/atomic"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/sirupsen/logrus"
)

// SolFailoverManager 负责对多个 Solana RPC 节点做简单故障转移/轮询
type SolFailoverManager struct {
	clients []*client.Client
	idx     uint32
	logger  *logrus.Logger
}

// NewSolFailoverFromChain 根据链名从配置创建 Sol 故障转移管理器
func NewSolFailoverFromChain(chainName string) (*SolFailoverManager, error) {
	chainCfg, ok := config.AppConfig.Blockchain.Chains[chainName]
	if !ok || !chainCfg.Enabled {
		return nil, fmt.Errorf("chain %s not enabled", chainName)
	}
	if len(chainCfg.RPCURLs) == 0 {
		return nil, fmt.Errorf("chain %s has no rpc_urls configured", chainName)
	}

	// 创建多个 RPC 客户端
	clients := make([]*client.Client, 0, len(chainCfg.RPCURLs))
	for _, endpoint := range chainCfg.RPCURLs {
		client := client.NewClient(endpoint)
		clients = append(clients, client)
	}

	return &SolFailoverManager{
		clients: clients,
		logger:  logrus.New(),
	}, nil
}

func (m *SolFailoverManager) nextClient() *client.Client {
	if len(m.clients) == 0 {
		return nil
	}
	i := atomic.AddUint32(&m.idx, 1)
	return m.clients[int(i)%len(m.clients)]
}

// callWithFailover 使用故障转移机制调用 RPC 方法
func (m *SolFailoverManager) callWithFailover(ctx context.Context, fn func(*client.Client) error) error {
	if len(m.clients) == 0 {
		return fmt.Errorf("no sol clients configured")
	}

	var lastErr error
	for range m.clients {
		client := m.nextClient()
		if client == nil {
			continue
		}

		if err := fn(client); err != nil {
			m.logger.WithError(err).Debug("RPC call failed, trying next client")
			lastErr = err
			continue
		}
		return nil
	}
	return lastErr
}

// GetSlot 获取最新 slot（用作"区块号"参考）
func (m *SolFailoverManager) GetSlot(ctx context.Context) (uint64, error) {
	var slot uint64
	err := m.callWithFailover(ctx, func(client *client.Client) error {
		var err error
		slot, err = client.GetSlot(ctx)
		return err
	})
	return slot, err
}

// PrioritizationFeeItem 优先费结构
type PrioritizationFeeItem struct {
	Slot              uint64 `json:"slot"`
	PrioritizationFee uint64 `json:"prioritizationFee"`
}

// GetRecentPrioritizationFees 返回最近一批优先费数据
func (m *SolFailoverManager) GetRecentPrioritizationFees(ctx context.Context) ([]PrioritizationFeeItem, error) {
	var fees []PrioritizationFeeItem
	err := m.callWithFailover(ctx, func(client *client.Client) error {
		// 获取最近的优先费用数据（传入空地址列表获取所有）
		recentFees, err := client.GetRecentPrioritizationFees(ctx, []common.PublicKey{})
		if err != nil {
			return err
		}

		// 转换为我们的结构
		fees = make([]PrioritizationFeeItem, len(recentFees))
		for i, fee := range recentFees {
			fees[i] = PrioritizationFeeItem{
				Slot:              fee.Slot,
				PrioritizationFee: fee.PrioritizationFee,
			}
		}
		return nil
	})
	return fees, err
}

// SendRawTransaction 发送原始交易
func (m *SolFailoverManager) SendRawTransaction(ctx context.Context, rawTx string) (string, error) {
	// TODO: 实现从 base58 字符串解析并发送交易
	// 这需要将 base58 字符串转换为 types.Transaction 对象
	return "", fmt.Errorf("SendRawTransaction not implemented yet")
}

// Close 关闭所有客户端连接
func (m *SolFailoverManager) Close() {
	// Solana Go SDK 的 RPC 客户端通常不需要显式关闭
	// 但我们可以在这里添加清理逻辑
}
