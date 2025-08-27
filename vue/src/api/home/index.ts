import request from '../request'
import { handleMockGetHomeStats } from '../mock/home'
import type { HomeApiResponse } from '@/types/home'
import type { ApiResponse } from '../types'

// ==================== API相关类型定义 ====================

// 获取首页统计数据请求参数
interface GetHomeStatsRequest {
  chain: string
}

// ==================== API函数实现 ====================

/**
 * 获取首页统计数据
 * @param data 请求参数，包含链类型
 * @returns 首页统计数据
 */
export function getHomeStats(data: GetHomeStatsRequest): Promise<HomeApiResponse> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getHomeStats')
    return handleMockGetHomeStats(data.chain)
  }
  
  console.log('🌐 使用真实API - getHomeStats')
  return request({
    url: '/api/v1/home/stats',
    method: 'GET',
    params: data
  })
}
