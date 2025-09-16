package script

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"blockChainBrowser/client/signer/internal/utils"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

// ScriptPath 记录脚本生成的地址
type ScriptPath struct {
	ID             string    `json:"id"`
	ScriptID       string    `json:"script_id"`
	Type           string    `json:"type"` // p2sh | p2wsh | p2tr(暂不支持)
	MainnetAddress string    `json:"mainnet_address"`
	TestnetAddress string    `json:"testnet_address"`
	CreatedAt      time.Time `json:"created_at"`
}

// PathManager 管理脚本地址记录
type PathManager struct {
	filePath string
	paths    []ScriptPath
}

// NewPathManager 创建路径管理器
func NewPathManager() *PathManager {
	homeDir, _ := os.UserHomeDir()
	dir := filepath.Join(homeDir, ".blockchain-signer")
	file := filepath.Join(dir, "scripts.path.json")
	_ = os.MkdirAll(dir, 0700)
	pm := &PathManager{filePath: file, paths: []ScriptPath{}}
	_ = pm.Load()
	return pm
}

// Load 读取记录
func (pm *PathManager) Load() error {
	if _, err := os.Stat(pm.filePath); os.IsNotExist(err) {
		return pm.Save()
	}
	b, err := os.ReadFile(pm.filePath)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		pm.paths = []ScriptPath{}
		return nil
	}
	return json.Unmarshal(b, &pm.paths)
}

// Save 保存记录
func (pm *PathManager) Save() error {
	b, err := json.MarshalIndent(pm.paths, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(pm.filePath, b, 0600)
}

// List 返回所有记录
func (pm *PathManager) List() []ScriptPath { return pm.paths }

// Add 添加一条记录
func (pm *PathManager) Add(p ScriptPath) error {
	p.ID = utils.GenerateID()
	p.CreatedAt = time.Now()
	pm.paths = append(pm.paths, p)
	return pm.Save()
}

// AssembleScript 将脚本内容(含操作码/地址/HEX)装配为脚本字节
func AssembleScript(content string) ([]byte, error) {
	builder := txscript.NewScriptBuilder()

	opCodes := map[string]byte{
		"OP_0":                   txscript.OP_0,
		"OP_1":                   txscript.OP_1,
		"OP_2":                   txscript.OP_2,
		"OP_3":                   txscript.OP_3,
		"OP_DUP":                 txscript.OP_DUP,
		"OP_HASH160":             txscript.OP_HASH160,
		"OP_EQUALVERIFY":         txscript.OP_EQUALVERIFY,
		"OP_EQUAL":               txscript.OP_EQUAL,
		"OP_CHECKSIG":            txscript.OP_CHECKSIG,
		"OP_CHECKMULTISIG":       txscript.OP_CHECKMULTISIG,
		"OP_IF":                  txscript.OP_IF,
		"OP_ELSE":                txscript.OP_ELSE,
		"OP_ENDIF":               txscript.OP_ENDIF,
		"OP_CHECKLOCKTIMEVERIFY": txscript.OP_CHECKLOCKTIMEVERIFY,
		"OP_CHECKSEQUENCEVERIFY": txscript.OP_CHECKSEQUENCEVERIFY,
	}

	decodeAny := func(tok string) (btcutil.Address, bool) {
		if a, err := btcutil.DecodeAddress(tok, &chaincfg.MainNetParams); err == nil {
			return a, true
		}
		if a, err := btcutil.DecodeAddress(tok, &chaincfg.TestNet3Params); err == nil {
			return a, true
		}
		return nil, false
	}

	tokens := strings.Fields(content)
	for _, tok := range tokens {
		if op, ok := opCodes[tok]; ok {
			builder.AddOp(op)
			continue
		}

		// HEX 数据
		if utils.IsHexString(tok) {
			hs := tok
			if len(hs) >= 2 && (hs[:2] == "0x" || hs[:2] == "0X") {
				hs = hs[2:]
			}
			data, err := hex.DecodeString(hs)
			if err != nil {
				return nil, fmt.Errorf("HEX解析失败: %v", err)
			}
			builder.AddData(data)
			continue
		}

		// BTC 地址数据（尝试主网/测试网）
		if addr, ok := decodeAny(tok); ok {
			switch a := addr.(type) {
			case *btcutil.AddressPubKeyHash:
				builder.AddData(a.Hash160()[:])
			case *btcutil.AddressScriptHash:
				builder.AddData(a.Hash160()[:])
			case *btcutil.AddressWitnessPubKeyHash:
				builder.AddData(a.WitnessProgram())
			case *btcutil.AddressWitnessScriptHash:
				builder.AddData(a.WitnessProgram())
			default:
				return nil, fmt.Errorf("不支持的地址类型: %T", a)
			}
			continue
		}

		return nil, fmt.Errorf("无法识别的脚本片段: %s", tok)
	}

	return builder.Script()
}

// DeriveAddressesFromContent 根据内容(HEX或可解析脚本)和类型生成主网/测试网地址
func DeriveAddressesFromContent(content string, addrType string) (mainnet string, testnet string, err error) {
	var scriptBytes []byte
	if utils.IsHexString(content) {
		hs := content
		if len(hs) >= 2 && (hs[:2] == "0x" || hs[:2] == "0X") {
			hs = hs[2:]
		}
		b, err := hex.DecodeString(hs)
		if err != nil {
			return "", "", fmt.Errorf("HEX解析失败: %v", err)
		}
		scriptBytes = b
	} else {
		b, err := AssembleScript(content)
		if err != nil {
			return "", "", err
		}
		scriptBytes = b
	}

	switch addrType {
	case "p2sh":
		mn, err := btcutil.NewAddressScriptHash(scriptBytes, &chaincfg.MainNetParams)
		if err != nil {
			return "", "", err
		}
		tn, err := btcutil.NewAddressScriptHash(scriptBytes, &chaincfg.TestNet3Params)
		if err != nil {
			return "", "", err
		}
		return mn.EncodeAddress(), tn.EncodeAddress(), nil
	case "p2wsh":
		h := sha256.Sum256(scriptBytes)
		mn, err := btcutil.NewAddressWitnessScriptHash(h[:], &chaincfg.MainNetParams)
		if err != nil {
			return "", "", err
		}
		tn, err := btcutil.NewAddressWitnessScriptHash(h[:], &chaincfg.TestNet3Params)
		if err != nil {
			return "", "", err
		}
		return mn.EncodeAddress(), tn.EncodeAddress(), nil
	default:
		return "", "", fmt.Errorf("不支持的地址类型: %s", addrType)
	}
}
