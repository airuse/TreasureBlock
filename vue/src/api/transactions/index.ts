import request from '../request'
import { 
  handleMockGetTransactions, 
  handleMockGetTransaction,
  handleMockSearchTransactions,
  handleMockGetParsedTransaction
} from '../mock/transactions'
import type { Transaction, ParsedContractResult } from '@/types'
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
 * 获取交易列表（需要认证）
 */
export function getTransactions(data: GetTransactionsRequest): Promise<PaginatedResponse<Transaction>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getTransactions')
    return handleMockGetTransactions(data)
  }
  
  console.log('🌐 使用真实API - getTransactions (认证接口)')
  return request({
    url: '/api/v1/transactions',
    method: 'GET',
    params: data
  })
}

/**
 * 获取交易列表（公开接口，有限制）
 */
export function getTransactionsPublic(data: GetTransactionsRequest): Promise<PaginatedResponse<Transaction>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getTransactionsPublic')
    return handleMockGetTransactions(data)
  }
  
  console.log('🌐 使用真实API - getTransactionsPublic (公开接口)')
  return request({
    url: '/api/no-auth/transactions',
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
    url: `/api/v1/transactions/${data.hash}`,
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
    url: '/api/v1/transactions/search',
    method: 'GET',
    params: data
  })
}

/**
 * 获取交易凭证
 */
export function getTransactionReceipt(hash: string): Promise<ApiResponse<any>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getTransactionReceipt')
    return Promise.resolve({
      success: true,
      data: {
        tx_hash: hash,
        status: 1,
        gas_used: 21000,
        cumulative_gas_used: 21000,
        transaction_index: 0,
        block_number: 9009097,
        logs_data: '[{"address":"0x...","topics":["0x..."],"data":"0x..."}]',
        contract_address: null
      }
    })
  }
  
  console.log('🌐 使用真实API - getTransactionReceipt')
  return request({
    url: `/api/v1/transactions/receipt/${hash}`,
    method: 'GET'
  })
}

/**
 * 获取交易解析结果（后端已预解析）
 */
export function getParsedTransaction(hash: string): Promise<ApiResponse<ParsedContractResult[]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getParsedTransaction')
    return handleMockGetParsedTransaction(hash)
  }
  
  console.log('🌐 使用真实API - getParsedTransaction')
  return request({
    url: `/api/v1/transactions/parsed/${hash}`,
    method: 'GET'
  })
}
