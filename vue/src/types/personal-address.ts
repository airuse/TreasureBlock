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
