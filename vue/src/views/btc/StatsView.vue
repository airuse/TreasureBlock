<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">比特币网络统计</h1>
      <div class="flex items-center space-x-4">
        <select 
          v-model="timeRange" 
          class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="7d">最近7天</option>
          <option value="30d">最近30天</option>
          <option value="90d">最近90天</option>
          <option value="1y">最近1年</option>
        </select>
      </div>
    </div>

    <!-- 关键指标卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <div class="card">
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div class="w-8 h-8 bg-orange-500 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                <path d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3z"/>
              </svg>
            </div>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-500">全网算力</p>
            <p class="text-2xl font-bold text-gray-900">{{ formatHashrate(currentStats.hashrate) }}</p>
            <p class="text-sm text-green-600">+2.3% 7天</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div class="w-8 h-8 bg-blue-500 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                <path d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zM8 7a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zM14 4a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z"/>
              </svg>
            </div>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-500">网络难度</p>
            <p class="text-2xl font-bold text-gray-900">{{ formatDifficulty(currentStats.difficulty) }}</p>
            <p class="text-sm text-red-600">-1.2% 7天</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div class="w-8 h-8 bg-green-500 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-13a1 1 0 10-2 0v.092a4.535 4.535 0 00-1.676.662C6.602 6.234 6 7.009 6 8c0 .99.602 1.765 1.324 2.246.48.32 1.054.545 1.676.662v1.941c-.391-.127-.68-.317-.843-.504a1 1 0 10-1.51 1.31c.562.649 1.413 1.076 2.353 1.253V15a1 1 0 102 0v-.092a4.535 4.535 0 001.676-.662C13.398 13.766 14 12.991 14 12c0-.99-.602-1.765-1.324-2.246A4.535 4.535 0 0011 9.092V7.151c.391.127.68.317.843.504a1 1 0 101.511-1.31c-.563-.649-1.413-1.076-2.354-1.253V5z" clip-rule="evenodd"/>
              </svg>
            </div>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-500">平均手续费</p>
            <p class="text-2xl font-bold text-gray-900">{{ formatAmount(currentStats.avgFee) }} BTC</p>
            <p class="text-sm text-green-600">-15.8% 7天</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div class="w-8 h-8 bg-purple-500 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                <path d="M11 17a1 1 0 001.447.894l4-2A1 1 0 0017 15V9.236a1 1 0 00-1.447-.894l-4 2a1 1 0 00-.553.894V17zM15.211 6.276a1 1 0 000-1.552l-4.764-3.368a1 1 0 00-1.447 0L4.789 4.724a1 1 0 000 1.552l4.764 3.368a1 1 0 001.447 0l4.764-3.368zM4.447 8.342A1 1 0 003 9.236V15a1 1 0 00.553.894l4 2A1 1 0 009 17v-5.764a1 1 0 00-.553-.894l-4-2z"/>
              </svg>
            </div>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-500">日交易量</p>
            <p class="text-2xl font-bold text-gray-900">{{ formatAmount(currentStats.dailyVolume) }} BTC</p>
            <p class="text-sm text-green-600">+8.7% 7天</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 图表区域 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 算力趋势 -->
      <div class="card">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-lg font-semibold text-gray-900">全网算力趋势</h2>
          <span class="text-sm text-gray-500">TH/s</span>
        </div>
        <div class="h-64 flex items-center justify-center bg-gray-50 rounded-lg">
          <div class="text-center">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
            </svg>
            <p class="mt-2 text-sm text-gray-500">算力趋势图表</p>
          </div>
        </div>
      </div>

      <!-- 难度调整历史 -->
      <div class="card">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-lg font-semibold text-gray-900">难度调整历史</h2>
          <span class="text-sm text-gray-500">T</span>
        </div>
        <div class="h-64 flex items-center justify-center bg-gray-50 rounded-lg">
          <div class="text-center">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"></path>
            </svg>
            <p class="mt-2 text-sm text-gray-500">难度变化图表</p>
          </div>
        </div>
      </div>

      <!-- 手续费趋势 -->
      <div class="card">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-lg font-semibold text-gray-900">平均手续费趋势</h2>
          <span class="text-sm text-gray-500">BTC</span>
        </div>
        <div class="h-64 flex items-center justify-center bg-gray-50 rounded-lg">
          <div class="text-center">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z"></path>
            </svg>
            <p class="mt-2 text-sm text-gray-500">手续费趋势图表</p>
          </div>
        </div>
      </div>

      <!-- 活跃地址统计 -->
      <div class="card">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-lg font-semibold text-gray-900">日活跃地址数</h2>
          <span class="text-sm text-gray-500">地址</span>
        </div>
        <div class="h-64 flex items-center justify-center bg-gray-50 rounded-lg">
          <div class="text-center">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>
            </svg>
            <p class="mt-2 text-sm text-gray-500">活跃地址图表</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 详细统计表格 -->
    <div class="card">
      <h2 class="text-lg font-semibold text-gray-900 mb-4">网络指标详情</h2>
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">指标</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">当前值</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">24小时变化</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">7天变化</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">历史最高</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="metric in detailedMetrics" :key="metric.name" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                {{ metric.name }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {{ metric.current }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm">
                <span :class="metric.change24h >= 0 ? 'text-green-600' : 'text-red-600'">
                  {{ metric.change24h >= 0 ? '+' : '' }}{{ metric.change24h }}%
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm">
                <span :class="metric.change7d >= 0 ? 'text-green-600' : 'text-red-600'">
                  {{ metric.change7d >= 0 ? '+' : '' }}{{ metric.change7d }}%
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {{ metric.ath }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { formatAmount, formatHashrate, formatDifficulty } from '@/utils/formatters'

// 响应式数据
const timeRange = ref('30d')

const currentStats = ref({
  hashrate: 500e12, // 500 TH/s
  difficulty: 50e12, // 50T
  avgFee: 0.00001, // 0.00001 BTC
  dailyVolume: 50000, // 50000 BTC
  activeAddresses: 500000,
  memPoolSize: 1500,
  avgBlockTime: 10.2
})

const detailedMetrics = ref([
  {
    name: '全网算力',
    current: '500 TH/s',
    change24h: 2.3,
    change7d: 5.7,
    ath: '680 TH/s'
  },
  {
    name: '网络难度',
    current: '50.0 T',
    change24h: -1.2,
    change7d: -2.8,
    ath: '62.5 T'
  },
  {
    name: '平均手续费',
    current: '0.00001 BTC',
    change24h: -15.8,
    change7d: -22.3,
    ath: '0.0015 BTC'
  },
  {
    name: '日交易量',
    current: '50,000 BTC',
    change24h: 8.7,
    change7d: 12.4,
    ath: '180,000 BTC'
  },
  {
    name: '活跃地址数',
    current: '500,000',
    change24h: 3.2,
    change7d: 7.8,
    ath: '1,200,000'
  },
  {
    name: '内存池大小',
    current: '1,500 MB',
    change24h: -12.4,
    change7d: -8.9,
    ath: '120,000 MB'
  },
  {
    name: '平均出块时间',
    current: '10.2 分钟',
    change24h: 1.8,
    change7d: -2.1,
    ath: '20.5 分钟'
  }
])

// 监听时间范围变化
watch(timeRange, (newRange) => {
  console.log('时间范围变更:', newRange)
  // 这里应该重新加载对应时间范围的数据
})

// 加载数据
const loadData = () => {
  // 模拟数据加载
  console.log('加载统计数据...')
}

onMounted(() => {
  loadData()
})
</script> 