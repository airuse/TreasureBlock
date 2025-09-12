package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"blockChainBrowser/server/config"
)

// BTCFailoverManager BTC API故障转移管理器
type BTCFailoverManager struct {
	rpcURLs     []string // JSON-RPC URLs (用于获取区块高度)
	restURLs    []string // REST API URLs (用于获取地址余额)
	rpcCurrent  int64
	restCurrent int64
	timeout     time.Duration
}

// BTCBalanceResponse BTC余额响应
type BTCBalanceResponse struct {
	Balance string `json:"balance"` // satoshi
	Height  uint64 `json:"height"`  // 区块高度
}

// NewBTCFailoverFromChain 基于链名创建BTC故障转移管理器
func NewBTCFailoverFromChain(chainName string) (*BTCFailoverManager, error) {
	chainCfg, ok := config.AppConfig.Blockchain.Chains[strings.ToLower(chainName)]
	if !ok {
		return nil, fmt.Errorf("chain not configured: %s", chainName)
	}

	// 获取JSON-RPC URLs
	rpcURLs := make([]string, 0)
	if len(chainCfg.RPCURLs) > 0 {
		rpcURLs = append(rpcURLs, chainCfg.RPCURLs...)
	}
	if chainCfg.RPCURL != "" {
		rpcURLs = append(rpcURLs, chainCfg.RPCURL)
	}

	// 获取REST API URLs
	restURLs := make([]string, 0)
	if len(chainCfg.RESTURLs) > 0 {
		restURLs = append(restURLs, chainCfg.RESTURLs...)
	}

	if len(rpcURLs) == 0 {
		return nil, fmt.Errorf("no RPC URLs configured for chain: %s", chainName)
	}

	if len(restURLs) == 0 {
		return nil, fmt.Errorf("no REST URLs configured for chain: %s", chainName)
	}

	return &BTCFailoverManager{
		rpcURLs:  rpcURLs,
		restURLs: restURLs,
		timeout:  20 * time.Second,
	}, nil
}

// nextRPC 轮询获取下一个RPC URL索引
func (m *BTCFailoverManager) nextRPC() string {
	if len(m.rpcURLs) == 1 {
		return m.rpcURLs[0]
	}
	idx := int(atomic.AddInt64(&m.rpcCurrent, 1)) % len(m.rpcURLs)
	return m.rpcURLs[idx]
}

// nextREST 轮询获取下一个REST URL索引
func (m *BTCFailoverManager) nextREST() string {
	if len(m.restURLs) == 1 {
		return m.restURLs[0]
	}
	idx := int(atomic.AddInt64(&m.restCurrent, 1)) % len(m.restURLs)
	return m.restURLs[idx]
}

// GetBalance 故障转移获取BTC地址余额
func (m *BTCFailoverManager) GetBalance(ctx context.Context, address string) (*BTCBalanceResponse, error) {
	return m.GetBalanceAtHeight(ctx, address, 0) // 0表示获取最新余额
}

// GetBalanceAtHeight 故障转移获取指定高度下BTC地址余额
func (m *BTCFailoverManager) GetBalanceAtHeight(ctx context.Context, address string, height uint64) (*BTCBalanceResponse, error) {
	var lastErr error
	deadline := time.Now().Add(m.timeout)

	for time.Now().Before(deadline) {
		url := m.nextREST()
		fmt.Printf("从REST API %s获取余额\n", url)
		balance, err := m.getBalanceFromRESTAPI(ctx, url, address, height)
		if err == nil {
			return balance, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("所有REST API都获取失败: %w", lastErr)
}

// GetLatestBlockHeight 故障转移获取最新区块高度
func (m *BTCFailoverManager) GetLatestBlockHeight(ctx context.Context) (uint64, error) {
	var lastErr error
	deadline := time.Now().Add(m.timeout)

	for time.Now().Before(deadline) {
		url := m.nextRPC()
		height, err := m.getLatestBlockHeightFromURL(ctx, url)
		if err == nil {
			return height, nil
		}
		lastErr = err
	}

	return 0, fmt.Errorf("所有JSON-RPC都获取最新高度失败: %w", lastErr)
}

// getLatestBlockHeightFromURL 从指定URL获取最新区块高度
func (m *BTCFailoverManager) getLatestBlockHeightFromURL(ctx context.Context, baseURL string) (uint64, error) {
	// 使用JSON-RPC格式
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "getblockcount",
		"params":  []interface{}{},
		"id":      1,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL, strings.NewReader(string(jsonData)))
	if err != nil {
		fmt.Printf("发送请求失败: %v\n", err)
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: m.timeout}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("读取响应体失败: %v\n", err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API返回错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// 解析JSON-RPC响应
	var rpcResp struct {
		Result uint64 `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return 0, err
	}

	if rpcResp.Error != nil {
		return 0, fmt.Errorf("RPC错误: %s", rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// getBalanceFromRESTAPI 从REST API获取余额
func (m *BTCFailoverManager) getBalanceFromRESTAPI(ctx context.Context, baseURL, address string, height uint64) (*BTCBalanceResponse, error) {
	// 构建REST API URL
	url := fmt.Sprintf("%s/address/%s", baseURL, address)

	fmt.Printf("使用REST API获取余额: %s\n", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: m.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("REST API返回错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析REST API响应
	var data struct {
		ChainStats struct {
			FundedTxoSum int64 `json:"funded_txo_sum"`
			SpentTxoSum  int64 `json:"spent_txo_sum"`
		} `json:"chain_stats"`
		MempoolStats struct {
			FundedTxoSum int64 `json:"funded_txo_sum"`
			SpentTxoSum  int64 `json:"spent_txo_sum"`
		} `json:"mempool_stats"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	// 计算余额（已确认 + 未确认）
	balance := data.ChainStats.FundedTxoSum - data.ChainStats.SpentTxoSum +
		data.MempoolStats.FundedTxoSum - data.MempoolStats.SpentTxoSum

	balanceStr := strconv.FormatInt(balance, 10)
	fmt.Printf("获取到的余额: %s satoshi\n", balanceStr)

	return &BTCBalanceResponse{
		Balance: balanceStr,
		Height:  height,
	}, nil
}

// GetBlockHash 故障转移获取区块哈希
func (m *BTCFailoverManager) GetBlockHash(ctx context.Context, blockNumber uint64) (string, error) {
	var lastErr error
	deadline := time.Now().Add(m.timeout)

	for time.Now().Before(deadline) {
		url := m.nextRPC()
		hash, err := m.getBlockHashFromURL(ctx, url, blockNumber)
		if err == nil {
			return hash, nil
		}
		lastErr = err
	}

	return "", fmt.Errorf("所有JSON-RPC都获取区块哈希失败: %w", lastErr)
}

// GetBlock 故障转移获取区块详情
func (m *BTCFailoverManager) GetBlock(ctx context.Context, blockHash string) (map[string]interface{}, error) {
	var lastErr error
	deadline := time.Now().Add(m.timeout)

	for time.Now().Before(deadline) {
		url := m.nextRPC()
		block, err := m.getBlockFromURL(ctx, url, blockHash)
		if err == nil {
			return block, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("所有JSON-RPC都获取区块详情失败: %w", lastErr)
}

// getBlockHashFromURL 从指定URL获取区块哈希
func (m *BTCFailoverManager) getBlockHashFromURL(ctx context.Context, baseURL string, blockNumber uint64) (string, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "getblockhash",
		"params":  []interface{}{blockNumber},
		"id":      1,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: m.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API返回错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析JSON-RPC响应
	var rpcResp struct {
		Result string `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return "", err
	}

	if rpcResp.Error != nil {
		return "", fmt.Errorf("RPC错误: %s", rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// getBlockFromURL 从指定URL获取区块详情
func (m *BTCFailoverManager) getBlockFromURL(ctx context.Context, baseURL string, blockHash string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "getblock",
		"params":  []interface{}{blockHash, 2}, // 2表示返回详细交易信息
		"id":      1,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: m.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析JSON-RPC响应
	var rpcResp struct {
		Result map[string]interface{} `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, err
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC错误: %s", rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// BTCMempoolTransaction Mempool中的BTC交易结构
type BTCMempoolTransaction struct {
	TxID    string
	Fee     int64   // 交易费 (satoshi)
	Size    int     // 交易大小 (bytes)
	FeeRate float64 // 费率 (satoshi per vbyte)
	Time    int64   // 进入Mempool的时间戳
}

// GetMempoolTransactions 故障转移获取Mempool中的交易
func (m *BTCFailoverManager) GetMempoolTransactions(ctx context.Context) ([]BTCMempoolTransaction, error) {
	var lastErr error
	deadline := time.Now().Add(m.timeout)

	for time.Now().Before(deadline) {
		url := m.nextRPC()
		txs, err := m.getMempoolTransactionsFromURL(ctx, url)
		if err == nil {
			return txs, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("所有JSON-RPC都获取Mempool交易失败: %w", lastErr)
}

// getMempoolTransactionsFromURL 从指定URL获取Mempool交易
func (m *BTCFailoverManager) getMempoolTransactionsFromURL(ctx context.Context, baseURL string) ([]BTCMempoolTransaction, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "getrawmempool",
		"params":  []interface{}{true}, // true表示返回详细交易信息
		"id":      1,
	}
	fmt.Printf("从JSON-RPC %s获取Mempool交易\n", baseURL)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL, strings.NewReader(string(jsonData)))
	if err != nil {
		fmt.Printf("发送请求失败: %v\n", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: m.timeout}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("读取响应体失败: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应体失败: %v\n", err)
		return nil, err
	}

	// 解析JSON-RPC响应
	var rpcResp struct {
		Result map[string]interface{} `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &rpcResp); err != nil {
		fmt.Printf("解析JSON-RPC响应失败: %v\n", err)
		return nil, err
	}

	if rpcResp.Error != nil {
		fmt.Printf("RPC错误: %s\n", rpcResp.Error.Message)
		return nil, fmt.Errorf("RPC错误: %s", rpcResp.Error.Message)
	}

	// 解析Mempool交易数据（优先使用 vsize 与 fees.base(BTC) 计算 sat/vB）
	var transactions []BTCMempoolTransaction
	for txID, txData := range rpcResp.Result {
		tx, ok := txData.(map[string]interface{})
		if !ok {
			fmt.Printf("交易数据格式错误: %v\n", txData)
			continue
		}

		// vsize（虚拟大小）
		vsize, _ := tx["vsize"].(float64)
		if vsize == 0 {
			// 兼容 size 字段
			vsize, _ = tx["size"].(float64)
		}

		// 费用在 fees.base，单位 BTC
		var feeBTC float64
		if feesAny, ok := tx["fees"]; ok {
			if feesMap, ok := feesAny.(map[string]interface{}); ok {
				if base, ok := feesMap["base"].(float64); ok {
					feeBTC = base
				}
			}
		}

		// 时间戳
		t, _ := tx["time"].(float64)

		// 计算费率：sat/vB
		var feeRate float64
		var feeSatoshi int64
		if feeBTC > 0 {
			feeSatoshi = int64(feeBTC * 1e8)
		}
		if vsize > 0 {
			feeRate = math.Round((float64(feeSatoshi)/vsize)*10) / 10
		}

		transactions = append(transactions, BTCMempoolTransaction{
			TxID:    txID,
			Fee:     feeSatoshi,
			Size:    int(vsize),
			FeeRate: feeRate,
			Time:    int64(t),
		})
	}

	return transactions, nil
}
