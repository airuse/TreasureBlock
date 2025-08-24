<template>
  <div class="space-y-6">
    <!-- 页面标题和返回按钮 -->
    <div class="flex items-center space-x-4">
      <router-link 
        to="/eth/addresses" 
        class="inline-flex items-center px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
        </svg>
        返回合约列表
      </router-link>
      <h1 class="text-2xl font-bold text-gray-900">合约详情</h1>
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
          加载合约信息中...
        </div>
      </div>
    </div>

    <!-- 合约信息 -->
    <div v-else-if="contract" class="space-y-6">
      <!-- 合约基本信息 -->
      <div class="card">
        <div class="flex items-start space-x-6">
          <!-- 合约Logo -->
          <div class="flex-shrink-0">
            <img 
              v-if="contract.contract_logo" 
              :src="contract.contract_logo" 
              alt="合约Logo" 
              class="w-24 h-24 rounded-lg object-cover border-2 border-gray-200"
            />
            <div v-else class="w-24 h-24 bg-gray-200 rounded-lg flex items-center justify-center">
              <svg class="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path>
              </svg>
            </div>
          </div>
          
          <!-- 合约基本信息 -->
          <div class="flex-1">
            <div class="flex items-center space-x-3 mb-4">
              <h2 class="text-xl font-bold text-gray-900">{{ contract.name || '未命名合约' }}</h2>
              <span v-if="contract.symbol" class="text-lg font-medium text-gray-600">({{ contract.symbol }})</span>
              <span :class="getStatusClass(contract.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                {{ getStatusText(contract.status) }}
              </span>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-500">合约地址</label>
                <p class="mt-1 text-sm text-gray-900 font-mono cursor-pointer hover:text-blue-600" @click="copyToClipboard(contract.address, $event)">
                  {{ contract.address }}
                </p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-500">合约类型</label>
                <p class="mt-1 text-sm text-gray-900">{{ getTypeText(contract.contract_type) }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-500">精度</label>
                <p class="mt-1 text-sm text-gray-900">{{ contract.decimals || 'N/A' }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-500">总供应量</label>
                <p class="mt-1 text-sm text-gray-900">{{ formatTotalSupply(contract.total_supply) }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-500">验证状态</label>
                <p class="mt-1 text-sm text-gray-900">
                  <span :class="contract.verified ? 'text-green-600' : 'text-red-600'">
                    {{ contract.verified ? '已验证' : '未验证' }}
                  </span>
                </p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-500">创建者</label>
                <p class="mt-1 text-sm text-gray-900 font-mono cursor-pointer hover:text-blue-600" @click="copyToClipboard(contract.creator, $event)">
                  {{ contract.creator || 'N/A' }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 合约详细信息 -->
      <div class="card">
        <h3 class="text-lg font-medium text-gray-900 mb-4">合约详细信息</h3>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <h4 class="text-sm font-medium text-gray-700 mb-2">接口</h4>
            <div class="bg-gray-50 p-3 rounded-lg">
              <div v-if="parsedInterfaces.length > 0" class="space-y-1">
                <span v-for="interfaceName in parsedInterfaces" :key="interfaceName" class="inline-block px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded mr-2 mb-2">
                  {{ interfaceName }}
                </span>
              </div>
              <p v-else class="text-sm text-gray-500">无接口信息</p>
            </div>
          </div>
          
          <div>
            <h4 class="text-sm font-medium text-gray-700 mb-2">方法</h4>
            <div class="bg-gray-50 p-3 rounded-lg">
              <div v-if="parsedMethods.length > 0" class="space-y-1">
                <span v-for="method in parsedMethods" :key="method" class="inline-block px-2 py-1 text-xs bg-green-100 text-green-800 rounded mr-2 mb-2">
                  {{ method }}
                </span>
              </div>
              <p v-else class="text-sm text-gray-500">无方法信息</p>
            </div>
          </div>
          
          <div>
            <h4 class="text-sm font-medium text-gray-700 mb-2">事件</h4>
            <div class="bg-gray-50 p-3 rounded-lg">
              <div v-if="parsedEvents.length > 0" class="space-y-1">
                <span v-for="event in parsedEvents" :key="event" class="inline-block px-2 py-1 text-xs bg-purple-100 text-purple-800 rounded mr-2 mb-2">
                  {{ event }}
                </span>
              </div>
              <p v-else class="text-sm text-gray-500">无事件信息</p>
            </div>
          </div>
          
          <div>
            <h4 class="text-sm font-medium text-gray-700 mb-2">元数据</h4>
            <div class="bg-gray-50 p-3 rounded-lg">
              <pre v-if="parsedMetadata" class="text-xs text-gray-700 whitespace-pre-wrap">{{ JSON.stringify(parsedMetadata, null, 2) }}</pre>
              <p v-else class="text-sm text-gray-500">无元数据信息</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 创建信息 -->
      <div class="card">
        <h3 class="text-lg font-medium text-gray-900 mb-4">创建信息</h3>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-500">创建交易</label>
            <p class="mt-1 text-sm text-gray-900 font-mono cursor-pointer hover:text-blue-600" @click="copyToClipboard(contract.creation_tx, $event)">
              {{ contract.creation_tx || 'N/A' }}
            </p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">创建区块</label>
            <p class="mt-1 text-sm text-gray-900">{{ contract.creation_block?.toLocaleString() || 'N/A' }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">创建时间</label>
            <p class="mt-1 text-sm text-gray-900">{{ formatTimestamp(contract.c_time) }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-500">最后更新</label>
            <p class="mt-1 text-sm text-gray-900">{{ formatTimestamp(contract.m_time) }}</p>
          </div>
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
        <p class="text-gray-500 mb-4">{{ errorMessage || '无法加载合约信息' }}</p>
        <button 
          @click="loadContractData" 
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
import { getContractByAddress } from '@/api/contracts'

// 路由参数
const route = useRoute()
const contractAddress = computed(() => route.params.address as string)

// 响应式数据
const contract = ref<any>(null)
const isLoading = ref(true)
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

// 计算属性
const parsedInterfaces = computed(() => {
  if (!contract.value?.interfaces) return []
  try {
    return JSON.parse(contract.value.interfaces)
  } catch {
    return []
  }
})

const parsedMethods = computed(() => {
  if (!contract.value?.methods) return []
  try {
    return JSON.parse(contract.value.methods)
  } catch {
    return []
  }
})

const parsedEvents = computed(() => {
  if (!contract.value?.events) return []
  try {
    return JSON.parse(contract.value.events)
  } catch {
    return []
  }
})

const parsedMetadata = computed(() => {
  if (!contract.value?.metadata) return null
  try {
    return JSON.parse(contract.value.metadata)
  } catch {
    return null
  }
})

// 加载合约数据
const loadContractData = async () => {
  try {
    isLoading.value = true
    errorMessage.value = ''
    
    const response = await getContractByAddress(contractAddress.value)
    
    if (response && response.success === true) {
      contract.value = response.data
    } else {
      throw new Error(response?.message || '获取合约信息失败')
    }
  } catch (error) {
    console.error('Failed to load contract:', error)
    errorMessage.value = error instanceof Error ? error.message : '加载合约信息失败'
  } finally {
    isLoading.value = false
  }
}

// 复制到剪贴板
const copyToClipboard = async (text: string, e?: MouseEvent) => {
  try {
    await navigator.clipboard.writeText(text)
    // 计算提示位置
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

// 格式化函数
const formatTimestamp = (timestamp: string) => {
  if (!timestamp) return 'N/A'
  try {
    const date = new Date(timestamp)
    return date.toLocaleString()
  } catch {
    return 'Invalid Date'
  }
}

const formatTotalSupply = (supply: string) => {
  if (!supply) return 'N/A'
  try {
    const num = BigInt(supply)
    return num.toLocaleString()
  } catch {
    return supply
  }
}

const getStatusClass = (status: number) => {
  switch (status) {
    case 1:
      return 'bg-green-100 text-green-800'
    case 0:
      return 'bg-red-100 text-red-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

const getStatusText = (status: number) => {
  switch (status) {
    case 1:
      return '启用'
    case 0:
      return '禁用'
    default:
      return '未知'
  }
}

const getTypeText = (type: string) => {
  switch (type?.toLowerCase()) {
    case 'erc20':
      return 'ERC-20 代币'
    case 'erc721':
      return 'ERC-721 NFT'
    case 'erc1155':
      return 'ERC-1155 多代币'
    case 'defi':
      return 'DeFi 协议'
    case 'dex':
      return 'DEX 交易所'
    case 'lending':
      return '借贷协议'
    default:
      return type || '未知'
  }
}

// 组件挂载时加载数据
onMounted(async () => {
  await loadContractData()
})
</script>

<style scoped>
.card {
  @apply bg-white shadow-sm rounded-lg border border-gray-200 p-4;
}
</style>
