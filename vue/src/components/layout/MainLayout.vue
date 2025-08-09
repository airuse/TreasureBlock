<template>
  <div class="min-h-screen bg-gray-50">
    <!-- 导航栏 -->
    <nav class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-1xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex items-center">
            <div class="flex-shrink-0 flex items-center">
              <div class="w-16 h-8 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg flex items-center justify-center mr-3">
                <span class="text-white font-bold text-xs">JDGBCB</span>
              </div>
              <h1 class="text-xl font-semibold text-gray-900">区块链浏览器</h1>
            </div>
          </div>
          
          <div class="flex items-center space-x-4">
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
            
            <!-- 链选择器 -->
            <div class="flex items-center space-x-2">
              <button 
                @click="switchChain('eth')" 
                :class="[
                  'px-3 py-2 rounded-md text-sm font-medium transition-colors flex items-center space-x-2',
                  currentChain === 'eth' 
                    ? 'bg-blue-100 text-blue-700' 
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                ]"
              >
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M11 17a1 1 0 001.447.894l4-2A1 1 0 0017 15V9.236a1 1 0 00-1.447-.894l-4 2a1 1 0 00-.553.894V17zM15.211 6.276a1 1 0 000-1.552l-4.764-3.368a1 1 0 00-1.447 0L4.789 4.724a1 1 0 000 1.552l4.764 3.368a1 1 0 001.447 0l4.764-3.368zM4.447 8.342A1 1 0 003 9.236V15a1 1 0 00.553.894l4 2A1 1 0 009 17v-5.764a1 1 0 00-.553-.894l-4-2z"/>
                </svg>
                <span>ETH</span>
              </button>
              <button 
                @click="switchChain('btc')" 
                :class="[
                  'px-3 py-2 rounded-md text-sm font-medium transition-colors flex items-center space-x-2',
                  currentChain === 'btc' 
                    ? 'bg-orange-100 text-orange-700' 
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                ]"
              >
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-13a1 1 0 10-2 0v.092a4.535 4.535 0 00-1.676.662C6.602 6.234 6 7.009 6 8c0 .99.602 1.765 1.324 2.246.48.32 1.054.545 1.676.662v1.941c-.391-.127-.68-.317-.843-.504a1 1 0 10-1.51 1.31c.562.649 1.413 1.076 2.353 1.253V15a1 1 0 102 0v-.092a4.535 4.535 0 001.676-.662C13.398 13.766 14 12.991 14 12c0-.99-.602-1.765-1.324-2.246A4.535 4.535 0 0011 9.092V7.151c.391.127.68.317.843.504a1 1 0 101.511-1.31c-.563-.649-1.413-1.076-2.354-1.253V5z" clip-rule="evenodd"/>
                </svg>
                <span>BTC</span>
              </button>
            </div>
            

          </div>
        </div>
      </div>
    </nav>

    <div class="flex">
      <!-- 侧边栏 -->
      <aside class="w-64 bg-white shadow-sm border-r border-gray-200 min-h-screen">
        <div class="p-4">
          <nav class="space-y-1">
            <router-link 
              v-for="item in menuItems" 
              :key="item.name"
              :to="item.path"
              class="sidebar-item"
              :class="{ active: $route.path === item.path || ($route.path.startsWith(item.path + '/') && item.name !== '首页') }"
            >
              <svg style="width: 20px; height: 20px; margin-right: 12px;" viewBox="0 0 20 20" fill="currentColor">
                <path v-if="item.name === '首页'" d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z"/>
                <path v-else-if="item.name === '区块'" d="M11 17a1 1 0 001.447.894l4-2A1 1 0 0017 15V9.236a1 1 0 00-1.447-.894l-4 2a1 1 0 00-.553.894V17zM15.211 6.276a1 1 0 000-1.552l-4.764-3.368a1 1 0 00-1.447 0L4.789 4.724a1 1 0 000 1.552l4.764 3.368a1 1 0 001.447 0l4.764-3.368zM4.447 8.342A1 1 0 003 9.236V15a1 1 0 00.553.894l4 2A1 1 0 009 17v-5.764a1 1 0 00-.553-.894l-4-2z"/>
                <path v-else-if="item.name === '交易'" fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-13a1 1 0 10-2 0v.092a4.535 4.535 0 00-1.676.662C6.602 6.234 6 7.009 6 8c0 .99.602 1.765 1.324 2.246.48.32 1.054.545 1.676.662v1.941c-.391-.127-.68-.317-.843-.504a1 1 0 10-1.51 1.31c.562.649 1.413 1.076 2.353 1.253V15a1 1 0 102 0v-.092a4.535 4.535 0 001.676-.662C13.398 13.766 14 12.991 14 12c0-.99-.602-1.765-1.324-2.246A4.535 4.535 0 0011 9.092V7.151c.391.127.68.317.843.504a1 1 0 101.511-1.31c-.563-.649-1.413-1.076-2.354-1.253V5z" clip-rule="evenodd"/>
                <path v-else-if="item.name === '地址'" d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3z"/>
                <path v-else-if="item.name === '统计'" d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zM8 7a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zM14 4a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z"/>
              </svg>
              {{ item.name }}
            </router-link>
          </nav>
        </div>
      </aside>

      <!-- 主内容区域 -->
      <main class="flex-1 p-6">
        <router-view />
      </main>
    </div>


  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useWebSocket } from '@/composables/useWebSocket'

const router = useRouter()

// 响应式数据
const currentChain = ref('eth')

// WebSocket状态
const { isConnected } = useWebSocket()

// 菜单项 - 根据当前链动态生成
const menuItems = computed(() => {
  const basePath = currentChain.value === 'eth' ? '/eth' : '/btc'
  return [
    { name: '首页', path: basePath },
    { name: '区块', path: `${basePath}/blocks` },
    { name: '交易', path: `${basePath}/transactions` },
    { name: '地址', path: `${basePath}/addresses` },
    { name: '统计', path: `${basePath}/statistics` },
]
})

// 链切换方法
const switchChain = (chain: string) => {
  currentChain.value = chain
  
  // 获取当前路径的页面类型
  const currentPath = router.currentRoute.value.path
  const currentPage = currentPath.split('/').pop() || ''
  
  // 构建新路径，保持当前页面类型
  const basePath = chain === 'eth' ? '/eth' : '/btc'
  let newPath = basePath
  
  // 如果不是首页，添加页面路径
  if (currentPage && currentPage !== 'eth' && currentPage !== 'btc') {
    newPath = `${basePath}/${currentPage}`
  }
  
  router.push(newPath)
}


</script> 