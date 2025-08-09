// WebSocket模拟服务器 - 仅用于开发测试
export class WebSocketMockServer {
  private clients: Set<WebSocket> = new Set()
  private interval: number | null = null

  constructor() {
    // 模拟WebSocket服务器
    this.startMockServer()
  }

  private startMockServer() {
    // 模拟定期发送数据
    this.interval = setInterval(() => {
      this.broadcastMockData()
    }, 5000) // 每5秒发送一次模拟数据
  }

  private broadcastMockData() {
    const mockData = this.generateMockData()
    this.clients.forEach(client => {
      if (client.readyState === WebSocket.OPEN) {
        client.dispatchEvent(new MessageEvent('message', {
          data: JSON.stringify(mockData)
        }))
      }
    })
  }

  private generateMockData() {
    const now = Date.now()
    const types = ['event', 'notification']
    const categories = ['transaction', 'block', 'stats', 'network']
    const chains = ['eth', 'btc']
    
    const type = types[Math.floor(Math.random() * types.length)]
    const category = categories[Math.floor(Math.random() * categories.length)]
    const chain = chains[Math.floor(Math.random() * chains.length)]

    let data: any = {}

    switch (category) {
      case 'transaction':
        data = {
          hash: `0x${Math.random().toString(16).substring(2, 34)}`,
          blockHeight: Math.floor(Math.random() * 1000000),
          timestamp: Math.floor(now / 1000),
          from: `0x${Math.random().toString(16).substring(2, 42)}`,
          to: `0x${Math.random().toString(16).substring(2, 42)}`,
          amount: Math.random() * 10,
          gasUsed: Math.floor(Math.random() * 100000),
          gasPrice: Math.floor(Math.random() * 1000000000),
          status: ['success', 'failed', 'pending'][Math.floor(Math.random() * 3)],
          nonce: Math.floor(Math.random() * 1000),
          input: '0x',
          totalTransactions: Math.floor(Math.random() * 1000000000)
        }
        break
      
      case 'block':
        data = {
          height: Math.floor(Math.random() * 1000000),
          timestamp: Math.floor(now / 1000),
          transactions: Math.floor(Math.random() * 2000) + 500,
          size: Math.floor(Math.random() * 1000000) + 500000,
          gasUsed: Math.floor(Math.random() * 15000000),
          gasLimit: 15000000,
          miner: `0x${Math.random().toString(16).substring(2, 42)}`,
          reward: Math.random() * 2,
          hash: `0x${Math.random().toString(16).substring(2, 66)}`,
          parentHash: `0x${Math.random().toString(16).substring(2, 66)}`,
          nonce: Math.random().toString(16).substring(2, 18),
          difficulty: Math.floor(Math.random() * 1000000000000),
          totalBlocks: Math.floor(Math.random() * 1000000)
        }
        break
      
      case 'stats':
        data = {
          totalBlocks: Math.floor(Math.random() * 1000000),
          totalTransactions: Math.floor(Math.random() * 1000000000),
          activeAddresses: Math.floor(Math.random() * 1000000),
          networkHashrate: Math.floor(Math.random() * 1000000000000000),
          avgGasPrice: Math.floor(Math.random() * 100000000000),
          avgBlockTime: Math.random() * 20 + 10,
          difficulty: Math.floor(Math.random() * 1000000000000000),
          dailyVolume: Math.random() * 100000
        }
        break
      
      case 'network':
        data = {
          type: 'heartbeat',
          timestamp: now
        }
        break
    }

    return {
      type,
      category,
      data,
      timestamp: now,
      chain
    }
  }

  // 模拟WebSocket连接
  public connect(url: string): Promise<WebSocket> {
    return new Promise((resolve) => {
      const mockWebSocket = new EventTarget() as any
      
      // 模拟WebSocket属性
      mockWebSocket.readyState = WebSocket.CONNECTING
      mockWebSocket.url = url
      
      // 模拟WebSocket方法
      mockWebSocket.send = (data: string) => {
        console.log('Mock WebSocket send:', data)
      }
      
      mockWebSocket.close = (code?: number, reason?: string) => {
        console.log('Mock WebSocket close:', code, reason)
        this.clients.delete(mockWebSocket)
        mockWebSocket.readyState = WebSocket.CLOSED
        mockWebSocket.dispatchEvent(new Event('close'))
      }
      
      // 模拟连接成功
      setTimeout(() => {
        mockWebSocket.readyState = WebSocket.OPEN
        this.clients.add(mockWebSocket)
        mockWebSocket.dispatchEvent(new Event('open'))
        resolve(mockWebSocket)
      }, 100)
    })
  }

  public destroy() {
    if (this.interval) {
      clearInterval(this.interval)
      this.interval = null
    }
    this.clients.clear()
  }
}

// 全局模拟服务器实例
let mockServer: WebSocketMockServer | null = null

// 创建模拟服务器
export function createMockServer(): WebSocketMockServer {
  if (!mockServer) {
    mockServer = new WebSocketMockServer()
  }
  return mockServer
}

// 销毁模拟服务器
export function destroyMockServer(): void {
  if (mockServer) {
    mockServer.destroy()
    mockServer = null
  }
}

// 重写WebSocket构造函数以使用模拟服务器
export function setupWebSocketMock(): void {
  const originalWebSocket = window.WebSocket
  
  window.WebSocket = class MockWebSocket extends originalWebSocket {
    constructor(url: string | URL, protocols?: string | string[]) {
      super(url, protocols)
      
      // 如果是我们的WebSocket URL，使用模拟服务器
      const urlString = url instanceof URL ? url.toString() : url
      if (urlString.includes('localhost:8080/ws')) {
        const mockServer = createMockServer()
        mockServer.connect(urlString).then((mockWs) => {
          // 将模拟WebSocket的事件转发到真实WebSocket
          mockWs.addEventListener('message', (event) => {
            this.dispatchEvent(new MessageEvent('message', { data: event.data }))
          })
          
          mockWs.addEventListener('open', () => {
            this.dispatchEvent(new Event('open'))
          })
          
          mockWs.addEventListener('close', () => {
            this.dispatchEvent(new Event('close'))
          })
          
          mockWs.addEventListener('error', () => {
            this.dispatchEvent(new Event('error'))
          })
        })
      }
    }
  } as typeof WebSocket
} 