# Toast 通知系统

## 概述

新的Toast通知系统解决了原有内联提示框的问题：
- ❌ **原有问题**：提示框会撑宽撑高页面，导致布局抖动
- ✅ **新方案**：固定位置的Toast通知，不影响页面布局

## 特性

- 🎯 **固定位置** - 右上角固定位置，不占用页面空间
- 🎨 **多种类型** - 支持成功、错误、警告、信息四种类型
- ⏱️ **自动消失** - 默认5秒后自动消失，可自定义时长
- 🖱️ **手动关闭** - 点击关闭按钮可立即关闭
- 📱 **响应式** - 适配不同屏幕尺寸
- 🎭 **平滑动画** - 进入和退出都有平滑的动画效果

## 使用方法

### 1. 基本用法

```typescript
import { showSuccess, showError, showWarning, showInfo } from '@/composables/useToast'

// 显示成功提示
showSuccess('操作成功！')

// 显示错误提示
showError('操作失败，请重试')

// 显示警告提示
showWarning('请注意这个操作')

// 显示信息提示
showInfo('这是一条信息')
```

### 2. 自定义显示时长

```typescript
// 显示3秒
showSuccess('操作成功！', 3000)

// 永久显示（不自动消失）
showInfo('重要信息', 0)
```

### 3. 高级用法

```typescript
import { addToast } from '@/composables/useToast'

// 自定义Toast
addToast({
  type: 'success',
  message: '自定义消息',
  duration: 8000
})
```

## 类型定义

```typescript
interface Toast {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  message: string
  duration?: number // 毫秒，0表示不自动消失
}
```

## 样式说明

- **成功 (Success)** - 绿色主题，✓ 图标
- **错误 (Error)** - 红色主题，✗ 图标  
- **警告 (Warning)** - 黄色主题，⚠ 图标
- **信息 (Info)** - 蓝色主题，ℹ 图标

## 迁移指南

### 从内联提示框迁移

**之前：**
```vue
<template>
  <div v-if="error" class="text-red-600 text-sm bg-red-50 p-3 rounded-md">
    {{ error }}
  </div>
  <div v-if="success" class="text-green-600 text-sm bg-green-50 p-3 rounded-md">
    {{ success }}
  </div>
</template>

<script>
const error = ref('')
const success = ref('')
</script>
```

**现在：**
```vue
<template>
  <!-- 移除所有内联提示框 -->
</template>

<script>
import { showSuccess, showError } from '@/composables/useToast'

// 移除 error 和 success 变量
// 直接使用 Toast 方法
showSuccess('操作成功！')
showError('操作失败！')
</script>
```

## 优势

1. **布局稳定** - 不会影响页面布局，避免抖动
2. **用户体验** - 统一的提示风格，更专业
3. **代码简洁** - 减少模板代码，逻辑更清晰
4. **性能更好** - 不需要频繁的DOM更新
5. **易于维护** - 集中管理所有提示逻辑

## 注意事项

- Toast组件已全局注册在App.vue中
- 使用Teleport渲染到body，确保层级正确
- 支持多个Toast同时显示
- 自动管理内存，组件卸载时清理定时器
