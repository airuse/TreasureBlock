import type { Block, Transaction, Address, NetworkStats, ApiResponse } from '@/types'

// API基础配置
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'https://localhost:8443'

// 通用请求函数
async function request<T>(endpoint: string, options?: RequestInit): Promise<ApiResponse<T>> {
  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
      ...options,
    })
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    
    const data = await response.json()
    return { success: true, data }
  } catch (error) {
    return {
      success: false,
      data: null as T,
      error: error instanceof Error ? error.message : 'Unknown error'
    }
  }
}

// 区块相关API
export const blocksApi = {
  // 获取区块列表
  getBlocks: (page: number = 1, pageSize: number = 25) =>
    request<Block[]>(`/blocks?page=${page}&pageSize=${pageSize}`),
  
  // 获取区块详情
  getBlock: (height: number) =>
    request<Block>(`/blocks/${height}`),
  
  // 搜索区块
  searchBlocks: (query: string) =>
    request<Block[]>(`/blocks/search?q=${encodeURIComponent(query)}`),
}

// 交易相关API
export const transactionsApi = {
  // 获取交易列表
  getTransactions: (page: number = 1, pageSize: number = 25, status?: string) => {
    const params = new URLSearchParams({
      page: page.toString(),
      pageSize: pageSize.toString(),
    })
    if (status) params.append('status', status)
    return request<Transaction[]>(`/transactions?${params}`)
  },
  
  // 获取交易详情
  getTransaction: (hash: string) =>
    request<Transaction>(`/transactions/${hash}`),
  
  // 搜索交易
  searchTransactions: (query: string) =>
    request<Transaction[]>(`/transactions/search?q=${encodeURIComponent(query)}`),
}

// 地址相关API
export const addressesApi = {
  // 获取地址列表
  getAddresses: (page: number = 1, pageSize: number = 25, type?: string) => {
    const params = new URLSearchParams({
      page: page.toString(),
      pageSize: pageSize.toString(),
    })
    if (type) params.append('type', type)
    return request<Address[]>(`/addresses?${params}`)
  },
  
  // 获取地址详情
  getAddress: (hash: string) =>
    request<Address>(`/addresses/${hash}`),
  
  // 搜索地址
  searchAddresses: (query: string) =>
    request<Address[]>(`/addresses/search?q=${encodeURIComponent(query)}`),
}

// 统计相关API
export const statsApi = {
  // 获取网络统计
  getNetworkStats: () =>
    request<NetworkStats>('/stats/network'),
  
  // 获取最新区块
  getLatestBlocks: (limit: number = 5) =>
    request<Block[]>(`/stats/latest-blocks?limit=${limit}`),
  
  // 获取最新交易
  getLatestTransactions: (limit: number = 5) =>
    request<Transaction[]>(`/stats/latest-transactions?limit=${limit}`),
}

// 模拟数据（当后端不可用时使用）
export const mockData = {
  // 模拟网络统计
  getMockNetworkStats: (): NetworkStats => ({
    totalBlocks: 18456789,
    totalTransactions: 987654321,
    activeAddresses: 1234567,
    networkHashrate: 2.5e12,
    dailyVolume: 150000e18,
    avgGasPrice: 25e9,
    avgBlockTime: 12.5,
    difficulty: 2.5e12,
  }),
  
  // 模拟区块数据
  getMockBlocks: (count: number = 25): Block[] => {
    const blocks: Block[] = []
    for (let i = 0; i < count; i++) {
      blocks.push({
        height: 18456789 - i,
        timestamp: Math.floor(Date.now() / 1000) - i * 12,
        transactions: Math.floor(Math.random() * 200) + 50,
        size: Math.floor(Math.random() * 1000000) + 500000,
        gasUsed: Math.floor(Math.random() * 15000000) + 5000000,
        gasLimit: 30000000,
        miner: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
        reward: 2e18 + Math.random() * 1e18,
        hash: `0x${Math.random().toString(16).substring(2, 66)}`,
        parentHash: `0x${Math.random().toString(16).substring(2, 66)}`,
        nonce: Math.random().toString(16).substring(2, 18),
        difficulty: 2.5e12,
      })
    }
    return blocks
  },
  
  // 模拟交易数据
  getMockTransactions: (count: number = 25): Transaction[] => {
    const transactions: Transaction[] = []
    for (let i = 0; i < count; i++) {
      const statuses: ('success' | 'failed' | 'pending')[] = ['success', 'failed', 'pending']
      transactions.push({
        hash: `0x${Math.random().toString(16).substring(2, 66)}`,
        blockHeight: 18456789 - Math.floor(i / 200),
        timestamp: Math.floor(Date.now() / 1000) - i * 12,
        from: `0x${Math.random().toString(16).substring(2, 42)}`,
        to: `0x${Math.random().toString(16).substring(2, 42)}`,
        amount: Math.random() * 10e18,
        gasUsed: Math.floor(Math.random() * 21000) + 21000,
        gasPrice: Math.floor(Math.random() * 20e9) + 20e9,
        status: statuses[Math.floor(Math.random() * statuses.length)],
        nonce: i,
        input: '0x',
      })
    }
    return transactions
  },
} 