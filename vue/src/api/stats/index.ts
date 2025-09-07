import request from '../request'
import { 
  handleMockGetNetworkStats, 
  handleMockGetLatestBlocks,
  handleMockGetLatestTransactions
} from '../mock/stats'
import type { NetworkStats } from '@/types'

// 使用统一的ApiResponse类型
import type { ApiResponse } from '../types'

// 请求参数类型
interface GetNetworkStatsRequest {
  chain?: string
  period?: string
}

interface GetLatestBlocksRequest {
  limit?: number
  chain?: string
}

interface GetLatestTransactionsRequest {
  limit?: number
  chain?: string
}

/**
 * 获取网络统计
 * @param data - 请求参数
 * @returns 返回结果
 */
export function getNetworkStats(data: GetNetworkStatsRequest): Promise<ApiResponse<any>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getNetworkStats')
    return handleMockGetNetworkStats(data)
  }
  
  return request({
    url: '/stats/network',
    method: 'GET',
    params: data
  })
}

/**
 * 获取最新区块
 * @param data - 请求参数
 * @returns 返回结果
 */
export function getLatestBlocks(data: GetLatestBlocksRequest): Promise<ApiResponse<any[]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getLatestBlocks')
    return handleMockGetLatestBlocks(data)
  }
  
  return request({
    url: '/stats/latest-blocks',
    method: 'GET',
    params: data
  })
}

/**
 * 获取最新交易
 * @param data - 请求参数
 * @returns 返回结果
 */
export function getLatestTransactions(data: GetLatestTransactionsRequest): Promise<ApiResponse<any[]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getLatestTransactions')
    return handleMockGetLatestTransactions(data)
  }
  
  return request({
    url: '/stats/latest-transactions',
    method: 'GET',
    params: data
  })
}
