import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { 
  login,
  register,
  getUserProfile,
  changePassword,
  refreshToken
} from '@/api/auth'
import { 
  createAPIKey,
  getAPIKeys,
  updateAPIKey,
  deleteAPIKey,
  createUserAddress,
  getUserAddresses,
  updateUserAddress,
  deleteUserAddress
} from '@/api/user'
import type { 
  UserProfile, 
  APIKey, 
  LoginRequest, 
  RegisterRequest,
  CreateAPIKeyRequest,
  GetAccessTokenRequest
} from '../types/auth'

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const isAuthenticated = ref(false)
  const user = ref<UserProfile | null>(null)
  const loginToken = ref<string | null>(null)
  const accessToken = ref<string | null>(null)
  const apiKeys = ref<APIKey[]>([]) // 确保初始化为空数组
  const currentAPIKey = ref<APIKey | null>(null)

  // 计算属性
  const hasValidTokens = computed(() => {
    return isAuthenticated.value && loginToken.value && accessToken.value
  })

  const isTokenExpired = computed(() => {
    if (!loginToken.value) return true
    // 这里可以添加令牌过期检查逻辑
    return false
  })

  // 初始化 - 从localStorage恢复状态
  const initialize = () => {
    const savedLoginToken = localStorage.getItem('loginToken')
    const savedAccessToken = localStorage.getItem('accessToken')
    const savedUser = localStorage.getItem('user')
    const savedAPIKeys = localStorage.getItem('apiKeys')
    const savedCurrentAPIKey = localStorage.getItem('currentAPIKey')

    if (savedLoginToken && savedUser) {
      loginToken.value = savedLoginToken
      user.value = JSON.parse(savedUser)
      isAuthenticated.value = true
    }

    if (savedAccessToken) {
      accessToken.value = savedAccessToken
    }

    if (savedAPIKeys) {
      try {
        apiKeys.value = JSON.parse(savedAPIKeys) || []
      } catch (error) {
        console.error('Failed to parse saved API keys:', error)
        apiKeys.value = []
      }
    } else {
      apiKeys.value = [] // 确保初始化为空数组
    }

    if (savedCurrentAPIKey) {
      try {
        currentAPIKey.value = JSON.parse(savedCurrentAPIKey)
      } catch (error) {
        console.error('Failed to parse saved current API key:', error)
        currentAPIKey.value = null
      }
    }
  }

  // 保存状态到localStorage
  const saveToStorage = () => {
    try {
      if (loginToken.value) {
        localStorage.setItem('loginToken', loginToken.value)
      } else {
        localStorage.removeItem('loginToken')
      }

      if (accessToken.value) {
        localStorage.setItem('accessToken', accessToken.value)
      } else {
        localStorage.removeItem('accessToken')
      }

      if (user.value) {
        localStorage.setItem('user', JSON.stringify(user.value))
      } else {
        localStorage.removeItem('user')
      }

      const keys: APIKey[] = Array.isArray(apiKeys.value) ? apiKeys.value : []
      localStorage.setItem('apiKeys', JSON.stringify(keys))

      if (currentAPIKey.value) {
        localStorage.setItem('currentAPIKey', JSON.stringify(currentAPIKey.value))
      } else {
        localStorage.removeItem('currentAPIKey')
      }
    } catch (err) {
      console.warn('saveToStorage failed:', err)
    }
  }

  // 清除localStorage
  const clearStorage = () => {
    localStorage.removeItem('loginToken')
    localStorage.removeItem('accessToken')
    localStorage.removeItem('user')
    localStorage.removeItem('apiKeys')
    localStorage.removeItem('currentAPIKey')
  }

  // 用户注册
  const registerUser = async (data: RegisterRequest) => {
    try {
      const response = await register(data)
      return response
    } catch (error) {
      console.error('Registration failed:', error)
      return { code: 500, message: 'Registration failed', data: null, timestamp: Date.now() }
    }
  }

  // 用户登录
  const loginUser = async (data: LoginRequest) => {
    try {
      const response = await login(data)
      
      if (response && response.code === 200 && response.data) {
        // 根据实际API响应结构调整
        user.value = (response.data as any).user || {
          id: (response.data as any).user_id,
          username: (response.data as any).username,
          email: (response.data as any).email,
          is_active: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        }
        loginToken.value = (response.data as any).access_token || (response.data as any).token
        isAuthenticated.value = true
        saveToStorage()
      }
      
      return response
    } catch (error) {
      console.error('Login failed:', error)
      return { code: 500, message: 'Login failed', data: null, timestamp: Date.now() }
    }
  }

  // 获取用户资料
  const fetchUserProfile = async () => {
    if (!loginToken.value) return { code: 401, message: 'Not authenticated', data: null, timestamp: Date.now() }
    
    try {
      const response = await getUserProfile({ token: loginToken.value })
      if (response && response.code === 200 && response.data) {
        user.value = response.data
        saveToStorage()
      }
      return response
    } catch (error) {
      console.error('Failed to fetch user profile:', error)
      return { code: 500, message: 'Failed to fetch user profile', data: null, timestamp: Date.now() }
    }
  }

  // 获取API密钥列表
  const fetchAPIKeys = async () => {
    if (!loginToken.value) return { code: 401, message: 'Not authenticated', data: null, timestamp: Date.now() }
    
    try {
      const response = await getAPIKeys({ token: loginToken.value })
      if (response && response.code === 200) {
        apiKeys.value = Array.isArray(response.data) ? response.data : []
        saveToStorage()
      }
      return response
    } catch (error) {
      console.error('Failed to fetch API keys:', error)
      return { code: 500, message: 'Failed to fetch API keys', data: null, timestamp: Date.now() }
    }
  }

  // 创建API密钥
  const createNewAPIKey = async (data: CreateAPIKeyRequest) => {
    if (!loginToken.value) return { code: 401, message: 'Not authenticated', data: null, timestamp: Date.now() }
    
    try {
      const response = await createAPIKey({ token: loginToken.value, ...data })
      if (response && response.code === 200) {
        // 重新获取API密钥列表
        await fetchAPIKeys()
      }
      return response
    } catch (error) {
      console.error('Failed to create API key:', error)
      return { code: 500, message: 'Failed to create API key', data: null, timestamp: Date.now() }
    }
  }

  // 修改密码
  const changeUserPassword = async (data: { current_password: string; new_password: string }) => {
    if (!loginToken.value) return { code: 401, message: 'Not authenticated', data: null, timestamp: Date.now() }
    
    try {
      const response = await changePassword({ token: loginToken.value, ...data })
      return response
    } catch (error) {
      console.error('Failed to change password:', error)
      return { code: 500, message: 'Failed to change password', data: null, timestamp: Date.now() }
    }
  }

  // 更新API密钥
  const updateExistingAPIKey = async (keyId: number, data: Partial<{ name: string; rate_limit: number; expires_at: string; is_active: boolean }>) => {
    if (!loginToken.value) return { code: 401, message: 'Not authenticated', data: null, timestamp: Date.now() }
    
    try {
      const response = await updateAPIKey({ token: loginToken.value, keyId, updateData: data })
      if (response && response.code === 200) {
        // 重新获取API密钥列表
        await fetchAPIKeys()
      }
      return response
    } catch (error) {
      console.error('Failed to update API key:', error)
      return { code: 500, message: 'Failed to update API key', data: null, timestamp: Date.now() }
    }
  }

  // 删除API密钥
  const deleteExistingAPIKey = async (keyId: number) => {
    if (!loginToken.value) return { code: 401, message: 'Not authenticated', data: null, timestamp: Date.now() }
    
    try {
      const response = await deleteAPIKey({ token: loginToken.value, keyId })
      if (response && response.code === 200) {
        // 重新获取API密钥列表
        await fetchAPIKeys()
      }
      return response
    } catch (error) {
      console.error('Failed to delete API key:', error)
      return { code: 500, message: 'Failed to delete API key', data: null, timestamp: Date.now() }
    }
  }

  // 创建用户地址
  const createNewUserAddress = async (data: { address: string; label: string; type: string }) => {
    if (!loginToken.value) return { code: 401, message: 'Not authenticated', data: null, timestamp: Date.now() }
    
    try {
      const response = await createUserAddress({ token: loginToken.value, ...data })
      return response
    } catch (error) {
      console.error('Failed to create user address:', error)
      return { code: 500, message: 'Failed to create user address', data: null, timestamp: Date.now() }
    }
  }

  // 获取用户地址列表
  const fetchUserAddresses = async () => {
    if (!loginToken.value) return { code: 401, message: 'Not authenticated', data: null, timestamp: Date.now() }
    
    try {
      const response = await getUserAddresses({ token: loginToken.value })
      return response
    } catch (error) {
      console.error('Failed to get user addresses:', error)
      return { code: 500, message: 'Failed to get user addresses', data: null, timestamp: Date.now() }
    }
  }

  // 更新用户地址
  const updateExistingUserAddress = async (addressId: number, data: Partial<{ label: string; type: string; is_active: boolean }>) => {
    if (!loginToken.value) return { code: 401, message: 'Not authenticated', data: null, timestamp: Date.now() }
    
    try {
      const response = await updateUserAddress({ token: loginToken.value, addressId, updateData: data })
      return response
    } catch (error) {
      console.error('Failed to update user address:', error)
      return { code: 500, message: 'Failed to update user address', data: null, timestamp: Date.now() }
    }
  }

  // 删除用户地址
  const deleteExistingUserAddress = async (addressId: number) => {
    if (!loginToken.value) return { code: 401, message: 'Not authenticated', data: null, timestamp: Date.now() }
    
    try {
      const response = await deleteUserAddress({ token: loginToken.value, addressId })
      return response
    } catch (error) {
      console.error('Failed to delete user address:', error)
      return { code: 500, message: 'Failed to delete user address', data: null, timestamp: Date.now() }
    }
  }

  // 获取权限类型列表
  const getPermissionTypes = async () => {
    try {
      // This function is not directly available in the new imports,
      // as the permission types are part of the user profile or fetched via getUserProfile.
      // For now, we'll return a placeholder or remove if not needed.
      // Assuming permission types are part of user profile or fetched via getUserProfile.
      // If they need to be fetched separately, this function needs to be re-added or refactored.
      // For now, returning an empty array as a placeholder.
      return { success: true, data: [] }; // Placeholder
    } catch (error) {
      console.error('Failed to get permission types:', error)
      throw error
    }
  }

  // 获取访问令牌
  const getAccessToken = async (apiKey: string, secretKey: string) => {
    try {
      // This function is not directly available in the new imports.
      // It's assumed to be handled by the backend or a separate auth mechanism.
      // For now, returning a placeholder response.
      return { success: true, data: { access_token: 'dummy_token' } }; // Placeholder
    } catch (error) {
      console.error('Failed to get access token:', error)
      throw error
    }
  }

  // 刷新访问令牌
  const refreshAccessToken = async () => {
    if (!loginToken.value) return false
    
    try {
      const response = await refreshToken({ loginToken: loginToken.value })
      if (response.code === 200) {
        loginToken.value = response.data.access_token
        saveToStorage()
        return true
      }
      return false
    } catch (error) {
      console.error('Failed to refresh access token:', error)
      return false
    }
  }

  // 检查并刷新令牌
  const checkAndRefreshToken = async () => {
    if (isTokenExpired.value) {
      try {
        const response = await refreshToken({ loginToken: loginToken.value! })
        if (response.code === 200) {
          loginToken.value = response.data.access_token
          saveToStorage()
          return true
        }
        return false
      } catch (error) {
        console.error('Failed to refresh token:', error)
        return false
      }
    }
    return true
  }

  // 设置当前API密钥
  const setCurrentAPIKey = (apiKey: APIKey) => {
    currentAPIKey.value = apiKey
    saveToStorage()
  }

  // 用户登出
  const logout = () => {
    isAuthenticated.value = false
    user.value = null
    loginToken.value = null
    accessToken.value = null
    apiKeys.value = []
    currentAPIKey.value = null
    
    clearStorage()
  }

  // 智能API调用 - 根据认证状态选择API
  const smartAPI = {
    // 智能获取区块列表
    async getBlocks(chain?: string, page: number = 1, pageSize: number = 10) {
      // 如果已认证且有访问令牌，使用认证API
      if (hasValidTokens.value && accessToken.value) {
        try {
          // This function is not directly available in the new imports.
          // It's assumed to be handled by the backend or a separate auth mechanism.
          // For now, returning a placeholder response.
          return { success: true, data: [] }; // Placeholder
        } catch (error: unknown) {
          // 如果是限流错误，提示用户登录解锁更多功能
          if (error instanceof Error && (error.message?.includes('请求过于频繁') || error.message?.includes('限流'))) {
            throw new Error('限流触发，请登录解锁更多功能！')
          }
          throw error
        }
      }
      
      // 否则使用公开API
      // This function is not directly available in the new imports.
      // It's assumed to be handled by the backend or a separate auth mechanism.
      // For now, returning a placeholder response.
      return { success: true, data: [] }; // Placeholder
    },

    // 智能获取交易列表
    async getTransactions(chain?: string, page: number = 1, pageSize: number = 50) {
      // 如果已认证且有访问令牌，使用认证API
      if (hasValidTokens.value && accessToken.value) {
        try {
          // This function is not directly available in the new imports.
          // It's assumed to be handled by the backend or a separate auth mechanism.
          // For now, returning a placeholder response.
          return { success: true, data: [] }; // Placeholder
        } catch (error: unknown) {
          // 如果是限流错误，提示用户登录解锁更多功能
          if (error instanceof Error && (error.message?.includes('请求过于频繁') || error.message?.includes('限流'))) {
            throw new Error('限流触发，请登录解锁更多功能！')
          }
          throw error
        }
      }
      
      // 否则使用公开API
      // This function is not directly available in the new imports.
      // It's assumed to be handled by the backend or a separate auth mechanism.
      // For now, returning a placeholder response.
      return { success: true, data: [] }; // Placeholder
    }
  }

  return {
    // 状态
    isAuthenticated,
    user,
    loginToken,
    accessToken,
    apiKeys,
    currentAPIKey,
    
    // 计算属性
    hasValidTokens,
    isTokenExpired,
    
    // 方法
    initialize,
    registerUser,
    loginUser,
    logout,
    fetchUserProfile,
    fetchAPIKeys,
    createNewAPIKey,
    changeUserPassword,
    updateExistingAPIKey,
    deleteExistingAPIKey,
    createNewUserAddress,
    fetchUserAddresses,
    updateExistingUserAddress,
    deleteExistingUserAddress,
    getPermissionTypes,
    getAccessToken,
    refreshAccessToken,
    setCurrentAPIKey,
    checkAndRefreshToken,
    
    // 智能API
    smartAPI,
  }
})
