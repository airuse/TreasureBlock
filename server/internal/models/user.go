package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"type:varchar(100);uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"type:varchar(255);not null"` // 密码不返回到前端
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time     `json:"last_login"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 关联的API密钥
	APIKeys []APIKey `json:"api_keys,omitempty" gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "users"
}

// APIKey API密钥模型
type APIKey struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id" gorm:"not null;index"`
	Name       string         `json:"name" gorm:"type:varchar(100);not null"` // 密钥名称，便于用户管理
	APIKey     string         `json:"api_key" gorm:"type:varchar(64);uniqueIndex;not null"`
	SecretKey  string         `json:"-" gorm:"type:varchar(128);not null"` // 不返回到前端
	IsActive   bool           `json:"is_active" gorm:"default:true"`
	ExpiresAt  *time.Time     `json:"expires_at"` // 过期时间，可以为空表示永久有效
	LastUsedAt *time.Time     `json:"last_used_at"`
	UsageCount int64          `json:"usage_count" gorm:"default:0"`
	RateLimit  int            `json:"rate_limit" gorm:"default:1000"` // 每小时请求限制
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 关联的用户
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (APIKey) TableName() string {
	return "api_keys"
}

// AccessToken 访问令牌记录模型（用于追踪令牌使用情况）
type AccessToken struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null;index"`
	APIKeyID  uint           `json:"api_key_id" gorm:"not null;index"`
	TokenHash string         `json:"-" gorm:"type:varchar(255);not null"` // JWT令牌的哈希值
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	IsRevoked bool           `json:"is_revoked" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 关联
	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	APIKey APIKey `json:"api_key,omitempty" gorm:"foreignKey:APIKeyID"`
}

func (AccessToken) TableName() string {
	return "access_tokens"
}

// RequestLog 请求日志模型（用于监控API使用情况）
type RequestLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null;index"`
	APIKeyID   uint      `json:"api_key_id" gorm:"not null;index"`
	Method     string    `json:"method" gorm:"type:varchar(10);not null"`
	Path       string    `json:"path" gorm:"type:varchar(500);not null"`
	StatusCode int       `json:"status_code" gorm:"not null"`
	Duration   int64     `json:"duration"` // 请求耗时(毫秒)
	IP         string    `json:"ip" gorm:"type:varchar(45)"`
	UserAgent  string    `json:"user_agent" gorm:"type:varchar(500)"`
	CreatedAt  time.Time `json:"created_at"`

	// 关联
	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	APIKey APIKey `json:"api_key,omitempty" gorm:"foreignKey:APIKeyID"`
}

func (RequestLog) TableName() string {
	return "request_logs"
}
