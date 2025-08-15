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
            <div class="relative">
              <button
                @click="togglePermissionDropdown"
                type="button"
                class="w-full px-3 py-2 text-left border border-gray-300 rounded-md bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 hover:border-gray-400 transition-colors"
              >
                <span v-if="newKeyForm.permissions.length === 0" class="text-gray-500">
                  请选择权限范围
                </span>
                <span v-else class="text-gray-900">
                  已选择 {{ newKeyForm.permissions.length }} 项权限
                </span>
                <svg 
                  class="absolute right-3 top-2.5 w-5 h-5 text-gray-400 transition-transform duration-200"
                  :class="{ 'rotate-180': showPermissionDropdown }"
                  fill="none" 
                  stroke="currentColor" 
                  viewBox="0 0 24 24"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              
              <!-- 权限下拉选项 -->
              <div v-if="showPermissionDropdown" class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-auto">
                <div class="p-2">
                  <div v-for="perm in permissionTypes" :key="perm.value" class="flex items-center p-2 hover:bg-gray-50 rounded cursor-pointer">
                    <input
                      v-model="newKeyForm.permissions"
                      type="checkbox"
                      :value="perm.value"
                      :id="'perm-' + perm.value"
                      class="mr-3 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                    />
                    <label :for="'perm-' + perm.value" class="text-sm text-gray-700 cursor-pointer flex-1">
                      {{ perm.description }}
                    </label>
                  </div>
                  
                  <!-- 如果没有权限选项，显示提示 -->
                  <div v-if="permissionTypes.length === 0" class="p-3 text-center text-gray-500 text-sm">
                    暂无可用权限
                  </div>
                </div>
              </div>
            </div>
            
            <!-- 已选择的权限标签 -->
            <div v-if="newKeyForm.permissions.length > 0" class="mt-2 flex flex-wrap gap-2">
              <span 
                v-for="perm in newKeyForm.permissions" 
                :key="perm"
                class="inline-flex items-center px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded-full"
              >
                {{ getPermissionDescription(perm) }}
                <button
                  @click="removePermission(perm)"
                  class="ml-1 text-blue-600 hover:text-blue-800"
                  type="button"
                >
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </span>
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

    <!-- 创建成功后的密钥显示模态框 -->
    <div v-if="showCreatedKeyModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-60">
      <div class="bg-white rounded-lg shadow-xl max-w-3xl w-full mx-4">
        <div class="flex items-center justify-between p-6 border-b border-gray-200">
          <h3 class="text-lg font-semibold text-green-600">🎉 API密钥创建成功！</h3>
          <button
            @click="closeCreatedKeyModal"
            class="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="p-6 space-y-6">
          <!-- 重要提醒 -->
          <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
            <div class="flex items-start">
              <svg class="w-5 h-5 text-yellow-400 mt-0.5 mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
              </svg>
              <div>
                <h4 class="text-sm font-medium text-yellow-800">重要提醒</h4>
                <p class="text-sm text-yellow-700 mt-1">
                  请立即保存您的Secret Key！关闭此窗口后，Secret Key将不再以明文形式显示。
                  为了安全起见，系统将自动加密存储Secret Key。
                </p>
              </div>
            </div>
          </div>

          <!-- API Key -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">API Key</label>
            <div class="flex items-center space-x-2">
              <code class="flex-1 p-3 bg-gray-50 border border-gray-300 rounded text-sm font-mono break-all">
                {{ createdKeyData.api_key }}
              </code>
              <button
                @click="copyToClipboard(createdKeyData.api_key)"
                class="px-3 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                title="复制API Key"
              >
                复制
              </button>
            </div>
          </div>

          <!-- Secret Key -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">Secret Key</label>
            <div class="flex items-center space-x-2">
              <code class="flex-1 p-3 bg-gray-50 border border-gray-300 rounded text-sm font-mono break-all">
                {{ createdKeyData.secret_key }}
              </code>
              <button
                @click="copyToClipboard(createdKeyData.secret_key)"
                class="px-3 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors"
                title="复制Secret Key"
              >
                复制
              </button>
            </div>
          </div>

          <!-- 使用说明 -->
          <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div class="flex items-start">
              <svg class="w-5 h-5 text-blue-400 mt-0.5 mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
              </svg>
              <div>
                <h4 class="text-sm font-medium text-blue-800">使用说明</h4>
                <p class="text-sm text-blue-700 mt-1">
                  请将这两个密钥安全地保存在您的地方。API Key用于标识您的身份，
                  Secret Key用于验证您的身份。请勿将Secret Key分享给任何人。
                </p>
              </div>
            </div>
          </div>
        </div>

        <div class="flex justify-end p-6 border-t border-gray-200">
          <button
            @click="closeCreatedKeyModal"
            class="px-6 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition-colors"
          >
            我已保存，关闭窗口
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import type { APIKey, PermissionConfig } from '@/types/auth'
import { showSuccess, showError } from '@/composables/useToast'

// 定义配置对象类型
interface PermissionType {
  value: string
  description: string
}

const props = defineProps<{
  isVisible: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const authStore = useAuthStore()

// 响应式数据
const isLoading = ref(false)
const apiKeys = ref<APIKey[]>([])

// 创建成功后的密钥数据
const showCreatedKeyModal = ref(false)
const createdKeyData = reactive({
  api_key: '',
  secret_key: ''
})

// 表单数据
const newKeyForm = reactive({
  name: '',
  permissions: [] as string[], // 初始为空，由watch动态设置
  expiresAt: getDefaultExpiryDate()
})

// 权限类型列表
const permissionTypes = ref<Array<{ value: string; description: string }>>([])

// 权限下拉框显示状态
const showPermissionDropdown = ref(false)

// 获取默认过期时间（一年后的今天）
function getDefaultExpiryDate(): string {
  const today = new Date()
  const oneYearLater = new Date(today.getFullYear() + 1, today.getMonth(), today.getDate())
  return oneYearLater.toISOString().split('T')[0]
}

// 监听模态框显示状态
watch(() => props.isVisible, (visible) => {
  if (visible) {
    loadAPIKeys()
    loadPermissionTypes()
    // 清空表单
    newKeyForm.name = ''
    newKeyForm.permissions = [] // 重置为空，让watch自动设置默认值
    newKeyForm.expiresAt = getDefaultExpiryDate()
  }
})

// 监听权限类型加载完成，设置默认权限
watch(permissionTypes, (newPermissions) => {
  if (newPermissions.length > 0 && newKeyForm.permissions.length === 0) {
    // 设置默认权限（选择前两个权限）
    newKeyForm.permissions = newPermissions.slice(0, 2).map(p => p.value)
  }
}, { immediate: true })

// 关闭模态框
const close = () => {
  emit('close')
}

// 点击外部关闭权限下拉框
const handleClickOutside = (event: Event) => {
  const target = event.target as Element
  if (!target.closest('.relative')) {
    showPermissionDropdown.value = false
  }
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
  } catch (err: unknown) {
    showError('加载API密钥失败: ' + (err instanceof Error ? err.message : '未知错误'))
  }
}

// 加载权限类型列表
const loadPermissionTypes = async () => {
  try {
    const response = await authStore.getPermissionTypes()
    if (response.success) {
      // 将PermissionConfig对象转换为前端需要的格式
      permissionTypes.value = (response.data || []).map((config: any) => ({
        value: config.config_value, // 使用config_value作为权限值
        description: config.config_name // 使用config_name作为描述
      }))
    }
  } catch (err: unknown) {
    console.error('Failed to load permission types:', err)
  }
}

// 切换权限下拉框显示状态
const togglePermissionDropdown = () => {
  showPermissionDropdown.value = !showPermissionDropdown.value
}

// 创建API密钥
const createAPIKey = async () => {
  try {
    if (!newKeyForm.name.trim()) {
      showError('请输入密钥名称')
      return
    }

    isLoading.value = true
    
    const response = await authStore.createNewAPIKey({
      name: newKeyForm.name.trim(),
      permissions: newKeyForm.permissions,
      expires_at: newKeyForm.expiresAt || undefined
    })
    
    if (response && response.code === 200) {
      // 保存创建的密钥数据用于显示
      createdKeyData.api_key = response.data?.api_key || ''
      createdKeyData.secret_key = response.data?.secret_key || ''
      
      // 显示创建成功模态框
      showCreatedKeyModal.value = true
      
      // 重新加载API密钥列表
      await loadAPIKeys()
      
      // 清空表单
      newKeyForm.name = ''
      newKeyForm.permissions = [] // 重置为空，让watch自动设置默认值
      newKeyForm.expiresAt = getDefaultExpiryDate()
      
      // 关闭权限下拉框
      showPermissionDropdown.value = false
      showSuccess('API密钥创建成功！')
    }
    
  } catch (err: unknown) {
    showError('创建失败: ' + (err instanceof Error ? err.message : '未知错误'))
  } finally {
    isLoading.value = false
  }
}

// 切换密钥状态
const toggleKeyStatus = async (key: APIKey) => {
  try {
    // 调用真实API更新密钥状态
    const response = await authStore.updateExistingAPIKey(key.id, {
      is_active: !key.is_active
    })
    
    if (response && response.code === 200) {
      showSuccess(`密钥已${key.is_active ? '禁用' : '启用'}`)
      await loadAPIKeys()
    }
    
  } catch (err: unknown) {
    showError('操作失败: ' + (err instanceof Error ? err.message : '未知错误'))
  }
}

// 删除密钥
const deleteKey = async (key: APIKey) => {
  if (!confirm(`确定要删除密钥"${key.name}"吗？此操作不可恢复。`)) {
    return
  }
  
  try {
    // 调用真实API删除密钥
    const response = await authStore.deleteExistingAPIKey(key.id)
    
    if (response && response.code === 200) {
      showSuccess('密钥已删除')
      await loadAPIKeys()
    }
    
  } catch (err: unknown) {
    showError('删除失败: ' + (err instanceof Error ? err.message : '未知错误'))
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
    showSuccess('已复制到剪贴板！')
  }).catch(err => {
    showError('复制失败: ' + err)
  })
}

// 获取状态样式
const getStatusClass = (isActive: boolean) => {
  return isActive 
    ? 'bg-green-100 text-green-800' 
    : 'bg-red-100 text-red-800'
}

// 获取权限描述
const getPermissionDescription = (value: string) => {
  const perm = permissionTypes.value.find(p => p.value === value);
  return perm ? perm.description : value;
};

// 移除权限
const removePermission = (value: string) => {
  newKeyForm.permissions = newKeyForm.permissions.filter(perm => perm !== value);
};

// 关闭创建成功模态框
const closeCreatedKeyModal = () => {
  showCreatedKeyModal.value = false
  createdKeyData.api_key = ''
  createdKeyData.secret_key = ''
}

// 组件挂载时加载数据
onMounted(() => {
  if (props.isVisible) {
    loadAPIKeys()
    loadPermissionTypes()
  }
  document.addEventListener('click', handleClickOutside)
})

// 组件卸载时移除监听
onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>
