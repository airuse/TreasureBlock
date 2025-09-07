package utils

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"blockChainBrowser/server/config"

	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthFailoverManager 极简以太坊RPC故障转移管理器
type EthFailoverManager struct {
	clients []*ethclient.Client
	current int64
}

// NewEthFailoverFromChain 基于链名创建故障转移管理器（读取 config.Blockchain.Chains）
func NewEthFailoverFromChain(chainName string) (*EthFailoverManager, error) {
	chainCfg, ok := config.AppConfig.Blockchain.Chains[strings.ToLower(chainName)]
	if !ok {
		return nil, fmt.Errorf("chain not configured: %s", chainName)
	}

	urls := make([]string, 0)
	if len(chainCfg.RPCURLs) > 0 {
		urls = append(urls, chainCfg.RPCURLs...)
	}
	if chainCfg.RPCURL != "" {
		urls = append(urls, chainCfg.RPCURL)
	}
	if len(urls) == 0 {
		return nil, fmt.Errorf("no RPC URLs configured for chain: %s", chainName)
	}

	clients := make([]*ethclient.Client, 0, len(urls))
	for _, u := range urls {
		cli, err := ethclient.Dial(u)
		if err == nil {
			clients = append(clients, cli)
		}
	}
	if len(clients) == 0 {
		return nil, fmt.Errorf("failed to connect any ETH RPC for chain: %s", chainName)
	}
	return &EthFailoverManager{clients: clients}, nil
}

// Close 关闭所有连接
func (m *EthFailoverManager) Close() {
	for _, c := range m.clients {
		c.Close()
	}
}

// next 轮询获取下一个客户端索引
func (m *EthFailoverManager) next() *ethclient.Client {
	if len(m.clients) == 1 {
		return m.clients[0]
	}
	idx := int(atomic.AddInt64(&m.current, 1)) % len(m.clients)
	return m.clients[idx]
}

// SendTransaction 故障转移发送交易
func (m *EthFailoverManager) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	fmt.Printf("🔷 开始发送交易: %s\n", tx.Hash().Hex())
	var lastErr error
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		cli := m.next()
		if err := cli.SendTransaction(ctx, tx); err == nil {
			fmt.Printf("✅ 交易发送成功: %s\n", tx.Hash().Hex())
			return nil
		} else {
			lastErr = err
			// 检查是否是余额不足错误
			if strings.Contains(err.Error(), "insufficient funds") {
				fmt.Printf("❌ 余额不足错误: %v\n", err)
				return fmt.Errorf("余额不足: %w", err)
			}
			fmt.Printf("⚠️ RPC调用失败: %v\n", err)
		}
	}
	fmt.Printf("🔷 发送交易失败,所有转移均不可用！: %v\n", lastErr)
	return lastErr
}

// TransactionByHash 故障转移查询交易
func (m *EthFailoverManager) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	var lastErr error
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		cli := m.next()
		tx, pending, err := cli.TransactionByHash(ctx, hash)
		if err == nil {
			return tx, pending, nil
		}
		lastErr = err
	}
	return nil, false, lastErr
}

// TransactionReceipt 故障转移查询收据
func (m *EthFailoverManager) TransactionReceipt(ctx context.Context, hash common.Hash) (*types.Receipt, error) {
	var lastErr error
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		cli := m.next()
		receipt, err := cli.TransactionReceipt(ctx, hash)
		if err == nil {
			return receipt, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// BlockNumber 故障转移查询最新区块
func (m *EthFailoverManager) BlockNumber(ctx context.Context) (uint64, error) {
	var lastErr error
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		cli := m.next()
		bn, err := cli.BlockNumber(ctx)
		if err == nil {
			return bn, nil
		}
		lastErr = err
	}
	return 0, lastErr
}

// BlockByHash 故障转移查询区块
func (m *EthFailoverManager) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	var lastErr error
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		cli := m.next()
		b, err := cli.BlockByHash(ctx, hash)
		if err == nil {
			return b, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// BlockByNumber 故障转移查询区块
func (m *EthFailoverManager) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	var lastErr error
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		cli := m.next()
		b, err := cli.BlockByNumber(ctx, number)
		if err == nil {
			return b, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// NonceAt 故障转移查询nonce
func (m *EthFailoverManager) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	var lastErr error
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		cli := m.next()
		n, err := cli.NonceAt(ctx, account, blockNumber)
		if err == nil {
			return n, nil
		}
		lastErr = err
	}
	return 0, lastErr
}

// EstimateGas 故障转移估算Gas
func (m *EthFailoverManager) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	var lastErr error
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		cli := m.next()
		gas, err := cli.EstimateGas(ctx, msg)
		if err == nil {
			return gas, nil
		}
		lastErr = err
	}
	return 0, lastErr
}

// BalanceAt 故障转移查询余额
func (m *EthFailoverManager) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	var lastErr error
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		cli := m.next()
		bal, err := cli.BalanceAt(ctx, account, blockNumber)
		if err == nil {
			return bal, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
