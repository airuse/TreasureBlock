// ==================== 收益模块业务实体类型定义 ====================

// 收益记录基础类型 - 匹配后端DTO
export interface EarningRecord {
  id: number
  user_id: number
  amount: number
  type: string
  source: string
  source_id?: number
  source_chain: string
  block_height?: number
  transaction_count?: number
  description?: string
  balance_before: number
  balance_after: number
  created_at: string
  updated_at: string
}

// 收益趋势数据点类型
export interface EarningsTrendPoint {
  timestamp: string
  amount: number
  block_height: number
  transaction_count: number
  source_chain: string
}

// 收益记录列表项类型 - 用于前端显示
export interface EarningRecordListItem {
  id: number
  block_hash: string
  block_height: number
  amount: number
  status: 'pending' | 'confirmed' | 'failed'
  timestamp: string
}

// 收益记录详情类型
export interface EarningRecordDetail extends EarningRecord {
  transaction_hash?: string
  gas_used?: number
  gas_price?: string
  network_fee?: number
  net_amount?: number
  description?: string
}

// 用户余额类型 - 匹配后端DTO
export interface UserBalance {
  id: number
  user_id: number
  balance: number
  total_earned: number
  total_spent: number
  last_earning_time?: string
  last_spend_time?: string
  created_at: string
  updated_at: string
}

// 收益统计类型 - 匹配后端DTO
export interface EarningsStats {
  user_id: number
  total_earnings: number
  total_spendings: number
  current_balance: number
  block_count: number
  transaction_count: number
}

// 转账记录类型
export interface TransferRecord {
  id: number
  from_user_id: number
  to_user_id: number
  amount: number
  status: 'pending' | 'completed' | 'failed'
  created_at: string
  completed_at?: string
  description?: string
}
