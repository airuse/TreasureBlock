import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'

const app = createApp(App)

app.use(createPinia())
app.use(router)

// 初始化stores
const pinia = app._context.provides.pinia
if (pinia) {
  const authStore = useAuthStore(pinia)
  authStore.initialize()
}

app.mount('#app')
