package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"type:varchar(191);uniqueIndex;not null"`
	Email     string    `json:"email" gorm:"type:varchar(191);uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"type:varchar(255);not null"`                   // 密码不返回给前端
	Role      string    `json:"role" gorm:"type:varchar(191);default:'user';not null"` // 用户角色：administrator, user
	Status    int       `json:"status" gorm:"default:1"`                               // 状态：1-启用，0-禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Role 角色模型
type Role struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"type:varchar(191);uniqueIndex;not null"` // 角色名称
	DisplayName string       `json:"display_name" gorm:"type:varchar(191);not null"`     // 显示名称
	Description string       `json:"description" gorm:"type:varchar(255)"`               // 角色描述
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`     // 角色权限
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Permission 权限模型
type Permission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"type:varchar(191);uniqueIndex;not null"` // 权限名称
	DisplayName string    `json:"display_name" gorm:"type:varchar(191);not null"`     // 显示名称
	Description string    `json:"description" gorm:"type:varchar(255)"`               // 权限描述
	Resource    string    `json:"resource" gorm:"type:varchar(64);not null"`          // 资源名称（如：contract, coinconfig）
	Action      string    `json:"action" gorm:"type:varchar(32);not null"`            // 操作类型（如：create, read, update, delete）
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RolePermission 角色权限关联表
type RolePermission struct {
	RoleID       uint `json:"role_id" gorm:"primaryKey"`
	PermissionID uint `json:"permission_id" gorm:"primaryKey"`
}

// UserRole 用户角色关联表（支持多角色）
type UserRole struct {
	UserID uint `json:"user_id" gorm:"primaryKey"`
	RoleID uint `json:"role_id" gorm:"primaryKey"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (Role) TableName() string {
	return "roles"
}

func (Permission) TableName() string {
	return "permissions"
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

func (UserRole) TableName() string {
	return "user_roles"
}

// APIKey API密钥模型
type APIKey struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Name        string         `json:"name" gorm:"type:varchar(100);not null"` // 密钥名称，便于用户管理
	APIKey      string         `json:"api_key" gorm:"type:varchar(64);uniqueIndex;not null"`
	SecretKey   string         `json:"-" gorm:"type:varchar(128);not null"`   // 不返回到前端
	Permissions string         `json:"permissions" gorm:"type:text;not null"` // JSON格式的权限数组
	RateLimit   int64          `json:"rate_limit" gorm:"default:1000"`        // 每小时请求限制
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	ExpiresAt   *time.Time     `json:"expires_at"` // 过期时间，可以为空表示永久有效
	LastUsedAt  *time.Time     `json:"last_used_at"`
	UsageCount  int64          `json:"usage_count" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// 关联的用户
	User User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
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

// RequestLog 请求日志模型
type RequestLog struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id" gorm:"index;not null"`
	APIKeyID   uint           `json:"api_key_id" gorm:"index;not null"`
	Method     string         `json:"method" gorm:"type:varchar(16);not null"`
	Path       string         `json:"path" gorm:"type:varchar(512);not null"`
	StatusCode int            `json:"status_code"`
	Duration   int64          `json:"duration"` // 毫秒
	IP         string         `json:"ip" gorm:"type:varchar(64)"`
	UserAgent  string         `json:"user_agent" gorm:"type:varchar(256)"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	APIKey APIKey `json:"api_key,omitempty" gorm:"foreignKey:APIKeyID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (RequestLog) TableName() string {
	return "request_logs"
}
