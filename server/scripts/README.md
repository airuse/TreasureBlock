# Gas费用单位迁移说明

## 问题描述

在之前的实现中，数据库中的 `max_priority_fee_per_gas` 和 `max_fee_per_gas` 字段存储的是 Gwei 单位，但代码期望的是 Wei 单位。这导致了 Gas 费用计算错误，交易成本被放大了 10^9 倍。

## 解决方案

1. **代码修复**：更新了 `user_transaction_service.go` 中的逻辑，自动检测并转换 Gwei 到 Wei
2. **数据迁移**：提供了迁移脚本将现有数据库记录从 Gwei 转换为 Wei
3. **单位统一**：确保所有 Gas 费用相关字段统一使用 Wei 单位

## 迁移步骤

### 方法1：使用迁移脚本（推荐）

```bash
# 进入脚本目录
cd server/scripts

# 执行迁移
./run_migration.sh
```

### 方法2：手动执行SQL

```sql
-- 备份数据（可选）
CREATE TABLE user_transactions_backup AS SELECT * FROM user_transactions;

-- 更新 max_priority_fee_per_gas
UPDATE user_transactions 
SET max_priority_fee_per_gas = CAST(CAST(max_priority_fee_per_gas AS DECIMAL(65,0)) * 1000000000 AS CHAR(100))
WHERE max_priority_fee_per_gas IS NOT NULL 
  AND max_priority_fee_per_gas != ''
  AND CAST(max_priority_fee_per_gas AS DECIMAL(65,0)) < 1000000000;

-- 更新 max_fee_per_gas
UPDATE user_transactions 
SET max_fee_per_gas = CAST(CAST(max_fee_per_gas AS DECIMAL(65,0)) * 1000000000 AS CHAR(100))
WHERE max_fee_per_gas IS NOT NULL 
  AND max_fee_per_gas != ''
  AND CAST(max_fee_per_gas AS DECIMAL(65,0)) < 1000000000;
```

## 转换规则

- **1 Gwei = 1,000,000,000 Wei (10^9)**
- 只转换小于 10^9 的值（典型的 Gwei 范围）
- 大于等于 10^9 的值认为是已经是 Wei 单位，不进行转换

## 验证结果

迁移完成后，可以执行以下查询验证结果：

```sql
SELECT 
    id,
    max_priority_fee_per_gas,
    max_fee_per_gas,
    status,
    created_at
FROM user_transactions 
WHERE max_priority_fee_per_gas IS NOT NULL 
   OR max_fee_per_gas IS NOT NULL
ORDER BY id DESC
LIMIT 10;
```

## 注意事项

1. **备份数据**：执行迁移前请务必备份数据库
2. **测试环境**：建议先在测试环境执行迁移
3. **监控日志**：迁移过程中注意观察日志输出
4. **回滚方案**：如有问题，可以使用备份表恢复数据

## 修复效果

- **修复前**：Gas费用 = 30 Gwei × GasLimit = 30 × 10^9 × GasLimit Wei（错误）
- **修复后**：Gas费用 = 30 Gwei × GasLimit = 30,000,000,000 × GasLimit Wei（正确）

交易成本从约 212 ETH 降低到约 0.00003 ETH，解决了余额不足的问题。
