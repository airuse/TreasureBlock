// 导入接口响应数据
import apiData from '../../../ApiDatas/addresses/addresses-v1.json'
import type { Address } from '@/types'

// 请求参数类型
interface GetAddressesRequest {
  page: number
  page_size: number  // 改为page_size以匹配API
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
  page_size?: number  // 改为page_size以匹配API
}

// 响应类型 - 使用新的格式
interface ApiResponse<T> {
  success: boolean
  message?: string
  data: T
  error?: string
}

interface PaginatedResponse<T> extends ApiResponse<T[]> {
  pagination: {
    page: number
    page_size: number  // 改为page_size以匹配API
    total: number
  }
}

/**
 * 模拟获取地址列表接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockGetAddresses = (data: GetAddressesRequest): Promise<PaginatedResponse<any>> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/addresses'].get.responses['200'].content['application/json'].example
      
      resolve({
        success: true,
        data: response.data || [],
        message: 'Success',
        pagination: {
          page: data.page,
          page_size: data.page_size,
          total: response.pagination?.total || 0
        }
      })
    }, 300) // 模拟网络延迟
  })
}

/**
 * 模拟获取地址详情接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockGetAddress = (data: GetAddressRequest): Promise<ApiResponse<any>> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/addresses/{hash}'].get.responses['200'].content['application/json'].example
      
      resolve({
        success: true,
        data: response.data || response,
        message: 'Success'
      })
    }, 200) // 模拟网络延迟
  })
}

/**
 * 模拟搜索地址接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockSearchAddresses = (data: SearchAddressesRequest): Promise<ApiResponse<any[]>> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/addresses/search'].get.responses['200'].content['application/json'].example
      
      resolve({
        success: true,
        data: response.data || response,
        message: 'Success'
      })
    }, 400) // 模拟网络延迟
  })
}
