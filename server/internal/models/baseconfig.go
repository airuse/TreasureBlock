package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseConfig 作为通用基础字典配置表，建议增加分级结构（如父子关系），以支持更复杂的字典层级和分组。
// 增加 ParentID 字段用于分级，Group 字段用于分组归类，Description 字段用于补充说明。
type BaseConfig struct {
	ID          uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Group       string         `gorm:"type:varchar(50);not null;default:'';column:group;comment:配置分组" json:"group"`
	ConfigNo    uint           `gorm:"type:int(11) unsigned;not null;column:no;comment:序号" json:"no"`
	ConfigType  uint8          `gorm:"type:tinyint(3) unsigned;not null;column:config_type;comment:配置类型" json:"config_type"`
	ConfigName  string         `gorm:"type:varchar(100);not null;column:config_name;comment:配置名称" json:"config_name"`
	ConfigKey   string         `gorm:"type:varchar(100);not null;column:config_key;comment:配置key" json:"config_key"`
	ConfigValue string         `gorm:"type:text;not null;column:config_value;comment:配置值" json:"config_value"`
	Description string         `gorm:"type:varchar(255);not null;default:'';column:description;comment:配置说明" json:"description"`
	CreatedAt   time.Time      `json:"created_at" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:created_at;comment:创建时间"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:updated_at;comment:更新时间"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (BaseConfig) TableName() string {
	return "base_config"
}
