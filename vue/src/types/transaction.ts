// 交易相关类型
export interface Transaction {
  hash: string
  blockHeight: number
  timestamp: number
  from: string
  to: string
  amount: number
  gasUsed: number
  gasPrice: number
  status: 'success' | 'failed' | 'pending'
  nonce: number
  input: string
} 