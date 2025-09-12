<template>
  <div class="space-y-6">
    <!-- 页面头部 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <div class="flex items-center justify-between">
          <div>
            <h1 class="text-2xl font-bold text-gray-900">交易历史</h1>
            <p class="text-sm text-gray-500">查看和管理您的交易记录</p>
          </div>
          <div class="flex items-center space-x-2">
            <div class="w-3 h-3 bg-orange-500 rounded-full"></div>
            <span class="text-sm text-gray-600">BTC 网络</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 实时费率信息 -->
    <div v-if="feeLevels" class="bg-white shadow rounded-lg">
      <div class="px-4 py-3">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-lg leading-6 font-medium text-gray-900">实时费率信息</h3>
          <div class="text-sm text-gray-500">
            最后更新: {{ formatTime(new Date(feeLevels.normal.last_updated * 1000)) }}
          </div>
        </div>
        <div class="flex flex-col lg:flex-row gap-3">
          <!-- 左侧：费率信息 -->
          <div class="lg:w-80 flex-shrink-0">
            <div class="space-y-1.5">
              <!-- 慢速费率 -->
              <div class="border border-gray-200 rounded-lg p-2.5">
                <div class="flex items-center justify-between mb-1">
                  <h4 class="text-sm font-medium text-gray-900">慢速</h4>
                  <span class="text-xs text-gray-500">0.5x 倍率</span>
                </div>
                <div class="text-sm text-gray-600">
                  <span class="font-mono">{{ feeLevels.slow.max_priority_fee }} sat/vB</span>
                </div>
              </div>
              
              <!-- 普通费率 -->
              <div class="border border-orange-200 bg-orange-50 rounded-lg p-2.5">
                <div class="flex items-center justify-between mb-1">
                  <h4 class="text-sm font-medium text-orange-900">普通</h4>
                  <span class="text-xs text-orange-600">1.0x 倍率</span>
                </div>
                <div class="text-sm text-orange-800">
                  <span class="font-mono">{{ feeLevels.normal.max_priority_fee }} sat/vB</span>
                </div>
              </div>
              
              <!-- 快速费率 -->
              <div class="border border-gray-200 rounded-lg p-2.5">
                <div class="flex items-center justify-between mb-1">
                  <h4 class="text-sm font-medium text-gray-900">快速</h4>
                  <span class="text-xs text-gray-500">2.0x 倍率</span>
                </div>
                <div class="text-sm text-gray-600">
                  <span class="font-mono">{{ feeLevels.fast.max_priority_fee }} sat/vB</span>
                </div>
              </div>
            </div>
          </div>
          
          <!-- 右侧：趋势图 -->
          <div class="flex-1 min-w-0">
            <!-- 费率趋势图 -->
            <div class="space-y-4">
              <!-- 费率图表 -->
              <div class="relative">
                <div class="text-sm font-medium text-gray-700 mb-2">费率趋势</div>
                <div class="h-32">
                  <canvas ref="feeChartCanvas" class="w-full h-full cursor-crosshair"></canvas>
                </div>
                <!-- 费率工具提示 -->
                <div 
                  ref="feeTooltip" 
                  class="absolute bg-gray-800 text-white text-xs px-2 py-1 rounded shadow-lg pointer-events-none opacity-0 transition-opacity duration-200 z-10"
                  style="transform: translate(-50%, -100%); margin-top: -8px;"
                >
                  <div class="font-medium">费率</div>
                  <div class="text-gray-300">Value: <span class="text-white font-mono" id="tooltip-fee-value">0</span> sat/vB</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 交易概览 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">交易概览</h3>
        <div class="grid grid-cols-1 md:grid-cols-5 gap-6">
          <div class="text-center">
            <div class="text-2xl font-bold text-gray-600">{{ totalTransactions }}</div>
            <div class="text-sm text-gray-500">总交易</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-yellow-600">{{ unsignedCount }}</div>
            <div class="text-sm text-gray-500">未签名</div>
          </div>
          <div class="text-center">
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-orange-600">{{ inProgressCount }}</div>
            <div class="text-sm text-gray-500">在途</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-green-600">{{ confirmedCount }}</div>
            <div class="text-sm text-gray-500">已确认</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 交易列表 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg leading-6 font-medium text-gray-900">交易历史</h3>
          <div class="flex space-x-2">
            <select v-model="selectedStatus" class="border border-gray-300 rounded-md px-3 py-2 text-sm">
              <option value="">全部状态</option>
              <option value="unsigned">未签名</option>
              <option value="in_progress">在途</option>
              <option value="packed">已打包</option>
              <option value="confirmed">已确认</option>
            </select>
            <button
              @click="showImportModal = true"
              class="px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-md hover:bg-green-700 transition-colors"
            >
              导入签名
            </button>
          </div>
        </div>
        
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">交易哈希</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">发送地址</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">接收地址</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">金额 (BTC)</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">创建时间</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="tx in filteredTransactions" :key="tx.id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono">
                    {{ tx.hash ? tx.hash.substring(0, 10) + '...' + tx.hash.substring(tx.hash.length - 8) : '未生成' }}
                  </code>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono">
                    {{ tx.from.substring(0, 10) }}...{{ tx.from.substring(tx.from.length - 8) }}
                  </code>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono">
                    {{ tx.to.substring(0, 10) }}...{{ tx.to.substring(tx.to.length - 8) }}
                  </code>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  {{ tx.amount }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="getStatusClass(tx.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                    {{ getStatusText(tx.status) }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {{ formatTime(tx.timestamp) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
                  <button
                    v-if="tx.status === 'unsigned'"
                    @click="exportTransaction(tx)"
                    class="text-blue-600 hover:text-blue-900"
                  >
                    导出交易
                  </button>
                  <button
                    v-if="tx.status === 'in_progress'"
                    @click="sendTransaction(tx)"
                    class="text-green-600 hover:text-green-900"
                  >
                    发送交易
                  </button>
                  <button
                    v-if="tx.status === 'in_progress' || tx.status === 'packed'"
                    @click="viewTransaction(tx)"
                    class="text-purple-600 hover:text-purple-900"
                  >
                    查看详情
                  </button>
                  <button
                    v-if="tx.status === 'confirmed'"
                    @click="viewTransaction(tx)"
                    class="text-gray-600 hover:text-gray-900"
                  >
                    查看详情
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 分页 -->
        <div class="flex items-center justify-between mt-4">
          <div class="text-sm text-gray-700">
            显示第 {{ (currentPage - 1) * pageSize + 1 }} 到 {{ Math.min(currentPage * pageSize, totalItems) }} 条，共 {{ totalItems }} 条记录
          </div>
          <div class="flex space-x-2">
            <button
              @click="prevPage"
              :disabled="currentPage === 1"
              class="px-3 py-2 border border-gray-300 rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              上一页
            </button>
            <button
              @click="nextPage"
              :disabled="currentPage >= totalPages"
              class="px-3 py-2 border border-gray-300 rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              下一页
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 导入签名模态框 -->
    <div v-if="showImportModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-medium text-gray-900">导入签名数据</h3>
        </div>
        
        <div class="px-6 py-4">
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">签名数据</label>
              <textarea
                v-model="importSignature"
                rows="6"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
                placeholder="请粘贴从离线程序导出的签名数据..."
              ></textarea>
            </div>
            <div class="bg-blue-50 border border-blue-200 rounded-md p-3">
              <div class="flex">
                <div class="flex-shrink-0">
                  <svg class="h-5 w-5 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
                  </svg>
                </div>
                <div class="ml-3">
                  <p class="text-sm text-blue-800">
                    请确保签名数据格式正确，导入后交易状态将变为"未发送"
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
          <button
            @click="showImportModal = false"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            取消
          </button>
          <button
            @click="importSignatureData"
            :disabled="!importSignature.trim()"
            class="px-4 py-2 bg-orange-600 text-white rounded-md hover:bg-orange-700 disabled:opacity-50"
          >
            导入签名
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import type { PersonalTransaction } from '@/types'
import { useChainWebSocket } from '@/composables/useWebSocket'
import type { FeeLevels } from '@/types'

// 响应式数据
const showImportModal = ref(false)
const selectedStatus = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const totalItems = ref(0)
const totalPages = ref(0)
const importSignature = ref('')

// WebSocket相关
const { subscribeChainEvent } = useChainWebSocket('btc')
// 收集本组件的取消订阅函数，避免重复回调
const wsUnsubscribes: Array<() => void> = []

// 费率数据
const feeLevels = ref<FeeLevels | null>(null)
const networkCongestion = ref<string>('normal')

// 费率历史数据存储（用于折线图）
const feeHistory = ref<Array<{
  timestamp: number
  fee: number
}>>([])

// 图表相关
const feeChartCanvas = ref<HTMLCanvasElement | null>(null)
const feeTooltip = ref<HTMLDivElement | null>(null)

// 交易统计
const totalTransactions = ref(18)
const unsignedCount = ref(2)
const inProgressCount = ref(1)
const confirmedCount = ref(14)

// 交易列表
const transactionsList = ref<PersonalTransaction[]>([
  {
    id: 1,
    hash: '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef',
    from: '1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa',
    to: '1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2',
    fromAddress: '1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa',
    toAddress: '1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2',
    amount: 0.01,
    fee: 0.0001,
    status: 'confirmed',
    timestamp: new Date('2024-01-15T10:30:00Z'),
    confirmations: 6
  },
  {
    id: 2,
    hash: null,
    from: '1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa',
    to: '1FvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2',
    fromAddress: '1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa',
    toAddress: '1FvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2',
    amount: 0.02,
    fee: 0.0001,
    status: 'in_progress',
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 4),
    confirmations: 0
  },
  {
    id: 3,
    hash: '1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef',
    from: '1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa',
    to: '1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2',
    fromAddress: '1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa',
    toAddress: '1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2',
    amount: 0.1,
    fee: 0.0001,
    status: 'in_progress',
    timestamp: new Date(Date.now() - 1000 * 60 * 30),
    confirmations: 0
  },
  {
    id: 4,
    hash: 'abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890',
    from: '1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa',
    to: '1FvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2',
    fromAddress: '1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa',
    toAddress: '1FvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2',
    amount: 0.03,
    fee: 0.0001,
    status: 'packed',
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24),
    confirmations: 0
  },
  {
    id: 5,
    hash: '567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234',
    from: '1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa',
    to: '1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2',
    fromAddress: '1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa',
    toAddress: '1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2',
    amount: 0.08,
    fee: 0.0001,
    status: 'confirmed',
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 * 3),
    confirmations: 6
  }
])

// 计算属性
const filteredTransactions = computed(() => {
  if (!selectedStatus.value) {
    return transactionsList.value
  }
  return transactionsList.value.filter(tx => tx.status === selectedStatus.value)
})

// 获取状态样式
const getStatusClass = (status: string) => {
  switch (status) {
    case 'unsigned': return 'bg-gray-100 text-gray-800'
    case 'in_progress': return 'bg-yellow-100 text-yellow-800'
    case 'packed': return 'bg-orange-100 text-orange-800'
    case 'confirmed': return 'bg-green-100 text-green-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

// 获取状态文本
const getStatusText = (status: string) => {
  switch (status) {
    case 'unsigned': return '未签名'
    case 'in_progress': return '在途'
    case 'packed': return '已打包'
    case 'confirmed': return '已确认'
    default: return '未知'
  }
}

// 格式化时间
const formatTime = (timestamp: Date | undefined) => {
  if (!timestamp) return '未知时间'
  return timestamp.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 导出交易
const exportTransaction = (tx: PersonalTransaction) => {
  // TODO: 实现导出交易功能
  console.log('导出交易:', tx)
}

// 发送交易
const sendTransaction = (tx: PersonalTransaction) => {
  // TODO: 实现发送交易功能
  console.log('发送交易:', tx)
}

// 查看交易
const viewTransaction = (tx: PersonalTransaction) => {
  // TODO: 实现查看交易详情功能
  console.log('查看交易:', tx)
}

// 导入签名数据
const importSignatureData = () => {
  // TODO: 实现导入签名功能
  console.log('导入签名:', importSignature.value)
  showImportModal.value = false
  importSignature.value = ''
}

// 分页方法
const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    loadTransactions()
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    loadTransactions()
  }
}

// 加载交易数据
const loadTransactions = async () => {
  // TODO: 从API加载真实数据
  totalItems.value = 18
  totalPages.value = Math.ceil(totalItems.value / pageSize.value)
}

// 监听状态筛选变化
watch(selectedStatus, () => {
  currentPage.value = 1
  loadTransactions()
})

// 添加费率历史数据
const addFeeHistory = (feeData: FeeLevels) => {
  const now = Date.now()
  const historyItem = {
    timestamp: now,
    fee: parseFloat(feeData.normal.max_priority_fee || '0')
  }
  
  // 添加到历史数据
  feeHistory.value.push(historyItem)
  
  // 只保留最近20条记录
  if (feeHistory.value.length > 20) {
    feeHistory.value = feeHistory.value.slice(-20)
  }
  
  // 更新图表
  updateChart()
}

// 费率鼠标移动事件处理
const handleFeeMouseMove = (event: MouseEvent) => {
  if (!feeChartCanvas.value || !feeTooltip.value || feeHistory.value.length === 0) return
  
  const canvas = feeChartCanvas.value
  const rect = canvas.getBoundingClientRect()
  const x = event.clientX - rect.left
  const y = event.clientY - rect.top
  
  // 计算数据点索引
  const padding = { top: 10, right: 10, bottom: 20, left: 40 }
  const chartWidth = rect.width - padding.left - padding.right
  const dataIndex = Math.round(((x - padding.left) / chartWidth) * (feeHistory.value.length - 1))
  
  // 确保索引在有效范围内
  if (dataIndex >= 0 && dataIndex < feeHistory.value.length) {
    const data = feeHistory.value[dataIndex]
    
    // 更新工具提示内容
    const feeElement = document.getElementById('tooltip-fee-value')
    if (feeElement) feeElement.textContent = data.fee.toFixed(1)
    
    // 计算相对于父容器的位置
    const parentRect = feeTooltip.value.parentElement?.getBoundingClientRect()
    
    if (parentRect) {
      const relativeX = event.clientX - parentRect.left
      const relativeY = event.clientY - parentRect.top
      
      feeTooltip.value.style.left = relativeX + 'px'
      feeTooltip.value.style.top = (relativeY - 10) + 'px'
      feeTooltip.value.style.opacity = '1'
    }
  }
}

// 费率鼠标离开事件处理
const handleFeeMouseLeave = () => {
  if (feeTooltip.value) {
    feeTooltip.value.style.opacity = '0'
  }
}

// 绘制费率图表
const drawFeeChart = (canvas: HTMLCanvasElement, data: number[], color: string, title: string) => {
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  
  // 设置canvas尺寸
  const rect = canvas.getBoundingClientRect()
  canvas.width = rect.width * window.devicePixelRatio
  canvas.height = rect.height * window.devicePixelRatio
  ctx.scale(window.devicePixelRatio, window.devicePixelRatio)
  
  // 清空画布
  ctx.clearRect(0, 0, rect.width, rect.height)
  
  if (data.length === 0) return
  
  // 移除之前的鼠标事件监听器
  canvas.removeEventListener('mousemove', handleFeeMouseMove)
  canvas.removeEventListener('mouseleave', handleFeeMouseLeave)
  
  // 添加新的鼠标事件监听器
  canvas.addEventListener('mousemove', handleFeeMouseMove)
  canvas.addEventListener('mouseleave', handleFeeMouseLeave)
  
  // 计算数据范围（增加自适应 padding，使小幅波动也能看见）
  const rawMin = Math.min(...data)
  const rawMax = Math.max(...data)
  let range = rawMax - rawMin
  if (range < 0.5) range = 0.5 // 最小范围，避免看起来成一条直线
  const pad = range * 0.1
  const minValue = rawMin - pad
  const maxValue = rawMax + pad
  
  // 设置边距
  const padding = { top: 10, right: 10, bottom: 20, left: 50 }
  const chartWidth = rect.width - padding.left - padding.right
  const chartHeight = rect.height - padding.top - padding.bottom
  
  // 绘制背景网格
  ctx.strokeStyle = '#f3f4f6'
  ctx.lineWidth = 1
  for (let i = 0; i <= 4; i++) {
    const y = padding.top + (chartHeight / 4) * i
    ctx.beginPath()
    ctx.moveTo(padding.left, y)
    ctx.lineTo(padding.left + chartWidth, y)
    ctx.stroke()
  }
  
  // 绘制折线
  ctx.strokeStyle = color
  ctx.lineWidth = 2
  ctx.beginPath()
  
  data.forEach((value, index) => {
    const x = data.length === 1 
      ? padding.left + chartWidth / 2
      : padding.left + (chartWidth / (data.length - 1)) * index
    const y = padding.top + chartHeight - ((value - minValue) / (maxValue - minValue)) * chartHeight
    
    if (index === 0) {
      ctx.moveTo(x, y)
    } else {
      ctx.lineTo(x, y)
    }
  })
  
  ctx.stroke()
  
  // 绘制数据点
  ctx.fillStyle = color
  data.forEach((value, index) => {
    const x = data.length === 1 
      ? padding.left + chartWidth / 2
      : padding.left + (chartWidth / (data.length - 1)) * index
    const y = padding.top + chartHeight - ((value - minValue) / (maxValue - minValue)) * chartHeight
    
    ctx.beginPath()
    ctx.arc(x, y, 2, 0, 2 * Math.PI)
    ctx.fill()
  })
  
  // 绘制Y轴标签（1位小数 + 单位）
  ctx.fillStyle = '#6b7280'
  ctx.font = '10px sans-serif'
  ctx.textAlign = 'right'
  for (let i = 0; i <= 4; i++) {
    const value = minValue + ((maxValue - minValue) / 4) * (4 - i)
    const y = padding.top + (chartHeight / 4) * i
    ctx.fillText(value.toFixed(1), padding.left - 5, y + 3)
  }
  // Y轴单位
  ctx.textAlign = 'left'
  ctx.fillText('sat/vB', 4, padding.top + 10)
}

// 更新折线图
const updateChart = () => {
  if (feeHistory.value.length === 0) return
  
  // 绘制费率图表
  if (feeChartCanvas.value) {
    const feeData = feeHistory.value.map(item => item.fee)
    drawFeeChart(
      feeChartCanvas.value, 
      feeData, 
      '#f97316', // 橙色
      'BTC费率'
    )
  }
}

// WebSocket监听
const setupWebSocketListeners = () => {
  // 监听费率更新
  const unsubNetwork = subscribeChainEvent('network', (message) => {
    if (message.action === 'fee_update' && message.data) {
      console.log('收到BTC费率更新:', message.data)
      feeLevels.value = message.data as unknown as FeeLevels
      
      // 添加历史数据
      addFeeHistory(message.data as unknown as FeeLevels)
      
      if (feeLevels.value?.normal?.network_congestion) {
        networkCongestion.value = feeLevels.value.normal.network_congestion
      }
    }
  })
  wsUnsubscribes.push(unsubNetwork)
}

onMounted(() => {
  loadTransactions()
  setupWebSocketListeners()
  
  // 监听窗口大小变化，重新绘制图表
  window.addEventListener('resize', updateChart)
  
  // 确保DOM完全渲染后再次更新图表
  setTimeout(() => {
    updateChart()
  }, 100)
})

onUnmounted(() => {
  // 组件卸载时取消订阅，避免重复注册导致一次数据多次回调
  wsUnsubscribes.forEach(unsub => { try { unsub() } catch {} })
  wsUnsubscribes.length = 0
  
  // 移除窗口大小变化监听
  window.removeEventListener('resize', updateChart)
})
</script>
