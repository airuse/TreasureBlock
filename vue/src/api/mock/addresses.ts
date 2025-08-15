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

// 响应类型 - 适配API实际返回的数据结构
interface ApiResponse<T> {
  code: number
  message: string
  data: T
  timestamp: number
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
        ...response,
        timestamp: Date.now()
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
        ...response,
        timestamp: Date.now()
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
        ...response,
        timestamp: Date.now()
      })
    }, 400) // 模拟网络延迟
  })
}
