package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

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
func DeriveBTCAddresses(privateKeyHex string) (p2wpkh string, p2wsh string, p2pkh string, p2sh string, err error) {
	privBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", "", "", "", fmt.Errorf("解析BTC私钥失败: %w", err)
	}
	priv, pub := btcsuitePrivFromBytes(privBytes)
	_ = priv // 私钥不直接使用

	compressedPub := pub.SerializeCompressed()
	pubKeyHash := btcutil.Hash160(compressedPub)

	// P2PKH (1开头)
	addrP2PKH, err := btcutil.NewAddressPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", "", fmt.Errorf("生成P2PKH地址失败: %w", err)
	}

	// P2WPKH (bc1q开头)
	addrWPKH, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", "", fmt.Errorf("生成P2WPKH地址失败: %w", err)
	}

	// P2WSH: witness script = <pubkey> OP_CHECKSIG
	p2pkScript, err := txscript.NewScriptBuilder().AddData(compressedPub).AddOp(txscript.OP_CHECKSIG).Script()
	if err != nil {
		return "", "", "", "", fmt.Errorf("生成P2PK脚本失败: %w", err)
	}
	wsh := sha256.Sum256(p2pkScript)
	addrWSH, err := btcutil.NewAddressWitnessScriptHash(wsh[:], &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", "", fmt.Errorf("生成P2WSH地址失败: %w", err)
	}

	// P2SH-P2WPKH (3开头): 对WPKH地址的见证程序做P2SH包装
	redeemScript, err := txscript.PayToAddrScript(addrWPKH)
	if err != nil {
		return "", "", "", "", fmt.Errorf("生成P2SH赎回脚本失败: %w", err)
	}
	redeemHash := btcutil.Hash160(redeemScript)
	addrP2SH, err := btcutil.NewAddressScriptHashFromHash(redeemHash, &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", "", fmt.Errorf("生成P2SH地址失败: %w", err)
	}

	return addrWPKH.EncodeAddress(), addrWSH.EncodeAddress(), addrP2PKH.EncodeAddress(), addrP2SH.EncodeAddress(), nil
}

// 使用btcec从私钥字节获取密钥对
func btcsuitePrivFromBytes(b []byte) (*btcec.PrivateKey, *btcec.PublicKey) {
	priv, pub := btcec.PrivKeyFromBytes(b)
	return priv, pub
}
