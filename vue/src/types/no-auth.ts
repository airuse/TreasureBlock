import type { ApiResponse, PaginatedResponse, PaginationRequest, SortRequest, SearchRequest } from '@/api/types'

// ==================== 游客模式专用类型定义 ====================

/**
 * 游客模式获取区块列表请求参数
 */
export interface NoAuthGetBlocksRequest extends PaginationRequest, SortRequest {
  chain?: string
}

/**
 * 游客模式获取区块详情请求参数
 */
export interface NoAuthGetBlockRequest {
  hash?: string
  height?: number
  chain?: string
}

/**
 * 游客模式搜索区块请求参数
 */
export interface NoAuthSearchBlocksRequest extends SearchRequest {
  // 继承SearchRequest的query, page, page_size
}

/**
 * 游客模式获取交易列表请求参数
 */
export interface NoAuthGetTransactionsRequest extends PaginationRequest, SortRequest {
  chain?: string
}

/**
 * 游客模式根据区块高度获取交易列表请求参数
 */
export interface NoAuthGetTransactionsByBlockRequest extends PaginationRequest {
  blockHeight: number
  chain?: string
}

/**
 * 游客模式获取首页统计数据请求参数
 */
export interface NoAuthGetHomeStatsRequest {
  chain?: string
}

/**
 * 游客模式获取合约列表请求参数
 */
export interface NoAuthGetContractsRequest extends PaginationRequest {
  chainName?: string
  contractType?: string
  status?: string
  search?: string
}

// ==================== 游客模式响应数据类型定义 ====================

/**
 * 游客模式区块列表响应
 */
export interface NoAuthBlocksResponse extends PaginatedResponse<any> {
  // 继承PaginatedResponse的data, pagination等字段
}

/**
 * 游客模式区块详情响应
 */
export interface NoAuthBlockResponse extends ApiResponse<any> {
  // 继承ApiResponse的success, data, message等字段
}

/**
 * 游客模式交易列表响应
 */
export interface NoAuthTransactionsResponse extends PaginatedResponse<any> {
  // 继承PaginatedResponse的data, pagination等字段
}

/**
 * 游客模式首页统计响应
 */
export interface NoAuthHomeStatsResponse extends ApiResponse<any> {
  // 继承ApiResponse的success, data, message等字段
}

/**
 * 游客模式合约列表响应
 */
export interface NoAuthContractsResponse extends ApiResponse<any[]> {
  // 继承ApiResponse的success, data, message等字段
}
