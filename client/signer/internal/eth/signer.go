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
	"github.com/ethereum/go-ethereum/rlp"
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

	// 获取解密密码（隐藏输入）
	password, err := utils.ReadPassword("请输入私钥解密密码: ")
	if err != nil {
		return "", fmt.Errorf("读取密码失败: %w", err)
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

	// 3. 使用私钥签名交易
	signer := types.NewLondonSigner(big.NewInt(1)) // 使用EIP-1559签名器
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %w", err)
	}

	// 4. 返回完整的签名交易
	rawTx, err := rlp.EncodeToBytes(signedTx)
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
	data, err := hex.DecodeString(strings.TrimPrefix(transaction.Data, "0x"))
	if err != nil {
		return nil, fmt.Errorf("解析交易数据失败: %w", err)
	}

	// 构建AccessList
	accessList := es.buildAccessList(transaction.AccessList)

	// 创建EIP-1559交易
	tx := &types.DynamicFeeTx{
		ChainID:    big.NewInt(chainID),
		Nonce:      transaction.Nonce,
		GasTipCap:  big.NewInt(0), // 设置为0，实际应用中应该从网络获取
		GasFeeCap:  big.NewInt(0), // 设置为0，实际应用中应该从网络获取
		Gas:        21000,         // 默认gas limit
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
