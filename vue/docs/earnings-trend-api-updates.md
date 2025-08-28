# 收益趋势API更新说明

## 🎯 更新概述

根据用户反馈，为收益趋势图表创建了专门的API接口，不再依赖前端过滤现有的收益记录列表。这样设计更加合理和高效。

## 🆕 新增接口

### 1. 后端新增接口

#### DTO类型定义 (`server/internal/dto/earnings_dto.go`)
```go
// EarningsTrendRequest 收益趋势请求DTO
type EarningsTrendRequest struct {
    Hours int `form:"hours" binding:"omitempty,min=1,max=24"` // 查询小时数，默认2小时
}

// EarningsTrendPoint 收益趋势数据点
type EarningsTrendPoint struct {
    Timestamp        string `json:"timestamp"`        // 时间戳 (HH:MM格式)
    Amount           int64  `json:"amount"`           // 收益数量
    BlockHeight      uint64 `json:"block_height"`     // 区块高度
    TransactionCount int64  `json:"transaction_count"` // 交易数量
    SourceChain      string `json:"source_chain"`     // 来源链
}
```

#### 服务层接口 (`server/internal/services/earnings_service.go`)
```go
// EarningsService 接口新增方法
GetEarningsTrend(ctx context.Context, userID uint64, hours int) ([]*dto.EarningsTrendPoint, error)

// 实现逻辑
func (s *earningsService) GetEarningsTrend(ctx context.Context, userID uint64, hours int) ([]*dto.EarningsTrendPoint, error) {
    // 1. 计算时间范围（默认2小时）
    // 2. 调用仓库层获取指定时间范围内的收益记录
    // 3. 过滤只包含收益增加的记录（type === "add" && source === "block_verification"）
    // 4. 转换为趋势数据点格式
    // 5. 按时间排序返回
}
```

#### 处理器接口 (`server/internal/handlers/earnings_handler.go`)
```go
// GetEarningsTrend 获取收益趋势数据
func (h *EarningsHandler) GetEarningsTrend(c *gin.Context) {
    // 1. 获取用户ID
    // 2. 解析查询参数（hours，默认2小时）
    // 3. 调用服务层获取趋势数据
    // 4. 返回JSON响应
}
```

#### 路由配置 (`server/internal/routes/routes.go`)
```go
// 收益相关路由
earnings := v1.Group("/earnings")
{
    // ... 其他接口
    earnings.GET("/trend", earningsHandler.GetEarningsTrend) // 获取收益趋势数据
}
```

### 2. 前端新增接口

#### 类型定义 (`vue/src/types/earnings.ts`)
```typescript
export interface EarningsTrendPoint {
  timestamp: string
  amount: number
  block_height: number
  transaction_count: number
  source_chain: string
}
```

#### API函数 (`vue/src/api/earnings/index.ts`)
```typescript
/**
 * 获取收益趋势数据
 */
export function getEarningsTrend(hours: number = 2): Promise<ApiResponse<EarningsTrendPoint[]>> {
  if (__USE_MOCK__) {
    return handleMockGetEarningsTrend(hours)
  }
  
  return request({
    url: '/api/v1/earnings/trend',
    method: 'GET',
    params: { hours }
  })
}
```

#### Mock数据处理 (`vue/src/api/mock/earnings.ts`)
```typescript
export const handleMockGetEarningsTrend = (hours: number = 2): Promise<any> => {
  // 生成模拟的趋势数据
  // 支持动态小时数参数
  // 返回标准化的趋势数据点
}
```

#### Mock数据示例 (`vue/ApiDatas/earnings/earnings-v1.json`)
```json
"/earnings/trend": {
  "get": {
    "summary": "获取收益趋势数据",
    "parameters": [
      {
        "name": "hours",
        "in": "query",
        "description": "查询小时数（1-24，默认2小时）",
        "required": false,
        "schema": {
          "type": "integer",
          "minimum": 1,
          "maximum": 24,
          "default": 2
        }
      }
    ],
    "responses": {
      "200": {
        "content": {
          "application/json": {
            "example": {
              "success": true,
              "message": "获取收益趋势成功",
              "data": [
                {
                  "timestamp": "15:30",
                  "amount": 193,
                  "block_height": 23201391,
                  "transaction_count": 193,
                  "source_chain": "eth"
                }
              ]
            }
          }
        }
      }
    }
  }
}
```

## 🔄 前端页面更新

### 图表数据加载逻辑
```typescript
// 更新前：依赖前端过滤
const recentRecords = earningsList.value.filter(record => {
  const recordTime = new Date(record.created_at)
  return recordTime >= twoHoursAgo && record.type === 'add'
})

// 更新后：调用专门接口
const trendResponse = await getEarningsTrend(2) // 默认2小时
if (trendResponse.success && trendResponse.data) {
  const trendData = trendResponse.data
  const labels = trendData.map(point => point.timestamp)
  const data = trendData.map(point => point.amount)
  // 创建图表...
}
```

### 组件生命周期
```typescript
onMounted(async () => {
  await loadUserData()        // 加载用户数据
  await loadEarnings()        // 加载收益记录
  await createEarningsChart() // 独立加载图表数据
})
```

## ✅ 优势对比

| 方面 | 更新前（前端过滤） | 更新后（专门接口） |
|------|-------------------|-------------------|
| **性能** | 需要加载完整收益记录，前端过滤 | 只加载需要的趋势数据 |
| **数据准确性** | 依赖前端数据状态 | 直接从数据库查询最新数据 |
| **可扩展性** | 难以支持不同时间范围 | 支持1-24小时动态查询 |
| **维护性** | 逻辑分散在前端 | 逻辑集中在后端服务 |
| **实时性** | 依赖收益记录列表的更新 | 独立的数据查询，更实时 |

## 🚀 使用方式

### 1. 获取默认2小时趋势
```typescript
const trendData = await getEarningsTrend()
```

### 2. 获取指定小时数趋势
```typescript
const trendData = await getEarningsTrend(6) // 获取6小时趋势
```

### 3. 后端API调用
```bash
# 默认2小时
GET /api/v1/earnings/trend

# 指定6小时
GET /api/v1/earnings/trend?hours=6
```

## 📊 数据格式

### 请求参数
- `hours` (可选): 查询小时数，范围1-24，默认2

### 响应数据
```typescript
{
  success: true,
  message: "获取收益趋势成功",
  data: [
    {
      timestamp: "15:30",        // 时间戳 (HH:MM)
      amount: 193,               // 收益数量 (TB)
      block_height: 23201391,    // 区块高度
      transaction_count: 193,    // 交易数量
      source_chain: "eth"        // 来源链
    }
  ]
}
```

## 🔧 技术实现细节

### 1. 时间范围计算
```go
endTime := time.Now()
startTime := endTime.Add(-time.Duration(hours) * time.Hour)
```

### 2. 数据过滤条件
```go
if record.Type == "add" && record.Source == "block_verification" {
    // 只包含扫块收益记录
}
```

### 3. 时间排序
```go
sort.Slice(trendPoints, func(i, j int) bool {
    timeI, _ := time.Parse("15:04", trendPoints[i].Timestamp)
    timeJ, _ := time.Parse("15:04", trendPoints[j].Timestamp)
    return timeI.Before(timeJ)
})
```

## 🎉 总结

通过创建专门的收益趋势API接口，我们实现了：

1. **性能优化**: 避免前端过滤大量数据
2. **数据准确性**: 直接从数据库获取最新趋势数据
3. **接口标准化**: 符合RESTful API设计原则
4. **可扩展性**: 支持动态时间范围查询
5. **维护性提升**: 前后端职责分离更清晰

这种设计模式可以应用到其他需要图表数据的场景，为后续功能扩展提供了良好的基础。
