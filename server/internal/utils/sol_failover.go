package utils

import (
	"blockChainBrowser/server/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/sirupsen/logrus"
)

// SolFailoverManager 负责对多个 Solana RPC 节点做简单故障转移/轮询
type SolFailoverManager struct {
	clients   []*client.Client
	idx       uint32
	logger    *logrus.Logger
	endpoints []string
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
		clients:   clients,
		logger:    logrus.New(),
		endpoints: chainCfg.RPCURLs,
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

// SendRawTransaction 发送原始交易（rawTx 为 base64 序列化字节串）
func (m *SolFailoverManager) SendRawTransaction(ctx context.Context, rawTx string) (string, error) {
	if len(m.endpoints) == 0 {
		return "", fmt.Errorf("no sol endpoints configured")
	}
	var lastErr error
	for i := 0; i < len(m.endpoints); i++ {
		endpoint := m.endpoints[(int(m.idx)+i)%len(m.endpoints)]
		txHash, err := m.sendRawTransactionOnce(ctx, endpoint, rawTx)
		if err == nil {
			return txHash, nil
		}
		lastErr = err
		m.logger.WithError(err).Warnf("sendRawTransaction failed on %s, trying next", endpoint)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("unknown error sending sol tx")
	}
	return "", lastErr
}

func (m *SolFailoverManager) sendRawTransactionOnce(ctx context.Context, endpoint string, rawBase64 string) (string, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "sendTransaction",
		"params":  []interface{}{rawBase64, map[string]interface{}{"encoding": "base64"}},
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var out struct {
		Result string `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &out); err != nil {
		return "", fmt.Errorf("decode response failed: %w, body=%s", err, string(respBytes))
	}
	if out.Error != nil {
		return "", fmt.Errorf("rpc error %d: %s", out.Error.Code, out.Error.Message)
	}
	if out.Result == "" {
		return "", fmt.Errorf("empty tx hash: %s", string(respBytes))
	}
	return out.Result, nil
}

// GetLatestBlockhash 获取最近区块哈希（用于构建未签名交易）
func (m *SolFailoverManager) GetLatestBlockhash(ctx context.Context) (string, error) {
	if len(m.endpoints) == 0 {
		return "", fmt.Errorf("no sol endpoints configured")
	}
	var lastErr error
	for i := 0; i < len(m.endpoints); i++ {
		endpoint := m.endpoints[(int(m.idx)+i)%len(m.endpoints)]
		bh, err := m.getLatestBlockhashOnce(ctx, endpoint)
		if err == nil {
			return bh, nil
		}
		lastErr = err
		m.logger.WithError(err).Warnf("getLatestBlockhash failed on %s, trying next", endpoint)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("unknown error getLatestBlockhash")
	}
	return "", lastErr
}

func (m *SolFailoverManager) getLatestBlockhashOnce(ctx context.Context, endpoint string) (string, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getLatestBlockhash",
		"params":  []interface{}{},
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var out struct {
		Result struct {
			Value struct {
				Blockhash string `json:"blockhash"`
			} `json:"value"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &out); err != nil {
		return "", fmt.Errorf("decode response failed: %w, body=%s", err, string(respBytes))
	}
	if out.Error != nil {
		return "", fmt.Errorf("rpc error %d: %s", out.Error.Code, out.Error.Message)
	}
	if out.Result.Value.Blockhash == "" {
		return "", fmt.Errorf("empty blockhash: %s", string(respBytes))
	}
	return out.Result.Value.Blockhash, nil
}

// GetAccountBalance 获取账户余额（返回 context.slot 与 lamports）
func (m *SolFailoverManager) GetAccountBalance(ctx context.Context, address string) (uint64, uint64, error) {
	if len(m.endpoints) == 0 {
		return 0, 0, fmt.Errorf("no sol endpoints configured")
	}
	type rpcReq struct {
		JSONRPC string        `json:"jsonrpc"`
		ID      int           `json:"id"`
		Method  string        `json:"method"`
		Params  []interface{} `json:"params"`
	}
	type rpcResp struct {
		Result struct {
			Context struct {
				Slot uint64 `json:"slot"`
			} `json:"context"`
			Value uint64 `json:"value"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	var lastErr error
	for i := 0; i < len(m.endpoints); i++ {
		endpoint := m.endpoints[(int(m.idx)+i)%len(m.endpoints)]
		payload := rpcReq{JSONRPC: "2.0", ID: 1, Method: "getBalance", Params: []interface{}{address}}
		body, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		respBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		var out rpcResp
		if err := json.Unmarshal(respBytes, &out); err != nil {
			lastErr = fmt.Errorf("decode response failed: %w, body=%s", err, string(respBytes))
			continue
		}
		if out.Error != nil {
			lastErr = fmt.Errorf("rpc error %d: %s", out.Error.Code, out.Error.Message)
			m.logger.WithError(lastErr).Warnf("getBalance failed on %s, trying next", endpoint)
			continue
		}
		return out.Result.Context.Slot, out.Result.Value, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("unknown error getBalance")
	}
	return 0, 0, lastErr
}

// GetTokenAccountsByOwner 获取指定所有者的Token账户列表
func (m *SolFailoverManager) GetTokenAccountsByOwner(ctx context.Context, owner string, mint *string) ([]TokenAccountInfo, error) {
	if len(m.endpoints) == 0 {
		return nil, fmt.Errorf("no sol endpoints configured")
	}

	type rpcReq struct {
		JSONRPC string        `json:"jsonrpc"`
		ID      int           `json:"id"`
		Method  string        `json:"method"`
		Params  []interface{} `json:"params"`
	}

	type rpcResp struct {
		Result struct {
			Context struct {
				Slot uint64 `json:"slot"`
			} `json:"context"`
			Value []struct {
				Account struct {
					Data struct {
						Parsed struct {
							Info struct {
								TokenAmount struct {
									Amount         string   `json:"amount"`
									Decimals       int      `json:"decimals"`
									UIAmount       *float64 `json:"uiAmount"`
									UIAmountString string   `json:"uiAmountString"`
								} `json:"tokenAmount"`
								Mint  string `json:"mint"`
								Owner string `json:"owner"`
							} `json:"info"`
							Type string `json:"type"`
						} `json:"parsed"`
					} `json:"data"`
					Executable bool   `json:"executable"`
					Lamports   uint64 `json:"lamports"`
					Owner      string `json:"owner"`
					RentEpoch  uint64 `json:"rentEpoch"`
				} `json:"account"`
				Pubkey string `json:"pubkey"`
			} `json:"value"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	var lastErr error
	for i := 0; i < len(m.endpoints); i++ {
		endpoint := m.endpoints[(int(m.idx)+i)%len(m.endpoints)]

		// 构建参数
		var params []interface{}

		if mint != nil {
			// 当指定mint时，使用mint过滤
			params = []interface{}{
				owner,
				map[string]interface{}{
					"mint": *mint,
				},
				map[string]interface{}{
					"encoding": "jsonParsed",
				},
			}
		} else {
			// 当不指定mint时，使用programId过滤
			params = []interface{}{
				owner,
				map[string]interface{}{
					"programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
				},
				map[string]interface{}{
					"encoding": "jsonParsed",
				},
			}
		}

		payload := rpcReq{JSONRPC: "2.0", ID: 1, Method: "getTokenAccountsByOwner", Params: params}
		body, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		respBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		var out rpcResp
		if err := json.Unmarshal(respBytes, &out); err != nil {
			lastErr = fmt.Errorf("decode response failed: %w, body=%s", err, string(respBytes))
			continue
		}
		if out.Error != nil {
			lastErr = fmt.Errorf("rpc error %d: %s", out.Error.Code, out.Error.Message)
			m.logger.WithError(lastErr).Warnf("getTokenAccountsByOwner failed on %s, trying next", endpoint)
			continue
		}

		// 转换结果
		var tokenAccounts []TokenAccountInfo
		fmt.Printf("RPC返回的账户数量: %d\n", len(out.Result.Value))
		for i, account := range out.Result.Value {
			fmt.Printf("账户 %d: Pubkey=%s, Type=%s\n", i, account.Pubkey, account.Account.Data.Parsed.Type)
			if account.Account.Data.Parsed.Type == "account" {
				fmt.Printf("  - Mint: %s\n", account.Account.Data.Parsed.Info.Mint)
				fmt.Printf("  - Owner: %s\n", account.Account.Data.Parsed.Info.Owner)
				fmt.Printf("  - Amount: %s\n", account.Account.Data.Parsed.Info.TokenAmount.Amount)

				tokenAccounts = append(tokenAccounts, TokenAccountInfo{
					Address:  account.Pubkey,
					Mint:     account.Account.Data.Parsed.Info.Mint,
					Owner:    account.Account.Data.Parsed.Info.Owner,
					Amount:   account.Account.Data.Parsed.Info.TokenAmount.Amount,
					Decimals: account.Account.Data.Parsed.Info.TokenAmount.Decimals,
					UIAmount: account.Account.Data.Parsed.Info.TokenAmount.UIAmount,
				})
			}
		}
		fmt.Printf("最终返回的Token账户数量: %d\n", len(tokenAccounts))
		return tokenAccounts, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("unknown error getTokenAccountsByOwner")
	}
	return nil, lastErr
}

// TokenAccountInfo Token账户信息
type TokenAccountInfo struct {
	Address  string   `json:"address"`
	Mint     string   `json:"mint"`
	Owner    string   `json:"owner"`
	Amount   string   `json:"amount"`
	Decimals int      `json:"decimals"`
	UIAmount *float64 `json:"uiAmount,omitempty"`
}

// Close 关闭所有客户端连接
func (m *SolFailoverManager) Close() {
	// Solana Go SDK 的 RPC 客户端通常不需要显式关闭
	// 但我们可以在这里添加清理逻辑
}
