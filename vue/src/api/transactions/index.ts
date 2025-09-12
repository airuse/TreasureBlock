import request from '../request'
import { 
  handleMockGetTransactions, 
  handleMockGetTransaction,
  handleMockSearchTransactions,
  handleMockGetParsedTransaction,
  handleMockGetBTCTransactionsByBlockHeight
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

interface GetBTCTransactionsByBlockHeightRequest extends PaginationRequest, SortRequest {
  blockHeight: number
  chain?: string
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
  
  return request({
    url: `/api/v1/transactions/parsed/${hash}`,
    method: 'GET'
  })
}

/**
 * 根据区块高度获取BTC交易列表
 */
export function getBTCTransactionsByBlockHeight(data: GetBTCTransactionsByBlockHeightRequest): Promise<PaginatedResponse<Transaction>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getBTCTransactionsByBlockHeight')
    // 复用通用交易列表的Mock逻辑以返回相同结构
    return handleMockGetBTCTransactionsByBlockHeight(data as any)
  }
  
  return request({
    url: `/api/v1/transactions/btc/block-height/${data.blockHeight}`,
    method: 'GET',
    params: data
  })
}

/**
 * 根据区块高度获取BTC交易列表（公开接口，有限制）
 */
export function getBTCTransactionsByBlockHeightPublic(data: GetBTCTransactionsByBlockHeightRequest): Promise<PaginatedResponse<Transaction>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getBTCTransactionsByBlockHeightPublic')
    return handleMockGetBTCTransactionsByBlockHeight(data as any)
  } 
  
  return request({
    url: `/api/no-auth/transactions/btc/block-height/${data.blockHeight}`,
    method: 'GET',
    params: data
  })
}