package utils

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"unicode"

	"golang.org/x/term"
)

// SHA256Hash 计算SHA256哈希
func SHA256Hash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}

// CopyToClipboard 复制文本到剪贴板
func CopyToClipboard(text string) error {
	var cmd string
	var args []string

	switch GetOS() {
	case "darwin": // macOS
		cmd = "pbcopy"
		args = []string{}
	case "linux":
		// 尝试使用xclip，如果失败则使用xsel
		cmd = "xclip"
		args = []string{"-selection", "clipboard"}
	case "windows":
		cmd = "clip"
		args = []string{}
	default:
		return fmt.Errorf("不支持的操作系统: %s", GetOS())
	}

	// 创建命令
	execCmd := exec.Command(cmd, args...)

	// 设置输入
	stdin, err := execCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("创建管道失败: %w", err)
	}

	// 启动命令
	if err := execCmd.Start(); err != nil {
		// 如果是Linux且xclip失败，尝试xsel
		if GetOS() == "linux" && cmd == "xclip" {
			cmd = "xsel"
			args = []string{"--clipboard", "--input"}
			execCmd = exec.Command(cmd, args...)
			stdin, err = execCmd.StdinPipe()
			if err != nil {
				return fmt.Errorf("创建管道失败: %w", err)
			}
			if err := execCmd.Start(); err != nil {
				return fmt.Errorf("启动命令失败: %w", err)
			}
		} else {
			return fmt.Errorf("启动命令失败: %w", err)
		}
	}

	// 写入数据
	go func() {
		defer stdin.Close()
		stdin.Write([]byte(text))
	}()

	// 等待命令完成
	if err := execCmd.Wait(); err != nil {
		return fmt.Errorf("命令执行失败: %w", err)
	}

	return nil
}

// SaveToFile 保存文本到文件
func SaveToFile(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// GetOS 获取操作系统类型
func GetOS() string {
	return runtime.GOOS
}

// IsWindows 判断是否为Windows系统
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsMacOS 判断是否为macOS系统
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux 判断是否为Linux系统
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// ReadPassword 读取隐藏输入的密码（不回显）
func ReadPassword(prompt string) (string, error) {
	if strings.TrimSpace(prompt) != "" {
		fmt.Print(prompt)
	}
	pwd, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("读取密码失败: %w", err)
	}
	return string(pwd), nil
}

// GenerateID 生成唯一ID
func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// ReplaceAll 替换字符串中的所有匹配项
func ReplaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// ReadLine 读取一整行（包含空格）
func ReadLine(prompt string) (string, error) {
	if strings.TrimSpace(prompt) != "" {
		fmt.Print(prompt)
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// IsHexString 判断是否为十六进制字符串（可带0x前缀）
func IsHexString(s string) bool {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}
	if len(s) == 0 || len(s)%2 != 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}
