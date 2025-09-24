package utils

import (
	"blockChainBrowser/server/config"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"regexp"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// RPCClientManager RPC客户端管理器
type RPCClientManager struct {
	ethFailovers map[string]*EthFailoverManager // 链名 -> ETH故障转移管理器
	btcFailovers map[string]*BTCFailoverManager // 链名 -> BTC故障转移管理器
	btcClients   map[string]*BitcoinRPCClient   // 链名 -> BTC客户端（保留用于兼容性）
	solFailovers map[string]*SolFailoverManager // 链名 -> SOL故障转移管理器
	logger       *logrus.Logger
}

// BitcoinRPCClient 比特币RPC客户端
type BitcoinRPCClient struct {
	config     *config.ChainConfig
	httpClient *http.Client
	baseURL    string
	username   string
	password   string
}

// BitcoinRPCRequest 比特币RPC请求结构
type BitcoinRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// BitcoinRPCResponse 比特币RPC响应结构
type BitcoinRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// SendTransactionRequest 发送交易请求
type SendTransactionRequest struct {
	Chain       string `json:"chain"`        // 链类型 (btc, eth)
	SignedTx    string `json:"signed_tx"`    // 已签名的交易数据
	TxHash      string `json:"tx_hash"`      // 交易哈希（可选）
	FromAddress string `json:"from_address"` // 发送地址
	ToAddress   string `json:"to_address"`   // 接收地址
	Amount      string `json:"amount"`       // 金额
	Fee         string `json:"fee"`          // 手续费
}

// SendTransactionResponse 发送交易响应
type SendTransactionResponse struct {
	Success   bool   `json:"success"`
	TxHash    string `json:"tx_hash"`
	Message   string `json:"message"`
	ErrorCode string `json:"error_code,omitempty"`
}

// NewRPCClientManager 创建RPC客户端管理器
func NewRPCClientManager() *RPCClientManager {
	manager := &RPCClientManager{
		ethFailovers: make(map[string]*EthFailoverManager),
		btcFailovers: make(map[string]*BTCFailoverManager),
		btcClients:   make(map[string]*BitcoinRPCClient),
		solFailovers: make(map[string]*SolFailoverManager),
		logger:       logrus.New(),
	}

	// 初始化所有配置的链客户端
	for chainName, chainConfig := range config.AppConfig.Blockchain.Chains {
		if !chainConfig.Enabled {
			continue
		}

		switch strings.ToLower(chainName) {
		case "eth", "ethereum":
			if fo, err := NewEthFailoverFromChain(chainName); err == nil {
				manager.ethFailovers[chainName] = fo
				manager.logger.Infof("Initialized ETH failover: %s", chainName)
			} else {
				manager.logger.Errorf("Failed to init ETH failover %s: %v", chainName, err)
			}
		case "bsc", "binance":
			// BSC使用ETH故障转移管理器（因为BSC兼容EVM）
			if fo, err := NewEthFailoverFromChain(chainName); err == nil {
				manager.ethFailovers[chainName] = fo
				manager.logger.Infof("Initialized BSC failover (using ETH client): %s", chainName)
			} else {
				manager.logger.Errorf("Failed to init BSC failover %s: %v", chainName, err)
			}
		case "btc", "bitcoin":
			// 使用BTC故障转移管理器
			if btcFailover, err := NewBTCFailoverFromChain(chainName); err == nil {
				manager.btcFailovers[chainName] = btcFailover
				manager.logger.Infof("Initialized BTC failover manager: %s", chainName)
			} else {
				manager.logger.Errorf("Failed to init BTC failover %s: %v", chainName, err)
			}
		case "sol", "solana":
			if fo, err := NewSolFailoverFromChain(chainName); err == nil {
				manager.solFailovers[chainName] = fo
				manager.logger.Infof("Initialized SOL failover manager: %s", chainName)
			} else {
				manager.logger.Errorf("Failed to init SOL failover %s: %v", chainName, err)
			}
		}
	}

	return manager
}

// SendTransaction 发送交易
func (m *RPCClientManager) SendTransaction(ctx context.Context, req *SendTransactionRequest) (*SendTransactionResponse, error) {
	chainName := strings.ToLower(req.Chain)

	switch chainName {
	case "eth", "ethereum":
		return m.sendEthTransaction(ctx, req)
	case "bsc", "binance":
		return m.sendBscTransaction(ctx, req)
	case "btc", "bitcoin":
		return m.sendBtcTransaction(ctx, req)
	default:
		return &SendTransactionResponse{
			Success:   false,
			Message:   fmt.Sprintf("不支持的链类型: %s", req.Chain),
			ErrorCode: "UNSUPPORTED_CHAIN",
		}, nil
	}
}

// sendEthTransaction 发送以太坊交易
func (m *RPCClientManager) sendEthTransaction(ctx context.Context, req *SendTransactionRequest) (*SendTransactionResponse, error) {
	// 获取ETH故障转移管理器
	fo, exists := m.ethFailovers["eth"]
	if !exists {
		// 尝试其他可能的键名
		for key, f := range m.ethFailovers {
			if strings.Contains(strings.ToLower(key), "eth") {
				fo = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return &SendTransactionResponse{
			Success:   false,
			Message:   "ETH RPC故障转移未初始化",
			ErrorCode: "RPC_CLIENT_NOT_AVAILABLE",
		}, nil
	}

	// 解析已签名的交易（兼容多种输入格式）
	tx, err := parseSignedEthTx(req.SignedTx)
	if err != nil {
		return &SendTransactionResponse{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: "INVALID_SIGNED_TX",
		}, nil
	}

	// 发送交易（故障转移）
	err = fo.SendTransaction(ctx, tx)
	if err != nil {
		m.logger.Errorf("发送ETH交易失败: %v", err)
		return &SendTransactionResponse{
			Success:   false,
			Message:   fmt.Sprintf("发送交易失败: %v", err),
			ErrorCode: "SEND_TX_FAILED",
		}, nil
	}

	txHash := tx.Hash().Hex()
	m.logger.Infof("ETH交易发送成功: %s", txHash)

	return &SendTransactionResponse{
		Success: true,
		TxHash:  txHash,
		Message: "交易发送成功",
	}, nil
}

// sendBscTransaction 发送BSC交易
func (m *RPCClientManager) sendBscTransaction(ctx context.Context, req *SendTransactionRequest) (*SendTransactionResponse, error) {
	// 获取BSC故障转移管理器
	fo, exists := m.ethFailovers["bsc"]
	if !exists {
		// 尝试其他可能的键名
		for key, f := range m.ethFailovers {
			if strings.Contains(strings.ToLower(key), "bsc") || strings.Contains(strings.ToLower(key), "binance") {
				fo = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return &SendTransactionResponse{
			Success:   false,
			Message:   "BSC RPC故障转移未初始化",
			ErrorCode: "RPC_CLIENT_NOT_AVAILABLE",
		}, nil
	}

	// 解析已签名的交易（BSC使用与ETH相同的格式）
	tx, err := parseSignedEthTx(req.SignedTx)
	if err != nil {
		return &SendTransactionResponse{
			Success:   false,
			Message:   err.Error(),
			ErrorCode: "INVALID_SIGNED_TX",
		}, nil
	}

	// 发送交易（故障转移）
	err = fo.SendTransaction(ctx, tx)
	if err != nil {
		m.logger.Errorf("发送BSC交易失败: %v", err)
		return &SendTransactionResponse{
			Success:   false,
			Message:   fmt.Sprintf("发送交易失败: %v", err),
			ErrorCode: "SEND_TX_FAILED",
		}, nil
	}

	txHash := tx.Hash().Hex()
	m.logger.Infof("BSC交易发送成功: %s", txHash)

	return &SendTransactionResponse{
		Success: true,
		TxHash:  txHash,
		Message: "交易发送成功",
	}, nil
}

// parseSignedEthTx 解析多种格式的已签名ETH交易为 types.Transaction
// 支持：
// - 原始RLP十六进制（带或不带0x）
// - JSON对象中包含 rawTransaction/raw/signedTx 字段
// - 检测到是32字节哈希时，明确报错
func parseSignedEthTx(input string) (*types.Transaction, error) {
	trimmed := strings.TrimSpace(input)
	// JSON 包装
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &obj); err == nil {
			// 可能字段名
			candidates := []string{"rawTransaction", "raw", "signedTx", "signed_tx"}
			for _, k := range candidates {
				if v, ok := obj[k]; ok {
					if s, ok2 := v.(string); ok2 {
						return parseSignedEthTx(s)
					}
				}
			}
			return nil, fmt.Errorf("未在JSON中找到原始已签名交易字段(rawTransaction/raw/signedTx)")
		}
		// JSON解析失败则继续按HEX处理
	}
	// 允许 0x 前缀
	hexStr := trimmed
	if strings.HasPrefix(hexStr, "0x") || strings.HasPrefix(hexStr, "0X") {
		hexStr = hexStr[2:]
	}
	// 仅十六进制字符
	if ok, _ := regexp.MatchString("^[0-9a-fA-F]+$", hexStr); !ok {
		return nil, fmt.Errorf("签名交易应为十六进制字符串或包含rawTransaction的JSON")
	}
	// 如果长度为64（32字节），很可能是交易哈希，而非原始交易
	if len(hexStr) == 64 {
		return nil, fmt.Errorf("收到看似交易哈希的值，而非原始已签名交易，请提供rawTransaction数据")
	}
	// 如果长度为130（65字节），很可能是裸签名(r||s||v)，并非原始交易
	if len(hexStr) == 130 {
		return nil, fmt.Errorf("收到看似签名组件(r||s||v)的值，而非原始已签名交易，请提供rawTransaction十六进制串")
	}

	// 解码RLP
	bytesData, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("解析交易数据失败: %v", err)
	}

	tx := &types.Transaction{}
	if err := tx.UnmarshalBinary(bytesData); err != nil {
		return nil, fmt.Errorf("反序列化交易失败: %v", err)
	}
	return tx, nil
}

// sendBtcTransaction 发送比特币交易
func (m *RPCClientManager) sendBtcTransaction(ctx context.Context, req *SendTransactionRequest) (*SendTransactionResponse, error) {
	// 获取BTC故障转移管理器
	fo, exists := m.btcFailovers["btc"]
	if !exists {
		// 尝试其他可能的键名
		for key, f := range m.btcFailovers {
			if strings.Contains(strings.ToLower(key), "btc") {
				fo = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return &SendTransactionResponse{
			Success:   false,
			Message:   "BTC RPC故障转移未初始化",
			ErrorCode: "RPC_CLIENT_NOT_AVAILABLE",
		}, nil
	}

	// 发送原始交易（故障转移）
	txHash, err := fo.SendRawTransaction(ctx, req.SignedTx)
	if err != nil {
		m.logger.Errorf("发送BTC交易失败: %v", err)
		return &SendTransactionResponse{
			Success:   false,
			Message:   fmt.Sprintf("发送交易失败: %v", err),
			ErrorCode: "SEND_TX_FAILED",
		}, nil
	}

	m.logger.Infof("BTC交易发送成功: %s", txHash)

	return &SendTransactionResponse{
		Success: true,
		TxHash:  txHash,
		Message: "交易发送成功",
	}, nil
}

// SendRawTransaction 发送原始交易（BTC）
func (c *BitcoinRPCClient) SendRawTransaction(ctx context.Context, rawTx string) (string, error) {
	// 准备RPC请求
	request := BitcoinRPCRequest{
		JSONRPC: "1.0",
		ID:      "1",
		Method:  "sendrawtransaction",
		Params:  []interface{}{rawTx},
	}

	// 发送请求
	response, err := c.callRPC(ctx, request)
	if err != nil {
		return "", fmt.Errorf("RPC调用失败: %w", err)
	}

	// 检查错误
	if response.Error != nil {
		return "", fmt.Errorf("RPC错误: %s (代码: %d)", response.Error.Message, response.Error.Code)
	}

	// 解析结果
	txHash, ok := response.Result.(string)
	if !ok {
		return "", fmt.Errorf("无效的响应格式")
	}

	return txHash, nil
}

// callRPC 调用RPC接口
func (c *BitcoinRPCClient) callRPC(ctx context.Context, request BitcoinRPCRequest) (*BitcoinRPCResponse, error) {
	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %d, 响应: %s", resp.StatusCode, string(responseBody))
	}

	// 解析响应
	var response BitcoinRPCResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &response, nil
}

// GetTransactionStatus 获取交易状态
func (m *RPCClientManager) GetTransactionStatus(ctx context.Context, chain, txHash string) (*TransactionStatus, error) {
	chainName := strings.ToLower(chain)

	switch chainName {
	case "eth", "ethereum":
		return m.getEthTransactionStatus(ctx, txHash)
	case "bsc", "binance":
		return m.getBscTransactionStatus(ctx, txHash)
	case "btc", "bitcoin":
		return m.getBtcTransactionStatus(ctx, txHash)
	default:
		return nil, fmt.Errorf("不支持的链类型: %s", chain)
	}
}

// TransactionStatus 交易状态
type TransactionStatus struct {
	TxHash        string `json:"tx_hash"`
	Status        string `json:"status"`        // pending, confirmed, failed
	BlockHeight   uint64 `json:"block_height"`  // 区块高度
	Confirmations uint64 `json:"confirmations"` // 确认数
	GasUsed       uint64 `json:"gas_used"`      // 使用的Gas（ETH）
	GasPrice      string `json:"gas_price"`     // Gas价格（ETH）
	ActualFee     string `json:"actual_fee"`    // 实际手续费
}

// getEthTransactionStatus 获取ETH交易状态
func (m *RPCClientManager) getEthTransactionStatus(ctx context.Context, txHash string) (*TransactionStatus, error) {
	client, exists := m.ethFailovers["eth"]
	if !exists {
		for key, f := range m.ethFailovers {
			if strings.Contains(strings.ToLower(key), "eth") {
				client = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return nil, fmt.Errorf("ETH RPC客户端未配置")
	}

	// 获取交易
	tx, isPending, err := client.TransactionByHash(ctx, common.HexToHash(txHash))
	if err != nil {
		return nil, fmt.Errorf("获取交易失败: %w", err)
	}

	status := &TransactionStatus{
		TxHash: txHash,
		Status: "pending",
	}

	if isPending {
		status.Status = "pending"
		return status, nil
	}

	// 获取交易收据
	receipt, err := client.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		status.Status = "failed"
		return status, nil
	}

	// 获取最新区块高度
	latestBlock, err := client.BlockNumber(ctx)
	if err != nil {
		status.Status = "confirmed"
		status.BlockHeight = receipt.BlockNumber.Uint64()
		status.Confirmations = 1
		return status, nil
	}

	status.Status = "confirmed"
	status.BlockHeight = receipt.BlockNumber.Uint64()
	status.Confirmations = latestBlock - receipt.BlockNumber.Uint64() + 1
	status.GasUsed = receipt.GasUsed
	status.GasPrice = tx.GasPrice().String()
	status.ActualFee = new(big.Int).Mul(tx.GasPrice(), big.NewInt(int64(receipt.GasUsed))).String()

	return status, nil
}

// getBscTransactionStatus 获取BSC交易状态
func (m *RPCClientManager) getBscTransactionStatus(ctx context.Context, txHash string) (*TransactionStatus, error) {
	// 获取BSC故障转移管理器
	client, exists := m.ethFailovers["bsc"]
	if !exists {
		// 尝试其他可能的键名
		for key, f := range m.ethFailovers {
			if strings.Contains(strings.ToLower(key), "bsc") || strings.Contains(strings.ToLower(key), "binance") {
				client = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return nil, fmt.Errorf("BSC RPC客户端未配置")
	}

	// 获取交易
	tx, isPending, err := client.TransactionByHash(ctx, common.HexToHash(txHash))
	if err != nil {
		return nil, fmt.Errorf("获取交易失败: %w", err)
	}

	status := &TransactionStatus{
		TxHash: txHash,
		Status: "pending",
	}

	if isPending {
		status.Status = "pending"
		return status, nil
	}

	// 获取交易收据
	receipt, err := client.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		status.Status = "failed"
		return status, nil
	}

	// 获取最新区块高度
	latestBlock, err := client.BlockNumber(ctx)
	if err != nil {
		status.Status = "confirmed"
		status.BlockHeight = receipt.BlockNumber.Uint64()
		status.Confirmations = 1
		return status, nil
	}

	status.Status = "confirmed"
	status.BlockHeight = receipt.BlockNumber.Uint64()
	status.Confirmations = latestBlock - receipt.BlockNumber.Uint64() + 1
	status.GasUsed = receipt.GasUsed
	status.GasPrice = tx.GasPrice().String()
	status.ActualFee = new(big.Int).Mul(tx.GasPrice(), big.NewInt(int64(receipt.GasUsed))).String()

	return status, nil
}

// getBtcTransactionStatus 获取BTC交易状态
func (m *RPCClientManager) getBtcTransactionStatus(ctx context.Context, txHash string) (*TransactionStatus, error) {
	// 获取BTC故障转移管理器
	fo, exists := m.btcFailovers["btc"]
	if !exists {
		// 尝试其他可能的键名
		for key, f := range m.btcFailovers {
			if strings.Contains(strings.ToLower(key), "btc") {
				fo = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return nil, fmt.Errorf("BTC RPC故障转移未初始化")
	}

	// 获取交易状态信息（故障转移）
	txInfo, err := fo.GetTransactionStatus(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("获取交易信息失败: %w", err)
	}

	status := &TransactionStatus{
		TxHash: txHash,
	}

	// 解析交易信息
	if txInfo.Confirmations >= 0 {
		status.Status = "confirmed"
		status.Confirmations = uint64(txInfo.Confirmations)
	} else {
		status.Status = "pending"
		status.Confirmations = 0
	}

	if txInfo.BlockHeight > 0 {
		status.BlockHeight = uint64(txInfo.BlockHeight)
	}

	return status, nil
}

// GetBlockNumber 获取最新区块号
func (m *RPCClientManager) GetBlockNumber(ctx context.Context, chain string) (uint64, error) {
	chainName := strings.ToLower(chain)

	switch chainName {
	case "eth", "ethereum":
		return m.getETHBlockNumber(ctx)
	case "bsc", "binance":
		return m.getBSCBlockNumber(ctx)
	case "btc", "bitcoin":
		return m.getBTCBlockNumber(ctx)
	default:
		return 0, fmt.Errorf("不支持的链类型: %s", chain)
	}
}

// getETHBlockNumber 获取ETH最新区块号
func (m *RPCClientManager) getETHBlockNumber(ctx context.Context) (uint64, error) {
	// 获取ETH故障转移管理器
	fo, exists := m.ethFailovers["eth"]
	if !exists {
		// 尝试其他可能的键名
		for key, f := range m.ethFailovers {
			if strings.Contains(strings.ToLower(key), "eth") {
				fo = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return 0, fmt.Errorf("未找到ETH故障转移管理器")
	}

	return fo.BlockNumber(ctx)
}

// getBSCBlockNumber 获取BSC最新区块号
func (m *RPCClientManager) getBSCBlockNumber(ctx context.Context) (uint64, error) {
	// 获取BSC故障转移管理器
	fo, exists := m.ethFailovers["bsc"]
	if !exists {
		// 尝试其他可能的键名
		for key, f := range m.ethFailovers {
			if strings.Contains(strings.ToLower(key), "bsc") || strings.Contains(strings.ToLower(key), "binance") {
				fo = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return 0, fmt.Errorf("未找到BSC故障转移管理器")
	}

	return fo.BlockNumber(ctx)
}

// getBTCBlockNumber 获取BTC最新区块号
func (m *RPCClientManager) getBTCBlockNumber(ctx context.Context) (uint64, error) {
	// 获取BTC故障转移管理器
	fo, exists := m.btcFailovers["btc"]
	if !exists {
		// 尝试其他可能的键名
		for key, f := range m.btcFailovers {
			if strings.Contains(strings.ToLower(key), "btc") {
				fo = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return 0, fmt.Errorf("未找到BTC故障转移管理器")
	}

	return fo.GetLatestBlockHeight(ctx)
}

// GetBlockByNumber 根据区块号获取区块
func (m *RPCClientManager) GetBlockByNumber(ctx context.Context, chain string, blockNumber *big.Int) (interface{}, error) {
	chainName := strings.ToLower(chain)

	switch chainName {
	case "eth", "ethereum":
		return m.getETHBlockByNumber(ctx, blockNumber)
	case "bsc", "binance":
		return m.getBSCBlockByNumber(ctx, blockNumber)
	case "btc", "bitcoin":
		return m.getBTCBlockByNumber(ctx, blockNumber)
	default:
		return nil, fmt.Errorf("不支持的链类型: %s", chain)
	}
}

// GetBlockByHash 根据区块哈希获取区块
func (m *RPCClientManager) GetBlockByHash(ctx context.Context, chain string, blockHash string) (interface{}, error) {
	chainName := strings.ToLower(chain)

	switch chainName {
	case "eth", "ethereum":
		return m.getETHBlockByHash(ctx, blockHash)
	case "bsc", "binance":
		return m.getBSCBlockByHash(ctx, blockHash)
	case "btc", "bitcoin":
		return m.getBTCBlockByHash(ctx, blockHash)
	default:
		return nil, fmt.Errorf("不支持的链类型: %s", chain)
	}
}

// getETHBlockByNumber 获取ETH区块
func (m *RPCClientManager) getETHBlockByNumber(ctx context.Context, blockNumber *big.Int) (*types.Block, error) {
	// 获取ETH故障转移管理器
	fo, exists := m.ethFailovers["eth"]
	if !exists {
		// 尝试其他可能的键名
		for key, f := range m.ethFailovers {
			if strings.Contains(strings.ToLower(key), "eth") {
				fo = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return nil, fmt.Errorf("未找到ETH故障转移管理器")
	}

	return fo.BlockByNumber(ctx, blockNumber)
}

// getBSCBlockByNumber 获取BSC区块
func (m *RPCClientManager) getBSCBlockByNumber(ctx context.Context, blockNumber *big.Int) (*types.Block, error) {
	// 获取BSC故障转移管理器
	fo, exists := m.ethFailovers["bsc"]
	if !exists {
		// 尝试其他可能的键名
		for key, f := range m.ethFailovers {
			if strings.Contains(strings.ToLower(key), "bsc") || strings.Contains(strings.ToLower(key), "binance") {
				fo = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return nil, fmt.Errorf("未找到BSC故障转移管理器")
	}

	return fo.BlockByNumber(ctx, blockNumber)
}

// getETHBlockByHash 获取ETH区块（通过哈希）
func (m *RPCClientManager) getETHBlockByHash(ctx context.Context, blockHash string) (*types.Block, error) {
	// 获取ETH故障转移管理器
	fo, exists := m.ethFailovers["eth"]
	if !exists {
		// 尝试其他可能的键名
		for key, f := range m.ethFailovers {
			if strings.Contains(strings.ToLower(key), "eth") {
				fo = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return nil, fmt.Errorf("未找到ETH故障转移管理器")
	}

	hash := common.HexToHash(blockHash)
	return fo.BlockByHash(ctx, hash)
}

// getBSCBlockByHash 获取BSC区块（通过哈希）
func (m *RPCClientManager) getBSCBlockByHash(ctx context.Context, blockHash string) (*types.Block, error) {
	// 获取BSC故障转移管理器
	fo, exists := m.ethFailovers["bsc"]
	if !exists {
		// 尝试其他可能的键名
		for key, f := range m.ethFailovers {
			if strings.Contains(strings.ToLower(key), "bsc") || strings.Contains(strings.ToLower(key), "binance") {
				fo = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return nil, fmt.Errorf("未找到BSC故障转移管理器")
	}

	hash := common.HexToHash(blockHash)
	return fo.BlockByHash(ctx, hash)
}

// getBTCBlockByNumber 获取BTC区块（通过区块号）
func (m *RPCClientManager) getBTCBlockByNumber(ctx context.Context, blockNumber *big.Int) (map[string]interface{}, error) {
	// 获取BTC故障转移管理器
	fo, exists := m.btcFailovers["btc"]
	if !exists {
		// 尝试其他可能的键名
		for key, f := range m.btcFailovers {
			if strings.Contains(strings.ToLower(key), "btc") {
				fo = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return nil, fmt.Errorf("未找到BTC故障转移管理器")
	}

	// 先通过区块号获取区块哈希
	blockHash, err := fo.GetBlockHash(ctx, blockNumber.Uint64())
	if err != nil {
		return nil, fmt.Errorf("获取区块哈希失败: %w", err)
	}

	// 再通过区块哈希获取区块详情
	return fo.GetBlock(ctx, blockHash)
}

// getBTCBlockByHash 获取BTC区块（通过哈希）
func (m *RPCClientManager) getBTCBlockByHash(ctx context.Context, blockHash string) (map[string]interface{}, error) {
	// 获取BTC故障转移管理器
	fo, exists := m.btcFailovers["btc"]
	if !exists {
		// 尝试其他可能的键名
		for key, f := range m.btcFailovers {
			if strings.Contains(strings.ToLower(key), "btc") {
				fo = f
				exists = true
				break
			}
		}
	}

	if !exists {
		return nil, fmt.Errorf("未找到BTC故障转移管理器")
	}

	return fo.GetBlock(ctx, blockHash)
}

// EstimateEthGas 估算以太坊交易的Gas上限
func (m *RPCClientManager) EstimateEthGas(ctx context.Context, from, to string, value *big.Int, data []byte) (uint64, error) {
	// 获取ETH故障转移管理器
	fo, exists := m.ethFailovers["eth"]
	if !exists {
		for key, f := range m.ethFailovers {
			if strings.Contains(strings.ToLower(key), "eth") {
				fo = f
				exists = true
				break
			}
		}
	}
	if !exists {
		return 0, fmt.Errorf("ETH RPC故障转移未初始化")
	}

	var toAddr *common.Address
	if to != "" {
		addr := common.HexToAddress(to)
		toAddr = &addr
	}

	msg := ethereum.CallMsg{
		From:  common.HexToAddress(from),
		To:    toAddr,
		Value: value,
		Data:  data,
	}
	// fmt.Printf("🔍 估算Gas: %+v\n", msg)
	// fmt.Printf("🔍 估算Gas: %+v\n", data)
	return fo.EstimateGas(ctx, msg)
}

// EstimateBscGas 估算BSC交易的Gas上限（EVM兼容）
func (m *RPCClientManager) EstimateBscGas(ctx context.Context, from, to string, value *big.Int, data []byte) (uint64, error) {
	fo, exists := m.ethFailovers["bsc"]
	if !exists {
		for key, f := range m.ethFailovers {
			if strings.Contains(strings.ToLower(key), "bsc") || strings.Contains(strings.ToLower(key), "binance") {
				fo = f
				exists = true
				break
			}
		}
	}
	if !exists {
		return 0, fmt.Errorf("BSC RPC故障转移未初始化")
	}

	var toAddr *common.Address
	if to != "" {
		addr := common.HexToAddress(to)
		toAddr = &addr
	}
	msg := ethereum.CallMsg{
		From:  common.HexToAddress(from),
		To:    toAddr,
		Value: value,
		Data:  data,
	}
	return fo.EstimateGas(ctx, msg)
}

// CallContract 调用合约方法（eth_call）
func (m *RPCClientManager) CallContract(ctx context.Context, from, to string, value *big.Int, data []byte, blockNumber *big.Int) ([]byte, error) {
	// 获取ETH故障转移管理器
	fo, exists := m.ethFailovers["eth"]
	if !exists {
		for key, f := range m.ethFailovers {
			if strings.Contains(strings.ToLower(key), "eth") {
				fo = f
				exists = true
				break
			}
		}
	}
	if !exists {
		return nil, fmt.Errorf("ETH RPC故障转移未初始化")
	}

	var toAddr *common.Address
	if to != "" {
		addr := common.HexToAddress(to)
		toAddr = &addr
	}

	msg := ethereum.CallMsg{
		From:  common.HexToAddress(from),
		To:    toAddr,
		Value: value,
		Data:  data,
	}

	return fo.CallContract(ctx, msg, blockNumber)
}

// CallContractOnChain 按链调用合约方法（eth_call）
func (m *RPCClientManager) CallContractOnChain(ctx context.Context, chain string, from, to string, value *big.Int, data []byte, blockNumber *big.Int) ([]byte, error) {
	chainLower := strings.ToLower(chain)

	var fo *EthFailoverManager
	var exists bool

	switch chainLower {
	case "eth", "ethereum":
		fo, exists = m.ethFailovers["eth"]
		if !exists {
			for key, f := range m.ethFailovers {
				if strings.Contains(strings.ToLower(key), "eth") {
					fo = f
					exists = true
					break
				}
			}
		}
	case "bsc", "binance":
		fo, exists = m.ethFailovers["bsc"]
		if !exists {
			for key, f := range m.ethFailovers {
				if strings.Contains(strings.ToLower(key), "bsc") || strings.Contains(strings.ToLower(key), "binance") {
					fo = f
					exists = true
					break
				}
			}
		}
	default:
		return nil, fmt.Errorf("不支持的EVM链: %s", chain)
	}

	if !exists || fo == nil {
		return nil, fmt.Errorf("%s RPC故障转移未初始化", strings.ToUpper(chainLower))
	}
	var toAddr *common.Address
	if to != "" {
		addr := common.HexToAddress(to)
		toAddr = &addr
	}

	msg := ethereum.CallMsg{
		From:  common.HexToAddress(from),
		To:    toAddr,
		Value: value,
		Data:  data,
	}

	return fo.CallContract(ctx, msg, blockNumber)
}

// GetBTCFailover 获取BTC故障转移管理器
func (m *RPCClientManager) GetBTCFailover(chain string) (*BTCFailoverManager, bool) {
	fo, exists := m.btcFailovers[chain]
	if !exists {
		for key, f := range m.btcFailovers {
			if strings.Contains(strings.ToLower(key), "btc") {
				fo = f
				exists = true
				break
			}
		}
	}
	return fo, exists
}

// GetSolanaClient 获取SOL故障转移管理器
func (m *RPCClientManager) GetSolanaClient(chain string) (*SolFailoverManager, error) {
	fo, exists := m.solFailovers[chain]
	if !exists {
		for key, f := range m.solFailovers {
			if strings.Contains(strings.ToLower(key), "sol") {
				fo = f
				exists = true
				break
			}
		}
	}
	if !exists {
		return nil, fmt.Errorf("SOL客户端未找到: %s", chain)
	}
	return fo, nil
}

// Close 关闭所有连接
func (m *RPCClientManager) Close() {
	for _, client := range m.ethFailovers {
		client.Close()
	}
	m.logger.Info("RPC客户端管理器已关闭")
}
