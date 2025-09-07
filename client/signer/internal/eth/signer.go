package eth

import (
	"blockChainBrowser/client/signer/internal/crypto"
	"blockChainBrowser/client/signer/internal/utils"
	"blockChainBrowser/client/signer/pkg"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// ETHSigner ETH签名器
type ETHSigner struct {
	cryptoManager *crypto.CryptoManager
}

// NewETHSigner 创建ETH签名器
func NewETHSigner(cryptoManager *crypto.CryptoManager) *ETHSigner {
	return &ETHSigner{
		cryptoManager: cryptoManager,
	}
}

// SignTransaction 签名ETH交易
func (es *ETHSigner) SignTransaction(transaction *pkg.TransactionData) (string, error) {
	fmt.Println("🔷 开始签名ETH交易...")

	// 1. 根据from地址查找对应的私钥
	keyManager := crypto.NewKeyManager(es.cryptoManager)
	if err := keyManager.LoadKeys(); err != nil {
		return "", fmt.Errorf("加载私钥失败: %w", err)
	}

	if !keyManager.HasKey(transaction.From) {
		return "", fmt.Errorf("未找到地址 %s 对应的私钥", transaction.From)
	}

	// 获取解密密码（隐藏输入）- 合并确认和密码输入步骤
	password, err := utils.ReadPassword("请确认此交易并输入私钥解密密码（无密码回车则视为取消）: ")
	if err != nil {
		return "", fmt.Errorf("读取密码失败: %w", err)
	}

	// 如果密码为空，视为取消操作
	if password == "" {
		return "", fmt.Errorf("操作已取消")
	}

	// 解密私钥
	privateKeyHex, err := keyManager.GetKey(transaction.From, password)
	if err != nil {
		return "", fmt.Errorf("解密私钥失败: %w", err)
	}

	// 解析私钥
	privateKey, err := ethcrypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("解析私钥失败: %w", err)
	}

	// 2. 构建EIP-1559交易结构
	tx, err := es.buildTransaction(transaction)
	if err != nil {
		return "", fmt.Errorf("构建交易失败: %w", err)
	}

	// 3. 使用私钥签名交易（严格按交易的链ID选择签名器）
	parsedChainID, perr := strconv.ParseInt(transaction.ChainID, 10, 64)
	if perr != nil {
		return "", fmt.Errorf("解析链ID失败: %w", perr)
	}
	signer := types.LatestSignerForChainID(big.NewInt(parsedChainID))
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %w", err)
	}

	// 4. 返回完整的签名交易（包含EIP-1559类型前缀）
	rawTx, err := signedTx.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("编码交易失败: %w", err)
	}

	return "0x" + hex.EncodeToString(rawTx), nil
}

// buildTransaction 构建交易结构
func (es *ETHSigner) buildTransaction(transaction *pkg.TransactionData) (*types.Transaction, error) {
	// 解析链ID
	chainID, err := strconv.ParseInt(transaction.ChainID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("解析链ID失败: %w", err)
	}

	// 解析接收地址
	to := common.HexToAddress(transaction.To)

	// 解析交易金额
	value, err := es.parseHexValue(transaction.Value)
	if err != nil {
		return nil, fmt.Errorf("解析交易金额失败: %w", err)
	}

	// 解析交易数据
	dataHex := transaction.Data
	if strings.HasPrefix(dataHex, "0x") {
		dataHex = dataHex[2:]
	}
	data, err := hex.DecodeString(dataHex)
	if err != nil {
		return nil, fmt.Errorf("解析交易数据失败: %w", err)
	}

	// 构建AccessList
	accessList := es.buildAccessList(transaction.AccessList)

	// 解析费率设置 - 费率必须提供，不能为空
	if transaction.MaxPriorityFeePerGas == "" {
		return nil, fmt.Errorf("MaxPriorityFeePerGas不能为空，请在导出交易时设置费率")
	}
	if transaction.MaxFeePerGas == "" {
		return nil, fmt.Errorf("MaxFeePerGas不能为空，请在导出交易时设置费率")
	}

	gasTipCap, err := strconv.ParseFloat(transaction.MaxPriorityFeePerGas, 64)
	if err != nil {
		return nil, fmt.Errorf("解析MaxPriorityFeePerGas失败: %w", err)
	}
	// if gasTipCap <= 0 {
	// 	return nil, fmt.Errorf("MaxPriorityFeePerGas必须大于0，当前值: %f", gasTipCap)
	// }

	gasFeeCap, err := strconv.ParseFloat(transaction.MaxFeePerGas, 64)
	if err != nil {
		return nil, fmt.Errorf("解析MaxFeePerGas失败: %w", err)
	}
	if gasFeeCap <= 0 {
		return nil, fmt.Errorf("MaxFeePerGas必须大于0，当前值: %f", gasFeeCap)
	}

	// 转换为wei
	gasTipCapWei := big.NewInt(int64(gasTipCap * 1e9))
	gasFeeCapWei := big.NewInt(int64(gasFeeCap * 1e9))

	// 创建EIP-1559交易，使用QR码中的费率设置
	// 优先使用从QR数据传入的 Gas 上限（如果提供）
	gasLimit := uint64(21000)
	if transaction.Gas > 0 {
		gasLimit = transaction.Gas
	}

	tx := &types.DynamicFeeTx{
		ChainID:    big.NewInt(chainID),
		Nonce:      transaction.Nonce,
		GasTipCap:  gasTipCapWei, // 使用QR码中的费率
		GasFeeCap:  gasFeeCapWei, // 使用QR码中的费率
		Gas:        gasLimit,     // 使用后端估算的gas limit
		To:         &to,
		Value:      value,
		Data:       data,
		AccessList: accessList,
	}

	return types.NewTx(tx), nil
}

// parseHexValue 解析十六进制金额
func (es *ETHSigner) parseHexValue(hexValue string) (*big.Int, error) {
	// 移除0x前缀
	if strings.HasPrefix(hexValue, "0x") {
		hexValue = hexValue[2:]
	}

	// 解析为big.Int
	value, ok := new(big.Int).SetString(hexValue, 16)
	if !ok {
		return nil, fmt.Errorf("无效的十六进制金额: %s", hexValue)
	}

	return value, nil
}

// buildAccessList 构建访问列表
func (es *ETHSigner) buildAccessList(accessList []pkg.AccessListItem) types.AccessList {
	var ethAccessList types.AccessList

	for _, item := range accessList {
		address := common.HexToAddress(item.Address)
		var storageKeys []common.Hash

		for _, key := range item.StorageKeys {
			storageKey := common.HexToHash(key)
			storageKeys = append(storageKeys, storageKey)
		}

		ethAccessList = append(ethAccessList, types.AccessTuple{
			Address:     address,
			StorageKeys: storageKeys,
		})
	}

	return ethAccessList
}

// DisplayTransaction 显示ETH交易详情
func (es *ETHSigner) DisplayTransaction(transaction *pkg.TransactionData) {
	fmt.Println("\n=== ETH交易详情 ===")
	fmt.Printf("交易ID: %d\n", transaction.ID)
	fmt.Printf("链ID: %s (Ethereum)\n", transaction.ChainID)
	fmt.Printf("Nonce: %d\n", transaction.Nonce)
	fmt.Printf("发送地址: %s\n", transaction.From)
	fmt.Printf("接收地址: %s\n", transaction.To)
	fmt.Printf("交易金额: %s wei\n", transaction.Value)
	fmt.Printf("交易数据: %s\n", transaction.Data)
	fmt.Printf("交易gas limit: %d\n", transaction.Gas)
	fmt.Printf("交易gas tip cap: %s\n", transaction.MaxPriorityFeePerGas)
	fmt.Printf("交易gas fee cap: %s\n", transaction.MaxFeePerGas)

	if len(transaction.AccessList) > 0 {
		fmt.Printf("访问列表: %d 项\n", len(transaction.AccessList))
		for i, item := range transaction.AccessList {
			fmt.Printf("  [%d] 地址: %s, 存储键: %d 个\n", i+1, item.Address, len(item.StorageKeys))
		}
	}
	fmt.Println("==================")
}

// ValidateTransaction 验证ETH交易
func (es *ETHSigner) ValidateTransaction(transaction *pkg.TransactionData) error {
	// TODO: 实现ETH交易验证逻辑
	// 1. 验证地址格式
	// 2. 验证金额格式
	// 3. 验证数据格式
	// 4. 验证AccessList格式

	fmt.Println("⚠️  ETH交易验证功能开发中...")
	return nil
}
