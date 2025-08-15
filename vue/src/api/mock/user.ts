// 导入接口响应数据
import apiData from '../../../ApiDatas/user/user-v1.json'
import type { 
  APIKey, 
  CreateAPIKeyRequest
} from '@/types/auth'
import type { UserAddress } from '@/types/address'

// 请求参数类型
interface GetAPIKeysRequest {
  token: string
}

interface CreateAPIKeyRequestData extends CreateAPIKeyRequest {
  token: string
}

interface UpdateAPIKeyRequest {
  token: string
  keyId: number
  updateData: Partial<APIKey>
}

interface DeleteAPIKeyRequest {
  token: string
  keyId: number
}

interface CreateUserAddressRequest {
  token: string
  address: string
  label: string
  type: string
  description?: string
}

interface GetUserAddressesRequest {
  token: string
}

interface UpdateUserAddressRequest {
  token: string
  addressId: number
  updateData: Partial<UserAddress>
}

interface DeleteUserAddressRequest {
  token: string
  addressId: number
}

// 响应类型 - 使用any避免复杂的类型匹配问题
interface ApiResponse<T> {
  code: number
  message: string
  data: T
  timestamp: number
}

/**
 * 模拟创建API密钥接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockCreateAPIKey = (data: CreateAPIKeyRequestData): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/api/user/api-keys'].get.responses['200'].content['application/json'].example
      
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300) // 模拟网络延迟
  })
}

/**
 * 模拟获取API密钥列表接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockGetAPIKeys = (data: GetAPIKeysRequest): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/api/user/api-keys'].get.responses['200'].content['application/json'].example
      
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 200) // 模拟网络延迟
  })
}

/**
 * 模拟更新API密钥接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockUpdateAPIKey = (data: UpdateAPIKeyRequest): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/api/user/api-keys/{keyId}'].delete.responses['200'].content['application/json'].example
      
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300) // 模拟网络延迟
  })
}

/**
 * 模拟删除API密钥接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockDeleteAPIKey = (data: DeleteAPIKeyRequest): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/api/user/api-keys/{keyId}'].delete.responses['200'].content['application/json'].example
      
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 200) // 模拟网络延迟
  })
}

/**
 * 模拟创建用户地址接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockCreateUserAddress = (data: CreateUserAddressRequest): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/api/user/addresses'].get.responses['200'].content['application/json'].example
      
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300) // 模拟网络延迟
  })
}

/**
 * 模拟获取用户地址列表接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockGetUserAddresses = (data: GetUserAddressesRequest): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/api/user/addresses'].get.responses['200'].content['application/json'].example
      
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 200) // 模拟网络延迟
  })
}

/**
 * 模拟更新用户地址接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockUpdateUserAddress = (data: UpdateUserAddressRequest): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/api/user/addresses/{addressId}'].delete.responses['200'].content['application/json'].example
      
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300) // 模拟网络延迟
  })
}

/**
 * 模拟删除用户地址接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockDeleteUserAddress = (data: DeleteUserAddressRequest): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/api/user/addresses/{addressId}'].delete.responses['200'].content['application/json'].example
      
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 200) // 模拟网络延迟
  })
}
