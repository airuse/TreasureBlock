package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// CryptoManager 加密管理器
type CryptoManager struct {
	// 加盐密钥 - 用于加密存储的私钥
	saltKey []byte
}

// NewCryptoManager 创建加密管理器
func NewCryptoManager() *CryptoManager {
	// 使用固定的盐值（实际应用中应该从配置文件或环境变量读取）
	salt := sha256.Sum256([]byte("blockchain-signer-salt-2024"))
	return &CryptoManager{
		saltKey: salt[:],
	}
}

// EncryptPrivateKey 加密私钥
func (cm *CryptoManager) EncryptPrivateKey(privateKey, password string) (string, error) {
	// 使用密码和盐值生成加密密钥
	key := cm.deriveKey(password)

	// 创建AES加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建AES加密器失败: %w", err)
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM模式失败: %w", err)
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成nonce失败: %w", err)
	}

	// 加密私钥
	ciphertext := gcm.Seal(nonce, nonce, []byte(privateKey), nil)

	// 返回十六进制编码的加密数据
	return hex.EncodeToString(ciphertext), nil
}

// DecryptPrivateKey 解密私钥
func (cm *CryptoManager) DecryptPrivateKey(encryptedKey, password string) (string, error) {
	// 解码十六进制数据
	ciphertext, err := hex.DecodeString(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("解码加密数据失败: %w", err)
	}

	// 使用密码和盐值生成解密密钥
	key := cm.deriveKey(password)

	// 创建AES解密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建AES解密器失败: %w", err)
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM模式失败: %w", err)
	}

	// 检查数据长度
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("加密数据长度不足")
	}

	// 提取nonce和密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密私钥
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}

	return string(plaintext), nil
}

// deriveKey 从密码派生加密密钥
func (cm *CryptoManager) deriveKey(password string) []byte {
	// 使用PBKDF2或简单的SHA256派生密钥
	// 这里使用简单的SHA256，实际应用中应该使用PBKDF2
	hash := sha256.New()
	hash.Write([]byte(password))
	hash.Write(cm.saltKey)
	return hash.Sum(nil)
}

// HashPassword 哈希密码
func (cm *CryptoManager) HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// VerifyPassword 验证密码
func (cm *CryptoManager) VerifyPassword(password, hash string) bool {
	return cm.HashPassword(password) == hash
}
