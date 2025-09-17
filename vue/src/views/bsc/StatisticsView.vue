<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">区块链统计</h1>
      <div class="text-sm text-gray-500">
        最后更新: {{ lastUpdateTime }}
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <div class="card">
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div class="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"></path>
              </svg>
            </div>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-500">24小时交易量</p>
            <p class="text-2xl font-semibold text-gray-900">{{ formatAmount(stats.dailyVolume) }} BNB</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div class="w-8 h-8 bg-green-100 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1"></path>
              </svg>
            </div>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-500">平均Gas价格</p>
            <p class="text-2xl font-semibold text-gray-900">{{ formatGasPrice(stats.avgGasPrice) }} Gwei</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div class="w-8 h-8 bg-purple-100 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
              </svg>
            </div>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-500">平均区块时间</p>
            <p class="text-2xl font-semibold text-gray-900">{{ stats.avgBlockTime }}s</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div class="w-8 h-8 bg-yellow-100 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
              </svg>
            </div>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-500">网络难度</p>
            <p class="text-2xl font-semibold text-gray-900">{{ formatDifficulty(stats.difficulty) }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 图表区域 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 交易量趋势 -->
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">24小时交易量趋势</h3>
        <div class="h-64 bg-gray-50 rounded-lg flex items-center justify-center">
          <div class="text-center">
            <svg class="w-12 h-12 text-gray-400 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
            </svg>
            <p class="text-gray-500">图表区域</p>
            <p class="text-sm text-gray-400">这里将显示交易量趋势图表</p>
          </div>
        </div>
      </div>

      <!-- Gas价格趋势 -->
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">Gas价格趋势</h3>
        <div class="h-64 bg-gray-50 rounded-lg flex items-center justify-center">
          <div class="text-center">
            <svg class="w-12 h-12 text-gray-400 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
            </svg>
            <p class="text-gray-500">图表区域</p>
            <p class="text-sm text-gray-400">这里将显示Gas价格趋势图表</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 详细统计表格 -->
    <div class="card">
      <h3 class="text-lg font-semibold text-gray-900 mb-4">网络统计详情</h3>
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">指标</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">当前值</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">24小时变化</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">7天变化</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="metric in networkMetrics" :key="metric.name">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                {{ metric.name }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {{ metric.currentValue }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm">
                <span :class="getChangeClass(metric.change24h)">
                  {{ formatChange(metric.change24h) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm">
                <span :class="getChangeClass(metric.change7d)">
                  {{ formatChange(metric.change7d) }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

// 响应式数据
const lastUpdateTime = ref('')

import type { NetworkMetricDisplay } from '@/types'

const networkMetrics = ref<NetworkMetricDisplay[]>([])

// 统计数据
const stats = ref({
  dailyVolume: 0,
  avgGasPrice: 0,
  avgBlockTime: 0,
  difficulty: 0
})

// 格式化函数
const formatAmount = (amount: number) => {
  return (amount / 1e18).toFixed(2)
}

const formatGasPrice = (gasPrice: number) => {
  return (gasPrice / 1e9).toFixed(1)
}

const formatDifficulty = (difficulty: number) => {
  if (difficulty >= 1e12) {
    return `${(difficulty / 1e12).toFixed(2)} T`
  } else if (difficulty >= 1e9) {
    return `${(difficulty / 1e9).toFixed(2)} G`
  } else if (difficulty >= 1e6) {
    return `${(difficulty / 1e6).toFixed(2)} M`
  } else {
    return difficulty.toLocaleString()
  }
}

const formatChange = (change: number) => {
  const sign = change >= 0 ? '+' : ''
  return `${sign}${change.toFixed(2)}%`
}

const getChangeClass = (change: number) => {
  if (change > 0) {
    return 'text-green-600'
  } else if (change < 0) {
    return 'text-red-600'
  } else {
    return 'text-gray-600'
  }
}

// 数据加载
const loadData = () => {
  // 模拟统计数据
  stats.value = {
    dailyVolume: 150000e18, // 150,000 BNB
    avgGasPrice: 25e9, // 25 Gwei
    avgBlockTime: 12.5,
    difficulty: 2.5e12 // 2.5 T
  }

  // 模拟网络指标
  networkMetrics.value = [
    {
      name: '总区块数',
      currentValue: '18,456,789',
      change24h: 2.1,
      change7d: 15.3
    },
    {
      name: '总交易数',
      currentValue: '987,654,321',
      change24h: 1.8,
      change7d: 12.7
    },
    {
      name: '活跃地址数',
      currentValue: '1,234,567',
      change24h: -0.5,
      change7d: 8.2
    },
    {
      name: '网络算力',
      currentValue: '2.5 TH/s',
      change24h: 0.3,
      change7d: 5.1
    },
    {
      name: '平均区块大小',
      currentValue: '1.2 MB',
      change24h: 1.2,
      change7d: 9.8
    },
    {
      name: '平均交易费用',
      currentValue: '0.002 BNB',
      change24h: -2.1,
      change7d: -8.5
    }
  ]

  lastUpdateTime.value = new Date().toLocaleString()
}

// 组件挂载时加载数据
onMounted(() => {
  loadData()
  
  // 每5分钟更新一次数据
  setInterval(() => {
    loadData()
  }, 300000)
})
</script> 