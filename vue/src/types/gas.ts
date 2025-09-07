// Gas费率相关类型定义

// Gas费率数据结构
export interface FeeData {
  chain: string
  base_fee: string
  max_priority_fee: string
  max_fee: string
  gas_price: string
  last_updated: number
  block_number: number
  network_congestion: string
}

// 费率等级
export interface FeeLevels {
  slow: FeeData
  normal: FeeData
  fast: FeeData
}

// Gas费率配置
export interface GasConfig {
  chain: string
  feeLevels: FeeLevels
  lastUpdated: number
}

// Gas费率统计
export interface GasStats {
  avgGasPrice: number
  avgPriorityFee: number
  networkCongestion: string
  lastUpdated: number
}
