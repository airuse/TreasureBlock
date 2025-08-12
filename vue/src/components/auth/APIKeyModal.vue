<template>
  <div v-if="isVisible" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-white rounded-lg shadow-xl max-w-4xl w-full mx-4 max-h-[90vh] overflow-y-auto">
      <!-- 头部 -->
      <div class="flex items-center justify-between p-6 border-b border-gray-200">
        <h2 class="text-xl font-semibold text-gray-900">API密钥管理</h2>
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
        <!-- 创建新API密钥 -->
        <div class="bg-gray-50 p-4 rounded-lg">
          <h3 class="text-lg font-medium text-gray-900 mb-4">创建新API密钥</h3>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">密钥名称</label>
              <input
                v-model="newKeyForm.name"
                type="text"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="例如：扫块客户端1"
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">过期时间</label>
              <input
                v-model="newKeyForm.expiresAt"
                type="date"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>
          
          <div class="mt-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">权限范围</label>
            <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
              <label v-for="perm in permissionTypes" :key="perm.value" class="flex items-center">
                <input
                  v-model="newKeyForm.permissions"
                  type="checkbox"
                  :value="perm.value"
                  class="mr-2"
                />
                <span class="text-sm">{{ perm.description }}</span>
              </label>
            </div>
          </div>
          
          <div class="mt-4">
            <button
              @click="createAPIKey"
              :disabled="!newKeyForm.name || newKeyForm.permissions.length === 0 || isLoading"
              class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {{ isLoading ? '创建中...' : '创建API密钥' }}
            </button>
          </div>
        </div>

        <!-- API密钥列表 -->
        <div>
          <h3 class="text-lg font-medium text-gray-900 mb-4">我的API密钥</h3>
          
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">名称</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">API Key</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Secret Key</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">权限</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200">
                <tr v-for="key in apiKeys" :key="key.id" class="hover:bg-gray-50">
                  <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    {{ key.name }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    <div class="flex items-center space-x-2">
                      <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono">{{ key.api_key }}</code>
                      <button
                        @click="copyToClipboard(key.api_key)"
                        class="text-gray-500 hover:text-gray-700 transition-colors"
                        title="复制API Key"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                      </button>
                    </div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    <div class="flex items-center space-x-2">
                      <code class="bg-gray-100 px-2 py-1 rounded text-xs font-mono">{{ key.secret_key || '未显示' }}</code>
                      <button
                        v-if="key.secret_key"
                        @click="copyToClipboard(key.secret_key)"
                        class="text-gray-500 hover:text-gray-700 transition-colors"
                        title="复制Secret Key"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                      </button>
                    </div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    <div class="flex flex-wrap gap-1">
                      <span v-for="perm in parsePermissions(key.permissions)" :key="perm" 
                            class="inline-flex px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded">
                        {{ perm }}
                      </span>
                    </div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <span :class="getStatusClass(key.is_active)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                      {{ key.is_active ? '活跃' : '已禁用' }}
                    </span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
                    <button
                      @click="toggleKeyStatus(key)"
                      :class="key.is_active ? 'text-red-600 hover:text-red-900' : 'text-green-600 hover:text-green-900'"
                    >
                      {{ key.is_active ? '禁用' : '启用' }}
                    </button>
                    <button
                      @click="deleteKey(key)"
                      class="text-red-600 hover:text-red-900"
                    >
                      删除
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          
          <div v-if="apiKeys.length === 0" class="text-center py-8 text-gray-500">
            暂无API密钥，请创建一个
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
      <div class="flex justify-end p-6 border-t border-gray-200">
        <button
          @click="close"
          class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
        >
          关闭
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import type { APIKey } from '@/types/auth'

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
const apiKeys = ref<APIKey[]>([])

// 表单数据
const newKeyForm = reactive({
  name: '',
  permissions: ['blocks:write', 'transactions:write'] as string[],
  expiresAt: ''
})

// 权限类型列表
const permissionTypes = ref<Array<{ value: string; description: string }>>([])

// 监听模态框显示状态
watch(() => props.isVisible, (visible) => {
  if (visible) {
    loadAPIKeys()
    loadPermissionTypes()
    // 清空表单
    newKeyForm.name = ''
    newKeyForm.permissions = ['blocks:write', 'transactions:write']
    newKeyForm.expiresAt = ''
    // 清空提示信息
    error.value = ''
    success.value = ''
  }
})

// 关闭模态框
const close = () => {
  emit('close')
}

// 加载API密钥列表
const loadAPIKeys = async () => {
  try {
    if (authStore.apiKeys.length > 0) {
      apiKeys.value = authStore.apiKeys
    } else {
      await authStore.fetchAPIKeys()
      apiKeys.value = authStore.apiKeys
    }
  } catch (err: any) {
    error.value = '加载API密钥失败: ' + err.message
  }
}

// 加载权限类型列表
const loadPermissionTypes = async () => {
  try {
    const response = await authStore.getPermissionTypes()
    if (response.success) {
      // 将BaseConfig对象转换为前端需要的格式
      permissionTypes.value = (response.data || []).map((config: any) => ({
        value: config.config_value,
        description: config.description
      }))
    }
  } catch (err: any) {
    console.error('Failed to load permission types:', err)
  }
}

// 创建API密钥
const createAPIKey = async () => {
  try {
    if (!newKeyForm.name.trim()) {
      error.value = '请输入密钥名称'
      return
    }

    isLoading.value = true
    error.value = ''
    
    const response = await authStore.createAPIKey({
      name: newKeyForm.name.trim(),
      permissions: newKeyForm.permissions,
      expires_at: newKeyForm.expiresAt || undefined
    })
    
    if (response.success) {
      success.value = 'API密钥创建成功！'
      await loadAPIKeys()
      
      // 延迟清空成功提示
      setTimeout(() => {
        success.value = ''
      }, 3000)
    }
    
  } catch (err: any) {
    error.value = '创建失败: ' + err.message
  } finally {
    isLoading.value = false
  }
}

// 切换密钥状态
const toggleKeyStatus = async (key: APIKey) => {
  try {
    // 调用真实API更新密钥状态
    const response = await authStore.updateAPIKey(key.id, {
      is_active: !key.is_active
    })
    
    if (response.success) {
      success.value = `密钥已${key.is_active ? '禁用' : '启用'}`
      await loadAPIKeys()
      
      setTimeout(() => {
        success.value = ''
      }, 2000)
    }
    
  } catch (err: any) {
    error.value = '操作失败: ' + err.message
  }
}

// 删除密钥
const deleteKey = async (key: APIKey) => {
  if (!confirm(`确定要删除密钥"${key.name}"吗？此操作不可恢复。`)) {
    return
  }
  
  try {
    // 调用真实API删除密钥
    const response = await authStore.deleteAPIKey(key.id)
    
    if (response.success) {
      success.value = '密钥已删除'
      await loadAPIKeys()
      
      setTimeout(() => {
        success.value = ''
      }, 2000)
    }
    
  } catch (err: any) {
    error.value = '删除失败: ' + err.message
  }
}

// 解析权限字符串为数组
const parsePermissions = (permissions: string[]) => {
  if (Array.isArray(permissions)) {
    return permissions.map(perm => {
      if (perm.includes(':')) {
        const [resource, action] = perm.split(':')
        return `${resource}:${action}`
      }
      return perm
    })
  }
  return []
}

// 复制到剪贴板
const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text).then(() => {
    success.value = '已复制到剪贴板！'
    setTimeout(() => {
      success.value = ''
    }, 2000)
  }).catch(err => {
    error.value = '复制失败: ' + err
    setTimeout(() => {
      error.value = ''
    }, 2000)
  })
}

// 获取状态样式
const getStatusClass = (isActive: boolean) => {
  return isActive 
    ? 'bg-green-100 text-green-800' 
    : 'bg-red-100 text-red-800'
}

// 组件挂载时加载数据
onMounted(() => {
  if (props.isVisible) {
    loadAPIKeys()
    loadPermissionTypes()
  }
})
</script>
