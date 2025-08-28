// 个人地址管理相关类型定义

// 个人地址基础类型（与API返回数据结构匹配）
export interface PersonalAddressItem {
  id: number
  address: string
  label: string
  type: string
  balance: number
  transaction_count: number
  is_active: boolean
  created_height: number
  created_at: string
  updated_at: string
}

// 个人地址列表项类型（与API返回数据结构匹配）
export interface PersonalAddressListItem {
  id: number
  address: string
  label: string
  type: string
  balance: number
  transaction_count: number
  is_active: boolean
  created_height: number
  created_at: string
  updated_at: string
}

// 个人地址详情类型（与编辑表单匹配）
export interface PersonalAddressDetail {
  id: number
  address: string
  label: string
  balance: number
  transactionCount: number
  status: string
  createdAt: string
  updatedAt: string
  type: string
  isActive: boolean
  notes?: string
  createdHeight: number
}

// 创建个人地址请求类型
export interface CreatePersonalAddressRequest {
  address: string
  label: string
  type?: string
  notes?: string
}

// 更新个人地址请求类型
export interface UpdatePersonalAddressRequest {
  label?: string
  type?: string
  notes?: string
  isActive?: boolean
}

// 地址类型枚举
export type PersonalAddressType = 'wallet' | 'contract' | 'exchange' | 'other'

// 地址状态枚举
export type PersonalAddressStatus = 'active' | 'inactive'
