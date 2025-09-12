import request from '../request'
import { 
  handleMockNoAuthGetBlocks,
  handleMockNoAuthSearchBlocks,
  handleMockNoAuthGetHomeStats,
  handleMockNoAuthGetContracts,
  handleMockNoAuthGetBlockByHeight,
  handleMockNoAuthGetTransactions,
  handleMockNoAuthGetTransactionsByBlockHeight
} from '../mock/no-auth'
import type { 
  NoAuthGetBlocksRequest,
  NoAuthGetBlockRequest,
  NoAuthSearchBlocksRequest,
  NoAuthGetTransactionsRequest,
  NoAuthGetTransactionsByBlockRequest,
  NoAuthGetHomeStatsRequest,
  NoAuthGetContractsRequest
} from '@/types'
import type { Block } from '@/types'
import type { Transaction } from '@/types'
import type { Contract } from '@/types'
import type { HomeOverview } from '@/types'
import type { ApiResponse, PaginatedResponse } from '../types'

// ==================== API函数实现 ====================

/**
 * 获取区块列表（游客模式，限制最多100个区块）
 */
export function getBlocks(data: NoAuthGetBlocksRequest): Promise<PaginatedResponse<Block>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getBlocks (游客模式)')
    return handleMockNoAuthGetBlocks(data)
  }
  
  console.log('🌐 游客模式API - getBlocks (限制最多100个区块)')
  return request({
    url: '/api/no-auth/blocks',
    method: 'GET',
    params: {
      ...data,
      page_size: Math.min(data.page_size || 25, 100) // 限制最大100个
    }
  })
}

/**
 * 根据区块高度获取区块详情（游客模式）
 */
export function getBlockByHeight(data: NoAuthGetBlockRequest): Promise<ApiResponse<Block>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getBlockByHeight (游客模式)')
    return handleMockNoAuthGetBlockByHeight(data)
  }
  
  console.log('🌐 游客模式API - getBlockByHeight')
  if (data.height) {
    return request({
      url: `/api/no-auth/blocks/height/${data.height}`,
      method: 'GET',
      params: { chain: data.chain }
    })
  } else {
    throw new Error('必须提供height参数')
  }
}

/**
 * 搜索区块（游客模式，限制最多20个结果）
 */
export function searchBlocks(data: NoAuthSearchBlocksRequest): Promise<PaginatedResponse<Block>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - searchBlocks (游客模式)')
    return handleMockNoAuthSearchBlocks(data)
  }
  
  console.log('🌐 游客模式API - searchBlocks (限制最多20个结果)')
  return request({
    url: '/api/no-auth/blocks/search',
    method: 'GET',
    params: {
      ...data,
      page_size: Math.min(data.page_size || 25, 20) // 限制最大20个
    }
  })
}

/**
 * 获取交易列表（游客模式，限制最多1000条交易）
 */
export function getTransactions(data: NoAuthGetTransactionsRequest): Promise<PaginatedResponse<Transaction>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getTransactions (游客模式)')
    return handleMockNoAuthGetTransactions(data)
  }
  
  console.log('🌐 游客模式API - getTransactions (限制最多1000条交易)')
  return request({
    url: '/api/no-auth/transactions',
    method: 'GET',
    params: {
      ...data,
      page_size: Math.min(data.page_size || 25, 1000) // 限制最大1000个
    }
  })
}

/**
 * 根据区块高度获取交易列表（游客模式，限制最多50条）
 */
export function getTransactionsByBlockHeight(data: NoAuthGetTransactionsByBlockRequest): Promise<ApiResponse<any>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getTransactionsByBlockHeight (游客模式)')
    return handleMockNoAuthGetTransactionsByBlockHeight(data)
  }
  
  console.log('🌐 游客模式API - getTransactionsByBlockHeight (限制最多50条)')
  return request({
    url: `/api/no-auth/transactions/block-height/${data.blockHeight}`,
    method: 'GET',
    params: { 
      chain: data.chain,
      page: data.page || 1,
      page_size: Math.min(data.page_size || 25, 50) // 限制最大50个
    }
  })
}

/**
 * 获取首页统计数据（游客模式）
 */
export function getHomeStats(data: NoAuthGetHomeStatsRequest): Promise<ApiResponse<HomeOverview>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getHomeStats (游客模式)')
    return handleMockNoAuthGetHomeStats(data)
  }
  
  console.log('🌐 游客模式API - getHomeStats')
  return request({
    url: '/api/no-auth/home/stats',
    method: 'GET',
    params: data
  })
}

/**
 * 获取比特币首页统计数据（游客模式）
 */
export function getBtcHomeStats(): Promise<ApiResponse<HomeOverview>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getBtcHomeStats (游客模式)')
    return handleMockNoAuthGetHomeStats({ chain: 'btc' } as any)
  }

  console.log('🌐 游客模式API - getBtcHomeStats')
  return request({
    url: '/api/no-auth/home/btc/stats',
    method: 'GET'
  })
}

/**
 * 获取合约列表（游客模式）
 */
export function getContracts(data: NoAuthGetContractsRequest): Promise<ApiResponse<Contract[]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getContracts (游客模式)')
    return handleMockNoAuthGetContracts(data)
  }
  
  console.log('🌐 游客模式API - getContracts')
  return request({
    url: '/api/no-auth/contracts',
    method: 'GET',
    params: {
      ...data,
      page_size: Math.min(data.page_size || 25, 100) // 限制最大100个
    }
  })
}

// ==================== 导出所有函数 ====================
export default {
  getBlocks,
  getBlockByHeight,
  searchBlocks,
  getTransactions,
  getTransactionsByBlockHeight,
  getHomeStats,
  getBtcHomeStats,
  getContracts
}
