<template>
  <div class="space-y-6">
    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <div class="bg-white overflow-hidden shadow rounded-lg">
        <div class="p-5">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <CurrencyDollarIcon class="h-6 w-6 text-green-400" />
            </div>
            <div class="ml-5 w-0 flex-1">
              <dl>
                <dt class="text-sm font-medium text-gray-500 truncate">总收益 (TB)</dt>
                <dd class="text-lg font-medium text-gray-900">{{ totalEarnings }}</dd>
              </dl>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white overflow-hidden shadow rounded-lg">
        <div class="p-5">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <MapPinIcon class="h-6 w-6 text-blue-400" />
            </div>
            <div class="ml-5 w-0 flex-1">
              <dl>
                <dt class="text-sm font-medium text-gray-500 truncate">管理地址</dt>
                <dd class="text-lg font-medium text-gray-900">{{ addressCount }}</dd>
              </dl>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white overflow-hidden shadow rounded-lg">
        <div class="p-5">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <ClockIcon class="h-6 w-6 text-yellow-400" />
            </div>
            <div class="ml-5 w-0 flex-1">
              <dl>
                <dt class="text-sm font-medium text-gray-500 truncate">待处理交易</dt>
                <dd class="text-lg font-medium text-gray-900">{{ pendingTransactions }}</dd>
              </dl>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white overflow-hidden shadow rounded-lg">
        <div class="p-5">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <CheckCircleIcon class="h-6 w-6 text-green-400" />
            </div>
            <div class="ml-5 w-0 flex-1">
              <dl>
                <dt class="text-sm font-medium text-gray-500 truncate">已完成交易</dt>
                <dd class="text-lg font-medium text-gray-900">{{ completedTransactions }}</dd>
              </dl>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 最近活动 -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">最近活动</h3>
        <div class="flow-root">
          <ul class="-mb-8">
            <li v-for="(activity, index) in recentActivities" :key="activity.id">
              <div class="relative pb-8">
                <span v-if="index !== recentActivities.length - 1" class="absolute top-4 left-4 -ml-px h-full w-0.5 bg-gray-200" aria-hidden="true"></span>
                <div class="relative flex space-x-3">
                  <div>
                    <span class="h-8 w-8 rounded-full flex items-center justify-center ring-8 ring-white"
                          :class="getActivityIconClass(activity.type)">
                      <component :is="getActivityIcon(activity.type)" class="h-5 w-5 text-white" />
                    </span>
                  </div>
                  <div class="min-w-0 flex-1 pt-1.5 flex justify-between space-x-4">
                    <div>
                      <p class="text-sm text-gray-500">{{ activity.description }}</p>
                    </div>
                    <div class="text-right text-sm whitespace-nowrap text-gray-500">
                      <time :datetime="activity.timestamp.toISOString()">{{ formatTime(activity.timestamp) }}</time>
                    </div>
                  </div>
                </div>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  CurrencyDollarIcon,
  MapPinIcon,
  ClockIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon
} from '@heroicons/vue/24/outline'

// 响应式数据
const totalEarnings = ref(0)
const addressCount = ref(0)
const pendingTransactions = ref(0)
const completedTransactions = ref(0)
const recentActivities = ref([
  {
    id: 1,
    type: 'earnings',
    description: '扫块成功，获得 0.5 TB 收益',
    timestamp: new Date(Date.now() - 1000 * 60 * 30) // 30分钟前
  },
  {
    id: 2,
    type: 'transaction',
    description: '交易 0x1234...5678 已确认',
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 2) // 2小时前
  },
  {
    id: 3,
    type: 'address',
    description: '新增地址 0xabcd...efgh',
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24) // 1天前
  }
])

// 获取活动图标
const getActivityIcon = (type: string) => {
  switch (type) {
    case 'earnings': return CurrencyDollarIcon
    case 'transaction': return CheckCircleIcon
    case 'address': return MapPinIcon
    default: return ExclamationTriangleIcon
  }
}

// 获取活动图标样式
const getActivityIconClass = (type: string) => {
  switch (type) {
    case 'earnings': return 'bg-green-500'
    case 'transaction': return 'bg-blue-500'
    case 'address': return 'bg-purple-500'
    default: return 'bg-gray-500'
  }
}

// 格式化时间
const formatTime = (timestamp: Date) => {
  const now = new Date()
  const diff = now.getTime() - timestamp.getTime()
  const minutes = Math.floor(diff / (1000 * 60))
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  return `${days}天前`
}

// 加载数据
const loadData = async () => {
  // TODO: 从API加载真实数据
  totalEarnings.value = 12.5
  addressCount.value = 3
  pendingTransactions.value = 2
  completedTransactions.value = 15
}

onMounted(() => {
  loadData()
})
</script>
