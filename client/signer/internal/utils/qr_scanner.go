package utils

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	_ "golang.org/x/image/bmp"
)

// QRScanner QR码扫描器
type QRScanner struct {
	reader gozxing.Reader
}

// NewQRScanner 创建QR码扫描器
func NewQRScanner() *QRScanner {
	return &QRScanner{
		reader: qrcode.NewQRCodeReader(),
	}
}

// ScanQRCodeFromFile 从文件扫描QR码
func (qs *QRScanner) ScanQRCodeFromFile(filename string) (string, error) {
	// 打开图片文件
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 解码图片
	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("解码图片失败: %w", err)
	}

	// 创建位图
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", fmt.Errorf("创建位图失败: %w", err)
	}

	// 扫描QR码
	result, err := qs.reader.Decode(bmp, nil)
	if err != nil {
		return "", fmt.Errorf("扫描QR码失败: %w", err)
	}

	return result.GetText(), nil
}

// ScanQRCodeFromImage 从图片对象扫描QR码
func (qs *QRScanner) ScanQRCodeFromImage(img image.Image) (string, error) {
	// 创建位图
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", fmt.Errorf("创建位图失败: %w", err)
	}

	// 扫描QR码
	result, err := qs.reader.Decode(bmp, nil)
	if err != nil {
		return "", fmt.Errorf("扫描QR码失败: %w", err)
	}

	return result.GetText(), nil
}

// IsQRCodeFile 检查文件是否为支持的图片格式
func IsQRCodeFile(filename string) bool {
	// 支持的图片格式
	supportedFormats := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp"}

	for _, format := range supportedFormats {
		if len(filename) >= len(format) && filename[len(filename)-len(format):] == format {
			return true
		}
	}

	return false
}

// GetSupportedFormats 获取支持的图片格式
func GetSupportedFormats() []string {
	return []string{".png", ".jpg", ".jpeg", ".gif", ".bmp"}
}
