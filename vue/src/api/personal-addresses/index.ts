import request from '../request'
import { 
  handleMockCreatePersonalAddress, 
  handleMockGetPersonalAddresses, 
  handleMockGetPersonalAddressById,
  handleMockUpdatePersonalAddress, 
  handleMockDeletePersonalAddress,
  handleMockGetAddressTransactions,
  handleMockGetAuthorizedAddresses
} from '../mock/personal-addresses'
import type { 
  PersonalAddressItem, 
  PersonalAddressDetail, 
  CreatePersonalAddressRequest, 
  UpdatePersonalAddressRequest,
  GetAuthorizedAddressesRequest,
  AuthorizedAddressesResponse
} from '@/types/personal-address'
import type { AddressTransactionsResponse } from '@/types/transaction'
import type { ApiResponse } from '../types'

// ==================== API相关类型定义 ====================

// 获取地址列表响应类型（与API实际返回结构匹配）
interface GetPersonalAddressesResponse {
  data: PersonalAddressItem[]
  message?: string
}

// ==================== API函数实现 ====================

/**
 * 创建个人地址
 */
export function createPersonalAddress(data: CreatePersonalAddressRequest): Promise<ApiResponse<PersonalAddressDetail>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - createPersonalAddress')
    return handleMockCreatePersonalAddress(data)
  }
  
  console.log('🌐 使用真实API - createPersonalAddress')
  return request({
    url: '/api/user/addresses',
    method: 'POST',
    data
  })
}

/**
 * 获取个人地址列表
 */
export function getPersonalAddresses(): Promise<ApiResponse<PersonalAddressItem[]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getPersonalAddresses')
    return handleMockGetPersonalAddresses()
  }
  
  console.log('🌐 使用真实API - getPersonalAddresses')
  return request({
    url: '/api/user/addresses',
    method: 'GET'
  })
}

/**
 * 根据ID获取个人地址详情
 */
export function getPersonalAddressById(id: number): Promise<ApiResponse<PersonalAddressDetail>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getPersonalAddressById')
    return handleMockGetPersonalAddressById(id)
  }
  
  console.log('🌐 使用真实API - getPersonalAddressById')
  return request({
    url: `/api/user/addresses/${id}`,
    method: 'GET'
  })
}

/**
 * 更新个人地址
 */
export function updatePersonalAddress(id: number, data: UpdatePersonalAddressRequest): Promise<ApiResponse<PersonalAddressDetail>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - updatePersonalAddress')
    return handleMockUpdatePersonalAddress(data)
  }
  
  console.log('🌐 使用真实API - updatePersonalAddress')
  return request({
    url: `/api/user/addresses/${id}`,
    method: 'PUT',
    data
  })
}

/**
 * 删除个人地址
 */
export function deletePersonalAddress(id: number): Promise<ApiResponse<null>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - deletePersonalAddress')
    return handleMockDeletePersonalAddress()
  }
  
  console.log('🌐 使用真实API - deletePersonalAddress')
  return request({
    url: `/api/user/addresses/${id}`,
    method: 'DELETE'
  })
}

/**
 * 获取地址相关的交易列表
 */
export function getAddressTransactions(
  address: string, 
  page: number = 1, 
  pageSize: number = 20, 
  chain?: string
): Promise<ApiResponse<AddressTransactionsResponse>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getAddressTransactions')
    return handleMockGetAddressTransactions()
  }
  
  console.log('🌐 使用真实API - getAddressTransactions')
  const params = new URLSearchParams({
    address,
    page: page.toString(),
    page_size: pageSize.toString()
  })
  
  if (chain) {
    params.append('chain', chain)
  }
  
  return request({
    url: `/api/user/addresses/transactions?${params.toString()}`,
    method: 'GET'
  })
}

/**
 * 查询授权关系
 */
export function getAuthorizedAddresses(data: GetAuthorizedAddressesRequest): Promise<ApiResponse<PersonalAddressItem[]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getAuthorizedAddresses')
    return handleMockGetAuthorizedAddresses(data)
  }
  
  console.log('🌐 使用真实API - getAuthorizedAddresses')
  return request({
    url: '/api/user/addresses/authorized',
    method: 'GET',
    params: data
  })
}
