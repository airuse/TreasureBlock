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
import { useChainWebSocket } from '@/composables/useWebSocket'
import { stats as statsApi } from '@/api'
import { blocks as blocksApi } from '@/api'
import { transactions as transactionsApi } from '@/api'
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

// WebSocket连接
const { subscribeChainEvent } = useChainWebSocket('btc')

// 加载数据
const loadData = async () => {
  try {
    // 获取统计数据
    const statsResponse = await statsApi.getNetworkStats({ chain: 'btc' })
    if (statsResponse && statsResponse.code === 200) {
      stats.value = {
        totalBlocks: statsResponse.data?.totalBlocks || 0,
        totalTransactions: statsResponse.data?.totalTransactions || 0,
        activeAddresses: statsResponse.data?.activeAddresses || 0,
        networkHashrate: statsResponse.data?.networkHashrate || 0,
        avgFee: statsResponse.data?.avgFee || 0,
        avgBlockTime: statsResponse.data?.avgBlockTime || 0,
        difficulty: statsResponse.data?.difficulty || 0,
        dailyVolume: statsResponse.data?.dailyVolume || 0
      }
    }

    // 获取最新区块
    const blocksResponse = await blocksApi.getBlocks({ 
      page: 1, 
      page_size: 5, 
      chain: 'btc' 
    })
    if (blocksResponse && blocksResponse.code === 200) {
      latestBlocks.value = blocksResponse.data || []
    }

    // 获取最新交易
    const transactionsResponse = await transactionsApi.getTransactions({ 
      page: 1, 
      page_size: 5, 
      chain: 'btc' 
    })
    if (transactionsResponse && transactionsResponse.code === 200) {
      latestTransactions.value = transactionsResponse.data || []
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

// WebSocket事件处理
const unsubscribeBlocks = subscribeChainEvent('block', (message) => {
  console.log('New block received:', message.data)
  // 更新最新区块数据
  if (message.data && message.data.height) {
    const blockData = message.data as unknown as WebSocketBlockMessage
    const newBlock: Block = {
      hash: blockData.hash || `0000000000000000000000000000000000000000000000000000000000000000${blockData.height.toString(16).padStart(8, '0')}`,
      number: blockData.height,
      height: blockData.height,
      timestamp: blockData.timestamp || Math.floor(Date.now() / 1000),
      transactions_count: blockData.transactions || 0,
      transactions: blockData.transactions || 0,
      size: blockData.size || 0,
      chain: 'btc'
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
  console.log('New transaction received:', message.data)
  // 更新最新交易数据
  if (message.data && message.data.hash) {
    const txData = message.data as unknown as WebSocketTransactionMessage
    const newTransaction: Transaction = {
      hash: txData.hash,
      block_hash: txData.blockHash || '0000000000000000000000000000000000000000000000000000000000000000',
      block_number: txData.blockNumber || 0,
      from_address: txData.from || '1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa',
      to_address: txData.to || '1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2',
      value: (txData.amount || 0).toString(),
      gas_price: (txData.fee || 0).toString(),
      gas_used: 1000,
      nonce: 0,
      timestamp: txData.timestamp || Math.floor(Date.now() / 1000),
      chain: 'btc'
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
  console.log('Stats update received:', message.data)
  // 更新统计信息
  if (message.data) {
    const statsData = message.data as unknown as WebSocketStatsMessage
    stats.value = {
      ...stats.value,
      ...statsData
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