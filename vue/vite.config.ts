import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import fs from 'fs'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  // 定义全局常量
  define: {
    __USE_MOCK__: false, // 启用真实API，使用真实数据
  },
  server: {
    https: {
      key: fs.readFileSync(path.resolve(__dirname, '../server/certs/localhost.key')),
      cert: fs.readFileSync(path.resolve(__dirname, '../server/certs/localhost.crt')),
    },
    port: 5173,
    host: 'localhost',
    proxy: {
      '/api': {
        target: 'https://localhost:8443',
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path
      }
    }
  },
})
