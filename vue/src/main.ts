import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

// 开发环境下设置WebSocket模拟
if (import.meta.env.DEV) {
  import('./utils/websocketMock').then(({ setupWebSocketMock }) => {
    setupWebSocketMock()
  })
}

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
