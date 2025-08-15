import request from '../request'
import { 
  handleMockCreateAPIKey, 
  handleMockGetAPIKeys,
  handleMockUpdateAPIKey,
  handleMockDeleteAPIKey,
  handleMockCreateUserAddress,
  handleMockGetUserAddresses,
  handleMockUpdateUserAddress,
  handleMockDeleteUserAddress
} from '../mock/user'
import type { 
  APIKey, 
  CreateAPIKeyRequest
} from '@/types/auth'
import type { UserAddress } from '@/types/address'

// 响应类型
interface ApiResponse<T> {
  code: number
  message: string
  data: T
  timestamp: number
}

/**
 * 创建API密钥
 * @param data - 请求参数
 * @returns 返回结果
 */
export function createAPIKey(data: CreateAPIKeyRequest & { token: string }): Promise<ApiResponse<APIKey>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - createAPIKey')
    return handleMockCreateAPIKey(data)
  }
  
  console.log('🌐 使用真实API - createAPIKey')
  const { token, ...keyData } = data
  return request({
    url: '/api/user/api-keys',
    method: 'POST',
    data: keyData,
    headers: {
      'Authorization': `Bearer ${token}`
    }
  })
}

/**
 * 获取API密钥列表
 * @param data - 请求参数
 * @returns 返回结果
 */
export function getAPIKeys(data: { token: string }): Promise<ApiResponse<APIKey[]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getAPIKeys')
    return handleMockGetAPIKeys(data)
  }
  
  console.log('🌐 使用真实API - getAPIKeys')
  return request({
    url: '/api/user/api-keys',
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${data.token}`
    }
  })
}

/**
 * 更新API密钥
 * @param data - 请求参数
 * @returns 返回结果
 */
export function updateAPIKey(data: { 
  token: string
  keyId: number
  updateData: Partial<APIKey>
}): Promise<ApiResponse<APIKey>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - updateAPIKey')
    return handleMockUpdateAPIKey(data)
  }
  
  console.log('🌐 使用真实API - updateAPIKey')
  const { token, keyId, updateData } = data
  return request({
    url: `/api/user/api-keys/${keyId}`,
    method: 'PUT',
    data: updateData,
    headers: {
      'Authorization': `Bearer ${token}`
    }
  })
}

/**
 * 删除API密钥
 * @param data - 请求参数
 * @returns 返回结果
 */
export function deleteAPIKey(data: { 
  token: string
  keyId: number
}): Promise<ApiResponse<{ message: string }>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - deleteAPIKey')
    return handleMockDeleteAPIKey(data)
  }
  
  console.log('🌐 使用真实API - deleteAPIKey')
  const { token, keyId } = data
  return request({
    url: `/api/user/api-keys/${keyId}`,
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  })
}

/**
 * 创建用户地址
 * @param data - 请求参数
 * @returns 返回结果
 */
export function createUserAddress(data: { 
  token: string
  address: string
  label: string
  type: string
  description?: string
}): Promise<ApiResponse<UserAddress>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - createUserAddress')
    return handleMockCreateUserAddress(data)
  }
  
  console.log('🌐 使用真实API - createUserAddress')
  const { token, ...addressData } = data
  return request({
    url: '/api/user/addresses',
    method: 'POST',
    data: addressData,
    headers: {
      'Authorization': `Bearer ${token}`
    }
  })
}

/**
 * 获取用户地址列表
 * @param data - 请求参数
 * @returns 返回结果
 */
export function getUserAddresses(data: { token: string }): Promise<ApiResponse<UserAddress[]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getUserAddresses')
    return handleMockGetUserAddresses(data)
  }
  
  console.log('🌐 使用真实API - getUserAddresses')
  return request({
    url: '/api/user/addresses',
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${data.token}`
    }
  })
}

/**
 * 更新用户地址
 * @param data - 请求参数
 * @returns 返回结果
 */
export function updateUserAddress(data: { 
  token: string
  addressId: number
  updateData: Partial<UserAddress>
}): Promise<ApiResponse<UserAddress>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - updateUserAddress')
    return handleMockUpdateUserAddress(data)
  }
  
  console.log('🌐 使用真实API - updateUserAddress')
  const { token, addressId, updateData } = data
  return request({
    url: `/api/user/addresses/${addressId}`,
    method: 'PUT',
    data: updateData,
    headers: {
      'Authorization': `Bearer ${token}`
    }
  })
}

/**
 * 删除用户地址
 * @param data - 请求参数
 * @returns 返回结果
 */
export function deleteUserAddress(data: { 
  token: string
  addressId: number
}): Promise<ApiResponse<{ message: string }>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - deleteUserAddress')
    return handleMockDeleteUserAddress(data)
  }
  
  console.log('🌐 使用真实API - deleteUserAddress')
  const { token, addressId } = data
  return request({
    url: `/api/user/addresses/${addressId}`,
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  })
}
