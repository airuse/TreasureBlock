import { ref } from 'vue'

export interface Toast {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  message: string
  duration?: number
}

const toasts = ref<Toast[]>([])

// 添加Toast
export const addToast = (toast: Omit<Toast, 'id'>) => {
  const id = Date.now().toString()
  const newToast: Toast = {
    id,
    ...toast,
    duration: toast.duration ?? 5000
  }
  
  toasts.value.push(newToast)
  
  // 自动移除
  if (newToast.duration && newToast.duration > 0) {
    setTimeout(() => {
      removeToast(id)
    }, newToast.duration)
  }
  
  return id
}

// 移除Toast
export const removeToast = (id: string) => {
  const index = toasts.value.findIndex(toast => toast.id === id)
  if (index > -1) {
    toasts.value.splice(index, 1)
  }
}

// 清空所有Toast
export const clearToasts = () => {
  toasts.value = []
}

// 便捷方法
export const showSuccess = (message: string, duration?: number) => {
  return addToast({ type: 'success', message, duration })
}

export const showError = (message: string, duration?: number) => {
  return addToast({ type: 'error', message, duration })
}

export const showWarning = (message: string, duration?: number) => {
  return addToast({ type: 'warning', message, duration })
}

export const showInfo = (message: string, duration?: number) => {
  return addToast({ type: 'info', message, duration })
}

// 导出toasts响应式引用
export { toasts }

// 默认导出
export default {
  toasts,
  addToast,
  removeToast,
  clearToasts,
  showSuccess,
  showError,
  showWarning,
  showInfo
}
