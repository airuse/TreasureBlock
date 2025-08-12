package dto

// CreateUserAddressRequest 创建用户地址请求
type CreateUserAddressRequest struct {
	Address string `json:"address" binding:"required"`
	Label   string `json:"label"`
	Type    string `json:"type" binding:"required,oneof=wallet contract exchange other"`
}

// UpdateUserAddressRequest 更新用户地址请求
type UpdateUserAddressRequest struct {
	Label    *string `json:"label"`
	Type     *string `json:"type" binding:"omitempty,oneof=wallet contract exchange other"`
	IsActive *bool   `json:"is_active"`
}

// UserAddressResponse 用户地址响应
type UserAddressResponse struct {
	ID               uint    `json:"id"`
	Address          string  `json:"address"`
	Label            string  `json:"label"`
	Type             string  `json:"type"`
	Balance          float64 `json:"balance"`
	TransactionCount int64   `json:"transaction_count"`
	IsActive         bool    `json:"is_active"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}
