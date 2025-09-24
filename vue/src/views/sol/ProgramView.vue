<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">Sol 程序维护</h1>
      <div class="flex items-center space-x-4">
        <div class="text-sm text-gray-500">
          共 {{ total.toLocaleString() }} 个程序
        </div>
        <button 
          @click="openCreate"
          class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
          </svg>
          新增程序
        </button>
      </div>
    </div>

    <!-- 搜索和筛选 -->
    <div class="card">
      <div class="flex flex-col sm:flex-row gap-4">
        <div class="flex-1">
          <label class="block text-sm font-medium text-gray-700 mb-2">搜索程序</label>
          <div class="relative">
            <input 
              v-model="keyword" 
              type="text" 
              placeholder="输入 Program ID、名称或别名..."
              class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              @keyup.enter="fetchList(1)"
            />
            <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <svg class="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
              </svg>
            </div>
          </div>
        </div>
        <div class="sm:w-48">
          <label class="block text-sm font-medium text-gray-700 mb-2">程序类型</label>
          <select 
            v-model="typeFilter" 
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">全部类型</option>
            <option value="system">系统程序</option>
            <option value="token">代币程序</option>
            <option value="dex">DEX 程序</option>
            <option value="defi">DeFi 程序</option>
            <option value="nft">NFT 程序</option>
            <option value="other">其他程序</option>
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
          </select>
        </div>
        <div class="sm:w-48">
          <label class="block text-sm font-medium text-gray-700 mb-2">每页显示</label>
          <select 
            v-model="pageSize" 
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            @change="fetchList(1)"
          >
            <option value="10">10</option>
            <option value="25">25</option>
            <option value="50">50</option>
            <option value="100">100</option>
          </select>
        </div>
      </div>
    </div>

    <!-- 程序列表 -->
    <div class="card">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Program ID</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">程序名称</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">分类</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">类型</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
        </tr>
      </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="item in list" :key="item.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center space-x-2">
                  <span class="text-blue-600 hover:text-blue-700 font-mono text-sm">
                    {{ item.program_id }}
                  </span>
                  <button 
                    @click="copyToClipboard(item.program_id)"
                    class="text-gray-400 hover:text-gray-600"
                    title="复制 Program ID"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 002 2v8a2 2 0 002 2z"></path>
                    </svg>
                  </button>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center space-x-2">
                  <span class="font-medium text-gray-900">{{ item.name || '未命名程序' }}</span>
                  <span v-if="item.alias" class="text-sm text-gray-500">({{ item.alias }})</span>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="text-sm text-gray-900">{{ item.category || '-' }}</span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getTypeClass(item.type)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                  {{ getTypeText(item.type) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getStatusClass(item.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                  {{ getStatusText(item.status) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                <div class="flex items-center space-x-2">
                  <button 
                    @click="openEdit(item)"
                    class="text-blue-600 hover:text-blue-800 text-xs"
                  >
                    编辑
                  </button>
                  <button 
                    @click="remove(item)"
                    class="text-red-600 hover:text-red-800 text-xs"
                  >
                    删除
                  </button>
                </div>
          </td>
        </tr>
        <tr v-if="list.length === 0">
              <td colspan="6" class="px-6 py-12 text-center text-gray-500">
                <div class="flex flex-col items-center">
                  <svg class="w-12 h-12 text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
                  </svg>
                  <p class="text-lg font-medium text-gray-900 mb-1">暂无程序数据</p>
                  <p class="text-sm text-gray-500">请添加新的程序或调整搜索条件</p>
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
          @click="fetchList(page-1)" 
          :disabled="page <= 1"
          class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          上一页
        </button>
        <button 
          @click="fetchList(page+1)" 
          :disabled="page >= totalPages"
          class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          下一页
        </button>
        </div>
      <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
        <div>
          <p class="text-sm text-gray-700">
            显示第 <span class="font-medium">{{ (page - 1) * pageSize + 1 }}</span> 到 
            <span class="font-medium">{{ Math.min(page * pageSize, total) }}</span> 条，
            共 <span class="font-medium">{{ total.toLocaleString() }}</span> 条记录
          </p>
        </div>
        <div>
          <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
            <button 
              @click="fetchList(page-1)" 
              :disabled="page <= 1"
              class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <svg class="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" />
              </svg>
            </button>
            
            <button 
              v-for="pageNum in visiblePages" 
              :key="pageNum"
              @click="fetchList(pageNum)"
              :class="[
                pageNum === page 
                  ? 'z-10 bg-blue-50 border-blue-500 text-blue-600' 
                  : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50',
                'relative inline-flex items-center px-4 py-2 border text-sm font-medium'
              ]"
            >
              {{ pageNum }}
            </button>
            
            <button 
              @click="fetchList(page+1)" 
              :disabled="page >= totalPages"
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

    <!-- 编辑程序模态框 -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
        <div class="relative top-16 mx-auto p-5 border w-11/12 max-w-5xl shadow-xl rounded-xl bg-white">
          <div class="flex justify-between items-center mb-4 pb-3 border-b">
            <h3 class="text-lg font-semibold text-gray-900">{{ editing?.id ? '编辑程序' : '新增程序' }}</h3>
            <button 
              @click="closeModal"
              class="text-gray-400 hover:text-gray-600 transition-colors"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
              </svg>
            </button>
          </div>
          
          <form @submit.prevent="submit" class="max-h-[65vh] overflow-y-auto pr-2">
            <!-- 基本信息 -->
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3 mb-4">
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-1">Program ID *</label>
                <input 
                  v-model="form.program_id" 
                  type="text" 
                  :disabled="!!editing?.id"
                  placeholder="输入 Program ID"
                  class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-1">程序名称 *</label>
                <input 
                  v-model="form.name" 
                  type="text" 
                  required
                  placeholder="输入程序名称"
                  class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-1">别名</label>
                <input 
                  v-model="form.alias" 
                  type="text" 
                  placeholder="输入别名"
                  class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-1">分类</label>
                <input 
                  v-model="form.category" 
                  type="text" 
                  placeholder="输入分类"
                  class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-1">类型</label>
                <select 
                  v-model="form.type" 
                  class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                >
                  <option value="">选择类型</option>
                  <option value="system">系统程序</option>
                  <option value="token">代币程序</option>
                  <option value="dex">DEX 程序</option>
                  <option value="defi">DeFi 程序</option>
                  <option value="nft">NFT 程序</option>
                  <option value="other">其他程序</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-1">状态</label>
                <select 
                  v-model="form.status" 
                  class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                >
                  <option value="active">活跃</option>
                  <option value="inactive">非活跃</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-1">版本</label>
                <input 
                  v-model="form.version" 
                  type="text" 
                  placeholder="输入版本号"
                  class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div class="flex items-center">
                <input 
                  type="checkbox" 
                  v-model="form.is_system" 
                  class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                />
                <label class="ml-2 block text-xs font-medium text-gray-700">系统程序</label>
              </div>
            </div>

            <!-- 描述 -->
            <div class="mb-4">
              <label class="block text-xs font-medium text-gray-700 mb-1">描述</label>
              <textarea 
                v-model="form.description" 
                rows="3"
                placeholder="输入程序描述"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              ></textarea>
            </div>

            <!-- 解析规则 -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-1">指令解析规则 (JSON)</label>
                <textarea 
                  v-model="instructionRulesRaw" 
                  rows="6"
                  placeholder="输入指令解析规则 JSON"
                  class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded font-mono focus:outline-none focus:ring-1 focus:ring-blue-500"
                ></textarea>
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-1">事件解析规则 (JSON)</label>
                <textarea 
                  v-model="eventRulesRaw" 
                  rows="6"
                  placeholder="输入事件解析规则 JSON"
                  class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded font-mono focus:outline-none focus:ring-1 focus:ring-blue-500"
                ></textarea>
              </div>
            </div>

            <!-- 样例数据 -->
            <div class="mb-4">
              <label class="block text-xs font-medium text-gray-700 mb-1">样例数据 (JSON)</label>
              <textarea 
                v-model="sampleDataRaw" 
                rows="6"
                placeholder="输入样例数据 JSON"
                class="block w-full px-2 py-1.5 text-xs border border-gray-300 rounded font-mono focus:outline-none focus:ring-1 focus:ring-blue-500"
              ></textarea>
            </div>
          </form>

          <!-- 底部按钮 -->
          <div class="flex justify-end space-x-3 pt-4 border-t">
            <button 
              type="button"
              @click="closeModal"
              class="px-4 py-2 text-xs font-medium text-gray-700 bg-gray-100 border border-gray-300 rounded hover:bg-gray-200 transition-colors"
            >
              取消
            </button>
            <button 
              @click="submit"
              class="px-4 py-2 text-xs font-medium text-white bg-blue-600 border border-transparent rounded hover:bg-blue-700 transition-colors"
            >
              保存
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { solPrograms } from '@/api'
import type { SolProgram } from '@/types'

const list = ref<SolProgram[]>([])
const page = ref(1)
const pageSize = ref(25)
const total = ref(0)
const keyword = ref('')
const typeFilter = ref('')
const statusFilter = ref('')

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

// 计算可见页码
const visiblePages = computed(() => {
  const pages = []
  const start = Math.max(1, page.value - 2)
  const end = Math.min(totalPages.value, page.value + 2)
  
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  return pages
})

async function fetchList(p = page.value) {
  page.value = p
  const res: any = await solPrograms.listPrograms({ page: page.value, page_size: pageSize.value, keyword: keyword.value })
  if (res?.success) {
    list.value = res.data
    total.value = res.meta?.total || 0
  }
}

const showModal = ref(false)
const editing = ref<SolProgram | null>(null)
const form = ref<SolProgram>({ program_id: '', name: '', status: 'active', is_system: false })
const instructionRulesRaw = ref('')
const eventRulesRaw = ref('')
const sampleDataRaw = ref('')

function openCreate() {
  editing.value = null
  form.value = { program_id: '', name: '', status: 'active', is_system: false }
  instructionRulesRaw.value = ''
  eventRulesRaw.value = ''
  sampleDataRaw.value = ''
  showModal.value = true
}

function openEdit(item: SolProgram) {
  editing.value = item
  form.value = { ...item }
  instructionRulesRaw.value = item.instruction_rules ? JSON.stringify(item.instruction_rules, null, 2) : ''
  eventRulesRaw.value = item.event_rules ? JSON.stringify(item.event_rules, null, 2) : ''
  sampleDataRaw.value = item.sample_data ? JSON.stringify(item.sample_data, null, 2) : ''
  showModal.value = true
}

function closeModal() { showModal.value = false }

async function submit() {
  try {
    const payload: SolProgram = {
      ...form.value,
      instruction_rules: parseJSONSafe(instructionRulesRaw.value),
      event_rules: parseJSONSafe(eventRulesRaw.value),
      sample_data: parseJSONSafe(sampleDataRaw.value)
    }
    let res: any
    if (editing.value?.id) {
      res = await solPrograms.updateProgram(editing.value.id, payload)
    } else {
      res = await solPrograms.createProgram(payload)
    }
    if (res?.success) {
      closeModal()
      fetchList(1)
    }
  } catch (e) {
    console.error(e)
  }
}

function parseJSONSafe(s: string) {
  if (!s) return undefined
  try { return JSON.parse(s) } catch { return undefined }
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
    system: 'bg-gray-100 text-gray-800',
    token: 'bg-green-100 text-green-800',
    dex: 'bg-blue-100 text-blue-800',
    defi: 'bg-purple-100 text-purple-800',
    nft: 'bg-pink-100 text-pink-800',
    other: 'bg-gray-100 text-gray-800'
  }
  return typeClasses[type || ''] || 'bg-gray-100 text-gray-800'
}

// 获取类型文本
function getTypeText(type?: string) {
  const typeTexts: Record<string, string> = {
    system: '系统程序',
    token: '代币程序',
    dex: 'DEX 程序',
    defi: 'DeFi 程序',
    nft: 'NFT 程序',
    other: '其他程序'
  }
  return typeTexts[type || ''] || type || '未知'
}

// 获取状态样式类
function getStatusClass(status?: string) {
  const statusClasses: Record<string, string> = {
    active: 'bg-green-100 text-green-800',
    inactive: 'bg-red-100 text-red-800'
  }
  return statusClasses[status || ''] || 'bg-gray-100 text-gray-800'
}

// 获取状态文本
function getStatusText(status?: string) {
  const statusTexts: Record<string, string> = {
    active: '活跃',
    inactive: '非活跃'
  }
  return statusTexts[status || ''] || status || '未知'
}

// 删除程序
async function remove(item: SolProgram) {
  if (confirm(`确定要删除程序 "${item.name}" 吗？`)) {
    try {
      if (item.id) {
        const res: any = await solPrograms.deleteProgram(item.id)
        if (res?.success) {
          fetchList(page.value)
        }
      }
    } catch (e) {
      console.error(e)
    }
  }
}

// 监听筛选器变化
watch([typeFilter, statusFilter], () => {
  fetchList(1)
})

fetchList(1)
</script>

<style scoped>
.card {
  background-color: white;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
  border-radius: 0.5rem;
  border: 1px solid #e5e7eb;
}
</style>


