package crypto

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// KeyInfo 私钥信息（支持多地址）
type KeyInfo struct {
	KeyID        string   `json:"key_id"`        // 唯一键ID（encrypted_key的SHA256）
	Addresses    []string `json:"addresses"`     // 关联的地址列表
	EncryptedKey string   `json:"encrypted_key"` // 加密后的私钥
	ChainType    string   `json:"chain_type"`    // 链类型 (eth, btc)
	CreatedAt    string   `json:"created_at"`    // 创建时间
	Description  string   `json:"description"`   // 描述
}

// KeyManager 私钥管理器
type KeyManager struct {
	cryptoManager *CryptoManager
	keysFile      string
	keys          map[string]*KeyInfo // keyID -> KeyInfo
}

// NewKeyManager 创建私钥管理器
func NewKeyManager(cryptoManager *CryptoManager) *KeyManager {
	homeDir, _ := os.UserHomeDir()
	keysFile := filepath.Join(homeDir, ".blockchain-signer", "keys.json")

	return &KeyManager{
		cryptoManager: cryptoManager,
		keysFile:      keysFile,
		keys:          make(map[string]*KeyInfo),
	}
}

// LoadKeys 加载私钥（兼容旧格式）
func (km *KeyManager) LoadKeys() error {
	// 确保目录存在
	dir := filepath.Dir(km.keysFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(km.keysFile); os.IsNotExist(err) {
		// 文件不存在，创建空文件
		return km.SaveKeys()
	}

	// 读取文件
	data, err := os.ReadFile(km.keysFile)
	if err != nil {
		return fmt.Errorf("读取私钥文件失败: %w", err)
	}

	// 优先尝试新格式：map[keyID]KeyInfo
	var newFormat map[string]*KeyInfo
	if err := json.Unmarshal(data, &newFormat); err == nil && newFormat != nil {
		km.keys = newFormat
		return nil
	}

	// 兼容旧格式：map[address]OldKeyInfo
	type OldKeyInfo struct {
		Address      string `json:"address"`
		EncryptedKey string `json:"encrypted_key"`
		ChainType    string `json:"chain_type"`
		CreatedAt    string `json:"created_at"`
		Description  string `json:"description"`
	}
	var oldFormat map[string]*OldKeyInfo
	if err := json.Unmarshal(data, &oldFormat); err != nil {
		return fmt.Errorf("解析私钥文件失败: %w", err)
	}

	grouped := make(map[string]*KeyInfo) // keyID -> KeyInfo
	for _, v := range oldFormat {
		keyID := sha256Hex(v.EncryptedKey)
		if existing, ok := grouped[keyID]; ok {
			if !contains(existing.Addresses, v.Address) {
				existing.Addresses = append(existing.Addresses, v.Address)
			}
		} else {
			grouped[keyID] = &KeyInfo{
				KeyID:        keyID,
				Addresses:    []string{v.Address},
				EncryptedKey: v.EncryptedKey,
				ChainType:    v.ChainType,
				CreatedAt:    v.CreatedAt,
				Description:  v.Description,
			}
		}
	}
	km.keys = grouped
	// 保存为新格式
	return km.SaveKeys()
}

// SaveKeys 保存私钥
func (km *KeyManager) SaveKeys() error {
	// 确保目录存在
	dir := filepath.Dir(km.keysFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 序列化为JSON（新格式）
	data, err := json.MarshalIndent(km.keys, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化私钥失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(km.keysFile, data, 0600); err != nil {
		return fmt.Errorf("写入私钥文件失败: %w", err)
	}

	return nil
}

// AddKey 添加私钥（单地址）
func (km *KeyManager) AddKey(address, privateKey, chainType, description, password string) error {
	return km.AddKeyWithAddresses([]string{address}, privateKey, chainType, description, password)
}

// AddKeyWithAddresses 添加私钥（多地址）
func (km *KeyManager) AddKeyWithAddresses(addresses []string, privateKey, chainType, description, password string) error {
	// 加密私钥
	encryptedKey, err := km.cryptoManager.EncryptPrivateKey(privateKey, password)
	if err != nil {
		return fmt.Errorf("加密私钥失败: %w", err)
	}

	// 生成KeyID
	keyID := sha256Hex(encryptedKey)

	if existing, ok := km.keys[keyID]; ok {
		existing.Description = description
		for _, addr := range addresses {
			if !contains(existing.Addresses, addr) {
				existing.Addresses = append(existing.Addresses, addr)
			}
		}
	} else {
		keyInfo := &KeyInfo{
			KeyID:        keyID,
			Addresses:    unique(addresses),
			EncryptedKey: encryptedKey,
			ChainType:    chainType,
			CreatedAt:    fmt.Sprintf("%d", os.Getpid()), // 简化时间戳
			Description:  description,
		}
		km.keys[keyID] = keyInfo
	}

	// 保存到文件
	if err := km.SaveKeys(); err != nil {
		return fmt.Errorf("保存私钥失败: %w", err)
	}

	return nil
}

// GetKey 获取私钥
func (km *KeyManager) GetKey(address, password string) (string, error) {
	info := km.findByAddress(address)
	if info == nil {
		return "", fmt.Errorf("未找到地址 %s 对应的私钥", address)
	}

	// 解密私钥
	privateKey, err := km.cryptoManager.DecryptPrivateKey(info.EncryptedKey, password)
	if err != nil {
		return "", fmt.Errorf("解密私钥失败: %w", err)
	}

	return privateKey, nil
}

// ListKeys 列出所有私钥
func (km *KeyManager) ListKeys() []*KeyInfo {
	var keys []*KeyInfo
	for _, key := range km.keys {
		keys = append(keys, key)
	}
	return keys
}

// RemoveKey 删除私钥（通过任意地址定位记录）
func (km *KeyManager) RemoveKey(address string) error {
	keyID := ""
	for id, info := range km.keys {
		if contains(info.Addresses, address) {
			keyID = id
			break
		}
	}
	if keyID == "" {
		return fmt.Errorf("未找到地址 %s 对应的私钥", address)
	}

	delete(km.keys, keyID)

	// 保存到文件
	if err := km.SaveKeys(); err != nil {
		return fmt.Errorf("保存私钥失败: %w", err)
	}

	return nil
}

// HasKey 检查是否有指定地址的私钥
func (km *KeyManager) HasKey(address string) bool {
	return km.findByAddress(address) != nil
}

// GetKeysByChain 根据链类型获取私钥
func (km *KeyManager) GetKeysByChain(chainType string) []*KeyInfo {
	var keys []*KeyInfo
	for _, key := range km.keys {
		if key.ChainType == chainType {
			keys = append(keys, key)
		}
	}
	return keys
}

// AddAlias 为现有密钥添加地址别名（共享同一加密私钥）
func (km *KeyManager) AddAlias(newAddress, existingAddress string) error {
	info := km.findByAddress(existingAddress)
	if info == nil {
		return fmt.Errorf("未找到原地址 %s 对应的私钥", existingAddress)
	}
	if !contains(info.Addresses, newAddress) {
		info.Addresses = append(info.Addresses, newAddress)
	}
	if err := km.SaveKeys(); err != nil {
		return fmt.Errorf("保存私钥失败: %w", err)
	}
	return nil
}

// 辅助函数
func (km *KeyManager) findByAddress(address string) *KeyInfo {
	for _, info := range km.keys {
		if contains(info.Addresses, address) {
			return info
		}
	}
	return nil
}

func contains(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func unique(list []string) []string {
	m := make(map[string]struct{})
	var out []string
	for _, v := range list {
		if _, ok := m[v]; ok {
			continue
		}
		m[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", sum[:])
}
