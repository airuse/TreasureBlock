<template>
  <div class="flex items-center gap-4 bg-white rounded-lg px-4 py-3 shadow-sm border border-gray-200">
    <slot name="label">
      <!-- 默认不显示label -->
    </slot>
    <div class="relative flex-1">
      <input
        :placeholder="placeholder"
        v-model="modelValueProxy"
        @input="$emit('update:modelValue', modelValueProxy)"
        class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
        type="text"
      />
      <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
        <svg class="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </div>
    </div>
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, defineProps, defineEmits } from 'vue'

const props = defineProps({
  modelValue: String,
  placeholder: {
    type: String,
    default: '请输入关键词...'
  }
})
const emit = defineEmits(['update:modelValue'])

const modelValueProxy = ref(props.modelValue)
watch(() => props.modelValue, v => modelValueProxy.value = v)
</script> 