<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">区块列表</h1>
      <div class="flex items-center space-x-4">
        <!-- WebSocket连接状态 -->
        <div class="flex items-center space-x-2">
          <div 
            :class="[
              'w-2 h-2 rounded-full',
              isConnected ? 'bg-green-400' : 'bg-red-400'
            ]"
          ></div>
          <span class="text-xs text-gray-500">
            {{ isConnected ? '实时连接' : '连接断开' }}
          </span>
        </div>
        
        <div class="text-sm text-gray-500">
          共 {{ totalBlocks.toLocaleString() }} 个区块
        </div>
        <div v-if="!authStore.isAuthenticated" class="text-xs text-orange-600 bg-orange-50 px-2 py-1 rounded">
          游客模式：仅显示100个区块
        </div>
      </div>
    </div>

    <!-- 搜索和筛选（与交易/地址页面风格一致） -->
    <div class="card">
    <div class="flex flex-col sm:flex-row gap-4">
      <div class="flex-1">
        <label class="block text-sm font-medium text-gray-700 mb-2">搜索区块</label>
        <div class="relative">
          <input 
            v-model="searchQuery" 
            type="text" 
            placeholder="输入区块高度或哈希..."
            class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <svg class="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
            </svg>
          </div>
        </div>
      </div>
      <div class="sm:w-48">
        <label class="block text-sm font-medium text-gray-700 mb-2">每页显示</label>
        <select 
          v-model="pageSize" 
          class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="10">10</option>
          <option value="25">25</option>
          <option value="50">50</option>
          <option value="100">100</option>
        </select>
        </div>
      </div>
    </div>

    <!-- 区块列表 -->
    <div class="card">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">区块高度</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">时间戳</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">交易数</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">大小</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Gas使用</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">矿工</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">奖励</th>
            </tr>
          </thead>
          <transition-group tag="tbody" name="block-fade" class="bg-white divide-y divide-gray-200">
            <template v-for="block in blocks" :key="block.height">
              <tr class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap">
                  <router-link :to="`/eth/blocks/${block.height}`" class="text-blue-600 hover:text-blue-700 font-medium">
                    #{{ block.height.toLocaleString() }}
                  </router-link>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {{ formatTimestamp(block.timestamp) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {{ block.transactions }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {{ formatBytes(block.size) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {{ formatGas(block.gasUsed, block.gasLimit) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <span class="font-mono">{{ block.miner }}</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <div>
                    <div class="font-medium">{{ formatMinerTip(block.miner_tip_eth) }} ETH</div>
                    <div v-if="block.burned_eth" class="text-xs text-gray-500">
                      燃烧: {{ formatBurnedEth(block.burned_eth) }} ETH
                    </div>
                  </div>
                </td>
              </tr>
            </template>
          </transition-group>
        </table>
      </div>

      <!-- 分页 -->
      <div class="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
        <div class="flex-1 flex justify-between sm:hidden">
          <button 
            @click="previousPage" 
            :disabled="currentPage === 1"
            class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            上一页
          </button>
          <button 
            @click="nextPage" 
            :disabled="currentPage >= totalPages"
            class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:cursor-not-allowed"
          >
            下一页
          </button>
        </div>
        <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
          <div>
            <p class="text-sm text-gray-700">
              显示第 <span class="font-medium">{{ (currentPage - 1) * pageSize + 1 }}</span> 到 
              <span class="font-medium">{{ Math.min(currentPage * pageSize, totalBlocks) }}</span> 条，
              共 <span class="font-medium">{{ totalBlocks.toLocaleString() }}</span> 条记录
            </p>
          </div>
          <div>
            <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
              <button 
                @click="previousPage" 
                :disabled="currentPage === 1"
                class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <svg class="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" />
                </svg>
              </button>
              
              <button 
                v-for="page in visiblePages" 
                :key="page"
                @click="goToPage(page)"
                :class="[
                  page === currentPage 
                    ? 'z-10 bg-blue-50 border-blue-500 text-blue-600' 
                    : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50',
                  'relative inline-flex items-center px-4 py-2 border text-sm font-medium'
                ]"
              >
                {{ page }}
              </button>
              
              <button 
                @click="nextPage" 
                :disabled="currentPage >= totalPages"
                class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <svg class="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
              </button>
            </nav>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useChainWebSocket } from '@/composables/useWebSocket'
import { useAuthStore } from '@/stores/auth'
import { blocks as blocksApi } from '@/api'

// 认证store
const authStore = useAuthStore()

// 响应式数据
const searchQuery = ref('')
const pageSize = ref(25)
const currentPage = ref(1)
const totalBlocks = ref(0)
const isLoading = ref(false)

// 定义区块类型
interface BlockData {
  height: number
  timestamp: number
  transactions: number
  size: number
  gasUsed: number
  gasLimit: number
  miner: string
  miner_tip_eth?: string | number // 矿工奖励（MinerTipEth）
  burned_eth?: string | number // 燃烧费用
}

const blocks = ref<BlockData[]>([])

// 计算属性
const totalPages = computed(() => Math.ceil(totalBlocks.value / pageSize.value))

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
const formatTimestamp = (timestamp: number) => {
  return new Date(timestamp * 1000).toLocaleString()
}

// 格式化矿工奖励（MinerTipEth）
const formatMinerTip = (minerTip: string | number | undefined): string => {
  if (minerTip === undefined || minerTip === null) {
    return '0'
  }
  
  if (typeof minerTip === 'string') {
    const parsed = parseFloat(minerTip)
    return isNaN(parsed) ? '0' : parsed.toFixed(6)
  }
  
  return minerTip.toFixed(6)
}

// 格式化燃烧费用
const formatBurnedEth = (burnedEth: string | number | undefined): string => {
  if (burnedEth === undefined || burnedEth === null) {
    return '0'
  }
  
  if (typeof burnedEth === 'string') {
    const parsed = parseFloat(burnedEth)
    return isNaN(parsed) ? '0' : parsed.toFixed(6)
  }
  
  return burnedEth.toFixed(6)
}

const formatAddress = (address: string) => {
  return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`
}

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatGas = (used: number, limit: number) => {
  const percentage = ((used / limit) * 100).toFixed(1)
  return `${used.toLocaleString()} / ${limit.toLocaleString()} (${percentage}%)`
}

const formatAmount = (amount: number) => {
  return (amount / 1e18).toFixed(6)
}

// 数据加载
const loadData = async () => {
  try {
    isLoading.value = true
    
    // 根据登录状态调用不同的API
    if (authStore.isAuthenticated) {
      // 已登录用户：调用 /v1/ 下的API
      const response = await blocksApi.getBlocks({ 
        page: currentPage.value, 
        page_size: pageSize.value, 
        chain: 'eth' 
      })
      
      if (response && response.success === true) {
        // 正确处理分页响应结构
        const responseData = response.data as any
        let blocksData: any[] = []
        let totalCount = 0
        
        // 检查不同的数据结构
        if (Array.isArray(responseData)) {
          // 如果直接返回数组
          blocksData = responseData
          totalCount = responseData.length
        } else if (responseData?.blocks && Array.isArray(responseData.blocks)) {
          // 如果返回 { blocks: [...], total: ... }
          blocksData = responseData.blocks
          totalCount = responseData.total || responseData.blocks.length
        } else if (responseData?.data && Array.isArray(responseData.data)) {
          // 如果返回 { data: [...], pagination: {...} }
          blocksData = responseData.data
          totalCount = responseData.pagination?.total || responseData.data.length
        } else {
          console.warn('未知的响应数据结构:', responseData)
          blocksData = []
          totalCount = 0
        }
        
        // 转换API返回的Block类型为组件需要的BlockData类型
        blocks.value = blocksData.map((block: any) => ({
          height: block.height || block.number,
          timestamp: typeof block.timestamp === 'string' ? new Date(block.timestamp).getTime() / 1000 : block.timestamp,
          transactions: block.transaction_count || block.transactions || 0,
          size: block.size,
          gasUsed: block.gas_used || block.gasUsed || 0,
          gasLimit: block.gas_limit || block.gasLimit || 0,
          miner: block.miner || '',
          miner_tip_eth: block.miner_tip_eth // 传递原始字段用于调试
        }))
        
        totalBlocks.value = totalCount
      } else {
        console.error('Failed to load blocks:', response?.message)
        // 如果API调用失败，使用默认数据
        totalBlocks.value = 18456789
        blocks.value = []
      }
    } else {
      // 未登录用户：调用 /no-auth/ 下的API（有限制）
      const response = await blocksApi.getBlocksPublic({ 
        page: currentPage.value, 
        page_size: Math.min(pageSize.value, 100), // 使用动态分页大小，但限制最大100个
        chain: 'eth' 
      })
      
      if (response && response.success === true) {
        // 正确处理分页响应结构
        const responseData = response.data as any
        let blocksData: any[] = []
        let totalCount = 0
        
        // 检查不同的数据结构
        if (Array.isArray(responseData)) {
          // 如果直接返回数组
          blocksData = responseData
          totalCount = responseData.length
        } else if (responseData?.blocks && Array.isArray(responseData.blocks)) {
          // 如果返回 { blocks: [...], total: ... }
          blocksData = responseData.blocks
          totalCount = responseData.total || responseData.blocks.length
        } else if (responseData?.data && Array.isArray(responseData.data)) {
          // 如果返回 { data: [...], pagination: {...} }
          blocksData = responseData.data
          totalCount = responseData.pagination?.total || responseData.data.length
        } else {
          console.warn('未知的响应数据结构:', responseData)
          blocksData = []
          totalCount = 0
        }
        
        // 转换API返回的Block类型为组件需要的BlockData类型
        blocks.value = blocksData.map((block: any) => ({
          height: block.height || block.number,
          timestamp: typeof block.timestamp === 'string' ? new Date(block.timestamp).getTime() / 1000 : block.timestamp,
          transactions: block.transaction_count || block.transactions || 0,
          size: block.size,
          gasUsed: block.gas_used || block.gasUsed || 0,
          gasLimit: block.gas_limit || block.gasLimit || 0,
          miner: block.miner || '',
          miner_tip_eth: block.miner_tip_eth // 传递原始字段用于调试
        }))
        
        totalBlocks.value = totalCount
      } else {
        console.error('Failed to load blocks:', response?.message)
        // 如果API调用失败，使用默认数据
        totalBlocks.value = 20
        blocks.value = []
      }
    }
  } catch (error) {
    console.error('Failed to load blocks:', error)
    // 如果API调用失败，使用默认数据
    if (authStore.isAuthenticated) {
      totalBlocks.value = 18456789
    } else {
      totalBlocks.value = 20
    }
    blocks.value = []
  } finally {
    isLoading.value = false
  }
}

// 分页方法
const previousPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    loadData()
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    loadData()
  }
}

const goToPage = (page: number) => {
  currentPage.value = page
  loadData()
}

// WebSocket集成
const { subscribeChainEvent, unsubscribeChainEvent, isConnected } = useChainWebSocket('eth')

// 存储取消订阅函数
let unsubscribeBlocks: (() => void) | null = null
let unsubscribeStats: (() => void) | null = null

function handleBlockCountUpdate(message: any) {
  if (message.data && typeof message.data.totalBlocks === 'number') {
    totalBlocks.value = message.data.totalBlocks
  }
}

function handleNewBlock(message: any) {
  // 只在第一页才处理新区块
  if (currentPage.value === 1 && message.data) {
    const newBlockHeight = message.data.height || message.data.number
    
    // 判断区块高度是否已存在
    const existingBlockIndex = blocks.value.findIndex(block => block.height === newBlockHeight)
    
    if (existingBlockIndex !== -1) {
      return
    }
    
    const newBlock: BlockData = {
      height: newBlockHeight,
      timestamp: typeof message.data.timestamp === 'string' 
        ? new Date(message.data.timestamp).getTime() / 1000 
        : message.data.timestamp,
      transactions: message.data.transaction_count || message.data.transactions || 0,
      size: message.data.size,
      gasUsed: message.data.gas_used || message.data.gasUsed || 0,
      gasLimit: message.data.gas_limit || message.data.gasLimit || 0,
      miner: message.data.miner || '',
      miner_tip_eth: message.data.miner_tip_eth
    }
    
    // 实现最新区块插入到列表头部的逻辑
    // 1. 先删除列表中最晚的一条（如果列表已满）
    if (blocks.value.length >= pageSize.value) {
      blocks.value.pop() // 删除最后一条
    }
    
    // 2. 在列表头部插入最新区块
    blocks.value.unshift(newBlock)
    
    // 3. 更新总数（如果后端没有实时更新）
    if (totalBlocks.value > 0) {
      totalBlocks.value++
    }
  }
}

function handleBlockUpdate(message: any) {
  if (message.data && message.action === 'update') {
    const updatedBlock = message.data
    const blockHeight = updatedBlock.height || updatedBlock.number
    
    if (blockHeight) {
      // 在列表中查找并更新对应区块
      const blockIndex = blocks.value.findIndex(block => block.height === blockHeight)
      
      if (blockIndex !== -1) {
        // 更新现有区块信息
        const existingBlock = blocks.value[blockIndex]
        
        // 只更新可能变化的字段
        if (updatedBlock.miner_tip_eth !== undefined) {
          existingBlock.miner_tip_eth = updatedBlock.miner_tip_eth
        }
        if (updatedBlock.burned_eth !== undefined) {
          // 如果前端需要burned_eth字段，可以在这里添加
        }
        if (updatedBlock.fee !== undefined) {
          // 如果前端需要fee字段，可以在这里添加
        }
      } else {
        console.log('⚠️ 区块不在当前列表中，无法更新:', blockHeight)
        console.log('当前列表中的区块高度:', blocks.value.map(b => b.height))
      }
    } else {
      console.warn('⚠️ 更新消息中缺少区块高度:', updatedBlock)
    }
  } else {
    console.warn('⚠️ 无效的更新消息格式:', message)
  }
}

function handleStatsUpdate(message: any) {
  if (message.data && typeof message.data.totalBlocks === 'number') {
    totalBlocks.value = message.data.totalBlocks
  }
}

onMounted(() => {
  loadData()
  
  // 订阅WebSocket事件
  console.log('🔌 开始订阅WebSocket事件...')
  
  // 订阅区块事件
  unsubscribeBlocks = subscribeChainEvent('block', (message) => {
    // 根据action区分创建和更新事件
    if (message.action === 'update') {
      handleBlockUpdate(message)
    } else {
      handleNewBlock(message)
    }
  })
  
  // 订阅统计事件
  unsubscribeStats = subscribeChainEvent('stats', handleStatsUpdate)
  
  console.log('✅ WebSocket事件订阅完成')
})

onUnmounted(() => {
  // 取消WebSocket订阅
  console.log('🔌 取消WebSocket订阅...')
  
  if (unsubscribeBlocks) {
    unsubscribeBlocks()
    unsubscribeBlocks = null
  }
  
  if (unsubscribeStats) {
    unsubscribeStats()
    unsubscribeStats = null
  }
  
  console.log('✅ WebSocket订阅已取消')
})

// 监听搜索查询
watch(searchQuery, (newQuery) => {
  if (newQuery) {
    // 实现搜索逻辑
    performSearch(newQuery)
  } else {
    // 清空搜索，重新加载默认数据
    currentPage.value = 1
    loadData()
  }
})

// 执行搜索
const performSearch = async (query: string) => {
  try {
    isLoading.value = true
    // 根据登录状态调用不同的搜索API
    if (authStore.isAuthenticated) {
      // 已登录用户：调用 /v1/ 下的搜索API
      const response = await blocksApi.searchBlocks({ 
        query: query,
        page: 1, 
        page_size: pageSize.value
      })
      
      if (response && response.success === true) {
        handleSearchResults(response.data, query)
      } else {
        // 搜索失败时显示空结果
        blocks.value = []
        totalBlocks.value = 0
      }
    } else {
      // 未登录用户：调用 /no-auth/ 下的搜索API
      console.log('👤 未登录用户，调用 /no-auth/ API 搜索区块')
      const response = await blocksApi.searchBlocksPublic({ 
        query: query,
        page: 1, 
        page_size: Math.min(pageSize.value, 20) // 限制为20个
      })
      
      if (response && response.success === true) {
        handleSearchResults(response.data, query)
      } else {
        console.error('搜索失败:', response?.message)
        // 搜索失败时显示空结果
        blocks.value = []
        totalBlocks.value = 0
      }
    }
  } catch (error) {
    console.error('搜索出错:', error)
    // 搜索出错时显示空结果
    blocks.value = []
    totalBlocks.value = 0
  } finally {
    isLoading.value = false
  }
}

// 处理搜索结果
const handleSearchResults = (responseData: any, query: string) => {
  let blocksData: any[] = []
  let totalCount = 0
  
  // 检查不同的数据结构
  if (Array.isArray(responseData)) {
    blocksData = responseData
    totalCount = responseData.length
  } else if (responseData?.blocks && Array.isArray(responseData.blocks)) {
    blocksData = responseData.blocks
    totalCount = responseData.total || responseData.blocks.length
  } else if (responseData?.data && Array.isArray(responseData.data)) {
    blocksData = responseData.data
    totalCount = responseData.pagination?.total || responseData.data.length
  } else {
    console.warn('未知的搜索响应数据结构:', responseData)
    blocksData = []
    totalCount = 0
  }
  
  // 转换搜索结果
  blocks.value = blocksData.map((block: any) => ({
    height: block.height || block.number,
    timestamp: typeof block.timestamp === 'string' ? new Date(block.timestamp).getTime() / 1000 : block.timestamp,
    transactions: block.transaction_count || block.transactions || 0,
    size: block.size,
    gasUsed: block.gas_used || block.gasUsed || 0,
    gasLimit: block.gas_limit || block.gasLimit || 0,
    miner: block.miner || '',
    miner_tip_eth: block.miner_tip_eth // 传递原始字段用于调试
  }))
  
  totalBlocks.value = totalCount
  currentPage.value = 1 // 搜索后重置到第一页
  
  console.log('✅ 搜索完成:', query, '找到', totalCount, '个结果')
}

// 监听页面大小变化
watch(pageSize, () => {
  currentPage.value = 1
  loadData()
})
</script> 

<style scoped>
.block-fade-enter-active, .block-fade-leave-active {
  transition: all 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}
.block-fade-enter-from {
  opacity: 0;
  transform: translateY(-30px);
}
.block-fade-enter-to {
  opacity: 1;
  transform: translateY(0);
}
.block-fade-leave-from {
  opacity: 1;
  transform: translateY(0);
}
.block-fade-leave-to {
  opacity: 0;
  transform: translateY(30px);
}
</style> 