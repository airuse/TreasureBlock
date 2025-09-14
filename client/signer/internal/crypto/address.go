package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// DeriveETHAddresses 从私钥派生ETH地址（小写与EIP-55校验两种）
func DeriveETHAddresses(privateKeyHex string) (lowercase string, checksummed string, err error) {
	priv, err := ethcrypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", "", fmt.Errorf("解析ETH私钥失败: %w", err)
	}
	addr := ethcrypto.PubkeyToAddress(priv.PublicKey)
	checksummed = addr.Hex()
	lowercase = "0x" + common.Bytes2Hex(addr.Bytes())
	return lowercase, checksummed, nil
}

// DeriveBTCAddresses 从私钥派生常见BTC地址（P2WPKH, P2WSH, P2PKH, P2SH-P2WPKH）
func DeriveBTCAddresses(privateKeyHex string, networkType string) (p2wpkh string, p2wsh string, p2pkh string, p2sh string, err error) {
	privBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", "", "", "", fmt.Errorf("解析BTC私钥失败: %w", err)
	}
	priv, pub := btcsuitePrivFromBytes(privBytes)
	_ = priv // 私钥不直接使用

	compressedPub := pub.SerializeCompressed()
	pubKeyHash := btcutil.Hash160(compressedPub)

	// 选择网络参数
	var params *chaincfg.Params
	switch networkType {
	case "testnet":
		params = &chaincfg.TestNet3Params
	default:
		params = &chaincfg.MainNetParams
	}

	// P2PKH (1开头)
	addrP2PKH, err := btcutil.NewAddressPubKeyHash(pubKeyHash, params)
	if err != nil {
		return "", "", "", "", fmt.Errorf("生成P2PKH地址失败: %w", err)
	}

	// P2WPKH (bc1q开头)
	addrWPKH, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, params)
	if err != nil {
		return "", "", "", "", fmt.Errorf("生成P2WPKH地址失败: %w", err)
	}

	// P2WSH: witness script = <pubkey> OP_CHECKSIG
	p2pkScript, err := txscript.NewScriptBuilder().AddData(compressedPub).AddOp(txscript.OP_CHECKSIG).Script()
	if err != nil {
		return "", "", "", "", fmt.Errorf("生成P2PK脚本失败: %w", err)
	}
	wsh := sha256.Sum256(p2pkScript)
	addrWSH, err := btcutil.NewAddressWitnessScriptHash(wsh[:], params)
	if err != nil {
		return "", "", "", "", fmt.Errorf("生成P2WSH地址失败: %w", err)
	}

	// P2SH-P2WPKH (3开头): 对WPKH地址的见证程序做P2SH包装
	redeemScript, err := txscript.PayToAddrScript(addrWPKH)
	if err != nil {
		return "", "", "", "", fmt.Errorf("生成P2SH赎回脚本失败: %w", err)
	}
	redeemHash := btcutil.Hash160(redeemScript)
	addrP2SH, err := btcutil.NewAddressScriptHashFromHash(redeemHash, params)
	if err != nil {
		return "", "", "", "", fmt.Errorf("生成P2SH地址失败: %w", err)
	}

	return addrWPKH.EncodeAddress(), addrWSH.EncodeAddress(), addrP2PKH.EncodeAddress(), addrP2SH.EncodeAddress(), nil
}

// BTCAddresses 包含所有BTC地址类型
type BTCAddresses struct {
	P2PKH  string
	P2WPKH string
	P2WSH  string
	P2SH   string
}

// DeriveAllBTCAddresses 从私钥派生所有BTC地址类型
func DeriveAllBTCAddresses(privateKeyHex string, networkType string) (*BTCAddresses, error) {
	privBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("解析BTC私钥失败: %w", err)
	}
	priv, pub := btcsuitePrivFromBytes(privBytes)
	_ = priv // 私钥不直接使用

	compressedPub := pub.SerializeCompressed()
	pubKeyHash := btcutil.Hash160(compressedPub)

	// 选择网络参数
	var params *chaincfg.Params
	switch networkType {
	case "testnet":
		params = &chaincfg.TestNet3Params
	default:
		params = &chaincfg.MainNetParams
	}

	// P2PKH (1开头)
	addrP2PKH, err := btcutil.NewAddressPubKeyHash(pubKeyHash, params)
	if err != nil {
		return nil, fmt.Errorf("生成P2PKH地址失败: %w", err)
	}

	// P2WPKH (bc1q开头)
	addrWPKH, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, params)
	if err != nil {
		return nil, fmt.Errorf("生成P2WPKH地址失败: %w", err)
	}

	// P2WSH: witness script = <pubkey> OP_CHECKSIG
	p2pkScript, err := txscript.NewScriptBuilder().AddData(compressedPub).AddOp(txscript.OP_CHECKSIG).Script()
	if err != nil {
		return nil, fmt.Errorf("生成P2PK脚本失败: %w", err)
	}
	wsh := sha256.Sum256(p2pkScript)
	addrWSH, err := btcutil.NewAddressWitnessScriptHash(wsh[:], params)
	if err != nil {
		return nil, fmt.Errorf("生成P2WSH地址失败: %w", err)
	}

	// P2SH-P2WPKH (3开头): 对WPKH地址的见证程序做P2SH包装
	redeemScript, err := txscript.PayToAddrScript(addrWPKH)
	if err != nil {
		return nil, fmt.Errorf("生成P2SH赎回脚本失败: %w", err)
	}
	redeemHash := btcutil.Hash160(redeemScript)
	addrP2SH, err := btcutil.NewAddressScriptHashFromHash(redeemHash, params)
	if err != nil {
		return nil, fmt.Errorf("生成P2SH地址失败: %w", err)
	}

	return &BTCAddresses{
		P2PKH:  addrP2PKH.EncodeAddress(),
		P2WPKH: addrWPKH.EncodeAddress(),
		P2WSH:  addrWSH.EncodeAddress(),
		P2SH:   addrP2SH.EncodeAddress(),
	}, nil
}

// DeriveCustomBTCAddresses 从私钥和自定义脚本派生BTC地址
func DeriveCustomBTCAddresses(privateKeyHex string, customScript string, networkType string) (p2sh string, p2wsh string, err error) {
	privBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", "", fmt.Errorf("解析BTC私钥失败: %w", err)
	}
	priv, pub := btcsuitePrivFromBytes(privBytes)
	_ = priv // 私钥不直接使用

	// 解析自定义脚本
	script, err := parseCustomScript(customScript, pub)
	if err != nil {
		return "", "", fmt.Errorf("解析自定义脚本失败: %w", err)
	}

	// 选择网络参数
	var params *chaincfg.Params
	switch networkType {
	case "testnet":
		params = &chaincfg.TestNet3Params
	default:
		params = &chaincfg.MainNetParams
	}

	// 生成P2SH地址
	scriptHash := btcutil.Hash160(script)
	addrP2SH, err := btcutil.NewAddressScriptHashFromHash(scriptHash, params)
	if err != nil {
		return "", "", fmt.Errorf("生成P2SH地址失败: %w", err)
	}

	// 生成P2WSH地址
	wsh := sha256.Sum256(script)
	addrP2WSH, err := btcutil.NewAddressWitnessScriptHash(wsh[:], params)
	if err != nil {
		return "", "", fmt.Errorf("生成P2WSH地址失败: %w", err)
	}

	return addrP2SH.EncodeAddress(), addrP2WSH.EncodeAddress(), nil
}

// GenerateScriptAddresses 生成脚本地址
func GenerateScriptAddresses(customScript string, selectedKeys []*KeyInfo, scriptType string) (string, string, error) {
	if len(selectedKeys) == 0 {
		return "", "", fmt.Errorf("没有选择私钥")
	}

	// 解析自定义脚本，替换占位符
	scriptBytes, err := parseCustomScriptWithKeys(customScript, selectedKeys)
	if err != nil {
		return "", "", fmt.Errorf("解析脚本失败: %w", err)
	}

	// 根据脚本类型生成地址
	switch scriptType {
	case "p2sh":
		return generateP2SHAddresses(scriptBytes)
	case "p2wsh":
		return generateP2WSHAddresses(scriptBytes)
	case "p2tr":
		return generateP2TRAddresses(scriptBytes)
	default:
		return "", "", fmt.Errorf("不支持的脚本类型: %s", scriptType)
	}
}

// parseCustomScriptWithKeys 解析自定义脚本并替换占位符
func parseCustomScriptWithKeys(scriptStr string, selectedKeys []*KeyInfo) ([]byte, error) {
	// 暂时实现一个简单的脚本生成逻辑
	// 根据脚本类型生成不同的脚本

	// 简单的2-of-3多签脚本示例
	if strings.Contains(scriptStr, "OP_2") && strings.Contains(scriptStr, "OP_3") && strings.Contains(scriptStr, "OP_CHECKMULTISIG") {
		// 生成2-of-3多签脚本
		script := []byte{
			0x52, // OP_2
			0x21, // 33字节公钥长度
			// 这里应该是真实的公钥，暂时使用占位符
			0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11,
			0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20, 0x21, 0x22,
			0x21, // 33字节公钥长度
			0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11,
			0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20, 0x21, 0x22,
			0x21, // 33字节公钥长度
			0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11,
			0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20, 0x21, 0x22,
			0x53, // OP_3
			0xae, // OP_CHECKMULTISIG
		}
		return script, nil
	}

	// 简单的P2PK脚本
	if strings.Contains(scriptStr, "OP_CHECKSIG") && !strings.Contains(scriptStr, "OP_2") {
		script := []byte{
			0x21, // 33字节公钥长度
			// 这里应该是真实的公钥，暂时使用占位符
			0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11,
			0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20, 0x21, 0x22,
			0xac, // OP_CHECKSIG
		}
		return script, nil
	}

	// 默认返回一个简单的脚本
	return []byte{0x51, 0x52, 0x53}, nil
}

// getPublicKeyFromKeyInfo 从私钥信息中获取公钥
func getPublicKeyFromKeyInfo(keyInfo *KeyInfo) (*btcec.PublicKey, error) {
	// 这里需要解密私钥并获取公钥
	// 暂时返回一个占位符公钥
	// 实际实现需要解密私钥，然后生成公钥
	return nil, fmt.Errorf("公钥获取功能待实现")
}

// generateP2SHAddresses 生成P2SH地址
func generateP2SHAddresses(scriptBytes []byte) (string, string, error) {
	// 主网P2SH
	mainnetParams := &chaincfg.MainNetParams
	mainnetAddr, err := btcutil.NewAddressScriptHash(scriptBytes, mainnetParams)
	if err != nil {
		return "", "", fmt.Errorf("生成主网P2SH地址失败: %w", err)
	}

	// 测试网P2SH
	testnetParams := &chaincfg.TestNet3Params
	testnetAddr, err := btcutil.NewAddressScriptHash(scriptBytes, testnetParams)
	if err != nil {
		return "", "", fmt.Errorf("生成测试网P2SH地址失败: %w", err)
	}

	return mainnetAddr.EncodeAddress(), testnetAddr.EncodeAddress(), nil
}

// generateP2WSHAddresses 生成P2WSH地址
func generateP2WSHAddresses(scriptBytes []byte) (string, string, error) {
	// 计算脚本的SHA256哈希
	scriptHash := sha256.Sum256(scriptBytes)

	// 主网P2WSH
	mainnetParams := &chaincfg.MainNetParams
	mainnetAddr, err := btcutil.NewAddressWitnessScriptHash(scriptHash[:], mainnetParams)
	if err != nil {
		return "", "", fmt.Errorf("生成主网P2WSH地址失败: %w", err)
	}

	// 测试网P2WSH
	testnetParams := &chaincfg.TestNet3Params
	testnetAddr, err := btcutil.NewAddressWitnessScriptHash(scriptHash[:], testnetParams)
	if err != nil {
		return "", "", fmt.Errorf("生成测试网P2WSH地址失败: %w", err)
	}

	return mainnetAddr.EncodeAddress(), testnetAddr.EncodeAddress(), nil
}

// generateP2TRAddresses 生成P2TR地址
func generateP2TRAddresses(scriptBytes []byte) (string, string, error) {
	// P2TR地址生成比较复杂，暂时返回占位符
	return "p2tr_mainnet_placeholder", "p2tr_testnet_placeholder", nil
}

// parseCustomScript 解析自定义脚本
func parseCustomScript(scriptStr string, pub *btcec.PublicKey) ([]byte, error) {
	// 支持的操作码映射
	opCodes := map[string]byte{
		"OP_DUP":                 0x76,
		"OP_HASH160":             0xa9,
		"OP_EQUALVERIFY":         0x88,
		"OP_CHECKSIG":            0xac,
		"OP_EQUAL":               0x87,
		"OP_0":                   0x00,
		"OP_1":                   0x51,
		"OP_2":                   0x52,
		"OP_3":                   0x53,
		"OP_CHECKMULTISIG":       0xae,
		"OP_IF":                  0x63,
		"OP_ELSE":                0x67,
		"OP_ENDIF":               0x68,
		"OP_CHECKLOCKTIMEVERIFY": 0xb1,
		"OP_CHECKSEQUENCEVERIFY": 0xb2,
	}

	// 如果脚本包含 <pubkey> 占位符，替换为实际公钥
	compressedPub := pub.SerializeCompressed()
	scriptStr = strings.ReplaceAll(scriptStr, "<pubkey>", hex.EncodeToString(compressedPub))

	// 如果脚本包含 <pubkeyhash> 占位符，替换为公钥哈希
	pubKeyHash := btcutil.Hash160(compressedPub)
	scriptStr = strings.ReplaceAll(scriptStr, "<pubkeyhash>", hex.EncodeToString(pubKeyHash))

	// 解析脚本
	var script []byte
	parts := strings.Fields(scriptStr)

	for _, part := range parts {
		// 检查是否是操作码
		if opCode, exists := opCodes[part]; exists {
			script = append(script, opCode)
		} else if strings.HasPrefix(part, "<") && strings.HasSuffix(part, ">") {
			// 处理占位符（已在上面处理）
			continue
		} else if len(part) > 0 {
			// 处理十六进制数据
			if hexData, err := hex.DecodeString(part); err == nil {
				// 添加长度前缀
				script = append(script, byte(len(hexData)))
				script = append(script, hexData...)
			} else {
				return nil, fmt.Errorf("无法解析脚本部分: %s", part)
			}
		}
	}

	return script, nil
}

// 使用btcec从私钥字节获取密钥对
func btcsuitePrivFromBytes(b []byte) (*btcec.PrivateKey, *btcec.PublicKey) {
	priv, pub := btcec.PrivKeyFromBytes(b)
	return priv, pub
}
