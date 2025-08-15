<template>
  <div class="min-h-screen bg-gray-50">
    <!-- 个人中心头部 -->
    <div class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="py-6">
          <div class="flex items-center justify-between">
            <div>
              <h1 class="text-2xl font-bold text-gray-900">个人中心</h1>
              <p class="mt-1 text-sm text-gray-500">管理您的账户、收益和交易</p>
            </div>
            <div class="flex items-center space-x-4">
              <div class="flex items-center space-x-2">
                <div class="w-3 h-3 bg-orange-500 rounded-full"></div>
                <span class="text-sm text-gray-600">BTC 网络</span>
              </div>
              <div class="flex items-center space-x-2">
                <div class="w-2 h-2 rounded-full bg-green-400"></div>
                <span class="text-sm text-gray-600">已连接</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
      <div class="flex space-x-6">
        <!-- 侧边导航 -->
        <aside class="w-64 flex-shrink-0">
          <nav class="space-y-1">
            <router-link
              v-for="item in navigation"
              :key="item.name"
              :to="item.path"
              class="group flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors"
              :class="[
                $route.path === item.path
                  ? 'bg-orange-50 text-orange-700 border-r-2 border-orange-700'
                  : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
              ]"
            >
              <component
                :is="item.icon"
                class="mr-3 h-5 w-5"
                :class="[
                  $route.path === item.path
                    ? 'text-orange-500'
                    : 'text-gray-400 group-hover:text-gray-500'
                ]"
              />
              {{ item.name }}
            </router-link>
          </nav>
        </aside>

        <!-- 主内容区域 -->
        <main class="flex-1">
          <router-view />
        </main>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  HomeIcon,
  CurrencyDollarIcon,
  MapPinIcon,
  ClockIcon
} from '@heroicons/vue/24/outline'

const navigation = computed(() => [
  {
    name: '概览',
    path: '/btc/personal',
    icon: HomeIcon
  },
  {
    name: '扫块收益',
    path: '/btc/personal/earnings',
    icon: CurrencyDollarIcon
  },
  {
    name: '个人地址',
    path: '/btc/personal/addresses',
    icon: MapPinIcon
  },
  {
    name: '交易历史',
    path: '/btc/personal/transactions',
    icon: ClockIcon
  }
])
</script>
