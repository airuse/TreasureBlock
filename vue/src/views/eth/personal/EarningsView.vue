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

    <!-- 收益图表 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">近1小时收益趋势</h3>
        <div class="h-96 bg-gray-50 rounded-lg p-6">
          <div ref="earningsChart" class="w-full h-full"></div>
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
                  <span class="font-mono text-blue-600">{{ earning.block_height }}</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  <span class="font-medium">{{ earning.transaction_count }}</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-green-600">
                  +{{ earning.amount }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                  <span class="text-xs">
                    {{ earning.balance_before }} → {{ earning.balance_after }}
                  </span>
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
  getEarningsRecordDetail,
  getEarningsTrend
} from '@/api/earnings'
import type { 
  UserBalance, 
  EarningRecord, 
  EarningsStats,
  EarningsTrendPoint
} from '@/types/earnings'
import { showSuccess, showError } from '@/composables/useToast'

// 响应式数据
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

// 收益趋势图表引用
const earningsChart = ref<HTMLDivElement>()

// 定时刷新相关
const refreshTimer = ref<NodeJS.Timeout | null>(null)
const REFRESH_INTERVAL = 30 * 1000 // 30秒

// 创建收益趋势图表
const createEarningsChart = async () => {
  try {
    // 调用专门的趋势接口获取数据
    const trendResponse = await getEarningsTrend(1) // 改为1小时
    
    if (trendResponse.success) {
      const trendData = trendResponse.data || []
      
      // 数据累加处理：按时间戳分组并累加amount
      const aggregatedData = aggregateTrendData(trendData)
      
      // 准备图表数据
      const labels = aggregatedData.map(point => point.timestamp)
      const data = aggregatedData.map(point => point.amount)
      
      // console.log('📊 原始数据点数量:', trendData.length)
      // console.log('📊 累加后数据点数量:', aggregatedData.length)
      // console.log('📊 累加后的数据:', aggregatedData)
      
      // 创建简单的SVG图表
      if (earningsChart.value) {
        const svg = createSVGChart(labels, data)
        earningsChart.value.innerHTML = svg
      }
    } else {
      console.error('获取收益趋势数据失败:', (trendResponse as any).error || trendResponse.message)
      // 显示空数据提示
      if (earningsChart.value) {
        earningsChart.value.innerHTML = createSVGChart([], [])
      }
    }
  } catch (error) {
    console.error('创建收益趋势图表失败:', error)
    // 显示错误提示
    if (earningsChart.value) {
      earningsChart.value.innerHTML = createSVGChart([], [])
    }
  }
}

// 数据累加处理函数
const aggregateTrendData = (trendData: any[]) => {
  // 使用Map按时间戳分组
  const timeGroupMap = new Map<string, { amount: number; count: number; blockHeights: number[]; transactionCounts: number[] }>()
  
  trendData.forEach(point => {
    const timestamp = point.timestamp
    const amount = point.amount || 0
    
    if (timeGroupMap.has(timestamp)) {
      // 累加已存在的时间戳数据
      const existing = timeGroupMap.get(timestamp)!
      existing.amount += amount
      existing.count += 1
      existing.blockHeights.push(point.block_height)
      existing.transactionCounts.push(point.transaction_count)
    } else {
      // 创建新的时间戳分组
      timeGroupMap.set(timestamp, {
        amount: amount,
        count: 1,
        blockHeights: [point.block_height],
        transactionCounts: [point.transaction_count]
      })
    }
  })
  
  // 转换为数组并按时间排序
  const aggregatedArray = Array.from(timeGroupMap.entries()).map(([timestamp, data]) => ({
    timestamp,
    amount: data.amount,
    count: data.count,
    blockHeights: data.blockHeights,
    transactionCounts: data.transactionCounts,
    // 计算平均值（可选）
    avgAmount: Math.round(data.amount / data.count),
    totalTransactions: data.transactionCounts.reduce((sum, count) => sum + count, 0)
  }))
  
  // 按时间戳排序（HH:MM格式）
  aggregatedArray.sort((a, b) => {
    const timeA = new Date(`2000-01-01 ${a.timestamp}`)
    const timeB = new Date(`2000-01-01 ${b.timestamp}`)
    return timeA.getTime() - timeB.getTime()
  })
  
  return aggregatedArray
}

// 启动定时刷新
const startAutoRefresh = () => {
  // 清除可能存在的旧定时器
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value)
  }
  
  // 设置新的定时器，每30秒刷新一次
  refreshTimer.value = setInterval(async () => {
    // console.log('🔄 自动刷新收益趋势图表...')
    await createEarningsChart()
  }, REFRESH_INTERVAL)
  
  // console.log('✅ 收益趋势图表自动刷新已启动，每30秒刷新一次')
}

// 停止定时刷新
const stopAutoRefresh = () => {
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value)
    refreshTimer.value = null
    // console.log('⏹️ 收益趋势图表自动刷新已停止')
  }
}

// 创建SVG图表
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
  
  // 生成完整的1小时时间轴（每1分钟一个点）
  const generateFullTimeAxis = () => {
    const now = new Date()
    const timePoints = []
    for (let i = 59; i >= 0; i--) { // 60个点，覆盖1小时
      const time = new Date(now.getTime() - i * 60 * 1000) // 每1分钟
      const hour = time.getHours()
      const minute = time.getMinutes()
      timePoints.push(`${hour.toString().padStart(2, '0')}:${minute.toString().padStart(2, '0')}`)
    }
    return timePoints
  }
  
  const fullTimeAxis = generateFullTimeAxis()
  
  // 将实际数据映射到完整时间轴上
  const mappedData = fullTimeAxis.map(timePoint => {
    const dataIndex = labels.findIndex(label => label === timePoint)
    return dataIndex >= 0 ? data[dataIndex] : 0 // 没有数据的时间点设为0
  })
  
  // 图表尺寸 - 使用容器真实像素尺寸，避免字体拉伸失真
  // 父容器是 Tailwind 的 h-96 (384px) 并有 p-6 (24px) 的内边距
  const parent = earningsChart.value
  const containerWidth = parent ? parent.clientWidth : 800
  const containerHeight = parent ? parent.clientHeight : 300
  const padding = { top: 30, right: 40, bottom: 40, left: 50 }
  const chartWidth = containerWidth - padding.left - padding.right
  const chartHeight = containerHeight - padding.top - padding.bottom
  
  const maxValue = Math.max(...mappedData) || 100
  const minValue = 0 // 从0开始，确保没有数据的时间点也能显示
  
  // 创建路径点
  const points = mappedData.map((value, index) => {
    const x = padding.left + (index / (fullTimeAxis.length - 1)) * chartWidth
    const y = padding.top + chartHeight - ((value - minValue) / (maxValue - minValue)) * chartHeight
    return `${x},${y}`
  }).join(' ')
  
  // 创建区域填充路径
  const areaPoints = [
    ...points.split(' ').map(point => point.split(',')[0] + ',' + point.split(',')[1]),
    ...points.split(' ').reverse().map(point => point.split(',')[0] + ',' + (containerHeight - padding.bottom))
  ].join(' ')
  
  // 生成Y轴标签 - 修正排序：最下边是0，最上边是最大值
  const yAxisLabels = Array.from({length: 6}, (_, i) => {
    const value = minValue + (i / 5) * (maxValue - minValue)
    // 修正Y坐标：i=0时y最大（顶部），i=5时y最小（底部）
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
      
      <!-- 背景网格线 -->
      <g stroke="rgba(0,0,0,0.08)" stroke-width="1" fill="none">
        ${yAxisLabels.map(label => 
          `<line x1="${padding.left}" y1="${label.y}" x2="${containerWidth - padding.right}" y2="${label.y}" />`
        ).join('')}
      </g>
      
      <!-- Y轴标签 -->
      <g>
        ${yAxisLabels.map(label => 
          `<text x="${padding.left - 8}" y="${label.y + 4}" text-anchor="end" font-size="10" fill="#6b7280">${label.value}</text>`
        ).join('')}
      </g>
      
      <!-- 区域填充 -->
      <polygon points="${areaPoints}" fill="url(#areaGradient)" />
      
      <!-- 折线 -->
      <polyline points="${points}" fill="none" stroke="rgb(59,130,246)" stroke-width="4" stroke-linecap="round" stroke-linejoin="round" />
      
      <!-- 数据点（只显示有数据的时间点） -->
      ${mappedData.map((value, index) => {
        if (value === 0) return '' // 跳过没有数据的时间点
        const x = padding.left + (index / (fullTimeAxis.length - 1)) * chartWidth
        const y = padding.top + chartHeight - ((value - minValue) / (maxValue - minValue)) * chartHeight
        return `<circle 
          cx="${x}" 
          cy="${y}" 
          r="5" 
          fill="white" 
          stroke="rgb(59,130,246)" 
          stroke-width="3"
          style="cursor: pointer; transition: all 0.2s ease;"
          data-time="${fullTimeAxis[index]}"
          data-value="${value}"
        />`
      }).join('')}
      
      <!-- X轴标签（30个单位间隔，每个间隔内放两个点） -->
      ${fullTimeAxis.map((label, index) => {
        // 每2个点显示一个标签，实现30个单位间隔
        if (index % 2 !== 0) return ''
        const x = padding.left + (index / (fullTimeAxis.length - 1)) * chartWidth
        const isDataPoint = mappedData[index] > 0
        const color = isDataPoint ? '#1f2937' : '#9ca3af'
        return `<text x="${x}" y="${containerHeight - padding.bottom + 14}" text-anchor="middle" font-size="10" fill="${color}">${label}</text>`
      }).join('')}
      
      
    </svg>
  `
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
const viewBlockDetails = (blockHeight: number) => {
  if (!blockHeight) {
    showError('无法查看区块详情：区块高度无效')
    return
  }
  // 跳转到区块详情页面
  const route = `/eth/blocks/${blockHeight}`
  console.log('跳转到区块详情:', route)
  // 使用 Vue Router 进行导航
  window.location.href = route
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
  try {
    // 加载收益记录列表
    const recordsResponse = await getUserEarningsRecords({
      page: currentPage.value,
      page_size: pageSize.value,
      period: selectedPeriod.value
    })
    
    if (recordsResponse.success) {
      // console.log('🔍 收益记录响应数据:', recordsResponse)
      // console.log('🔍 收益记录数据类型:', typeof recordsResponse.data)
      // console.log('🔍 收益记录列表:', recordsResponse.data)
      // console.log('🔍 收益记录是否为数组:', Array.isArray(recordsResponse.data))
      
      // 安全检查：后端返回的是 {pagination: {...}, records: Array}
      if (!recordsResponse.data || !Array.isArray(recordsResponse.data.records)) {
        earningsList.value = []
        totalItems.value = 0
        totalPages.value = 0
        return
      }
      
      // 直接使用后端返回的数据，类型已经匹配
      earningsList.value = recordsResponse.data.records
      
      totalItems.value = recordsResponse.data.pagination.total
      totalPages.value = Math.ceil(totalItems.value / pageSize.value)
      
      // console.log('🔍 转换后的收益记录:', earningsList.value)
    } else {
      showError(`获取收益记录失败: ${recordsResponse.message || '未知错误'}`)
    }
  } catch (error) {
    console.error('加载收益数据失败:', error)
    showError(`加载收益数据失败: ${error instanceof Error ? error.message : '未知错误'}`)
  }
}

// 监听周期变化
watch(selectedPeriod, () => {
  currentPage.value = 1
  loadEarnings()
})

// 加载用户余额和统计数据
const loadUserData = async () => {
  try {
    // 并行加载用户余额和收益统计
    const [balanceResponse, statsResponse] = await Promise.all([
      getUserBalance(),
      getUserEarningsStats()
    ])
    
    if (balanceResponse.success) {
      const balance = balanceResponse.data
      // console.log('🔍 用户余额数据:', balance)
      
      // 设置当前余额
      currentBalance.value = balance.balance || 0
      // 暂时使用总收益，后续可以从统计接口获取今日数据
      todayEarnings.value = balance.total_earned || 0
    }
    
    if (statsResponse.success) {
      const stats = statsResponse.data
      // console.log('🔍 收益统计数据:', stats)
      
      // 设置总扫块交易数
      totalTransactionCount.value = stats.transaction_count || 0
      
      // 如果余额接口没有返回今日收益，使用统计接口
      if (todayEarnings.value === 0) {
        todayEarnings.value = stats.total_earnings || 0
      }
    }
  } catch (error) {
    console.error('加载用户数据失败:', error)
    showError(`加载用户数据失败: ${error instanceof Error ? error.message : '未知错误'}`)
  }
}

// 组件挂载时加载数据
onMounted(async () => {
  await loadUserData()
  await loadEarnings()
  await createEarningsChart() // 独立加载图表数据
  
  // 启动自动刷新
  startAutoRefresh()
})

// 组件卸载时清理定时器
onUnmounted(() => {
  stopAutoRefresh()
})
</script>
