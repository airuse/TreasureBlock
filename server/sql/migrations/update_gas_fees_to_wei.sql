-- 更新Gas费用单位从Gwei到Wei的迁移脚本
-- 执行时间：2024-01-XX
-- 说明：将user_transactions表中的max_priority_fee_per_gas和max_fee_per_gas从Gwei单位转换为Wei单位

-- 备份原始数据（可选）
-- CREATE TABLE user_transactions_backup AS SELECT * FROM user_transactions;

-- 更新max_priority_fee_per_gas：将Gwei转换为Wei (乘以10^9)
UPDATE user_transactions 
SET max_priority_fee_per_gas = CAST(CAST(max_priority_fee_per_gas AS DECIMAL(65,0)) * 1000000000 AS CHAR(100))
WHERE max_priority_fee_per_gas IS NOT NULL 
  AND max_priority_fee_per_gas != ''
  AND CAST(max_priority_fee_per_gas AS DECIMAL(65,0)) < 1000000000; -- 小于10^9的值认为是Gwei

-- 更新max_fee_per_gas：将Gwei转换为Wei (乘以10^9)
UPDATE user_transactions 
SET max_fee_per_gas = CAST(CAST(max_fee_per_gas AS DECIMAL(65,0)) * 1000000000 AS CHAR(100))
WHERE max_fee_per_gas IS NOT NULL 
  AND max_fee_per_gas != ''
  AND CAST(max_fee_per_gas AS DECIMAL(65,0)) < 1000000000; -- 小于10^9的值认为是Gwei

-- 验证更新结果
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
