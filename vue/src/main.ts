import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
// Polyfill Buffer for browser usage (needed by @solana/spl-token)
import { Buffer } from 'buffer'

import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'

// 初始化WebSocket连接管理
import { createWebSocketManager, setupVisibilityHandler } from './utils/websocket'

const app = createApp(App)

// Expose Buffer globally for libs expecting Node Buffer
;(window as any).Buffer = (window as any).Buffer || Buffer

// 在应用启动时创建全局WebSocket管理器
const wsManager = createWebSocketManager({
  url: import.meta.env.VITE_WS_BASE_URL || 'wss://localhost:8443/ws',
  autoReconnect: true,
  reconnectInterval: 3000,
  maxReconnectAttempts: 5,
  heartbeatInterval: 30000
})

// 设置页面可见性处理
setupVisibilityHandler(wsManager)

app.use(createPinia())
app.use(router)

// 初始化stores
const pinia = app._context.provides.pinia
if (pinia) {
  const authStore = useAuthStore(pinia)
  authStore.initialize()
}

app.mount('#app')
