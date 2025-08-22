import request from '../request'
import { 
  handleMockGetBlocks, 
  handleMockGetBlock,
  handleMockSearchBlocks
} from '../mock/blocks'
import type { Block } from '@/types'
import type { ApiResponse, PaginatedResponse, PaginationRequest, SortRequest, SearchRequest } from '../types'

// ==================== API相关类型定义 ====================

// 请求参数类型
interface GetBlocksRequest extends PaginationRequest, SortRequest {
  chain?: string
}

interface GetBlockRequest {
  hash?: string
  height?: number
  chain?: string
}

interface SearchBlocksRequest extends SearchRequest {
  // 继承SearchRequest的query, page, page_size
}

// ==================== API函数实现 ====================

/**
 * 获取区块列表（需要认证）
 */
export function getBlocks(data: GetBlocksRequest): Promise<PaginatedResponse<Block>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getBlocks')
    return handleMockGetBlocks(data)
  }
  
  console.log('🌐 使用真实API - getBlocks (认证接口)')
  return request({
    url: '/api/v1/blocks',
    method: 'GET',
    params: data
  })
}

/**
 * 获取区块列表（公开接口，有限制）
 */
export function getBlocksPublic(data: GetBlocksRequest): Promise<PaginatedResponse<Block>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getBlocksPublic')
    return handleMockGetBlocks(data)
  }
  
  console.log('🌐 使用真实API - getBlocksPublic (公开接口)')
  return request({
    url: '/api/no-auth/blocks',
    method: 'GET',
    params: data
  })
}

/**
 * 获取区块详情
 */
export function getBlock(data: GetBlockRequest): Promise<ApiResponse<Block>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getBlock')
    return handleMockGetBlock(data)
  }
  
  console.log('🌐 使用真实API - getBlock')
  if (data.hash) {
    return request({
      url: `/api/v1/blocks/hash/${data.hash}`,
      method: 'GET',
      params: { chain: data.chain }
    })
  } else if (data.height) {
    return request({
      url: `/api/v1/blocks/height/${data.height}`,
      method: 'GET',
      params: { chain: data.chain }
    })
  } else {
    throw new Error('必须提供hash或height参数')
  }
}

/**
 * 获取区块详情（公开接口，有限制）
 */
export function getBlockPublic(data: GetBlockRequest): Promise<ApiResponse<Block>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getBlockPublic')
    return handleMockGetBlock(data)
  }
  
  console.log('🌐 使用真实API - getBlockPublic (公开接口)')
  if (data.hash) {
    return request({
      url: `/api/no-auth/blocks/hash/${data.hash}`,
      method: 'GET',
      params: { chain: data.chain }
    })
  } else if (data.height) {
    return request({
      url: `/api/no-auth/blocks/height/${data.height}`,
      method: 'GET',
      params: { chain: data.chain }
    })
  } else {
    throw new Error('必须提供hash或height参数')
  }
}

/**
 * 获取区块交易列表（需要认证）
 */
export function getBlockTransactions(data: { height: number; chain: string; page?: number; page_size?: number }): Promise<ApiResponse<any>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getBlockTransactions')
    return handleMockGetBlock(data)
  }
  
  console.log('🌐 使用真实API - getBlockTransactions')
  return request({
    url: `/api/v1/transactions/block-height/${data.height}`,
    method: 'GET',
    params: { 
      chain: data.chain,
      page: data.page || 1,
      page_size: data.page_size || 100
    }
  })
}

/**
 * 获取区块交易列表（公开接口，有限制）
 */
export function getBlockTransactionsPublic(data: { height: number; chain: string; page?: number; page_size?: number }): Promise<ApiResponse<any>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getBlockTransactionsPublic')
    return handleMockGetBlock(data)
  }
  
  console.log('🌐 使用真实API - getBlockTransactionsPublic (公开接口)')
  return request({
    url: `/api/no-auth/transactions/block-height/${data.height}`,
    method: 'GET',
    params: { 
      chain: data.chain,
      page: data.page || 1,
      page_size: data.page_size || 50
    }
  })
}

/**
 * 搜索区块
 */
export function searchBlocks(data: SearchBlocksRequest): Promise<PaginatedResponse<Block>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - searchBlocks')
    return handleMockGetBlocks(data) as Promise<PaginatedResponse<Block>>
  }
  
  console.log('🌐 使用真实API - searchBlocks')
  return request({
    url: '/api/v1/blocks/search',
    method: 'GET',
    params: data
  })
}

/**
 * 搜索区块（公开接口，有限制）
 */
export function searchBlocksPublic(data: SearchBlocksRequest): Promise<PaginatedResponse<Block>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - searchBlocksPublic')
    return handleMockGetBlocks(data) as Promise<PaginatedResponse<Block>>
  }
  
  console.log('🌐 使用真实API - searchBlocksPublic (公开接口)')
  return request({
    url: '/api/no-auth/blocks/search',
    method: 'GET',
    params: data
  })
}
