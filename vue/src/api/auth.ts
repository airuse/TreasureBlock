import { request, authenticatedRequest } from './client'
import type { Block, Transaction, UserAddress } from '@/types'

// 认证API（需要LoginToken）
export const authAPI = {
  // 用户注册
  async register(data: { username: string; email: string; password: string }): Promise<{ success: boolean; message: string; data: { user_id: number; username: string; email: string; token: string; expires_at: number } }> {
    return request('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  // 用户登录
  async login(data: { username: string; password: string }): Promise<{ success: boolean; message: string; data: { user_id: number; username: string; email: string; token: string; expires_at: number } }> {
    return request('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  // 获取用户资料
  async getUserProfile(token: string): Promise<{ success: boolean; message: string; data: { id: number; username: string; email: string; is_active: boolean; last_login?: string; created_at: string; updated_at: string } }> {
    return authenticatedRequest('/api/auth/profile', token)
  },

  // 修改密码
  async changePassword(token: string, data: { current_password: string; new_password: string }): Promise<{ success: boolean; message: string }> {
    return authenticatedRequest('/api/auth/change-password', token, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  },

  // 刷新访问令牌
  async refreshToken(loginToken: string): Promise<{ success: boolean; message: string; data: { token: string } }> {
    return authenticatedRequest('/api/auth/refresh', loginToken)
  },

  // API密钥管理
  async createAPIKey(token: string, data: { name: string; permissions: string[]; expires_at?: string }): Promise<{ success: boolean; message: string; data: { id: number; name: string; api_key: string; secret_key: string; rate_limit: number; expires_at?: string; created_at: string } }> {
    return authenticatedRequest('/api/user/api-keys', token, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  async getAPIKeys(token: string): Promise<{ success: boolean; message: string; data: Array<{ id: number; name: string; api_key: string; secret_key?: string; permissions: string[]; is_active: boolean; usage_count: number; last_used_at?: string; expires_at?: string; created_at: string; updated_at: string }> }> {
    return authenticatedRequest('/api/user/api-keys', token)
  },

  async updateAPIKey(token: string, keyId: number, data: Partial<{ name: string; permissions: string[]; is_active: boolean; expires_at: string }>): Promise<{ success: boolean; message: string; data: { id: number; name: string; api_key: string; secret_key?: string; permissions: string[]; is_active: boolean; usage_count: number; last_used_at?: string; expires_at?: string; created_at: string; updated_at: string } }> {
    return authenticatedRequest(`/api/user/api-keys/${keyId}`, token, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  },

  async deleteAPIKey(token: string, keyId: number): Promise<{ success: boolean; message: string }> {
    return authenticatedRequest(`/api/user/api-keys/${keyId}`, token, {
      method: 'DELETE',
    })
  },

  // 获取访问令牌
  async getAccessToken(apiKey: string, secretKey: string): Promise<{ success: boolean; message: string; data: { access_token: string; token_type: string; expires_in: number; expires_at: number } }> {
    return request('/api/auth/access-token', {
      method: 'POST',
      body: JSON.stringify({ api_key: apiKey, secret_key: secretKey }),
    })
  },

  // 用户地址管理
  async createUserAddress(token: string, data: { address: string; label: string; type: string }): Promise<{ success: boolean; message: string; data: UserAddress }> {
    return authenticatedRequest('/api/user/addresses', token, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  async getUserAddresses(token: string): Promise<{ success: boolean; message: string; data: UserAddress[] }> {
    return authenticatedRequest('/api/user/addresses', token)
  },

  async updateUserAddress(token: string, addressId: number, data: Partial<{ label: string; type: string; is_active: boolean }>): Promise<{ success: boolean; message: string; data: UserAddress }> {
    return authenticatedRequest(`/api/user/addresses/${addressId}`, token, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  },

  async deleteUserAddress(token: string, addressId: number): Promise<{ success: boolean; message: string }> {
    return authenticatedRequest(`/api/user/addresses/${addressId}`, token, {
      method: 'DELETE',
    })
  },

  // 获取权限类型列表
  async getPermissionTypes(): Promise<{ success: boolean; message: string; data: Array<{ value: string; description: string }> }> {
    return request('/api/base-configs/group/api_permissions')
  },
}

// 公开API（无需认证）
export const publicAPI = {
  // 获取区块列表（限制版本）
  async getBlocks(chain?: string, page: number = 1, pageSize: number = 10): Promise<{ success: boolean; message: string; data: Block[]; pagination: { page: number; page_size: number; total: number } }> {
    const params = new URLSearchParams()
    if (chain) params.append('chain', chain)
    params.append('page', page.toString())
    params.append('page_size', pageSize.toString())
    
    return request(`/api/no-auth/blocks?${params.toString()}`)
  },

  // 获取交易列表（限制版本）
  async getTransactions(chain?: string, page: number = 1, pageSize: number = 50): Promise<{ success: boolean; message: string; data: Transaction[]; pagination: { page: number; page_size: number; total: number } }> {
    const params = new URLSearchParams()
    if (chain) params.append('chain', chain)
    params.append('page', page.toString())
    params.append('page_size', pageSize.toString())
    
    return request(`/api/no-auth/transactions?${params.toString()}`)
  },
}

// 认证API（需要AccessToken）
export const authenticatedAPI = {
  // 获取区块列表（完整版本）
  async getBlocks(accessToken: string, chain?: string, page: number = 1, pageSize: number = 100): Promise<{ success: boolean; message: string; data: Block[]; pagination: { page: number; page_size: number; total: number } }> {
    const params = new URLSearchParams()
    if (chain) params.append('chain', chain)
    params.append('page', page.toString())
    params.append('page_size', pageSize.toString())
    
    return authenticatedRequest(`/api/v1/blocks?${params.toString()}`, accessToken)
  },

  // 获取交易列表（完整版本）
  async getTransactions(accessToken: string, chain?: string, page: number = 1, pageSize: number = 1000): Promise<{ success: boolean; message: string; data: Transaction[]; pagination: { page: number; page_size: number; total: number } }> {
    const params = new URLSearchParams()
    if (chain) params.append('chain', chain)
    params.append('page', page.toString())
    params.append('page_size', pageSize.toString())
    
    return authenticatedRequest(`/api/v1/transactions?${params.toString()}`, accessToken)
  },

  // 获取区块详情
  async getBlockByHash(accessToken: string, hash: string): Promise<{ success: boolean; message: string; data: Block }> {
    return authenticatedRequest(`/api/v1/blocks/hash/${hash}`, accessToken)
  },

  // 获取交易详情
  async getTransactionByHash(accessToken: string, hash: string): Promise<{ success: boolean; message: string; data: Transaction }> {
    return authenticatedRequest(`/api/v1/transactions/${hash}`, accessToken)
  },

  // 获取地址交易
  async getTransactionsByAddress(accessToken: string, address: string, page: number = 1, pageSize: number = 100): Promise<{ success: boolean; message: string; data: Transaction[]; pagination: { page: number; page_size: number; total: number } }> {
    const params = new URLSearchParams()
    params.append('page', page.toString())
    params.append('page_size', pageSize.toString())
    
    return authenticatedRequest(`/api/v1/addresses/${address}/transactions?${params.toString()}`, accessToken)
  },
}
