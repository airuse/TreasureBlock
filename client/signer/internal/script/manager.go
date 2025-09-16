package script

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"blockChainBrowser/client/signer/internal/utils"
)

// NewScriptManager 创建脚本管理器
func NewScriptManager() *ScriptManager {
	homeDir, _ := os.UserHomeDir()
	scriptDir := filepath.Join(homeDir, ".blockchain-signer")
	filePath := filepath.Join(scriptDir, "scripts.json")

	// 确保目录存在
	if err := os.MkdirAll(scriptDir, 0700); err != nil {
		fmt.Printf("❌ 创建脚本目录失败: %v\n", err)
		return nil
	}

	sm := &ScriptManager{
		scripts:   []Script{},
		templates: DefaultTemplates,
		filePath:  filePath,
	}

	// 加载现有脚本
	if err := sm.LoadScripts(); err != nil {
		fmt.Printf("⚠️  加载脚本失败: %v\n", err)
	}

	return sm
}

// LoadScripts 加载脚本
func (sm *ScriptManager) LoadScripts() error {
	if _, err := os.Stat(sm.filePath); os.IsNotExist(err) {
		// 文件不存在，创建空文件
		return sm.SaveScripts()
	}

	data, err := os.ReadFile(sm.filePath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		sm.scripts = []Script{}
		return nil
	}

	return json.Unmarshal(data, &sm.scripts)
}

// SaveScripts 保存脚本
func (sm *ScriptManager) SaveScripts() error {
	data, err := json.MarshalIndent(sm.scripts, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sm.filePath, data, 0600)
}

// AddScript 添加脚本
func (sm *ScriptManager) AddScript(script Script) error {
	script.ID = utils.GenerateID()
	script.CreatedAt = time.Now()
	script.UpdatedAt = time.Now()

	sm.scripts = append(sm.scripts, script)
	return sm.SaveScripts()
}

// GetScript 获取脚本
func (sm *ScriptManager) GetScript(id string) (*Script, error) {
	for i, script := range sm.scripts {
		if script.ID == id {
			return &sm.scripts[i], nil
		}
	}
	return nil, fmt.Errorf("脚本不存在: %s", id)
}

// ListScripts 列出所有脚本
func (sm *ScriptManager) ListScripts() []Script {
	return sm.scripts
}

// UpdateScript 更新脚本
func (sm *ScriptManager) UpdateScript(id string, updatedScript Script) error {
	for i, script := range sm.scripts {
		if script.ID == id {
			updatedScript.ID = id
			updatedScript.CreatedAt = script.CreatedAt
			updatedScript.UpdatedAt = time.Now()
			sm.scripts[i] = updatedScript
			return sm.SaveScripts()
		}
	}
	return fmt.Errorf("脚本不存在: %s", id)
}

// DeleteScript 删除脚本
func (sm *ScriptManager) DeleteScript(id string) error {
	for i, script := range sm.scripts {
		if script.ID == id {
			sm.scripts = append(sm.scripts[:i], sm.scripts[i+1:]...)
			return sm.SaveScripts()
		}
	}
	return fmt.Errorf("脚本不存在: %s", id)
}

// GetTemplates 获取模板列表
func (sm *ScriptManager) GetTemplates() []ScriptTemplate {
	return sm.templates
}

// GetTemplate 获取指定模板
func (sm *ScriptManager) GetTemplate(id string) (*ScriptTemplate, error) {
	for _, template := range sm.templates {
		if template.ID == id {
			return &template, nil
		}
	}
	return nil, fmt.Errorf("模板不存在: %s", id)
}

// GenerateScriptFromTemplate 从模板生成脚本
func (sm *ScriptManager) GenerateScriptFromTemplate(templateID string, parameters map[string]string) (string, error) {
	template, err := sm.GetTemplate(templateID)
	if err != nil {
		return "", err
	}

	content := template.Template
	for param, value := range parameters {
		placeholder := "{{" + param + "}}"
		content = utils.ReplaceAll(content, placeholder, value)
	}

	return content, nil
}
