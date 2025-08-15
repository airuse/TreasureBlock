import request from '../request'
import { 
  handleMockGetTransactions, 
  handleMockGetTransaction,
  handleMockSearchTransactions
} from '../mock/transactions'
import type { Transaction } from '@/types'
import type { ApiResponse, PaginatedResponse, PaginationRequest, SortRequest, SearchRequest } from '../types'

// ==================== API相关类型定义 ====================

// 请求参数类型
interface GetTransactionsRequest extends PaginationRequest, SortRequest {
  status?: string
  chain?: string
}

interface GetTransactionRequest {
  hash: string
}

interface SearchTransactionsRequest extends SearchRequest {
  // 继承SearchRequest的query, page, page_size
}

// ==================== API函数实现 ====================

/**
 * 获取交易列表
 */
export function getTransactions(data: GetTransactionsRequest): Promise<PaginatedResponse<Transaction>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getTransactions')
    return handleMockGetTransactions(data)
  }
  
  console.log('🌐 使用真实API - getTransactions')
  return request({
    url: '/transactions',
    method: 'GET',
    params: data
  })
}

/**
 * 获取交易详情
 */
export function getTransaction(data: GetTransactionRequest): Promise<ApiResponse<Transaction>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getTransaction')
    return handleMockGetTransaction(data)
  }
  
  console.log('🌐 使用真实API - getTransaction')
  return request({
    url: `/transactions/${data.hash}`,
    method: 'GET'
  })
}

/**
 * 搜索交易
 */
export function searchTransactions(data: SearchTransactionsRequest): Promise<PaginatedResponse<Transaction>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - searchTransactions')
    return handleMockSearchTransactions(data) as Promise<PaginatedResponse<Transaction>>
  }
  
  console.log('🌐 使用真实API - searchTransactions')
  return request({
    url: '/transactions/search',
    method: 'GET',
    params: data
  })
}
