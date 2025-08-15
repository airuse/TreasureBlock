// 统计相关类型
export interface NetworkMetric {
  totalBlocks: number
  totalTransactions: number
  activeAddresses: number
  networkHashrate: number
  avgGasPrice: number
  avgBlockTime: number
  difficulty: number
  dailyVolume: number
}

// 网络指标显示类型（用于页面展示）
export interface NetworkMetricDisplay {
  name: string
  currentValue: string
  change24h: number
  change7d: number
}

// 区块统计类型
export interface BlockStats {
  totalBlocks: number
  avgBlockTime: number
  avgBlockSize: number
  avgGasUsed: number
  avgGasLimit: number
  avgReward: number
}

// 交易统计类型
export interface TransactionStats {
  totalTransactions: number
  avgTransactionValue: number
  avgGasPrice: number
  avgGasUsed: number
  successRate: number
  pendingCount: number
}

// 地址统计类型
export interface AddressStats {
  totalAddresses: number
  activeAddresses: number
  newAddresses: number
  topAddresses: Array<{
    address: string
    balance: number
    transactionCount: number
  }>
}

// 网络状态类型
export interface NetworkStatus {
  isOnline: boolean
  lastBlockTime: number
  syncStatus: 'syncing' | 'synced' | 'error'
  peerCount: number
  version: string
}
