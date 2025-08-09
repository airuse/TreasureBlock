// 区块相关类型
export interface Block {
  height: number
  timestamp: number
  transactions: number
  size: number
  gasUsed: number
  gasLimit: number
  miner: string
  reward: number
  hash: string
  parentHash: string
  nonce: string
  difficulty: number
} 