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
            <div class="w-3 h-3 bg-blue-500 rounded-full"></div>
            <span class="text-sm text-gray-600">ETH 网络</span>
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
            class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 transition-colors"
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
            <div class="text-2xl font-bold text-green-600">{{ totalBalance.toFixed(4) }}</div>
            <div class="text-sm text-gray-500">总余额 (ETH)</div>
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
        <div v-else-if="addressesList.length === 0" class="text-center py-8">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">暂无地址</h3>
          <p class="mt-1 text-sm text-gray-500">开始添加您的第一个地址</p>
          <div class="mt-6">
            <button
              @click="showAddAddressModal = true"
              class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
            >
              添加地址
            </button>
          </div>
        </div>

        <!-- 地址列表 -->
        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">地址</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">标签</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">类型</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">余额 (ETH)</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">交易数</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">创建高度</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="address in addressesList" :key="address.id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <div class="flex items-center space-x-2">
                    <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono">
                      {{ formatAddress(address.address) }}
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
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <span :class="getTypeClass(address.type)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                    {{ getTypeText(address.type) }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  {{ address.balance.toFixed(4) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {{ address.transaction_count }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-gray-100 text-gray-800">
                    #{{ address.created_height }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="getStatusClass(address.is_active ? 'active' : 'inactive')" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                    {{ getStatusText(address.is_active ? 'active' : 'inactive') }}
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
                  <button
                    @click="removeAddress(address)"
                    class="text-red-600 hover:text-red-900"
                  >
                    删除
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
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="0x..."
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">标签</label>
              <input
                v-model="newAddress.label"
                type="text"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="例如：主钱包"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">类型</label>
              <select
                v-model="newAddress.type"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="wallet">钱包</option>
                <option value="contract">合约</option>
                <option value="exchange">交易所</option>
                <option value="other">其他</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">备注</label>
              <textarea
                v-model="newAddress.notes"
                rows="3"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="可选备注信息"
              ></textarea>
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
            :disabled="!canAddAddress || addingAddress"
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            <span v-if="addingAddress">添加中...</span>
            <span v-else>添加</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 编辑地址模态框 -->
    <div v-if="showEditAddressModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-medium text-gray-900">编辑地址</h3>
        </div>
        
        <div class="px-6 py-4">
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">地址</label>
              <input
                v-model="editingAddress.address"
                type="text"
                class="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50"
                readonly
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">标签</label>
              <input
                v-model="editingAddress.label"
                type="text"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="例如：主钱包"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">类型</label>
              <select
                v-model="editingAddress.type"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="wallet">钱包</option>
                <option value="contract">合约</option>
                <option value="exchange">交易所</option>
                <option value="other">其他</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">备注</label>
              <textarea
                v-model="editingAddress.notes"
                rows="3"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="可选备注信息"
              ></textarea>
            </div>
            <div>
              <label class="flex items-center">
                <input
                  v-model="editingAddress.isActive"
                  type="checkbox"
                  class="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50"
                />
                <span class="ml-2 text-sm text-gray-700">启用地址</span>
              </label>
            </div>
          </div>
        </div>
        
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
          <button
            @click="showEditAddressModal = false"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            取消
          </button>
          <button
            @click="updateAddress"
            :disabled="!canUpdateAddress || updatingAddress"
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            <span v-if="updatingAddress">更新中...</span>
            <span v-else>更新</span>
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
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="0x..."
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">金额 (ETH)</label>
              <input
                v-model="newTransaction.amount"
                type="number"
                step="0.000001"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="0.001"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Gas价格 (Gwei)</label>
              <input
                v-model="newTransaction.gasPrice"
                type="number"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="20"
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
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
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
import { useRouter } from 'vue-router'
import { showSuccess, showError, showInfo } from '@/composables/useToast'
import { 
  createPersonalAddress, 
  getPersonalAddresses, 
  updatePersonalAddress, 
  deletePersonalAddress 
} from '@/api/personal-addresses'
import type { 
  PersonalAddressItem, 
  PersonalAddressDetail, 
  CreatePersonalAddressRequest, 
  UpdatePersonalAddressRequest 
} from '@/types/personal-address'

// 初始化路由
const router = useRouter()

// 响应式数据
const loading = ref(false)
const showAddAddressModal = ref(false)
const showCreateTransactionModal = ref(false)
const showEditAddressModal = ref(false)
const addingAddress = ref(false)
const updatingAddress = ref(false)

// 地址统计
const addressCount = ref(0)
const totalBalance = ref(0)
const totalTransactions = ref(0)
const availableEarnings = ref(8.2)

// 新地址表单
const newAddress = ref<CreatePersonalAddressRequest>({
  address: '',
  label: '',
  type: 'wallet',
  notes: ''
})

// 编辑地址表单
const editingAddress = ref<PersonalAddressDetail>({
  id: 0,
  address: '',
  label: '',
  balance: 0,
  transactionCount: 0,
  status: 'active',
  createdAt: '',
  updatedAt: '',
  type: 'wallet',
  isActive: true,
  notes: '',
  createdHeight: 0
})

// 新交易表单
const newTransaction = ref({
  fromAddress: '',
  toAddress: '',
  amount: '',
  gasPrice: '20'
})

// 地址列表
const addressesList = ref<PersonalAddressItem[]>([])

// 计算属性
const canAddAddress = computed(() => {
  return newAddress.value.address && newAddress.value.label
})

const canUpdateAddress = computed(() => {
  return editingAddress.value.label
})

const canCreateTransaction = computed(() => {
  return availableEarnings.value >= 1 && 
         newTransaction.value.toAddress && 
         newTransaction.value.amount && 
         newTransaction.value.gasPrice
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

// 获取类型样式
const getTypeClass = (type: string) => {
  switch (type) {
    case 'wallet': return 'bg-blue-100 text-blue-800'
    case 'contract': return 'bg-purple-100 text-purple-800'
    case 'exchange': return 'bg-orange-100 text-orange-800'
    case 'other': return 'bg-gray-100 text-gray-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

// 获取类型文本
const getTypeText = (type: string) => {
  switch (type) {
    case 'wallet': return '钱包'
    case 'contract': return '合约'
    case 'exchange': return '交易所'
    case 'other': return '其他'
    default: return '未知'
  }
}

// 格式化地址显示
const formatAddress = (address: string) => {
  if (address.length <= 10) return address
  return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`
}

// 复制到剪贴板
const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text).then(() => {
    showSuccess('地址已复制到剪贴板')
  }).catch(() => {
    showError('复制失败')
  })
}

// 加载地址列表
const loadAddresses = async () => {
  loading.value = true
  try {
    const response = await getPersonalAddresses()
    if (response.success) {
      addressesList.value = response.data || []
      updateStats()
    } else {
      showError(response.message || '获取地址列表失败')
    }
  } catch (error) {
    console.error('加载地址列表失败:', error)
    showError('获取地址列表失败')
  } finally {
    loading.value = false
  }
}

// 更新统计数据
const updateStats = () => {
  addressCount.value = addressesList.value.length
  totalBalance.value = addressesList.value.reduce((sum, addr) => sum + addr.balance, 0)
  totalTransactions.value = addressesList.value.reduce((sum, addr) => sum + addr.transaction_count, 0)
}

// 添加地址
const addAddress = async () => {
  if (!canAddAddress.value) return
  
  addingAddress.value = true
  try {
    const response = await createPersonalAddress(newAddress.value)
    if (response.success) {
      showSuccess('地址添加成功')
      showAddAddressModal.value = false
      newAddress.value = { address: '', label: '', type: 'wallet', notes: '' }
      await loadAddresses() // 重新加载列表
    } else {
      showError(response.message || '添加地址失败')
    }
  } catch (error) {
    console.error('添加地址失败:', error)
    showError('添加地址失败')
  } finally {
    addingAddress.value = false
  }
}

// 编辑地址
const editAddress = (address: PersonalAddressItem) => {
  editingAddress.value = {
    id: address.id,
    address: address.address,
    label: address.label,
    balance: address.balance,
    transactionCount: address.transaction_count,
    status: address.is_active ? 'active' : 'inactive',
    createdAt: address.created_at,
    updatedAt: address.updated_at,
    type: address.type,
    isActive: address.is_active,
    notes: '',
    createdHeight: address.created_height
  }
  showEditAddressModal.value = true
}

// 更新地址
const updateAddress = async () => {
  if (!canUpdateAddress.value) return
  
  updatingAddress.value = true
  try {
    const updateData: UpdatePersonalAddressRequest = {
      label: editingAddress.value.label,
      type: editingAddress.value.type,
      notes: editingAddress.value.notes,
      isActive: editingAddress.value.isActive
    }
    
    const response = await updatePersonalAddress(editingAddress.value.id, updateData)
    if (response.success) {
      showSuccess('地址更新成功')
      showEditAddressModal.value = false
      await loadAddresses() // 重新加载列表
    } else {
      showError(response.message || '更新地址失败')
    }
  } catch (error) {
    console.error('更新地址失败:', error)
    showError('更新地址失败')
  } finally {
    updatingAddress.value = false
  }
}

// 删除地址
const removeAddress = async (address: PersonalAddressItem) => {
  if (!confirm(`确定要删除地址 ${address.label} 吗？`)) return
  
  try {
    const response = await deletePersonalAddress(address.id)
    if (response.success) {
      showSuccess('地址删除成功')
      await loadAddresses() // 重新加载列表
    } else {
      showError(response.message || '删除地址失败')
    }
  } catch (error) {
    console.error('删除地址失败:', error)
    showError('删除地址失败')
  }
}

// 创建交易
const createTransaction = (address: PersonalAddressItem) => {
  newTransaction.value.fromAddress = address.address
  showCreateTransactionModal.value = true
}

// 提交交易
const submitTransaction = () => {
  // TODO: 调用API创建交易
  console.log('创建交易:', newTransaction.value)
  showSuccess('交易创建成功')
  showCreateTransactionModal.value = false
  newTransaction.value = { fromAddress: '', toAddress: '', amount: '', gasPrice: '20' }
}

// 查看交易
const viewTransactions = (address: PersonalAddressItem) => {
  // 跳转到地址交易详情页面
  console.log('查看地址交易:', address.address)
  console.log('Using router:', router) // 添加调试日志
  router.push({
    path: '/eth/address-transactions',
    query: { address: address.address }
  })
}

onMounted(() => {
  loadAddresses()
})
</script>
