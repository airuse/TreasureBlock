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
}

interface SearchBlocksRequest extends SearchRequest {
  // 继承SearchRequest的query, page, page_size
}

// ==================== API函数实现 ====================

/**
 * 获取区块列表
 */
export function getBlocks(data: GetBlocksRequest): Promise<PaginatedResponse<Block>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getBlocks')
    return handleMockGetBlocks(data)
  }
  
  console.log('🌐 使用真实API - getBlocks')
  return request({
    url: '/blocks',
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
      url: `/blocks/hash/${data.hash}`,
      method: 'GET'
    })
  } else if (data.height) {
    return request({
      url: `/blocks/${data.height}`,
      method: 'GET'
    })
  } else {
    throw new Error('必须提供hash或height参数')
  }
}

/**
 * 搜索区块
 */
export function searchBlocks(data: SearchBlocksRequest): Promise<PaginatedResponse<Block>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - searchBlocks')
    return handleMockSearchBlocks(data) as Promise<PaginatedResponse<Block>>
  }
  
  console.log('🌐 使用真实API - searchBlocks')
  return request({
    url: '/blocks/search',
    method: 'GET',
    params: data
  })
}
