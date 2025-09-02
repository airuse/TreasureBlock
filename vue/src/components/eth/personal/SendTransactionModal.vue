<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        <div class="px-6 py-4 border-b border-gray-200">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-medium text-gray-900">发送交易</h3>
            <button
              @click="$emit('close')"
              class="text-gray-400 hover:text-gray-600 transition-colors"
            >
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
        
        <div class="px-6 py-4">
          <!-- 交易信息概览 -->
          <div class="mb-6 p-4 bg-gray-50 rounded-lg">
            <h4 class="text-sm font-medium text-gray-700 mb-3">交易信息</h4>
            <div class="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span class="text-gray-500">发送地址:</span>
                <div class="font-mono text-gray-900 break-all">{{ transaction.from_address }}</div>
              </div>
              <div>
                <span class="text-gray-500">接收地址:</span>
                <div class="font-mono text-gray-900 break-all">{{ transaction.to_address }}</div>
              </div>
              <div>
                <span class="text-gray-500">交易金额:</span>
                <div class="text-gray-900">{{ transaction.amount }} {{ transaction.symbol }}</div>
              </div>
              <div>
                <span class="text-gray-500">Gas限制:</span>
                <div class="text-gray-900">{{ transaction.gas_limit }}</div>
              </div>
            </div>
          </div>

          <!-- 手续费设置 -->
          <div class="mb-6">
            <h4 class="text-sm font-medium text-gray-700 mb-3">手续费设置 (Type2)</h4>
            
            <!-- 手续费模式选择 -->
            <div class="mb-4">
              <div class="flex space-x-4">
                <label class="flex items-center">
                  <input
                    type="radio"
                    v-model="feeMode"
                    value="auto"
                    class="mr-2 text-blue-600"
                  />
                  <span class="text-sm text-gray-700">自动模式</span>
                </label>
                <label class="flex items-center">
                  <input
                    type="radio"
                    v-model="feeMode"
                    value="manual"
                    class="mr-2 text-blue-600"
                  />
                  <span class="text-sm text-gray-700">手动模式</span>
                </label>
              </div>
            </div>

            <!-- 自动模式 -->
            <div v-if="feeMode === 'auto'" class="space-y-3">
              <div class="grid grid-cols-3 gap-3">
                <label class="flex items-center p-3 border border-gray-200 rounded-lg cursor-pointer hover:border-blue-300">
                  <input
                    type="radio"
                    v-model="autoFeeSpeed"
                    value="slow"
                    class="mr-2 text-blue-600"
                  />
                  <div>
                    <div class="font-medium text-gray-900">慢速</div>
                    <div class="text-xs text-gray-500">{{ autoFeeRates.slow }} Gwei</div>
                  </div>
                </label>
                
                <label class="flex items-center p-3 border border-gray-200 rounded-lg cursor-pointer hover:border-blue-300">
                  <input
                    type="radio"
                    v-model="autoFeeSpeed"
                    value="normal"
                    class="mr-2 text-blue-600"
                  />
                  <div>
                    <div class="font-medium text-gray-900">普通</div>
                    <div class="text-xs text-gray-500">{{ autoFeeRates.normal }} Gwei</div>
                  </div>
                </label>
                
                <label class="flex items-center p-3 border border-gray-200 rounded-lg cursor-pointer hover:border-blue-300">
                  <input
                    type="radio"
                    v-model="autoFeeSpeed"
                    value="fast"
                    class="mr-2 text-blue-600"
                  />
                  <div>
                    <div class="font-medium text-gray-900">快速</div>
                    <div class="text-xs text-gray-500">{{ autoFeeRates.fast }} Gwei</div>
                  </div>
                </label>
              </div>
              
              <!-- 自动模式费用预览 -->
              <div class="p-3 bg-blue-50 border border-blue-200 rounded-md">
                <div class="text-sm text-blue-800">
                  <div class="flex justify-between">
                    <span>矿工费:</span>
                    <span>{{ calculateAutoMinerFee }} ETH</span>
                  </div>
                  <div class="flex justify-between">
                    <span>最大手续费:</span>
                    <span>{{ calculateAutoMaxFee }} ETH</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- 手动模式 -->
            <div v-if="feeMode === 'manual'" class="space-y-4">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">矿工费 (Gwei)</label>
                  <input
                    v-model="manualFee.maxPriorityFeePerGas"
                    type="number"
                    step="0.1"
                    class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="1.5"
                  />
                </div>
                
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">最大手续费 (Gwei)</label>
                  <input
                    v-model="manualFee.maxFeePerGas"
                    type="number"
                    step="0.1"
                    class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="20"
                  />
                </div>
              </div>
              
              <!-- 手动模式费用预览 -->
              <div class="p-3 bg-green-50 border border-green-200 rounded-md">
                <div class="text-sm text-green-800">
                  <div class="flex justify-between">
                    <span>矿工费:</span>
                    <span>{{ calculateManualMinerFee }} ETH</span>
                  </div>
                  <div class="flex justify-between">
                    <span>最大手续费:</span>
                    <span>{{ calculateManualMaxFee }} ETH</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 交易确认 -->
          <div class="mb-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
            <div class="flex items-start">
              <svg class="h-5 w-5 text-yellow-400 mr-2 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
              </svg>
              <div class="text-sm text-yellow-800">
                <div class="font-medium mb-1">确认交易</div>
                <div>请仔细检查交易信息，确认无误后点击发送。交易发送后将无法撤销。</div>
              </div>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div class="flex justify-end space-x-3">
            <button
              type="button"
              @click="$emit('close')"
              class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 transition-colors"
            >
              取消
            </button>
            <button
              @click="handleSendTransaction"
              :disabled="isSending"
              class="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {{ isSending ? '发送中...' : '发送交易' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { sendTransaction } from '@/api/user-transactions'
import type { UserTransaction } from '@/types'

interface Props {
  show: boolean
  transaction: UserTransaction
}

interface Emits {
  (e: 'close'): void
  (e: 'sent', transaction: any): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 手续费模式
const feeMode = ref<'auto' | 'manual'>('auto')
const autoFeeSpeed = ref<'slow' | 'normal' | 'fast'>('normal')

// 自动模式费率（Gwei）
const autoFeeRates = {
  slow: 1.5,
  normal: 2.0,
  fast: 2.5
}

// 手动模式费率
const manualFee = ref({
  maxPriorityFeePerGas: '1.5', // 矿工费
  maxFeePerGas: '20' // 最大手续费
})

// 发送状态
const isSending = ref(false)

// 计算属性
const calculateAutoMinerFee = computed(() => {
  const gasPrice = autoFeeRates[autoFeeSpeed.value]
  const gasLimit = props.transaction.gas_limit || 21000
  const fee = (gasPrice * gasLimit) / 1e9
  return fee.toFixed(6)
})

const calculateAutoMaxFee = computed(() => {
  const gasPrice = autoFeeRates[autoFeeSpeed.value] * 1.1 // 增加10%作为最大手续费
  const gasLimit = props.transaction.gas_limit || 21000
  const fee = (gasPrice * gasLimit) / 1e9
  return fee.toFixed(6)
})

const calculateManualMinerFee = computed(() => {
  const maxPriorityFee = parseFloat(manualFee.value.maxPriorityFeePerGas) || 0
  const gasLimit = props.transaction.gas_limit || 21000
  const fee = (maxPriorityFee * gasLimit) / 1e9
  return fee.toFixed(6)
})

const calculateManualMaxFee = computed(() => {
  const maxFee = parseFloat(manualFee.value.maxFeePerGas) || 0
  const gasLimit = props.transaction.gas_limit || 21000
  const fee = (maxFee * gasLimit) / 1e9
  return fee.toFixed(6)
})

// 监听手续费模式变化
watch(feeMode, (newMode) => {
  if (newMode === 'auto') {
    autoFeeSpeed.value = 'normal'
  }
})

// 发送交易
const handleSendTransaction = async () => {
  try {
    isSending.value = true
    
    // 准备手续费数据
    let feeData = {}
    if (feeMode.value === 'auto') {
      const gasPrice = autoFeeRates[autoFeeSpeed.value]
      feeData = {
        maxPriorityFeePerGas: gasPrice.toString(),
        maxFeePerGas: (gasPrice * 1.1).toString()
      }
    } else {
      feeData = {
        maxPriorityFeePerGas: manualFee.value.maxPriorityFeePerGas,
        maxFeePerGas: manualFee.value.maxFeePerGas
      }
    }
    
    // 调用发送交易API
    const response = await sendTransaction(props.transaction.id)
    
    if (response.success) {
      emit('sent', response.data)
      emit('close')
    } else {
      alert('发送交易失败: ' + response.message)
    }
  } catch (error) {
    console.error('发送交易失败:', error)
    alert('发送交易失败，请重试')
  } finally {
    isSending.value = false
  }
}
</script>
