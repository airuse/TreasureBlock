import request from '../request'
import { 
  handleMockCreateUserTransaction,
  handleMockGetUserTransactions,
  handleMockGetUserTransactionStats,
  handleMockGetUserTransactionById,
  handleMockUpdateUserTransaction,
  handleMockDeleteUserTransaction,
  handleMockExportTransaction,
  handleMockImportSignature,
  handleMockSendTransaction
} from '../mock/user-transactions'
import type { 
  CreateUserTransactionRequest, 
  UpdateUserTransactionRequest, 
  UserTransactionListResponse,
  UserTransaction,
  ExportTransactionResponse,
  ImportSignatureRequest,
  SendTransactionRequest,
  UserTransactionStatsResponse
} from '@/types'
import type { ApiResponse, PaginatedResponse, PaginationRequest, SortRequest, SearchRequest } from '../types'

// ==================== SOL 专用类型定义 ====================
export interface SolExportTransactionResponse {
  id: number
  chain: string
  recent_blockhash: string
  fee_payer: string
  version: string
  instructions: Array<Record<string, any>>
  context?: Record<string, any>
}

// ==================== API相关类型定义 ====================

// 获取用户交易列表请求参数 - 继承通用类型
interface GetUserTransactionsRequest extends PaginationRequest, SortRequest {
  status?: string
  chain?: string
  query?: string
}

// ==================== API函数实现 ====================

/**
 * 创建用户交易
 */
export function createUserTransaction(data: CreateUserTransactionRequest): Promise<ApiResponse<UserTransaction>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - createUserTransaction')
    return handleMockCreateUserTransaction(data)
  }
  
  return request({
    url: '/api/user/transactions',
    method: 'POST',
    data
  })
}

/**
 * 获取用户交易列表
 */
export function getUserTransactions(params: GetUserTransactionsRequest): Promise<ApiResponse<UserTransactionListResponse>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getUserTransactions')
    return handleMockGetUserTransactions(params)
  }
  
  return request({
    url: '/api/user/transactions',
    method: 'GET',
    params
  })
}

/**
 * 获取用户交易统计
 */
export function getUserTransactionStats(params?: { chain?: string }): Promise<ApiResponse<UserTransactionStatsResponse>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getUserTransactionStats')
    return handleMockGetUserTransactionStats()
  }
  
  return request({
    url: '/api/user/transactions/stats',
    method: 'GET',
    params
  })
}

/**
 * 根据ID获取用户交易
 */
export function getUserTransactionById(id: number): Promise<ApiResponse<UserTransaction>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getUserTransactionById')
    return handleMockGetUserTransactionById(id)
  }
  
  return request({
    url: `/api/user/transactions/${id}`,
    method: 'GET'
  })
}

/**
 * 更新用户交易
 */
export function updateUserTransaction(id: number, data: UpdateUserTransactionRequest): Promise<ApiResponse<UserTransaction>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - updateUserTransaction')
    return handleMockUpdateUserTransaction(id, data)
  }
  
  return request({
    url: `/api/user/transactions/${id}`,
    method: 'PUT',
    data
  })
}

/**
 * 删除用户交易
 */
export function deleteUserTransaction(id: number): Promise<ApiResponse<null>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - deleteUserTransaction')
    return handleMockDeleteUserTransaction(id)
  }
  
  return request({
    url: `/api/user/transactions/${id}`,
    method: 'DELETE'
  })
}

/**
 * 导出交易
 */
export function exportTransaction(id: number, feeData?: any): Promise<ApiResponse<ExportTransactionResponse>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - exportTransaction')
    return handleMockExportTransaction(id)
  }
  
  return request({
    url: `/api/user/transactions/${id}/export`,
    method: 'POST',
    data: feeData || {}
  })
}

// ==================== SOL 专用接口 ====================

/**
 * SOL: 导出未签名交易
 */
export function exportSolUnsigned(id: number): Promise<ApiResponse<SolExportTransactionResponse>> {
  return request({
    url: `/api/user/transactions/${id}/sol/export-unsigned`,
    method: 'GET'
  })
}

/**
 * SOL: 导入签名（base64 原始交易）
 */
export function importSolSignature(id: number, data: { id: number; signed_base64: string }): Promise<ApiResponse<any>> {
  return request({
    url: `/api/user/transactions/${id}/sol/import-signature`,
    method: 'POST',
    data
  })
}

/**
 * SOL: 发送交易
 */
export function sendSolTransaction(id: number): Promise<ApiResponse<any>> {
  return request({
    url: `/api/user/transactions/${id}/sol/send`,
    method: 'POST'
  })
}

/**
 * 导入签名
 */
export function importSignature(id: number, data: ImportSignatureRequest): Promise<ApiResponse<UserTransaction>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - importSignature')
    return handleMockImportSignature(id, data)
  }
  
  return request({
    url: `/api/user/transactions/${id}/import-signature`,
    method: 'POST',
    data
  })
}

/**
 * 发送交易
 */
export function sendTransaction(id: number): Promise<ApiResponse<UserTransaction>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - sendTransaction')
    return handleMockSendTransaction(id)
  }
  
  return request({
    url: `/api/user/transactions/${id}/send`,
    method: 'POST'
  })
}
