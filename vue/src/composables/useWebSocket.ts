import { ref, watch, onMounted, onUnmounted } from 'vue'
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
      // 获取或创建WebSocket管理器
      let wsManager = getWebSocketManager()
      
      if (!wsManager) {
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

      // 连接WebSocket
      await wsManager.connect()
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
    // 注意：这里不销毁全局管理器，因为其他组件可能还在使用
    // 只清理当前组件的订阅
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
  const { subscribe, send, ...rest } = useWebSocket({
    url: import.meta.env.VITE_WS_BASE_URL || 'wss://localhost:8443/ws'
  })

  // 订阅特定链的事件
  const subscribeChainEvent = (
    category: 'transaction' | 'block' | 'address' | 'stats' | 'network',
    callback: (message: WebSocketMessage) => void
  ) => {
    // 等待WebSocket连接建立后再发送订阅消息
    const sendSubscribeWhenReady = () => {
      if (rest.manager?.value && rest.isConnected?.value) {
        const success = rest.manager.value.sendSubscribe(category, chain)
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
        if (rest.isConnected?.value) {
          // 添加延迟，确保状态完全同步
          setTimeout(() => {
            rest.manager?.value?.sendSubscribe(category, chain)
          }, 200)
        }
      }
      
      // 监听连接状态变化
      watch(rest.isConnected, (connected) => {
        if (connected) {
          checkConnection()
        }
      })
    }
    
    // 然后订阅本地事件
    return subscribe(`event:${category}`, (message) => {
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
    if (rest.manager?.value) {
      rest.manager.value.sendSubscribe(category, chain)
    }
    
    // 然后订阅本地事件
    return subscribe(`notification:${category}`, (message) => {
      // 只处理当前链的消息
      if (message.chain === chain || !message.chain) {
        callback(message)
      }
    })
  }

  // 发送链特定消息
  const sendChainMessage = (
    type: 'event' | 'notification',
    category: 'transaction' | 'block' | 'address' | 'stats' | 'network',
    data: Record<string, unknown>
  ) => {
    return send({
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
    if (rest.manager?.value && rest.isConnected?.value) {
      const success = rest.manager.value.sendUnsubscribe(category, chain)
    } else {
      console.warn('⚠️ WebSocket未连接，无法发送取消订阅消息')
    }
  }

  return {
    ...rest,
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