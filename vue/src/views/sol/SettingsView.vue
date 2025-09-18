<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">Solana 系统设置</h1>
    </div>

    <!-- 设置表单 -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- 左侧设置分类 -->
      <div class="lg:col-span-1">
        <div class="card">
          <nav class="space-y-1">
            <button 
              v-for="section in settingsSections" 
              :key="section.id"
              @click="activeSection = section.id"
              :class="[
                activeSection === section.id 
                  ? 'bg-blue-50 text-blue-700 border-blue-500' 
                  : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900 border-transparent',
                'w-full flex items-center px-4 py-2 text-sm font-medium border-l-4 transition-colors duration-200'
              ]"
            >
              <component :is="section.icon" class="mr-3 h-5 w-5" />
              {{ section.name }}
            </button>
          </nav>
        </div>
      </div>

      <!-- 右侧设置内容 -->
      <div class="lg:col-span-2">
        <!-- 显示设置 -->
        <div v-if="activeSection === 'display'" class="card">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">显示设置</h3>
          <div class="space-y-6">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">主题模式</label>
              <select 
                v-model="settings.theme" 
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="light">浅色模式</option>
                <option value="dark">深色模式</option>
                <option value="auto">跟随系统</option>
              </select>
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">语言</label>
              <select 
                v-model="settings.language" 
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="zh-CN">简体中文</option>
                <option value="en-US">English</option>
              </select>
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">每页显示数量</label>
              <select 
                v-model="settings.pageSize" 
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="10">10</option>
                <option value="25">25</option>
                <option value="50">50</option>
                <option value="100">100</option>
              </select>
            </div>
            
            <div class="flex items-center">
              <input 
                v-model="settings.autoRefresh" 
                type="checkbox" 
                class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label class="ml-2 block text-sm text-gray-900">启用自动刷新</label>
            </div>
            
            <div v-if="settings.autoRefresh">
              <label class="block text-sm font-medium text-gray-700 mb-2">刷新间隔（秒）</label>
              <input 
                v-model="settings.refreshInterval" 
                type="number" 
                min="5" 
                max="300"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>
        </div>

        <!-- 网络设置 -->
        <div v-if="activeSection === 'network'" class="card">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">网络设置</h3>
          <div class="space-y-6">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">区块链网络</label>
              <select 
                v-model="settings.network" 
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="mainnet">以太坊主网</option>
                <option value="goerli">Goerli测试网</option>
                <option value="sepolia">Sepolia测试网</option>
                <option value="custom">自定义网络</option>
              </select>
            </div>
            
            <div v-if="settings.network === 'custom'">
              <label class="block text-sm font-medium text-gray-700 mb-2">RPC端点</label>
              <input 
                v-model="settings.customRpc" 
                type="text" 
                placeholder="https://your-rpc-endpoint.com"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">API端点</label>
              <input 
                v-model="settings.apiEndpoint" 
                type="text" 
                placeholder="https://api.example.com"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            
            <div class="flex items-center">
              <input 
                v-model="settings.useWebSocket" 
                type="checkbox" 
                class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label class="ml-2 block text-sm text-gray-900">使用WebSocket连接</label>
            </div>
          </div>
        </div>

        <!-- 通知设置 -->
        <div v-if="activeSection === 'notifications'" class="card">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">通知设置</h3>
          <div class="space-y-6">
            <div class="flex items-center">
              <input 
                v-model="settings.emailNotifications" 
                type="checkbox" 
                class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label class="ml-2 block text-sm text-gray-900">启用邮件通知</label>
            </div>
            
            <div v-if="settings.emailNotifications">
              <label class="block text-sm font-medium text-gray-700 mb-2">邮箱地址</label>
              <input 
                v-model="settings.emailAddress" 
                type="email" 
                placeholder="your@email.com"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            
            <div class="flex items-center">
              <input 
                v-model="settings.browserNotifications" 
                type="checkbox" 
                class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label class="ml-2 block text-sm text-gray-900">启用浏览器通知</label>
            </div>
            
            <div class="flex items-center">
              <input 
                v-model="settings.priceAlerts" 
                type="checkbox" 
                class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label class="ml-2 block text-sm text-gray-900">价格变动提醒</label>
            </div>
            
            <div class="flex items-center">
              <input 
                v-model="settings.blockAlerts" 
                type="checkbox" 
                class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label class="ml-2 block text-sm text-gray-900">新区块提醒</label>
            </div>
          </div>
        </div>

        <!-- 安全设置 -->
        <div v-if="activeSection === 'security'" class="card">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">安全设置</h3>
          <div class="space-y-6">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">会话超时（分钟）</label>
              <input 
                v-model="settings.sessionTimeout" 
                type="number" 
                min="5" 
                max="1440"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            
            <div class="flex items-center">
              <input 
                v-model="settings.requirePassword" 
                type="checkbox" 
                class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label class="ml-2 block text-sm text-gray-900">敏感操作需要密码确认</label>
            </div>
            
            <div class="flex items-center">
              <input 
                v-model="settings.logActivity" 
                type="checkbox" 
                class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label class="ml-2 block text-sm text-gray-900">记录用户活动日志</label>
            </div>
            
            <div>
              <button @click="exportSettings" class="btn-secondary">
                导出设置
              </button>
              <button @click="importSettings" class="btn-secondary ml-2">
                导入设置
              </button>
            </div>
          </div>
        </div>

        <!-- 关于 -->
        <div v-if="activeSection === 'about'" class="card">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">关于系统</h3>
          <div class="space-y-4">
            <div>
              <h4 class="text-sm font-medium text-gray-700">版本信息</h4>
              <p class="text-sm text-gray-900">区块链浏览器 v1.0.0</p>
            </div>
            
            <div>
              <h4 class="text-sm font-medium text-gray-700">技术栈</h4>
              <p class="text-sm text-gray-900">Vue 3 + TypeScript + TailwindCSS</p>
            </div>
            
            <div>
              <h4 class="text-sm font-medium text-gray-700">开发团队</h4>
              <p class="text-sm text-gray-900">区块链浏览器开发团队</p>
            </div>
            
            <div>
              <h4 class="text-sm font-medium text-gray-700">许可证</h4>
              <p class="text-sm text-gray-900">MIT License</p>
            </div>
            
            <div>
              <button @click="checkUpdate" class="btn-secondary">
                检查更新
              </button>
            </div>
          </div>
        </div>

        <!-- 保存按钮 -->
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="resetSettings" class="btn-secondary">
            重置
          </button>
          <button @click="saveSettings" class="btn-primary">
            保存设置
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { 
  EyeIcon, 
  GlobeAltIcon, 
  BellIcon, 
  ShieldCheckIcon,
  InformationCircleIcon
} from '@heroicons/vue/20/solid'

// 响应式数据
const activeSection = ref('display')

const settings = reactive({
  // 显示设置
  theme: 'light',
  language: 'zh-CN',
  pageSize: 25,
  autoRefresh: true,
  refreshInterval: 30,
  
  // 网络设置
  network: 'mainnet',
  customRpc: '',
  apiEndpoint: 'https://api.example.com',
  useWebSocket: false,
  
  // 通知设置
  emailNotifications: false,
  emailAddress: '',
  browserNotifications: true,
  priceAlerts: false,
  blockAlerts: false,
  
  // 安全设置
  sessionTimeout: 30,
  requirePassword: true,
  logActivity: true
})

// 设置分类
const settingsSections = [
  { id: 'display', name: '显示设置', icon: EyeIcon },
  { id: 'network', name: '网络设置', icon: GlobeAltIcon },
  { id: 'notifications', name: '通知设置', icon: BellIcon },
  { id: 'security', name: '安全设置', icon: ShieldCheckIcon },
  { id: 'about', name: '关于', icon: InformationCircleIcon }
]

// 方法
const saveSettings = () => {
  // 这里应该保存设置到本地存储或后端
  localStorage.setItem('blockchain-browser-settings', JSON.stringify(settings))
  console.log('设置已保存')
}

const resetSettings = () => {
  // 重置为默认设置
  Object.assign(settings, {
    theme: 'light',
    language: 'zh-CN',
    pageSize: 25,
    autoRefresh: true,
    refreshInterval: 30,
    network: 'mainnet',
    customRpc: '',
    apiEndpoint: 'https://api.example.com',
    useWebSocket: false,
    emailNotifications: false,
    emailAddress: '',
    browserNotifications: true,
    priceAlerts: false,
    blockAlerts: false,
    sessionTimeout: 30,
    requirePassword: true,
    logActivity: true
  })
}

const exportSettings = () => {
  const dataStr = JSON.stringify(settings, null, 2)
  const dataBlob = new Blob([dataStr], { type: 'application/json' })
  const url = URL.createObjectURL(dataBlob)
  const link = document.createElement('a')
  link.href = url
  link.download = 'blockchain-browser-settings.json'
  link.click()
  URL.revokeObjectURL(url)
}

const importSettings = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'
  input.onchange = (e) => {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = (e) => {
        try {
          const importedSettings = JSON.parse(e.target?.result as string)
          Object.assign(settings, importedSettings)
          console.log('设置已导入')
        } catch (error) {
          console.error('导入设置失败:', error)
        }
      }
      reader.readAsText(file)
    }
  }
  input.click()
}

const checkUpdate = () => {
  console.log('检查更新...')
  // 这里应该实现检查更新逻辑
}
</script> 