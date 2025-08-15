// 地址相关类型
export interface Address {
  address: string
  balance: string
  transaction_count: number
  chain: string
}

// 用户地址类型
export interface UserAddress {
  id: number
  address: string
  label: string
  type: string
  balance: number
  transaction_count: number
  is_active: boolean
  created_at: string
  updated_at: string
}

// 地址详细信息类型（用于BTC地址页面）
export interface AddressData {
  address: string
  balance: number
  txCount: number
  utxoCount: number
  type: string
  firstSeen: number
  lastSeen: number
  transactions: Array<{
    hash: string
    timestamp: number
    type: string
    amount: number
    balance: number
  }>
  utxos: Array<{
    txHash: string
    vout: number
    amount: number
    confirmations: number
    timestamp: number
  }>
}

// 个人地址类型（用于个人地址管理页面）
export interface PersonalAddress {
  id: number
  address: string
  label: string
  balance: number
  transactionCount: number
  status: string
}

// 地址类型枚举
export type AddressType = 'wallet' | 'contract' | 'exchange' | 'other'

// 地址状态枚举
export type AddressStatus = 'active' | 'inactive' 