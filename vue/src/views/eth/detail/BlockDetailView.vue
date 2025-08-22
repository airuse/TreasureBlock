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

    <!-- 全局轻提示：复制成功（跟随点击位置） -->
    <div v-if="showToast" class="fixed z-50 bg-gray-900 text-white text-sm px-3 py-2 rounded shadow pointer-events-none" :style="toastStyle">
      {{ toastMessage || '已复制到剪贴板' }}
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
            <p class="mt-1 text-sm text-gray-900">{{ block.transaction_count?.toLocaleString() || block.transactions?.toLocaleString() || 'N/A' }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">区块大小</label>
            <p class="mt-1 text-sm text-gray-900">{{ formatBytes(block.size || block.stripped_size) }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">Gas使用</label>
            <p class="mt-1 text-sm text-gray-900">{{ formatGas(block.gas_used || block.gasUsed, block.gas_limit || block.gasLimit) }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">矿工地址</label>
            <p class="mt-1 text-sm text-gray-900 font-mono cursor-pointer hover:text-blue-600" @click="copyToClipboard(block.miner || block.miner_address, $event)">
              {{ block.miner || block.miner_address || 'N/A' }}
            </p>
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
            共 {{ totalCount }} 笔交易 (第 {{ currentPage }}/{{ totalPages }} 页)
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
            <!-- 交易基本信息 -->
            <div class="flex items-center justify-between">
              <div class="flex-1">
                <div class="flex items-center space-x-4 mb-2">
                  <span class="font-mono text-sm text-gray-600 cursor-pointer hover:text-blue-600" title="点击复制" @click="copyToClipboard(tx.tx_id || tx.hash, $event)">
                    {{ tx.tx_id || tx.hash || 'N/A' }}
                  </span>
                  <span class="text-sm text-gray-500">{{ formatTimestamp(tx.ctime || tx.timestamp) }}</span>
                </div>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm">
                  <div>
                    <span class="text-gray-500">从: </span>
                    <span class="font-mono text-blue-600 cursor-pointer hover:text-blue-800" title="点击复制" @click="copyToClipboard(tx.address_from || tx.from, $event)">
                      {{ tx.address_from || tx.from || 'N/A' }}
                    </span>
                  </div>
                  <div>
                    <span class="text-gray-500">到: </span>
                    <span class="font-mono text-blue-600 cursor-pointer hover:text-blue-800" title="点击复制" @click="copyToClipboard(tx.address_to || tx.to, $event)">
                      {{ tx.address_to || tx.to || 'N/A' }}
                    </span>
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
              <div class="flex items-center space-x-2">
                <span :class="getStatusClass(tx.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                  {{ getStatusText(tx.status) }}
                </span>
                <button 
                  @click="toggleTransactionExpansion(tx.tx_id || tx.hash)"
                  class="inline-flex items-center px-2 py-1 text-xs font-medium text-gray-600 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
                >
                  <svg v-if="!expandedTransactions[tx.tx_id || tx.hash]" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
                  </svg>
                  <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7"></path>
                  </svg>
                </button>
              </div>
            </div>

            <!-- 展开的交易凭证信息 -->
            <div v-if="expandedTransactions[tx.tx_id || tx.hash]" class="mt-4 pt-4 border-t border-gray-200">
              <!-- 未登录用户提示 -->
              <div v-if="!authStore.isAuthenticated" class="text-center py-4 text-gray-500">
                请登录后查看交易凭证信息
              </div>
              
              <!-- 已登录用户显示凭证信息 -->
              <div v-else-if="loadingReceipts[tx.tx_id || tx.hash]" class="text-center py-4">
                <div class="inline-flex items-center px-4 py-2 text-sm text-gray-600">
                  <svg class="animate-spin -ml-1 mr-3 h-4 w-4 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  加载凭证信息中...
                </div>
              </div>
              
              <div v-else-if="transactionReceipts[tx.tx_id || tx.hash]" class="space-y-4">
                <h4 class="text-sm font-medium text-gray-900 border-b border-gray-200 pb-2">交易详情</h4>
                
                <!-- 交易状态和区块信息 -->
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                  <div>
                    <span class="text-gray-500">状态: </span>
                    <span :class="getReceiptStatusClass(transactionReceipts[tx.tx_id || tx.hash].status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full ml-2">
                      {{ getReceiptStatusText(transactionReceipts[tx.tx_id || tx.hash].status) }}
                    </span>
                  </div>
                  <div>
                    <span class="text-gray-500">区块: </span>
                    <span class="text-gray-600">{{ transactionReceipts[tx.tx_id || tx.hash].block_number?.toLocaleString() || 'N/A' }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">区块内位置: </span>
                    <span class="text-gray-600">{{ transactionReceipts[tx.tx_id || tx.hash].transaction_index || 'N/A' }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">时间戳: </span>
                    <span class="text-gray-600">{{ formatTimestamp(tx.ctime || tx.timestamp) }}</span>
                  </div>
                </div>

                <!-- Gas 费用信息 -->
                <div class="bg-gray-50 p-4 rounded-lg">
                  <h5 class="text-sm font-medium text-gray-900 mb-3">Gas 费用信息</h5>
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
                    <div>
                      <span class="text-gray-500">Gas 使用: </span>
                      <span class="text-gray-600">{{ transactionReceipts[tx.tx_id || tx.hash].gas_used?.toLocaleString() || 'N/A' }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">累计 Gas: </span>
                      <span class="text-gray-600">{{ transactionReceipts[tx.tx_id || tx.hash].cumulative_gas_used?.toLocaleString() || 'N/A' }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">Gas 价格: </span>
                      <span class="text-gray-600">{{ formatGasPrice(tx.gas_price || tx.gasPrice) }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">交易费用: </span>
                      <span class="text-gray-600">{{ formatTransactionFee(tx.gas_price || tx.gasPrice, transactionReceipts[tx.tx_id || tx.hash].gas_used) }}</span>
                    </div>
                  </div>
                </div>

                <!-- 交易属性 -->
                <div class="bg-gray-50 p-4 rounded-lg">
                  <h5 class="text-sm font-medium text-gray-900 mb-3">交易属性</h5>
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
                    <div>
                      <span class="text-gray-500">交易类型: </span>
                      <span class="text-gray-600">{{ getTransactionTypeText(tx.type || tx.tx_type) }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">Nonce: </span>
                      <span class="text-gray-600">{{ tx.nonce || 'N/A' }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">输入数据: </span>
                      <span class="text-gray-600">{{ formatInputData(tx.input || tx.data) }}</span>
                    </div>
                    <div v-if="transactionReceipts[tx.tx_id || tx.hash].contract_address">
                      <span class="text-gray-500">合约地址: </span>
                      <span class="font-mono text-blue-600 cursor-pointer hover:text-blue-800" @click="copyToClipboard(transactionReceipts[tx.tx_id || tx.hash].contract_address, $event)">
                        {{ transactionReceipts[tx.tx_id || tx.hash].contract_address }}
                      </span>
                    </div>
                  </div>
                </div>

                <!-- 交易日志 -->
                <div v-if="transactionReceipts[tx.tx_id || tx.hash].logs_data" class="bg-gray-50 p-4 rounded-lg">
                  <h5 class="text-sm font-medium text-gray-900 mb-3">交易日志</h5>
                  <div class="bg-white p-3 rounded border overflow-x-auto max-w-full">
                    <pre class="text-xs text-gray-700 whitespace-pre-wrap break-all max-w-full">{{ formatLogsData(transactionReceipts[tx.tx_id || tx.hash].logs_data) }}</pre>
                  </div>
                </div>
              </div>

              <div v-else class="text-center py-4 text-gray-500 text-sm">
                暂无凭证信息
              </div>
            </div>
          </div>

          <!-- 分页控件 -->
          <div v-if="totalPages > 1" class="mt-6 flex justify-center">
            <nav class="flex items-center space-x-2">
              <button 
                @click="changePage(currentPage - 1)" 
                :disabled="currentPage <= 1"
                class="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                上一页
              </button>
              
              <div class="flex items-center space-x-1">
                <span v-for="page in visiblePages" :key="page" 
                      @click="changePage(page)"
                      :class="[
                        'px-3 py-2 text-sm font-medium rounded-md cursor-pointer',
                        page === currentPage 
                          ? 'bg-blue-600 text-white' 
                          : 'text-gray-500 bg-white border border-gray-300 hover:bg-gray-50'
                      ]"
                >
                  {{ page }}
                </span>
              </div>
              
              <button 
                @click="changePage(currentPage + 1)" 
                :disabled="currentPage >= totalPages"
                class="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                下一页
              </button>
            </nav>
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

// 复制提示（跟随点击位置）
const showToast = ref(false)
const toastMessage = ref('')
const toastX = ref<number | null>(null)
const toastY = ref<number | null>(null)
const toastStyle = computed(() => {
  if (toastX.value !== null && toastY.value !== null) {
    return { top: `${toastY.value}px`, left: `${toastX.value}px` }
  }
  return { top: '16px', right: '16px' }
})
let toastTimer: any = null

// 分页相关数据
const currentPage = ref(1)
const pageSize = ref(20)
const totalCount = ref(0)
const totalPages = ref(1)

// 交易展开相关数据
const expandedTransactions = ref<Record<string, boolean>>({})
const loadingReceipts = ref<Record<string, boolean>>({})
const transactionReceipts = ref<Record<string, any>>({})

// 计算属性
const isFilteredByBlock = computed(() => {
  // 检查交易是否按区块筛选
  if (transactions.value.length === 0) return false
  
  // 如果第一个交易有区块高度字段，说明是按区块筛选的
  const firstTx = transactions.value[0]
  return !!(firstTx.blockHeight || firstTx.block_number || firstTx.block_height)
})

// 分页计算属性
const visiblePages = computed(() => {
  const pages = []
  const start = Math.max(1, currentPage.value - 2)
  const end = Math.min(totalPages.value, currentPage.value + 2)
  
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  return pages
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
        page: currentPage.value,
        page_size: pageSize.value
      })
      
      if (response && response.success === true) {
        console.log('📊 后端返回交易数据:', response.data)
        
        // 新API直接返回交易数据，不需要过滤
        const responseData = response.data as any
        console.log('🔍 解析API返回数据:', responseData)
        
        if (responseData?.transactions && Array.isArray(responseData.transactions)) {
          transactions.value = responseData.transactions
          
          // 尝试多种可能的字段名
          totalCount.value = responseData.total_count || responseData.total || responseData.totalCount || responseData.totalTransactions || responseData.transaction_count || 0
          
          // 如果总数还是0，但有交易数据，说明可能是单页返回所有数据
          if (totalCount.value === 0 && transactions.value.length > 0) {
            // 尝试从区块信息中获取交易总数
            if (block.value && block.value.transaction_count) {
              totalCount.value = block.value.transaction_count
              console.log('📊 从区块信息获取交易总数:', totalCount.value)
            } else if (block.value && block.value.transactions) {
              totalCount.value = block.value.transactions
              console.log('📊 从区块信息获取交易总数:', totalCount.value)
            } else {
              totalCount.value = transactions.value.length
              console.log('⚠️ 后端未返回总数，使用当前页交易数量作为总数')
            }
          }
          
          totalPages.value = Math.max(1, Math.ceil(totalCount.value / pageSize.value))
          console.log('✅ 成功加载区块交易:', transactions.value.length, '笔交易，总计:', totalCount.value, '页数:', totalPages.value)
        } else {
          console.warn('API返回数据格式异常:', responseData)
          transactions.value = []
          totalCount.value = 0
          totalPages.value = 1
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
        page: currentPage.value,
        page_size: pageSize.value
      })
      
      if (response && response.success === true) {
        console.log('📊 后端返回交易数据:', response.data)
        
        // 新API直接返回交易数据，不需要过滤
        const responseData = response.data as any
        console.log('🔍 解析API返回数据:', responseData)
        
        if (responseData?.transactions && Array.isArray(responseData.transactions)) {
          transactions.value = responseData.transactions
          
          // 尝试多种可能的字段名
          totalCount.value = responseData.total_count || responseData.total || responseData.totalCount || responseData.totalTransactions || responseData.transaction_count || 0
          
          // 如果总数还是0，但有交易数据，说明可能是单页返回所有数据
          if (totalCount.value === 0 && transactions.value.length > 0) {
            // 尝试从区块信息中获取交易总数
            if (block.value && block.value.transaction_count) {
              totalCount.value = block.value.transaction_count
              console.log('📊 从区块信息获取交易总数:', totalCount.value)
            } else if (block.value && block.value.transactions) {
              totalCount.value = block.value.transactions
              console.log('📊 从区块信息获取交易总数:', totalCount.value)
            } else {
              totalCount.value = transactions.value.length
              console.log('⚠️ 后端未返回总数，使用当前页交易数量作为总数')
            }
          }
          
          totalPages.value = Math.max(1, Math.ceil(totalCount.value / pageSize.value))
          console.log('✅ 成功加载区块交易:', transactions.value.length, '笔交易，总计:', totalCount.value, '页数:', totalPages.value)
        } else {
          console.warn('API返回数据格式异常:', responseData)
          transactions.value = []
          totalCount.value = 0
          totalPages.value = 1
        }
      } else {
        throw new Error(response?.message || '获取交易信息失败')
      }
    }
  } catch (error) {
    console.error('Failed to load transactions:', error)
    transactions.value = []
    totalCount.value = 0
    totalPages.value = 1
  } finally {
    loadingTransactions.value = false
  }
}

// 复制到剪贴板（支持传入点击事件以定位提示位置）
const copyToClipboard = async (text: string, e?: MouseEvent) => {
  try {
    await navigator.clipboard.writeText(text)
    // 计算提示位置（相对视口，稍微偏移）
    if (e) {
      const offset = 12
      toastX.value = Math.min(window.innerWidth - 16, e.clientX + offset)
      toastY.value = Math.min(window.innerHeight - 16, e.clientY + offset)
    } else {
      toastX.value = null
      toastY.value = null
    }
    toastMessage.value = '已复制到剪贴板'
    showToast.value = true
    if (toastTimer) clearTimeout(toastTimer)
    toastTimer = setTimeout(() => {
      showToast.value = false
      toastTimer = null
    }, 1200)
  } catch (err) {
    console.error('复制失败:', err)
  }
}

// 分页切换
const changePage = async (page: number) => {
  if (page < 1 || page > totalPages.value) return
  
  currentPage.value = page
  await loadTransactions()
}

// 切换交易展开状态
const toggleTransactionExpansion = async (txHash: string) => {
  if (!txHash) return
  
  const isExpanded = expandedTransactions.value[txHash]
  expandedTransactions.value[txHash] = !isExpanded
  
  // 如果展开且还没有加载凭证，且用户已登录，则加载
  if (!isExpanded && !transactionReceipts.value[txHash] && authStore.isAuthenticated) {
    await loadTransactionReceipt(txHash)
  }
}

// 加载交易凭证
const loadTransactionReceipt = async (txHash: string) => {
  if (!txHash || transactionReceipts.value[txHash]) return
  
  try {
    loadingReceipts.value[txHash] = true
    
    // 调用API获取凭证
    const response = await transactionsApi.getTransactionReceipt(txHash)
    
    if (response && response.success === true) {
      transactionReceipts.value[txHash] = response.data
      console.log('✅ 成功加载交易凭证:', txHash, response.data)
    } else {
      console.warn('获取交易凭证失败:', response?.message)
    }
  } catch (error) {
    console.error('Failed to load transaction receipt:', error)
  } finally {
    loadingReceipts.value[txHash] = false
  }
}

// 凭证状态样式
const getReceiptStatusClass = (status: number) => {
  switch (status) {
    case 0:
      return 'bg-red-100 text-red-800'
    case 1:
      return 'bg-green-100 text-green-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

// 凭证状态文本
const getReceiptStatusText = (status: number) => {
  switch (status) {
    case 0:
      return 'Failed'
    case 1:
      return 'Success'
    default:
      return 'Unknown'
  }
}

// 格式化日志数据
const formatLogsData = (logsData: string) => {
  try {
    if (typeof logsData === 'string') {
      const parsed = JSON.parse(logsData)
      return JSON.stringify(parsed, null, 2)
    }
    return JSON.stringify(logsData, null, 2)
  } catch (error) {
    return logsData || 'Invalid logs data'
  }
}

// 格式化Gas价格
const formatGasPrice = (gasPrice: number | string) => {
  if (!gasPrice) return 'N/A'
  
  const price = typeof gasPrice === 'string' ? parseInt(gasPrice, 16) : gasPrice
  if (price === 0) return '0 Gwei'
  
  const gwei = price / 1e9
  if (gwei >= 1) {
    return `${gwei.toFixed(2)} Gwei`
  } else {
    return `${(gwei * 1000).toFixed(2)} Mwei`
  }
}

// 格式化交易费用
const formatTransactionFee = (gasPrice: number | string, gasUsed: number) => {
  if (!gasPrice || !gasUsed) return 'N/A'
  
  const price = typeof gasPrice === 'string' ? parseInt(gasPrice, 16) : gasPrice
  const fee = price * gasUsed
  
  if (fee === 0) return '0 ETH'
  
  const eth = fee / 1e18
  if (eth < 0.001) {
    return `${(eth * 1000).toFixed(6)} mETH`
  } else {
    return `${eth.toFixed(6)} ETH`
  }
}

// 获取交易类型文本
const getTransactionTypeText = (type: number | string) => {
  if (!type) return 'Legacy'
  
  const txType = typeof type === 'string' ? parseInt(type, 16) : type
  
  switch (txType) {
    case 0:
      return 'Legacy'
    case 1:
      return 'EIP-2930'
    case 2:
      return 'EIP-1559'
    case 3:
      return 'EIP-4844'
    default:
      return `Type ${txType}`
  }
}

// 格式化输入数据
const formatInputData = (input: string) => {
  if (!input || input === '0x') return '0x (No input data)'
  
  if (input.length <= 66) {
    return input
  }
  
  return `${input.substring(0, 32)}...${input.substring(input.length - 32)}`
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
