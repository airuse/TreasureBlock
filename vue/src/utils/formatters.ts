// 格式化时间戳
export const formatTimestamp = (timestamp: number): string => {
  // 检查时间戳是否为毫秒级（13位数字）
  const isMilliseconds = timestamp > 1000000000000
  const date = isMilliseconds ? new Date(timestamp) : new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
}

// 格式化哈希值（显示前8位和后8位）
export const formatHash = (hash: string): string => {
  if (!hash || hash.length < 16) return hash
  return `${hash.substring(0, 8)}...${hash.substring(hash.length - 8)}`
}

// 格式化完整哈希值（不截断）
export const formatFullHash = (hash: string): string => {
  if (!hash) return ''
  return hash
}

// 格式化地址（显示前6位和后4位）
export const formatAddress = (address: string): string => {
  if (!address || address.length < 10) return address
  return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`
}

// 格式化字节大小
export const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 格式化算力
export const formatHashrate = (hashrate: number): string => {
  if (hashrate === 0) return '0 H/s'
  const k = 1000
  const sizes = ['H/s', 'KH/s', 'MH/s', 'GH/s', 'TH/s']
  const i = Math.floor(Math.log(hashrate) / Math.log(k))
  return parseFloat((hashrate / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 格式化ETH金额
export const formatAmount = (amount: number): string => {
  return (amount / 1e18).toFixed(6)
}

// 格式化Gas费用
export const formatGasFee = (gasUsed: number, gasPrice: number): string => {
  const fee = gasUsed * gasPrice
  return `${(fee / 1e18).toFixed(6)} ETH`
}

// 格式化Gas使用率
export const formatGasUsage = (used: number, limit: number): string => {
  const percentage = ((used / limit) * 100).toFixed(1)
  return `${used.toLocaleString()} / ${limit.toLocaleString()} (${percentage}%)`
}

// 格式化难度
export const formatDifficulty = (difficulty: number): string => {
  if (difficulty >= 1e12) {
    return `${(difficulty / 1e12).toFixed(2)} T`
  } else if (difficulty >= 1e9) {
    return `${(difficulty / 1e9).toFixed(2)} G`
  } else if (difficulty >= 1e6) {
    return `${(difficulty / 1e6).toFixed(2)} M`
  } else {
    return difficulty.toLocaleString()
  }
}

// 格式化变化百分比
export const formatChange = (change: number): string => {
  const sign = change >= 0 ? '+' : ''
  return `${sign}${change.toFixed(2)}%`
}

// 格式化数字（添加千位分隔符）
export const formatNumber = (num: number): string => {
  return num.toLocaleString()
}

// 格式化BTC手续费
export const formatFee = (fee: number): string => {
  return fee.toFixed(8)
} 