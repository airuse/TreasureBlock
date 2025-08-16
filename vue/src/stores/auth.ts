import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { 
  UserProfile, 
  APIKey, 
  LoginRequest, 
  RegisterRequest,
  CreateAPIKeyRequest,
  GetAccessTokenRequest
} from '../types/auth'

export const useAuthStore = defineStore('auth', () => {
  // 状态 - 只保留需要持久化的数据
  const isAuthenticated = ref(false)
  const user = ref<UserProfile | null>(null)
  const loginToken = ref<string | null>(null)
  const accessToken = ref<string | null>(null)
  const permissionTypes = ref<Array<{ config_value: string; config_name: string }>>([])

  // 计算属性
  const hasValidTokens = computed(() => {
    return isAuthenticated.value && loginToken.value && accessToken.value
  })

  const isTokenExpired = computed(() => {
    if (!loginToken.value) return true
    return false
  })

  // 初始化 - 从localStorage恢复状态
  const initialize = () => {
    const savedLoginToken = localStorage.getItem('loginToken')
    const savedAccessToken = localStorage.getItem('accessToken')
    const savedUser = localStorage.getItem('user')
    const savedPermissionTypes = localStorage.getItem('permissionTypes')

    if (savedLoginToken && savedUser) {
      loginToken.value = savedLoginToken
      user.value = JSON.parse(savedUser)
      isAuthenticated.value = true
    }

    if (savedAccessToken) {
      accessToken.value = savedAccessToken
    }

    if (savedPermissionTypes) {
      try {
        permissionTypes.value = JSON.parse(savedPermissionTypes) || []
      } catch (error) {
        console.error('Failed to parse saved permission types:', error)
        permissionTypes.value = []
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

      localStorage.setItem('permissionTypes', JSON.stringify(permissionTypes.value))
    } catch (err) {
      console.warn('saveToStorage failed:', err)
    }
  }

  // 清除localStorage
  const clearStorage = () => {
    localStorage.removeItem('loginToken')
    localStorage.removeItem('accessToken')
    localStorage.removeItem('user')
    localStorage.removeItem('permissionTypes')
  }

  // 设置用户认证状态 - 用于登录成功后
  const setAuthState = (userData: UserProfile, token: string) => {
    user.value = userData
    loginToken.value = token
    isAuthenticated.value = true
    saveToStorage()
  }

  // 设置权限类型数据 - 用于从API获取后保存到本地
  const setPermissionTypes = (permissions: Array<{ config_value: string; config_name: string }>) => {
    permissionTypes.value = permissions
    saveToStorage()
  }

  // 检查本地是否有权限数据
  const hasLocalPermissionTypes = computed(() => {
    return permissionTypes.value.length > 0
  })

  // 获取本地权限数据
  const getLocalPermissionTypes = () => {
    return permissionTypes.value
  }

  // 用户登出
  const logout = () => {
    isAuthenticated.value = false
    user.value = null
    loginToken.value = null
    accessToken.value = null
    permissionTypes.value = []
    
    clearStorage()
  }

  return {
    // 状态
    isAuthenticated,
    user,
    loginToken,
    accessToken,
    permissionTypes,
    
    // 计算属性
    hasValidTokens,
    isTokenExpired,
    hasLocalPermissionTypes,
    
    // 方法
    initialize,
    setAuthState,
    setPermissionTypes,
    getLocalPermissionTypes,
    logout,
  }
})
