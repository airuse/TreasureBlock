import { useAuthStore } from '@/stores/auth'
import * as authApi from './index'
import * as noAuthApi from './no-auth'

/**
 * API选择器 - 根据用户登录状态自动选择API模块
 * 
 * 使用方式：
 * const api = useApiSelector()
 * const blocks = await api.blocks.getBlocks({ page: 1, page_size: 25, chain: 'eth' })
 */
export function useApiSelector() {
  const authStore = useAuthStore()
  
  // 根据登录状态返回对应的API模块
  if (authStore.isAuthenticated) {
    console.log('🔐 用户已登录，使用认证API (/api/v1)')
    return authApi
  } else {
    console.log('👤 用户未登录，使用游客API (/api/no-auth)')
    return noAuthApi
  }
}

/**
 * 获取当前应该使用的API基础URL
 */
export function getCurrentApiBaseUrl(): string {
  const authStore = useAuthStore()
  return authStore.isAuthenticated ? '/api/v1' : '/api/no-auth'
}

/**
 * 检查当前用户是否有权限访问某个功能
 */
export function hasPermission(permission: string): boolean {
  const authStore = useAuthStore()
  return authStore.isAuthenticated
}

/**
 * 获取当前用户的API限制信息
 */
export function getCurrentApiLimits() {
  const authStore = useAuthStore()
  
  if (authStore.isAuthenticated) {
    return {
      blocks: '无限制',
      transactions: '无限制',
      search: '无限制',
      contracts: '无限制'
    }
  } else {
    return {
      blocks: '最多100个',
      transactions: '最多1000条',
      search: '最多20个结果',
      contracts: '最多100个'
    }
  }
}

export default useApiSelector
