<template>
  <div class="space-y-6">
    <!-- 页面标题和返回按钮 -->
    <div class="flex items-center space-x-4">
      <router-link 
        to="/eth/blocks" 
        class="inline-flex items-center px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
        </svg>
        返回区块列表
      </router-link>
      <h1 class="text-2xl font-bold text-gray-900">区块详情 #{{ blockHeight }}</h1>
    </div>

    <!-- 加载状态 -->
    <div v-if="isLoading" class="card">
      <div class="text-center py-8">
        <div class="inline-flex items-center px-4 py-2 text-sm text-gray-600">
          <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          加载区块信息中...
        </div>
      </div>
    </div>

    <!-- 区块信息 -->
    <div v-else-if="block" class="space-y-6">
      <!-- 区块基本信息 -->
      <div class="card">
        <h2 class="text-lg font-medium text-gray-900 mb-4">区块信息</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-500">区块高度</label>
            <p class="mt-1 text-sm text-gray-900">#{{ block.height?.toLocaleString() }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">时间戳</label>
            <p class="mt-1 text-sm text-gray-900">{{ formatTimestamp(block.timestamp) }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">交易数量</label>
            <p class="mt-1 text-sm text-gray-900">{{ block.transactions?.toLocaleString() }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">区块大小</label>
            <p class="mt-1 text-sm text-gray-900">{{ formatBytes(block.size) }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">Gas使用</label>
            <p class="mt-1 text-sm text-gray-900">{{ formatGas(block.gasUsed, block.gasLimit) }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">矿工地址</label>
            <p class="mt-1 text-sm text-gray-900 font-mono">{{ formatAddress(block.miner) }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">区块奖励</label>
            <p class="mt-1 text-sm text-gray-900">{{ formatAmount(block.reward) }} ETH</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">难度</label>
            <p class="mt-1 text-sm text-gray-900">{{ block.difficulty?.toLocaleString() || 'N/A' }}</p>
          </div>
        </div>
      </div>

      <!-- 交易列表 -->
      <div class="card">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-lg font-medium text-gray-900">交易列表</h2>
          <div class="text-sm text-gray-500">
            共 {{ transactions.length }} 笔交易
          </div>
        </div>

        <!-- 交易范围说明 -->
        <div v-if="transactions.length > 0" class="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-md">
          <div class="flex items-center">
            <svg class="w-5 h-5 text-blue-600 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <span class="text-sm text-blue-800">
              <span v-if="isFilteredByBlock">显示区块 #{{ blockHeight }} 的交易</span>
              <span v-else>显示所有交易（后端暂不支持按区块筛选）</span>
            </span>
          </div>
        </div>

        <!-- 交易加载状态 -->
        <div v-if="loadingTransactions" class="text-center py-8">
          <div class="inline-flex items-center px-4 py-2 text-sm text-gray-600">
            <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            加载交易中...
          </div>
        </div>

        <!-- 交易列表 -->
        <div v-else-if="transactions.length > 0" class="space-y-3">
          <div v-for="tx in transactions" :key="tx.id" class="bg-gray-50 p-4 rounded-lg">
            <div class="flex items-center justify-between">
              <div class="flex-1">
                <div class="flex items-center space-x-4 mb-2">
                  <span class="font-mono text-sm text-gray-600">{{ formatHash(tx.tx_id || tx.hash) }}</span>
                  <span class="text-sm text-gray-500">{{ formatTimestamp(tx.ctime || tx.timestamp) }}</span>
                </div>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm">
                  <div>
                    <span class="text-gray-500">从: </span>
                    <span class="font-mono text-blue-600">{{ formatAddress(tx.address_from || tx.from) }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">到: </span>
                    <span class="font-mono text-blue-600">{{ formatAddress(tx.address_to || tx.to) }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">金额: </span>
                    <span class="font-medium">{{ formatAmount(tx.amount || tx.value) }} ETH</span>
                  </div>
                  <div>
                    <span class="text-gray-500">Gas: </span>
                    <span class="text-gray-600">{{ tx.gas_used?.toLocaleString() || tx.gasUsed?.toLocaleString() || 'N/A' }}</span>
                  </div>
                </div>
              </div>
              <span :class="getStatusClass(tx.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                {{ getStatusText(tx.status) }}
              </span>
            </div>
          </div>
        </div>

        <!-- 无交易状态 -->
        <div v-else class="text-center py-8 text-gray-500">
          该区块暂无交易
        </div>
      </div>
    </div>

    <!-- 错误状态 -->
    <div v-else class="card">
      <div class="text-center py-8">
        <div class="text-red-600 mb-2">
          <svg class="w-12 h-12 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"></path>
          </svg>
        </div>
        <h3 class="text-lg font-medium text-gray-900 mb-2">加载失败</h3>
        <p class="text-gray-500 mb-4">{{ errorMessage || '无法加载区块信息' }}</p>
        <button 
          @click="loadBlockData" 
          class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
        >
          重试
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { blocks as blocksApi } from '@/api'
import { transactions as transactionsApi } from '@/api'

// 路由参数
const route = useRoute()
const blockHeight = computed(() => route.params.height as string)

// 认证store
const authStore = useAuthStore()

// 响应式数据
const block = ref<any>(null)
const transactions = ref<any[]>([])
const isLoading = ref(true)
const loadingTransactions = ref(true)
const errorMessage = ref('')

// 计算属性
const isFilteredByBlock = computed(() => {
  // 检查交易是否按区块筛选
  if (transactions.value.length === 0) return false
  
  // 如果第一个交易有区块高度字段，说明是按区块筛选的
  const firstTx = transactions.value[0]
  return !!(firstTx.blockHeight || firstTx.block_number || firstTx.block_height)
})

// 格式化函数
const formatTimestamp = (timestamp: string | number) => {
  if (!timestamp) return 'N/A'
  
  let date: Date
  if (typeof timestamp === 'string') {
    // 处理ISO格式字符串
    date = new Date(timestamp)
  } else {
    // 处理Unix时间戳
    date = new Date(timestamp * 1000)
  }
  
  // 检查日期是否有效
  if (isNaN(date.getTime())) {
    return 'Invalid Date'
  }
  
  return date.toLocaleString()
}

const formatAddress = (address: string) => {
  if (!address) return 'N/A'
  return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`
}

const formatBytes = (bytes: number) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatGas = (used: number, limit: number) => {
  if (!used || !limit) return 'N/A'
  const percentage = ((used / limit) * 100).toFixed(1)
  return `${used.toLocaleString()} / ${limit.toLocaleString()} (${percentage}%)`
}

const formatAmount = (amount: number) => {
  if (!amount) return '0'
  return (amount / 1e18).toFixed(6)
}

const formatHash = (hash: string) => {
  if (!hash) return 'N/A'
  return `${hash.substring(0, 10)}...${hash.substring(hash.length - 10)}`
}

const getStatusClass = (status: number) => {
  switch (status) {
    case 0:
      return 'bg-gray-100 text-gray-800'
    case 1:
      return 'bg-green-100 text-green-800'
    case 2:
      return 'bg-red-100 text-red-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

const getStatusText = (status: number) => {
  switch (status) {
    case 0:
      return 'Pending'
    case 1:
      return 'Success'
    case 2:
      return 'Failed'
    default:
      return 'Unknown'
  }
}

// 加载区块数据
const loadBlockData = async () => {
  try {
    isLoading.value = true
    errorMessage.value = ''
    
    // 根据登录状态调用不同的API
    if (authStore.isAuthenticated) {
      // 已登录用户：调用 /v1/ 下的API
      console.log('🔐 已登录用户，调用 /v1/ API 获取区块详情')
      const response = await blocksApi.getBlock({ 
        height: parseInt(blockHeight.value), 
        chain: 'eth' 
      })
      
      if (response && response.success === true) {
        console.log('📊 后端返回区块数据:', response.data)
        block.value = response.data
      } else {
        throw new Error(response?.message || '获取区块信息失败')
      }
    } else {
      // 未登录用户：调用 /no-auth/ 下的API（有限制）
      console.log('👤 未登录用户，调用 /no-auth/ API 获取区块详情（有限制）')
      const response = await blocksApi.getBlockPublic({ 
        height: parseInt(blockHeight.value), 
        chain: 'eth' 
      })
      
      if (response && response.success === true) {
        console.log('📊 后端返回区块数据:', response.data)
        block.value = response.data
      } else {
        throw new Error(response?.message || '获取区块信息失败')
      }
    }
  } catch (error) {
    console.error('Failed to load block:', error)
    errorMessage.value = error instanceof Error ? error.message : '加载区块信息失败'
  } finally {
    isLoading.value = false
  }
}

// 加载交易数据
const loadTransactions = async () => {
  try {
    loadingTransactions.value = true
    
    // 根据登录状态调用不同的API
    if (authStore.isAuthenticated) {
      // 已登录用户：调用 /v1/ 下的API
      console.log('🔐 已登录用户，调用 /v1/ API 获取区块交易')
      const response = await blocksApi.getBlockTransactions({
        height: parseInt(blockHeight.value),
        chain: 'eth',
        page: 1,
        page_size: 100
      })
      
      if (response && response.success === true) {
        console.log('📊 后端返回交易数据:', response.data)
        
        // 新API直接返回交易数据，不需要过滤
        const responseData = response.data as any
        if (responseData?.transactions && Array.isArray(responseData.transactions)) {
          transactions.value = responseData.transactions
          console.log('✅ 成功加载区块交易:', transactions.value.length, '笔交易')
        } else {
          console.warn('API返回数据格式异常:', responseData)
          transactions.value = []
        }
      } else {
        throw new Error(response?.message || '获取交易信息失败')
      }
    } else {
      // 未登录用户：调用 /no-auth/ 下的API（有限制）
      console.log('👤 未登录用户，调用 /no-auth/ API 获取区块交易（有限制）')
      const response = await blocksApi.getBlockTransactionsPublic({
        height: parseInt(blockHeight.value),
        chain: 'eth',
        page: 1,
        page_size: 50
      })
      
      if (response && response.success === true) {
        console.log('📊 后端返回交易数据:', response.data)
        
        // 新API直接返回交易数据，不需要过滤
        const responseData = response.data as any
        if (responseData?.transactions && Array.isArray(responseData.transactions)) {
          transactions.value = responseData.transactions
          console.log('✅ 成功加载区块交易:', transactions.value.length, '笔交易')
        } else {
          console.warn('API返回数据格式异常:', responseData)
          transactions.value = []
        }
      } else {
        throw new Error(response?.message || '获取交易信息失败')
      }
    }
  } catch (error) {
    console.error('Failed to load transactions:', error)
    transactions.value = []
  } finally {
    loadingTransactions.value = false
  }
}

// 监听路由参数变化
onMounted(async () => {
  await loadBlockData()
  if (block.value) {
    await loadTransactions()
  }
})
</script>

<style scoped>
.card {
  @apply bg-white shadow-sm rounded-lg border border-gray-200 p-6;
}
</style>
