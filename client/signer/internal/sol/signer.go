package sol

import (
	"blockChainBrowser/client/signer/internal/crypto"
	"blockChainBrowser/client/signer/internal/utils"
	"blockChainBrowser/client/signer/pkg"
	"encoding/base64"
	"errors"
	"fmt"

	solcommon "github.com/blocto/solana-go-sdk/common"
	solsys "github.com/blocto/solana-go-sdk/program/system"
	soltypes "github.com/blocto/solana-go-sdk/types"
)

// SOLSigner Solana 签名器
type SOLSigner struct {
	cryptoManager *crypto.CryptoManager
}

func NewSOLSigner(cryptoManager *crypto.CryptoManager) *SOLSigner {
	return &SOLSigner{cryptoManager: cryptoManager}
}

// SignTransaction 组装并签名 SOL 交易
// 输入为从二维码解析的 TransactionData 或直接的 SolanaUnsigned JSON
func (ss *SOLSigner) SignTransaction(unsigned *pkg.SolanaUnsigned) (string, error) {
	fmt.Println("🟣 开始签名SOL交易...")
	if unsigned == nil {
		return "", errors.New("未提供未签名数据")
	}
	if unsigned.FeePayer == "" {
		return "", errors.New("缺少 fee_payer")
	}

	// 加载并匹配 fee_payer 私钥
	keyManager := crypto.NewKeyManager(ss.cryptoManager)
	if err := keyManager.LoadKeys(); err != nil {
		return "", fmt.Errorf("加载私钥失败: %w", err)
	}
	if !keyManager.HasKey(unsigned.FeePayer) {
		return "", fmt.Errorf("未找到地址 %s 对应的私钥", unsigned.FeePayer)
	}
	password, err := utils.ReadPassword("请确认此交易并输入私钥解密密码（无密码回车则视为取消）: ")
	if err != nil {
		return "", fmt.Errorf("读取密码失败: %w", err)
	}
	if password == "" {
		return "", fmt.Errorf("操作已取消")
	}
	privateKeyHex, err := keyManager.GetKey(unsigned.FeePayer, password)
	if err != nil {
		return "", fmt.Errorf("解密私钥失败: %w", err)
	}
	// 此处要求 keys.json 中存储 base64 的 Solana 私钥（64字节）
	raw, berr := base64.StdEncoding.DecodeString(privateKeyHex)
	if berr != nil {
		return "", fmt.Errorf("无法解析Solana私钥(需base64): %w", berr)
	}
	if len(raw) != 64 {
		return "", fmt.Errorf("Solana私钥长度应为64字节(base64解码后)，当前=%d", len(raw))
	}
	kp, err := soltypes.AccountFromBytes(raw)
	if err != nil {
		return "", fmt.Errorf("构造Solana账户失败: %w", err)
	}

	// 构建交易消息
	var ins []soltypes.Instruction
	for _, plan := range unsigned.Instructions {
		t, _ := plan["type"].(string)
		switch t {
		case "system_transfer":
			params, _ := plan["params"].(map[string]interface{})
			from := solcommon.PublicKeyFromString(fmt.Sprintf("%v", params["from"]))
			to := solcommon.PublicKeyFromString(fmt.Sprintf("%v", params["to"]))
			lamportsStr := fmt.Sprintf("%v", params["lamports"])
			var lamports uint64
			fmt.Sscanf(lamportsStr, "%d", &lamports)
			ix := solsys.Transfer(solsys.TransferParam{From: from, To: to, Amount: lamports})
			ins = append(ins, ix)
		default:
			return "", fmt.Errorf("不支持的指令类型: %s", t)
		}
	}

	msg := soltypes.NewMessage(soltypes.NewMessageParam{
		FeePayer:        solcommon.PublicKeyFromString(unsigned.FeePayer),
		RecentBlockhash: unsigned.RecentBlockhash,
		Instructions:    ins,
	})

	// 签名
	tx, err := soltypes.NewTransaction(soltypes.NewTransactionParam{
		Message: msg,
		Signers: []soltypes.Account{kp},
	})
	if err != nil {
		return "", fmt.Errorf("构建交易失败: %w", err)
	}

	// 编码为 base64，供后端广播
	raw, err = tx.Serialize()
	if err != nil {
		return "", fmt.Errorf("序列化交易失败: %w", err)
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}

// Display 显示 SOL 交易信息（基于 unsigned 计划）
func (ss *SOLSigner) Display(unsigned *pkg.SolanaUnsigned) {
	fmt.Println("\n=== SOL 交易详情 ===")
	fmt.Printf("交易ID: %d\n", unsigned.ID)
	fmt.Printf("Fee Payer: %s\n", unsigned.FeePayer)
	fmt.Printf("Recent Blockhash: %s\n", unsigned.RecentBlockhash)
	fmt.Printf("指令数: %d\n", len(unsigned.Instructions))
	for i, ins := range unsigned.Instructions {
		fmt.Printf("  [%d] type=%v program=%v params=%v\n", i+1, ins["type"], ins["program_id"], ins["params"])
	}
}
