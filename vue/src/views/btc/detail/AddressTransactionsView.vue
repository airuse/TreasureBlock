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
                class="text-orange-600 hover:text-orange-800 text-sm flex items-center"
              >
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
                </svg>
                返回个人地址
              </button>
            </div>
            <h1 class="text-2xl font-bold text-gray-900">地址UTXO详情</h1>
            <p class="mt-1 text-sm text-gray-500">
              查看地址 <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono break-all">{{ formatAddress(address) }}</code> 的所有UTXO记录
            </p>
          </div>
          <div class="flex items-center space-x-2">
            <div class="w-3 h-3 bg-orange-500 rounded-full"></div>
            <span class="text-sm text-gray-600">BTC 网络</span>
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
                class="absolute -top-2 -right-2 p-1 bg-orange-500 text-white rounded-full hover:bg-orange-600 transition-colors"
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
        <div class="grid grid-cols-1 md:grid-cols-5 gap-6">
          <div class="text-center p-4 bg-green-50 rounded-lg border border-green-200">
            <div class="text-3xl font-bold text-green-600">{{ totalUTXOs }}</div>
            <div class="text-sm text-green-700 font-medium">总UTXO数</div>
          </div>
          <div class="text-center p-4 bg-blue-50 rounded-lg border border-blue-200">
            <div class="text-3xl font-bold text-blue-600">{{ totalValueBTC }}</div>
            <div class="text-sm text-blue-700 font-medium">总余额 (BTC)</div>
          </div>
          <div class="text-center p-4 bg-purple-50 rounded-lg border border-purple-200">
            <div class="text-3xl font-bold text-purple-600">{{ totalValueSatoshi }}</div>
            <div class="text-sm text-purple-700 font-medium">总余额 (聪)</div>
          </div>
          <div class="text-center p-4 bg-orange-50 rounded-lg border border-orange-200">
            <div class="text-3xl font-bold text-orange-600">{{ coinbaseCount }}</div>
            <div class="text-sm text-orange-700 font-medium">Coinbase UTXO</div>
          </div>
          <div class="text-center p-4 bg-yellow-50 rounded-lg border border-yellow-200">
            <div class="text-3xl font-bold text-yellow-600">{{ pendingSpentCount }}</div>
            <div class="text-sm text-yellow-700 font-medium">打包中 UTXO</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 筛选和排序控制 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-4 sm:space-y-0">
          <div class="flex items-center space-x-4">
            <div class="flex items-center space-x-2">
              <label class="text-sm font-medium text-gray-700">排序方式:</label>
              <select
                v-model="sortBy"
                @change="sortUTXOs"
                class="px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
              >
                <option value="value_desc">金额 (高到低)</option>
                <option value="value_asc">金额 (低到高)</option>
                <option value="height_desc">区块高度 (高到低)</option>
                <option value="height_asc">区块高度 (低到高)</option>
                <option value="time_desc">创建时间 (新到旧)</option>
                <option value="time_asc">创建时间 (旧到新)</option>
              </select>
            </div>
            <div class="flex items-center space-x-2">
              <label class="text-sm font-medium text-gray-700">脚本类型:</label>
              <select
                v-model="scriptTypeFilter"
                @change="filterUTXOs"
                class="px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
              >
                <option value="">全部</option>
                <option value="pubkeyhash">P2PKH</option>
                <option value="scripthash">P2SH</option>
                <option value="witness_v0_keyhash">P2WPKH</option>
                <option value="witness_v0_scripthash">P2WSH</option>
                <option value="witness_v1_taproot">P2TR</option>
                <option value="nulldata">Null Data</option>
                <option value="pubkey">PubKey</option>
                <option value="multisig">MultiSig</option>
                <option value="nonstandard">Non-Standard</option>
              </select>
            </div>
            <div class="flex items-center space-x-2">
              <label class="text-sm font-medium text-gray-700">状态:</label>
              <select
                v-model="statusFilter"
                @change="filterUTXOs"
                class="px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
              >
                <option value="">全部</option>
                <option value="unspent">未花费</option>
                <option value="spent">打包中</option>
                <option value="confirmed_spent">已确认花费</option>
              </select>
            </div>
          </div>
          
          <div class="flex items-center space-x-2">
            <button
              @click="refreshUTXOs"
              :disabled="loading"
              class="px-4 py-2 bg-orange-600 text-white text-sm font-medium rounded-md hover:bg-orange-700 disabled:opacity-50 transition-colors"
            >
              <svg v-if="loading" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              {{ loading ? '刷新中...' : '刷新UTXO' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- UTXO列表 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">UTXO记录</h3>
        
        <!-- 加载状态 -->
        <div v-if="loading" class="text-center py-8">
          <div class="inline-flex items-center px-4 py-2 font-semibold leading-6 text-sm text-white bg-orange-600 rounded-md">
            <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            加载中...
          </div>
        </div>

        <!-- 空状态 -->
        <div v-else-if="!filteredUTXOs || filteredUTXOs.length === 0" class="text-center py-8">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">暂无UTXO记录</h3>
          <p class="mt-1 text-sm text-gray-500">该地址暂无未花费的输出</p>
        </div>

        <!-- UTXO列表 -->
        <div v-else class="space-y-4">
          <div
            v-for="utxo in filteredUTXOs"
            :key="`${utxo.tx_id}-${utxo.vout_index}`"
            class="border border-gray-200 rounded-lg p-4 hover:bg-gray-50 transition-colors"
          >
            <div class="flex items-start justify-between">
              <div class="flex-1">
                <div class="flex items-center space-x-3 mb-3">
                  <h4 class="text-sm font-medium text-gray-900">
                    UTXO: {{ formatAddress(utxo.tx_id) }}:{{ utxo.vout_index }}
                  </h4>
                  <span :class="getUTXOStatusClass(utxo)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                    {{ getUTXOStatusText(utxo) }}
                  </span>
                  <span v-if="utxo.is_coinbase" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-yellow-100 text-yellow-800">
                    Coinbase
                  </span>
                </div>
                
                <!-- 基础信息 -->
                <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm mb-4">
                  <div>
                    <span class="text-gray-500">区块高度:</span>
                    <span class="ml-2 font-mono text-gray-900">#{{ utxo.block_height }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">金额:</span>
                    <span class="ml-2 font-mono text-gray-900">{{ formatBTCAmount(utxo.value_satoshi) }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">脚本类型:</span>
                    <span class="ml-2 font-mono text-gray-900">{{ formatScriptType(utxo.script_type) }}</span>
                  </div>
                </div>

                <!-- 交易ID信息 -->
                <div class="space-y-3 mb-4">
                  <div>
                    <div class="flex items-center space-x-2 mb-2">
                      <span class="text-sm font-medium text-gray-700">交易ID</span>
                    </div>
                    <div class="relative">
                      <div class="font-mono text-sm p-3 rounded-lg break-all bg-gray-50 text-gray-700 border border-gray-200">
                        {{ utxo.tx_id }}
                      </div>
                      <button
                        @click="copyToClipboard(utxo.tx_id)"
                        class="absolute top-2 right-2 p-1.5 bg-white/80 hover:bg-white text-gray-600 hover:text-gray-800 rounded transition-colors"
                        title="复制交易ID"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                      </button>
                    </div>
                  </div>
                </div>

                <!-- 花费信息 -->
                <div v-if="utxo.spent_tx_id" class="mt-3 pt-3 border-t border-gray-200">
                  <h5 class="text-xs font-medium text-gray-700 mb-2">花费信息</h5>
                  <div class="space-y-2">
                    <div>
                      <span class="text-gray-500 text-xs">花费交易:</span>
                      <span class="ml-2 font-mono text-xs text-gray-900">{{ formatAddress(utxo.spent_tx_id) }}</span>
                    </div>
                    <div v-if="utxo.spent_height">
                      <span class="text-gray-500 text-xs">花费高度:</span>
                      <span class="ml-2 font-mono text-xs text-gray-900">#{{ utxo.spent_height }}</span>
                    </div>
                    <div v-if="utxo.spent_at">
                      <span class="text-gray-500 text-xs">花费时间:</span>
                      <span class="ml-2 font-mono text-xs text-gray-900">{{ formatDateTime(utxo.spent_at) }}</span>
                    </div>
                    <div v-if="utxo.status === 'spent'" class="mt-2 p-2 bg-orange-50 border border-orange-200 rounded text-xs text-orange-700">
                      <div class="flex items-center">
                        <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 19.5c-.77.833.192 2.5 1.732 2.5z" />
                        </svg>
                        此UTXO刚被打包，尚未达到安全确认高度，请谨慎使用
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 时间信息 -->
                <div class="mt-3 pt-3 border-t border-gray-200 text-xs text-gray-500">
                  <span>创建时间: {{ formatDateTime(utxo.ctime) }}</span>
                  <span class="ml-4">更新时间: {{ formatDateTime(utxo.mtime) }}</span>
                </div>
              </div>

              <!-- 操作按钮 -->
              <div class="flex flex-col space-y-2 ml-4">
                <button
                  @click="copyToClipboard(utxo.tx_id)"
                  class="text-orange-600 hover:text-orange-900 text-sm"
                  title="复制交易ID"
                >
                  复制交易ID
                </button>
                <button
                  @click="viewBlock(utxo.block_height)"
                  class="text-green-600 hover:text-green-900 text-sm"
                  title="查看区块"
                >
                  查看区块
                </button>
                <button
                  v-if="utxo.spent_tx_id"
                  @click="viewTransaction(utxo.spent_tx_id)"
                  class="text-blue-600 hover:text-blue-900 text-sm"
                  title="查看花费交易"
                >
                  查看花费交易
                </button>
              </div>
            </div>
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
import { getAddressUTXOs } from '@/api/personal-addresses'
import type { BTCUTXO } from '@/types/personal-address'

// 路由和响应式数据
const route = useRoute()
const router = useRouter()

// 响应式数据
const loading = ref(false)
const utxos = ref<BTCUTXO[]>([])
const sortBy = ref('value_desc')
const scriptTypeFilter = ref('')
const statusFilter = ref('')

// 从路由参数获取地址
const address = computed(() => route.query.address as string || '')

// 计算属性
const totalUTXOs = computed(() => utxos.value.length)

const totalValueSatoshi = computed(() => {
  return utxos.value.reduce((sum, utxo) => sum + utxo.value_satoshi, 0)
})

const totalValueBTC = computed(() => {
  const satoshi = totalValueSatoshi.value
  return (satoshi / 100000000).toFixed(8)
})

const coinbaseCount = computed(() => {
  return utxos.value.filter(utxo => utxo.is_coinbase).length
})

const pendingSpentCount = computed(() => {
  return utxos.value.filter(utxo => utxo.status === 'spent').length
})

const filteredUTXOs = computed(() => {
  let filtered = [...utxos.value]
  
  // 按脚本类型过滤
  if (scriptTypeFilter.value) {
    filtered = filtered.filter(utxo => utxo.script_type === scriptTypeFilter.value)
  }
  
  // 按状态过滤
  if (statusFilter.value) {
    filtered = filtered.filter(utxo => {
      switch (statusFilter.value) {
        case 'unspent':
          return !utxo.spent_tx_id && utxo.status !== 'spent'
        case 'spent':
          return utxo.status === 'spent'
        case 'confirmed_spent':
          return utxo.spent_tx_id && utxo.status !== 'spent'
        default:
          return true
      }
    })
  }
  
  // 排序
  switch (sortBy.value) {
    case 'value_desc':
      filtered.sort((a, b) => b.value_satoshi - a.value_satoshi)
      break
    case 'value_asc':
      filtered.sort((a, b) => a.value_satoshi - b.value_satoshi)
      break
    case 'height_desc':
      filtered.sort((a, b) => b.block_height - a.block_height)
      break
    case 'height_asc':
      filtered.sort((a, b) => a.block_height - b.block_height)
      break
    case 'time_desc':
      filtered.sort((a, b) => new Date(b.ctime).getTime() - new Date(a.ctime).getTime())
      break
    case 'time_asc':
      filtered.sort((a, b) => new Date(a.ctime).getTime() - new Date(b.ctime).getTime())
      break
  }
  
  return filtered
})

// 获取UTXO列表
const loadUTXOs = async () => {
  if (!address.value) {
    showError('地址参数缺失')
    return
  }

  loading.value = true
  try {
    const response = await getAddressUTXOs(address.value)
    
    if (response.success) {
      utxos.value = response.data || []
    } else {
      showError(response.message || '获取UTXO列表失败')
      utxos.value = []
    }
  } catch (error) {
    console.error('获取UTXO列表失败:', error)
    showError('获取UTXO列表失败')
    utxos.value = []
  } finally {
    loading.value = false
  }
}

// 刷新UTXO
const refreshUTXOs = () => {
  loadUTXOs()
}

// 排序UTXO
const sortUTXOs = () => {
  // 排序逻辑在computed中处理
}

// 过滤UTXO
const filterUTXOs = () => {
  // 过滤逻辑在computed中处理
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
    path: '/btc/blocks',
    query: { height: height.toString() }
  })
}

// 查看交易
const viewTransaction = (txId: string) => {
  router.push({
    path: '/btc/transactions',
    query: { txId: txId }
  })
}

// 返回个人地址页面
const goBack = () => {
  router.push('/btc/personal/addresses')
}

// 格式化地址显示
const formatAddress = (addr: string) => {
  if (!addr || addr.length <= 10) return addr
  return `${addr.substring(0, 6)}...${addr.substring(addr.length - 4)}`
}

// 格式化BTC金额
const formatBTCAmount = (satoshi: number) => {
  const btc = satoshi / 100000000
  return `${btc.toFixed(8)} BTC`
}

// 格式化日期时间
const formatDateTime = (dateStr: string) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

// 格式化脚本类型显示
const formatScriptType = (scriptType: string) => {
  const typeMap: { [key: string]: string } = {
    'pubkeyhash': 'P2PKH',
    'scripthash': 'P2SH', 
    'witness_v0_keyhash': 'P2WPKH',
    'witness_v0_scripthash': 'P2WSH',
    'witness_v1_taproot': 'P2TR',
    'nulldata': 'Null Data',
    'pubkey': 'PubKey',
    'multisig': 'MultiSig',
    'nonstandard': 'Non-Standard'
  }
  return typeMap[scriptType] || scriptType
}

// 获取UTXO状态样式
const getUTXOStatusClass = (utxo: BTCUTXO) => {
  if (utxo.status === 'spent') {
    return 'bg-orange-100 text-orange-800' // 刚被打包但未达到安全高度
  }
  if (utxo.spent_tx_id) {
    return 'bg-red-100 text-red-800' // 已确认花费
  }
  return 'bg-green-100 text-green-800' // 未花费
}

// 获取UTXO状态文本
const getUTXOStatusText = (utxo: BTCUTXO) => {
  if (utxo.status === 'spent') {
    return '打包中' // 刚被打包但未达到安全高度
  }
  if (utxo.spent_tx_id) {
    return '已花费' // 已确认花费
  }
  return '未花费' // 未花费
}

// 监听路由参数变化
watch(() => route.query.address, (newAddress) => {
  if (newAddress && newAddress !== address.value) {
    loadUTXOs()
  }
})

// 组件挂载时加载数据
onMounted(() => {
  if (address.value) {
    loadUTXOs()
  }
})
</script>
