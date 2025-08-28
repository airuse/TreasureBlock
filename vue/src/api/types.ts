// ==================== 通用API类型定义 ====================

// 基础响应类型
export interface ApiResponse<T> {
  success: boolean
  message?: string
  data: T
  error?: string
}

// 分页响应类型
export interface PaginatedResponse<T> extends ApiResponse<T[]> {
  pagination: {
    page: number
    page_size: number
    total: number
  }
}

// 后端实际返回的分页响应类型
export interface BackendPaginatedResponse<T> extends ApiResponse<{
  pagination: {
    page: number
    page_size: number
    total: number
  }
  records: T[]
}> {}

// 分页请求参数
export interface PaginationRequest {
  page: number
  page_size: number
}

// 排序参数
export interface SortRequest {
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

// 搜索参数
export interface SearchRequest {
  query: string
  page?: number
  page_size?: number
}


