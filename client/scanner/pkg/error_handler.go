package pkg

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// HandleFatalError 处理致命错误，直接退出程序
func HandleFatalError(err error, operation string) {
	logrus.Errorf("🚨 致命错误 - %s: %v", operation, err)

	// 发送钉钉通知（通过延迟导入避免循环依赖）
	sendDingTalkFatalError(operation, err)

	// 根据错误类型提供解决建议
	var suggestion string
	errStr := strings.ToUpper(err.Error())

	switch {
	case strings.Contains(errStr, "BLOCK_VERIFICATION_FAILED"):
		suggestion = "请检查RPA节点配置，确保节点有效且不存在测试节点和正式节点混用情况"
	case strings.Contains(errStr, "TX_UPLOAD_FAILED"):
		suggestion = "请检查网络连接和API配置，确保能够正常上传交易数据"
	case strings.Contains(errStr, "TIMEOUT_EXCEEDED"):
		suggestion = "请增加RPA节点数量或优化网络配置，减少超时情况"
	default:
		suggestion = "请检查配置并重试，如果问题持续存在请联系技术支持"
	}

	// 显示错误信息
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🚨 扫块程序遇到致命错误，程序将退出")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("错误类型: %s\n", operation)
	fmt.Printf("错误详情: %v\n", err)
	fmt.Printf("解决建议: %s\n", suggestion)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("💡 提示：修复配置问题后，重新启动扫块程序")
	fmt.Println(strings.Repeat("=", 80))

	// 等待用户确认
	fmt.Print("\n按回车键退出程序...")
	fmt.Scanln()

	// 退出程序
	os.Exit(1)
}

// sendDingTalkFatalError 发送钉钉致命错误通知
func sendDingTalkFatalError(operation string, err error) {
	SendDingTalkFatalError(operation, err)
}
