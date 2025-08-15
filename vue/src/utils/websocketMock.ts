// 定义MockWebSocket类型
interface MockWebSocket extends EventTarget {
  url: string
  readyState: number
  protocol: string
  extensions: string
  bufferedAmount: number
  onopen: ((event: Event) => void) | null
  onclose: ((event: CloseEvent) => void) | null
  onmessage: ((event: MessageEvent) => void) | null
  onerror: ((event: Event) => void) | null
  close(code?: number, reason?: string): void
  send(data: string | ArrayBufferLike | Blob | ArrayBufferView): void
  addEventListener(type: string, listener: EventListener | null, options?: boolean | AddEventListenerOptions): void
  removeEventListener(type: string, listener: EventListener | null, options?: boolean | EventListenerOptions): void
  dispatchEvent(event: Event): boolean
}

class MockWebSocketManager {
  private mockConnections: Map<string, MockWebSocket> = new Map()
  private messageHandlers: Map<string, ((message: any) => void)[]> = new Map()
  private connectionHandlers: Map<string, (() => void)[]> = new Map()
  private disconnectionHandlers: Map<string, (() => void)[]> = new Map()

  constructor() {
    // 模拟网络延迟
    this.simulateNetworkLatency()
  }

  // 模拟网络延迟
  private simulateNetworkLatency() {
    setInterval(() => {
      // 随机模拟网络抖动
      if (Math.random() < 0.1) {
        this.triggerRandomEvent()
      }
    }, 5000)
  }

  // 触发随机事件
  private triggerRandomEvent() {
    const events = ['block', 'transaction', 'address']
    const randomEvent = events[Math.floor(Math.random() * events.length)]
    
    if (this.messageHandlers.has(randomEvent)) {
      const handlers = this.messageHandlers.get(randomEvent)!
      handlers.forEach(handler => {
        setTimeout(() => {
          handler(this.generateMockData(randomEvent))
        }, Math.random() * 1000)
      })
    }
  }

  // 生成模拟数据
  private generateMockData(eventType: string) {
    switch (eventType) {
      case 'block':
        return {
          type: 'block',
          data: {
            height: Math.floor(Math.random() * 1000000) + 1000000,
            timestamp: Date.now(),
            transactions: Math.floor(Math.random() * 1000) + 100,
            size: Math.floor(Math.random() * 1000000) + 1000000,
            gasUsed: Math.floor(Math.random() * 10000000) + 10000000,
            gasLimit: Math.floor(Math.random() * 20000000) + 20000000,
            miner: '0x' + Math.random().toString(16).substr(2, 40),
            reward: Math.random() * 2 + 2
          }
        }
      case 'transaction':
        return {
          type: 'transaction',
          data: {
            hash: '0x' + Math.random().toString(16).substr(2, 64),
            from: '0x' + Math.random().toString(16).substr(2, 40),
            to: '0x' + Math.random().toString(16).substr(2, 40),
            value: Math.random() * 1000,
            gasPrice: Math.floor(Math.random() * 20e9) + 20e9,
            gasUsed: Math.floor(Math.random() * 100000) + 21000,
            timestamp: Date.now()
          }
        }
      case 'address':
        return {
          type: 'address',
          data: {
            address: '0x' + Math.random().toString(16).substr(2, 40),
            balance: Math.random() * 10000,
            transactionCount: Math.floor(Math.random() * 10000) + 100
          }
        }
      default:
        return { type: 'unknown', data: {} }
    }
  }

  // 连接WebSocket
  public connect(url: string): Promise<MockWebSocket> {
    return new Promise((resolve) => {
      // 模拟连接延迟
      setTimeout(() => {
        const mockWebSocket = new EventTarget() as MockWebSocket
        
        // 设置基本属性
        mockWebSocket.url = url
        mockWebSocket.readyState = 1 // OPEN
        mockWebSocket.protocol = ''
        mockWebSocket.extensions = ''
        mockWebSocket.bufferedAmount = 0
        
        // 设置事件处理器
        mockWebSocket.onopen = null
        mockWebSocket.onclose = null
        mockWebSocket.onmessage = null
        mockWebSocket.onerror = null
        
        // 实现close方法
        mockWebSocket.close = (code?: number, reason?: string) => {
          mockWebSocket.readyState = 3 // CLOSED
          this.mockConnections.delete(url)
          
          // 触发关闭事件
          const closeEvent = new CloseEvent('close', {
            code: code || 1000,
            reason: reason || 'Normal closure',
            wasClean: true
          })
          mockWebSocket.dispatchEvent(closeEvent)
        }
        
        // 实现send方法
        mockWebSocket.send = (data: string | ArrayBufferLike | Blob | ArrayBufferView) => {
          // 模拟发送延迟
          setTimeout(() => {
            // 这里可以处理发送的数据
            console.log('Mock WebSocket sent:', data)
          }, Math.random() * 100)
        }
        
        // 添加到连接列表
        this.mockConnections.set(url, mockWebSocket)
        
        // 触发连接事件
        const openEvent = new Event('open')
        mockWebSocket.dispatchEvent(openEvent)
        
        resolve(mockWebSocket)
      }, Math.random() * 500 + 100)
    })
  }

  // 订阅链上事件
  public subscribeChainEvent(eventType: string, handler: (message: any) => void) {
    if (!this.messageHandlers.has(eventType)) {
      this.messageHandlers.set(eventType, [])
    }
    this.messageHandlers.get(eventType)!.push(handler)
    
    // 返回取消订阅函数
    return () => {
      const handlers = this.messageHandlers.get(eventType)!
      const index = handlers.indexOf(handler)
      if (index > -1) {
        handlers.splice(index, 1)
      }
    }
  }

  // 模拟接收消息
  public simulateMessage(eventType: string, message: any) {
    if (this.messageHandlers.has(eventType)) {
      const handlers = this.messageHandlers.get(eventType)!
      handlers.forEach(handler => {
        setTimeout(() => {
          handler(message)
        }, Math.random() * 100)
      })
    }
  }

  // 模拟连接状态变化
  public simulateConnectionChange(url: string, isConnected: boolean) {
    const mockWs = this.mockConnections.get(url)
    if (mockWs) {
      if (isConnected) {
        mockWs.readyState = 1 // OPEN
        const openEvent = new Event('open')
        mockWs.dispatchEvent(openEvent)
      } else {
        mockWs.readyState = 3 // CLOSED
        const closeEvent = new CloseEvent('close', {
          code: 1006,
          reason: 'Abnormal closure',
          wasClean: false
        })
        mockWs.dispatchEvent(closeEvent)
      }
    }
  }

  // 获取连接状态
  public getConnectionStatus(url: string): boolean {
    const mockWs = this.mockConnections.get(url)
    return mockWs ? mockWs.readyState === 1 : false
  }

  // 关闭所有连接
  public closeAllConnections() {
    this.mockConnections.forEach((mockWs, url) => {
      mockWs.close()
    })
    this.mockConnections.clear()
  }

  // 模拟网络错误
  public simulateNetworkError(url: string) {
    const mockWs = this.mockConnections.get(url)
    if (mockWs) {
      const errorEvent = new Event('error')
      mockWs.dispatchEvent(errorEvent)
    }
  }

  // 模拟消息延迟
  public simulateMessageDelay(eventType: string, delay: number) {
    if (this.messageHandlers.has(eventType)) {
      const handlers = this.messageHandlers.get(eventType)!
      handlers.forEach(handler => {
        setTimeout(() => {
          handler(this.generateMockData(eventType))
        }, delay)
      })
    }
  }

  // 获取统计信息
  public getStats() {
    return {
      activeConnections: this.mockConnections.size,
      totalHandlers: Array.from(this.messageHandlers.values()).reduce((sum, handlers) => sum + handlers.length, 0),
      supportedEvents: Array.from(this.messageHandlers.keys())
    }
  }
}

// 创建全局实例
const mockWebSocketManager = new MockWebSocketManager()

// 导出函数
export const connectMockWebSocket = (url: string) => mockWebSocketManager.connect(url)
export const subscribeChainEvent = (eventType: string, handler: (message: any) => void) => mockWebSocketManager.subscribeChainEvent(eventType, handler)
export const simulateMessage = (eventType: string, message: any) => mockWebSocketManager.simulateMessage(eventType, message)
export const simulateConnectionChange = (url: string, isConnected: boolean) => mockWebSocketManager.simulateConnectionChange(url, isConnected)
export const getConnectionStatus = (url: string) => mockWebSocketManager.getConnectionStatus(url)
export const closeAllConnections = () => mockWebSocketManager.closeAllConnections()
export const simulateNetworkError = (url: string) => mockWebSocketManager.simulateNetworkError(url)
export const simulateMessageDelay = (eventType: string, delay: number) => mockWebSocketManager.simulateMessageDelay(eventType, delay)
export const getStats = () => mockWebSocketManager.getStats()

// 默认导出
export default mockWebSocketManager 