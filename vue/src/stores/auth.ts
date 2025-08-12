import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { 
  authAPI, 
  publicAPI, 
  authenticatedAPI 
} from '@/api/auth'
import type { 
  AuthState, 
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
  const register = async (data: RegisterRequest) => {
    try {
      const response = await authAPI.register(data)
      return response
    } catch (error) {
      console.error('Registration failed:', error)
      throw error
    }
  }

  // 用户登录
  const login = async (data: LoginRequest) => {
    try {
      const response = await authAPI.login(data)
      
      if (response.success) {
        // 保存登录信息
        loginToken.value = response.data.token
        user.value = {
          id: response.data.user_id,
          username: response.data.username,
          email: response.data.email,
          is_active: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        }
        isAuthenticated.value = true

        // 获取用户完整资料
        await fetchUserProfile()
        
        // 获取API密钥列表
        await fetchAPIKeys()
        
        // 保存到localStorage
        saveToStorage()
      }
      
      return response
    } catch (error) {
      console.error('Login failed:', error)
      throw error
    }
  }

  // 获取用户资料
  const fetchUserProfile = async () => {
    if (!loginToken.value) return
    
    try {
      const response = await authAPI.getUserProfile(loginToken.value)
      if (response.success && response.data) {
        user.value = response.data
        saveToStorage()
      }
    } catch (error) {
      console.error('Failed to fetch user profile:', error)
    }
  }

  // 获取API密钥列表
  const fetchAPIKeys = async () => {
    if (!loginToken.value) return
    
    try {
      const response = await authAPI.getAPIKeys(loginToken.value)
      if (response.success) {
        apiKeys.value = Array.isArray(response.data) ? response.data : []
        saveToStorage()
      }
    } catch (error) {
      console.error('Failed to fetch API keys:', error)
      apiKeys.value = [] // 出错时设置为空数组
    }
  }

  // 创建API密钥
  const createAPIKey = async (data: CreateAPIKeyRequest) => {
    if (!loginToken.value) {
      throw new Error('未登录')
    }
    
    try {
      const response = await authAPI.createAPIKey(loginToken.value, data)
      if (response.success) {
        // 重新获取API密钥列表
        await fetchAPIKeys()
      }
      return response
    } catch (error) {
      console.error('Failed to create API key:', error)
      throw error
    }
  }

  // 修改密码
  const changePassword = async (data: { current_password: string; new_password: string }) => {
    if (!loginToken.value) {
      throw new Error('未登录')
    }
    
    try {
      const response = await authAPI.changePassword(loginToken.value, data.current_password, data.new_password)
      return response
    } catch (error) {
      console.error('Failed to change password:', error)
      throw error
    }
  }

  // 更新API密钥
  const updateAPIKey = async (keyId: number, data: Partial<{ name: string; rate_limit: number; expires_at: string; is_active: boolean }>) => {
    if (!loginToken.value) {
      throw new Error('未登录')
    }
    
    try {
      const response = await authAPI.updateAPIKey(loginToken.value, keyId, data)
      if (response.success) {
        // 重新获取API密钥列表
        await fetchAPIKeys()
      }
      return response
    } catch (error) {
      console.error('Failed to update API key:', error)
      throw error
    }
  }

  // 删除API密钥
  const deleteAPIKey = async (keyId: number) => {
    if (!loginToken.value) {
      throw new Error('未登录')
    }
    
    try {
      const response = await authAPI.deleteAPIKey(loginToken.value, keyId)
      if (response.success) {
        // 重新获取API密钥列表
        await fetchAPIKeys()
      }
      return response
    } catch (error) {
      console.error('Failed to delete API key:', error)
      throw error
    }
  }

  // 用户地址管理
  const createUserAddress = async (data: { address: string; label: string; type: string }) => {
    if (!loginToken.value) {
      throw new Error('未登录')
    }
    
    try {
      const response = await authAPI.createUserAddress(loginToken.value, data)
      return response
    } catch (error) {
      console.error('Failed to create user address:', error)
      throw error
    }
  }

  const getUserAddresses = async () => {
    if (!loginToken.value) {
      throw new Error('未登录')
    }
    
    try {
      const response = await authAPI.getUserAddresses(loginToken.value)
      return response
    } catch (error) {
      console.error('Failed to get user addresses:', error)
      throw error
    }
  }

  const updateUserAddress = async (addressId: number, data: Partial<{ label: string; type: string; is_active: boolean }>) => {
    if (!loginToken.value) {
      throw new Error('未登录')
    }
    
    try {
      const response = await authAPI.updateUserAddress(loginToken.value, addressId, data)
      return response
    } catch (error) {
      console.error('Failed to update user address:', error)
      throw error
    }
  }

  const deleteUserAddress = async (addressId: number) => {
    if (!loginToken.value) {
      throw new Error('未登录')
    }
    
    try {
      const response = await authAPI.deleteUserAddress(loginToken.value, addressId)
      return response
    } catch (error) {
      console.error('Failed to delete user address:', error)
      throw error
    }
  }

  // 获取权限类型列表
  const getPermissionTypes = async () => {
    try {
      const response = await authAPI.getPermissionTypes()
      return response
    } catch (error) {
      console.error('Failed to get permission types:', error)
      throw error
    }
  }

  // 获取访问令牌
  const getAccessToken = async (apiKey: string, secretKey: string) => {
    try {
      const data: GetAccessTokenRequest = { api_key: apiKey, secret_key: secretKey }
      const response = await authAPI.getAccessToken(data)
      
      if (response.success) {
        accessToken.value = response.data.access_token
        saveToStorage()
      }
      
      return response
    } catch (error) {
      console.error('Failed to get access token:', error)
      throw error
    }
  }

  // 刷新访问令牌
  const refreshAccessToken = async () => {
    if (!currentAPIKey.value || !currentAPIKey.value.secret_key) return false
    
    try {
      const response = await getAccessToken(
        currentAPIKey.value.api_key, 
        currentAPIKey.value.secret_key
      )
      return response.success
    } catch (error) {
      console.error('Failed to refresh access token:', error)
      return false
    }
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

  // 检查令牌是否过期并自动刷新
  const checkAndRefreshTokens = async () => {
    if (isTokenExpired.value) {
      try {
        const response = await authAPI.refreshToken(loginToken.value!)
        if (response.success) {
          loginToken.value = response.data.token
          saveToStorage()
          return true
        }
      } catch (error) {
        console.error('Token refresh failed:', error)
        logout()
        return false
      }
    }
    return true
  }

  // 智能API调用 - 根据认证状态选择API
  const smartAPI = {
    // 智能获取区块列表
    async getBlocks(chain?: string, page: number = 1, pageSize: number = 10) {
      // 如果已认证且有访问令牌，使用认证API
      if (hasValidTokens.value && accessToken.value) {
        try {
          return await authenticatedAPI.getBlocks(accessToken.value, chain, page, pageSize)
        } catch (error: any) {
          // 如果是限流错误，提示用户登录解锁更多功能
          if (error.message?.includes('请求过于频繁') || error.message?.includes('限流')) {
            throw new Error('限流触发，请登录解锁更多功能！')
          }
          throw error
        }
      }
      
      // 否则使用公开API
      return await publicAPI.getBlocks(chain, page, pageSize)
    },

    // 智能获取交易列表
    async getTransactions(chain?: string, page: number = 1, pageSize: number = 50) {
      // 如果已认证且有访问令牌，使用认证API
      if (hasValidTokens.value && accessToken.value) {
        try {
          return await authenticatedAPI.getTransactions(accessToken.value, chain, page, pageSize)
        } catch (error: any) {
          // 如果是限流错误，提示用户登录解锁更多功能
          if (error.message?.includes('请求过于频繁') || error.message?.includes('限流')) {
            throw new Error('限流触发，请登录解锁更多功能！')
          }
          throw error
        }
      }
      
      // 否则使用公开API
      return await publicAPI.getTransactions(chain, page, pageSize)
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
    register,
    login,
    logout,
    fetchUserProfile,
    fetchAPIKeys,
    createAPIKey,
    changePassword,
    updateAPIKey,
    deleteAPIKey,
    createUserAddress,
    getUserAddresses,
    updateUserAddress,
    deleteUserAddress,
    getPermissionTypes,
    getAccessToken,
    refreshAccessToken,
    setCurrentAPIKey,
    checkAndRefreshTokens,
    
    // 智能API
    smartAPI,
  }
})
