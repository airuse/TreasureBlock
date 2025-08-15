import axios from 'axios'
import type { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from 'axios'

// 创建axios实例
const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'https://localhost:8443',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // 添加认证token
    const token = localStorage.getItem('access_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    
    console.log(`🌐 API请求: ${config.method?.toUpperCase()} ${config.url}`)
    return config
  },
  (error: any) => {
    console.error('❌ 请求拦截器错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    console.log(`✅ API响应: ${response.status} ${response.config.url}`)
    return response.data
  },
  (error: any) => {
    console.error('❌ API响应错误:', error)
    
    // 处理特定错误状态
    if (error.response?.status === 429) {
      console.warn('⚠️ 请求频率限制')
    } else if (error.response?.status === 401) {
      console.warn('⚠️ 认证失败，请重新登录')
      // 可以在这里处理token过期逻辑
    }
    
    return Promise.reject(error)
  }
)

export default request
