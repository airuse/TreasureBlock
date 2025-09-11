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
            <div class="w-3 h-3 bg-orange-500 rounded-full"></div>
            <span class="text-sm text-gray-600">BTC 网络</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 收益概览（对齐 ETH 页面） -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">收益概览</h3>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div class="text-center">
            <div class="text-3xl font-bold text-green-600">{{ currentBalance }}</div>
            <div class="text-sm text-gray-500">当前余额 (TB)</div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-blue-600">{{ todayEarnings }}</div>
            <div class="text-sm text-gray-500">今日收益 (TB)</div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-purple-600">{{ totalTransactionCount }}</div>
            <div class="text-sm text-gray-500">总扫块交易数</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 收益图表（对齐 ETH 页面，使用 SVG 渲染） -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">近1小时收益趋势</h3>
        <div class="h-96 bg-gray-50 rounded-lg p-6">
          <div ref="earningsChart" class="w-full h-full"></div>
        </div>
      </div>
    </div>

    <!-- 收益记录（对齐 ETH 列结构） -->
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
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">区块高度</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">交易数量</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">收益 (TB)</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">余额变化</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="earning in earningsList" :key="earning.id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {{ formatTime(new Date(earning.created_at)) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <span class="font-mono text-blue-600">{{ earning.block_height || 0 }}</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <span class="font-medium">{{ earning.transaction_count || 0 }}</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-green-600">
                  +{{ earning.amount }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                  <span class="text-xs">{{ earning.balance_before }} → {{ earning.balance_after }}</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                  <button
                    @click="viewBlockDetails(earning.block_height || 0)"
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
import { ref, onMounted, watch, nextTick, onUnmounted } from 'vue'
import { 
  getUserBalance, 
  getUserEarningsRecords, 
  getUserEarningsStats,
  getEarningsTrend
} from '@/api/earnings'
import type { 
  UserBalance, 
  EarningRecord, 
  EarningsStats,
  EarningsTrendPoint
} from '@/types/earnings'

// 响应式数据（对齐 ETH 页面）
const currentBalance = ref(0)
const todayEarnings = ref(0)
const totalTransactionCount = ref(0)
const selectedPeriod = ref(30)
const currentPage = ref(1)
const pageSize = ref(10)
const totalItems = ref(0)
const totalPages = ref(0)

// 收益记录列表
const earningsList = ref<EarningRecord[]>([])

// 趋势图 DOM 引用
const earningsChart = ref<HTMLDivElement>()

// 自动刷新控制（与 ETH 保持一致）
const refreshTimer = ref<NodeJS.Timeout | null>(null)
const REFRESH_INTERVAL = 30 * 1000

// 创建 SVG 趋势图（复用 ETH 的思路）
const createSVGChart = (labels: string[], data: number[]) => {
  if (data.length === 0) {
    return `
      <div class="flex items-center justify-center h-full">
        <div class="text-center text-gray-500">
          <svg class="mx-auto h-12 w-12 text-gray-400 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2zm0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          <p>暂无收益数据</p>
          <p class="text-sm">近1小时内没有扫块收益记录</p>
        </div>
      </div>
    `
  }

  // 生成完整1小时轴并映射（与 ETH 一致）
  const generateFullTimeAxis = () => {
    const now = new Date()
    const timePoints: string[] = []
    for (let i = 59; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 60 * 1000)
      const hour = time.getHours().toString().padStart(2, '0')
      const minute = time.getMinutes().toString().padStart(2, '0')
      timePoints.push(`${hour}:${minute}`)
    }
    return timePoints
  }

  const fullTimeAxis = generateFullTimeAxis()
  const mappedData = fullTimeAxis.map(timePoint => {
    const idx = labels.findIndex(label => label === timePoint)
    return idx >= 0 ? data[idx] : 0
  })

  const parent = earningsChart.value
  const containerWidth = parent ? parent.clientWidth : 800
  const containerHeight = parent ? parent.clientHeight : 300
  const padding = { top: 30, right: 40, bottom: 40, left: 50 }
  const chartWidth = containerWidth - padding.left - padding.right
  const chartHeight = containerHeight - padding.top - padding.bottom

  const maxValue = Math.max(...mappedData) || 100
  const minValue = 0

  const points = mappedData.map((value, index) => {
    const x = padding.left + (index / (fullTimeAxis.length - 1)) * chartWidth
    const y = padding.top + chartHeight - ((value - minValue) / (maxValue - minValue)) * chartHeight
    return `${x},${y}`
  }).join(' ')

  const areaPoints = [
    ...points.split(' ').map(point => point.split(',')[0] + ',' + point.split(',')[1]),
    ...points.split(' ').reverse().map(point => point.split(',')[0] + ',' + (containerHeight - padding.bottom))
  ].join(' ')

  const yAxisLabels = Array.from({length: 6}, (_, i) => {
    const value = minValue + (i / 5) * (maxValue - minValue)
    const y = padding.top + ((5 - i) / 5) * chartHeight
    return { value: Math.round(value), y }
  })

  return `
    <svg width="${containerWidth}" height="${containerHeight}" style="cursor: crosshair;">
      <defs>
        <linearGradient id="areaGradient" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" style="stop-color:rgba(59,130,246,0.4);stop-opacity:1" />
          <stop offset="100%" style="stop-color:rgba(59,130,246,0.05);stop-opacity:1" />
        </linearGradient>
      </defs>

      <g stroke="rgba(0,0,0,0.08)" stroke-width="1" fill="none">
        ${yAxisLabels.map(label => 
          `<line x1="${padding.left}" y1="${label.y}" x2="${containerWidth - padding.right}" y2="${label.y}" />`
        ).join('')}
      </g>

      <g>
        ${yAxisLabels.map(label => 
          `<text x="${padding.left - 8}" y="${label.y + 4}" text-anchor="end" font-size="10" fill="#6b7280">${label.value}</text>`
        ).join('')}
      </g>

      <polygon points="${areaPoints}" fill="url(#areaGradient)" />
      <polyline points="${points}" fill="none" stroke="rgb(59,130,246)" stroke-width="4" stroke-linecap="round" stroke-linejoin="round" />

      ${mappedData.map((value, index) => {
        if (value === 0) return ''
        const x = padding.left + (index / (fullTimeAxis.length - 1)) * chartWidth
        const y = padding.top + chartHeight - ((value - minValue) / (maxValue - minValue)) * chartHeight
        return `<circle cx="${x}" cy="${y}" r="5" fill="white" stroke="rgb(59,130,246)" stroke-width="3" />`
      }).join('')}

      ${fullTimeAxis.map((label, index) => {
        if (index % 2 !== 0) return ''
        const x = padding.left + (index / (fullTimeAxis.length - 1)) * chartWidth
        const isDataPoint = mappedData[index] > 0
        const color = isDataPoint ? '#1f2937' : '#9ca3af'
        return `<text x="${x}" y="${containerHeight - padding.bottom + 14}" text-anchor="middle" font-size="10" fill="${color}">${label}</text>`
      }).join('')}
    </svg>
  `
}

// 聚合趋势数据（复用 ETH 思路）
const aggregateTrendData = (trendData: any[]) => {
  const timeGroupMap = new Map<string, { amount: number; count: number }>()
  trendData.forEach(point => {
    const timestamp = point.timestamp
    const amount = point.amount || 0
    if (timeGroupMap.has(timestamp)) {
      const existing = timeGroupMap.get(timestamp)!
      existing.amount += amount
      existing.count += 1
    } else {
      timeGroupMap.set(timestamp, { amount, count: 1 })
    }
  })
  const aggregatedArray = Array.from(timeGroupMap.entries()).map(([timestamp, data]) => ({
    timestamp,
    amount: data.amount
  }))
  aggregatedArray.sort((a, b) => {
    const timeA = new Date(`2000-01-01 ${a.timestamp}`)
    const timeB = new Date(`2000-01-01 ${b.timestamp}`)
    return timeA.getTime() - timeB.getTime()
  })
  return aggregatedArray
}

// 创建趋势图
const createEarningsChart = async () => {
  try {
    const trendResponse = await getEarningsTrend(1)
    if (trendResponse.success) {
      const trendData = (trendResponse.data || []).filter((p: any) => p.source_chain === 'btc')
      const aggregated = aggregateTrendData(trendData)
      const labels = aggregated.map(p => p.timestamp)
      const data = aggregated.map(p => p.amount)
      if (earningsChart.value) {
        const svg = createSVGChart(labels, data)
        earningsChart.value.innerHTML = svg
      }
    } else {
      if (earningsChart.value) {
        earningsChart.value.innerHTML = createSVGChart([], [])
      }
    }
  } catch (e) {
    if (earningsChart.value) {
      earningsChart.value.innerHTML = createSVGChart([], [])
    }
  }
}

// 启动/停止自动刷新（保持与 ETH 一致的行为）
const startAutoRefresh = () => {
  if (refreshTimer.value) clearInterval(refreshTimer.value)
  refreshTimer.value = setInterval(async () => {
    await createEarningsChart()
  }, REFRESH_INTERVAL)
}
const stopAutoRefresh = () => {
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value)
    refreshTimer.value = null
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

// 查看区块详情（使用区块高度，保持与 ETH 列行为一致）
const viewBlockDetails = (blockHeight: number) => {
  if (!blockHeight) return
  const route = `/btc/blocks/${blockHeight}`
  window.location.href = route
}

// 分页
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

// 加载收益记录（仅 BTC）
const loadEarnings = async () => {
  const params: any = {
    page: currentPage.value,
    page_size: pageSize.value,
    period: selectedPeriod.value,
    chain: 'btc'
  }
  const resp = await getUserEarningsRecords(params)
  if (resp.success && resp.data && Array.isArray(resp.data.records)) {
    // 后端 DTO 已包含 source_chain 字段
    const btcRecords = resp.data.records.filter((r: any) => r.source_chain === 'btc') as EarningRecord[]
    earningsList.value = btcRecords
    totalItems.value = resp.data.pagination.total
    totalPages.value = Math.ceil(totalItems.value / pageSize.value)
    // 计算今日收益（基于已加载记录）
    const startOfToday = new Date(); startOfToday.setHours(0,0,0,0)
    todayEarnings.value = earningsList.value
      .filter(r => new Date(r.created_at) >= startOfToday)
      .reduce((sum, r) => sum + (r.amount || 0), 0)
  } else {
    earningsList.value = []
    totalItems.value = 0
    totalPages.value = 0
    todayEarnings.value = 0
  }
}

// 加载用户余额与统计（对齐 ETH 页面）
const loadUserData = async () => {
  try {
    const [balanceResponse, statsResponse] = await Promise.all([
      getUserBalance({ chain: 'btc' }),
      getUserEarningsStats()
    ])
    if (balanceResponse.success) {
      const balance = balanceResponse.data as unknown as UserBalance
      currentBalance.value = (balance as any)?.balance || 0
    }
    if (statsResponse.success) {
      const stats = statsResponse.data as unknown as EarningsStats
      totalTransactionCount.value = (stats as any)?.transaction_count || 0
    }
  } catch (e) {
    // 忽略错误，界面显示为0
  }
}

// 监听周期变化
watch(selectedPeriod, () => {
  currentPage.value = 1
  loadEarnings()
})

// 生命周期
onMounted(async () => {
  await loadUserData()
  await loadEarnings()
  await createEarningsChart()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>
