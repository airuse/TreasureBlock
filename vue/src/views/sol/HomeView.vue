<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import type { HomeOverview, HomeBlockSummary, HomeTransactionSummary } from '@/types'
import { 
  formatTimestamp, 
  formatHash, 
  formatFullHash,
  formatNumber, 
  formatDifficulty 
} from '@/utils/formatters'

// 响应式数据
const stats = ref<HomeOverview>({
  totalBlocks: 0,
  totalTransactions: 0,
  baseFee: 0,
  dailyVolume: 0,
  avgGasPrice: 0,
  avgBlockTime: 0
})
const latestBlocks = ref<HomeBlockSummary[]>([])
const latestTransactions = ref<HomeTransactionSummary[]>([])
const loading = ref(false)
const error = ref('')

// 移除WebSocket相关代码

// 格式化字节大小
const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 格式化金额
const formatAmount = (amount: number | string) => {
  // 如果amount是字符串，先转换为数字
  const numAmount = typeof amount === 'string' ? parseFloat(amount) : amount
  // 检查是否为有效数字
  if (isNaN(numAmount)) {
    return '0.000000'
  }
  return numAmount.toFixed(6)
}

// 格式化Gas价格
const formatGasPrice = (gasPrice: number | string | undefined) => {
  if (gasPrice === undefined || gasPrice === null) {
    return '0'
  }
  // 如果gasPrice是字符串，先转换为数字
  const numGasPrice = typeof gasPrice === 'string' ? parseFloat(gasPrice) : gasPrice
  // 检查是否为有效数字
  if (isNaN(numGasPrice)) {
    return '0'
  }
  // 转换为Gwei (1 Gwei = 10^9 Wei)
  return (numGasPrice / 1e9).toFixed(2)
}

// 格式化大数值，处理科学计数法
const formatLargeNumber = (value: number | string | undefined) => {
  if (value === undefined || value === null) {
    return '0'
  }
  // 如果value是字符串，先转换为数字
  const numValue = typeof value === 'string' ? parseFloat(value) : value
  // 检查是否为有效数字
  if (isNaN(numValue)) {
    return '0'
  }
  
  // 处理大数值
  if (numValue >= 1e9) {
    return (numValue / 1e9).toFixed(2) + ' SOL'
  } else if (numValue >= 1e6) {
    return (numValue / 1e6).toFixed(2) + ' MSOL'
  } else if (numValue >= 1e3) {
    return (numValue / 1e3).toFixed(2) + ' KSOL'
  } else {
    return numValue.toFixed(6) + ' SOL'
  }
}

// 格式化lamports到SOL
const formatLamportsToSol = (lamports: number | string | undefined) => {
  if (lamports === undefined || lamports === null) {
    return '0.000000'
  }
  const numLamports = typeof lamports === 'string' ? parseFloat(lamports) : lamports
  if (isNaN(numLamports)) {
    return '0.000000'
  }
  return (numLamports / 1e9).toFixed(6)
}

// 认证store
const authStore = useAuthStore()

// 加载首页数据
const loadHomeData = async () => {
  loading.value = true
  error.value = ''
  
  try {
    console.log('🔍 开始加载首页数据...')
    
    // 根据登录状态调用不同的API
    let response
    if (authStore.isAuthenticated) {
      // 已登录用户：调用 /v1/ 下的API
      const { home } = await import('@/api')
      response = await home.getHomeStats({ chain: 'sol' })
    } else {
      // 未登录用户：调用 /no-auth/ 下的API
      const { noAuth } = await import('@/api')
      response = await noAuth.getHomeStats({ chain: 'sol' })
    }
    
    console.log('📡 API响应:', response)
    
    if (response.success && response.data) {
      console.log('✅ 数据加载成功，开始处理数据...')
      
      // 处理不同的响应数据结构
      if ('overview' in response.data) {
        // 标准响应结构：{ overview, latestBlocks, latestTransactions }
        stats.value = response.data.overview || {}
        latestBlocks.value = response.data.latestBlocks || []
        latestTransactions.value = response.data.latestTransactions || []
      } else {
        // 直接返回HomeOverview结构
        stats.value = response.data as HomeOverview
        latestBlocks.value = []
        latestTransactions.value = []
      }
      
      console.log('📊 处理后的数据:', {
        stats: stats.value,
        blocks: latestBlocks.value,
        transactions: latestTransactions.value
      })
    } else {
      console.warn('⚠️ API返回失败:', response)
      error.value = response.message || '获取数据失败'
    }
  } catch (err) {
    console.error('❌ 加载首页数据失败:', err)
    error.value = '网络错误，请稍后重试'
  } finally {
    loading.value = false
  }
}

// 移除WebSocket事件处理，改为定时刷新数据
const refreshInterval = ref<number | null>(null)

// 定时刷新数据（每30秒刷新一次）
const startAutoRefresh = () => {
  refreshInterval.value = window.setInterval(() => {
    loadHomeData()
  }, 30000) // 30秒
}

// 停止自动刷新
const stopAutoRefresh = () => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
    refreshInterval.value = null
  }
}

// 初始化数据
onMounted(() => {
  loadHomeData()
  startAutoRefresh() // 启动自动刷新
})

// 页面卸载时清理定时器
onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">Solana 区块链概览</h1>
      <div class="flex items-center space-x-4">
        <button 
          @click="loadHomeData" 
          :disabled="loading"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
        >
          {{ loading ? '加载中...' : '刷新数据' }}
        </button>
        <div class="text-sm text-gray-500">
          最后更新: {{ new Date().toLocaleString() }}
        </div>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
      <strong class="font-bold">错误:</strong>
      <span class="block sm:inline">{{ error }}</span>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div class="card">
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 rounded-lg">
            <svg class="h-6 w-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"></path>
            </svg>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">总区块数</p>
            <p class="text-2xl font-bold text-gray-900">{{ formatNumber(stats.totalBlocks || 0) }}</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center">
          <div class="p-2 bg-green-100 rounded-lg">
            <svg class="h-6 w-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1"></path>
            </svg>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">总交易数</p>
            <p class="text-2xl font-bold text-gray-900">{{ formatNumber(stats.totalTransactions || 0) }}</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 rounded-lg">
            <svg class="h-6 w-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
            </svg>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">最新区块Base Fee</p>
            <p class="text-2xl font-bold text-gray-900">{{ formatGasPrice(stats.baseFee) }} Gwei</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 最新区块和交易 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 最新区块 -->
      <div class="card">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-lg font-semibold text-gray-900">最新区块</h2>
          <router-link to="sol/blocks" class="text-blue-600 hover:text-blue-700 text-sm">
            查看全部
          </router-link>
        </div>
        <div class="space-y-2">
          <div 
            v-for="block in (latestBlocks || [])" 
            :key="block.height"
            class="flex items-center justify-between py-3 border-b border-gray-100 last:border-b-0"
          >
            <div class="flex items-center space-x-3">
              <div class="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                <svg class="h-4 w-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"></path>
                </svg>
              </div>
              <div>
                <p class="text-sm font-medium text-gray-900">#{{ (block.height || 0).toLocaleString() }}</p>
                <p class="text-sm text-gray-500">{{ formatTimestamp(block.timestamp) }}</p>
              </div>
            </div>
            <div class="text-right">
              <p class="text-sm font-medium text-gray-900">{{ block.transactions_count || 0 }} 交易</p>
              <p class="text-sm text-gray-500">{{ formatBytes(block.size || 0) }}</p>
            </div>
          </div>
          <!-- 空数据提示 -->
          <div v-if="!latestBlocks || latestBlocks.length === 0" class="text-center py-8 text-gray-500">
            暂无区块数据
          </div>
        </div>
      </div>

      <!-- 最新交易 -->
      <div class="card">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-lg font-semibold text-gray-900">最新交易</h2>
          <router-link to="sol/transactions" class="text-blue-600 hover:text-blue-700 text-sm">
            查看全部
          </router-link>
        </div>
        <div class="space-y-2">
          <div 
            v-for="tx in (latestTransactions || [])" 
            :key="tx.hash"
            class="flex items-center justify-between py-3 border-b border-gray-100 last:border-b-0"
          >
            <div class="flex items-center space-x-3">
              <div class="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
                <svg class="h-4 w-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1"></path>
                </svg>
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-gray-900 truncate">{{ formatFullHash(tx.hash) }}</p>
                <p class="text-sm text-gray-500">
                  {{ formatTimestamp(tx.timestamp) }}
                  <router-link 
                    v-if="tx.height" 
                    :to="`/sol/blocks/${tx.height}`" 
                    class="ml-2 text-blue-600 hover:text-blue-800 underline"
                  >
                    #{{ tx.height.toLocaleString() }}
                  </router-link>
                </p>
              </div>
            </div>
            <div class="text-right ml-4">
              <p class="text-sm font-medium text-gray-900">{{ formatGasPrice(tx.gas_price) }} Gwei</p>
              <p class="text-sm text-gray-500">{{ formatLamportsToSol(tx.amount || 0) }} SOL</p>
            </div>
          </div>
          <!-- 空数据提示 -->
          <div v-if="!latestTransactions || latestTransactions.length === 0" class="text-center py-8 text-gray-500">
            暂无交易数据
          </div>
        </div>
      </div>
    </div>

    <!-- 网络状态 -->
    <div class="card">
      <h2 class="text-lg font-semibold text-gray-900 mb-4">网络状态</h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="text-center">
          <p class="text-sm text-gray-600">十分钟内平均Gas价格</p>
          <p class="text-lg font-semibold text-gray-900">{{ formatGasPrice(stats.avgGasPrice) }} Gwei</p>
        </div>
        <div class="text-center">
          <p class="text-sm text-gray-600">十分钟内平均出块时间</p>
          <p class="text-lg font-semibold text-gray-900">{{ (stats.avgBlockTime || 0).toFixed(1) }} 秒</p>
        </div>
        <div class="text-center">
          <p class="text-sm text-gray-600">十分钟内交易量</p>
          <p class="text-lg font-semibold text-gray-900">{{ formatLamportsToSol(stats.dailyVolume) }} SOL</p>
        </div>
      </div>
    </div>
  </div>
</template>
