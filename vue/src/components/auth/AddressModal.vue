<template>
  <Teleport to="body">
    <div v-if="isVisible" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-4xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        <!-- 头部 -->
        <div class="flex items-center justify-between p-6 border-b border-gray-200">
          <h2 class="text-xl font-semibold text-gray-900">地址管理</h2>
          <button
            @click="close"
            class="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- 内容 -->
        <div class="p-6 space-y-6">
          <!-- 添加新地址 -->
          <div class="bg-gray-50 p-4 rounded-lg">
            <h3 class="text-lg font-medium text-gray-900 mb-4">添加新地址</h3>
            
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">地址</label>
                <input
                  v-model="newAddressForm.address"
                  type="text"
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="0x..."
                />
              </div>
              
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">标签</label>
                <input
                  v-model="newAddressForm.label"
                  type="text"
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="例如：我的钱包"
                />
              </div>
              
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">类型</label>
                <select
                  v-model="newAddressForm.type"
                  class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="wallet">钱包地址</option>
                  <option value="contract">合约地址</option>
                  <option value="exchange">交易所地址</option>
                  <option value="other">其他</option>
                </select>
              </div>
            </div>
            
            <div class="mt-4">
              <button
                @click="addAddress"
                :disabled="!newAddressForm.address || isLoading"
                class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                {{ isLoading ? '添加中...' : '添加地址' }}
              </button>
            </div>
          </div>

          <!-- 地址列表 -->
          <div>
            <h3 class="text-lg font-medium text-gray-900 mb-4">我的地址</h3>
            
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200">
                <thead class="bg-gray-50">
                  <tr>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">地址</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">标签</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">类型</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">余额</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">交易数</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
                  </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                  <tr v-for="addr in addresses" :key="addr.id" class="hover:bg-gray-50">
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <div class="flex items-center space-x-2">
                        <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono">
                          {{ formatAddress(addr.address) }}
                        </code>
                        <button
                          @click="copyAddress(addr.address)"
                          class="text-blue-600 hover:text-blue-800"
                          title="复制地址"
                        >
                          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                          </svg>
                        </button>
                      </div>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <span v-if="addr.label" class="inline-flex px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded">
                        {{ addr.label }}
                      </span>
                      <span v-else class="text-gray-400">-</span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <span :class="getTypeClass(addr.type)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                        {{ getTypeText(addr.type) }}
                      </span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {{ formatBalance(addr.balance) }} ETH
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {{ addr.transaction_count.toLocaleString() }}
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
                      <button
                        @click="editAddress(addr)"
                        class="text-blue-600 hover:text-blue-900"
                      >
                        编辑
                      </button>
                      <button
                        @click="removeAddress(addr)"
                        class="text-red-600 hover:text-red-900"
                      >
                        删除
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            
            <div v-if="addresses.length === 0" class="text-center py-8 text-gray-500">
              暂无地址，请添加一个
            </div>
          </div>

          <!-- 说明 -->
          <div class="bg-green-50 p-4 rounded-lg">
            <h3 class="text-lg font-medium text-green-900 mb-2">说明</h3>
            <div class="text-sm text-green-800 space-y-2">
              <p>• 添加地址后，该地址的交易记录将在地址菜单中显示</p>
              <p>• 您可以给地址添加标签，方便识别和管理</p>
              <p>• 支持钱包地址、合约地址、交易所地址等多种类型</p>
              <p>• 删除地址后，相关交易记录将不再显示</p>
            </div>
          </div>

          <!-- 错误提示 -->
          <div v-if="error" class="text-red-600 text-sm bg-red-50 p-3 rounded-md">
            {{ error }}
          </div>

          <!-- 成功提示 -->
          <div v-if="success" class="text-green-600 text-sm bg-green-50 p-3 rounded-md">
            {{ success }}
          </div>
        </div>

        <!-- 底部按钮 -->
        <div class="flex justify-end p-6 border-t border-gray-200">
          <button
            @click="close"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            关闭
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { 
  getUserAddresses,
  createUserAddress,
  updateUserAddress,
  deleteUserAddress
} from '@/api/user'
import type { UserAddress } from '@/types/address'
import { showSuccess, showError } from '@/composables/useToast'

const props = defineProps<{
  isVisible: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const authStore = useAuthStore()

// 响应式数据
const isLoading = ref(false)
const error = ref('')
const success = ref('')

const addresses = ref<UserAddress[]>([])

// 表单数据
const newAddressForm = reactive({
  address: '',
  label: '',
  type: 'wallet' as const
})

// 监听模态框显示状态
watch(() => props.isVisible, (visible) => {
  if (visible) {
    loadAddresses()
    // 清空表单
    newAddressForm.address = ''
    newAddressForm.label = ''
    newAddressForm.type = 'wallet'
    // 清空提示信息
    error.value = ''
    success.value = ''
  }
})

// 关闭模态框
const close = () => {
  emit('close')
}

// 加载地址列表
const loadAddresses = async () => {
  try {
    isLoading.value = true
    const response = await getUserAddresses({ token: authStore.loginToken || '' })
    
    if (response && response.success === true) {
      addresses.value = Array.isArray(response.data) ? response.data : []
    } else {
      showError('加载地址列表失败')
    }
  } catch (error: any) {
    console.error('Failed to load user addresses:', error)
    showError('加载地址列表失败: ' + (error?.message || '未知错误'))
  } finally {
    isLoading.value = false
  }
}

// 添加地址
const addAddress = async () => {
  if (!newAddressForm.address.trim() || !newAddressForm.label.trim()) {
    showError('请填写完整的地址信息')
    return
  }

  try {
    isLoading.value = true
    
    const response = await createUserAddress({
      token: authStore.loginToken || '',
      address: newAddressForm.address.trim(),
      label: newAddressForm.label.trim(),
      type: newAddressForm.type || 'wallet'
    })
    
    if (response && response.success === true) {
      showSuccess('地址创建成功！')
      // 重新加载地址列表
      await loadAddresses()
      // 清空表单
      newAddressForm.address = ''
      newAddressForm.label = ''
      newAddressForm.type = 'wallet'
      // 关闭创建表单
      // showCreateForm.value = false // This line was removed from the new_code, so it's removed here.
    } else {
      showError(response?.message || '创建失败')
    }
  } catch (error: any) {
    console.error('Failed to create address:', error)
    showError('创建失败: ' + (error?.message || '未知错误'))
  } finally {
    isLoading.value = false
  }
}

// 编辑地址
const editAddress = async (address: UserAddress) => {
  // 这里可以实现编辑功能
  const newLabel = prompt('请输入新的标签:', address.label)
  if (newLabel !== null && newLabel !== address.label) {
    try {
      const response = await updateUserAddress({
        token: authStore.loginToken || '',
        addressId: address.id,
        updateData: {
          label: newLabel
        }
      })
      
      if (response && response.success === true) {
        showSuccess('地址更新成功！')
        await loadAddresses()
        setTimeout(() => {
          success.value = ''
        }, 2000)
      }
    } catch (err: unknown) {
      error.value = '更新失败: ' + (err instanceof Error ? err.message : '未知错误')
    }
  }
}

// 删除地址
const removeAddress = async (address: UserAddress) => {
  if (!confirm(`确定要删除地址"${address.label || address.address}"吗？`)) {
    return
  }
  
  try {
    isLoading.value = true
    
    const response = await deleteUserAddress({
      token: authStore.loginToken || '',
      addressId: address.id
    })
    
    if (response && response.success === true) {
      showSuccess('地址已删除')
      await loadAddresses()
      
      setTimeout(() => {
        success.value = ''
      }, 2000)
    }
    
  } catch (err: unknown) {
    error.value = '删除失败: ' + (err instanceof Error ? err.message : '未知错误')
  } finally {
    isLoading.value = false
  }
}

// 复制地址
const copyAddress = async (address: string) => {
  try {
    await navigator.clipboard.writeText(address)
    showSuccess('地址已复制到剪贴板')
    setTimeout(() => {
      success.value = ''
    }, 2000)
  } catch {
    error.value = '复制失败'
  }
}

// 格式化地址显示
const formatAddress = (address: string) => {
  if (address.length <= 10) return address
  return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`
}

// 格式化余额
const formatBalance = (balance: number) => {
  return balance.toFixed(4)
}

// 获取类型样式
const getTypeClass = (type: string) => {
  switch (type) {
    case 'wallet':
      return 'bg-blue-100 text-blue-800'
    case 'contract':
      return 'bg-purple-100 text-purple-800'
    case 'exchange':
      return 'bg-green-100 text-green-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

// 获取类型文本
const getTypeText = (type: string) => {
  switch (type) {
    case 'wallet':
      return '钱包'
    case 'contract':
      return '合约'
    case 'exchange':
      return '交易所'
    default:
      return '其他'
  }
}
</script>
