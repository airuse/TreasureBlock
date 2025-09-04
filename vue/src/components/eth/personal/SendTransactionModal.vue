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

          <!-- 费率信息显示 -->
          <div class="mb-6 p-4 bg-gray-50 rounded-lg">
            <h4 class="text-sm font-medium text-gray-700 mb-3">交易费率信息</h4>
            <div class="text-sm text-gray-600">
              <div class="flex justify-between">
                <span>矿工费:</span>
                <span>{{ transaction.max_priority_fee_per_gas || '2' }} Gwei</span>
              </div>
              <div class="flex justify-between">
                <span>最大手续费:</span>
                <span>{{ transaction.max_fee_per_gas || '30' }} Gwei</span>
              </div>
              <div class="mt-2 text-xs text-gray-500">
                费率已在导出交易时设置，无法修改
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
import { ref } from 'vue'
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

// 发送状态
const isSending = ref(false)

// 发送交易
const handleSendTransaction = async () => {
  try {
    isSending.value = true
    
    // 调用发送交易API（费率已在导出时设置）
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
