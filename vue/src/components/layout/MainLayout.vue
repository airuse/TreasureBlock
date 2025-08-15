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

            <!-- 认证状态 -->
            <div class="flex items-center space-x-3">
              <!-- 未登录状态 -->
              <div v-if="!authStore.isAuthenticated" class="flex items-center space-x-2">
                <span class="text-sm text-gray-600">游客模式</span>
                <button
                  @click="showLoginModal = true"
                  class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 transition-colors"
                >
                  登录
                </button>
              </div>

              <!-- 已登录状态 -->
              <div v-else class="flex items-center space-x-3">
                <!-- 用户信息 -->
                <div class="flex items-center space-x-2">
                  <div class="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                    <span class="text-blue-600 font-medium text-sm">
                      {{ authStore.user?.username?.charAt(0)?.toUpperCase() || 'U' }}
                    </span>
                  </div>
                  <div class="text-sm">
                    <div class="font-medium text-gray-900">{{ authStore.user?.username }}</div>
                    <div class="text-gray-500">{{ authStore.user?.email }}</div>
                  </div>
                </div>

                <!-- 用户菜单 -->
                <div class="relative">
                  <button
                    @click="showUserMenu = !showUserMenu"
                    class="p-2 text-gray-400 hover:text-gray-600 transition-colors"
                  >
                    <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
                    </svg>
                  </button>

                  <!-- 下拉菜单 -->
                  <div
                    v-if="showUserMenu"
                    class="absolute right-0 mt-2 w-56 bg-white rounded-md shadow-lg py-1 z-50 border border-gray-200"
                  >
                    <!-- 个人中心 -->
                    <button
                      @click="showProfileModal = true; showUserMenu = false"
                      class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 flex items-center"
                    >
                      <svg class="w-4 h-4 mr-3 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                      </svg>
                      个人中心
                    </button>
                    
                    <!-- 地址管理 -->
                    <button
                      @click="showAddressModal = true; showUserMenu = false"
                      class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 flex items-center"
                    >
                      <svg class="w-4 h-4 mr-3 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                      </svg>
                      地址管理
                    </button>
                    
                    <!-- API密钥管理 -->
                    <button
                      @click="showAPIKeyModal = true; showUserMenu = false"
                      class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 flex items-center"
                    >
                      <svg class="w-4 h-4 mr-3 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                      </svg>
                      API密钥管理
                    </button>
                    
                    <div class="border-t border-gray-200 my-1"></div>
                    
                    <!-- 退出登录 -->
                    <button
                      @click="handleLogout"
                      class="block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 flex items-center"
                    >
                      <svg class="w-4 h-4 mr-3 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                      </svg>
                      退出登录
                    </button>
                  </div>
                </div>
              </div>
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

            <!-- 个人中心菜单（仅登录用户可见） -->
            <div v-if="authStore.isAuthenticated" class="pt-4 border-t border-gray-200">
              <div class="px-3 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wider">
                个人中心
              </div>
              <router-link 
                v-for="item in personalMenuItems" 
                :key="item.name"
                :to="item.path"
                class="sidebar-item"
                :class="{ active: $route.path.startsWith(item.path) }"
              >
                <svg style="width: 20px; height: 20px; margin-right: 12px;" viewBox="0 0 20 20" fill="currentColor">
                  <path v-if="item.name === '扫块收益'" fill-rule="evenodd" d="M4 4a2 2 0 00-2 2v4a2 2 0 002 2V6h10a2 2 0 00-2-2H4zm2 6a2 2 0 012-2h8a2 2 0 012 2v4a2 2 0 01-2 2H8a2 2 0 01-2-2v-4zm6 4a2 2 0 100 4 2 2 0 000-4z" clip-rule="evenodd"/>
                  <path v-else-if="item.name === '个人地址'" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"/>
                  <path v-else-if="item.name === '交易历史'" fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd"/>
                </svg>
                {{ item.name }}
              </router-link>
            </div>
          </nav>
        </div>
      </aside>

      <!-- 主内容区域 -->
      <main class="flex-1 p-6">
        <router-view />
      </main>
      <!-- 登录模态框 -->
      <LoginModal 
        :isVisible="showLoginModal"
        @close="showLoginModal = false"
        @success="showLoginModal = false"
      />
      
      <!-- 个人中心模态框 -->
      <ProfileModal 
        :isVisible="showProfileModal"
        @close="showProfileModal = false"
      />
      
      <!-- API密钥管理模态框 -->
      <APIKeyModal 
        :isVisible="showAPIKeyModal"
        @close="showAPIKeyModal = false"
      />
      
      <!-- 地址管理模态框 -->
      <AddressModal 
        :isVisible="showAddressModal"
        @close="showAddressModal = false"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useWebSocket } from '@/composables/useWebSocket'
import { useAuthStore } from '@/stores/auth'
import LoginModal from '../auth/LoginModal.vue'
import ProfileModal from '../auth/ProfileModal.vue'
import APIKeyModal from '../auth/APIKeyModal.vue'
import AddressModal from '../auth/AddressModal.vue'

const router = useRouter()

// 认证
const authStore = useAuthStore()

// 初始化认证状态
onMounted(() => {
  authStore.initialize()
})

const showLoginModal = ref(false)
const showUserMenu = ref(false)
const showAPIKeyModal = ref(false)
const showProfileModal = ref(false)
const showAddressModal = ref(false) // Added for address management

const handleLogout = () => {
  authStore.logout()
  showUserMenu.value = false
}

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

// 个人中心菜单项
const personalMenuItems = computed(() => {
  const basePath = currentChain.value === 'eth' ? '/eth' : '/btc'
  return [
    { name: '扫块收益', path: `${basePath}/personal/earnings` },
    { name: '个人地址', path: `${basePath}/personal/addresses` },
    { name: '交易历史', path: `${basePath}/personal/transactions` },
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

<style scoped>
.sidebar-item {
  @apply flex items-center px-3 py-2 text-sm font-medium text-gray-600 rounded-md hover:text-gray-900 hover:bg-gray-100 transition-colors;
}

.sidebar-item.active {
  @apply bg-blue-50 text-blue-700;
}

.sidebar-item.active:hover {
  @apply bg-blue-100;
}
</style>
