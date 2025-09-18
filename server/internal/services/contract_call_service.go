package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"blockChainBrowser/server/internal/utils"
)

// ContractCallService 合约调用服务接口
type ContractCallService interface {
	// 调用合约的只读方法
	CallBalanceOf(ctx context.Context, contractAddress, accountAddress string, blockNumber *big.Int) (*big.Int, error)
	CallAllowance(ctx context.Context, contractAddress, ownerAddress, spenderAddress string, blockNumber *big.Int) (*big.Int, error)
	// 指定链的只读方法（用于多EVM链）
	CallBalanceOfOnChain(ctx context.Context, chain, contractAddress, accountAddress string, blockNumber *big.Int) (*big.Int, error)
	CallAllowanceOnChain(ctx context.Context, chain, contractAddress, ownerAddress, spenderAddress string, blockNumber *big.Int) (*big.Int, error)
	// 通用合约调用方法
	CallContractMethod(ctx context.Context, contractAddress, callData string, blockNumber *big.Int) ([]byte, error)
}

type contractCallService struct {
	rpcManager *utils.RPCClientManager
}

// NewContractCallService 创建合约调用服务
func NewContractCallService(rpcManager *utils.RPCClientManager) ContractCallService {
	return &contractCallService{
		rpcManager: rpcManager,
	}
}

// CallBalanceOf 调用合约的 balanceOf 方法
func (s *contractCallService) CallBalanceOf(ctx context.Context, contractAddress, accountAddress string, blockNumber *big.Int) (*big.Int, error) {
	// balanceOf(address) 的函数选择器是 0x70a08231
	methodSelector := "0x70a08231"

	// 将地址参数编码为32字节（64个十六进制字符）
	// 去掉0x前缀，左填充0到64个字符
	addressParam := strings.TrimPrefix(accountAddress, "0x")
	addressParam = strings.ToLower(addressParam)
	// 确保地址是40个字符（20字节）
	if len(addressParam) != 40 {
		return nil, fmt.Errorf("invalid address length: %d", len(addressParam))
	}
	// 左填充0到64个字符
	addressParam = strings.Repeat("0", 24) + addressParam

	// 构造调用数据
	callData := methodSelector + addressParam

	// 调用合约
	result, err := s.CallContractMethod(ctx, contractAddress, callData, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	// 解析返回值（32字节的uint256）
	if len(result) != 32 {
		return nil, fmt.Errorf("invalid balanceOf result length: %d", len(result))
	}

	// 转换为big.Int
	balance := new(big.Int).SetBytes(result)
	return balance, nil
}

// CallBalanceOfOnChain 按链调用 balanceOf（EVM兼容，如 eth/bsc）
func (s *contractCallService) CallBalanceOfOnChain(ctx context.Context, chain, contractAddress, accountAddress string, blockNumber *big.Int) (*big.Int, error) {
	methodSelector := "0x70a08231"
	addressParam := strings.TrimPrefix(accountAddress, "0x")
	addressParam = strings.ToLower(addressParam)
	if len(addressParam) != 40 {
		return nil, fmt.Errorf("invalid address length: %d", len(addressParam))
	}
	addressParam = strings.Repeat("0", 24) + addressParam
	callData := methodSelector + addressParam

	result, err := s.rpcManager.CallContractOnChain(ctx, chain, "0x0000000000000000000000000000000000000000", contractAddress, big.NewInt(0), mustDecodeHex(callData), blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf on %s: %w", chain, err)
	}
	if len(result) != 32 {
		return nil, fmt.Errorf("invalid balanceOf result length: %d", len(result))
	}
	return new(big.Int).SetBytes(result), nil
}

// CallAllowance 调用合约的 allowance 方法
func (s *contractCallService) CallAllowance(ctx context.Context, contractAddress, ownerAddress, spenderAddress string, blockNumber *big.Int) (*big.Int, error) {
	// allowance(address,address) 的函数选择器是 0xdd62ed3e
	methodSelector := "0xdd62ed3e"

	// 编码两个地址参数
	ownerParam := s.encodeAddress(ownerAddress)
	spenderParam := s.encodeAddress(spenderAddress)

	// 构造调用数据
	callData := methodSelector + ownerParam + spenderParam

	// 调用合约
	result, err := s.CallContractMethod(ctx, contractAddress, callData, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to call allowance: %w", err)
	}
	// 解析返回值（32字节的uint256）
	if len(result) != 32 {
		return nil, fmt.Errorf("invalid allowance result length: %d", len(result))
	}

	// 转换为big.Int
	allowance := new(big.Int).SetBytes(result)
	return allowance, nil
}

// CallAllowanceOnChain 按链调用 allowance（EVM兼容，如 eth/bsc）
func (s *contractCallService) CallAllowanceOnChain(ctx context.Context, chain, contractAddress, ownerAddress, spenderAddress string, blockNumber *big.Int) (*big.Int, error) {
	methodSelector := "0xdd62ed3e"
	ownerParam := s.encodeAddress(ownerAddress)
	spenderParam := s.encodeAddress(spenderAddress)
	callData := methodSelector + ownerParam + spenderParam

	result, err := s.rpcManager.CallContractOnChain(ctx, chain, "0x0000000000000000000000000000000000000000", contractAddress, big.NewInt(0), mustDecodeHex(callData), blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to call allowance on %s: %w", chain, err)
	}
	if len(result) != 32 {
		return nil, fmt.Errorf("invalid allowance result length: %d", len(result))
	}
	return new(big.Int).SetBytes(result), nil
}

func mustDecodeHex(s string) []byte {
	b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil {
		panic(err)
	}
	return b
}

// CallContractMethod 通用合约调用方法
func (s *contractCallService) CallContractMethod(ctx context.Context, contractAddress, callData string, blockNumber *big.Int) ([]byte, error) {
	// 将十六进制字符串转换为字节数组
	data, err := hex.DecodeString(strings.TrimPrefix(callData, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid call data: %w", err)
	}

	// 调用合约（使用零地址作为from，因为这是只读调用）
	fromAddress := "0x0000000000000000000000000000000000000000"

	result, err := s.rpcManager.CallContract(ctx, fromAddress, contractAddress, big.NewInt(0), data, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}

	return result, nil
}

// encodeAddress 将地址编码为32字节的十六进制字符串
func (s *contractCallService) encodeAddress(address string) string {
	// 去掉0x前缀
	addr := strings.TrimPrefix(address, "0x")
	addr = strings.ToLower(addr)

	// 确保地址是40个字符（20字节）
	if len(addr) != 40 {
		panic(fmt.Sprintf("invalid address length: %d", len(addr)))
	}

	// 左填充0到64个字符
	return strings.Repeat("0", 24) + addr
}

// ContractCallResult 合约调用结果
type ContractCallResult struct {
	MethodName   string   `json:"method_name"`
	ContractAddr string   `json:"contract_addr"`
	AccountAddr  string   `json:"account_addr,omitempty"`
	OwnerAddr    string   `json:"owner_addr,omitempty"`
	SpenderAddr  string   `json:"spender_addr,omitempty"`
	Result       *big.Int `json:"result"`
	BlockNumber  *big.Int `json:"block_number"`
	CallData     string   `json:"call_data"`
	RawResult    string   `json:"raw_result"`
}

// ParseContractCallResult 解析合约调用结果
func (s *contractCallService) ParseContractCallResult(methodName, contractAddr, accountAddr string, result []byte, blockNumber *big.Int, callData string) *ContractCallResult {
	return &ContractCallResult{
		MethodName:   methodName,
		ContractAddr: contractAddr,
		AccountAddr:  accountAddr,
		Result:       new(big.Int).SetBytes(result),
		BlockNumber:  blockNumber,
		CallData:     callData,
		RawResult:    "0x" + hex.EncodeToString(result),
	}
}
