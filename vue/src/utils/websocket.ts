import { ref, computed } from 'vue'

// WebSocket消息类型定义 - 与后端三级分类保持一致
export interface WebSocketMessage {
  type: 'event' | 'notification'  // 第一级别：事件或通知
  category: 'transaction' | 'block' | 'address' | 'stats' | 'network'  // 第二级别：数据类型
  data: Record<string, unknown>  // 第三级别：真实数据
  timestamp: number
  chain: 'eth' | 'btc'  // 区块链类型
}

// WebSocket事件类型
export interface WebSocketEvent {
  type: string
  data: WebSocketMessage
}

// WebSocket配置选项
export interface WebSocketOptions {
  url: string
  autoReconnect?: boolean
  reconnectInterval?: number
  maxReconnectAttempts?: number
  heartbeatInterval?: number
}

// WebSocket状态枚举
export enum WebSocketStatus {
  CONNECTING = 'connecting',
  CONNECTED = 'connected',
  DISCONNECTED = 'disconnected',
  RECONNECTING = 'reconnecting',
  ERROR = 'error'
}

// 定义回调函数类型
type WebSocketCallback = (message: WebSocketMessage) => void
type PromiseResolve<T> = (value: T | PromiseLike<T>) => void
type PromiseReject = (reason?: unknown) => void

class WebSocketManager {
  private ws: WebSocket | null = null
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null
  private reconnectAttempts = 0
  private eventListeners: Map<string, Set<WebSocketCallback>> = new Map()
  private options: WebSocketOptions
  private isManualClose = false

  // 响应式状态
  public status = ref<WebSocketStatus>(WebSocketStatus.DISCONNECTED)
  public isConnected = computed(() => this.status.value === WebSocketStatus.CONNECTED)
  public isConnecting = computed(() => this.status.value === WebSocketStatus.CONNECTING)
  public isReconnecting = computed(() => this.status.value === WebSocketStatus.RECONNECTING)

  constructor(options: WebSocketOptions) {
    this.options = {
      autoReconnect: true,
      reconnectInterval: 3000,
      maxReconnectAttempts: 5,
      heartbeatInterval: 30000,
      ...options
    }
  }

  // 连接WebSocket
  public connect(): Promise<void> {
    return new Promise((resolve: PromiseResolve<void>, reject: PromiseReject) => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        resolve()
        return
      }

      this.isManualClose = false
      this.status.value = WebSocketStatus.CONNECTING

      try {
        this.ws = new WebSocket(this.options.url)
        this.setupEventHandlers(resolve, reject)
      } catch (error) {
        this.status.value = WebSocketStatus.ERROR
        reject(error)
      }
    })
  }

  // 设置事件处理器
  private setupEventHandlers(resolve: PromiseResolve<void>, reject: PromiseReject) {
    if (!this.ws) return

    this.ws.onopen = () => {
      console.log('WebSocket connected')
      this.status.value = WebSocketStatus.CONNECTED
      this.reconnectAttempts = 0
      this.startHeartbeat()
      resolve()
    }

    this.ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data)
        this.handleMessage(message)
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
      }
    }

    this.ws.onclose = (event) => {
      console.log('WebSocket closed:', event.code, event.reason)
      this.status.value = WebSocketStatus.DISCONNECTED
      this.stopHeartbeat()
      
      if (!this.isManualClose && this.options.autoReconnect) {
        this.scheduleReconnect()
      }
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      this.status.value = WebSocketStatus.ERROR
      reject(error)
    }
  }

  // 处理接收到的消息
  private handleMessage(message: WebSocketMessage) {
    // 验证消息格式
    if (!this.validateMessage(message)) {
      console.warn('Invalid message format:', message)
      return
    }

    // 触发对应的事件监听器
    const eventKey = `${message.type}:${message.category}`
    const listeners = this.eventListeners.get(eventKey)
    
    if (listeners) {
      listeners.forEach(listener => {
        try {
          listener(message)
        } catch (error) {
          console.error('Error in event listener:', error)
        }
      })
    }
  }

  // 验证消息格式
  private validateMessage(message: unknown): message is WebSocketMessage {
    return (
      message &&
      typeof message === 'object' &&
      message !== null &&
      'type' in message &&
      typeof (message as WebSocketMessage).type === 'string' &&
      ['event', 'notification'].includes((message as WebSocketMessage).type) &&
      'category' in message &&
      typeof (message as WebSocketMessage).category === 'string' &&
      ['transaction', 'block', 'address', 'stats', 'network'].includes((message as WebSocketMessage).category) &&
      'data' in message &&
      (message as WebSocketMessage).data !== undefined &&
      'timestamp' in message &&
      typeof (message as WebSocketMessage).timestamp === 'number' &&
      'chain' in message &&
      typeof (message as WebSocketMessage).chain === 'string' &&
      ['eth', 'btc'].includes((message as WebSocketMessage).chain)
    ) as boolean
  }

  // 发送消息
  public send(message: WebSocketMessage): boolean {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket is not connected')
      return false
    }

    try {
      this.ws.send(JSON.stringify(message))
      return true
    } catch (error) {
      console.error('Failed to send message:', error)
      return false
    }
  }

  // 发送订阅消息
  public sendSubscribe(category: string, chain: string): boolean {
    const subscribeMessage = {
      type: 'subscribe',
      category: category,
      chain: chain
    }
    
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket is not connected')
      return false
    }

    try {
      this.ws.send(JSON.stringify(subscribeMessage))
      return true
    } catch (error) {
      console.error('Failed to send subscribe message:', error)
      return false
    }
  }

  // 发送取消订阅消息
  public sendUnsubscribe(category: string, chain: string): boolean {
    const unsubscribeMessage = {
      type: 'unsubscribe',
      category: category,
      chain: chain
    }
    
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket is not connected')
      return false
    }

    try {
      this.ws.send(JSON.stringify(unsubscribeMessage))
      return true
    } catch (error) {
      console.error('Failed to send unsubscribe message:', error)
      return false
    }
  }

  // 订阅事件
  public subscribe(eventKey: string, callback: WebSocketCallback): () => void {
    if (!this.eventListeners.has(eventKey)) {
      this.eventListeners.set(eventKey, new Set())
    }
    
    this.eventListeners.get(eventKey)!.add(callback)

    // 返回取消订阅的函数
    return () => {
      const listeners = this.eventListeners.get(eventKey)
      if (listeners) {
        listeners.delete(callback)
        if (listeners.size === 0) {
          this.eventListeners.delete(eventKey)
        }
      }
    }
  }

  // 取消订阅
  public unsubscribe(eventKey: string, callback: WebSocketCallback): void {
    const listeners = this.eventListeners.get(eventKey)
    if (listeners) {
      listeners.delete(callback)
      if (listeners.size === 0) {
        this.eventListeners.delete(eventKey)
      }
    }
  }

  // 开始心跳检测
  private startHeartbeat() {
    if (this.options.heartbeatInterval && this.options.heartbeatInterval > 0) {
      this.heartbeatTimer = setInterval(() => {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
          this.send({
            type: 'event',
            category: 'network',
            data: { type: 'heartbeat' },
            timestamp: Date.now(),
            chain: 'eth'
          })
        }
      }, this.options.heartbeatInterval)
    }
  }

  // 停止心跳检测
  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  // 安排重连
  private scheduleReconnect() {
    if (this.reconnectAttempts >= this.options.maxReconnectAttempts!) {
      console.error('Max reconnection attempts reached')
      return
    }

    this.reconnectAttempts++
    this.status.value = WebSocketStatus.RECONNECTING

    console.log(`Scheduling reconnection attempt ${this.reconnectAttempts}/${this.options.maxReconnectAttempts}`)

    this.reconnectTimer = setTimeout(() => {
      this.connect().catch(error => {
        console.error('Reconnection failed:', error)
        this.scheduleReconnect()
      })
    }, this.options.reconnectInterval)
  }

  // 断开连接
  public disconnect(): void {
    this.isManualClose = true
    this.stopHeartbeat()
    
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }

    if (this.ws) {
      this.ws.close(1000, 'Manual disconnect')
      this.ws = null
    }

    this.status.value = WebSocketStatus.DISCONNECTED
    this.reconnectAttempts = 0
  }

  // 清理资源
  public destroy(): void {
    this.disconnect()
    this.eventListeners.clear()
  }

  // 获取连接状态
  public getStatus(): WebSocketStatus {
    return this.status.value
  }

  // 获取重连次数
  public getReconnectAttempts(): number {
    return this.reconnectAttempts
  }
}

// 全局WebSocket实例
let globalWebSocketManager: WebSocketManager | null = null

// 创建WebSocket管理器
export function createWebSocketManager(options: WebSocketOptions): WebSocketManager {
  if (globalWebSocketManager) {
    globalWebSocketManager.destroy()
  }
  
  globalWebSocketManager = new WebSocketManager(options)
  return globalWebSocketManager
}

// 获取全局WebSocket管理器
export function getWebSocketManager(): WebSocketManager | null {
  return globalWebSocketManager
}

// 销毁全局WebSocket管理器
export function destroyWebSocketManager(): void {
  if (globalWebSocketManager) {
    globalWebSocketManager.destroy()
    globalWebSocketManager = null
  }
}

// 页面可见性变化处理
let visibilityChangeHandler: (() => void) | null = null

// 设置页面可见性监听
export function setupVisibilityHandler(manager: WebSocketManager): void {
  if (visibilityChangeHandler) {
    document.removeEventListener('visibilitychange', visibilityChangeHandler)
  }

  visibilityChangeHandler = () => {
    if (document.hidden) {
      // 页面隐藏时，可以选择断开连接或保持连接
      console.log('Page hidden')
    } else {
      // 页面显示时，确保连接正常
      console.log('Page visible')
      if (!manager.isConnected.value) {
        manager.connect().catch(console.error)
      }
    }
  }

  document.addEventListener('visibilitychange', visibilityChangeHandler)
}

// 清理页面可见性监听
export function cleanupVisibilityHandler(): void {
  if (visibilityChangeHandler) {
    document.removeEventListener('visibilitychange', visibilityChangeHandler)
    visibilityChangeHandler = null
  }
}

export default WebSocketManager 