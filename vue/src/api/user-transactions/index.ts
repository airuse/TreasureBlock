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
  
  console.log('🌐 使用真实API - createUserTransaction')
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
  
  console.log('🌐 使用真实API - getUserTransactions')
  return request({
    url: '/api/user/transactions',
    method: 'GET',
    params
  })
}

/**
 * 获取用户交易统计
 */
export function getUserTransactionStats(): Promise<ApiResponse<UserTransactionStatsResponse>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getUserTransactionStats')
    return handleMockGetUserTransactionStats()
  }
  
  console.log('🌐 使用真实API - getUserTransactionStats')
  return request({
    url: '/api/user/transactions/stats',
    method: 'GET'
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
  
  console.log('🌐 使用真实API - getUserTransactionById')
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
  
  console.log('🌐 使用真实API - updateUserTransaction')
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
  
  console.log('🌐 使用真实API - deleteUserTransaction')
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
  
  console.log('🌐 使用真实API - exportTransaction')
  return request({
    url: `/api/user/transactions/${id}/export`,
    method: 'POST',
    data: feeData || {}
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
  
  console.log('🌐 使用真实API - importSignature')
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
  
  console.log('🌐 使用真实API - sendTransaction')
  return request({
    url: `/api/user/transactions/${id}/send`,
    method: 'POST'
  })
}
