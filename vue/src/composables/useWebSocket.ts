import { ref, watch, onMounted, onUnmounted, computed } from 'vue'
import {
  WebSocketStatus,
  createWebSocketManager,
  getWebSocketManager,
  setupVisibilityHandler
} from '@/utils/websocket'
import type { 
  WebSocketOptions, 
  WebSocketMessage
} from '@/utils/websocket'
import type WebSocketManager from '@/utils/websocket'

// WebSocket组合式函数
export function useWebSocket(options?: Partial<WebSocketOptions>) {
  const manager = ref<WebSocketManager | null>(null)
  const status = ref<WebSocketStatus>(WebSocketStatus.DISCONNECTED)
  const isConnected = ref(false)
  const isConnecting = ref(false)
  const isReconnecting = ref(false)
  const reconnectAttempts = ref(0)

  // 默认配置
  const defaultOptions: WebSocketOptions = {
    url: import.meta.env.VITE_WS_BASE_URL || 'wss://localhost:8443/ws', // 使用HTTP协议避免TLS问题
    autoReconnect: true,
    reconnectInterval: 3000,
    maxReconnectAttempts: 5,
    heartbeatInterval: 30000
  }

  const finalOptions = { ...defaultOptions, ...options }

  // 初始化WebSocket管理器
  const initWebSocket = async () => {
    try {
      // 获取全局WebSocket管理器
      let wsManager = getWebSocketManager()
      
      if (!wsManager) {
        // 如果没有全局管理器，创建一个（这种情况不应该发生，因为main.ts中已经创建）
        console.warn('No global WebSocket manager found, creating one...')
        wsManager = createWebSocketManager(finalOptions)
        setupVisibilityHandler(wsManager)
      }

      manager.value = wsManager

      // 监听状态变化
      watch(wsManager.status, (newStatus) => {
        status.value = newStatus
        isConnected.value = wsManager.isConnected.value
        isConnecting.value = wsManager.isConnecting.value
        isReconnecting.value = wsManager.isReconnecting.value
        reconnectAttempts.value = wsManager.getReconnectAttempts()
      }, { immediate: true })

      // 只有在没有连接时才连接
      if (!wsManager.isConnected.value) {
        await wsManager.connect()
      }
    } catch (error) {
      console.error('Failed to initialize WebSocket:', error)
    }
  }

  // 订阅事件
  const subscribe = (eventKey: string, callback: (message: WebSocketMessage) => void) => {
    if (!manager.value) {
      return () => {}
    }

    return manager.value.subscribe(eventKey, callback)
  }

  // 发送消息
  const send = (message: WebSocketMessage): boolean => {
    if (!manager.value) {
      return false
    }

    return manager.value.send(message)
  }

  // 断开连接
  const disconnect = () => {
    if (manager.value) {
      manager.value.disconnect()
    }
  }

  // 重新连接
  const reconnect = async () => {
    if (manager.value) {
      try {
        await manager.value.connect()
      } catch (error) {
        console.error('Failed to reconnect:', error)
      }
    }
  }

  // 获取连接状态
  const getConnectionStatus = () => {
    if (!manager.value) return WebSocketStatus.DISCONNECTED
    return manager.value.getStatus()
  }

  // 页面挂载时初始化
  onMounted(() => {
    initWebSocket()
  })

  // 页面卸载时清理
  onUnmounted(() => {
    // 清理当前组件的订阅
    if (manager.value) {
      // 注意：这里不销毁全局管理器，因为其他组件可能还在使用
      // 只清理当前组件的引用
      manager.value = null
    }
  })

  return {
    // 状态
    status,
    isConnected,
    isConnecting,
    isReconnecting,
    reconnectAttempts,
    
    // 方法
    subscribe,
    send,
    disconnect,
    reconnect,
    getConnectionStatus,
    
    // 管理器实例（高级用法）
    manager
  }
}

// 特定链的WebSocket组合式函数
export function useChainWebSocket(chain: 'eth' | 'btc') {
  // 使用全局WebSocket管理器，避免重复创建
  const manager = getWebSocketManager()
  
  if (!manager) {
    console.warn('WebSocket manager not found, creating one...')
    createWebSocketManager({
      url: import.meta.env.VITE_WS_BASE_URL || 'wss://localhost:8443/ws',
      autoReconnect: true,
      reconnectInterval: 3000,
      maxReconnectAttempts: 5,
      heartbeatInterval: 30000
    })
  }
  
  // 获取全局管理器的状态
  const status = computed(() => manager?.getStatus() || WebSocketStatus.DISCONNECTED)
  const isConnected = computed(() => manager?.isConnected.value || false)
  const isConnecting = computed(() => manager?.isConnecting.value || false)
  const isReconnecting = computed(() => manager?.isReconnecting.value || false)
  const reconnectAttempts = computed(() => manager?.getReconnectAttempts() || 0)

  // 订阅特定链的事件
  const subscribeChainEvent = (
    category: 'transaction' | 'block' | 'address' | 'stats' | 'network',
    callback: (message: WebSocketMessage) => void
  ) => {
    if (!manager) {
      console.warn('WebSocket manager not available')
      return () => {}
    }
    
    // 等待WebSocket连接建立后再发送订阅消息
    const sendSubscribeWhenReady = () => {
      if (manager && manager.isConnected.value) {
        const success = manager.sendSubscribe(category, chain)
        return success
      } else {
        return false
      }
    }
    
    // 立即尝试发送订阅消息
    let subscribeSent = sendSubscribeWhenReady()
    
    // 如果发送失败，等待连接建立后重试
    if (!subscribeSent) {
      const checkConnection = () => {
        if (manager && manager.isConnected.value) {
          // 添加延迟，确保状态完全同步
          setTimeout(() => {
            manager.sendSubscribe(category, chain)
          }, 200)
        }
      }
      
      // 监听连接状态变化
      watch(isConnected, (connected) => {
        if (connected) {
          checkConnection()
        }
      })
    }
    
    // 然后订阅本地事件
    return manager.subscribe(`event:${category}`, (message) => {
      // 只处理当前链的消息
      if (message.chain === chain || !message.chain) {
        callback(message)
      }
    })
  }

  // 订阅特定链的通知
  const subscribeChainNotification = (
    category: 'transaction' | 'block' | 'address' | 'stats' | 'network',
    callback: (message: WebSocketMessage) => void
  ) => {
    // 先发送订阅消息到服务器
    if (manager) {
      manager.sendSubscribe(category, chain)
    }
    
    // 然后订阅本地事件
    return manager?.subscribe(`notification:${category}`, (message) => {
      // 只处理当前链的消息
      if (message.chain === chain || !message.chain) {
        callback(message)
      }
    }) || (() => {})
  }

  // 发送链特定消息
  const sendChainMessage = (
    type: 'event' | 'notification',
    category: 'transaction' | 'block' | 'address' | 'stats' | 'network',
    data: Record<string, unknown>
  ) => {
    if (!manager) return false
    
    return manager.send({
      type,
      category,
      data,
      timestamp: Date.now(),
      chain
    })
  }

  // 取消订阅特定链的事件
  const unsubscribeChainEvent = (
    category: 'transaction' | 'block' | 'address' | 'stats' | 'network'
  ) => {
    // 发送取消订阅消息到服务器
    if (manager && manager.isConnected.value) {
      const success = manager.sendUnsubscribe(category, chain)
      if (!success) {
        console.warn('⚠️ WebSocket未连接，无法发送取消订阅消息')
      }
    } else {
      console.warn('⚠️ WebSocket未连接，无法发送取消订阅消息')
    }
  }

  return {
    status,
    isConnected,
    isConnecting,
    isReconnecting,
    reconnectAttempts,
    subscribeChainEvent,
    subscribeChainNotification,
    sendChainMessage,
    unsubscribeChainEvent
  }
}

// 自动刷新数据的组合式函数
export function useAutoRefresh<T>(
  fetchData: () => Promise<T>,
  refreshInterval: number = 30000
) {
  const data = ref<T | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  let refreshTimer: ReturnType<typeof setInterval> | null = null

  const loadData = async () => {
    try {
      loading.value = true
      error.value = null
      data.value = await fetchData()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load data'
    } finally {
      loading.value = false
    }
  }

  const startAutoRefresh = () => {
    stopAutoRefresh()
    refreshTimer = setInterval(loadData, refreshInterval)
  }

  const stopAutoRefresh = () => {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
  }

  // 页面挂载时开始自动刷新
  onMounted(() => {
    loadData()
    startAutoRefresh()
  })

  // 页面卸载时停止自动刷新
  onUnmounted(() => {
    stopAutoRefresh()
  })

  return {
    data,
    loading,
    error,
    loadData,
    startAutoRefresh,
    stopAutoRefresh
  }
} 