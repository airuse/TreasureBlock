// 区块相关类型
export interface Block {
  // 通用字段
  hash: string
  number: number
  height?: number // 兼容BTC
  timestamp: string | number
  transactions_count: number
  transactions?: number // 兼容BTC
  size: number
  chain: string
  
  // ETH特有字段
  gasUsed?: number
  gasLimit?: number
  miner?: string
  reward?: number
  parentHash?: string
  nonce?: string
  difficulty?: number
  
  // BTC特有字段
  version?: number
  merkleRoot?: string
  bits?: string
  weight?: number
  strippedSize?: number
}

// 区块列表项类型（用于列表展示）
export interface BlockListItem {
  hash: string
  number: number
  height?: number
  timestamp: string | number
  transactions_count: number
  transactions?: number
  size: number
  chain: string
  gasUsed?: number
  gasLimit?: number
  miner?: string
  reward?: number
}

// 区块详情类型（用于详情展示）
export interface BlockDetail extends Block {
  // 可以添加更多详情字段
  confirmations?: number
  nextHash?: string
  previousHash?: string
} 