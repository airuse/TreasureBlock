package sol

import (
	"blockChainBrowser/client/signer/internal/crypto"
	"blockChainBrowser/client/signer/internal/utils"
	"blockChainBrowser/client/signer/pkg"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	solcommon "github.com/blocto/solana-go-sdk/common"
	solata "github.com/blocto/solana-go-sdk/program/associated_token_account"
	solsys "github.com/blocto/solana-go-sdk/program/system"
	soltoken "github.com/blocto/solana-go-sdk/program/token"
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
		case "spl_approve":
			// 授权操作：授权 delegate 可以花费 authority 的代币
			params, _ := plan["params"].(map[string]interface{})
			authority := solcommon.PublicKeyFromString(fmt.Sprintf("%v", params["authority"]))
			delegate := solcommon.PublicKeyFromString(fmt.Sprintf("%v", params["delegate"]))
			mint := solcommon.PublicKeyFromString(fmt.Sprintf("%v", params["mint"]))
			amountStr := fmt.Sprintf("%v", params["amount"])
			amount, err := strconv.ParseUint(amountStr, 10, 64)
			if err != nil {
				return "", fmt.Errorf("解析授权数量失败: %w", err)
			}

			// 计算授权者的 ATA 地址
			authorityATA, _, err := solcommon.FindAssociatedTokenAddress(authority, mint)
			if err != nil {
				return "", fmt.Errorf("计算授权者ATA地址失败: %w", err)
			}

			// 构建SPL代币授权指令
			ix := soltoken.Approve(soltoken.ApproveParam{
				From:   authorityATA, // 授权者的代币账户
				To:     delegate,     // 被授权者地址
				Auth:   authority,    // 授权者钱包地址
				Amount: amount,       // 授权数量
			})
			ins = append(ins, ix)

		case "spl_transfer_checked":
			params, _ := plan["params"].(map[string]interface{})
			fromWallet := solcommon.PublicKeyFromString(fmt.Sprintf("%v", params["from_owner"]))
			toWallet := solcommon.PublicKeyFromString(fmt.Sprintf("%v", params["to_owner"]))
			mint := solcommon.PublicKeyFromString(fmt.Sprintf("%v", params["mint"]))
			amountStr := fmt.Sprintf("%v", params["amount"])
			amount, err := strconv.ParseUint(amountStr, 10, 64)
			if err != nil {
				return "", fmt.Errorf("解析代币数量失败: %w", err)
			}

			// 获取代币精度（从交易数据中获取，如果没有则使用默认值）
			decimals := uint8(6) // 默认精度
			if decimalsParam, exists := params["decimals"]; exists {
				if decimalsFloat, ok := decimalsParam.(float64); ok {
					decimals = uint8(decimalsFloat)
				} else if decimalsInt, ok := decimalsParam.(int); ok {
					decimals = uint8(decimalsInt)
				}
			}

			// 计算 Associated Token Account (ATA) 地址
			// ATA = findProgramAddress([owner, token_program_id, mint], associated_token_program_id)
			fromATA, _, err := solcommon.FindAssociatedTokenAddress(fromWallet, mint)
			if err != nil {
				return "", fmt.Errorf("计算发送者ATA地址失败: %w", err)
			}

			toATA, _, err := solcommon.FindAssociatedTokenAddress(toWallet, mint)
			if err != nil {
				return "", fmt.Errorf("计算接收者ATA地址失败: %w", err)
			}

			// 检查是否需要创建接收者的ATA账户
			needCreateToATA := false
			if createParam, exists := params["create"]; exists {
				if createInt, ok := createParam.(int); ok && createInt == 1 {
					needCreateToATA = true
				} else if createFloat, ok := createParam.(float64); ok && createFloat == 1 {
					needCreateToATA = true
				}
			}

			// 发送方 ATA 不创建；仅允许为接收方创建

			// 如果需要创建接收者的ATA账户，添加创建指令
			if needCreateToATA && fromWallet.String() != toWallet.String() {
				createToATAIx := solata.CreateAssociatedTokenAccount(solata.CreateAssociatedTokenAccountParam{
					Funder:                 solcommon.PublicKeyFromString(unsigned.FeePayer),
					Owner:                  toWallet,
					Mint:                   mint,
					AssociatedTokenAccount: toATA,
				})
				ins = append(ins, createToATAIx)
			}

			// 选择授权账户：优先使用 delegate_auth，否则回退 from_owner
			authWallet := fromWallet
			if da, ok := params["delegate_auth"]; ok {
				authWallet = solcommon.PublicKeyFromString(fmt.Sprintf("%v", da))
			}

			// 构建SPL代币转账指令
			// TransferChecked 需要 from, to, mint, auth, amount, decimals
			// Auth 为被授权者(delegate)或from_owner
			ix := soltoken.TransferChecked(soltoken.TransferCheckedParam{
				From:     fromATA,
				To:       toATA,
				Mint:     mint,
				Auth:     authWallet,
				Amount:   amount,
				Decimals: decimals,
			})
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
	s, _ := json.Marshal(ins)
	fmt.Printf("DEBUG: 构建交易指令: %v\n", string(s))

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
