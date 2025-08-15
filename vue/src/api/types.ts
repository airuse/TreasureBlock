// ==================== 通用API类型定义 ====================

// 基础响应类型
export interface ApiResponse<T> {
  code: number
  message: string
  data: T
  timestamp: number
}

// 分页响应类型
export interface PaginatedResponse<T> extends ApiResponse<T[]> {
  pagination: {
    page: number
    page_size: number
    total: number
  }
}

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
