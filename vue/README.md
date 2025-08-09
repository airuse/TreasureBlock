# 区块链浏览器前端

这是一个基于Vue 3的区块链浏览器前端项目，采用现代化的技术栈和简洁的设计风格。

## 技术栈

- **Vue 3** - 渐进式JavaScript框架
- **TypeScript** - 类型安全的JavaScript超集
- **Vue Router** - 官方路由管理器
- **Pinia** - 状态管理库
- **TailwindCSS** - 实用优先的CSS框架
- **Heroicons** - 精美的SVG图标库
- **Vite** - 下一代前端构建工具

## 项目结构

```
src/
├── assets/          # 静态资源
├── components/      # 公共组件
├── layouts/         # 布局组件
├── router/          # 路由配置
├── stores/          # 状态管理
├── views/           # 页面组件
└── main.ts          # 应用入口
```

## 功能特性

### 🏠 首页
- 区块链概览统计
- 最新区块列表
- 最新交易列表
- 实时数据更新

### 📦 区块管理
- 区块列表浏览
- 区块详情查看
- 搜索和筛选功能
- 分页显示

### 💰 交易管理
- 交易列表浏览
- 交易详情查看
- 交易状态显示
- 高级搜索功能

### 👥 地址管理
- 地址列表浏览
- 地址详情查看
- 地址类型分类
- 标签管理

### 📊 统计分析
- 网络统计指标
- 趋势图表显示
- 实时数据监控
- 历史数据对比

### ⚙️ 系统设置
- 显示设置
- 网络配置
- 通知管理
- 安全设置

## 设计特点

### 🎨 简洁实用
- 采用简洁的设计语言
- 注重用户体验
- 信息层次清晰
- 操作流程优化

### 👨‍💻 程序员友好
- 技术数据展示
- 专业术语使用
- 调试信息支持
- 开发者工具集成

### 📱 响应式设计
- 支持多设备访问
- 自适应布局
- 移动端优化
- 触摸友好

## 开发指南

### 环境要求

- Node.js 18+
- npm 9+

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run dev
```

### 构建生产版本

```bash
npm run build
```

### 代码检查

```bash
npm run lint
```

### 代码格式化

```bash
npm run format
```

## 项目配置

### 环境变量

创建 `.env.local` 文件：

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_NETWORK=mainnet
VITE_APP_TITLE=区块链浏览器
```

### TailwindCSS配置

项目使用自定义的区块链主题色彩：

```javascript
// tailwind.config.js
theme: {
  extend: {
    colors: {
      'blockchain': {
        50: '#f0f9ff',
        100: '#e0f2fe',
        200: '#bae6fd',
        300: '#7dd3fc',
        400: '#38bdf8',
        500: '#0ea5e9',
        600: '#0284c7',
        700: '#0369a1',
        800: '#075985',
        900: '#0c4a6e',
      }
    }
  }
}
```

## 组件说明

### MainLayout
主布局组件，包含：
- 顶部导航栏
- 左侧菜单栏
- 右侧内容区域
- 登录模态框

### 页面组件
- `HomeView` - 首页概览
- `BlocksView` - 区块列表
- `TransactionsView` - 交易列表
- `AddressesView` - 地址列表
- `StatisticsView` - 统计分析
- `SettingsView` - 系统设置

## 数据格式

### 区块数据
```typescript
interface Block {
  height: number
  timestamp: number
  transactions: number
  size: number
  gasUsed: number
  gasLimit: number
  miner: string
  reward: number
}
```

### 交易数据
```typescript
interface Transaction {
  hash: string
  blockHeight: number
  timestamp: number
  from: string
  to: string
  amount: number
  gasUsed: number
  gasPrice: number
  status: 'success' | 'failed' | 'pending'
}
```

## 部署说明

### 构建部署

1. 构建生产版本：
```bash
npm run build
```

2. 部署到服务器：
```bash
# 将 dist 目录部署到 Web 服务器
```

### Docker部署

```dockerfile
FROM node:18-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License

## 联系方式

如有问题或建议，请通过以下方式联系：

- 项目Issues
- 邮箱：dev@example.com

---

**注意**：这是一个前端演示项目，目前使用模拟数据。在实际部署时需要连接后端API服务。
