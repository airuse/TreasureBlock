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
                <p class="text-sm font-medium text-gray-900">#{{ block.height.toLocaleString() }}</p>
                <p class="text-sm text-gray-500">{{ formatTimestamp(block.timestamp) }}</p>
              </div>
            </div>
            <div class="text-right">
              <p class="text-sm font-medium text-gray-900">{{ block.transactions }} 交易</p>
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
                <p class="text-sm text-gray-500">{{ formatTimestamp(tx.timestamp) }}</p>
              </div>
            </div>
            <div class="text-right">
              <p class="text-sm font-medium text-gray-900">0.000000 BTC</p>
              <p class="text-sm text-gray-500">{{ formatAmount(tx.amount) }} BTC</p>
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
import { useChainWebSocket } from '@/composables/useWebSocket'

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
interface Block {
  height: number
  timestamp: number
  transactions: number
  size: number
}

interface Transaction {
  hash: string
  timestamp: number
  amount: number
  fee: number
}

const latestBlocks = ref<Block[]>([])
const latestTransactions = ref<Transaction[]>([])

// WebSocket连接
const { subscribeChainEvent, subscribeChainNotification } = useChainWebSocket('btc')

// 加载数据
const loadData = () => {
  // 模拟BTC数据
  stats.value = {
    totalBlocks: 850000,
    totalTransactions: 850000000,
    activeAddresses: 500000,
    networkHashrate: 500e12, // 500 TH/s
    avgFee: 0.00001, // 0.00001 BTC
    avgBlockTime: 10, // 10分钟
    difficulty: 50e12, // 50T
    dailyVolume: 50000 // 50000 BTC
  }

  // 模拟最新区块
  latestBlocks.value = Array.from({ length: 5 }, (_, i) => ({
    height: 850000 - i,
    timestamp: Math.floor(Date.now() / 1000) - i * 600,
    transactions: Math.floor(Math.random() * 2000) + 500,
    size: Math.floor(Math.random() * 1000000) + 500000
  }))

  // 模拟最新交易
  latestTransactions.value = Array.from({ length: 5 }, (_, i) => ({
    hash: `0x${Math.random().toString(16).substring(2, 34)}`,
    timestamp: Math.floor(Date.now() / 1000) - i * 60,
    amount: Math.random() * 10,
    fee: Math.random() * 0.001
  }))
}

// WebSocket事件处理
const unsubscribeBlocks = subscribeChainEvent('block', (message) => {
  console.log('New block received:', message.data)
  // 更新最新区块数据
  if (message.data && message.data.height) {
    const newBlock: Block = {
      height: message.data.height,
      timestamp: message.data.timestamp || Math.floor(Date.now() / 1000),
      transactions: message.data.transactions || 0,
      size: message.data.size || 0
    }
    
    // 将新区块添加到列表开头
    latestBlocks.value.unshift(newBlock)
    // 保持最多5个区块
    if (latestBlocks.value.length > 5) {
      latestBlocks.value = latestBlocks.value.slice(0, 5)
    }
    
    // 更新统计信息
    stats.value.totalBlocks = message.data.totalBlocks || stats.value.totalBlocks
  }
})

const unsubscribeTransactions = subscribeChainEvent('transaction', (message) => {
  console.log('New transaction received:', message.data)
  // 更新最新交易数据
  if (message.data && message.data.hash) {
    const newTransaction: Transaction = {
      hash: message.data.hash,
      timestamp: message.data.timestamp || Math.floor(Date.now() / 1000),
      amount: message.data.amount || 0,
      fee: message.data.fee || 0
    }
    
    // 将新交易添加到列表开头
    latestTransactions.value.unshift(newTransaction)
    // 保持最多5个交易
    if (latestTransactions.value.length > 5) {
      latestTransactions.value = latestTransactions.value.slice(0, 5)
    }
    
    // 更新统计信息
    stats.value.totalTransactions = message.data.totalTransactions || stats.value.totalTransactions
  }
})

const unsubscribeStats = subscribeChainEvent('stats', (message) => {
  console.log('Stats update received:', message.data)
  // 更新统计信息
  if (message.data) {
    stats.value = {
      ...stats.value,
      ...message.data
    }
  }
})

onMounted(() => {
  loadData()
})

onUnmounted(() => {
  // 清理WebSocket订阅
  unsubscribeBlocks()
  unsubscribeTransactions()
  unsubscribeStats()
})
</script> 