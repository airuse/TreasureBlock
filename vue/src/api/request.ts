import axios from 'axios'
import type { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from 'axios'

// 创建axios实例
const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // 添加认证token - 优先使用loginToken，如果没有则使用access_token
    const loginToken = localStorage.getItem('loginToken')
    const accessToken = localStorage.getItem('access_token')
    const token = loginToken || accessToken
    
    if (token && config.headers) {
      // 根据后端要求设置token格式
      if (loginToken) {
        // 如果使用loginToken，可能需要特殊格式
        config.headers.Authorization = `Bearer ${token}`
        // 或者根据后端要求设置其他header
        // config.headers['X-Auth-Token'] = token
      } else {
        config.headers.Authorization = `Bearer ${token}`
      }
    }
    
    // console.log(`🌐 API请求: ${config.method?.toUpperCase()} ${config.url}`)
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
    // console.log(`✅ API响应: ${response.status} ${response.config.url}`)
    return response.data
  },
  (error: any) => {
    console.error('❌ API响应错误:', error)
    
    // 处理特定错误状态
    if (error.response?.status === 429) {
      console.warn('⚠️ 请求频率限制')
      // 统一处理限流错误，显示用户友好的提示
      showRateLimitError()
    } else if (error.response?.status === 401) {
      console.warn('⚠️ 认证失败，请重新登录')
      // 清除本地存储的token
      localStorage.removeItem('loginToken')
      localStorage.removeItem('access_token')
      
      // 触发登录模态框显示
      showLoginModal()
    }
    
    return Promise.reject(error)
  }
)

// 显示限流错误的函数
function showRateLimitError() {
  // 动态导入toast组件，避免循环依赖
  import('@/composables/useToast').then(({ showError }) => {
    showError('请求过于频繁，请稍后再试')
  }).catch(() => {
    // 如果导入失败，使用console.warn作为备选方案
    console.warn('⚠️ 请求过于频繁，请稍后再试')
  })
}

// 显示登录模态框的函数
function showLoginModal() {
  // 通过自定义事件触发登录模态框显示
  window.dispatchEvent(new CustomEvent('show-login-modal'))
}

export default request
