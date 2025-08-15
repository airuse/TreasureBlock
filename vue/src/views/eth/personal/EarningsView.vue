<template>
  <div class="space-y-6">
    <!-- 页面头部 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <div class="flex items-center justify-between">
          <div>
            <h1 class="text-2xl font-bold text-gray-900">扫块收益</h1>
            <p class="mt-1 text-sm text-gray-500">查看和管理您的扫块收益</p>
          </div>
          <div class="flex items-center space-x-2">
            <div class="w-3 h-3 bg-blue-500 rounded-full"></div>
            <span class="text-sm text-gray-600">ETH 网络</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 收益概览 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">收益概览</h3>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div class="text-center">
            <div class="text-3xl font-bold text-green-600">{{ totalEarnings }}</div>
            <div class="text-sm text-gray-500">总收益 (TB)</div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-blue-600">{{ todayEarnings }}</div>
            <div class="text-sm text-gray-500">今日收益 (TB)</div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-purple-600">{{ monthlyEarnings }}</div>
            <div class="text-sm text-gray-500">本月收益 (TB)</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 收益图表 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">收益趋势</h3>
        <div class="h-64 bg-gray-50 rounded-lg flex items-center justify-center">
          <div class="text-center text-gray-500">
            <ChartBarIcon class="mx-auto h-12 w-12 text-gray-400 mb-2" />
            <p>收益图表将在这里显示</p>
            <p class="text-sm">支持按日、周、月查看收益趋势</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 收益记录 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg leading-6 font-medium text-gray-900">收益记录</h3>
          <div class="flex space-x-2">
            <select v-model="selectedPeriod" class="border border-gray-300 rounded-md px-3 py-2 text-sm">
              <option value="7">最近7天</option>
              <option value="30">最近30天</option>
              <option value="90">最近90天</option>
            </select>
          </div>
        </div>
        
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">时间</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">区块哈希</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">收益 (TB)</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="earning in earningsList" :key="earning.id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {{ formatTime(earning.timestamp) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono">
                    {{ earning.blockHash.substring(0, 10) }}...{{ earning.blockHash.substring(earning.blockHash.length - 8) }}
                  </code>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-green-600">
                  +{{ earning.amount }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="getStatusClass(earning.status)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                    {{ getStatusText(earning.status) }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                  <button
                    @click="viewBlockDetails(earning.blockHash)"
                    class="text-blue-600 hover:text-blue-900"
                  >
                    查看区块
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ChartBarIcon } from '@heroicons/vue/24/outline'

// 响应式数据
const totalEarnings = ref(12.5)
const todayEarnings = ref(0.8)
const monthlyEarnings = ref(3.2)
const selectedPeriod = ref(30)
const currentPage = ref(1)
const pageSize = ref(10)
const totalItems = ref(0)
const totalPages = ref(0)

// 收益记录列表
const earningsList = ref([
  {
    id: 1,
    timestamp: new Date(Date.now() - 1000 * 60 * 30),
    blockHash: '0x1234567890abcdef1234567890abcdef12345678',
    amount: 0.5,
    status: 'confirmed'
  },
  {
    id: 2,
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 2),
    blockHash: '0xabcdef1234567890abcdef1234567890abcdef12',
    amount: 0.3,
    status: 'pending'
  },
  {
    id: 3,
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24),
    blockHash: '0x567890abcdef1234567890abcdef1234567890ab',
    amount: 0.7,
    status: 'confirmed'
  }
])

// 获取状态样式
const getStatusClass = (status: string) => {
  switch (status) {
    case 'confirmed': return 'bg-green-100 text-green-800'
    case 'pending': return 'bg-yellow-100 text-yellow-800'
    case 'failed': return 'bg-red-100 text-red-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

// 获取状态文本
const getStatusText = (status: string) => {
  switch (status) {
    case 'confirmed': return '已确认'
    case 'pending': return '待确认'
    case 'failed': return '失败'
    default: return '未知'
  }
}

// 格式化时间
const formatTime = (timestamp: Date) => {
  return timestamp.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 查看区块详情
const viewBlockDetails = (blockHash: string) => {
  // TODO: 跳转到区块详情页面
  console.log('查看区块:', blockHash)
}

// 分页方法
const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    loadEarnings()
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    loadEarnings()
  }
}

// 加载收益数据
const loadEarnings = async () => {
  // TODO: 从API加载真实数据
  totalItems.value = 25
  totalPages.value = Math.ceil(totalItems.value / pageSize.value)
}

// 监听周期变化
watch(selectedPeriod, () => {
  currentPage.value = 1
  loadEarnings()
})

onMounted(() => {
  loadEarnings()
})
</script>
