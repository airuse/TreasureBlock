package main

import (
	"blockChainBrowser/client/signer/internal/btc"
	"blockChainBrowser/client/signer/internal/crypto"
	"blockChainBrowser/client/signer/internal/eth"
	"blockChainBrowser/client/signer/internal/utils"
	"blockChainBrowser/client/signer/pkg"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/skip2/go-qrcode"
)

// 内嵌密码哈希 - 用于验证系统访问权限
const EMBEDDED_PASSWORD_HASH = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" // "hello"的SHA256哈希

// 辅助：格式化地址列表
func formatAddresses(addresses []string) string {
	if len(addresses) == 0 {
		return "(无地址)"
	}
	return strings.Join(addresses, ", ")
}

// 辅助：安全获取第一个地址
func firstAddress(addresses []string) (string, bool) {
	if len(addresses) == 0 {
		return "", false
	}
	return addresses[0], true
}

func main() {
	fmt.Println("=== 区块链交易签名程序 ===")
	fmt.Println("版本: 1.0.0")
	fmt.Println("支持: ETH, BTC")
	fmt.Println("=========================")

	// 验证系统密码
	if !verifySystemPassword() {
		fmt.Println("❌ 系统密码验证失败，程序退出")
		os.Exit(1)
	}

	fmt.Println("✅ 系统密码验证成功")

	// 初始化加密模块
	cryptoManager := crypto.NewCryptoManager()

	// 初始化ETH签名器
	ethSigner := eth.NewETHSigner(cryptoManager)

	// 初始化BTC签名器
	btcSigner := btc.NewBTCSigner(cryptoManager)

	// 主菜单循环
	for {
		showMainMenu()

		var choice string
		fmt.Print("请选择操作: ")
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			handleQRCodeImport(ethSigner, btcSigner)
		case "2":
			handlePrivateKeyImport(cryptoManager)
		case "3":
			handleKeyManagement(cryptoManager)
		case "4":
			handleSystemSettings()
		case "5":
			fmt.Println("👋 感谢使用，程序退出")
			os.Exit(0)
		default:
			fmt.Println("❌ 无效选择，请重新输入")
		}

		fmt.Println("\n按回车键继续...")
		fmt.Scanln()
	}
}

// 验证系统密码
func verifySystemPassword() bool {
	password, err := utils.ReadPassword("请输入系统密码: ")
	if err != nil {
		fmt.Println("读取密码失败:", err)
		return false
	}

	// 计算输入密码的SHA256哈希
	hash := utils.SHA256Hash(password)
	return hash == EMBEDDED_PASSWORD_HASH
}

// 显示主菜单
func showMainMenu() {
	fmt.Println("\n=== 主菜单 ===")
	fmt.Println("1. 导入QR码并签名")
	fmt.Println("2. 导入私钥")
	fmt.Println("3. 密钥管理")
	fmt.Println("4. 系统设置")
	fmt.Println("5. 退出程序")
	fmt.Println("===============")
}

// 选择链类型
func selectChainType() string {
	fmt.Println("\n请选择链类型:")
	fmt.Println("1. ETH (Ethereum)")
	fmt.Println("2. BTC (Bitcoin)")
	fmt.Print("请选择 (1-2): ")

	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		return "eth"
	case "2":
		return "btc"
	default:
		fmt.Println("❌ 无效选择")
		return ""
	}
}

// 处理QR码导入和签名
func handleQRCodeImport(ethSigner *eth.ETHSigner, btcSigner *btc.BTCSigner) {
	fmt.Println("\n=== QR码导入和签名 ===")

	// 获取QR码文件路径
	fmt.Print("请输入QR码图片文件路径: ")
	var filePath string
	fmt.Scanln(&filePath)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("❌ 文件不存在: %s\n", filePath)
		return
	}

	// 检查文件格式
	if !utils.IsQRCodeFile(filePath) {
		fmt.Printf("❌ 不支持的图片格式，支持的格式: %v\n", utils.GetSupportedFormats())
		return
	}

	// 扫描QR码
	scanner := utils.NewQRScanner()
	qrData, err := scanner.ScanQRCodeFromFile(filePath)
	if err != nil {
		fmt.Printf("❌ QR码扫描失败: %v\n", err)
		return
	}

	fmt.Printf("✅ QR码扫描成功\n")
	fmt.Printf("扫描到的数据: %s\n", qrData)

	// 解析QR码数据
	transaction, err := pkg.ParseQRCodeData(qrData)
	if err != nil {
		fmt.Printf("❌ QR码数据解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ QR码数据解析成功\n")
	fmt.Printf("交易ID: %d\n", transaction.ID)
	fmt.Printf("链类型: %s\n", transaction.Type)
	fmt.Printf("发送地址: %s\n", transaction.From)

	// 根据QR码中的类型字段自动选择签名器
	if transaction.IsETH() {
		fmt.Println("🔷 自动识别为ETH交易，使用ETH签名器")
		signETHTransaction(ethSigner, transaction)
	} else if transaction.IsBTC() {
		fmt.Println("🟠 自动识别为BTC交易，使用BTC签名器")
		signBTCTransaction(btcSigner, transaction)
	} else {
		fmt.Printf("❌ 不支持的链类型: %s\n", transaction.Type)
		return
	}
}

// 签名ETH交易
func signETHTransaction(ethSigner *eth.ETHSigner, transaction *pkg.TransactionData) {
	fmt.Println("\n=== ETH交易签名 ===")

	// 显示交易详情
	ethSigner.DisplayTransaction(transaction)

	// 执行签名（确认步骤已合并到密码输入中）
	signedTx, err := ethSigner.SignTransaction(transaction)
	if err != nil {
		fmt.Printf("❌ ETH交易签名失败: %v\n", err)
		return
	}

	fmt.Println("✅ ETH交易签名成功")
	fmt.Printf("签名结果: %s\n", signedTx)

	// 提供导出选项
	handleSignatureExport(signedTx, transaction)
}

// 签名BTC交易
func signBTCTransaction(btcSigner *btc.BTCSigner, transaction *pkg.TransactionData) {
	fmt.Println("\n=== BTC交易签名 ===")

	// 显示交易详情
	btcSigner.DisplayTransaction(transaction)

	// 确认签名
	fmt.Print("\n是否确认签名此交易? (y/N): ")
	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "y" && confirm != "Y" {
		fmt.Println("❌ 用户取消签名")
		return
	}

	// 执行签名
	signedTx, err := btcSigner.SignTransaction(transaction)
	if err != nil {
		fmt.Printf("❌ BTC交易签名失败: %v\n", err)
		return
	}

	fmt.Println("✅ BTC交易签名成功")
	fmt.Printf("签名结果: %s\n", signedTx)

	// 提供导出选项
	handleSignatureExport(signedTx, transaction)
}

// 处理签名导出
func handleSignatureExport(signedTx string, transaction *pkg.TransactionData) {
	fmt.Println("\n=== 签名导出 ===")
	fmt.Println("1. 复制到剪贴板 (JSON: {id, signer})")
	fmt.Println("2. 保存到文件 (JSON)")
	fmt.Println("3. 生成并展示QR码")
	fmt.Println("4. 返回主菜单")

	fmt.Print("请选择导出方式: ")
	var choice string
	fmt.Scanln(&choice)

	// 构造导出JSON
	exportObj := map[string]interface{}{
		"id":     transaction.ID,
		"signer": signedTx,
	}
	exportJSON, _ := json.Marshal(exportObj)

	switch choice {
	case "1":
		if err := utils.CopyToClipboard(string(exportJSON)); err != nil {
			fmt.Printf("❌ 复制到剪贴板失败: %v\n", err)
		} else {
			fmt.Println("✅ 已复制JSON到剪贴板")
		}
	case "2":
		filename := fmt.Sprintf("signed_tx_%d.json", transaction.ID)
		if err := utils.SaveToFile(filename, string(exportJSON)); err != nil {
			fmt.Printf("❌ 保存到文件失败: %v\n", err)
		} else {
			fmt.Printf("✅ 已保存JSON到文件: %s\n", filename)
		}
	case "3":
		pngName := fmt.Sprintf("signed_tx_%d.png", transaction.ID)
		if err := qrcode.WriteFile(string(exportJSON), qrcode.Medium, 320, pngName); err != nil {
			fmt.Printf("❌ 生成QR码失败: %v\n", err)
		} else {
			fmt.Printf("✅ 已生成签名QR码: %s\n", pngName)
			// 尝试用系统默认查看器打开
			openCmd := ""
			args := []string{}
			if utils.IsMacOS() {
				openCmd = "open"
				args = []string{pngName}
			} else if utils.IsWindows() {
				openCmd = "rundll32"
				args = []string{"url.dll,FileProtocolHandler", pngName}
			} else if utils.IsLinux() {
				openCmd = "xdg-open"
				args = []string{pngName}
			}
			if openCmd != "" {
				if err := exec.Command(openCmd, args...).Start(); err != nil {
					fmt.Printf("⚠️  无法自动打开图片，请手动查看: %s\n", pngName)
				}
			}
		}
	case "4":
		return
	default:
		fmt.Println("❌ 无效选择")
	}
}

// 处理私钥导入
func handlePrivateKeyImport(cryptoManager *crypto.CryptoManager) {
	fmt.Println("\n=== 私钥导入 ===")

	// 选择链类型
	chainType := selectChainType()
	if chainType == "" {
		return
	}

	// 创建私钥管理器
	keyManager := crypto.NewKeyManager(cryptoManager)
	if err := keyManager.LoadKeys(); err != nil {
		fmt.Printf("❌ 加载私钥失败: %v\n", err)
		return
	}

	// 获取私钥
	fmt.Print("请输入私钥 (十六进制格式，不带0x前缀): ")
	var privateKey string
	fmt.Scanln(&privateKey)

	// 验证私钥格式
	if len(privateKey) != 64 {
		fmt.Println("❌ 私钥格式错误，应该是64位十六进制字符")
		return
	}

	// 获取描述
	fmt.Print("请输入描述 (可选): ")
	var description string
	fmt.Scanln(&description)

	// 获取加密密码（隐藏输入）
	password, err := utils.ReadPassword("请输入加密密码: ")
	if err != nil {
		fmt.Println("读取密码失败:", err)
		return
	}

	// 根据链类型自动派生地址并保存
	if chainType == "eth" {
		lower, checksum, err := crypto.DeriveETHAddresses(privateKey)
		if err != nil {
			fmt.Printf("❌ 生成ETH地址失败: %v\n", err)
			return
		}
		if err := keyManager.AddKey(lower, privateKey, chainType, description, password); err != nil {
			fmt.Printf("❌ 添加私钥失败: %v\n", err)
			return
		}
		if err := keyManager.AddAlias(checksum, lower); err != nil {
			fmt.Printf("⚠️  添加校验地址别名失败: %v\n", err)
		}
		fmt.Println("✅ 私钥导入成功 (ETH)")
		fmt.Printf("地址: %s\n", lower)
		fmt.Printf("校验地址: %s\n", checksum)
	} else if chainType == "btc" {
		wpkh, wsh, p2pkh, p2sh, err := crypto.DeriveBTCAddresses(privateKey)
		if err != nil {
			fmt.Printf("❌ 生成BTC地址失败: %v\n", err)
			return
		}
		if err := keyManager.AddKey(wpkh, privateKey, chainType, description, password); err != nil {
			fmt.Printf("❌ 添加私钥失败: %v\n", err)
			return
		}
		for _, alias := range []string{wsh, p2pkh, p2sh} {
			if err := keyManager.AddAlias(alias, wpkh); err != nil {
				fmt.Printf("⚠️  添加地址别名失败: %v\n", err)
			}
		}
		fmt.Println("✅ 私钥导入成功 (BTC)")
		fmt.Printf("P2WPKH: %s\n", wpkh)
		fmt.Printf("P2WSH: %s\n", wsh)
		fmt.Printf("P2PKH: %s\n", p2pkh)
		fmt.Printf("P2SH-P2WPKH: %s\n", p2sh)
	}

	fmt.Printf("链类型: %s\n", chainType)
	fmt.Printf("描述: %s\n", description)
}

// 处理密钥管理
func handleKeyManagement(cryptoManager *crypto.CryptoManager) {
	fmt.Println("\n=== 密钥管理 ===")

	// 创建私钥管理器
	keyManager := crypto.NewKeyManager(cryptoManager)
	if err := keyManager.LoadKeys(); err != nil {
		fmt.Printf("❌ 加载私钥失败: %v\n", err)
		return
	}

	for {
		showKeyManagementMenu()

		var choice string
		fmt.Print("请选择操作: ")
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			listKeys(keyManager)
		case "2":
			listKeysByChain(keyManager, "eth")
		case "3":
			listKeysByChain(keyManager, "btc")
		case "4":
			removeKey(keyManager)
		case "5":
			exportKey(keyManager, cryptoManager)
		case "6":
			addAddressAlias(keyManager)
		case "7":
			return
		default:
			fmt.Println("❌ 无效选择，请重新输入")
		}

		fmt.Println("\n按回车键继续...")
		fmt.Scanln()
	}
}

// 显示密钥管理菜单
func showKeyManagementMenu() {
	fmt.Println("\n=== 密钥管理菜单 ===")
	fmt.Println("1. 查看所有私钥")
	fmt.Println("2. 查看ETH私钥")
	fmt.Println("3. 查看BTC私钥")
	fmt.Println("4. 删除私钥")
	fmt.Println("5. 导出私钥")
	fmt.Println("6. 添加地址别名")
	fmt.Println("7. 返回主菜单")
	fmt.Println("==================")
}

// 列出所有私钥
func listKeys(keyManager *crypto.KeyManager) {
	fmt.Println("\n=== 私钥列表 ===")

	keys := keyManager.ListKeys()
	if len(keys) == 0 {
		fmt.Println("暂无私钥")
		return
	}

	for i, key := range keys {
		fmt.Printf("%d. 地址: %s\n", i+1, formatAddresses(key.Addresses))
		fmt.Printf("   链类型: %s\n", key.ChainType)
		fmt.Printf("   描述: %s\n", key.Description)
		fmt.Printf("   创建时间: %s\n", key.CreatedAt)
		fmt.Println("   ---")
	}
}

// 按链类型列出私钥
func listKeysByChain(keyManager *crypto.KeyManager, chainType string) {
	chainName := "ETH"
	if chainType == "btc" {
		chainName = "BTC"
	}

	fmt.Printf("\n=== %s私钥列表 ===\n", chainName)

	keys := keyManager.GetKeysByChain(chainType)
	if len(keys) == 0 {
		fmt.Printf("暂无%s私钥\n", chainName)
		return
	}

	for i, key := range keys {
		fmt.Printf("%d. 地址: %s\n", i+1, formatAddresses(key.Addresses))
		fmt.Printf("   链类型: %s\n", key.ChainType)
		fmt.Printf("   描述: %s\n", key.Description)
		fmt.Printf("   创建时间: %s\n", key.CreatedAt)
		fmt.Println("   ---")
	}
}

// 删除私钥
func removeKey(keyManager *crypto.KeyManager) {
	fmt.Println("\n=== 删除私钥 ===")

	keys := keyManager.ListKeys()
	if len(keys) == 0 {
		fmt.Println("暂无私钥可删除")
		return
	}

	// 显示私钥列表
	for i, key := range keys {
		fmt.Printf("%d. [%s] (%s)\n", i+1, formatAddresses(key.Addresses), key.ChainType)
	}

	fmt.Print("请选择要删除的私钥编号: ")
	var choice string
	fmt.Scanln(&choice)

	// 验证选择
	var index int
	if _, err := fmt.Sscanf(choice, "%d", &index); err != nil || index < 1 || index > len(keys) {
		fmt.Println("❌ 无效选择")
		return
	}

	selectedKey := keys[index-1]

	// 地址校验
	addr, ok := firstAddress(selectedKey.Addresses)
	if !ok {
		fmt.Println("❌ 该条目没有任何地址，无法删除")
		return
	}

	// 确认删除
	fmt.Printf("确认删除这些地址的私钥? [%s] (y/N): ", formatAddresses(selectedKey.Addresses))
	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "y" && confirm != "Y" {
		fmt.Println("❌ 取消删除")
		return
	}

	// 删除私钥（使用第一个地址定位）
	if err := keyManager.RemoveKey(addr); err != nil {
		fmt.Printf("❌ 删除私钥失败: %v\n", err)
		return
	}

	fmt.Println("✅ 私钥删除成功")
}

// 导出私钥
func exportKey(keyManager *crypto.KeyManager, cryptoManager *crypto.CryptoManager) {
	fmt.Println("\n=== 导出私钥 ===")

	keys := keyManager.ListKeys()
	if len(keys) == 0 {
		fmt.Println("暂无私钥可导出")
		return
	}

	// 显示私钥列表
	for i, key := range keys {
		fmt.Printf("%d. [%s] (%s)\n", i+1, formatAddresses(key.Addresses), key.ChainType)
	}

	fmt.Print("请选择要导出的私钥编号: ")
	var choice string
	fmt.Scanln(&choice)

	// 验证选择
	var index int
	if _, err := fmt.Sscanf(choice, "%d", &index); err != nil || index < 1 || index > len(keys) {
		fmt.Println("❌ 无效选择")
		return
	}

	selectedKey := keys[index-1]
	addr, ok := firstAddress(selectedKey.Addresses)
	if !ok {
		fmt.Println("❌ 该条目没有任何地址，无法导出")
		return
	}

	// 获取解密密码（隐藏输入）
	password, err := utils.ReadPassword("请输入解密密码: ")
	if err != nil {
		fmt.Println("读取密码失败:", err)
		return
	}

	// 解密私钥
	privateKey, err := keyManager.GetKey(addr, password)
	if err != nil {
		fmt.Printf("❌ 解密私钥失败: %v\n", err)
		return
	}

	// 显示私钥
	fmt.Println("\n=== 私钥信息 ===")
	fmt.Printf("地址: [%s]\n", formatAddresses(selectedKey.Addresses))
	fmt.Printf("链类型: %s\n", selectedKey.ChainType)
	fmt.Printf("私钥: %s\n", privateKey)
	fmt.Println("================")

	// 提供导出选项
	fmt.Println("\n导出选项:")
	fmt.Println("1. 复制到剪贴板")
	fmt.Println("2. 保存到文件")
	fmt.Println("3. 返回")

	fmt.Print("请选择: ")
	var exportChoice string
	fmt.Scanln(&exportChoice)

	switch exportChoice {
	case "1":
		if err := utils.CopyToClipboard(privateKey); err != nil {
			fmt.Printf("❌ 复制到剪贴板失败: %v\n", err)
		} else {
			fmt.Println("✅ 私钥已复制到剪贴板")
		}
	case "2":
		prefix := addr
		if len(prefix) > 8 {
			prefix = addr[:8]
		}
		filename := fmt.Sprintf("private_key_%s.txt", prefix)
		if err := utils.SaveToFile(filename, privateKey); err != nil {
			fmt.Printf("❌ 保存到文件失败: %v\n", err)
		} else {
			fmt.Printf("✅ 私钥已保存到文件: %s\n", filename)
		}
	case "3":
		return
	default:
		fmt.Println("❌ 无效选择")
	}
}

// 添加地址别名
func addAddressAlias(keyManager *crypto.KeyManager) {
	fmt.Println("\n=== 添加地址别名 ===")
	keys := keyManager.ListKeys()
	if len(keys) == 0 {
		fmt.Println("暂无私钥")
		return
	}
	// 选择已有地址
	for i, key := range keys {
		fmt.Printf("%d. [%s] (%s)\n", i+1, formatAddresses(key.Addresses), key.ChainType)
	}
	fmt.Print("请选择已有地址编号: ")
	var choice string
	fmt.Scanln(&choice)
	var index int
	if _, err := fmt.Sscanf(choice, "%d", &index); err != nil || index < 1 || index > len(keys) {
		fmt.Println("❌ 无效选择")
		return
	}
	from := keys[index-1]
	fromAddr, ok := firstAddress(from.Addresses)
	if !ok {
		fmt.Println("❌ 该条目没有任何地址，无法添加别名")
		return
	}
	// 输入新地址
	fmt.Print("请输入要添加的地址别名: ")
	var alias string
	fmt.Scanln(&alias)
	if alias == "" {
		fmt.Println("❌ 地址不能为空")
		return
	}
	if err := keyManager.AddAlias(alias, fromAddr); err != nil {
		fmt.Printf("❌ 添加地址别名失败: %v\n", err)
		return
	}
	fmt.Println("✅ 地址别名添加成功")
}

// 处理系统设置
func handleSystemSettings() {
	fmt.Println("\n=== 系统设置 ===")

	for {
		showSystemSettingsMenu()

		var choice string
		fmt.Print("请选择操作: ")
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			showSystemInfo()
		case "2":
			changeSystemPassword()
		case "3":
			showSecurityInfo()
		case "4":
			return
		default:
			fmt.Println("❌ 无效选择，请重新输入")
		}

		fmt.Println("\n按回车键继续...")
		fmt.Scanln()
	}
}

// 显示系统设置菜单
func showSystemSettingsMenu() {
	fmt.Println("\n=== 系统设置菜单 ===")
	fmt.Println("1. 系统信息")
	fmt.Println("2. 修改系统密码")
	fmt.Println("3. 安全信息")
	fmt.Println("4. 返回主菜单")
	fmt.Println("==================")
}

// 显示系统信息
func showSystemInfo() {
	fmt.Println("\n=== 系统信息 ===")
	fmt.Printf("程序版本: 1.0.0\n")
	fmt.Printf("支持链: ETH, BTC\n")
	fmt.Printf("操作系统: %s\n", utils.GetOS())
	fmt.Printf("Go版本: %s\n", runtime.Version())
	fmt.Println("================")
}

// 修改系统密码
func changeSystemPassword() {
	fmt.Println("\n=== 修改系统密码 ===")
	fmt.Println("⚠️  系统密码修改功能开发中...")
	fmt.Println("当前密码: hello")
	fmt.Println("密码哈希: 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824")
	fmt.Println("==================")
}

// 显示安全信息
func showSecurityInfo() {
	fmt.Println("\n=== 安全信息 ===")
	fmt.Println("✅ 内嵌密码验证")
	fmt.Println("✅ 私钥加密存储")
	fmt.Println("✅ AES-GCM加密算法")
	fmt.Println("✅ 加盐处理")
	fmt.Println("✅ 完全离线运行")
	fmt.Println("✅ 私钥不存储明文")
	fmt.Println("================")
}
