// 导入接口响应数据
import apiData from '../../../ApiDatas/auth/auth-v1.json'
import type { 
  LoginRequest, 
  RegisterRequest, 
  UserProfile, 
  GetAccessTokenResponse
} from '@/types/auth'

// 响应类型 - 使用any避免复杂的类型匹配问题
interface ApiResponse<T> {
  code: number
  message: string
  data: T
  timestamp: number
}

/**
 * 模拟用户登录接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockLogin = (data: LoginRequest): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/api/auth/login'].post.responses['200'].content['application/json'].example
      
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300) // 模拟网络延迟
  })
}

/**
 * 模拟用户注册接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockRegister = (data: RegisterRequest): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/api/auth/register'].post.responses['200'].content['application/json'].example
      
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 400) // 模拟网络延迟
  })
}

/**
 * 模拟获取用户资料接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockGetUserProfile = (data: { token: string }): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/api/auth/profile'].get.responses['200'].content['application/json'].example
      
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 200) // 模拟网络延迟
  })
}

/**
 * 模拟修改密码接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockChangePassword = (data: { 
  token: string
  current_password: string
  new_password: string 
}): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/api/auth/change-password'].put.responses['200'].content['application/json'].example
      
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300) // 模拟网络延迟
  })
}

/**
 * 模拟刷新访问令牌接口
 * @param data - 请求参数
 * @returns 返回结果
 */
export const handleMockRefreshToken = (data: { loginToken: string }): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      // 从API文档中提取对应接口的响应数据
      const response = apiData.paths['/api/auth/refresh'].post.responses['200'].content['application/json'].example
      
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 200) // 模拟网络延迟
  })
}
