// 个人地址管理相关类型定义

// 个人地址基础类型（与API返回数据结构匹配）
export interface PersonalAddressItem {
  id: number
  address: string
  label: string
  type: string
  contract_id?: number
  authorized_addresses?: { [address: string]: { allowance: string } }
  notes?: string
  balance?: string
  contract_balance?: string
  contract_balance_height?: number
  transaction_count: number
  utxo_count?: number  // UTXO数量（仅BTC使用）
  is_active: boolean
  balance_height: number
  created_at: string
  updated_at: string
}

// 个人地址列表项类型（与API返回数据结构匹配）
export interface PersonalAddressListItem {
  id: number
  address: string
  label: string
  type: string
  contract_id?: number
  authorized_addresses?: { [address: string]: { allowance: string } }
  notes?: string
  balance?: string
  contract_balance?: string
  contract_balance_height?: number
  transaction_count: number
  utxo_count?: number  // UTXO数量（仅BTC使用）
  is_active: boolean
  balance_height: number
  created_at: string
  updated_at: string
}

// 个人地址详情类型（与编辑表单匹配）
export interface PersonalAddressDetail {
  id: number
  address: string
  label: string
  balance?: string
  transactionCount: number
  status: string
  createdAt: string
  updatedAt: string
  type: string
  contract_id?: number  // 使用与后端一致的字段名
  authorized_addresses?: string[]
  isActive: boolean
  notes?: string
  balanceHeight: number
}

// 创建个人地址请求类型
export interface CreatePersonalAddressRequest {
  address: string
  label: string
  chain?: string  // 区块链类型：eth, btc, sol, other
  type?: string
  contract_id?: number  // 使用与后端一致的字段名
  authorized_addresses?: string[]
  notes?: string
}

// 更新个人地址请求类型
export interface UpdatePersonalAddressRequest {
  label?: string
  type?: string
  contract_id?: number  // 使用与后端一致的字段名
  authorized_addresses?: string[]
  contract_balance?: string
  contract_balance_height?: number
  notes?: string
  isActive?: boolean
}

// 地址类型枚举
export type PersonalAddressType = 'wallet' | 'contract' | 'authorized_contract' | 'exchange' | 'other'

// 地址状态枚举
export type PersonalAddressStatus = 'active' | 'inactive'

// 授权地址查询请求类型
export interface GetAuthorizedAddressesRequest {
  spender_address: string
}

// 授权地址响应类型
export interface AuthorizedAddressesResponse {
  authorized_addresses: PersonalAddressItem[]
}

// BTC UTXO类型定义
export interface BTCUTXO {
  id: number
  chain: string
  tx_id: string
  vout_index: number
  block_height: number
  block_id?: number
  address: string
  script_pub_key: string
  script_type: string
  is_coinbase: boolean
  value_satoshi: number
  spent_tx_id?: string
  spent_vin_index?: number
  spent_height?: number
  spent_at?: string
  ctime: string
  mtime: string
}
