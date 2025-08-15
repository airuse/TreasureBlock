// WebSocket消息类型定义
export interface WebSocketBlockMessage {
  height: number
  hash?: string
  timestamp?: number
  transactions?: number
  size?: number
  totalBlocks?: number
  // ETH特有字段
  gasUsed?: number
  gasLimit?: number
  miner?: string
  reward?: number
  parentHash?: string
  nonce?: string
  difficulty?: number
}

export interface WebSocketTransactionMessage {
  hash: string
  blockHash?: string
  blockNumber?: number
  blockHeight?: number
  from?: string
  to?: string
  amount?: number
  fee?: number
  timestamp?: number
  totalTransactions?: number
  // ETH特有字段
  gasUsed?: number
  gasPrice?: number
  status?: string
  nonce?: number
  input?: string
}

export interface WebSocketStatsMessage {
  totalBlocks?: number
  totalTransactions?: number
  activeAddresses?: number
  networkHashrate?: number
  avgFee?: number
  avgBlockTime?: number
  difficulty?: number
  dailyVolume?: number
  // ETH特有字段
  avgGasPrice?: number
}
