<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">地址列表</h1>
      <div class="flex items-center space-x-4">
        <div class="text-sm text-gray-500">
          共 {{ totalAddresses.toLocaleString() }} 个地址
        </div>
      </div>
    </div>

    <!-- 搜索和筛选 -->
    <div class="card">
      <div class="flex flex-col sm:flex-row gap-4">
        <div class="flex-1">
          <label class="block text-sm font-medium text-gray-700 mb-2">搜索地址</label>
          <div class="relative">
            <input 
              v-model="searchQuery" 
              type="text" 
              placeholder="输入地址哈希..."
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
          <label class="block text-sm font-medium text-gray-700 mb-2">地址类型</label>
          <select 
            v-model="typeFilter" 
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">全部</option>
            <option value="contract">合约</option>
            <option value="wallet">钱包</option>
          </select>
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

    <!-- 地址列表 -->
    <div class="card">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">地址</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">类型</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">余额</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">交易数</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">最后活动</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">标签</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="address in addresses" :key="address.hash" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap">
                <router-link :to="`/addresses/${address.hash}`" class="text-blue-600 hover:text-blue-700 font-mono text-sm">
                  {{ formatAddress(address.hash) }}
                </router-link>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getTypeClass(address.type)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                  {{ getTypeText(address.type) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {{ formatAmount(address.balance) }} ETH
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {{ address.transactionCount.toLocaleString() }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {{ formatTimestamp(address.lastActivity) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                <span v-if="address.label" class="inline-flex px-2 py-1 text-xs font-medium bg-gray-100 text-gray-800 rounded">
                  {{ address.label }}
                </span>
                <span v-else class="text-gray-400">-</span>
              </td>
            </tr>
          </tbody>
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
            class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            下一页
          </button>
        </div>
        <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
          <div>
            <p class="text-sm text-gray-700">
              显示第 <span class="font-medium">{{ (currentPage - 1) * pageSize + 1 }}</span> 到 
              <span class="font-medium">{{ Math.min(currentPage * pageSize, totalAddresses) }}</span> 条，
              共 <span class="font-medium">{{ totalAddresses.toLocaleString() }}</span> 条记录
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
import { ref, computed, onMounted, watch } from 'vue'

// 响应式数据
const searchQuery = ref('')
const typeFilter = ref('')
const pageSize = ref(25)
const currentPage = ref(1)
const totalAddresses = ref(0)
const addresses = ref([])

// 计算属性
const totalPages = computed(() => Math.ceil(totalAddresses.value / pageSize.value))

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
const formatAddress = (address: string) => {
  return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`
}

const formatTimestamp = (timestamp: number) => {
  return new Date(timestamp * 1000).toLocaleString()
}

const formatAmount = (amount: number) => {
  return (amount / 1e18).toFixed(6)
}

const getTypeClass = (type: string) => {
  switch (type) {
    case 'contract':
      return 'bg-purple-100 text-purple-800'
    case 'wallet':
      return 'bg-blue-100 text-blue-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

const getTypeText = (type: string) => {
  switch (type) {
    case 'contract':
      return '合约'
    case 'wallet':
      return '钱包'
    default:
      return '未知'
  }
}

// 数据加载
const loadData = () => {
  // 模拟数据
  totalAddresses.value = 1234567
  
  const startIndex = (currentPage.value - 1) * pageSize.value
  const endIndex = startIndex + pageSize.value
  
  addresses.value = []
  
  for (let i = startIndex; i < endIndex; i++) {
    const types = ['contract', 'wallet']
    const randomType = types[Math.floor(Math.random() * types.length)]
    const labels = ['交易所', '矿池', 'DeFi协议', 'NFT市场', null]
    const randomLabel = labels[Math.floor(Math.random() * labels.length)]
    
    addresses.value.push({
      hash: `0x${Math.random().toString(16).substring(2, 42)}`,
      type: randomType,
      balance: Math.random() * 1000e18,
      transactionCount: Math.floor(Math.random() * 10000) + 1,
      lastActivity: Math.floor(Date.now() / 1000) - Math.floor(Math.random() * 86400 * 30),
      label: randomLabel
    })
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

// 监听搜索查询
watch(searchQuery, (newQuery) => {
  if (newQuery) {
    // 这里应该实现搜索逻辑
    console.log('搜索:', newQuery)
  }
})

// 监听类型筛选
watch(typeFilter, () => {
  currentPage.value = 1
  loadData()
})

// 监听页面大小变化
watch(pageSize, () => {
  currentPage.value = 1
  loadData()
})

// 组件挂载时加载数据
onMounted(() => {
  loadData()
})
</script> 