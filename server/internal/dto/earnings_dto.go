package dto

import "time"

// EarningsRecordResponse 收益记录响应DTO
type EarningsRecordResponse struct {
	ID               uint64    `json:"id"`
	UserID           uint64    `json:"user_id"`
	Amount           int64     `json:"amount"`
	Type             string    `json:"type"`
	Source           string    `json:"source"`
	SourceID         *uint64   `json:"source_id"`
	SourceChain      string    `json:"source_chain"`
	BlockHeight      *uint64   `json:"block_height"`
	TransactionCount *int64    `json:"transaction_count"`
	Description      string    `json:"description"`
	BalanceBefore    int64     `json:"balance_before"`
	BalanceAfter     int64     `json:"balance_after"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// UserBalanceResponse 用户余额响应DTO
type UserBalanceResponse struct {
	ID              uint64     `json:"id"`
	UserID          uint64     `json:"user_id"`
	Balance         int64      `json:"balance"`
	TotalEarned     int64      `json:"total_earned"`
	TotalSpent      int64      `json:"total_spent"`
	LastEarningTime *time.Time `json:"last_earning_time"`
	LastSpendTime   *time.Time `json:"last_spend_time"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// EarningsStatsResponse 收益统计响应DTO
type EarningsStatsResponse struct {
	UserID           uint64 `json:"user_id"`
	TotalEarnings    int64  `json:"total_earnings"`
	TotalSpendings   int64  `json:"total_spendings"`
	CurrentBalance   int64  `json:"current_balance"`
	BlockCount       int64  `json:"block_count"`
	TransactionCount int64  `json:"transaction_count"`
}

// TransferRequest 转账请求DTO
type TransferRequest struct {
	ToUserID    uint64 `json:"to_user_id" binding:"required"`
	Amount      int64  `json:"amount" binding:"required,gt=0"`
	Description string `json:"description"`
}

// TransferResponse 转账响应DTO
type TransferResponse struct {
	FromUserID        uint64    `json:"from_user_id"`
	ToUserID          uint64    `json:"to_user_id"`
	Amount            int64     `json:"amount"`
	Description       string    `json:"description"`
	FromBalanceBefore int64     `json:"from_balance_before"`
	FromBalanceAfter  int64     `json:"from_balance_after"`
	ToBalanceBefore   int64     `json:"to_balance_before"`
	ToBalanceAfter    int64     `json:"to_balance_after"`
	TransferTime      time.Time `json:"transfer_time"`
}

// EarningsRecordListRequest 收益记录列表请求DTO
type EarningsRecordListRequest struct {
	Page      int    `form:"page" binding:"omitempty,gt=0"`
	PageSize  int    `form:"page_size" binding:"omitempty,gt=0,lte=100"`
	Type      string `form:"type"`       // 类型过滤：add, decrease
	Source    string `form:"source"`     // 来源过滤
	Chain     string `form:"chain"`      // 链过滤
	StartDate string `form:"start_date"` // 开始日期 YYYY-MM-DD
	EndDate   string `form:"end_date"`   // 结束日期 YYYY-MM-DD
}

// BlockEarningsInfo 区块收益信息（用于扫块时计算收益）
type BlockEarningsInfo struct {
	BlockID          uint64 `json:"block_id"`
	BlockHeight      uint64 `json:"block_height"`
	Chain            string `json:"chain"`
	TransactionCount int64  `json:"transaction_count"`
	EarningsAmount   int64  `json:"earnings_amount"`
}

// EarningsTrendRequest 收益趋势请求DTO
type EarningsTrendRequest struct {
	Hours int `form:"hours" binding:"omitempty,min=1,max=24"` // 查询小时数，默认2小时
}

// EarningsTrendPoint 收益趋势数据点
type EarningsTrendPoint struct {
	Timestamp       string `json:"timestamp"`        // 时间戳 (HH:MM格式)
	Amount          int64  `json:"amount"`           // 收益数量
	BlockHeight     uint64 `json:"block_height"`     // 区块高度
	TransactionCount int64 `json:"transaction_count"` // 交易数量
	SourceChain     string `json:"source_chain"`     // 来源链
}
