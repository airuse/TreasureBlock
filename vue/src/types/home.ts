// 首页概览数据类型
export interface HomeOverview {
  totalBlocks: number
  totalTransactions: number
  baseFee: number
  dailyVolume: number
  avgGasPrice: number
  avgBlockTime: number
}

// 首页区块摘要类型
export interface HomeBlockSummary {
  height: number
  hash: string
  timestamp: number
  transactions_count: number
  size: number
  miner?: string
  chain: string
}

// 首页交易摘要类型
export interface HomeTransactionSummary {
  hash: string
  timestamp: number
  amount: string | number  // 可能是字符串或数字
  from: string
  to: string
  gas_price?: string | number  // 可能是字符串或数字
  gas_used?: number
  chain: string
  height?: number  // 区块高度
}

// 首页统计数据响应类型
export interface HomeStatsResponse {
  overview: HomeOverview
  latestBlocks: HomeBlockSummary[]
  latestTransactions: HomeTransactionSummary[]
}

// 首页API响应类型
export interface HomeApiResponse {
  success: boolean
  data: HomeStatsResponse
  message?: string
}
