# 扫块收益系统 (T币系统)

## 概述

扫块收益系统是区块链浏览器的核心功能之一，用户通过验证区块可以获得T币奖励。T币的数量根据区块中的交易数量动态计算，1个交易对应1个T币。

## 核心功能

### 1. 收益计算规则
- **扫块收益**：每验证一个区块，用户获得的T币数量 = 区块中的交易数量
- **收益类型**：
  - `add`：增加收益（扫块验证、转账接收）
  - `decrease`：减少收益（转账支出、业务消耗）

### 2. 数据库表结构

#### 收益流水表 (earnings_records)
```sql
CREATE TABLE earnings_records (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,                    -- 用户ID
    amount BIGINT NOT NULL,                              -- T币数量
    type VARCHAR(20) NOT NULL,                           -- 类型：add/decrease
    source VARCHAR(50) NOT NULL,                         -- 来源：block_verification/transfer_out等
    source_id BIGINT UNSIGNED NULL,                      -- 来源ID（如区块ID）
    source_chain VARCHAR(50) NULL,                       -- 来源链名称
    block_height BIGINT UNSIGNED NULL,                   -- 相关区块高度
    transaction_count BIGINT NULL DEFAULT 0,             -- 相关交易数量
    description VARCHAR(255) NOT NULL DEFAULT '',        -- 描述信息
    balance_before BIGINT NOT NULL,                      -- 操作前余额
    balance_after BIGINT NOT NULL,                       -- 操作后余额
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

#### 用户余额表 (user_balances)
```sql
CREATE TABLE user_balances (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL UNIQUE,            -- 用户ID（唯一）
    balance BIGINT NOT NULL DEFAULT 0,                  -- 当前余额
    total_earned BIGINT NOT NULL DEFAULT 0,             -- 累计获得的T币
    total_spent BIGINT NOT NULL DEFAULT 0,              -- 累计消耗的T币
    last_earning_time TIMESTAMP NULL,                   -- 最后一次获得收益时间
    last_spend_time TIMESTAMP NULL,                     -- 最后一次消耗时间
    version BIGINT NOT NULL DEFAULT 0,                  -- 版本号（乐观锁）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

## API接口

### 1. 获取用户余额
```http
GET /api/v1/earnings/balance
Authorization: Bearer <JWT_TOKEN>
```

**响应示例：**
```json
{
    "success": true,
    "message": "获取余额成功",
    "data": {
        "id": 1,
        "user_id": 123,
        "balance": 1500,
        "total_earned": 2000,
        "total_spent": 500,
        "last_earning_time": "2024-01-15T10:30:00Z",
        "last_spend_time": "2024-01-14T09:15:00Z",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
    }
}
```

### 2. 获取收益记录列表
```http
GET /api/v1/earnings/records?page=1&page_size=20&type=add&chain=ethereum
Authorization: Bearer <JWT_TOKEN>
```

**查询参数：**
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认20，最大100）
- `type`: 类型过滤（add/decrease）
- `source`: 来源过滤
- `chain`: 链过滤
- `start_date`: 开始日期（YYYY-MM-DD）
- `end_date`: 结束日期（YYYY-MM-DD）

**响应示例：**
```json
{
    "success": true,
    "message": "获取收益记录成功",
    "data": {
        "records": [
            {
                "id": 1,
                "user_id": 123,
                "amount": 10,
                "type": "add",
                "source": "block_verification",
                "source_id": 12345,
                "source_chain": "ethereum",
                "block_height": 18500000,
                "transaction_count": 10,
                "description": "扫块收益 - 区块高度: 18500000, 交易数量: 10",
                "balance_before": 1490,
                "balance_after": 1500,
                "created_at": "2024-01-15T10:30:00Z",
                "updated_at": "2024-01-15T10:30:00Z"
            }
        ],
        "pagination": {
            "page": 1,
            "page_size": 20,
            "total": 150
        }
    }
}
```

### 3. 获取收益统计
```http
GET /api/v1/earnings/stats
Authorization: Bearer <JWT_TOKEN>
```

**响应示例：**
```json
{
    "success": true,
    "message": "获取收益统计成功",
    "data": {
        "user_id": 123,
        "total_earnings": 2000,
        "total_spendings": 500,
        "current_balance": 1500,
        "block_count": 200,
        "transaction_count": 2000
    }
}
```

### 4. 转账T币
```http
POST /api/v1/earnings/transfer
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
    "to_user_id": 456,
    "amount": 100,
    "description": "转账给朋友"
}
```

**响应示例：**
```json
{
    "success": true,
    "message": "转账成功",
    "data": {
        "from_user_id": 123,
        "to_user_id": 456,
        "amount": 100,
        "description": "转账给朋友",
        "from_balance_before": 1500,
        "from_balance_after": 1400,
        "to_balance_before": 200,
        "to_balance_after": 300,
        "transfer_time": "2024-01-15T11:00:00Z"
    }
}
```

## 业务流程

### 1. 扫块收益流程

1. 用户调用区块验证接口：`POST /api/v1/blocks/{blockID}/verify`
2. 系统验证区块成功后：
   - 获取区块中的交易数量
   - 计算收益金额（交易数量 × 1）
   - 更新用户余额
   - 创建收益流水记录

### 2. 转账流程

1. 用户调用转账接口：`POST /api/v1/earnings/transfer`
2. 系统验证：
   - 检查余额是否足够
   - 验证目标用户存在
   - 验证转账金额 > 0
3. 执行转账：
   - 扣除发送方余额
   - 增加接收方余额
   - 创建双方的流水记录

### 3. 业务消耗流程

1. 业务系统调用消耗接口（通过服务层）
2. 系统验证余额足够
3. 扣除用户余额并记录流水

## 安全特性

### 1. 乐观锁机制
- 用户余额表使用版本号实现乐观锁
- 防止并发更新导致的数据不一致

### 2. 事务保证
- 余额更新和流水记录创建在同一事务中
- 确保数据一致性

### 3. 权限控制
- 所有接口都需要JWT认证
- 用户只能操作自己的数据

## 使用示例

### 扫块获得收益
```bash
# 验证区块（自动获得收益）
curl -X POST "http://localhost:8080/api/v1/blocks/12345/verify" \
     -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 查看余额
curl -X GET "http://localhost:8080/api/v1/earnings/balance" \
     -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 转账T币
```bash
curl -X POST "http://localhost:8080/api/v1/earnings/transfer" \
     -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
         "to_user_id": 456,
         "amount": 100,
         "description": "转账给朋友"
     }'
```

### 查看收益记录
```bash
curl -X GET "http://localhost:8080/api/v1/earnings/records?page=1&page_size=10&type=add" \
     -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 注意事项

1. **T币只能通过扫块获得**：系统不提供直接增加T币的接口
2. **T币只能通过业务消耗**：除了转账外，T币主要通过业务功能消耗
3. **数据迁移**：首次部署需要运行数据库迁移文件
4. **性能优化**：大量并发时建议使用连接池和缓存
5. **监控告警**：建议对余额异常变化进行监控

## 扩展功能

未来可以考虑的扩展：
- T币兑换功能
- 收益排行榜
- 批量转账
- 定期收益结算
- 收益提现功能
