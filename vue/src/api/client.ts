const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'https://localhost:8443'

// 通用请求函数
export async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
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
export async function authenticatedRequest<T>(endpoint: string, token: string, options: RequestInit = {}): Promise<T> {
  return request<T>(endpoint, {
    ...options,
    headers: {
      'Authorization': `Bearer ${token}`,
      ...options.headers,
    },
  })
}
