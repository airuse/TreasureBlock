<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">地址查询</h1>
    </div>

    <!-- 搜索框 -->
    <div class="card">
      <div class="flex flex-col sm:flex-row gap-4">
        <div class="flex-1">
          <label class="block text-sm font-medium text-gray-700 mb-2">比特币地址</label>
          <input 
            v-model="addressInput" 
            type="text" 
            placeholder="输入比特币地址 (1..., 3..., bc1...)"
            class="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <div class="flex items-end">
          <button 
            @click="searchAddress" 
            :disabled="!addressInput.trim()"
            class="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            查询
          </button>
        </div>
      </div>
    </div>

    <!-- 地址详情 -->
    <div v-if="addressData" class="space-y-6">
      <!-- 基本信息 -->
      <div class="card">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">地址详情</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <div>
            <p class="text-sm font-medium text-gray-500">余额</p>
            <p class="text-2xl font-bold text-gray-900">{{ formatAmount(addressData.balance) }} BTC</p>
            <p class="text-sm text-gray-500">${{ formatFiat(addressData.balance * 45000) }}</p>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-500">交易次数</p>
            <p class="text-2xl font-bold text-gray-900">{{ addressData.txCount.toLocaleString() }}</p>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-500">UTXO 数量</p>
            <p class="text-2xl font-bold text-gray-900">{{ addressData.utxoCount }}</p>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-500">地址类型</p>
            <p class="text-2xl font-bold text-gray-900">{{ addressData.type }}</p>
          </div>
        </div>
        <div class="mt-4 grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <p class="text-sm font-medium text-gray-500">首次活跃</p>
            <p class="text-lg text-gray-900">{{ formatTimestamp(addressData.firstSeen) }}</p>
          </div>
          <div>
            <p class="text-sm font-medium text-gray-500">最近活跃</p>
            <p class="text-lg text-gray-900">{{ formatTimestamp(addressData.lastSeen) }}</p>
          </div>
        </div>
      </div>

      <!-- 标签页切换 -->
      <div class="card">
        <div class="border-b border-gray-200">
          <nav class="-mb-px flex space-x-8">
            <button
              @click="activeTab = 'transactions'"
              :class="[
                activeTab === 'transactions'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
                'whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm'
              ]"
            >
              交易记录 ({{ addressData.txCount }})
            </button>
            <button
              @click="activeTab = 'utxos'"
              :class="[
                activeTab === 'utxos'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
                'whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm'
              ]"
            >
              UTXO 列表 ({{ addressData.utxoCount }})
            </button>
          </nav>
        </div>

        <!-- 交易记录标签页 -->
        <div v-if="activeTab === 'transactions'" class="mt-4">
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">交易哈希</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">时间</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">类型</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">金额</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">余额</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200">
                <tr v-for="tx in addressData.transactions" :key="tx.hash" class="hover:bg-gray-50">
                  <td class="px-6 py-4 whitespace-nowrap">
                    <router-link :to="`/btc/transactions/${tx.hash}`" class="text-blue-600 hover:text-blue-700 font-mono text-sm">
                      {{ formatHash(tx.hash) }}
                    </router-link>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {{ formatTimestamp(tx.timestamp) }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <span :class="[
                      tx.type === '转入' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800',
                      'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium'
                    ]">
                      {{ tx.type }}
                    </span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm">
                    <span :class="tx.type === '转入' ? 'text-green-600' : 'text-red-600'">
                      {{ tx.type === '转入' ? '+' : '-' }}{{ formatAmount(Math.abs(tx.amount)) }} BTC
                    </span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {{ formatAmount(tx.balance) }} BTC
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- UTXO 列表标签页 -->
        <div v-if="activeTab === 'utxos'" class="mt-4">
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">交易哈希</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">输出索引</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">金额</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">确认数</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">创建时间</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200">
                <tr v-for="utxo in addressData.utxos" :key="`${utxo.txHash}:${utxo.vout}`" class="hover:bg-gray-50">
                  <td class="px-6 py-4 whitespace-nowrap">
                    <router-link :to="`/btc/transactions/${utxo.txHash}`" class="text-blue-600 hover:text-blue-700 font-mono text-sm">
                      {{ formatHash(utxo.txHash) }}
                    </router-link>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {{ utxo.vout }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {{ formatAmount(utxo.amount) }} BTC
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {{ utxo.confirmations }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {{ formatTimestamp(utxo.timestamp) }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="card text-center py-12">
      <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
      </svg>
      <h3 class="mt-2 text-sm font-medium text-gray-900">输入地址开始查询</h3>
      <p class="mt-1 text-sm text-gray-500">支持 Legacy (1...)、Script Hash (3...) 和 Bech32 (bc1...) 格式</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { formatTimestamp, formatHash, formatAmount } from '@/utils/formatters'
import type { AddressData } from '@/types'

// 响应式数据
const addressInput = ref('')
const addressData = ref<AddressData | null>(null)
const activeTab = ref('transactions')

// 格式化法币金额
const formatFiat = (amount: number) => {
  return amount.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

// 获取地址类型
const getAddressType = (address: string) => {
  if (address.startsWith('1')) return 'Legacy'
  if (address.startsWith('3')) return 'Script Hash'
  if (address.startsWith('bc1')) return 'Bech32'
  return 'Unknown'
}

// 搜索地址
const searchAddress = () => {
  if (!addressInput.value.trim()) return

  // 模拟地址数据
  const txCount = Math.floor(Math.random() * 1000) + 10
  const utxoCount = Math.floor(Math.random() * 50) + 1
  const balance = Math.random() * 10 + 0.001
  
  addressData.value = {
    address: addressInput.value,
    balance: balance,
    txCount: txCount,
    utxoCount: utxoCount,
    type: getAddressType(addressInput.value),
    firstSeen: Math.floor(Date.now() / 1000) - 86400 * 365,
    lastSeen: Math.floor(Date.now() / 1000) - 3600,
    transactions: Array.from({ length: Math.min(20, txCount) }, (_, i) => {
      const isIncoming = Math.random() > 0.5
      const amount = Math.random() * 2
      return {
        hash: Array.from({ length: 64 }, () => Math.floor(Math.random() * 16).toString(16)).join(''),
        timestamp: Math.floor(Date.now() / 1000) - i * 3600 * 24,
        type: isIncoming ? '转入' : '转出',
        amount: isIncoming ? amount : -amount,
        balance: balance + (isIncoming ? amount : -amount) * Math.random()
      }
    }),
    utxos: Array.from({ length: utxoCount }, (_, i) => ({
      txHash: Array.from({ length: 64 }, () => Math.floor(Math.random() * 16).toString(16)).join(''),
      vout: i,
      amount: Math.random() * 1 + 0.001,
      confirmations: Math.floor(Math.random() * 1000) + 1,
      timestamp: Math.floor(Date.now() / 1000) - Math.random() * 86400 * 30
    }))
  }
}
</script> 