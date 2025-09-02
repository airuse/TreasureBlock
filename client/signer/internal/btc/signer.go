package btc

import (
	"blockChainBrowser/client/signer/internal/crypto"
	"blockChainBrowser/client/signer/internal/utils"
	"blockChainBrowser/client/signer/pkg"
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

// BTCSigner BTC签名器
type BTCSigner struct {
	cryptoManager *crypto.CryptoManager
}

// NewBTCSigner 创建BTC签名器
func NewBTCSigner(cryptoManager *crypto.CryptoManager) *BTCSigner {
	return &BTCSigner{
		cryptoManager: cryptoManager,
	}
}

// SignTransaction 签名BTC交易
func (bs *BTCSigner) SignTransaction(transaction *pkg.TransactionData) (string, error) {
	fmt.Println("🟠 开始签名BTC交易...")

	// 1. 根据from地址查找对应的私钥
	keyManager := crypto.NewKeyManager(bs.cryptoManager)
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
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("解析私钥失败: %w", err)
	}

	privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

	// 2. 构建BTC交易结构
	tx, err := bs.buildTransaction(transaction)
	if err != nil {
		return "", fmt.Errorf("构建交易失败: %w", err)
	}

	// 3. 使用私钥签名交易
	signedTx, err := bs.signTransaction(tx, privateKey, transaction)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %w", err)
	}

	// 4. 返回完整的签名交易
	return bs.serializeTransaction(signedTx), nil
}

// buildTransaction 构建BTC交易结构
func (bs *BTCSigner) buildTransaction(transaction *pkg.TransactionData) (*wire.MsgTx, error) {
	// 创建新交易
	tx := wire.NewMsgTx(wire.TxVersion)

	// 解析交易金额 (从satoshi转换为BTC)
	value, err := bs.parseValue(transaction.Value)
	if err != nil {
		return nil, fmt.Errorf("解析交易金额失败: %w", err)
	}

	// 添加输出
	toAddress, err := btcutil.DecodeAddress(transaction.To, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("解析接收地址失败: %w", err)
	}

	pkScript, err := txscript.PayToAddrScript(toAddress)
	if err != nil {
		return nil, fmt.Errorf("创建输出脚本失败: %w", err)
	}

	tx.AddTxOut(wire.NewTxOut(value, pkScript))

	// 注意：这里简化了输入处理，实际应用中需要处理UTXO
	// 添加一个占位符输入
	prevTxHash, _ := chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	tx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(prevTxHash, 0), nil, nil))

	return tx, nil
}

// signTransaction 签名BTC交易
func (bs *BTCSigner) signTransaction(tx *wire.MsgTx, privateKey *btcec.PrivateKey, transaction *pkg.TransactionData) (*wire.MsgTx, error) {
	// 获取发送地址
	fromAddress, err := btcutil.DecodeAddress(transaction.From, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("解析发送地址失败: %w", err)
	}

	// 创建P2PKH脚本
	pkScript, err := txscript.PayToAddrScript(fromAddress)
	if err != nil {
		return nil, fmt.Errorf("创建输入脚本失败: %w", err)
	}

	// 签名第一个输入
	sigScript, err := txscript.SignatureScript(tx, 0, pkScript, txscript.SigHashAll, privateKey, true)
	if err != nil {
		return nil, fmt.Errorf("创建签名脚本失败: %w", err)
	}

	tx.TxIn[0].SignatureScript = sigScript

	return tx, nil
}

// parseValue 解析交易金额
func (bs *BTCSigner) parseValue(valueStr string) (int64, error) {
	// 移除0x前缀
	if strings.HasPrefix(valueStr, "0x") {
		valueStr = valueStr[2:]
	}

	// 解析为十六进制
	value, err := strconv.ParseInt(valueStr, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("无效的十六进制金额: %s", valueStr)
	}

	return value, nil
}

// serializeTransaction 序列化交易
func (bs *BTCSigner) serializeTransaction(tx *wire.MsgTx) string {
	var buf bytes.Buffer
	tx.Serialize(&buf)
	return hex.EncodeToString(buf.Bytes())
}

// DisplayTransaction 显示BTC交易详情
func (bs *BTCSigner) DisplayTransaction(transaction *pkg.TransactionData) {
	fmt.Println("\n=== BTC交易详情 ===")
	fmt.Printf("交易ID: %d\n", transaction.ID)
	fmt.Printf("链ID: %s (Bitcoin)\n", transaction.ChainID)
	fmt.Printf("发送地址: %s\n", transaction.From)
	fmt.Printf("接收地址: %s\n", transaction.To)
	fmt.Printf("交易金额: %s satoshi\n", transaction.Value)
	fmt.Println("==================")
}

// ValidateTransaction 验证BTC交易
func (bs *BTCSigner) ValidateTransaction(transaction *pkg.TransactionData) error {
	// TODO: 实现BTC交易验证逻辑
	// 1. 验证地址格式
	// 2. 验证金额格式
	// 3. 验证UTXO

	fmt.Println("⚠️  BTC交易验证功能开发中...")
	return nil
}
