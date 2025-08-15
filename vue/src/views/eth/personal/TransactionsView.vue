<template>
  <div class="space-y-6">
    <!-- 页面头部 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <div class="flex items-center justify-between">
          <div>
            <h1 class="text-2xl font-bold text-gray-900">交易历史</h1>
            <p class="mt-1 text-sm text-gray-500">查看和管理您的交易记录</p>
          </div>
          <div class="flex items-center space-x-2">
            <div class="w-3 h-3 bg-blue-500 rounded-full"></div>
            <span class="text-sm text-gray-600">ETH 网络</span>
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
            <div class="text-2xl font-bold text-blue-600">{{ unsentCount }}</div>
            <div class="text-sm text-gray-500">未发送</div>
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
              <option value="unsent">未发送</option>
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
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">金额 (ETH)</th>
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
                    v-if="tx.status === 'unsent'"
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
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
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
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            导入签名
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import type { PersonalTransaction } from '@/types'

// 响应式数据
const showImportModal = ref(false)
const selectedStatus = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const totalItems = ref(0)
const totalPages = ref(0)
const importSignature = ref('')

// 交易统计
const totalTransactions = ref(25)
const unsignedCount = ref(3)
const unsentCount = ref(2)
const inProgressCount = ref(1)
const confirmedCount = ref(19)

// 交易列表
const transactionsList = ref<PersonalTransaction[]>([
  {
    id: 1,
    hash: '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef',
    from: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
    to: '0x8ba1f109551bD432803012645Hac136c22C177e9',
    fromAddress: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
    toAddress: '0x8ba1f109551bD432803012645Hac136c22C177e9',
    amount: 0.1,
    fee: 0.00042, // gasPrice * gasUsed / 1e9
    gasPrice: 20,
    gasUsed: 21000,
    status: 'confirmed',
    timestamp: new Date('2024-01-15T10:30:00Z'),
    confirmations: 12
  },
  {
    id: 2,
    hash: null,
    from: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
    to: '0x1234567890123456789012345678901234567890',
    fromAddress: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
    toAddress: '0x1234567890123456789012345678901234567890',
    amount: 0.05,
    fee: 0.00042,
    status: 'unsent',
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 4),
    confirmations: 0
  },
  {
    id: 3,
    hash: '0x1234567890abcdef1234567890abcdef12345678',
    from: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
    to: '0x8ba1f109551bD432803012645Hac136c22C177e9',
    fromAddress: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
    toAddress: '0x8ba1f109551bD432803012645Hac136c22C177e9',
    amount: 0.2,
    fee: 0.00042,
    status: 'in_progress',
    timestamp: new Date(Date.now() - 1000 * 60 * 30),
    confirmations: 0
  },
  {
    id: 4,
    hash: '0xabcdef1234567890abcdef1234567890abcdef12',
    from: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
    to: '0x1234567890123456789012345678901234567890',
    fromAddress: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
    toAddress: '0x1234567890123456789012345678901234567890',
    amount: 0.15,
    fee: 0.00042,
    status: 'packed',
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24),
    confirmations: 0
  },
  {
    id: 5,
    hash: '0x567890abcdef1234567890abcdef1234567890ab',
    from: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
    to: '0x8ba1f109551bD432803012645Hac136c22C177e9',
    fromAddress: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
    toAddress: '0x8ba1f109551bD432803012645Hac136c22C177e9',
    amount: 0.3,
    fee: 0.00042,
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
    case 'unsent': return 'bg-blue-100 text-blue-800'
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
    case 'unsent': return '未发送'
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
  totalItems.value = 25
  totalPages.value = Math.ceil(totalItems.value / pageSize.value)
}

// 监听状态筛选变化
watch(selectedStatus, () => {
  currentPage.value = 1
  loadTransactions()
})

onMounted(() => {
  loadTransactions()
})
</script>
