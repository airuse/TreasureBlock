# 项目结构说明

## 📁 目录结构

```
vue/
├── src/
│   ├── api/                    # API服务层
│   │   └── index.ts           # API接口定义和模拟数据
│   ├── assets/                 # 静态资源
│   │   └── main.css           # 全局样式
│   ├── components/             # 组件目录
│   │   ├── layout/            # 布局组件
│   │   │   └── MainLayout.vue # 主布局组件
│   │   └── ui/                # UI组件（待扩展）
│   ├── router/                 # 路由配置
│   │   └── index.ts           # 路由定义
│   ├── stores/                 # 状态管理（Pinia）
│   ├── types/                  # TypeScript类型定义
│   │   └── index.ts           # 全局类型定义
│   ├── utils/                  # 工具函数
│   │   └── formatters.ts      # 格式化工具函数
│   ├── views/                  # 页面组件
│   │   ├── pages/             # 页面级组件
│   │   │   ├── HomeView.vue   # 首页
│   │   │   ├── BlocksView.vue # 区块列表页
│   │   │   ├── TransactionsView.vue # 交易列表页
│   │   │   ├── AddressesView.vue    # 地址列表页
│   │   │   ├── StatisticsView.vue   # 统计页面
│   │   │   └── SettingsView.vue     # 设置页面
│   │   └── detail/            # 详情页面（待扩展）
│   ├── App.vue                 # 根组件
│   └── main.ts                 # 应用入口
├── public/                     # 公共静态资源
├── dist/                       # 构建输出目录
├── node_modules/               # 依赖包
├── .env.local                  # 本地环境变量
├── index.html                  # HTML模板
├── package.json                # 项目配置
├── tailwind.config.js          # TailwindCSS配置
├── postcss.config.js           # PostCSS配置
├── tsconfig.json               # TypeScript配置
├── vite.config.ts              # Vite配置
├── README.md                   # 项目说明
└── PROJECT_STRUCTURE.md        # 项目结构说明（本文件）
```

## 🏗️ 架构设计

### 分层架构
- **API层** (`src/api/`): 负责与后端通信，包含接口定义和模拟数据
- **组件层** (`src/components/`): 可复用的UI组件
- **页面层** (`src/views/`): 页面级组件，组合各种组件
- **工具层** (`src/utils/`): 通用工具函数
- **类型层** (`src/types/`): TypeScript类型定义

### 组件分类
- **Layout组件**: 布局相关，如导航栏、侧边栏
- **UI组件**: 基础UI组件，如按钮、表格、卡片
- **页面组件**: 完整的页面视图
- **详情组件**: 详情页面（待扩展）

## 📝 文件命名规范

### 组件文件
- 使用PascalCase命名
- 页面组件以`View`结尾
- 布局组件以`Layout`结尾
- 基础组件使用描述性名称

### 工具文件
- 使用camelCase命名
- 按功能分类，如`formatters.ts`、`validators.ts`

### 类型文件
- 使用camelCase命名
- 全局类型定义在`types/index.ts`

## 🔧 开发规范

### 导入顺序
1. Vue相关导入
2. 第三方库导入
3. 本地组件导入
4. 工具函数导入
5. 类型导入

### 组件结构
```vue
<template>
  <!-- 模板内容 -->
</template>

<script setup lang="ts">
// 导入
import { ref, computed } from 'vue'
import type { ComponentType } from '@/types'

// 类型定义
interface Props {
  // 属性定义
}

// 响应式数据
const data = ref()

// 计算属性
const computedValue = computed(() => {
  // 计算逻辑
})

// 方法
const handleClick = () => {
  // 处理逻辑
}
</script>

<style scoped>
/* 样式 */
</style>
```

## 🚀 扩展指南

### 添加新页面
1. 在`src/views/pages/`创建页面组件
2. 在`src/router/index.ts`添加路由
3. 在`src/types/index.ts`添加相关类型
4. 在`src/api/index.ts`添加API接口

### 添加新组件
1. 在`src/components/ui/`创建UI组件
2. 在`src/components/layout/`创建布局组件
3. 导出组件并在需要的地方导入使用

### 添加新工具函数
1. 在`src/utils/`创建工具文件
2. 导出函数并在需要的地方导入使用

## 📊 数据流

```
API层 → 页面组件 → UI组件 → 用户交互
  ↑                                    ↓
类型定义 ← 工具函数 ← 格式化 ← 数据处理
```

## 🎨 样式规范

### CSS类命名
- 使用TailwindCSS工具类
- 自定义类使用BEM命名法
- 组件样式使用scoped

### 主题色彩
- 主色：蓝色系 (`blue-*`)
- 成功：绿色系 (`green-*`)
- 警告：黄色系 (`yellow-*`)
- 错误：红色系 (`red-*`)

## 🔍 调试指南

### 开发工具
- Vue DevTools: 组件调试
- 浏览器开发者工具: 网络、控制台
- TypeScript: 类型检查

### 常见问题
1. **路由问题**: 检查`router/index.ts`配置
2. **类型错误**: 检查`types/index.ts`定义
3. **样式问题**: 检查TailwindCSS配置
4. **API问题**: 检查`api/index.ts`接口

## 📚 相关文档

- [Vue 3 官方文档](https://vuejs.org/)
- [Vue Router 文档](https://router.vuejs.org/)
- [TailwindCSS 文档](https://tailwindcss.com/)
- [TypeScript 文档](https://www.typescriptlang.org/) 