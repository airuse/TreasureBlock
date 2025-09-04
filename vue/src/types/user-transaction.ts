// 用户交易相关类型
export interface UserTransaction {
  id: number
  user_id: number
  chain: string
  symbol: string
  from_address: string
  to_address: string
  amount: string
  fee: string
  gas_limit?: number
  gas_price?: string
  nonce?: number
  status: string
  
  // EIP-1559费率字段
  max_priority_fee_per_gas?: string
  max_fee_per_gas?: string
  tx_hash?: string
  block_height?: number
  confirmations?: number
  error_msg?: string
  remark: string
  created_at: string
  updated_at: string
  
  // ERC-20代币交易相关字段
  transaction_type?: string
  contract_operation_type?: string
  token_contract_address?: string
  token_name?: string
  token_decimals?: number
  
  // QR码导出相关字段
  chain_id?: string
  tx_data?: string
  access_list?: string
  
  // 签名组件
  v?: string
  r?: string
  s?: string
}

// 创建用户交易请求
export interface CreateUserTransactionRequest {
  chain: string
  symbol: string
  from_address: string
  to_address: string
  amount: string
  fee: string
  gas_limit?: number
  gas_price?: string
  nonce?: number
  remark?: string
  
  // ERC-20代币交易相关字段
  transaction_type?: string
  contract_operation_type?: string
  token_contract_address?: string
}

// 更新用户交易请求
export interface UpdateUserTransactionRequest {
  status?: string
  tx_hash?: string
  unsigned_tx?: string
  signed_tx?: string
  block_height?: number
  confirmations?: number
  error_msg?: string
  remark?: string
}

// 用户交易列表响应
export interface UserTransactionListResponse {
  transactions: UserTransaction[]
  total: number
  page: number
  page_size: number
  has_more: boolean
}

// 导出交易响应
export interface ExportTransactionResponse {
  unsigned_tx: string
  chain: string
  symbol: string
  from_address: string
  to_address: string
  amount: string
  fee: string
  gas_limit?: number
  gas_price?: string
  nonce?: number
  
  // EIP-1559费率字段
  max_priority_fee_per_gas?: string
  max_fee_per_gas?: string
  
  // QR码数据
  chain_id?: string
  tx_data?: string
  access_list?: string
}

// 导入签名请求
export interface ImportSignatureRequest {
  id: number
  signed_tx: string
  v?: string | null
  r?: string | null
  s?: string | null
}

// 发送交易请求
export interface SendTransactionRequest {
  id: number
}

// 用户交易统计响应
export interface UserTransactionStatsResponse {
  total_transactions: number
  draft_count: number
  unsigned_count: number
  in_progress_count: number
  packed_count: number
  confirmed_count: number
  failed_count: number
}

// 用户交易状态类型
export type UserTransactionStatus = 'draft' | 'unsigned' | 'in_progress' | 'packed' | 'confirmed' | 'failed'

// 用户交易状态文本映射
export const USER_TRANSACTION_STATUS_TEXT: Record<UserTransactionStatus, string> = {
  draft: '草稿',
  unsigned: '未签名',
  in_progress: '在途',
  packed: '已打包',
  confirmed: '已确认',
  failed: '失败'
}

// 用户交易状态样式映射
export const USER_TRANSACTION_STATUS_CLASS: Record<UserTransactionStatus, string> = {
  draft: 'bg-gray-100 text-gray-800',
  unsigned: 'bg-yellow-100 text-yellow-800',
  in_progress: 'bg-orange-100 text-orange-800',
  packed: 'bg-purple-100 text-purple-800',
  confirmed: 'bg-green-100 text-green-800',
  failed: 'bg-red-100 text-red-800'
}
