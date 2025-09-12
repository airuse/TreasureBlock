<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">比特币概览</h1>
      <div class="text-sm text-gray-500">
        最后更新: {{ new Date().toLocaleString() }}
      </div>
    </div>

    <!-- 关键指标卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <div class="card">
        <div class="flex items-center">
          <div class="p-2 bg-orange-100 rounded-lg">
            <CubeIcon class="h-6 w-6 text-orange-600" />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">总区块数</p>
            <p class="text-2xl font-bold text-gray-900">{{ formatNumber(stats.totalBlocks) }}</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center">
          <div class="p-2 bg-green-100 rounded-lg">
            <CurrencyDollarIcon class="h-6 w-6 text-green-600" />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">总交易数</p>
            <p class="text-2xl font-bold text-gray-900">{{ formatNumber(stats.totalTransactions) }}</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 rounded-lg">
            <UserGroupIcon class="h-6 w-6 text-purple-600" />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">活跃地址</p>
            <p class="text-2xl font-bold text-gray-900">{{ formatNumber(stats.activeAddresses) }}</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 rounded-lg">
            <ChartBarIcon class="h-6 w-6 text-blue-600" />
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">网络算力</p>
            <p class="text-2xl font-bold text-gray-900">{{ formatHashrate(stats.networkHashrate) }}</p>
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
          <router-link to="/btc/blocks" class="text-blue-600 hover:text-blue-700 text-sm">
            查看全部
          </router-link>
        </div>
        <div class="space-y-2">
          <div 
            v-for="block in latestBlocks" 
            :key="block.height"
            class="flex items-center justify-between py-3 border-b border-gray-100 last:border-b-0"
          >
            <div class="flex items-center space-x-3">
              <div class="w-8 h-8 bg-orange-100 rounded-full flex items-center justify-center">
                <CubeIcon class="h-4 w-4 text-orange-600" />
              </div>
              <div>
                <p class="text-sm font-medium text-gray-900">#{{ (block.height || block.number || 0).toLocaleString() }}</p>
                <p class="text-sm text-gray-500">{{ formatTimestamp(typeof block.timestamp === 'string' ? parseInt(block.timestamp) : block.timestamp) }}</p>
              </div>
            </div>
            <div class="text-right">
              <p class="text-sm font-medium text-gray-900">{{ block.transactions_count || block.transactions || 0 }} 交易</p>
              <p class="text-sm text-gray-500">{{ formatBytes(block.size) }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 最新交易 -->
      <div class="card">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-lg font-semibold text-gray-900">最新交易</h2>
          <router-link to="/btc/transactions" class="text-blue-600 hover:text-blue-700 text-sm">
            查看全部
          </router-link>
        </div>
        <div class="space-y-2">
          <div 
            v-for="tx in latestTransactions" 
            :key="tx.hash"
            class="flex items-center justify-between py-3 border-b border-gray-100 last:border-b-0"
          >
            <div class="flex items-center space-x-3">
              <div class="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
                <CurrencyDollarIcon class="h-4 w-4 text-green-600" />
              </div>
              <div>
                <p class="text-sm font-medium text-gray-900">{{ formatHash(tx.hash) }}</p>
                <p class="text-sm text-gray-500">{{ formatTimestamp(typeof tx.timestamp === 'string' ? parseInt(tx.timestamp) : tx.timestamp) }}</p>
              </div>
            </div>
            <div class="text-right">
              <p class="text-sm font-medium text-gray-900">0.000000 BTC</p>
              <p class="text-sm text-gray-500">{{ formatAmount(tx.amount || 0) }} BTC</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 网络状态 -->
    <div class="card">
      <h2 class="text-lg font-semibold text-gray-900 mb-4">网络状态</h2>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="text-center">
          <p class="text-sm text-gray-600">平均手续费</p>
          <p class="text-lg font-semibold text-gray-900">{{ formatFee(stats.avgFee) }} BTC</p>
        </div>
        <div class="text-center">
          <p class="text-sm text-gray-600">平均出块时间</p>
          <p class="text-lg font-semibold text-gray-900">{{ stats.avgBlockTime }} 分钟</p>
        </div>
        <div class="text-center">
          <p class="text-sm text-gray-600">当前难度</p>
          <p class="text-lg font-semibold text-gray-900">{{ formatDifficulty(stats.difficulty) }}</p>
        </div>
        <div class="text-center">
          <p class="text-sm text-gray-600">日交易量</p>
          <p class="text-lg font-semibold text-gray-900">{{ formatAmount(stats.dailyVolume) }} BTC</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { 
  CubeIcon, 
  CurrencyDollarIcon, 
  UserGroupIcon, 
  ChartBarIcon 
} from '@heroicons/vue/20/solid'
import { formatNumber, formatTimestamp, formatHash, formatBytes, formatAmount, formatFee, formatHashrate, formatDifficulty } from '@/utils/formatters'
import { useAuthStore } from '@/stores/auth'
import { home, noAuth, blocks as blocksApi, transactions as transactionsApi } from '@/api'
import type { Block, Transaction } from '@/types'

// 响应式数据
const stats = ref({
  totalBlocks: 0,
  totalTransactions: 0,
  activeAddresses: 0,
  networkHashrate: 0,
  avgFee: 0,
  avgBlockTime: 0,
  difficulty: 0,
  dailyVolume: 0
})

// 定义类型
import type { WebSocketBlockMessage, WebSocketTransactionMessage, WebSocketStatsMessage } from '@/types'

const latestBlocks = ref<Block[]>([])
const latestTransactions = ref<Transaction[]>([])

// 认证store
const authStore = useAuthStore()

// 加载数据
const loadData = async () => {
  try {
    // 获取首页统计数据（优先使用新的BTC首页接口）
    let statsResponse
    if (authStore.isAuthenticated) {
      statsResponse = await home.getBtcHomeStats()
    } else {
      statsResponse = await noAuth.getBtcHomeStats()
    }
    if (statsResponse && statsResponse.success === true) {
      const data = (statsResponse as any).data || {}
      const overview = data.overview || data
      stats.value = {
        totalBlocks: overview.totalBlocks || 0,
        totalTransactions: overview.totalTransactions || 0,
        activeAddresses: overview.activeAddresses || 0,
        networkHashrate: overview.networkHashrate || 0,
        avgFee: overview.avgFee || 0,
        avgBlockTime: overview.avgBlockTime || 0,
        difficulty: overview.difficulty || 0,
        dailyVolume: overview.dailyVolume || 0
      }
      // 最新区块/交易（优先使用首页接口返回的数据）
      if (Array.isArray(data.latestBlocks) && data.latestBlocks.length > 0) {
        latestBlocks.value = data.latestBlocks
      }
      if (Array.isArray(data.latestTransactions) && data.latestTransactions.length > 0) {
        latestTransactions.value = data.latestTransactions
      }
    }

    // 获取最新区块（仅在首页接口未返回时回退）
    if (!latestBlocks.value || latestBlocks.value.length === 0) {
      const blocksResponse = await blocksApi.getBlocks({ 
        page: 1, 
        page_size: 5, 
        chain: 'btc' 
      })
      if (blocksResponse && blocksResponse.success === true) {
        latestBlocks.value = blocksResponse.data || []
      }
    }

    // 获取最新交易（仅在首页接口未返回时回退）
    if (!latestTransactions.value || latestTransactions.value.length === 0) {
      const transactionsResponse = await transactionsApi.getTransactions({ 
        page: 1, 
        page_size: 5, 
        chain: 'btc' 
      })
      if (transactionsResponse && transactionsResponse.success === true) {
        latestTransactions.value = transactionsResponse.data || []
      }
    }
  } catch (error) {
    console.error('Failed to load data:', error)
    // 如果API调用失败，使用默认数据
    stats.value = {
      totalBlocks: 850000,
      totalTransactions: 850000000,
      activeAddresses: 500000,
      networkHashrate: 500e12,
      avgFee: 0.00001,
      avgBlockTime: 10,
      difficulty: 50e12,
      dailyVolume: 50000
    }
  }
}

// 定时刷新（每30秒）
const refreshInterval = ref<number | null>(null)

onMounted(() => {
  loadData()
  // 启动自动刷新
  refreshInterval.value = window.setInterval(() => {
    loadData()
  }, 30000)
})

onUnmounted(() => {
  // 清理定时器
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
    refreshInterval.value = null
  }
})
</script> 