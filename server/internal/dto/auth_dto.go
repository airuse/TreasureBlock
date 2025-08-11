package dto

import "time"

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"john_doe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6,max=100" example:"password123"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	UserID    uint   `json:"user_id" example:"1"`
	Username  string `json:"username" example:"john_doe"`
	Email     string `json:"email" example:"john@example.com"`
	Token     string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt int64  `json:"expires_at" example:"1672531200"`
}

// CreateAPIKeyRequest 创建API密钥请求
type CreateAPIKeyRequest struct {
	Name      string     `json:"name" binding:"required,min=1,max=100" example:"Production API Key"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
	RateLimit int        `json:"rate_limit" binding:"min=1,max=10000" example:"1000"`
}

// CreateAPIKeyResponse 创建API密钥响应
type CreateAPIKeyResponse struct {
	ID        uint       `json:"id" example:"1"`
	Name      string     `json:"name" example:"Production API Key"`
	APIKey    string     `json:"api_key" example:"ak_1234567890abcdef"`
	SecretKey string     `json:"secret_key" example:"sk_abcdef1234567890"` // 只在创建时返回一次
	ExpiresAt *time.Time `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
	RateLimit int        `json:"rate_limit" example:"1000"`
	CreatedAt time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// APIKeyResponse API密钥响应（不包含SecretKey）
type APIKeyResponse struct {
	ID         uint       `json:"id" example:"1"`
	Name       string     `json:"name" example:"Production API Key"`
	APIKey     string     `json:"api_key" example:"ak_1234567890abcdef"`
	IsActive   bool       `json:"is_active" example:"true"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" example:"2024-01-01T12:00:00Z"`
	UsageCount int64      `json:"usage_count" example:"100"`
	RateLimit  int        `json:"rate_limit" example:"1000"`
	CreatedAt  time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt  time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// GetAccessTokenRequest 获取访问令牌请求
type GetAccessTokenRequest struct {
	APIKey    string `json:"api_key" binding:"required" example:"ak_1234567890abcdef"`
	SecretKey string `json:"secret_key" binding:"required" example:"sk_abcdef1234567890"`
}

// GetAccessTokenResponse 获取访问令牌响应
type GetAccessTokenResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string `json:"token_type" example:"Bearer"`
	ExpiresIn   int64  `json:"expires_in" example:"3600"` // 秒数
	ExpiresAt   int64  `json:"expires_at" example:"1672531200"`
}

// UpdateAPIKeyRequest 更新API密钥请求
type UpdateAPIKeyRequest struct {
	Name      *string `json:"name,omitempty" binding:"omitempty,min=1,max=100" example:"Updated API Key"`
	IsActive  *bool   `json:"is_active,omitempty" example:"false"`
	RateLimit *int    `json:"rate_limit,omitempty" binding:"omitempty,min=1,max=10000" example:"2000"`
}

// UserProfileResponse 用户资料响应
type UserProfileResponse struct {
	ID        uint      `json:"id" example:"1"`
	Username  string    `json:"username" example:"john_doe"`
	Email     string    `json:"email" example:"john@example.com"`
	IsActive  bool      `json:"is_active" example:"true"`
	LastLogin *time.Time `json:"last_login,omitempty" example:"2024-01-01T12:00:00Z"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"oldpassword"`
	NewPassword     string `json:"new_password" binding:"required,min=6,max=100" example:"newpassword123"`
}

// APIUsageStatsResponse API使用统计响应
type APIUsageStatsResponse struct {
	TotalRequests   int64 `json:"total_requests" example:"1000"`
	TodayRequests   int64 `json:"today_requests" example:"50"`
	ThisHourRequests int64 `json:"this_hour_requests" example:"5"`
	AvgResponseTime float64 `json:"avg_response_time" example:"120.5"` // 毫秒
}
