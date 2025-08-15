// 认证相关类型定义

// 用户注册请求
export interface RegisterRequest {
  username: string
  email: string
  password: string
}

// 用户登录请求
export interface LoginRequest {
  username: string
  password: string
}

// 登录响应
export interface LoginResponse {
  user_id: number
  username: string
  email: string
  token: string
  expires_at: number
}

// 用户资料
export interface UserProfile {
  id: number
  username: string
  email: string
  is_active: boolean
  last_login?: string
  created_at: string
  updated_at: string
}

// API密钥
export interface APIKey {
  id: number
  name: string
  api_key: string
  secret_key?: string
  permissions: string[] // 权限范围
  is_active: boolean
  usage_count: number
  last_used_at?: string
  expires_at?: string
  created_at: string
  updated_at: string
}

// 创建API密钥请求
export interface CreateAPIKeyRequest {
  name: string
  permissions: string[] // 权限范围
  expires_at?: string // 过期时间
}

// 创建API密钥响应
export interface CreateAPIKeyResponse {
  id: number
  name: string
  api_key: string
  secret_key: string
  rate_limit: number
  expires_at?: string
  created_at: string
}

// 获取访问令牌请求
export interface GetAccessTokenRequest {
  api_key: string
  secret_key: string
}

// 获取访问令牌响应
export interface GetAccessTokenResponse {
  access_token: string
  token_type: string
  expires_in: number
  expires_at: number
}

// 认证状态
export interface AuthState {
  isAuthenticated: boolean
  user: UserProfile | null
  loginToken: string | null
  accessToken: string | null
  apiKeys: APIKey[]
  currentAPIKey: APIKey | null
}

// 限流错误响应
export interface RateLimitError {
  success: false
  error: string
  message: string
  retry_after: string
}

// 权限配置项
export interface PermissionConfig {
  id: number
  group: string
  no: number
  config_type: number
  config_name: string
  config_key: string
  config_value: string
  description: string
  created_at: string
  updated_at: string
  deleted_at?: string
}
