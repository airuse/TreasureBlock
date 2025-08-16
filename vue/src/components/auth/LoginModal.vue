<template>
  <div v-if="isVisible" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
      <!-- 头部 -->
      <div class="flex items-center justify-between p-6 border-b border-gray-200">
        <h2 class="text-xl font-semibold text-gray-900">
          {{ isLogin ? '登录' : '注册' }}
        </h2>
        <button
          @click="close"
          class="text-gray-400 hover:text-gray-600 transition-colors"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- 表单 -->
      <form @submit.prevent="handleSubmit" class="p-6 space-y-4">
        <!-- 用户名 -->
        <div>
          <label for="username" class="block text-sm font-medium text-gray-700 mb-1">
            用户名
          </label>
          <input
            id="username"
            v-model="form.username"
            type="text"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="请输入用户名"
          />
        </div>

        <!-- 邮箱（仅注册时显示） -->
        <div v-if="!isLogin">
          <label for="email" class="block text-sm font-medium text-gray-700 mb-1">
            邮箱
          </label>
          <input
            id="email"
            v-model="form.email"
            type="email"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="请输入邮箱"
          />
        </div>

        <!-- 密码 -->
        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 mb-1">
            密码
          </label>
          <input
            id="password"
            v-model="form.password"
            type="password"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="请输入密码"
          />
        </div>

        <!-- 确认密码（仅注册时显示） -->
        <div v-if="!isLogin">
          <label for="confirmPassword" class="block text-sm font-medium text-gray-700 mb-1">
            确认密码
          </label>
          <input
            id="confirmPassword"
            v-model="form.confirmPassword"
            type="password"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="请再次输入密码"
          />
        </div>

        <!-- 提交按钮 -->
        <button
          type="submit"
          :disabled="isLoading"
          class="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          <span v-if="isLoading" class="flex items-center justify-center">
            <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            处理中...
          </span>
          <span v-else>
            {{ isLogin ? '登录' : '注册' }}
          </span>
        </button>

        <!-- 切换模式 -->
        <div class="text-center">
          <button
            type="button"
            @click="toggleMode"
            class="text-blue-600 hover:text-blue-800 text-sm transition-colors"
          >
            {{ isLogin ? '没有账号？点击注册' : '已有账号？点击登录' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { login, register } from '@/api/auth'
import type { LoginRequest, RegisterRequest } from '@/types/auth'
import { showSuccess, showError } from '@/composables/useToast'

interface Props {
  isVisible: boolean
  defaultMode?: 'login' | 'register'
}

interface Emits {
  (e: 'close'): void
  (e: 'success'): void
}

const props = withDefaults(defineProps<Props>(), {
  defaultMode: 'login'
})

const emit = defineEmits<Emits>()

const authStore = useAuthStore()

// 响应式数据
const isLogin = ref(props.defaultMode === 'login') // 根据props初始化
const isLoading = ref(false)

// 表单数据
const form = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

// 监听模式变化，重置表单
watch(isLogin, () => {
  resetForm()
})

// 重置表单
const resetForm = () => {
  form.username = ''
  form.email = ''
  form.password = ''
  form.confirmPassword = ''
}

// 切换模式
const toggleMode = () => {
  isLogin.value = !isLogin.value
}

// 关闭模态框
const close = () => {
  emit('close')
}

// 处理表单提交
const handleSubmit = async () => {
  // 验证表单
  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    if (isLogin.value) {
      // 登录
      const loginData: LoginRequest = {
        username: form.username,
        password: form.password
      }
      
      const response = await login(loginData)
      
      if (response && response.success === true && response.data) {
        // 设置认证状态到stores
        const userData = (response.data as any).user || {
          id: (response.data as any).user_id,
          username: (response.data as any).username,
          email: (response.data as any).email,
          is_active: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        }
        const token = (response.data as any).access_token || (response.data as any).token
        
        authStore.setAuthState(userData, token)
        
        showSuccess('登录成功！')
        emit('success')
        close()
      } else {
        showError(response?.message || '登录失败，请检查用户名或密码')
      }
    } else {
      // 注册
      const registerData: RegisterRequest = {
        username: form.username,
        email: form.email,
        password: form.password
      }
      
      const response = await register(registerData)
      
      if (response && response.success === true) {
        showSuccess('注册成功！请使用新账号登录。')
        // 注册成功后自动切换到登录模式
        isLogin.value = true
      } else {
        showError(response?.message || '注册失败，请重试')
      }
    }
  } catch (err: unknown) {
    showError('操作失败，请重试')
  } finally {
    isLoading.value = false
  }
}

// 验证表单
const validateForm = (): boolean => {
  if (!form.username.trim()) {
    showError('请输入用户名')
    return false
  }

  if (!isLogin.value) {
    if (!form.email.trim()) {
      showError('请输入邮箱')
      return false
    }

    if (!form.email.includes('@')) {
      showError('请输入有效的邮箱地址')
      return false
    }

    if (form.password.length < 6) {
      showError('密码长度至少6位')
      return false
    }

    if (form.password !== form.confirmPassword) {
      showError('两次输入的密码不一致')
      return false
    }
  }

  if (!form.password.trim()) {
    showError('请输入密码')
    return false
  }

  return true
}
</script>
