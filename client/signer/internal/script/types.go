package script

import (
	"time"
)

// ScriptType 脚本类型
type ScriptType string

const (
	ScriptTypeManual   ScriptType = "manual"   // 手动创建
	ScriptTypeTemplate ScriptType = "template" // 从模板创建
)

// ScriptTemplate 脚本模板
type ScriptTemplate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Template    string   `json:"template"`   // 模板内容
	Parameters  []string `json:"parameters"` // 参数列表
}

// Script 脚本结构
type Script struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        ScriptType        `json:"type"`
	Content     string            `json:"content"`    // 脚本内容（操作码+数据）
	Parameters  map[string]string `json:"parameters"` // 参数值
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ScriptManager 脚本管理器
type ScriptManager struct {
	scripts   []Script
	templates []ScriptTemplate
	filePath  string
}

// 预定义模板
var DefaultTemplates = []ScriptTemplate{
	{
		ID:          "btc_2of3_multisig",
		Name:        "BTC 2-of-3 多签脚本",
		Description: "比特币2-of-3多重签名脚本模板",
		Template:    "OP_2 {{address1}} {{address2}} {{address3}} OP_3 OP_CHECKMULTISIG",
		Parameters:  []string{"address1", "address2", "address3"},
	},
	{
		ID:          "btc_3of5_multisig",
		Name:        "BTC 3-of-5 多签脚本",
		Description: "比特币3-of-5多重签名脚本模板",
		Template:    "OP_3 {{address1}} {{address2}} {{address3}} {{address4}} {{address5}} OP_5 OP_CHECKMULTISIG",
		Parameters:  []string{"address1", "address2", "address3", "address4", "address5"},
	},
	{
		ID:          "btc_timelock",
		Name:        "BTC 时间锁脚本",
		Description: "比特币时间锁定脚本模板",
		Template:    "OP_IF {{timelock}} OP_CHECKLOCKTIMEVERIFY OP_DROP {{address}} OP_CHECKSIG OP_ELSE {{address}} OP_CHECKSIG OP_ENDIF",
		Parameters:  []string{"timelock", "address"},
	},
}
