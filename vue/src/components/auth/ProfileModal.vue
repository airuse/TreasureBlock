<template>
  <div v-if="isVisible" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
      <!-- 头部 -->
      <div class="flex items-center justify-between p-6 border-b border-gray-200">
        <h2 class="text-xl font-semibold text-gray-900">个人中心</h2>
        <button
          @click="close"
          class="text-gray-400 hover:text-gray-600 transition-colors"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- 内容 -->
      <div class="p-6 space-y-6">
        <!-- 头像设置 -->
        <div class="flex items-center space-x-4">
          <div class="w-20 h-20 bg-blue-100 rounded-full flex items-center justify-center">
            <span class="text-blue-600 font-bold text-2xl">
              {{ authStore.user?.username?.charAt(0)?.toUpperCase() || 'U' }}
            </span>
          </div>
          <div>
            <h3 class="text-lg font-medium text-gray-900">头像</h3>
            <p class="text-sm text-gray-500">点击上传新头像</p>
            <button class="mt-2 px-4 py-2 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700">
              更换头像
            </button>
          </div>
        </div>

        <!-- 基本信息 -->
        <div class="space-y-4">
          <h3 class="text-lg font-medium text-gray-900">基本信息</h3>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">用户名</label>
              <input
                v-model="profileForm.username"
                type="text"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="请输入用户名"
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">邮箱</label>
              <input
                v-model="profileForm.email"
                type="email"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="请输入邮箱"
              />
            </div>
          </div>
        </div>

        <!-- 修改密码 -->
        <div class="space-y-4">
          <h3 class="text-lg font-medium text-gray-900">修改密码</h3>
          
          <div class="space-y-3">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">当前密码</label>
              <input
                v-model="passwordForm.currentPassword"
                type="password"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="请输入当前密码"
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">新密码</label>
              <input
                v-model="passwordForm.newPassword"
                type="password"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="请输入新密码"
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">确认新密码</label>
              <input
                v-model="passwordForm.confirmPassword"
                type="password"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="请再次输入新密码"
              />
            </div>
          </div>
        </div>

        <!-- 错误提示 -->
        <div v-if="error" class="text-red-600 text-sm bg-red-50 p-3 rounded-md">
          {{ error }}
        </div>

        <!-- 成功提示 -->
        <div v-if="success" class="text-green-600 text-sm bg-green-50 p-3 rounded-md">
          {{ success }}
        </div>
      </div>

      <!-- 底部按钮 -->
      <div class="flex justify-end space-x-3 p-6 border-t border-gray-200">
        <button
          @click="close"
          class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
        >
          取消
        </button>
        <button
          @click="saveProfile"
          :disabled="isLoading"
          class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
        >
          {{ isLoading ? '保存中...' : '保存' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{
  isVisible: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const authStore = useAuthStore()

// 响应式数据
const isLoading = ref(false)
const error = ref('')
const success = ref('')

// 表单数据
const profileForm = reactive({
  username: '',
  email: ''
})

const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 监听模态框显示状态，初始化表单数据
watch(() => props.isVisible, (visible) => {
  if (visible && authStore.user) {
    profileForm.username = authStore.user.username || ''
    profileForm.email = authStore.user.email || ''
    // 清空密码表单
    passwordForm.currentPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
    // 清空提示信息
    error.value = ''
    success.value = ''
  }
})

// 关闭模态框
const close = () => {
  emit('close')
}

// 保存个人资料
const saveProfile = async () => {
  try {
    isLoading.value = true
    error.value = ''
    success.value = ''

    // 验证密码表单
    if (passwordForm.newPassword || passwordForm.confirmPassword) {
      if (!passwordForm.currentPassword) {
        error.value = '请输入当前密码'
        return
      }
      if (passwordForm.newPassword !== passwordForm.confirmPassword) {
        error.value = '新密码与确认密码不匹配'
        return
      }
      if (passwordForm.newPassword.length < 6) {
        error.value = '新密码长度不能少于6位'
        return
      }
    }

    // 调用真实API保存个人资料和密码
    if (passwordForm.newPassword) {
      // 修改密码
      await authStore.changePassword({
        current_password: passwordForm.currentPassword,
        new_password: passwordForm.newPassword
      })
    }

    // 更新用户资料（如果需要的话，这里可以添加更新用户资料的API）
    // 目前后端没有更新用户资料的接口，所以只处理密码修改
    
    success.value = '保存成功！'
    
    // 延迟关闭模态框
    setTimeout(() => {
      close()
    }, 1500)
    
  } catch (err: any) {
    error.value = err.message || '保存失败'
  } finally {
    isLoading.value = false
  }
}
</script>
