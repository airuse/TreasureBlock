# WebSocket 实现文档

## 概述

本项目实现了一个完整的WebSocket系统，支持实时数据推送、自动重连、资源管理等功能。WebSocket采用三级分类的消息结构，确保数据的有序性和可扩展性。

## 消息结构

### 三级分类系统

```typescript
interface WebSocketMessage {
  type: 'event' | 'notification'  // 第一级别：事件或通知
  category: 'transaction' | 'block' | 'address' | 'stats' | 'network'  // 第二级别：数据类型
  data: any  // 第三级别：真实数据
  timestamp: number
  chain?: 'eth' | 'btc'  // 可选：指定链类型
}
```

### 消息类型说明

- **第一级别 (type)**：
  - `event`: 实时事件，如新区块、新交易
  - `notification`: 系统通知，如网络状态变化

- **第二级别 (category)**：
  - `transaction`: 交易相关数据
  - `block`: 区块相关数据
  - `address`: 地址相关数据
  - `stats`: 统计信息
  - `network`: 网络状态

- **第三级别 (data)**：具体的JSON数据

## 核心功能

### 1. WebSocket管理器 (WebSocketManager)

```typescript
import { createWebSocketManager } from '@/utils/websocket'

const manager = createWebSocketManager({
  url: 'ws://localhost:8080/ws',
  autoReconnect: true,
  reconnectInterval: 3000,
  maxReconnectAttempts: 5,
  heartbeatInterval: 30000
})
```

**主要特性：**
- 自动重连机制
- 心跳检测
- 事件订阅/取消订阅
- 连接状态管理
- 资源自动清理

### 2. Vue组合式函数

#### 基础用法

```typescript
import { useWebSocket } from '@/composables/useWebSocket'

const { 
  isConnected, 
  subscribe, 
  send, 
  disconnect 
} = useWebSocket({
  url: 'ws://localhost:8080/ws'
})
```

#### 链特定用法

```typescript
import { useChainWebSocket } from '@/composables/useWebSocket'

const { 
  subscribeChainEvent, 
  subscribeChainNotification,
  sendChainMessage 
} = useChainWebSocket('eth')
```

### 3. 事件订阅

```typescript
// 订阅特定事件
const unsubscribe = subscribeChainEvent('block', (message) => {
  console.log('New block:', message.data)
  // 处理新区块数据
})

// 订阅通知
const unsubscribeNotification = subscribeChainNotification('stats', (message) => {
  console.log('Stats update:', message.data)
  // 处理统计信息更新
})

// 组件卸载时取消订阅
onUnmounted(() => {
  unsubscribe()
  unsubscribeNotification()
})
```

### 4. 发送消息

```typescript
// 发送链特定消息
sendChainMessage('event', 'block', {
  height: 123456,
  hash: '0x...',
  timestamp: Date.now()
})
```

## 在组件中使用

### 首页实时数据更新

```vue
<script setup lang="ts">
import { useChainWebSocket } from '@/composables/useWebSocket'

const { subscribeChainEvent } = useChainWebSocket('btc')

// 订阅新区块事件
const unsubscribeBlocks = subscribeChainEvent('block', (message) => {
  if (message.data && message.data.height) {
    // 更新最新区块列表
    latestBlocks.value.unshift(message.data)
    if (latestBlocks.value.length > 5) {
      latestBlocks.value = latestBlocks.value.slice(0, 5)
    }
  }
})

// 订阅新交易事件
const unsubscribeTransactions = subscribeChainEvent('transaction', (message) => {
  if (message.data && message.data.hash) {
    // 更新最新交易列表
    latestTransactions.value.unshift(message.data)
    if (latestTransactions.value.length > 5) {
      latestTransactions.value = latestTransactions.value.slice(0, 5)
    }
  }
})

// 订阅统计信息更新
const unsubscribeStats = subscribeChainEvent('stats', (message) => {
  if (message.data) {
    stats.value = { ...stats.value, ...message.data }
  }
})

onUnmounted(() => {
  unsubscribeBlocks()
  unsubscribeTransactions()
  unsubscribeStats()
})
</script>
```

## 连接状态管理

### 导航栏状态显示

```vue
<template>
  <div class="flex items-center space-x-2">
    <div 
      :class="[
        'w-2 h-2 rounded-full',
        isConnected ? 'bg-green-400' : 'bg-red-400'
      ]"
    ></div>
    <span class="text-sm text-gray-600">
      {{ isConnected ? '网络正常' : '连接失败' }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { useWebSocket } from '@/composables/useWebSocket'

const { isConnected } = useWebSocket()
</script>
```

## 资源管理

### 自动清理机制

1. **组件卸载时**：自动取消事件订阅
2. **页面隐藏时**：可选择断开连接或保持连接
3. **应用关闭时**：自动销毁WebSocket管理器

### 内存泄漏防护

```typescript
// 在组件中使用时，确保在onUnmounted中清理
onUnmounted(() => {
  // 取消所有订阅
  unsubscribeBlocks()
  unsubscribeTransactions()
  unsubscribeStats()
})
```

## 开发环境模拟

在开发环境中，系统会自动启用WebSocket模拟服务器，每5秒发送一次模拟数据：

```typescript
// main.ts
if (import.meta.env.DEV) {
  import('./utils/websocketMock').then(({ setupWebSocketMock }) => {
    setupWebSocketMock()
  })
}
```

## 配置选项

### WebSocketOptions

```typescript
interface WebSocketOptions {
  url: string                    // WebSocket服务器地址
  autoReconnect?: boolean        // 是否自动重连 (默认: true)
  reconnectInterval?: number     // 重连间隔 (默认: 3000ms)
  maxReconnectAttempts?: number  // 最大重连次数 (默认: 5)
  heartbeatInterval?: number     // 心跳间隔 (默认: 30000ms)
}
```

## 错误处理

### 连接错误

```typescript
const { status, reconnect } = useWebSocket()

// 监听连接状态
watch(status, (newStatus) => {
  if (newStatus === WebSocketStatus.ERROR) {
    console.error('WebSocket连接失败')
    // 可以尝试手动重连
    reconnect()
  }
})
```

### 消息解析错误

系统会自动处理消息解析错误，并在控制台输出警告信息。

## 最佳实践

1. **事件订阅**：在组件挂载时订阅，卸载时取消订阅
2. **状态管理**：使用响应式状态管理连接状态
3. **错误处理**：实现适当的错误处理和用户提示
4. **性能优化**：避免在事件回调中执行重操作
5. **资源清理**：确保在组件卸载时清理所有资源

## 扩展性

系统设计具有良好的扩展性：

- 可以轻松添加新的消息类型和分类
- 支持多链数据推送
- 可以扩展为支持多房间/频道的系统
- 支持自定义消息验证和处理逻辑 