import { ref, computed } from 'vue'

// WebSocket消息类型定义 - 与后端三级分类保持一致
export interface WebSocketMessage {
  type: 'event' | 'notification'  // 第一级别：事件或通知
  category: 'transaction' | 'block' | 'address' | 'stats' | 'network'  // 第二级别：数据类型
  action?: 'create' | 'update' | 'delete' | 'fee_update' | 'status_update'  // 第三级别：动作类型
  data: Record<string, unknown>  // 第四级别：真实数据
  timestamp: number
  chain: 'eth' | 'btc'  // 区块链类型
}

// 费率数据结构
export interface FeeData {
  chain: string
  base_fee: string  // Base Fee (Wei)
  max_priority_fee: string  // Max Priority Fee (Wei)
  max_fee: string  // Max Fee (Wei)
  gas_price: string  // Legacy Gas Price (Wei)
  last_updated: number  // 最后更新时间戳
  block_number: number  // 当前区块号
  network_congestion: string  // 网络拥堵状态
}

// 费率等级
export interface FeeLevels {
  slow: FeeData
  normal: FeeData
  fast: FeeData
}

// 交易状态更新数据
export interface TransactionStatusUpdate {
  id: number
  status: string
  tx_hash?: string
  block_height?: number
  confirmations?: number
  error_msg?: string
  updated_at: string
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
        console.log('WebSocket已连接，无需重复连接')
        resolve()
        return
      }

      this.isManualClose = false
      this.status.value = WebSocketStatus.CONNECTING
      console.log(`正在连接WebSocket: ${this.options.url}`)

      try {
        this.ws = new WebSocket(this.options.url)
        this.setupEventHandlers(resolve, reject)
      } catch (error) {
        console.error('WebSocket连接失败:', error)
        this.status.value = WebSocketStatus.ERROR
        
        // 如果是初始连接失败，也尝试重连
        if (!this.isManualClose && this.options.autoReconnect) {
          this.scheduleReconnect()
        }
        
        reject(error)
      }
    })
  }

  // 设置事件处理器
  private setupEventHandlers(resolve: PromiseResolve<void>, reject: PromiseReject) {
    if (!this.ws) return

    this.ws.onopen = () => {
      console.log('WebSocket连接成功')
      this.status.value = WebSocketStatus.CONNECTED
      this.reconnectAttempts = 0
      this.startHeartbeat()
      
      // 延迟一点时间再resolve，确保状态更新完成
      setTimeout(() => {
        resolve()
      }, 100)
    }

    this.ws.onmessage = (event) => {
      try {
        // 尝试解析JSON消息
        const message = JSON.parse(event.data)
        
        // 处理ping消息（服务端发送的JSON格式）
        if (message.type === 'ping') {
          // 响应pong消息
          const pongMessage = {
            type: 'pong',
            data: 'pong',
            timestamp: Date.now()
          }
          this.ws?.send(JSON.stringify(pongMessage))
          return
        }
        
        // 处理其他消息
        this.handleMessage(message)
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
        // 如果不是JSON，可能是其他类型的消息，记录但不报错
        if (typeof event.data === 'string' && event.data !== 'ping') {
          console.warn('Received non-JSON message:', event.data)
        }
      }
    }

    this.ws.onclose = (event) => {
      console.log(`WebSocket连接关闭: 代码=${event.code}, 原因=${event.reason || '未知'}`)
      this.status.value = WebSocketStatus.DISCONNECTED
      this.stopHeartbeat()
      
      if (!this.isManualClose && this.options.autoReconnect) {
        console.log('WebSocket连接意外断开，将尝试重连...')
        this.scheduleReconnect()
      } else if (this.isManualClose) {
        console.log('WebSocket手动断开，不进行重连')
      }
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      this.status.value = WebSocketStatus.ERROR
      
      // 连接失败时也触发重连机制
      if (!this.isManualClose && this.options.autoReconnect) {
        this.scheduleReconnect()
      }
      
      reject(error)
    }
  }

  // 处理接收到的消息
  private handleMessage(message: any) {
    // 处理错误消息
    if (message.type === 'error') {
      console.warn('WebSocket error message:', message.error)
      return
    }
    
    // 处理订阅响应
    if (message.type === 'subscribed' || message.type === 'unsubscribed') {
      console.log('Subscription response:', message)
      return
    }
    
    // 处理pong响应
    if (message.type === 'pong') {
      return
    }

    // 验证消息格式（只对event和notification类型进行严格验证）
    if (message.type === 'event' || message.type === 'notification') {
      if (!this.validateMessage(message)) {
        console.warn('Invalid message format:', message)
        return
      }
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
      category: category as any, // 转换为MessageCategory类型
      chain: chain as any        // 转换为ChainType类型
    }
    
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      return false
    }

    try {
      this.ws.send(JSON.stringify(subscribeMessage))
      return true
    } catch (error) {
      return false
    }
  }

  // 发送取消订阅消息
  public sendUnsubscribe(category: string, chain: string): boolean {
    const unsubscribeMessage = {
      type: 'unsubscribe',
      category: category as any, // 转换为MessageCategory类型
      chain: chain as any        // 转换为ChainType类型
    }
    
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      return false
    }

    try {
      this.ws.send(JSON.stringify(unsubscribeMessage))
      return true
    } catch (error) {
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
    // 移除前端主动发送心跳的逻辑，只响应服务端的ping
    // 心跳检测现在完全由服务端控制
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
      console.warn(`WebSocket重连失败，已达到最大重连次数 ${this.options.maxReconnectAttempts}`)
      this.status.value = WebSocketStatus.ERROR
      return
    }

    this.reconnectAttempts++
    this.status.value = WebSocketStatus.RECONNECTING
    
    console.log(`WebSocket重连中... (第${this.reconnectAttempts}次尝试，${this.options.reconnectInterval}ms后重试)`)

    this.reconnectTimer = setTimeout(() => {
      this.connect().catch(error => {
        console.error(`WebSocket重连失败 (第${this.reconnectAttempts}次):`, error)
        // 继续尝试重连
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
    } else {
      // 页面显示时，确保连接正常
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