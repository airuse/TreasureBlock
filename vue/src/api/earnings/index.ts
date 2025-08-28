import request from '../request'
import { 
  handleMockGetUserBalance, 
  handleMockGetUserEarningsRecords, 
  handleMockGetEarningsRecordDetail, 
  handleMockGetUserEarningsStats, 
  handleMockTransferTCoins,
  handleMockGetEarningsTrend
} from '../mock/earnings'
import type { 
  EarningRecord, 
  EarningRecordListItem, 
  EarningRecordDetail, 
  UserBalance, 
  EarningsStats, 
  TransferRecord,
  EarningsTrendPoint
} from '@/types/earnings'
import type { ApiResponse, PaginatedResponse, BackendPaginatedResponse, PaginationRequest, SortRequest } from '../types'

// ==================== API相关类型定义 ====================

// 获取收益记录列表请求参数
interface GetEarningsRecordsRequest extends PaginationRequest, SortRequest {
  period?: number // 时间周期：7, 30, 90天
  status?: 'pending' | 'confirmed' | 'failed'
  start_date?: string
  end_date?: string
}

// 转账T币请求参数
interface TransferTCoinsRequest {
  to_user_id: number
  amount: number
  description?: string
}

// ==================== API函数实现 ====================

/**
 * 获取用户余额
 */
export function getUserBalance(): Promise<ApiResponse<UserBalance>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getUserBalance')
    return handleMockGetUserBalance()
  }
  
  console.log('🌐 使用真实API - getUserBalance')
  return request({
    url: '/api/v1/earnings/balance',
    method: 'GET'
  })
}

/**
 * 获取用户收益记录列表
 */
export function getUserEarningsRecords(data: GetEarningsRecordsRequest): Promise<BackendPaginatedResponse<EarningRecord>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getUserEarningsRecords')
    return handleMockGetUserEarningsRecords(data)
  }
  
  console.log('🌐 使用真实API - getUserEarningsRecords')
  return request({
    url: '/api/v1/earnings/records',
    method: 'GET',
    params: data
  })
}

/**
 * 获取收益记录详情
 */
export function getEarningsRecordDetail(id: number): Promise<ApiResponse<EarningRecordDetail>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getEarningsRecordDetail')
    return handleMockGetEarningsRecordDetail(id)
  }
  
  console.log('🌐 使用真实API - getEarningsRecordDetail')
  return request({
    url: `/api/v1/earnings/records/${id}`,
    method: 'GET'
  })
}

/**
 * 获取用户收益统计
 */
export function getUserEarningsStats(): Promise<ApiResponse<EarningsStats>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getUserEarningsStats')
    return handleMockGetUserEarningsStats()
  }
  
  console.log('🌐 使用真实API - getUserEarningsStats')
  return request({
    url: '/api/v1/earnings/stats',
    method: 'GET'
  })
}

/**
 * 获取收益趋势数据
 */
export function getEarningsTrend(hours: number = 2): Promise<ApiResponse<EarningsTrendPoint[]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getEarningsTrend')
    return handleMockGetEarningsTrend(hours)
  }
  
  console.log('🌐 使用真实API - getEarningsTrend')
  return request({
    url: '/api/v1/earnings/trend',
    method: 'GET',
    params: { hours }
  })
}

/**
 * 转账T币
 */
export function transferTCoins(data: TransferTCoinsRequest): Promise<ApiResponse<TransferRecord>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - transferTCoins')
    return handleMockTransferTCoins(data)
  }
  
  console.log('🌐 使用真实API - transferTCoins')
  return request({
    url: '/api/v1/earnings/transfer',
    method: 'POST',
    data
  })
}
