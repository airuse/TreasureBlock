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
        <div class="grid grid-cols-1 md:grid-cols-4 lg:grid-cols-8 gap-4">
          <div class="text-center">
            <div class="text-2xl font-bold text-gray-600">{{ totalTransactions }}</div>
            <div class="text-sm text-gray-500">总交易</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-gray-500">{{ draftCount }}</div>
            <div class="text-sm text-gray-500">草稿</div>
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
            <div class="text-2xl font-bold text-purple-600">{{ packedCount }}</div>
            <div class="text-sm text-gray-500">已打包</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-green-600">{{ confirmedCount }}</div>
            <div class="text-sm text-gray-500">已确认</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-red-600">{{ failedCount }}</div>
            <div class="text-sm text-gray-500">失败</div>
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
              <option value="draft">草稿</option>
              <option value="unsigned">未签名</option>
              <option value="unsent">未发送</option>
              <option value="in_progress">在途</option>
              <option value="packed">已打包</option>
              <option value="confirmed">已确认</option>
              <option value="failed">失败</option>
            </select>
            <button
              @click="showCreateModal = true"
              class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 transition-colors"
            >
              新建交易
            </button>
          </div>
        </div>
        
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">交易哈希</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">交易类型</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">发送地址</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">接收地址</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">金额</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">创建时间</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="tx in filteredTransactions" :key="tx.id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <code v-if="tx.tx_hash" 
                        class="bg-gray-100 px-2 py-1 rounded text-xs font-mono cursor-pointer hover:bg-gray-200 transition-colors"
                        :title="tx.tx_hash"
                        @click="copyToClipboard(tx.tx_hash)">
                    {{ tx.tx_hash.substring(0, 10) + '...' + tx.tx_hash.substring(tx.tx_hash.length - 8) }}
                  </code>
                  <span v-else class="text-gray-400">未生成</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <div class="flex flex-col">
                    <span class="font-medium">{{ getTransactionTypeText(tx) }}</span>
                    <span v-if="tx.transaction_type === 'token' && tx.token_contract_address" 
                          class="text-xs text-gray-500 font-mono cursor-pointer hover:text-gray-700 transition-colors"
                          :title="tx.token_contract_address"
                          @click="copyToClipboard(tx.token_contract_address)">
                      {{ tx.token_contract_address.substring(0, 8) }}...{{ tx.token_contract_address.substring(tx.token_contract_address.length - 6) }}
                    </span>
                    <span v-if="tx.contract_operation_type" class="text-xs text-blue-600">
                      {{ getContractOperationText(tx.contract_operation_type) }}
                    </span>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono cursor-pointer hover:bg-gray-200 transition-colors" 
                        :title="tx.from_address"
                        @click="copyToClipboard(tx.from_address)">
                    {{ tx.from_address.substring(0, 10) }}...{{ tx.from_address.substring(tx.from_address.length - 8) }}
                  </code>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono cursor-pointer hover:bg-gray-200 transition-colors" 
                        :title="tx.to_address"
                        @click="copyToClipboard(tx.to_address)">
                    {{ tx.to_address.substring(0, 10) }}...{{ tx.to_address.substring(tx.to_address.length - 8) }}
                  </code>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  <div class="flex flex-col">
                    <span>{{ formatAmount(tx.amount, tx.symbol, tx.token_decimals) }} {{ tx.symbol }}</span>
                    <span v-if="tx.transaction_type === 'token' && tx.token_name" class="text-xs text-gray-500">
                      {{ tx.token_name }}
                    </span>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="getStatusClass(tx.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                    {{ getStatusText(tx.status) }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {{ formatTime(tx.created_at) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
                  <button
                    v-if="tx.status === 'draft' || tx.status === 'unsigned'"
                    @click="editTransaction(tx)"
                    class="text-indigo-600 hover:text-indigo-900"
                  >
                    编辑
                  </button>
                  <button
                    v-if="tx.status === 'draft' || tx.status === 'unsigned'"
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
                    v-if="tx.status === 'in_progress' || tx.status === 'packed' || tx.status === 'confirmed' || tx.status === 'failed'"
                    @click="viewTransaction(tx)"
                    class="text-purple-600 hover:text-purple-900"
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

    <!-- 新建交易模态框 -->
    <CreateTransactionModal
      :show="showCreateModal"
      :isEditMode="isEditMode"
      :transaction="selectedTransaction"
      @close="handleModalClose"
      @created="handleTransactionCreated"
      @updated="handleTransactionUpdated"
    />

    <!-- 发送交易模态框 -->
    <SendTransactionModal
      v-if="selectedTransaction"
      :show="showSendModal"
      :transaction="selectedTransaction"
      @close="showSendModal = false"
      @sent="handleTransactionSent"
    />

    <!-- QR码预览模态框 -->
    <div v-if="showQRModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-medium text-gray-900">交易QR码</h3>
            <button
              @click="showQRModal = false"
              class="text-gray-400 hover:text-gray-600 transition-colors"
            >
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
        
        <div class="px-6 py-4">
          <div class="text-center">
            <div class="mb-4">
              <h4 class="text-md font-medium text-gray-900 mb-2">交易数据</h4>
            </div>
            
            <div class="flex justify-center mb-4">
              <div v-if="qrCodeDataURL" class="bg-white p-4 rounded-lg border-2 border-gray-200">
                <img :src="qrCodeDataURL" alt="交易QR码" class="max-w-full h-auto" />
              </div>
              <div v-else class="bg-gray-100 p-8 rounded-lg">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
                <p class="text-sm text-gray-500 mt-2">生成中...</p>
              </div>
            </div>
            
            <div class="text-left bg-gray-50 p-4 rounded-lg mb-4">
              <h5 class="text-sm font-medium text-gray-900 mb-2">交易信息：</h5>
              <div class="text-xs text-gray-600 space-y-1">
                <div>交易ID: {{ selectedQRTransaction?.id }}</div>
                <div>链类型: {{ selectedQRTransaction?.chain?.toUpperCase() }}</div>
                <div>币种: {{ selectedQRTransaction?.symbol }}</div>
                <div>状态: {{ getStatusText(selectedQRTransaction?.status || '') }}</div>
                <div>创建时间: {{ formatTime(selectedQRTransaction?.created_at) }}</div>
              </div>
            </div>
          </div>
        </div>
        
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
          <button
            @click="showQRModal = false"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            关闭
          </button>
          <button
            @click="downloadQRCode"
            :disabled="!qrCodeDataURL"
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            下载QR码
          </button>
        </div>
      </div>
    </div>

    <!-- 导入签名模态框 -->
    <div v-if="showImportModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-medium text-gray-900">导入签名数据</h3>
        </div>
        
        <div class="px-6 py-4">
          <div class="space-y-4">
            <!-- 选择交易 -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">选择要导入签名的交易</label>
              <select v-model="selectedImportTransactionId" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="">请选择交易</option>
                <option v-for="tx in transactionsList.filter(t => t.status === 'unsigned')" :key="tx.id" :value="tx.id">
                  ID: {{ tx.id }} - {{ tx.from_address.substring(0, 10) }}... → {{ tx.to_address.substring(0, 10) }}... ({{ tx.amount }} {{ tx.symbol }})
                </option>
              </select>
            </div>
            
            <!-- 签名数据 -->
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
                    支持导入签名数据：完整的签名交易字符串或包含v,r,s字段的JSON格式。导入后交易状态将变为"未发送"
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
            :disabled="!importSignature.trim() || !selectedImportTransactionId"
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
import type { UserTransaction, UserTransactionStatsResponse } from '@/types'
import CreateTransactionModal from '@/components/eth/personal/CreateTransactionModal.vue'
import SendTransactionModal from '@/components/eth/personal/SendTransactionModal.vue'
import { getUserTransactions, getUserTransactionStats, exportTransaction as exportTransactionAPI, sendTransaction as sendTransactionAPI, importSignature as importSignatureAPI } from '@/api/user-transactions'

// 响应式数据
const showCreateModal = ref(false)
const showImportModal = ref(false)
const showSendModal = ref(false)
const selectedStatus = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const totalItems = ref(0)
const totalPages = ref(0)
const importSignature = ref('')
const selectedTransaction = ref<UserTransaction | null>(null)
const selectedImportTransactionId = ref<number | ''>('')
const isEditMode = ref(false) // 是否为编辑模式

// QR码相关状态
const showQRModal = ref(false)
const qrCodeDataURL = ref<string>('')
const selectedQRTransaction = ref<UserTransaction | null>(null)

// 交易统计
const totalTransactions = ref(0)
const unsignedCount = ref(0)
const unsentCount = ref(0)
const inProgressCount = ref(0)
const confirmedCount = ref(0)
const draftCount = ref(0)
const packedCount = ref(0)
const failedCount = ref(0)

// 交易列表
const transactionsList = ref<UserTransaction[]>([])

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
    case 'draft': return 'bg-gray-100 text-gray-800'
    case 'unsigned': return 'bg-yellow-100 text-yellow-800'
    case 'unsent': return 'bg-blue-100 text-blue-800'
    case 'in_progress': return 'bg-orange-100 text-orange-800'
    case 'packed': return 'bg-purple-100 text-purple-800'
    case 'confirmed': return 'bg-green-100 text-green-800'
    case 'failed': return 'bg-red-100 text-red-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

// 获取状态文本
const getStatusText = (status: string) => {
  switch (status) {
    case 'draft': return '草稿'
    case 'unsigned': return '未签名'
    case 'unsent': return '未发送'
    case 'in_progress': return '在途'
    case 'packed': return '已打包'
    case 'confirmed': return '已确认'
    case 'failed': return '失败'
    default: return '未知'
  }
}

// 获取交易类型文本
const getTransactionTypeText = (tx: UserTransaction) => {
  // 如果是查询余额操作，显示为"查询余额"
  if (tx.contract_operation_type === 'balanceOf') {
    return `${tx.symbol} 查询余额`
  }
  
  if (tx.transaction_type === 'coin' || tx.transaction_type === 'native') {
    return 'ETH 转账'
  } else if (tx.transaction_type === 'token') {
    return `${tx.symbol} 代币转账`
  } else if (tx.symbol === 'ETH') {
    return 'ETH 转账'
  } else {
    return `${tx.symbol} 代币转账`
  }
}

// 获取合约操作类型文本
const getContractOperationText = (type: string) => {
  switch (type) {
    case 'transfer': return '转账'
    case 'approve': return '授权'
    case 'transferFrom': return '代币转移'
    case 'mint': return '铸造'
    case 'burn': return '销毁'
    case 'setApprovalForAll': return '设置授权'
    case 'transferOwnership': return '转让所有权'
    default: return type
  }
}

// 格式化时间
const formatTime = (timestamp: string | undefined) => {
  if (!timestamp) return '未知时间'
  return new Date(timestamp).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 格式化金额
const formatAmount = (amount: string, symbol: string, decimals: number | undefined) => {
  const numAmount = parseFloat(amount)
  if (isNaN(numAmount)) return amount
  
  console.log(`格式化金额: amount=${amount}, symbol=${symbol}, decimals=${decimals}, numAmount=${numAmount}`)
  
  // 如果明确提供了精度，使用提供的精度
  if (decimals !== undefined && decimals >= 0) {
    const factor = Math.pow(10, decimals)
    const readableAmount = numAmount / factor
    const result = readableAmount.toFixed(Math.min(decimals, 8))
    console.log(`使用提供精度: factor=${factor}, readableAmount=${readableAmount}, result=${result}`)
    return result
  }
  
  // 如果没有提供精度，根据币种和数值特征智能判断
  if (symbol === 'ETH') {
    // ETH使用18位精度
    const factor = Math.pow(10, 18)
    const readableAmount = numAmount / factor
    const result = readableAmount.toFixed(8)
    console.log(`ETH精度: factor=${factor}, readableAmount=${readableAmount}, result=${result}`)
    return result
  } else if (symbol === 'USDC' || symbol === 'USDT') {
    // USDC/USDT使用6位精度
    const factor = Math.pow(10, 6)
    const readableAmount = numAmount / factor
    const result = readableAmount.toFixed(6)
    console.log(`USDC/USDT精度: factor=${factor}, readableAmount=${readableAmount}, result=${result}`)
    return result
  } else if (symbol === 'DAI') {
    // DAI使用18位精度
    const factor = Math.pow(10, 18)
    const readableAmount = numAmount / factor
    const result = readableAmount.toFixed(8)
    console.log(`DAI精度: factor=${factor}, readableAmount=${readableAmount}, result=${result}`)
    return result
  } else {
    // 其他代币，尝试智能判断精度
    // 如果数值很大（超过10^12），可能是原始精度，需要转换
    if (numAmount > Math.pow(10, 12)) {
      // 尝试常见的精度：6, 8, 18
      const possibleDecimals = [6, 8, 18]
      for (const dec of possibleDecimals) {
        const factor = Math.pow(10, dec)
        const readableAmount = numAmount / factor
        // 如果转换后的数值在合理范围内（0.000001 到 1000000），使用这个精度
        if (readableAmount >= 0.000001 && readableAmount <= 1000000) {
          const result = readableAmount.toFixed(Math.min(dec, 8))
          console.log(`智能判断精度: 使用${dec}位精度, factor=${factor}, readableAmount=${readableAmount}, result=${result}`)
          return result
        }
      }
    }
    
    // 如果无法确定，直接返回原始值
    console.log(`无法确定精度，返回原始值: ${amount}`)
    return amount
  }
}

// 复制到剪贴板
const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    // 使用更友好的提示方式
    const toast = document.createElement('div')
    toast.className = 'fixed top-4 right-4 bg-green-500 text-white px-4 py-2 rounded-md shadow-lg z-50 transition-opacity duration-300'
    toast.textContent = '地址已复制到剪贴板！'
    document.body.appendChild(toast)
    
    // 3秒后自动消失
    setTimeout(() => {
      toast.style.opacity = '0'
      setTimeout(() => {
        document.body.removeChild(toast)
      }, 300)
    }, 3000)
  } catch (err) {
    console.error('复制失败:', err)
    // 降级方案：使用传统方法
    const textArea = document.createElement('textarea')
    textArea.value = text
    document.body.appendChild(textArea)
    textArea.select()
    try {
      document.execCommand('copy')
      alert('地址已复制到剪贴板！')
    } catch (fallbackErr) {
      alert('复制失败，请手动复制：' + text)
    }
    document.body.removeChild(textArea)
  }
}

// 导出交易
const exportTransaction = async (tx: UserTransaction) => {
  try {
    const response = await exportTransactionAPI(tx.id)
    if (response.success) {
      // 显示QR码预览模态框
      selectedQRTransaction.value = tx
      showQRModal.value = true
      qrCodeDataURL.value = '' // 重置QR码
      
      // 异步生成QR码
      generateQRCode(response.data, tx)
      
      console.log('导出交易成功:', response.data)
    } else {
      alert('导出交易失败: ' + response.message)
    }
  } catch (error) {
    console.error('导出交易失败:', error)
    alert('导出交易失败，请重试')
  }
}

// 生成QR码（用于预览）
const generateQRCode = async (transactionData: any, tx: UserTransaction) => {
  try {
    // 动态导入QRCode库
    const QRCode = await import('qrcode')
    
    // 创建精简的交易数据结构，只包含签名必需的核心字段
    const minimalTxData = createMinimalTransactionData(tx, transactionData)
    
    // 将精简数据转换为JSON字符串
    const transactionJson = JSON.stringify(minimalTxData, null, 0) // 不格式化，减少字符数
    
    
    
    console.log('极简交易数据:', minimalTxData)
    console.log('QR码数据长度:', transactionJson.length, '字符')
    console.log('数据结构说明: 只包含签名必需字段，Gas相关字段由签名程序自动填充')
    console.log('数据来源分析:', {
      nonce: {
        apiNonce: transactionData.nonce,
        txNonce: tx.nonce,
        finalNonce: minimalTxData.nonce,
        source: '从区块链实时获取'
      },
      chainId: {
        apiChainId: transactionData.chain_id,
        txChain: tx.chain,
        finalChainId: minimalTxData.chainId,
        source: transactionData.chain_id ? '后端保存' : '前端计算'
      },
      txData: {
        apiTxData: transactionData.tx_data,
        generatedData: generateContractData(tx, transactionData),
        finalData: minimalTxData.data,
        source: transactionData.tx_data ? '后端保存' : '前端生成'
      },
      accessList: {
        apiAccessList: transactionData.access_list,
        finalAccessList: minimalTxData.accessList,
        source: transactionData.access_list ? '后端保存' : '前端生成'
      }
    })
    
    if (minimalTxData.accessList) {
      console.log('包含AccessList: 用于优化ERC-20代币交易的Gas成本')
      console.log('AccessList详情:', minimalTxData.accessList)
    } else {
      console.log('不包含AccessList: ETH转账或简单操作')
    }
    
    console.log('示例数据结构:')
    console.log('ETH转账:', {
      id: 123,
      chainId: '1',
      nonce: 42, // 从区块链实时获取的nonce
      from: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6', // 发送地址，用于签名程序自动匹配私钥
      to: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
      value: '0xde0b6b3a7640000', // 十六进制格式 (1 ETH = 10^18 wei)
      data: '0x'
    })
    console.log('ERC-20查询余额:', {
      id: 124,
      chainId: '1',
      nonce: 43, // 从区块链实时获取的nonce
      from: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6', // 发送地址，用于签名程序自动匹配私钥
      to: '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48',
      value: '0x0', // 代币操作value为0
      data: '0x70a08231000000000000000000000000742d35cc6634c0532925a3b8d4c9db96c4b4d8b6' // balanceOf(address)
    })
    console.log('ERC-20代币转账:', {
      id: 125,
      chainId: '1',
      nonce: 44, // 从区块链实时获取的nonce
      from: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6', // 发送地址，用于签名程序自动匹配私钥
      to: '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48',
      value: '0x0', // 代币操作value为0
      data: '0xa9059cbb000000000000000000000000742d35cc6634c0532925a3b8d4c9db96c4b4d8b600000000000000000000000000000000000000000000000000000000000f4240', // transfer(address,uint256)
      accessList: '可选，用于优化Gas成本'
    })
    console.log('ERC-20授权转账:', {
      id: 126,
      chainId: '1',
      nonce: 45, // 从区块链实时获取的nonce
      from: '0x26248Ec61fC83a24F958faF435f8254ce65D08d9', // 发送地址，用于签名程序自动匹配私钥
      to: '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48',
      value: '0x0', // 代币操作value为0
      data: '0x23b872dd00000000000000000000000026248ec61fc83a24f958faf435f8254ce65d08d9000000000000000000000000320b0306fa5917cb12109d5404b5fd00056b37b500000000000000000000000000000000000000000000000000000000d09dc300', // transferFrom(address,address,uint256)
      accessList: '可选，用于优化Gas成本'
    })
    
    // 生成QR码配置 - 使用更高的错误纠正级别
    const qrOptions = {
      type: 'image/png' as const,
      quality: 0.92,
      margin: 4, // 增加边距
      color: {
        dark: '#000000',
        light: '#FFFFFF'
      },
      width: 512, // 增加尺寸提高识别率
      errorCorrectionLevel: 'H' as const // 使用最高错误纠正级别
    }
    
    // 生成QR码数据URL
    const qrDataURL = await QRCode.toDataURL(transactionJson, qrOptions)
    qrCodeDataURL.value = qrDataURL
    
    console.log('QR码生成成功:', qrDataURL)
  } catch (error) {
    console.error('生成QR码失败:', error)
    qrCodeDataURL.value = ''
    alert('QR码生成失败，请重试')
  }
}

// 创建精简的交易数据结构
const createMinimalTransactionData = (tx: UserTransaction, fullData: any) => {
  // 优先使用后端保存的完整数据，确保数据一致性
  const minimalData: any = {
    // 交易标识
    id: tx.id,
    
    // 链信息 - 优先使用后端保存的chainId
    chainId: fullData.chain_id || (tx.chain === 'eth' ? '1' : tx.chain),
    
    // 交易核心字段
    nonce: fullData.nonce || tx.nonce || 0, // 优先使用API返回的nonce
    from: tx.from_address, // 添加from字段用于签名程序自动匹配私钥
    to: tx.transaction_type === 'token' && tx.token_contract_address ? tx.token_contract_address : tx.to_address,
    value: tx.transaction_type === 'token' ? '0x0' : convertToHexString(tx.amount || '0'), // 代币转账value为0，ETH转账使用十六进制格式
    data: fullData.tx_data || generateContractData(tx, fullData) // 优先使用后端保存的tx_data
  }
  
  // 添加AccessList - 优先使用后端保存的accessList
  if (fullData.access_list && fullData.access_list !== '[]') {
    try {
      minimalData.accessList = JSON.parse(fullData.access_list)
    } catch (error) {
      console.warn('解析AccessList失败，使用前端生成:', error)
      // 如果解析失败，使用前端生成的AccessList
      if (tx.transaction_type === 'token' && tx.token_contract_address) {
        const accessList = generateAccessListForTokenTransfer(tx)
        if (accessList && accessList.length > 0) {
          minimalData.accessList = accessList
        }
      }
    }
  } else if (tx.transaction_type === 'token' && tx.token_contract_address) {
    // 如果后端没有AccessList，使用前端生成
    const accessList = generateAccessListForTokenTransfer(tx)
    if (accessList && accessList.length > 0) {
      minimalData.accessList = accessList
    }
  }
  
  return minimalData
}

// 转换金额为十六进制格式
const convertToHexString = (amount: string) => {
  // 如果是浮点数，转换为wei单位（18位精度）
  const numAmount = parseFloat(amount)
  if (isNaN(numAmount)) return '0x0'
  
  // 转换为wei单位（乘以10^18）
  const weiAmount = Math.floor(numAmount * Math.pow(10, 18))
  
  // 转换为十六进制字符串
  const hexString = weiAmount.toString(16)
  return '0x' + hexString
}

// 根据操作类型生成合约调用数据
const generateContractData = (tx: UserTransaction, fullData: any) => {
  // 如果有完整的data，优先使用
  if (fullData.data && fullData.data !== '0x') {
    return fullData.data
  }
  
  // 如果是代币交易，根据操作类型生成data
  if (tx.transaction_type === 'token' && tx.token_contract_address) {
    switch (tx.contract_operation_type) {
      case 'balanceOf':
        // balanceOf(address) 函数调用
        return generateBalanceOfData(tx.from_address)
        
      case 'transfer':
        // transfer(address,uint256) 函数调用
        return generateTransferData(tx.to_address, tx.amount)
        
      case 'approve':
        // approve(address,uint256) 函数调用
        return generateApproveData(tx.to_address, tx.amount)
        
      case 'transferFrom':
        // transferFrom(address,address,uint256) 函数调用
        return generateTransferFromData(tx.from_address, tx.to_address, tx.amount)
        
      default:
        return '0x'
    }
  }
  
  // ETH转账，data为空
  return '0x'
}

// 生成balanceOf函数调用数据
const generateBalanceOfData = (address: string) => {
  // balanceOf(address) 函数选择器: 0x70a08231
  const functionSelector = '0x70a08231'
  // 地址参数（32字节，右对齐）
  const addressParam = address.slice(2).padStart(64, '0')
  return functionSelector + addressParam
}

// 生成transfer函数调用数据
const generateTransferData = (toAddress: string, amount: string) => {
  // transfer(address,uint256) 函数选择器: 0xa9059cbb
  const functionSelector = '0xa9059cbb'
  // 接收地址参数（32字节，右对齐）
  const toParam = toAddress.slice(2).padStart(64, '0')
  // 金额参数（32字节，转换为十六进制wei）
  const amountHex = convertToHexString(amount)
  const amountParam = amountHex.slice(2).padStart(64, '0')
  return functionSelector + toParam + amountParam
}

// 生成approve函数调用数据
const generateApproveData = (spenderAddress: string, amount: string) => {
  // approve(address,uint256) 函数选择器: 0x095ea7b3
  const functionSelector = '0x095ea7b3'
  // 被授权者地址参数（32字节，右对齐）
  const spenderParam = spenderAddress.slice(2).padStart(64, '0')
  // 授权金额参数（32字节，转换为十六进制wei）
  const amountHex = convertToHexString(amount)
  const amountParam = amountHex.slice(2).padStart(64, '0')
  return functionSelector + spenderParam + amountParam
}

// 生成transferFrom函数调用数据
const generateTransferFromData = (fromAddress: string, toAddress: string, amount: string) => {
  // transferFrom(address,address,uint256) 函数选择器: 0x23b872dd
  const functionSelector = '0x23b872dd'
  // 发送者地址参数（32字节，右对齐）
  const fromParam = fromAddress.slice(2).padStart(64, '0')
  // 接收者地址参数（32字节，右对齐）
  const toParam = toAddress.slice(2).padStart(64, '0')
  // 金额参数（32字节，转换为十六进制wei）
  const amountHex = convertToHexString(amount)
  const amountParam = amountHex.slice(2).padStart(64, '0')
  return functionSelector + fromParam + toParam + amountParam
}

// 为代币转账生成AccessList
const generateAccessListForTokenTransfer = (tx: UserTransaction) => {
  if (!tx.token_contract_address) return null
  
  const accessList = []
  
  // 根据合约操作类型生成不同的AccessList
  switch (tx.contract_operation_type) {
    case 'transfer':
      // 标准transfer操作，通常只需要访问余额存储槽
      accessList.push({
        address: tx.token_contract_address,
        storageKeys: [
          // 发送者余额存储槽 (keccak256(abi.encodePacked(sender, balanceOf_slot)))
          `0x${keccak256(`0x${tx.from_address.slice(2).padStart(64, '0')}0000000000000000000000000000000000000000000000000000000000000002`).slice(2)}`,
          // 接收者余额存储槽
          `0x${keccak256(`0x${tx.to_address.slice(2).padStart(64, '0')}0000000000000000000000000000000000000000000000000000000000000002`).slice(2)}`
        ]
      })
      break
      
    case 'approve':
      // approve操作，需要访问allowance存储槽
      accessList.push({
        address: tx.token_contract_address,
        storageKeys: [
          // allowance存储槽 (keccak256(abi.encodePacked(owner, spender, allowance_slot)))
          `0x${keccak256(`0x${tx.from_address.slice(2).padStart(64, '0')}${tx.to_address.slice(2).padStart(64, '0')}0000000000000000000000000000000000000000000000000000000000000003`).slice(2)}`
        ]
      })
      break
      
    case 'transferFrom':
      // transferFrom操作，需要访问发送者、接收者余额和allowance
      accessList.push({
        address: tx.token_contract_address,
        storageKeys: [
          // 发送者余额
          `0x${keccak256(`0x${tx.from_address.slice(2).padStart(64, '0')}0000000000000000000000000000000000000000000000000000000000000002`).slice(2)}`,
          // 接收者余额
          `0x${keccak256(`0x${tx.to_address.slice(2).padStart(64, '0')}0000000000000000000000000000000000000000000000000000000000000002`).slice(2)}`,
          // allowance
          `0x${keccak256(`0x${tx.from_address.slice(2).padStart(64, '0')}${tx.to_address.slice(2).padStart(64, '0')}0000000000000000000000000000000000000000000000000000000000000003`).slice(2)}`
        ]
      })
      break
      
    default:
      // 其他操作类型，不添加AccessList
      return null
  }
  
  return accessList
}

// 使用crypto-js实现keccak256（用于生成存储槽）
const keccak256 = (input: string) => {
  try {
    // 动态导入crypto-js
    const CryptoJS = require('crypto-js')
    
    // 移除0x前缀并转换为字节数组
    const hexString = input.startsWith('0x') ? input.slice(2) : input
    const wordArray = CryptoJS.enc.Hex.parse(hexString)
    
    // 计算keccak256哈希
    const hash = CryptoJS.SHA3(wordArray, { outputLength: 256 })
    
    return '0x' + hash.toString(CryptoJS.enc.Hex)
  } catch (error) {
    console.warn('keccak256计算失败，使用占位符:', error)
    // 如果计算失败，返回占位符
    return '0x' + '0'.repeat(64)
  }
}

// 下载QR码
const downloadQRCode = () => {
  if (!qrCodeDataURL.value || !selectedQRTransaction.value) return
  
  const tx = selectedQRTransaction.value
  const link = document.createElement('a')
  link.href = qrCodeDataURL.value
  link.download = `transaction_${tx.id}_${tx.chain}_${tx.symbol}_qr.png`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  
  // 显示成功提示
  const toast = document.createElement('div')
  toast.className = 'fixed top-4 right-4 bg-green-500 text-white px-4 py-2 rounded-md shadow-lg z-50 transition-opacity duration-300'
  toast.textContent = 'QR码已下载！'
  document.body.appendChild(toast)
  
  setTimeout(() => {
    toast.style.opacity = '0'
    setTimeout(() => {
      document.body.removeChild(toast)
    }, 300)
  }, 3000)
}

// 发送交易
const sendTransaction = (tx: UserTransaction) => {
  selectedTransaction.value = tx
  showSendModal.value = true
}

// 查看交易
const viewTransaction = (tx: UserTransaction) => {
  // 显示交易详情
  console.log('查看交易详情:', tx)
  
  let details = `交易详情:

ID: ${tx.id}
状态: ${getStatusText(tx.status)}
链类型: ${tx.chain.toUpperCase()}
币种: ${tx.symbol}
${tx.contract_operation_type === 'balanceOf' ? '查询地址' : '发送地址'}: ${tx.from_address}
${tx.contract_operation_type === 'balanceOf' ? '' : `接收地址: ${tx.to_address}
金额: ${formatAmount(tx.amount, tx.symbol, tx.token_decimals)} ${tx.symbol}`}
Gas限制: ${tx.gas_limit || '未设置'}
Gas价格: ${tx.gas_price || '未设置'} Gwei
Nonce: ${tx.nonce || '自动获取'}
交易哈希: ${tx.tx_hash || '未生成'}
区块高度: ${tx.block_height || '未确认'}
确认数: ${tx.confirmations || 0}
备注: ${tx.remark || '无'}
创建时间: ${formatTime(tx.created_at)}
更新时间: ${formatTime(tx.updated_at)}`

  // 添加ERC-20相关信息
  if (tx.transaction_type === 'token') {
    details += `

=== ERC-20 代币信息 ===
交易类型: 代币转账
合约操作: ${getContractOperationText(tx.contract_operation_type || '')}
代币合约地址: ${tx.token_contract_address || '未设置'}
代币名称: ${tx.token_name || '未设置'}
代币精度: ${tx.token_decimals || '未设置'}`
  } else {
    details += `

=== 交易类型 ===
交易类型: ETH转账`
  }
  
  alert(details)
}

// 编辑交易
const editTransaction = (tx: UserTransaction) => {
  selectedTransaction.value = tx
  isEditMode.value = true
  showCreateModal.value = true // 使用新建交易模态框进行编辑
}

// 导入签名数据
const importSignatureData = async () => {
  try {
    if (!selectedImportTransactionId.value) {
      alert('请选择要导入签名的交易')
      return
    }
    
    const id = selectedImportTransactionId.value as number
    
    // 解析签名数据
    const signatureData = parseSignatureData(importSignature.value)
    if (!signatureData) {
      alert('签名数据格式错误，请检查数据格式')
      return
    }
    
    // 调用导入签名API
    const response = await importSignatureAPI(id, { 
      id, 
      signed_tx: signatureData.signedTx,
      v: signatureData.v,
      r: signatureData.r,
      s: signatureData.s
    })
    
    if (response.success) {
      console.log('导入签名成功:', response.data)
      alert('导入签名成功！')
      loadTransactions()
      loadTransactionStats()
      showImportModal.value = false
      importSignature.value = ''
      selectedImportTransactionId.value = ''
    } else {
      alert('导入签名失败: ' + response.message)
    }
  } catch (error) {
    console.error('导入签名失败:', error)
    alert('导入签名失败，请重试')
  }
}

// 解析签名数据
const parseSignatureData = (signatureText: string) => {
  try {
    // 尝试解析JSON格式的签名数据
    const data = JSON.parse(signatureText)
    
    // 检查是否包含签名字段
    if (data.v && data.r && data.s) {
      return {
        signedTx: data.signedTx || signatureText, // 如果有完整的签名交易，使用它
        v: data.v,
        r: data.r,
        s: data.s
      }
    }
    
    // 如果只是签名交易字符串
    if (typeof data === 'string' || data.signedTx) {
      return {
        signedTx: data.signedTx || data,
        v: null,
        r: null,
        s: null
      }
    }
    
    return null
  } catch (error) {
    // 如果不是JSON格式，假设是直接的签名交易字符串
    if (signatureText.startsWith('0x') && signatureText.length > 100) {
      return {
        signedTx: signatureText,
        v: null,
        r: null,
        s: null
      }
    }
    
    console.error('解析签名数据失败:', error)
    return null
  }
}

// 处理交易创建成功
const handleTransactionCreated = (transaction: any) => {
  console.log('交易创建成功:', transaction)
  // 刷新交易列表和统计
  loadTransactions()
  loadTransactionStats()
  isEditMode.value = false // 关闭编辑模式
  selectedTransaction.value = null // 清除选中的交易
}

// 处理交易发送成功
const handleTransactionSent = (transaction: any) => {
  console.log('交易发送成功:', transaction)
  // 刷新交易列表和统计
  loadTransactions()
  loadTransactionStats()
}

// 处理模态框关闭
const handleModalClose = () => {
  showCreateModal.value = false
  isEditMode.value = false
  selectedTransaction.value = null
}

// 处理交易更新
const handleTransactionUpdated = (transaction: any) => {
  console.log('交易更新成功:', transaction)
  // 刷新交易列表和统计
  loadTransactions()
  loadTransactionStats()
  isEditMode.value = false // 关闭编辑模式
  selectedTransaction.value = null // 清除选中的交易
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
  try {
    const response = await getUserTransactions({
      page: currentPage.value,
      page_size: pageSize.value,
      status: selectedStatus.value
    })
    
    if (response.success) {
      transactionsList.value = response.data.transactions
      totalItems.value = response.data.total
      totalPages.value = Math.ceil(totalItems.value / pageSize.value)
    }
  } catch (error) {
    console.error('加载交易数据失败:', error)
  }
}

// 加载交易统计
const loadTransactionStats = async () => {
  try {
    const response = await getUserTransactionStats()
    
    if (response.success) {
      const stats = response.data
      totalTransactions.value = stats.total_transactions
      draftCount.value = stats.draft_count
      unsignedCount.value = stats.unsigned_count
      unsentCount.value = stats.unsent_count
      inProgressCount.value = stats.in_progress_count
      packedCount.value = stats.packed_count
      confirmedCount.value = stats.confirmed_count
      failedCount.value = stats.failed_count
    }
  } catch (error) {
    console.error('加载交易统计失败:', error)
  }
}

// 监听状态筛选变化
watch(selectedStatus, () => {
  currentPage.value = 1
  loadTransactions()
})

// 监听模态框状态变化
watch(showCreateModal, (newVal) => {
  if (!newVal) {
    // 模态框关闭时重置编辑状态
    isEditMode.value = false
    selectedTransaction.value = null
  }
})

onMounted(() => {
  loadTransactions()
  loadTransactionStats()
})
</script>
