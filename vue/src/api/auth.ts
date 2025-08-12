import type { 
  RegisterRequest, 
  LoginRequest, 
  LoginResponse, 
  UserProfile, 
  APIKey, 
  CreateAPIKeyRequest, 
  CreateAPIKeyResponse,
  GetAccessTokenRequest,
  GetAccessTokenResponse
} from '../types/auth'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'https://localhost:8443'

// 通用请求函数
async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`
  
  const config: RequestInit = {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  }

  try {
    const response = await fetch(url, config)
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      // 特判公开API限流
      if (response.status === 429) {
        const msg = (errorData && (errorData.message || errorData.error)) || '请求过于频繁'
        throw new Error(`RATE_LIMITED:${msg}`)
      }
      throw new Error((errorData && (errorData.error || errorData.message)) || `HTTP error! status: ${response.status}`)
    }
    
    return await response.json()
  } catch (error) {
    console.error('API request failed:', error)
    throw error
  }
}

// 带认证的请求函数
async function authenticatedRequest<T>(endpoint: string, token: string, options: RequestInit = {}): Promise<T> {
  return request<T>(endpoint, {
    ...options,
    headers: {
      'Authorization': `Bearer ${token}`,
      ...options.headers,
    },
  })
}

// 认证相关API
export const authAPI = {
  // 用户注册
  async register(data: RegisterRequest): Promise<{ success: boolean; message: string; data: UserProfile }> {
    return request('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  // 用户登录
  async login(data: LoginRequest): Promise<{ success: boolean; message: string; data: LoginResponse }> {
    return request('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  // 刷新令牌
  async refreshToken(token: string): Promise<{ success: boolean; message: string; data: LoginResponse }> {
    return authenticatedRequest('/api/auth/refresh', token, {
      method: 'POST',
    })
  },

  // 获取用户资料
  async getUserProfile(token: string): Promise<{ success: boolean; message: string; data: UserProfile }> {
    return authenticatedRequest('/api/user/profile', token)
  },

  // 修改密码
  async changePassword(token: string, currentPassword: string, newPassword: string): Promise<{ success: boolean; message: string }> {
    return authenticatedRequest('/api/user/change-password', token, {
      method: 'POST',
      body: JSON.stringify({
        current_password: currentPassword,
        new_password: newPassword,
      }),
    })
  },

  // 创建API密钥
  async createAPIKey(token: string, data: CreateAPIKeyRequest): Promise<{ success: boolean; message: string; data: CreateAPIKeyResponse }> {
    return authenticatedRequest('/api/user/api-keys', token, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  // 获取API密钥列表
  async getAPIKeys(token: string): Promise<{ success: boolean; message: string; data: APIKey[] }> {
    return authenticatedRequest('/api/user/api-keys', token)
  },

  // 更新API密钥
  async updateAPIKey(token: string, keyId: number, data: Partial<CreateAPIKeyRequest>): Promise<{ success: boolean; message: string; data: APIKey }> {
    return authenticatedRequest(`/api/user/api-keys/${keyId}`, token, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  },

  // 删除API密钥
  async deleteAPIKey(token: string, keyId: number): Promise<{ success: boolean; message: string }> {
    return authenticatedRequest(`/api/user/api-keys/${keyId}`, token, {
      method: 'DELETE',
    })
  },

  // 获取访问令牌
  async getAccessToken(data: GetAccessTokenRequest): Promise<{ success: boolean; message: string; data: GetAccessTokenResponse }> {
    return request('/api/auth/token', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  // 用户地址管理
  async createUserAddress(token: string, data: { address: string; label: string; type: string }): Promise<{ success: boolean; message: string; data: any }> {
    return authenticatedRequest('/api/user/addresses', token, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  },

  async getUserAddresses(token: string): Promise<{ success: boolean; message: string; data: any[] }> {
    return authenticatedRequest('/api/user/addresses', token)
  },

  async updateUserAddress(token: string, addressId: number, data: Partial<{ label: string; type: string; is_active: boolean }>): Promise<{ success: boolean; message: string; data: any }> {
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
  async getBlocks(chain?: string, page: number = 1, pageSize: number = 10): Promise<any> {
    const params = new URLSearchParams()
    if (chain) params.append('chain', chain)
    params.append('page', page.toString())
    params.append('page_size', pageSize.toString())
    
    return request(`/api/no-auth/blocks?${params.toString()}`)
  },

  // 获取交易列表（限制版本）
  async getTransactions(chain?: string, page: number = 1, pageSize: number = 50): Promise<any> {
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
  async getBlocks(accessToken: string, chain?: string, page: number = 1, pageSize: number = 100): Promise<any> {
    const params = new URLSearchParams()
    if (chain) params.append('chain', chain)
    params.append('page', page.toString())
    params.append('page_size', pageSize.toString())
    
    return authenticatedRequest(`/api/v1/blocks?${params.toString()}`, accessToken)
  },

  // 获取交易列表（完整版本）
  async getTransactions(accessToken: string, chain?: string, page: number = 1, pageSize: number = 1000): Promise<any> {
    const params = new URLSearchParams()
    if (chain) params.append('chain', chain)
    params.append('page', page.toString())
    params.append('page_size', pageSize.toString())
    
    return authenticatedRequest(`/api/v1/transactions?${params.toString()}`, accessToken)
  },

  // 获取区块详情
  async getBlockByHash(accessToken: string, hash: string): Promise<any> {
    return authenticatedRequest(`/api/v1/blocks/hash/${hash}`, accessToken)
  },

  // 获取交易详情
  async getTransactionByHash(accessToken: string, hash: string): Promise<any> {
    return authenticatedRequest(`/api/v1/transactions/${hash}`, accessToken)
  },

  // 获取地址交易
  async getTransactionsByAddress(accessToken: string, address: string, page: number = 1, pageSize: number = 100): Promise<any> {
    const params = new URLSearchParams()
    params.append('page', page.toString())
    params.append('page_size', pageSize.toString())
    
    return authenticatedRequest(`/api/v1/addresses/${address}/transactions?${params.toString()}`, accessToken)
  },
}
