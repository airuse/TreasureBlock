<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { 
  CubeIcon, 
  CurrencyDollarIcon, 
  UserGroupIcon, 
  ChartBarIcon 
} from '@heroicons/vue/20/solid'
import type { Block, Transaction, NetworkStats, WebSocketBlockMessage, WebSocketTransactionMessage, WebSocketStatsMessage } from '@/types'
import { mockData } from '@/api'
import { 
  formatTimestamp, 
  formatHash, 
  formatBytes, 
  formatHashrate, 
  formatAmount, 
  formatDifficulty 
} from '@/utils/formatters'
import { useChainWebSocket } from '@/composables/useWebSocket'

// 响应式数据
const stats = ref<NetworkStats>(mockData.getMockNetworkStats())
const latestBlocks = ref<Block[]>([])
const latestTransactions = ref<Transaction[]>([])

// WebSocket连接
const { subscribeChainEvent } = useChainWebSocket('eth')

// 格式化数字
const formatNumber = (num: number): string => {
  return num.toLocaleString()
}

// WebSocket事件处理
const unsubscribeBlocks = subscribeChainEvent('block', (message) => {
  console.log('New ETH block received:', message.data)
  // 更新最新区块数据
  if (message.data && message.data.height) {
    const blockData = message.data as unknown as WebSocketBlockMessage
    const newBlock: Block = {
      hash: blockData.hash || `0x${blockData.height.toString(16).padStart(64, '0')}`,
      number: blockData.height,
      height: blockData.height,
      timestamp: blockData.timestamp || Math.floor(Date.now() / 1000),
      transactions_count: blockData.transactions || 0,
      transactions: blockData.transactions || 0,
      size: blockData.size || 0,
      chain: 'eth',
      gasUsed: blockData.gasUsed || 0,
      gasLimit: blockData.gasLimit || 0,
      miner: blockData.miner || '',
      reward: blockData.reward || 0,
      parentHash: blockData.parentHash || '',
      nonce: blockData.nonce || '',
      difficulty: blockData.difficulty || 0
    }
    
    // 将新区块添加到列表开头
    latestBlocks.value.unshift(newBlock)
    // 保持最多5个区块
    if (latestBlocks.value.length > 5) {
      latestBlocks.value = latestBlocks.value.slice(0, 5)
    }
    
    // 更新统计信息
    stats.value.totalBlocks = blockData.totalBlocks || stats.value.totalBlocks
  }
})

const unsubscribeTransactions = subscribeChainEvent('transaction', (message) => {
  console.log('New ETH transaction received:', message.data)
  // 更新最新交易数据
  if (message.data && message.data.hash) {
    const txData = message.data as unknown as WebSocketTransactionMessage
    const newTransaction: Transaction = {
      hash: txData.hash,
      blockHeight: txData.blockHeight || 0,
      timestamp: txData.timestamp || Math.floor(Date.now() / 1000),
      from: txData.from || '',
      to: txData.to || '',
      amount: txData.amount || 0,
      gasUsed: txData.gasUsed || 0,
      gas_price: (txData.gasPrice || 0).toString(),
      gas_used: txData.gasUsed || 0,
      nonce: txData.nonce || 0,
      input: txData.input || '',
      block_hash: '',
      block_number: 0,
      from_address: txData.from || '',
      to_address: txData.to || '',
      value: (txData.amount || 0).toString(),
      chain: 'eth',
      status: (txData.status === 'success' || txData.status === 'failed' || txData.status === 'pending') ? txData.status : 'success'
    }
    
    // 将新交易添加到列表开头
    latestTransactions.value.unshift(newTransaction)
    // 保持最多5个交易
    if (latestTransactions.value.length > 5) {
      latestTransactions.value = latestTransactions.value.slice(0, 5)
    }
    
    // 更新统计信息
    stats.value.totalTransactions = txData.totalTransactions || stats.value.totalTransactions
  }
})

const unsubscribeStats = subscribeChainEvent('stats', (message) => {
  console.log('ETH stats update received:', message.data)
  // 更新统计信息
  if (message.data) {
    const statsData = message.data as unknown as WebSocketStatsMessage
    stats.value = {
      ...stats.value,
      ...statsData
    }
  }
})

// 初始化数据
onMounted(() => {
  latestBlocks.value = mockData.getMockBlocks(5)
  latestTransactions.value = mockData.getMockTransactions(5)
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
