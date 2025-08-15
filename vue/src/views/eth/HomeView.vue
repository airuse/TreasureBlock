<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useWebSocket } from '@/composables/useWebSocket'
import type { Block, Transaction, NetworkStats, WebSocketBlockMessage } from '@/types'
import { 
  formatTimestamp, 
  formatHash, 
  formatNumber, 
  formatDifficulty 
} from '@/utils/formatters'

// 响应式数据
const stats = ref<NetworkStats>({
  totalTransactions: 0,
  totalBlocks: 0,
  activeAddresses: 0,
  networkHashrate: 0,
  dailyVolume: 0,
  avgGasPrice: 0,
  avgBlockTime: 0,
  difficulty: 0
})
const latestBlocks = ref<Block[]>([])
const latestTransactions = ref<Transaction[]>([])

// WebSocket连接
const { subscribe } = useWebSocket()

// 格式化数字
const formatHashrate = (hashrate: number) => {
  if (hashrate >= 1e12) return `${(hashrate / 1e12).toFixed(2)} TH/s`
  if (hashrate >= 1e9) return `${(hashrate / 1e9).toFixed(2)} GH/s`
  if (hashrate >= 1e6) return `${(hashrate / 1e6).toFixed(2)} MH/s`
  if (hashrate >= 1e3) return `${(hashrate / 1e3).toFixed(2)} KH/s`
  return `${hashrate.toFixed(2)} H/s`
}

// 格式化字节大小
const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 格式化金额
const formatAmount = (amount: number) => {
  return amount.toFixed(6)
}

// WebSocket事件处理
const unsubscribeBlocks = subscribe('block', (message: any) => {
  if (message.data) {
    latestBlocks.value.unshift(message.data as Block)
    if (latestBlocks.value.length > 10) {
      latestBlocks.value = latestBlocks.value.slice(0, 10)
    }
  }
})

const unsubscribeTransactions = subscribe('transaction', (message: any) => {
  if (message.data) {
    latestTransactions.value.unshift(message.data as Transaction)
    if (latestTransactions.value.length > 10) {
      latestTransactions.value = latestTransactions.value.slice(0, 10)
    }
  }
})

const unsubscribeStats = subscribe('stats', (message: any) => {
  if (message.data) {
    stats.value = message.data as NetworkStats
  }
})

// 初始化数据
onMounted(() => {
  // 初始化空数据，实际数据将通过WebSocket或API获取
  latestBlocks.value = []
  latestTransactions.value = []
})

onUnmounted(() => {
  // 清理WebSocket订阅
  unsubscribeBlocks()
  unsubscribeTransactions()
  unsubscribeStats()
})
</script>

<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">区块链概览</h1>
      <div class="text-sm text-gray-500">
        最后更新: {{ new Date().toLocaleString() }}
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <div class="card">
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 rounded-lg">
            <CubeIcon class="h-6 w-6 text-blue-600" />
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
          <div class="p-2 bg-orange-100 rounded-lg">
            <ChartBarIcon class="h-6 w-6 text-orange-600" />
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
          <router-link to="/blocks" class="text-blue-600 hover:text-blue-700 text-sm">
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
              <div class="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                <CubeIcon class="h-4 w-4 text-blue-600" />
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
          <router-link to="/transactions" class="text-blue-600 hover:text-blue-700 text-sm">
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
              <p class="text-sm font-medium text-gray-900">0.000000 ETH</p>
              <p class="text-sm text-gray-500">{{ formatAmount(tx.amount || 0) }} ETH</p>
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
          <p class="text-sm text-gray-600">平均Gas价格</p>
          <p class="text-lg font-semibold text-gray-900">{{ (stats.avgGasPrice / 1e9).toFixed(0) }} Gwei</p>
        </div>
        <div class="text-center">
          <p class="text-sm text-gray-600">平均出块时间</p>
          <p class="text-lg font-semibold text-gray-900">{{ stats.avgBlockTime.toFixed(1) }} 秒</p>
        </div>
        <div class="text-center">
          <p class="text-sm text-gray-600">当前难度</p>
          <p class="text-lg font-semibold text-gray-900">{{ formatDifficulty(stats.difficulty) }}</p>
        </div>
        <div class="text-center">
          <p class="text-sm text-gray-600">日交易量</p>
          <p class="text-lg font-semibold text-gray-900">{{ formatAmount(stats.dailyVolume) }} ETH</p>
        </div>
      </div>
    </div>
  </div>
</template>
