<template>
  <div class="space-y-6">
    <!-- 页面标题和返回按钮 -->
    <div class="flex items-center space-x-4">
      <button 
        @click="goBack"
        class="inline-flex items-center px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
        </svg>
        返回
      </button>
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

    <!-- 合约信息（概览） -->
    <div v-else-if="contract" class="space-y-6">
      <!-- 合约基本信息 -->
      <div class="card">
        <div class="flex items-start space-x-6">
          <!-- 合约Logo -->
          <div class="flex-shrink-0">
            <img v-if="contract.contract_logo" :src="contract.contract_logo" alt="合约Logo" class="w-24 h-24 rounded-lg object-cover border-2 border-gray-200" />
            <div v-else class="w-24 h-24 bg-gray-200 rounded-lg flex items-center justify-center">
              <svg class="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"></path>
              </svg>
            </div>
          </div>
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
                <p class="mt-1 text-sm text-gray-900 font-mono cursor-pointer hover:text-blue-600" @click="copyToClipboard(contract.address, $event)">{{ contract.address }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-500">合约类型</label>
                <p class="mt-1 text-sm text-gray-900">{{ getTypeText(contract.contract_type) }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-500">精度</label>
                <p class="mt-1 text-sm text-gray-900">{{ contract.decimals ?? 'N/A' }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-500">总供应量</label>
                <p class="mt-1 text-sm text-gray-900">{{ formatTotalSupply(contract.total_supply) }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-500">验证状态</label>
                <p class="mt-1 text-sm text-gray-900">
                  <span :class="contract.verified ? 'text-green-600' : 'text-red-600'">{{ contract.verified ? '已验证' : '未验证' }}</span>
                </p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-500">创建者</label>
                <p class="mt-1 text-sm text-gray-900 font-mono cursor-pointer hover:text-blue-600" @click="copyToClipboard(contract.creator, $event)">{{ contract.creator || 'N/A' }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 编辑合约字段（只读展示） -->
      <div class="card">
        <h3 class="text-lg font-medium text-gray-900 mb-4">编辑合约字段（只读展示）</h3>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
          <Field label="地址" :value="contract.address" mono />
          <Field label="合约名称" :value="contract.name" />
          <Field label="合约符号" :value="contract.symbol" />
          <Field label="合约类型" :value="getTypeText(contract.contract_type)" />
          <Field label="精度" :value="contract.decimals ?? 'N/A'" />
          <Field label="总供应量" :value="formatTotalSupply(contract.total_supply)" />
          <Field label="是否ERC-20" :value="contract.is_erc20 ? '是' : '否'" />
          <Field label="是否已验证" :value="contract.verified ? '已验证' : '未验证'" />
          <Field label="创建者" :value="contract.creator" mono />
          <Field label="创建交易哈希" :value="contract.creation_tx" mono />
          <Field label="创建区块" :value="contract.creation_block?.toString()" />
          <Field class="md:col-span-3" label="接口 (JSON/数组)" :value="JSON.stringify(parsedInterfaces, null, 2)" pre />
          <Field class="md:col-span-3" label="方法 (JSON/数组)" :value="JSON.stringify(parsedMethods, null, 2)" pre />
          <Field class="md:col-span-3" label="事件 (JSON/数组)" :value="JSON.stringify(parsedEvents, null, 2)" pre />
          <Field class="md:col-span-3" label="元数据 (JSON 对象)" :value="parsedMetadata ? JSON.stringify(parsedMetadata, null, 2) : ''" pre />
        </div>
      </div>

      <!-- 币种信息字段（只读展示，来自后端 coin config） -->
      <div class="card">
        <h3 class="text-lg font-medium text-gray-900 mb-4">维护币种信息字段（只读展示）</h3>
        <div v-if="coinConfig" class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
            <Field label="合约地址" :value="coinConfig.contract_address" mono />
            <Field label="币种名称" :value="coinConfig.name" />
            <Field label="币种符号" :value="coinConfig.symbol" />
            <Field label="类型" :value="formatCoinType(coinConfig.coin_type)" />
            <Field label="精度" :value="coinConfig.precision?.toString()" />
            <Field label="精度别名" :value="coinConfig.decimals?.toString()" />
            <Field label="状态" :value="coinConfig.status === 1 ? '启用' : '禁用'" />
            <Field label="Logo URL" :value="coinConfig.logo_url" />
            <Field label="是否已验证" :value="coinConfig.is_verified ? '已验证' : '未验证'" />
            <Field class="md:col-span-3" label="描述" :value="coinConfig.description || ''" />
          </div>

          <!-- 解析配置列表 -->
          <div>
            <h4 class="text-md font-semibold text-gray-900 mb-2">解析配置 parser_configs</h4>
            <div v-if="Array.isArray(parserConfigs) && parserConfigs.length > 0" class="space-y-4">
              <div v-for="(cfg, idx) in parserConfigs" :key="cfg.id || idx" class="border rounded-md p-3 bg-gray-50">
                <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
                  <Field label="ID" :value="(cfg.id ?? '').toString()" />
                  <Field label="函数签名" :value="cfg.function_signature || ''" mono />
                  <Field label="函数名" :value="cfg.function_name || ''" />
                  <Field label="函数描述" :value="cfg.function_description || ''" class="md:col-span-3" />
                  <Field label="显示格式" :value="cfg.display_format || ''" class="md:col-span-3" />
                  <Field label="是否启用" :value="cfg.is_active ? '是' : '否'" />
                  <Field label="优先级" :value="(cfg.priority ?? 0).toString()" />
                  <Field label="解析类型" :value="cfg.logs_parser_type || 'input_data'" />
                  <Field label="事件签名" :value="cfg.event_signature || ''" mono class="md:col-span-3" />
                  <Field label="事件名称" :value="cfg.event_name || ''" />
                  <Field label="事件描述" :value="cfg.event_description || ''" class="md:col-span-2" />
                </div>
                <div class="mt-3">
                  <h5 class="text-sm font-medium text-gray-700 mb-1">参数配置 param_config</h5>
                  <div v-if="Array.isArray(cfg.param_config) && cfg.param_config.length > 0" class="overflow-x-auto">
                    <table class="min-w-full text-xs">
                      <thead>
                        <tr class="text-gray-500">
                          <th class="text-left pr-4 pb-1">name</th>
                          <th class="text-left pr-4 pb-1">type</th>
                          <th class="text-left pr-4 pb-1">offset</th>
                          <th class="text-left pr-4 pb-1">length</th>
                          <th class="text-left pr-4 pb-1">description</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="(p, i) in cfg.param_config" :key="i" class="text-gray-800">
                          <td class="pr-4 py-0.5">{{ p.name }}</td>
                          <td class="pr-4 py-0.5">{{ p.type }}</td>
                          <td class="pr-4 py-0.5">{{ p.offset }}</td>
                          <td class="pr-4 py-0.5">{{ p.length }}</td>
                          <td class="pr-4 py-0.5">{{ p.description }}</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                  <p v-else class="text-xs text-gray-500">无</p>
                </div>
                <div class="mt-3">
                  <h5 class="text-sm font-medium text-gray-700 mb-1">日志参数配置 logs_param_config</h5>
                  <div v-if="Array.isArray(cfg.logs_param_config) && cfg.logs_param_config.length > 0" class="overflow-x-auto">
                    <table class="min-w-full text-xs">
                      <thead>
                        <tr class="text-gray-500">
                          <th class="text-left pr-4 pb-1">name</th>
                          <th class="text-left pr-4 pb-1">type</th>
                          <th class="text-left pr-4 pb-1">topic_index</th>
                          <th class="text-left pr-4 pb-1">data_index</th>
                          <th class="text-left pr-4 pb-1">description</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="(lp, j) in cfg.logs_param_config" :key="j" class="text-gray-800">
                          <td class="pr-4 py-0.5">{{ lp.name }}</td>
                          <td class="pr-4 py-0.5">{{ lp.type }}</td>
                          <td class="pr-4 py-0.5">{{ lp.topic_index ?? '-' }}</td>
                          <td class="pr-4 py-0.5">{{ lp.data_index ?? '-' }}</td>
                          <td class="pr-4 py-0.5">{{ lp.description }}</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                  <p v-else class="text-xs text-gray-500">无</p>
                </div>
                <div class="mt-3 grid grid-cols-1 md:grid-cols-2 gap-3">
                  <Field label="parser_rules" :value="cfg.parser_rules || {}" pre />
                  <Field label="logs_parser_rules" :value="cfg.logs_parser_rules || {}" pre />
                  <Field class="md:col-span-2" label="logs_display_format" :value="cfg.logs_display_format || ''" />
                </div>
              </div>
            </div>
            <div v-else class="text-sm text-gray-500">暂无解析配置</div>
          </div>
        </div>
        <div v-else class="text-sm text-gray-500">暂无币种配置信息</div>
      </div>

      <!-- 创建信息 -->
      <div class="card">
        <h3 class="text-lg font-medium text-gray-900 mb-4">创建信息</h3>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-500">创建交易</label>
            <p class="mt-1 text-sm text-gray-900 font-mono cursor-pointer hover:text-blue-600" @click="copyToClipboard(contract.creation_tx, $event)">{{ contract.creation_tx || 'N/A' }}</p>
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
        <button @click="loadContractData" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700">重试</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, defineComponent, h } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getContractByAddress } from '@/api/contracts'
import { getCoinConfigMaintenance } from '@/api/coinconfig'

// 本地展示组件 Field（只读）
const Field = defineComponent({
  name: 'Field',
  props: {
    label: { type: String, required: true },
    value: { type: [String, Number, Object, Array], default: '' },
    mono: { type: Boolean, default: false },
    pre: { type: Boolean, default: false },
    class: { type: String, default: '' }
  },
  setup(props) {
    const valueText = () => {
      if (props.value === null || props.value === undefined) return ''
      if (typeof props.value === 'string' || typeof props.value === 'number') return String(props.value)
      try { return JSON.stringify(props.value) } catch { return String(props.value) }
    }
    return () => h('div', { class: props.class || '' }, [
      h('label', { class: 'block text-sm font-medium text-gray-500 mb-1' }, props.label),
      props.pre
        ? h('pre', { class: 'text-xs text-gray-800 whitespace-pre-wrap bg-gray-50 p-2 rounded border' }, valueText())
        : h('p', { class: `mt-1 text-sm text-gray-900 ${props.mono ? 'font-mono break-all' : ''}` }, valueText())
    ])
  }
})

// 路由参数
const route = useRoute()
const router = useRouter()
const contractAddress = computed(() => route.params.address as string)

// 响应式数据
const contract = ref<any>(null)
const isLoading = ref(true)
const errorMessage = ref('')

// 币种配置
const coinConfig = ref<any | null>(null)
const parserConfigs = ref<any[]>([])

// 返回方法
const goBack = () => {
  // 检查是否有来源页面信息
  const from = route.query.from as string
  if (from) {
    // 如果有来源页面信息，跳转到指定页面
    router.push(from)
  } else {
    // 否则使用浏览器历史记录返回
    router.go(-1)
  }
}

// 加载合约数据
const loadContractData = async () => {
  try {
    isLoading.value = true
    errorMessage.value = ''
    const response: any = await getContractByAddress(contractAddress.value)
    if (response && response.success === true) {
      contract.value = response.data
      // 加载币种配置（基于合约地址）
      try {
        const cfgRes: any = await getCoinConfigMaintenance(contractAddress.value)
        if (cfgRes?.success) {
          coinConfig.value = cfgRes.data?.coin_config || null
          parserConfigs.value = cfgRes.data?.parser_configs || []
        } else {
          coinConfig.value = null
          parserConfigs.value = []
        }
      } catch {
        coinConfig.value = null
        parserConfigs.value = []
      }
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

// 解析型计算属性（提供安全默认值）
const parsedInterfaces = computed<string[]>(() => {
  const val = contract.value?.interfaces
  if (!val) return []
  try { return Array.isArray(val) ? val : JSON.parse(val) } catch { return [] }
})
const parsedMethods = computed<string[]>(() => {
  const val = contract.value?.methods
  if (!val) return []
  try { return Array.isArray(val) ? val : JSON.parse(val) } catch { return [] }
})
const parsedEvents = computed<string[]>(() => {
  const val = contract.value?.events
  if (!val) return []
  try { return Array.isArray(val) ? val : JSON.parse(val) } catch { return [] }
})
const parsedMetadata = computed<Record<string, any> | null>(() => {
  const val = contract.value?.metadata
  if (!val) return null
  try { return typeof val === 'object' ? val : JSON.parse(val) } catch { return null }
})

// 复制到剪贴板
const copyToClipboard = async (text: string, e?: MouseEvent) => {
  try {
    await navigator.clipboard.writeText(text)
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

const formatCoinType = (type: number) => {
  switch (type) {
    case 0:
      return '原生币'
    case 1:
      return 'ERC-20'
    case 2:
      return 'ERC-223'
    case 3:
      return 'ERC-721'
    case 4:
      return 'ERC-1155'
    default:
      return '未知'
  }
}

// 组件挂载
onMounted(async () => {
  await loadContractData()
})
</script>

<style scoped>
.card {
  @apply bg-white shadow-sm rounded-lg border border-gray-200 p-4;
}
</style>
