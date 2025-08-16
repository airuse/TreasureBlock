import request from '../request'
import { 
  handleMockGetAddresses, 
  handleMockGetAddress,
  handleMockSearchAddresses
} from '../mock/addresses'
import type { Address } from '@/types'

// 使用统一的ApiResponse类型
import type { ApiResponse } from '../types'

// 请求参数类型
interface GetAddressesRequest {
  page: number
  page_size: number
  type?: string
  chain?: string
  sortBy?: string
  sortOrder?: string
}

interface GetAddressRequest {
  hash: string
}

interface SearchAddressesRequest {
  query: string
  page?: number
  page_size?: number
}

// 分页响应类型
interface PaginatedResponse<T> extends ApiResponse<T[]> {
  pagination: {
    page: number
    page_size: number
    total: number
  }
}

/**
 * 获取地址列表
 */
export function getAddresses(data: GetAddressesRequest): Promise<PaginatedResponse<Address>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getAddresses')
    return handleMockGetAddresses(data)
  }
  
  console.log('🌐 使用真实API - getAddresses')
  return request({
    url: '/api/v1/addresses',
    method: 'GET',
    params: data
  })
}

/**
 * 获取地址详情
 */
export function getAddress(data: GetAddressRequest): Promise<ApiResponse<Address>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getAddress')
    return handleMockGetAddress(data)
  }
  
  console.log('🌐 使用真实API - getAddress')
  return request({
    url: `/api/v1/addresses/${data.hash}`,
    method: 'GET'
  })
}

/**
 * 搜索地址
 */
export function searchAddresses(data: SearchAddressesRequest): Promise<PaginatedResponse<Address>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - searchAddresses')
    return handleMockSearchAddresses(data) as Promise<PaginatedResponse<Address>>
  }
  
  console.log('🌐 使用真实API - searchAddresses')
  return request({
    url: '/api/v1/addresses/search',
    method: 'GET',
    params: data
  })
}
