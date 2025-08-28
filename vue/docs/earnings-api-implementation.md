# 收益模块API实现说明

## 📁 文件结构

按照API文件结构说明规范，收益模块包含以下文件：

```
vue/src/
├── api/
│   ├── earnings/
│   │   └── index.ts          # API函数实现 + 模块相关类型定义
│   └── mock/
│       └── earnings.ts        # Mock数据处理
├── types/
│   └── earnings.ts            # 业务实体类型定义
└── ApiDatas/
    └── earnings/
        └── earnings-v1.json   # API文档数据
```

## 🚀 功能特性

### 1. 用户余额管理
- 获取用户总余额、可用余额、冻结余额
- 获取总收益、今日收益、本月收益
- 实时更新余额信息

### 2. 收益记录管理
- 分页获取收益记录列表
- 支持按时间周期筛选（7天、30天、90天）
- 支持按状态筛选（待确认、已确认、失败）
- 获取收益记录详情

### 3. 收益统计分析
- 总收益统计
- 今日/本月收益统计
- 区块数量统计（总数、已确认、待确认、失败）
- 平均每块收益

### 4. T币转账功能
- 用户间T币转账
- 转账状态跟踪
- 转账记录管理

## 🔧 使用方法

### 在Vue组件中使用

```typescript
import { 
  getUserBalance, 
  getUserEarningsRecords, 
  getUserEarningsStats 
} from '@/api/earnings'

// 获取用户余额
const balance = await getUserBalance()

// 获取收益记录列表
const records = await getUserEarningsRecords({
  page: 1,
  page_size: 10,
  period: 30
})

// 获取收益统计
const stats = await getUserEarningsStats()
```

### Mock数据切换

在 `vite.config.ts` 中配置：

```typescript
define: {
  __USE_MOCK__: false, // false: 使用真实API, true: 使用Mock数据
}
```

## 📊 数据流程

1. **页面加载** → 调用 `loadUserData()` 和 `loadEarnings()`
2. **用户余额** → 调用 `getUserBalance()` API
3. **收益统计** → 调用 `getUserEarningsStats()` API
4. **收益记录** → 调用 `getUserEarningsRecords()` API
5. **数据展示** → 更新页面响应式数据

## 🎯 已实现的页面

- `EarningsView.vue` - 收益概览页面
  - 显示用户余额信息
  - 显示收益统计数据
  - 显示收益记录列表
  - 支持分页和筛选

## 🔄 后续扩展

1. **收益图表** - 集成Chart.js或ECharts显示收益趋势
2. **转账功能** - 实现T币转账界面
3. **收益详情** - 点击记录查看详细信息
4. **实时更新** - 集成WebSocket获取实时收益数据
5. **导出功能** - 支持导出收益记录为CSV/Excel

## 📝 注意事项

1. 所有API调用都包含错误处理
2. 使用TypeScript确保类型安全
3. 支持Mock数据和真实API的无缝切换
4. 遵循RESTful API设计规范
5. 包含完整的请求参数验证
