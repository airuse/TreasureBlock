<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">SOL 铸币地址管理</h1>
      <div class="flex items-center space-x-4">
        <div class="text-sm text-gray-500">
          共 {{ totalAddresses.toLocaleString() }} 个铸币地址
        </div>
        <button 
          v-if="isAdmin"
          @click="showAddAddressModal = true"
          class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
          </svg>
          添加铸币地址
        </button>
      </div>
    </div>

    <!-- 搜索和筛选 -->
    <div class="card">
      <div class="flex flex-col sm:flex-row gap-4">
        <div class="flex-1">
          <label class="block text-sm font-medium text-gray-700 mb-2">搜索铸币地址</label>
          <div class="relative">
            <input 
              v-model="searchQuery" 
              type="text" 
              placeholder="输入铸币地址哈希或名称..."
              class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <svg class="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
              </svg>
            </div>
          </div>
        </div>
        <div class="sm:w-48">
          <label class="block text-sm font-medium text-gray-700 mb-2">代币类型</label>
          <select 
            v-model="typeFilter" 
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">全部类型</option>
            <option value="spl">SPL 代币</option>
            <option value="nft">NFT 代币</option>
            <option value="meme">Meme 代币</option>
            <option value="utility">实用代币</option>
            <option value="governance">治理代币</option>
            <option value="other">其他代币</option>
          </select>
        </div>
        <div class="sm:w-48">
          <label class="block text-sm font-medium text-gray-700 mb-2">状态</label>
          <select 
            v-model="statusFilter" 
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">全部状态</option>
            <option value="active">活跃</option>
            <option value="inactive">非活跃</option>
            <option value="paused">暂停</option>
          </select>
        </div>
        <div class="sm:w-48">
          <label class="block text-sm font-medium text-gray-700 mb-2">每页显示</label>
          <select 
            v-model="pageSize" 
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="10">10</option>
            <option value="25">25</option>
            <option value="50">50</option>
            <option value="100">100</option>
          </select>
        </div>
      </div>
    </div>

    <!-- 地址列表 -->
    <div class="card">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">铸币地址</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">代币名称</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">符号</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">类型</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">精度</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="address in addresses" :key="address.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center space-x-2">
                  <span class="text-blue-600 hover:text-blue-700 font-mono text-sm">
                    {{ address.address }}
                  </span>
                  <button 
                    @click="copyToClipboard(address.address)"
                    class="text-gray-400 hover:text-gray-600"
                    title="复制地址"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 002 2v8a2 2 0 002 2z"></path>
                    </svg>
                  </button>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center space-x-2">
                  <img 
                    v-if="address.contract_logo" 
                    :src="address.contract_logo" 
                    :alt="address.name"
                    class="w-6 h-6 rounded-full"
                  />
                  <span class="font-medium text-gray-900">{{ address.name || '未命名代币' }}</span>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="text-sm text-gray-900">{{ address.symbol || '-' }}</span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getTypeClass(address.contract_type)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                  {{ getTypeText(address.contract_type) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="text-sm text-gray-900">{{ address.decimals || 0 }}</span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getStatusClass(address.status === 1 ? 'active' : 'inactive')" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                  {{ getStatusText(address.status === 1 ? 'active' : 'inactive') }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                <div class="flex items-center space-x-2">
                  <button 
                    @click="viewDetails(address)"
                    class="text-blue-600 hover:text-blue-800 text-xs"
                  >
                    查看
                  </button>
                  <button 
                    v-if="isAdmin"
                    @click="editAddress(address)"
                    class="text-green-600 hover:text-green-800 text-xs"
                  >
                    编辑
                  </button>
                  <button 
                    v-if="isAdmin"
                    @click="deleteAddress(address)"
                    class="text-red-600 hover:text-red-800 text-xs"
                  >
                    删除
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="addresses.length === 0">
              <td colspan="7" class="px-6 py-12 text-center text-gray-500">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
                  </svg>
                  <p class="text-lg font-medium text-gray-900 mb-1">暂无铸币地址数据</p>
                  <p class="text-sm text-gray-500">请添加新的铸币地址或调整搜索条件</p>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 分页 -->
    <div class="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
      <div class="flex-1 flex justify-between sm:hidden">
        <button 
          @click="fetchAddresses(currentPage - 1)" 
          :disabled="currentPage <= 1"
          class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          上一页
        </button>
        <button 
          @click="fetchAddresses(currentPage + 1)" 
          :disabled="currentPage >= totalPages"
          class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          下一页
        </button>
      </div>
      <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
        <div>
          <p class="text-sm text-gray-700">
            显示第 <span class="font-medium">{{ (currentPage - 1) * pageSize + 1 }}</span> 到 
            <span class="font-medium">{{ Math.min(currentPage * pageSize, totalAddresses) }}</span> 条，
            共 <span class="font-medium">{{ totalAddresses.toLocaleString() }}</span> 条记录
          </p>
        </div>
        <div>
          <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
            <button 
              @click="fetchAddresses(currentPage - 1)" 
              :disabled="currentPage <= 1"
              class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <svg class="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" />
              </svg>
            </button>
            
            <button 
              v-for="pageNum in visiblePages" 
              :key="pageNum"
              @click="fetchAddresses(pageNum)"
              :class="[
                pageNum === currentPage 
                  ? 'z-10 bg-blue-50 border-blue-500 text-blue-600' 
                  : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50',
                'relative inline-flex items-center px-4 py-2 border text-sm font-medium'
              ]"
            >
              {{ pageNum }}
            </button>
            
            <button 
              @click="fetchAddresses(currentPage + 1)" 
              :disabled="currentPage >= totalPages"
              class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <svg class="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
              </svg>
            </button>
          </nav>
        </div>
      </div>
    </div>

    <!-- 添加/编辑地址模态框 -->
    <Teleport to="body">
      <div v-if="showAddAddressModal || showEditModal" class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
        <div class="relative top-16 mx-auto p-5 border w-11/12 max-w-4xl shadow-xl rounded-xl bg-white">
          <div class="flex justify-between items-center mb-4 pb-3 border-b">
            <h3 class="text-lg font-semibold text-gray-900">{{ editingAddress ? '编辑铸币地址' : '添加铸币地址' }}</h3>
            <button 
              @click="closeModal"
              class="text-gray-400 hover:text-gray-600 transition-colors"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
              </svg>
            </button>
          </div>
          
          <form @submit.prevent="submitAddress" class="max-h-[65vh] overflow-y-auto pr-2">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">关联程序</label>
                <select 
                  v-model="addressForm.program_id"
                  class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">选择程序（可选）</option>
                  <option v-for="p in programOptions" :key="p.program_id" :value="p.program_id">
                    {{ p.name || p.program_id }}
                  </option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">铸币地址 *</label>
                <input 
                  v-model="addressForm.address" 
                  type="text" 
                  required
                  :disabled="!!editingAddress"
                  placeholder="输入铸币地址"
                  class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">代币名称 *</label>
                <input 
                  v-model="addressForm.name" 
                  type="text" 
                  required
                  placeholder="输入代币名称"
                  class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">代币符号 *</label>
                <input 
                  v-model="addressForm.symbol" 
                  type="text" 
                  required
                  placeholder="输入代币符号"
                  class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">精度</label>
                <input 
                  v-model.number="addressForm.decimals" 
                  type="number" 
                  min="0"
                  max="18"
                  placeholder="输入精度"
                  class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">代币类型</label>
                <select 
                  v-model="addressForm.type" 
                  class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">选择类型</option>
                  <option value="spl">SPL 代币</option>
                  <option value="nft">NFT 代币</option>
                  <option value="meme">Meme 代币</option>
                  <option value="utility">实用代币</option>
                  <option value="governance">治理代币</option>
                  <option value="other">其他代币</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">状态</label>
                <select 
                  v-model="addressForm.status" 
                  class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="active">活跃</option>
                  <option value="inactive">非活跃</option>
                  <option value="paused">暂停</option>
                </select>
              </div>
            </div>

            <div class="mb-4">
              <label class="block text-sm font-medium text-gray-700 mb-2">代币描述</label>
              <textarea 
                v-model="addressForm.description" 
                rows="3"
                placeholder="输入代币描述"
                class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              ></textarea>
            </div>

            <div class="mb-4">
              <label class="block text-sm font-medium text-gray-700 mb-2">Logo URL</label>
              <input 
                v-model="addressForm.logo" 
                type="url" 
                placeholder="输入Logo URL"
                class="block w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </form>

          <!-- 底部按钮 -->
          <div class="flex justify-end space-x-3 pt-4 border-t">
            <button 
              type="button"
              @click="closeModal"
              class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 border border-gray-300 rounded-md hover:bg-gray-200 transition-colors"
            >
              取消
            </button>
            <button 
              @click="submitAddress"
              class="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700 transition-colors"
            >
              {{ editingAddress ? '更新' : '添加' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { contracts, solPrograms } from '@/api'
import { useAuthStore } from '@/stores/auth'
import type { Contract } from '@/types'

// 认证状态
const authStore = useAuthStore()
const isAdmin = computed(() => authStore.isAuthenticated)

// 数据状态
const addresses = ref<Contract[]>([])
const totalAddresses = ref(0)
const currentPage = ref(1)
const pageSize = ref(25)
const searchQuery = ref('')
const typeFilter = ref('')
const statusFilter = ref('')

// 模态框状态
const showAddAddressModal = ref(false)
const showEditModal = ref(false)
const editingAddress = ref<Contract | null>(null)

// 程序下拉选项
const programOptions = ref<{ program_id: string; name?: string }[]>([])

// 表单数据
const addressForm = ref({
  address: '',
  name: '',
  symbol: '',
  decimals: 9,
  type: '',
  status: 'active',
  description: '',
  logo: '',
  program_id: ''
})

// 计算属性
const totalPages = computed(() => Math.max(1, Math.ceil(totalAddresses.value / pageSize.value)))

const visiblePages = computed(() => {
  const pages = []
  const start = Math.max(1, currentPage.value - 2)
  const end = Math.min(totalPages.value, currentPage.value + 2)
  
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  return pages
})

// 方法
async function fetchAddresses(page = 1) {
  try {
    currentPage.value = page
    const params = {
      page: page,
      page_size: pageSize.value,
      chainName: 'sol',
      search: searchQuery.value,
      contractType: typeFilter.value,
      status: statusFilter.value
    }
    
    const response = await contracts.getContracts(params)
    if (response.success) {
      addresses.value = response.data || []
      totalAddresses.value = (response as any).meta?.total || 0
    }
  } catch (error) {
    console.error('获取地址列表失败:', error)
  }
}

function openAddModal() {
  editingAddress.value = null
  addressForm.value = {
    address: '',
    name: '',
    symbol: '',
    decimals: 9,
    type: '',
    status: 'active',
    description: '',
    logo: '',
    program_id: ''
  }
  showAddAddressModal.value = true
}

function editAddress(address: Contract) {
  editingAddress.value = address
  addressForm.value = {
    address: address.address || '',
    name: address.name || '',
    symbol: address.symbol || '',
    decimals: address.decimals || 9,
    type: address.contract_type || '',
    status: address.status === 1 ? 'active' : 'inactive',
    description: (() => {
      try {
        if (address.metadata && typeof address.metadata === 'string') {
          const parsed = JSON.parse(address.metadata)
          return parsed.description || ''
        }
        return ''
      } catch {
        return ''
      }
    })(),
    logo: address.contract_logo || '',
    program_id: address.program_id || ''
  }
  showEditModal.value = true
}

function closeModal() {
  showAddAddressModal.value = false
  showEditModal.value = false
  editingAddress.value = null
}

async function submitAddress() {
  try {
    const contractData = {
      chain_name: 'sol',
      address: addressForm.value.address,  // 修复：使用 address 而不是 contract_address
      name: addressForm.value.name,        // 修复：使用 name 而不是 contract_name
      symbol: addressForm.value.symbol,    // 修复：使用 symbol 而不是 contract_symbol
      contract_type: addressForm.value.type || 'spl',
      decimals: addressForm.value.decimals,
      metadata: addressForm.value.description ? JSON.stringify({ description: addressForm.value.description }) : '',  // 修复：将 description 保存到 metadata
      contract_logo: addressForm.value.logo,  // 修复：使用 contract_logo 而不是 logo_url
      status: addressForm.value.status === 'active' ? 1 : 0,
      program_id: addressForm.value.program_id || ''
    }
    
    let response
    if (editingAddress.value) {
      response = await contracts.createOrUpdateContract(contractData)
    } else {
      response = await contracts.createOrUpdateContract(contractData)
    }
    
    if (response.success) {
      closeModal()
      await fetchAddresses(currentPage.value)
    }
  } catch (error) {
    console.error('保存地址失败:', error)
  }
}

async function deleteAddress(address: Contract) {
  if (confirm(`确定要删除铸币地址 "${address.name}" 吗？`)) {
    try {
      if (address.address) {
        const response = await contracts.deleteContract(address.address)
        if (response.success) {
          await fetchAddresses(currentPage.value)
        }
      }
    } catch (error) {
      console.error('删除地址失败:', error)
    }
  }
}

function viewDetails(address: Contract) {
  // 跳转到地址详情页面，传递来源页面信息
  router.push(`/sol/addresses/${address.address}?from=/sol/addresses`)
}

// 复制到剪贴板
async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    // 这里可以添加成功提示
  } catch (err) {
    console.error('复制失败:', err)
  }
}

// 获取类型样式类
function getTypeClass(type?: string) {
  const typeClasses: Record<string, string> = {
    spl: 'bg-blue-100 text-blue-800',
    nft: 'bg-purple-100 text-purple-800',
    meme: 'bg-pink-100 text-pink-800',
    utility: 'bg-green-100 text-green-800',
    governance: 'bg-yellow-100 text-yellow-800',
    other: 'bg-gray-100 text-gray-800'
  }
  return typeClasses[type || ''] || 'bg-gray-100 text-gray-800'
}

// 获取类型文本
function getTypeText(type?: string) {
  const typeTexts: Record<string, string> = {
    spl: 'SPL 代币',
    nft: 'NFT 代币',
    meme: 'Meme 代币',
    utility: '实用代币',
    governance: '治理代币',
    other: '其他代币'
  }
  return typeTexts[type || ''] || type || '未知'
}

// 获取状态样式类
function getStatusClass(status?: string) {
  const statusClasses: Record<string, string> = {
    active: 'bg-green-100 text-green-800',
    inactive: 'bg-red-100 text-red-800',
    paused: 'bg-yellow-100 text-yellow-800'
  }
  return statusClasses[status || ''] || 'bg-gray-100 text-gray-800'
}

// 获取状态文本
function getStatusText(status?: string) {
  const statusTexts: Record<string, string> = {
    active: '活跃',
    inactive: '非活跃',
    paused: '暂停'
  }
  return statusTexts[status || ''] || status || '未知'
}

// 监听筛选器变化
watch([searchQuery, typeFilter, statusFilter, pageSize], () => {
  fetchAddresses(1)
}, { deep: true })

// 路由
const router = useRouter()

// 初始化
onMounted(() => {
  fetchAddresses(1)
  // 预加载程序选项（只取前200条，避免过大）
  solPrograms.listPrograms({ page: 1, page_size: 200 }).then((res: any) => {
    if (res?.success && Array.isArray(res.data)) {
      programOptions.value = res.data.map((x: any) => ({ program_id: x.program_id, name: x.name }))
    }
  }).catch(() => {})
})
</script>

<style scoped>
.card {
  background-color: white;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
  border-radius: 0.5rem;
  border: 1px solid #e5e7eb;
}
</style>
