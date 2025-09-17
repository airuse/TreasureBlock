<template>
  <div class="space-y-6">
    <!-- 页面头部 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <div class="flex items-center justify-between">
          <div>
            <div class="flex items-center space-x-2 mb-2">
              <button
                @click="goBack"
                class="text-blue-600 hover:text-blue-800 text-sm flex items-center"
              >
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
                </svg>
                返回个人地址
              </button>
            </div>
            <h1 class="text-2xl font-bold text-gray-900">地址交易详情</h1>
            <p class="mt-1 text-sm text-gray-500">
              查看地址 <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono break-all">{{ formatAddress(address) }}</code> 的所有交易记录
            </p>
          </div>
          <div class="flex items-center space-x-2">
            <div class="w-3 h-3 bg-blue-500 rounded-full"></div>
            <span class="text-sm text-gray-600">BNB 网络</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 地址信息概览 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <!-- 地址单独显示 -->
        <div class="mb-6">
          <div class="text-center">
            <div class="relative inline-block">
              <div class="text-lg font-mono text-gray-900 break-all bg-gray-50 p-3 rounded-lg max-w-full">
                {{ address }}
              </div>
              <button
                @click="copyToClipboard(address)"
                class="absolute -top-2 -right-2 p-1 bg-blue-500 text-white rounded-full hover:bg-blue-600 transition-colors"
                title="复制地址"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
              </button>
            </div>
            <div class="text-sm text-gray-500 mt-2">地址</div>
          </div>
        </div>
        
        <!-- 统计信息 -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div class="text-center p-4 bg-green-50 rounded-lg border border-green-200">
            <div class="text-3xl font-bold text-green-600">{{ totalTransactions }}</div>
            <div class="text-sm text-green-700 font-medium">总交易数</div>
          </div>
          <div class="text-center p-4 bg-purple-50 rounded-lg border border-purple-200">
            <div class="text-3xl font-bold text-purple-600">{{ currentPage }}</div>
            <div class="text-sm text-purple-700 font-medium">当前页</div>
          </div>
          <div class="text-center p-4 bg-orange-50 rounded-lg border border-orange-200">
            <div class="text-3xl font-bold text-orange-600">{{ totalPages }}</div>
            <div class="text-sm text-orange-700 font-medium">总页数</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 筛选和分页控制 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-4 sm:space-y-0">
          <div class="flex items-center space-x-4">
          </div>
          
          <div class="flex items-center space-x-2">
            <button
              @click="goToPage(currentPage - 1)"
              :disabled="currentPage <= 1"
              class="px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              上一页
            </button>
            <span class="text-sm text-gray-700">
              第 {{ currentPage }} 页，共 {{ totalPages }} 页
            </span>
            <button
              @click="goToPage(currentPage + 1)"
              :disabled="currentPage >= totalPages"
              class="px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              下一页
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 交易列表 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">交易记录</h3>
        
        <!-- 加载状态 -->
        <div v-if="loading" class="text-center py-8">
          <div class="inline-flex items-center px-4 py-2 font-semibold leading-6 text-sm text-white bg-blue-600 rounded-md">
            <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            加载中...
          </div>
        </div>

        <!-- 空状态 -->
        <div v-else-if="!transactions || transactions.length === 0" class="text-center py-8">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">暂无交易记录</h3>
          <p class="mt-1 text-sm text-gray-500">该地址暂无相关交易</p>
        </div>

        <!-- 交易列表 -->
        <div v-else class="space-y-4">
          <div
            v-for="tx in transactions"
            :key="tx.id"
            class="border border-gray-200 rounded-lg p-4 hover:bg-gray-50 transition-colors"
          >
            <div class="flex items-start justify-between">
              <div class="flex-1">
                <div class="flex items-center space-x-3 mb-3">
                  <h4 class="text-sm font-medium text-gray-900">
                    交易哈希: {{ formatAddress(tx.tx_id) }}
                  </h4>
                  <span :class="getStatusClass(tx.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                    {{ getStatusText(tx.status) }}
                  </span>
                </div>
                
                <!-- 基础信息 -->
                <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm mb-4">
                  <div>
                    <span class="text-gray-500">区块高度:</span>
                    <span class="ml-2 font-mono text-gray-900">#{{ tx.height }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">区块位置:</span>
                    <span class="ml-2 font-mono text-gray-900">{{ tx.block_index }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">金额:</span>
                    <span class="ml-2 font-mono text-gray-900">{{ formatAmount(tx.amount, tx.symbol) }}</span>
                  </div>
                </div>

                <!-- 地址信息 - 单独显示，更美观 -->
                <div class="space-y-3 mb-4">
                  <!-- 发送方 -->
                  <div>
                    <div class="flex items-center space-x-2 mb-2">
                      <span class="text-sm font-medium text-gray-700">发送方</span>
                      <span class="text-xs text-gray-500">(From)</span>
                    </div>
                    <div class="relative">
                      <div 
                        :class="getAddressClass(tx.address_from, 'from')" 
                        class="font-mono text-sm p-3 rounded-lg break-all transition-all duration-200 hover:shadow-md"
                      >
                        {{ tx.address_from }}
                      </div>
                      <button
                        @click="copyToClipboard(tx.address_from)"
                        class="absolute top-2 right-2 p-1.5 bg-white/80 hover:bg-white text-gray-600 hover:text-gray-800 rounded transition-colors"
                        title="复制地址"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                      </button>
                    </div>
                  </div>

                  <!-- 接收方 -->
                  <div>
                    <div class="flex items-center space-x-2 mb-2">
                      <span class="text-sm font-medium text-gray-700">接收方</span>
                      <span class="text-xs text-gray-500">(To)</span>
                    </div>
                    <div class="relative">
                      <div 
                        :class="getAddressClass(tx.address_to, 'to')" 
                        class="font-mono text-sm p-3 rounded-lg break-all transition-all duration-200 hover:shadow-md"
                      >
                        {{ tx.address_to }}
                      </div>
                      <button
                        @click="copyToClipboard(tx.address_to)"
                        class="absolute top-2 right-2 p-1.5 bg-white/80 hover:bg-white text-gray-600 hover:text-gray-800 rounded transition-colors"
                        title="复制地址"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                      </button>
                    </div>
                  </div>
                </div>


                <!-- Gas信息 -->
                <div class="mt-3 pt-3 border-t border-gray-200">
                  <h5 class="text-xs font-medium text-gray-700 mb-2">Gas 信息</h5>
                  <div class="grid grid-cols-2 md:grid-cols-4 gap-3 text-xs">
                    <div>
                      <span class="text-gray-500">限制:</span>
                      <span class="ml-1 font-mono text-gray-900">{{ tx.gas_limit }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">价格:</span>
                      <span class="ml-1 font-mono text-gray-900">{{ formatGasPrice(tx.gas_price) }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">已用:</span>
                      <span class="ml-1 font-mono text-gray-900">{{ tx.gas_used }}</span>
                    </div>
                    <div>
                      <span class="text-gray-500">有效价格:</span>
                      <span class="ml-1 font-mono text-gray-900">{{ formatGasPrice(tx.effective_gas_price) }}</span>
                    </div>
                  </div>
                  
                  <!-- EIP-1559 相关字段 -->
                  <div v-if="tx.max_fee_per_gas || tx.max_priority_fee_per_gas" class="mt-2 pt-2 border-t border-gray-100">
                    <h6 class="text-xs font-medium text-gray-600 mb-1">EIP-1559 费用</h6>
                    <div class="grid grid-cols-2 gap-3 text-xs">
                      <div>
                        <span class="text-gray-500">最高费用:</span>
                        <span class="ml-1 font-mono text-gray-900">{{ formatGasPrice(tx.max_fee_per_gas) }}</span>
                      </div>
                      <div>
                        <span class="text-gray-500">最高小费:</span>
                        <span class="ml-1 font-mono text-gray-900">{{ formatGasPrice(tx.max_priority_fee_per_gas) }}</span>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 合约信息 -->
                <div v-if="tx.contract_addr" class="mt-3 pt-3 border-t border-gray-200">
                  <h5 class="text-xs font-medium text-gray-700 mb-2">合约信息</h5>
                  <div class="text-xs">
                    <span class="text-gray-500">合约地址:</span>
                    <span class="ml-2 font-mono text-gray-900">{{ formatAddress(tx.contract_addr) }}</span>
                  </div>
                </div>

                <!-- 时间信息 -->
                <div class="mt-3 pt-3 border-t border-gray-200 text-xs text-gray-500">
                  <span>创建时间: {{ tx.ctime }}</span>
                  <span class="ml-4">更新时间: {{ tx.mtime }}</span>
                </div>
              </div>

              <!-- 操作按钮 -->
              <div class="flex flex-col space-y-2 ml-4">
                <button
                  @click="copyToClipboard(tx.tx_id)"
                  class="text-blue-600 hover:text-blue-900 text-sm"
                  title="复制交易哈希"
                >
                  复制哈希
                </button>
                <button
                  @click="viewBlock(tx.height)"
                  class="text-green-600 hover:text-green-900 text-sm"
                  title="查看区块"
                >
                  查看区块
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 分页信息 -->
        <div v-if="totalPages > 1" class="mt-6 flex items-center justify-between">
          <div class="text-sm text-gray-700">
            显示第 {{ (currentPage - 1) * pageSize + 1 }} - {{ Math.min(currentPage * pageSize, totalTransactions) }} 条，
            共 {{ totalTransactions }} 条记录
          </div>
          <div class="flex items-center space-x-2">
            <button
              @click="goToPage(1)"
              :disabled="currentPage <= 1"
              class="px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              首页
            </button>
            <button
              @click="goToPage(currentPage - 1)"
              :disabled="currentPage <= 1"
              class="px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              上一页
            </button>
            <span class="px-3 py-2 text-sm text-gray-700">
              {{ currentPage }} / {{ totalPages }}
            </span>
            <button
              @click="goToPage(currentPage + 1)"
              :disabled="currentPage >= totalPages"
              class="px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              下一页
            </button>
            <button
              @click="goToPage(totalPages)"
              :disabled="currentPage >= totalPages"
              class="px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              末页
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showSuccess, showError } from '@/composables/useToast'
import { getAddressTransactions } from '@/api/personal-addresses'
import type { AddressTransactionResponse, AddressTransactionsResponse } from '@/types/transaction'

// 路由和响应式数据
const route = useRoute()
const router = useRouter()

// 响应式数据
const loading = ref(false)
const transactions = ref<AddressTransactionResponse[]>([])
const totalTransactions = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const selectedChain = ref('')

// 从路由参数获取地址
const address = computed(() => route.query.address as string || '')

// 计算属性
const totalPages = computed(() => Math.ceil(totalTransactions.value / pageSize.value))

// 获取交易列表
const loadTransactions = async () => {
  if (!address.value) {
    showError('地址参数缺失')
    return
  }

  loading.value = true
  try {
    const response = await getAddressTransactions(
      address.value,
      currentPage.value,
      pageSize.value,
      selectedChain.value || undefined
    )
    
    if (response.success) {
      // 确保 transactions 始终是数组，避免 null 值
      transactions.value = response.data.transactions || []
      totalTransactions.value = response.data.total || 0
    } else {
      showError(response.message || '获取交易列表失败')
      // 失败时重置为空数组
      transactions.value = []
      totalTransactions.value = 0
    }
  } catch (error) {
    console.error('获取交易列表失败:', error)
    showError('获取交易列表失败')
    // 异常时也重置为空数组
    transactions.value = []
    totalTransactions.value = 0
  } finally {
    loading.value = false
  }
}

// 处理链类型变化
const handleChainChange = () => {
  currentPage.value = 1
  loadTransactions()
}

// 跳转到指定页
const goToPage = (page: number) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  loadTransactions()
}

// 复制到剪贴板
const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text).then(() => {
    showSuccess('已复制到剪贴板')
  }).catch(() => {
    showError('复制失败')
  })
}

// 查看区块
const viewBlock = (height: number) => {
  router.push({
    path: '/bsc/blocks',
    query: { height: height.toString() }
  })
}

// 返回个人地址页面
const goBack = () => {
  router.push('/bsc/personal/addresses')
}

// 格式化地址显示
const formatAddress = (addr: string) => {
  if (!addr || addr.length <= 10) return addr
  return `${addr.substring(0, 6)}...${addr.substring(addr.length - 4)}`
}

// 格式化金额
const formatAmount = (amount: string, symbol: string) => {
  if (!amount) return '0'
  // 这里可以根据不同代币的精度进行格式化
  const num = parseFloat(amount) / Math.pow(10, 18) // 假设18位精度
  return `${num.toFixed(6)} ${symbol}`
}

// 格式化Gas价格
const formatGasPrice = (gasPrice: string) => {
  if (!gasPrice) return '0'
  const num = parseFloat(gasPrice) / Math.pow(10, 9) // 转换为Gwei
  return `${num.toFixed(2)} Gwei`
}

// 获取状态样式
const getStatusClass = (status: number) => {
  switch (status) {
    case 1: return 'bg-green-100 text-green-800'
    case 2: return 'bg-red-100 text-red-800'
    case 0: return 'bg-gray-100 text-gray-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

// 获取状态文本
const getStatusText = (status: number) => {
  switch (status) {
    case 1: return '成功'
    case 2: return '失败'
    case 0: return '未知'
    default: return '未知'
  }
}

// 获取地址样式类
const getAddressClass = (txAddress: string, type: 'from' | 'to') => {
  // 如果是我关注的地址（当前页面显示的地址）
  if (txAddress.toLowerCase() === address.value.toLowerCase()) {
    if (type === 'from') {
      // 发送方是我关注的地址 - 雅红色背景
      return 'bg-red-50 text-red-800 border border-red-200 shadow-sm'
    } else {
      // 接收方是我关注的地址 - 雅蓝色背景
      return 'bg-blue-50 text-blue-800 border border-blue-200 shadow-sm'
    }
  }
  // 不是关注的地址 - 默认样式
  return 'bg-gray-50 text-gray-700 border border-gray-200'
}

// 监听路由参数变化
watch(() => route.query.address, (newAddress) => {
  if (newAddress && newAddress !== address.value) {
    currentPage.value = 1
    loadTransactions()
  }
})

// 组件挂载时加载数据
onMounted(() => {
  // 根据当前路径设置链类型
  selectedChain.value = 'bsc'
  if (address.value) {
    loadTransactions()
  }
})
</script>
