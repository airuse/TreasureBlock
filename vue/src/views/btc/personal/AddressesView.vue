<template>
  <div class="space-y-6">
    <!-- 页面头部 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <div class="flex items-center justify-between">
          <div>
            <h1 class="text-2xl font-bold text-gray-900">个人地址</h1>
            <p class="mt-1 text-sm text-gray-500">管理您的个人地址和创建交易</p>
          </div>
          <div class="flex items-center space-x-2">
            <div class="w-3 h-3 bg-orange-500 rounded-full"></div>
            <span class="text-sm text-gray-600">BTC 网络</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 地址概览 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg leading-6 font-medium text-gray-900">个人地址</h3>
          <button
            @click="showAddAddressModal = true"
            class="px-4 py-2 bg-orange-600 text-white text-sm font-medium rounded-md hover:bg-orange-700 transition-colors"
          >
            添加地址
          </button>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div class="text-center">
            <div class="text-2xl font-bold text-blue-600">{{ addressCount }}</div>
            <div class="text-sm text-gray-500">管理地址</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-green-600">{{ totalBalance }}</div>
            <div class="text-sm text-gray-500">总余额 (BTC)</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-purple-600">{{ totalTransactions }}</div>
            <div class="text-sm text-gray-500">总交易数</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 地址列表 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">地址列表</h3>
        
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">地址</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">标签</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">余额 (BTC)</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">交易数</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="address in addressesList" :key="address.id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <div class="flex items-center space-x-2">
                    <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono">
                      {{ address.address.substring(0, 10) }}...{{ address.address.substring(address.address.length - 8) }}
                    </code>
                    <button
                      @click="copyToClipboard(address.address)"
                      class="text-gray-400 hover:text-gray-600"
                      title="复制地址"
                    >
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                      </svg>
                    </button>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {{ address.label || '-' }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  {{ address.balance }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {{ address.transactionCount }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="getStatusClass(address.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                    {{ getStatusText(address.status) }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
                  <button
                    @click="createTransaction(address)"
                    class="text-blue-600 hover:text-blue-900"
                  >
                    创建交易
                  </button>
                  <button
                    @click="viewTransactions(address)"
                    class="text-green-600 hover:text-green-900"
                  >
                    查看交易
                  </button>
                  <button
                    @click="editAddress(address)"
                    class="text-gray-600 hover:text-gray-900"
                  >
                    编辑
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- 添加地址模态框 -->
    <div v-if="showAddAddressModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-medium text-gray-900">添加地址</h3>
        </div>
        
        <div class="px-6 py-4">
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">地址</label>
              <input
                v-model="newAddress.address"
                type="text"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
                placeholder="1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">标签</label>
              <input
                v-model="newAddress.label"
                type="text"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
                placeholder="例如：主钱包"
              />
            </div>
          </div>
        </div>
        
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
          <button
            @click="showAddAddressModal = false"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            取消
          </button>
          <button
            @click="addAddress"
            class="px-4 py-2 bg-orange-600 text-white rounded-md hover:bg-orange-700"
          >
            添加
          </button>
        </div>
      </div>
    </div>

    <!-- 创建交易模态框 -->
    <div v-if="showCreateTransactionModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-medium text-gray-900">创建交易</h3>
        </div>
        
        <div class="px-6 py-4">
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">发送地址</label>
              <input
                v-model="newTransaction.fromAddress"
                type="text"
                class="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50"
                readonly
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">接收地址</label>
              <input
                v-model="newTransaction.toAddress"
                type="text"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
                placeholder="1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">金额 (BTC)</label>
              <input
                v-model="newTransaction.amount"
                type="number"
                step="0.00000001"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
                placeholder="0.001"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">手续费 (sat/byte)</label>
              <input
                v-model="newTransaction.feeRate"
                type="number"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
                placeholder="5"
              />
            </div>
            <div class="bg-yellow-50 border border-yellow-200 rounded-md p-3">
              <div class="flex">
                <div class="flex-shrink-0">
                  <svg class="h-5 w-5 text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
                  </svg>
                </div>
                <div class="ml-3">
                  <p class="text-sm text-yellow-800">
                    创建交易需要消耗 1 TB 收益。当前可用收益：{{ availableEarnings }} TB
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
          <button
            @click="showCreateTransactionModal = false"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            取消
          </button>
          <button
            @click="submitTransaction"
            :disabled="!canCreateTransaction"
            class="px-4 py-2 bg-orange-600 text-white rounded-md hover:bg-orange-700 disabled:opacity-50"
          >
            创建交易
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { PersonalAddress } from '@/types'

// 响应式数据
const showAddAddressModal = ref(false)
const showCreateTransactionModal = ref(false)
const addressCount = ref(2)
const totalBalance = ref(0.15)
const totalTransactions = ref(8)
const availableEarnings = ref(6.8)

// 新地址表单
const newAddress = ref({
  address: '',
  label: ''
})

// 新交易表单
const newTransaction = ref({
  fromAddress: '',
  toAddress: '',
  amount: '',
  feeRate: '5'
})

// 地址列表
const addressesList = ref([
  {
    id: 1,
    address: '1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa',
    label: '主钱包',
    balance: 0.1,
    transactionCount: 5,
    status: 'active'
  },
  {
    id: 2,
    address: '1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2',
    label: '交易钱包',
    balance: 0.05,
    transactionCount: 3,
    status: 'active'
  }
])

// 计算属性
const canCreateTransaction = computed(() => {
  return availableEarnings.value >= 1 && 
         newTransaction.value.toAddress && 
         newTransaction.value.amount && 
         newTransaction.value.feeRate
})

// 获取状态样式
const getStatusClass = (status: string) => {
  switch (status) {
    case 'active': return 'bg-green-100 text-green-800'
    case 'inactive': return 'bg-gray-100 text-gray-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

// 获取状态文本
const getStatusText = (status: string) => {
  switch (status) {
    case 'active': return '活跃'
    case 'inactive': return '非活跃'
    default: return '未知'
  }
}

// 复制到剪贴板
const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text).then(() => {
    // 可以添加成功提示
    console.log('已复制到剪贴板')
  })
}

// 添加地址
const addAddress = () => {
  // TODO: 调用API添加地址
  console.log('添加地址:', newAddress.value)
  showAddAddressModal.value = false
  newAddress.value = { address: '', label: '' }
}

// 创建交易
const createTransaction = (address: PersonalAddress) => {
  newTransaction.value.fromAddress = address.address
  showCreateTransactionModal.value = true
}

// 提交交易
const submitTransaction = () => {
  // TODO: 调用API创建交易
  console.log('创建交易:', newTransaction.value)
  showCreateTransactionModal.value = false
  newTransaction.value = { fromAddress: '', toAddress: '', amount: '', feeRate: '5' }
}

// 查看交易
const viewTransactions = (address: PersonalAddress) => {
  // TODO: 跳转到交易历史页面，筛选该地址
  console.log('查看地址交易:', address.address)
}

// 编辑地址
const editAddress = (address: PersonalAddress) => {
  // TODO: 实现编辑地址功能
  console.log('编辑地址:', address)
}

onMounted(() => {
  // 加载数据
})
</script>
