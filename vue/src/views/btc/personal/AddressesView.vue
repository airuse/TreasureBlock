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
        <div v-else-if="addressesList.length === 0" class="text-center py-8">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">暂无地址</h3>
          <p class="mt-1 text-sm text-gray-500">开始添加您的第一个地址</p>
          <div class="mt-6">
            <button
              @click="showAddAddressModal = true"
              class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-orange-600 hover:bg-orange-700"
            >
              添加地址
            </button>
          </div>
        </div>

        <!-- 地址概览 -->
        <div v-else class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div class="text-center">
            <div class="text-2xl font-bold text-blue-600">{{ addressCount }}</div>
            <div class="text-sm text-gray-500">管理地址</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-green-600">{{ totalBalance }}</div>
            <div class="text-sm text-gray-500">总余额 (聪)</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-purple-600">{{ totalUtxos }}</div>
            <div class="text-sm text-gray-500">总UTXO数</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 地址列表 -->
    <div v-if="!loading && addressesList.length > 0" class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">地址列表</h3>
        
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">地址</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">标签</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">余额 (BTC)</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">UTXO数量</th>
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
                  <div class="group relative inline-block">
                    <span class="text-gray-900">{{ address.label || '-' }}</span>
                    <div v-if="address.notes && address.notes.trim() !== ''" 
                         class="absolute left-0 mt-1 px-3 py-2 bg-gray-800 text-white text-xs rounded-lg opacity-0 group-hover:opacity-100 transition-opacity duration-200 z-10 min-w-[12rem] max-w-[28rem] break-words">
                      {{ address.notes }}
                      <div class="absolute -top-1 left-4 w-0 h-0 border-l-4 border-r-4 border-b-4 border-transparent border-b-gray-800"></div>
                    </div>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  {{ formatBtcBalance(address.balance || '0') }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                    {{ address.utxo_count || 0 }} UTXO
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="getStatusClass(address.is_active ? 'active' : 'inactive')" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                    {{ getStatusText(address.is_active ? 'active' : 'inactive') }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
                  <button
                    @click="refreshBalance(address)"
                    class="text-blue-600 hover:text-blue-900"
                  >
                    更新余额
                  </button>
                  <button
                    @click="viewUTXOs(address)"
                    class="text-purple-600 hover:text-purple-900"
                  >
                    查看UTXO
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
    <Teleport to="body">
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
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">类型</label>
                <select
                  v-model="newAddress.type"
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
                >
                  <option value="wallet">钱包</option>
                  <option value="exchange">交易所</option>
                  <option value="other">其他</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">备注</label>
                <textarea
                  v-model="newAddress.notes"
                  rows="3"
                  maxlength="500"
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
                  placeholder="可选备注信息（最多500字符）"
                ></textarea>
                <div class="mt-1 text-xs text-gray-500 text-right">
                  {{ (newAddress.notes || '').length }}/500
                </div>
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
              class="px-4 py-2 bg-orange-600 text-white rounded-md hover:bg-orange-700 disabled:opacity-50"
            >
              <span v-if="addingAddress">添加中...</span>
              <span v-else>添加</span>
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- 编辑地址模态框 -->
    <Teleport to="body">
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
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
                  placeholder="例如：主钱包"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">类型</label>
                <select
                  v-model="editingAddress.type"
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
                >
                  <option value="wallet">钱包</option>
                  <option value="exchange">交易所</option>
                  <option value="other">其他</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">备注</label>
                <textarea
                  v-model="editingAddress.notes"
                  rows="3"
                  maxlength="500"
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-orange-500"
                  placeholder="可选备注信息（最多500字符）"
                ></textarea>
                <div class="mt-1 text-xs text-gray-500 text-right">
                  {{ (editingAddress.notes || '').length }}/500
                </div>
              </div>
              <div>
                <label class="flex items-center">
                  <input
                    v-model="editingAddress.isActive"
                    type="checkbox"
                    class="rounded border-gray-300 text-orange-600 shadow-sm focus:border-orange-300 focus:ring focus:ring-orange-200 focus:ring-opacity-50"
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
              class="px-4 py-2 bg-orange-600 text-white rounded-md hover:bg-orange-700 disabled:opacity-50"
            >
              <span v-if="updatingAddress">更新中...</span>
              <span v-else>更新</span>
            </button>
          </div>
        </div>
      </div>
    </Teleport>

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
  deletePersonalAddress, 
  refreshPersonalAddressBalance 
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
const showEditAddressModal = ref(false)
const addingAddress = ref(false)
const updatingAddress = ref(false)


// 新地址表单
const newAddress = ref<CreatePersonalAddressRequest>({
  address: '',
  label: '',
  type: 'wallet',
  notes: '',
  chain: 'btc'
})

// 编辑地址表单
const editingAddress = ref<PersonalAddressDetail>({
  id: 0,
  address: '',
  label: '',
  balance: '0',
  transactionCount: 0,
  status: 'active',
  createdAt: '',
  updatedAt: '',
  type: 'wallet',
  isActive: true,
  notes: '',
  balanceHeight: 0
})


// 地址列表
const addressesList = ref<PersonalAddressItem[]>([])

// 计算属性
const addressCount = computed(() => addressesList.value.length)

const totalBalance = computed(() => {
  return addressesList.value.reduce((sum, addr) => {
    const balance = parseFloat(addr.balance || '0')
    return sum + balance
  }, 0).toFixed(0)
})

const totalUtxos = computed(() => {
  return addressesList.value.reduce((sum, addr) => sum + (addr.utxo_count || 0), 0)
})

// 计算属性
const canAddAddress = computed(() => {
  return newAddress.value.address && newAddress.value.label
})

const canUpdateAddress = computed(() => {
  return editingAddress.value.label
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

// 格式化地址显示
const formatAddress = (address: string) => {
  if (address.length <= 10) return address
  return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`
}

// 格式化BTC余额（从satoshi转换为BTC）
const formatBtcBalance = (balance: string) => {
  if (!balance || balance === '0') return '0.00000000'
  const satoshi = parseFloat(balance)
  if (satoshi === 0) return '0.00000000'
  // 1 BTC = 100,000,000 satoshi
  const btc = satoshi / 100000000
  return btc.toFixed(8)
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
    const response = await getPersonalAddresses("btc")
    if (response.success) {
      addressesList.value = response.data || []
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

// 添加地址
const addAddress = async () => {
  if (!canAddAddress.value) return
  
  addingAddress.value = true
  try {
    // 为BTC地址添加chain字段
    const addressData = {
      ...newAddress.value,
      chain: 'btc'
    }
    
    const response = await createPersonalAddress(addressData)
    if (response.success) {
      showSuccess('地址添加成功')
      showAddAddressModal.value = false
      newAddress.value = { address: '', label: '', type: 'wallet', notes: '', chain: 'btc' }
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
    balance: address.balance || '0',
    transactionCount: address.transaction_count,
    status: address.is_active ? 'active' : 'inactive',
    createdAt: address.created_at,
    updatedAt: address.updated_at,
    type: address.type,
    isActive: address.is_active,
    notes: address.notes || '',
    balanceHeight: address.balance_height
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

// 刷新余额
const refreshBalance = async (address: PersonalAddressItem) => {
  try {
    const response = await refreshPersonalAddressBalance(address.id)
    if (!response.success) throw new Error(response.message || 'error')
    showSuccess('余额刷新成功')
    await loadAddresses()
  } catch (e) {
    showError('余额刷新失败')
  }
}


// 查看UTXO
const viewUTXOs = (address: PersonalAddressItem) => {
  // 跳转到地址UTXO详情页面
  router.push({
    path: '/btc/address-transactions',
    query: { address: address.address }
  })
}

onMounted(() => {
  loadAddresses()
})
</script>
