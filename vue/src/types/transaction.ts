// 交易相关类型
export interface Transaction {
  // 通用字段
  hash: string
  block_hash: string
  block_number: number
  blockHeight?: number // 兼容BTC
  from_address: string
  to_address: string
  from?: string // 兼容BTC
  to?: string // 兼容BTC
  value: string
  amount?: number // 兼容BTC
  gas_price: string
  gasUsed?: number // 兼容BTC
  gas_used: number
  nonce: number
  timestamp: string | number
  chain: string
  
  // BTC特有字段
  status?: 'success' | 'failed' | 'pending'
  input?: string
  fee?: number
  confirmations?: number
  
  // ETH特有字段
  gas_limit?: number
  max_fee_per_gas?: string
  max_priority_fee_per_gas?: string
}

// 个人交易类型（用于个人交易管理页面）
export interface PersonalTransaction {
  id: number
  hash: string | null // 允许为null（未生成的交易）
  from: string
  to: string
  fromAddress?: string // 兼容页面中的字段名
  toAddress?: string // 兼容页面中的字段名
  amount: number
  fee: number
  gasPrice?: number // ETH特有
  gasUsed?: number // ETH特有
  status: string
  timestamp: Date
  createdAt?: Date // 兼容某些页面使用的字段名
  confirmations: number
}

// 交易状态类型
export type TransactionStatus = 'unsigned' | 'in_progress' | 'packed' | 'confirmed'

// 交易列表项类型（用于列表展示）
export interface TransactionListItem {
  hash: string
  block_number: number
  blockHeight?: number
  from_address: string
  to_address: string
  from?: string
  to?: string
  value: string
  amount?: number
  timestamp: string | number
  chain: string
  status?: string
  confirmations?: number
} 

// 交易相关类型定义

// 地址交易响应类型
export interface AddressTransactionResponse {
  id: number
  tx_id: string
  height: number
  block_index: number
  address_from: string
  address_to: string
  amount: string
  gas_limit: number
  gas_price: string
  gas_used: number
  max_fee_per_gas: string
  max_priority_fee_per_gas: string
  effective_gas_price: string
  fee: string
  status: number
  confirm: number
  chain: string
  symbol: string
  contract_addr: string
  ctime: string
  mtime: string
}

// 地址交易列表响应类型
export interface AddressTransactionsResponse {
  transactions: AddressTransactionResponse[]
  total: number
  page: number
  page_size: number
  has_more: boolean
}

// 获取地址交易请求类型
export interface GetAddressTransactionsRequest {
  address: string
  page: number
  page_size: number
  chain?: string
} 

// 合约解析结果类型（后端预解析返回）
export interface ParsedContractResult {
  id: number
  tx_hash: string
  contract_address: string
  chain: string
  block_number: number
  log_index: number
  event_signature: string
  event_name: string
  from_address: string
  to_address: string
  amount_wei: string
  token_decimals: number
  token_symbol: string
}