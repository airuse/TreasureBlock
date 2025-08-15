# API开发规范与结构说明 (TypeScript版本)

## 📁 标准文件结构

```
vue/src/api/
├── [模块名]/
│   └── index.ts          # API函数实现 + 模块相关类型定义
├── mock/
│   └── [模块名].ts       # Mock数据处理
├── types.ts              # 通用API类型定义
├── index.ts              # API统一导出入口
└── request.ts            # axios请求配置

vue/src/types/           # 只保留业务实体类型
├── index.ts
├── block.ts             # 业务实体类型
├── transaction.ts       # 业务实体类型
├── address.ts           # 业务实体类型
├── stats.ts             # 业务实体类型
├── auth.ts              # 业务实体类型
└── user.ts              # 业务实体类型

vue/ApiDatas/                 # Mock数据源文件夹
└── [模块名]/
    └── [模块名]-[版本号].json  # API文档数据
```

## 🛠 代码生成步骤

### 第一步：创建业务实体类型文件 (src/types/[模块名].ts)
```typescript
// 基础业务实体类型
export interface [数据类型名] {
  id: string | number
  name: string
  // 其他业务字段...
}

// 列表项类型
export interface [数据类型名]ListItem {
  id: string | number
  name: string
  // 列表展示需要的字段...
}

// 详情类型
export interface [数据类型名]Detail extends [数据类型名] {
  description?: string
  createdAt?: string
  updatedAt?: string
}
```

### 第二步：创建Mock处理文件 (src/api/mock/[模块名].ts)
```typescript
import apiData from '../../../ApiDatas/[模块名]/[模块名]-[版本号].json'

/**
 * 模拟[接口功能]接口
 */
export const handleMock[接口名] = (data: any): Promise<any> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const response = apiData.paths['/[模块名]/[接口名]'].responses['200'].content['application/json'].example
      resolve({
        ...response,
        timestamp: Date.now()
      })
    }, 300)
  })
}
```

### 第三步：创建API实现文件 (src/api/[模块名]/index.ts)
```typescript
import request from '../request'
import { handleMock[接口名] } from '../mock/[模块名]'
import type { [数据类型名] } from '@/types'
import type { ApiResponse, PaginatedResponse, PaginationRequest, SortRequest, SearchRequest } from '../types'

// ==================== API相关类型定义 ====================

// 请求参数类型 - 继承通用类型
interface [接口名]Request extends PaginationRequest, SortRequest {
  // 模块特有参数
  param1?: string
  param2?: number
}

// ==================== API函数实现 ====================

/**
 * [接口功能描述]
 */
export function [接口名](data: [接口名]Request): Promise<ApiResponse<[数据类型名]>> {
  if (__USE_MOCK__) {
    console.log('🔧 使用Mock数据 - [接口名]')
    return handleMock[接口名](data)
  }
  
  console.log('🌐 使用真实API - [接口名]')
  return request({
    url: '/[模块名]/[接口名]',
    method: '[http方法]',
    data  // POST请求
    // params  // GET请求
  })
}
```

### 第四步：更新导出文件

在 `src/api/index.ts` 中添加：
```typescript
import * as [模块名] from './[模块名]'

export {
  // ... 其他模块
  [模块名]
}

export default {
  // ... 其他模块
  [模块名]
}
```

在 `src/types/index.ts` 中添加：
```typescript
export * from './[模块名]'
```

## 🎨 命名规范

- **模块名**：小写英文，如 blocks、transactions、addresses、stats、auth、user
- **接口名**：驼峰命名，动词开头，如 getBlocks、getBlock、searchBlocks
- **类型名**：业务实体类型 `[数据类型名]`，请求类型 `[接口名]Request`
- **Mock函数**：`handleMock` + 接口名
- **通用类型**：继承 `PaginationRequest`、`SortRequest`、`SearchRequest` 等 