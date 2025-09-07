import request from '../request'
import { 
  handleMockLogin, 
  handleMockRegister,
  handleMockGetUserProfile,
  handleMockChangePassword,
  handleMockRefreshToken
} from '../mock/auth'
import type { 
  LoginRequest, 
  RegisterRequest, 
  UserProfile, 
  CreateAPIKeyRequest,
  GetAccessTokenRequest,
  GetAccessTokenResponse
} from '@/types/auth'
import type { ApiResponse } from '../types'

/**
 * 获取权限类型列表
 * @returns 返回结果
 */
export function getPermissionTypes(): Promise<ApiResponse<any[]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getPermissionTypes')
    return Promise.resolve({
      success: true,
      data: [
        { config_value: 'blocks:read', config_name: '区块读取权限' },
        { config_value: 'blocks:write', config_name: '区块写入权限' },
        { config_value: 'transactions:read', config_name: '交易读取权限' },
        { config_value: 'transactions:write', config_name: '交易写入权限' },
        { config_value: 'addresses:read', config_name: '地址读取权限' },
        { config_value: 'addresses:write', config_name: '地址写入权限' }
      ],
      message: 'Success'
    })
  }
  
  return request({
    url: '/api/permissions',
    method: 'GET'
  })
}

/**
 * 用户登录
 * @param data - 请求参数
 * @returns 返回结果
 */
export function login(data: LoginRequest): Promise<ApiResponse<GetAccessTokenResponse>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - login')
    return handleMockLogin(data)
  }
  
  return request({
    url: '/api/auth/login',
    method: 'POST',
    data
  })
}

/**
 * 用户注册
 * @param data - 请求参数
 * @returns 返回结果
 */
export function register(data: RegisterRequest): Promise<ApiResponse<UserProfile>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - register')
    return handleMockRegister(data)
  }
  
  return request({
    url: '/api/auth/register',
    method: 'POST',
    data
  })
}

/**
 * 获取用户资料
 * @param data - 请求参数
 * @returns 返回结果
 */
export function getUserProfile(data: { token: string }): Promise<ApiResponse<UserProfile>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - getUserProfile')
    return handleMockGetUserProfile(data)
  }
  
  return request({
    url: '/api/auth/profile',
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${data.token}`
    }
  })
}

/**
 * 修改密码
 * @param data - 请求参数
 * @returns 返回结果
 */
export function changePassword(data: { 
  token: string
  current_password: string
  new_password: string 
}): Promise<ApiResponse<{ message: string }>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - changePassword')
    return handleMockChangePassword(data)
  }
  
  const { token, ...passwordData } = data
  return request({
    url: '/api/auth/change-password',
    method: 'PUT',
    data: passwordData,
    headers: {
      'Authorization': `Bearer ${token}`
    }
  })
}

/**
 * 刷新访问令牌
 * @param data - 请求参数
 * @returns 返回结果
 */
export function refreshToken(data: { loginToken: string }): Promise<ApiResponse<GetAccessTokenResponse>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - refreshToken')
    return handleMockRefreshToken(data)
  }
  
  return request({
    url: '/api/auth/refresh',
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${data.loginToken}`
    }
  })
}
