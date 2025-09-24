import request from '../request'
import { 
  handleMockGetGasRates, 
  handleMockGetAllGasRates
} from '../mock/gas'
import type { FeeLevels, GasConfig } from '@/types'
import type { ApiResponse } from '../types'

// ==================== API相关类型定义 ====================

// 获取Gas费率请求参数
interface GetGasRatesRequest {
  chain: string
}

// 获取所有链Gas费率响应
interface GetAllGasRatesResponse {
  [chain: string]: FeeLevels
}

// ==================== API函数实现 ====================

/**
 * 获取指定链的Gas费率
 */
export function getGasRates(data: GetGasRatesRequest): Promise<ApiResponse<FeeLevels>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getGasRates')
    return handleMockGetGasRates(data)
  }
  
  return request({
    url: '/api/user/gas',
    method: 'GET',
    params: data
  })
}

/**
 * 获取所有链的Gas费率
 */
export function getAllGasRates(): Promise<ApiResponse<GetAllGasRatesResponse>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getAllGasRates')
    return handleMockGetAllGasRates({})
  }
  
  return request({
    url: '/api/user/gas/all',
    method: 'GET'
  })
}

/**
 * 获取SOL缓存费率（无鉴权，页面初始加载用）
 */
export function getSOLGasRatesCached(): Promise<ApiResponse<FeeLevels>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getSOLGasRatesCached')
    return handleMockGetGasRates({ chain: 'sol' })
  }
  
  return request({
    url: '/api/no-auth/gas/sol',
    method: 'GET'
  })
}
