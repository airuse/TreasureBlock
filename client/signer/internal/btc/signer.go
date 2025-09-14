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

	// 1. 加载密钥管理器
	keyManager := crypto.NewKeyManager(bs.cryptoManager)
	if err := keyManager.LoadKeys(); err != nil {
		return "", fmt.Errorf("加载私钥失败: %w", err)
	}

	// 2. 智能地址匹配和私钥选择
	selectedKeys, err := bs.selectKeysForTransaction(transaction, keyManager)
	if err != nil {
		return "", fmt.Errorf("选择私钥失败: %w", err)
	}

	// 3. 构建BTC交易结构
	tx, err := bs.buildTransactionFromMsgTx(transaction)
	if err != nil {
		return "", fmt.Errorf("构建交易失败: %w", err)
	}

	// 4. 使用选中的私钥签名交易
	signedTx, err := bs.signTransactionWithKeys(tx, selectedKeys, transaction)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %w", err)
	}

	// 5. 返回完整的签名交易
	return bs.serializeTransaction(signedTx), nil
}

// buildTransactionFromMsgTx 从MsgTx构建BTC交易结构
func (bs *BTCSigner) buildTransactionFromMsgTx(transaction *pkg.TransactionData) (*wire.MsgTx, error) {
	if transaction.MsgTx == nil {
		return nil, fmt.Errorf("MsgTx数据为空")
	}

	// 创建新交易
	tx := wire.NewMsgTx(wire.TxVersion)
	tx.Version = transaction.MsgTx.Version
	tx.LockTime = transaction.MsgTx.LockTime

	// 添加交易输入
	for _, txIn := range transaction.MsgTx.TxIn {
		prevTxHash, err := chainhash.NewHashFromStr(txIn.Txid)
		if err != nil {
			return nil, fmt.Errorf("解析前一个交易哈希失败: %w", err)
		}

		outPoint := wire.NewOutPoint(prevTxHash, uint32(txIn.Vout))
		newTxIn := wire.NewTxIn(outPoint, nil, nil)
		newTxIn.Sequence = txIn.Sequence
		tx.AddTxIn(newTxIn)
	}

	// 添加交易输出
	for _, txOut := range transaction.MsgTx.TxOut {
		// 根据地址类型创建输出脚本
		var pkScript []byte
		var err error

		// 尝试解析地址（支持主网和测试网）
		address, err := bs.decodeAddress(txOut.Address)
		if err != nil {
			return nil, fmt.Errorf("解析输出地址失败 %s: %w", txOut.Address, err)
		}

		pkScript, err = txscript.PayToAddrScript(address)
		if err != nil {
			return nil, fmt.Errorf("创建输出脚本失败: %w", err)
		}

		tx.AddTxOut(wire.NewTxOut(txOut.ValueSatoshi, pkScript))
	}

	return tx, nil
}

// SelectedKey 选中的私钥信息
type SelectedKey struct {
	KeyInfo    *crypto.KeyInfo
	PrivateKey *btcec.PrivateKey
	Address    string
}

// selectKeysForTransaction 智能选择用于交易的私钥
func (bs *BTCSigner) selectKeysForTransaction(transaction *pkg.TransactionData, keyManager *crypto.KeyManager) ([]*SelectedKey, error) {
	fromAddress := transaction.Address
	if fromAddress == "" {
		fromAddress = transaction.From
	}

	// 1. 尝试自动匹配地址
	if keyManager.HasKey(fromAddress) {
		fmt.Printf("✅ 自动找到地址 %s 对应的私钥\n", fromAddress)
		return bs.selectSingleKey(fromAddress, keyManager)
	}

	// 2. 如果自动匹配失败，显示所有可用的BTC私钥供用户选择
	fmt.Printf("❌ 未找到地址 %s 对应的私钥\n", fromAddress)
	fmt.Println("请从以下可用的BTC私钥中选择:")

	btcKeys := keyManager.GetKeysByChain("btc")
	if len(btcKeys) == 0 {
		return nil, fmt.Errorf("没有可用的BTC私钥")
	}

	// 显示可用的私钥
	for i, key := range btcKeys {
		fmt.Printf("%d. 私钥ID: %s\n", i+1, key.KeyID[:8]+"...")
		fmt.Printf("   地址: %v\n", key.Addresses)
		fmt.Printf("   描述: %s\n", key.Description)
		fmt.Printf("   创建时间: %s\n", key.CreatedAt)
		fmt.Println("   " + strings.Repeat("-", 50))
	}

	// 获取用户选择
	fmt.Print("请选择私钥编号 (用逗号分隔多个私钥，如: 1,3,5): ")
	var selection string
	fmt.Scanln(&selection)

	selectedIndices := bs.parseSelection(selection, len(btcKeys))
	if len(selectedIndices) == 0 {
		return nil, fmt.Errorf("未选择任何私钥")
	}

	// 返回选中的私钥
	var selectedKeys []*SelectedKey
	for _, idx := range selectedIndices {
		key := btcKeys[idx]
		selectedKey, err := bs.selectKeyByInfo(key, keyManager)
		if err != nil {
			return nil, fmt.Errorf("选择私钥失败: %w", err)
		}
		selectedKeys = append(selectedKeys, selectedKey)
	}

	return selectedKeys, nil
}

// selectSingleKey 选择单个私钥
func (bs *BTCSigner) selectSingleKey(address string, keyManager *crypto.KeyManager) ([]*SelectedKey, error) {
	// 获取解密密码
	password, err := utils.ReadPassword("请输入私钥解密密码: ")
	if err != nil {
		return nil, fmt.Errorf("读取密码失败: %w", err)
	}

	// 解密私钥
	privateKeyHex, err := keyManager.GetKey(address, password)
	if err != nil {
		return nil, fmt.Errorf("解密私钥失败: %w", err)
	}

	// 解析私钥
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

	// 获取KeyInfo
	keyInfo := keyManager.FindByAddress(address)
	if keyInfo == nil {
		return nil, fmt.Errorf("未找到私钥信息")
	}

	return []*SelectedKey{{
		KeyInfo:    keyInfo,
		PrivateKey: privateKey,
		Address:    address,
	}}, nil
}

// selectKeyByInfo 根据KeyInfo选择私钥
func (bs *BTCSigner) selectKeyByInfo(keyInfo *crypto.KeyInfo, keyManager *crypto.KeyManager) (*SelectedKey, error) {
	// 获取解密密码
	password, err := utils.ReadPassword(fmt.Sprintf("请输入私钥 %s 的解密密码: ", keyInfo.KeyID[:8]+"..."))
	if err != nil {
		return nil, fmt.Errorf("读取密码失败: %w", err)
	}

	// 解密私钥
	privateKey, err := keyManager.GetCryptoManager().DecryptPrivateKey(keyInfo.EncryptedKey, password)
	if err != nil {
		return nil, fmt.Errorf("解密私钥失败: %w", err)
	}

	// 解析私钥
	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	btcPrivateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

	// 使用第一个地址作为主地址
	mainAddress := keyInfo.Addresses[0]

	return &SelectedKey{
		KeyInfo:    keyInfo,
		PrivateKey: btcPrivateKey,
		Address:    mainAddress,
	}, nil
}

// parseSelection 解析用户选择的私钥编号
func (bs *BTCSigner) parseSelection(selection string, maxCount int) []int {
	var indices []int
	parts := strings.Split(selection, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if idx, err := strconv.Atoi(part); err == nil && idx >= 1 && idx <= maxCount {
			indices = append(indices, idx-1) // 转换为0基索引
		}
	}
	return indices
}

// signTransactionWithKeys 使用选中的私钥签名交易
func (bs *BTCSigner) signTransactionWithKeys(tx *wire.MsgTx, selectedKeys []*SelectedKey, transaction *pkg.TransactionData) (*wire.MsgTx, error) {
	// 对于简单的P2PKH或P2WPKH，使用第一个私钥签名所有输入
	if len(selectedKeys) == 1 {
		return bs.signTransactionWithSingleKey(tx, selectedKeys[0], transaction)
	}

	// 对于复杂的多签脚本，需要更复杂的签名逻辑
	// 这里简化处理，使用第一个私钥签名所有输入
	fmt.Println("⚠️  多私钥签名功能开发中，使用第一个私钥签名所有输入")
	return bs.signTransactionWithSingleKey(tx, selectedKeys[0], transaction)
}

// signTransactionWithSingleKey 使用单个私钥签名交易
func (bs *BTCSigner) signTransactionWithSingleKey(tx *wire.MsgTx, selectedKey *SelectedKey, transaction *pkg.TransactionData) (*wire.MsgTx, error) {
	// 获取发送地址
	fromAddress := transaction.Address
	if fromAddress == "" {
		fromAddress = transaction.From
	}

	address, err := bs.decodeAddress(fromAddress)
	if err != nil {
		return nil, fmt.Errorf("解析发送地址失败: %w", err)
	}

	// 创建输出脚本
	pkScript, err := txscript.PayToAddrScript(address)
	if err != nil {
		return nil, fmt.Errorf("创建输入脚本失败: %w", err)
	}

	// 签名所有输入
	for i := range tx.TxIn {
		sigScript, err := txscript.SignatureScript(tx, i, pkScript, txscript.SigHashAll, selectedKey.PrivateKey, true)
		if err != nil {
			return nil, fmt.Errorf("创建签名脚本失败 (输入 %d): %w", i, err)
		}

		tx.TxIn[i].SignatureScript = sigScript
	}

	return tx, nil
}

// decodeAddress 解码BTC地址（支持主网和测试网）
func (bs *BTCSigner) decodeAddress(address string) (btcutil.Address, error) {
	// 尝试主网
	if addr, err := btcutil.DecodeAddress(address, &chaincfg.MainNetParams); err == nil {
		return addr, nil
	}

	// 尝试测试网
	if addr, err := btcutil.DecodeAddress(address, &chaincfg.TestNet3Params); err == nil {
		return addr, nil
	}

	return nil, fmt.Errorf("无法解析地址: %s", address)
}

// buildTransaction 构建BTC交易结构（保留用于向后兼容）
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
	fromAddress := transaction.Address
	if fromAddress == "" {
		fromAddress = transaction.From
	}

	address, err := bs.decodeAddress(fromAddress)
	if err != nil {
		return nil, fmt.Errorf("解析发送地址失败: %w", err)
	}

	// 创建输出脚本
	pkScript, err := txscript.PayToAddrScript(address)
	if err != nil {
		return nil, fmt.Errorf("创建输入脚本失败: %w", err)
	}

	// 签名所有输入
	for i := range tx.TxIn {
		sigScript, err := txscript.SignatureScript(tx, i, pkScript, txscript.SigHashAll, privateKey, true)
		if err != nil {
			return nil, fmt.Errorf("创建签名脚本失败 (输入 %d): %w", i, err)
		}

		tx.TxIn[i].SignatureScript = sigScript
	}

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
	fmt.Printf("链类型: %s (Bitcoin)\n", transaction.Type)
	fmt.Printf("发送地址: %s\n", transaction.Address)

	if transaction.MsgTx != nil {
		fmt.Printf("交易版本: %d\n", transaction.MsgTx.Version)
		fmt.Printf("锁定时间: %d\n", transaction.MsgTx.LockTime)

		fmt.Printf("\n交易输入 (%d个):\n", len(transaction.MsgTx.TxIn))
		for i, txIn := range transaction.MsgTx.TxIn {
			fmt.Printf("  %d. 前交易: %s, 输出索引: %d, 序列号: %d\n",
				i+1, txIn.Txid, txIn.Vout, txIn.Sequence)
		}

		fmt.Printf("\n交易输出 (%d个):\n", len(transaction.MsgTx.TxOut))
		totalOutput := int64(0)
		for i, txOut := range transaction.MsgTx.TxOut {
			fmt.Printf("  %d. 地址: %s, 金额: %d satoshi (%.8f BTC)\n",
				i+1, txOut.Address, txOut.ValueSatoshi, float64(txOut.ValueSatoshi)/1e8)
			totalOutput += txOut.ValueSatoshi
		}
		fmt.Printf("\n总输出金额: %d satoshi (%.8f BTC)\n", totalOutput, float64(totalOutput)/1e8)
	}

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
