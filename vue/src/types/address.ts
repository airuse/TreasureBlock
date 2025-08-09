// 地址相关类型
export interface Address {
  hash: string
  type: 'contract' | 'wallet'
  balance: number
  transactionCount: number
  lastActivity: number
  label?: string
} 