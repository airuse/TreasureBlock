import request from '../request'
import { 
  handleMockCreatePersonalAddress, 
  handleMockGetPersonalAddresses, 
  handleMockGetPersonalAddressById,
  handleMockUpdatePersonalAddress, 
  handleMockDeletePersonalAddress,
  handleMockGetAddressTransactions,
  handleMockGetAuthorizedAddresses,
  handleMockGetUserAddressesByPending
} from '../mock/personal-addresses'
import type { 
  PersonalAddressItem, 
  PersonalAddressDetail, 
  CreatePersonalAddressRequest, 
  UpdatePersonalAddressRequest,
  GetAuthorizedAddressesRequest,
  AuthorizedAddressesResponse,
  BTCUTXO,
  UserAddressPendingItem
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
  
  return request({
    url: '/api/user/addresses',
    method: 'POST',
    data
  })
}

/**
 * 获取个人地址列表
 */
export function getPersonalAddresses(chain: string): Promise<ApiResponse<PersonalAddressItem[]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getPersonalAddresses')
    return handleMockGetPersonalAddresses()
  }
  
  return request({
    url: `/api/user/addresses/chain/${chain}`,
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
  
  return request({
    url: '/api/user/addresses/authorized',
    method: 'GET',
    params: data
  })
}

/**
 * 刷新个人地址（钱包/合约/授权）余额
 */
export function refreshPersonalAddressBalance(id: number): Promise<ApiResponse<PersonalAddressDetail>> {
  return request({
    url: `/api/user/addresses/${id}/refresh-balance`,
    method: 'POST'
  })
}

/**
 * 获取地址的UTXO列表（仅BTC）
 */
export function getAddressUTXOs(address: string): Promise<ApiResponse<BTCUTXO[]>> {
  return request({
    url: '/api/user/addresses/utxos',
    method: 'GET',
    params: { address }
  })
}

/**
 * 获取用户所有在途交易地址
 */
export function getUserAddressesByPending(chain: string): Promise<ApiResponse<UserAddressPendingItem[]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getUserAddressesByPending')
    return handleMockGetUserAddressesByPending(chain)
  }
  
  return request({
    url: '/api/user/addresses/pending',
    method: 'GET',
    params: { chain }
  })
}
