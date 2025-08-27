# 首页API接口文档

## 概述

首页API提供了区块链浏览器的首页统计数据，包括概览信息、最新区块和最新交易。这些数据可以通过REST API获取，也可以通过WebSocket实时更新。

## API端点

### 获取首页统计数据

**接口地址**: `GET /api/v1/home/stats`

**请求参数**:
- `chain` (string, 必需): 区块链类型，支持 `eth` 和 `btc`

**请求示例**:
```bash
GET /api/v1/home/stats?chain=eth
```

**响应格式**:
```json
{
  "success": true,
  "data": {
    "overview": {
      "totalBlocks": 18500000,
      "totalTransactions": 1250000000,
      "activeAddresses": 45000000,
      "networkHashrate": 850000000000000,
      "dailyVolume": 1250000.5,
      "avgGasPrice": 25000000000,
      "avgBlockTime": 12.5,
      "difficulty": 9223372036854775807
    },
    "latestBlocks": [
      {
        "height": 18500000,
        "hash": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
        "timestamp": 1693123456789,
        "transactions_count": 156,
        "size": 245760,
        "miner": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
        "chain": "eth"
      }
    ],
    "latestTransactions": [
      {
        "hash": "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
        "timestamp": 1693123456789,
        "amount": "0.5",
        "from": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
        "to": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
        "gas_price": "25000000000",
        "gas_used": 21000,
        "chain": "eth"
      }
    ]
  },
  "message": "成功获取首页统计数据"
}
```

## 数据结构

### Overview (概览数据)

| 字段 | 类型 | 描述 |
|------|------|------|
| totalBlocks | int64 | 总区块数 |
| totalTransactions | int64 | 总交易数 |
| activeAddresses | int64 | 活跃地址数（最近24小时） |
| networkHashrate | int64 | 网络算力（H/s） |
| dailyVolume | float64 | 日交易量 |
| avgGasPrice | int64 | 平均Gas价格（Wei） |
| avgBlockTime | float64 | 平均出块时间（秒） |
| difficulty | int64 | 当前难度 |

### BlockSummary (区块摘要)

| 字段 | 类型 | 描述 |
|------|------|------|
| height | int64 | 区块高度 |
| hash | string | 区块哈希 |
| timestamp | int64 | 时间戳（毫秒） |
| transactions_count | int | 交易数量 |
| size | int64 | 区块大小（字节） |
| miner | string | 矿工地址 |
| chain | string | 区块链类型 |

### TransactionSummary (交易摘要)

| 字段 | 类型 | 描述 |
|------|------|------|
| hash | string | 交易哈希 |
| timestamp | int64 | 时间戳（毫秒） |
| amount | string | 交易金额 |
| from | string | 发送方地址 |
| to | string | 接收方地址 |
| gas_price | string | Gas价格 |
| gas_used | int | 使用的Gas |
| chain | string | 区块链类型 |

## 错误响应

### 400 Bad Request
```json
{
  "success": false,
  "error": "链类型参数缺失"
}
```

### 500 Internal Server Error
```json
{
  "success": false,
  "error": "获取概览数据失败: 数据库连接错误"
}
```

## WebSocket实时更新

首页数据支持通过WebSocket实时更新，客户端可以订阅以下频道：

### 订阅区块更新
```json
{
  "type": "subscribe",
  "category": "block",
  "chain": "eth"
}
```

### 订阅交易更新
```json
{
  "type": "subscribe",
  "category": "transaction",
  "chain": "eth"
}
```

### 订阅统计更新
```json
{
  "type": "subscribe",
  "category": "stats",
  "chain": "eth"
}
```

## 实现细节

### 后端架构

1. **HomeHandler**: 处理首页API请求
2. **StatsService**: 提供统计数据服务
3. **BlockService**: 提供区块数据服务
4. **TransactionService**: 提供交易数据服务

### 数据获取策略

1. **概览数据**: 从多个服务聚合获取
2. **最新区块**: 限制返回3个最新区块
3. **最新交易**: 限制返回3个最新交易
4. **缓存策略**: 统计数据可以缓存5分钟

### 性能优化

1. **并发查询**: 并行获取不同类型的数据
2. **数据限制**: 限制返回的数据量
3. **错误处理**: 优雅处理部分数据获取失败的情况

## 使用示例

### JavaScript/TypeScript
```typescript
import { getHomeStats } from '@/api/home'

// 获取ETH首页数据
const loadHomeData = async () => {
  try {
    const response = await getHomeStats({ chain: 'eth' })
    if (response.success && response.data) {
      const { overview, latestBlocks, latestTransactions } = response.data.data
      // 处理数据...
    }
  } catch (error) {
    console.error('获取首页数据失败:', error)
  }
}
```

### cURL
```bash
# 获取ETH首页数据
curl -X GET "http://localhost:8080/api/v1/home/stats?chain=eth" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 获取BTC首页数据
curl -X GET "http://localhost:8080/api/v1/home/stats?chain=btc" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 注意事项

1. **认证要求**: 需要有效的JWT访问令牌
2. **限流保护**: 受API限流策略保护
3. **数据实时性**: 统计数据可能有5-10分钟的延迟
4. **链类型支持**: 目前仅支持ETH和BTC
5. **数据格式**: 所有数值字段都使用字符串类型避免精度丢失

## 未来扩展

1. **更多链支持**: 添加其他区块链的支持
2. **历史数据**: 提供历史统计数据
3. **自定义时间范围**: 支持自定义统计时间范围
4. **数据导出**: 支持数据导出功能
5. **图表数据**: 提供图表所需的聚合数据
