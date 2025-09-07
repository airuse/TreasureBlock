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
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// RPCClientManager RPC客户端管理器
type RPCClientManager struct {
	ethFailovers map[string]*EthFailoverManager // 链名 -> ETH故障转移管理器
	btcClients   map[string]*BitcoinRPCClient   // 链名 -> BTC客户端
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
		btcClients:   make(map[string]*BitcoinRPCClient),
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
		case "btc", "bitcoin":
			btcClient := &BitcoinRPCClient{
				config:     &chainConfig,
				httpClient: &http.Client{Timeout: 30 * time.Second},
				baseURL:    chainConfig.RPCURL,
				username:   chainConfig.Username,
				password:   chainConfig.Password,
			}
			manager.btcClients[chainName] = btcClient
			manager.logger.Infof("Initialized BTC RPC client: %s", chainName)
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
	// 获取BTC客户端
	client, exists := m.btcClients["btc"]
	if !exists {
		// 尝试其他可能的键名
		for key, cli := range m.btcClients {
			if strings.Contains(strings.ToLower(key), "btc") {
				client = cli
				exists = true
				break
			}
		}
	}

	if !exists {
		return &SendTransactionResponse{
			Success:   false,
			Message:   "BTC RPC客户端未配置或连接失败",
			ErrorCode: "RPC_CLIENT_NOT_AVAILABLE",
		}, nil
	}

	// 发送原始交易
	txHash, err := client.SendRawTransaction(ctx, req.SignedTx)
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

// getBtcTransactionStatus 获取BTC交易状态
func (m *RPCClientManager) getBtcTransactionStatus(ctx context.Context, txHash string) (*TransactionStatus, error) {
	client, exists := m.btcClients["btc"]
	if !exists {
		for key, cli := range m.btcClients {
			if strings.Contains(strings.ToLower(key), "btc") {
				client = cli
				exists = true
				break
			}
		}
	}

	if !exists {
		return nil, fmt.Errorf("BTC RPC客户端未配置")
	}

	// 获取交易信息
	txInfo, err := client.GetTransaction(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("获取交易信息失败: %w", err)
	}

	status := &TransactionStatus{
		TxHash: txHash,
	}

	// 解析交易信息
	if txInfo["confirmations"] != nil {
		confirmations, ok := txInfo["confirmations"].(float64)
		if ok && confirmations > 0 {
			status.Status = "confirmed"
			status.Confirmations = uint64(confirmations)
		} else {
			status.Status = "pending"
			status.Confirmations = 0
		}
	} else {
		status.Status = "pending"
		status.Confirmations = 0
	}

	if txInfo["blockheight"] != nil {
		blockHeight, ok := txInfo["blockheight"].(float64)
		if ok {
			status.BlockHeight = uint64(blockHeight)
		}
	}

	return status, nil
}

// GetTransaction 获取交易信息（BTC）
func (c *BitcoinRPCClient) GetTransaction(ctx context.Context, txHash string) (map[string]interface{}, error) {
	request := BitcoinRPCRequest{
		JSONRPC: "1.0",
		ID:      "1",
		Method:  "gettransaction",
		Params:  []interface{}{txHash},
	}

	response, err := c.callRPC(ctx, request)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC错误: %s (代码: %d)", response.Error.Message, response.Error.Code)
	}

	result, ok := response.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无效的响应格式")
	}

	return result, nil
}

// GetBlockCount 获取最新区块高度（BTC）
func (c *BitcoinRPCClient) GetBlockCount(ctx context.Context) (uint64, error) {
	request := BitcoinRPCRequest{
		JSONRPC: "1.0",
		ID:      "1",
		Method:  "getblockcount",
		Params:  []interface{}{},
	}

	response, err := c.callRPC(ctx, request)
	if err != nil {
		return 0, err
	}

	if response.Error != nil {
		return 0, fmt.Errorf("RPC错误: %s (代码: %d)", response.Error.Message, response.Error.Code)
	}

	count, ok := response.Result.(float64)
	if !ok {
		return 0, fmt.Errorf("无效的响应格式")
	}

	return uint64(count), nil
}

// GetBlockNumber 获取最新区块号
func (m *RPCClientManager) GetBlockNumber(ctx context.Context, chain string) (uint64, error) {
	chainName := strings.ToLower(chain)

	switch chainName {
	case "eth", "ethereum":
		return m.getETHBlockNumber(ctx)
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

// getBTCBlockNumber 获取BTC最新区块号
func (m *RPCClientManager) getBTCBlockNumber(ctx context.Context) (uint64, error) {
	// 获取BTC客户端
	client, exists := m.btcClients["btc"]
	if !exists {
		for key, cli := range m.btcClients {
			if strings.Contains(strings.ToLower(key), "btc") {
				client = cli
				exists = true
				break
			}
		}
	}

	if !exists {
		return 0, fmt.Errorf("未找到BTC客户端")
	}

	return client.GetBlockCount(ctx)
}

// GetBlockByNumber 根据区块号获取区块
func (m *RPCClientManager) GetBlockByNumber(ctx context.Context, chain string, blockNumber *big.Int) (*types.Block, error) {
	chainName := strings.ToLower(chain)

	switch chainName {
	case "eth", "ethereum":
		return m.getETHBlockByNumber(ctx, blockNumber)
	case "btc", "bitcoin":
		return nil, fmt.Errorf("BTC不支持通过区块号获取区块，请使用GetBlockByHash")
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
	fmt.Printf("🔍 估算Gas: %+v\n", msg)
	fmt.Printf("🔍 估算Gas: %+v\n", data)
	return fo.EstimateGas(ctx, msg)
}

// Close 关闭所有连接
func (m *RPCClientManager) Close() {
	for _, client := range m.ethFailovers {
		client.Close()
	}
	m.logger.Info("RPC客户端管理器已关闭")
}
