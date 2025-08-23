-- 回滚 transaction_receipts 表结构更改

-- 1. 移除添加的索引
DROP INDEX idx_transaction_receipts_effective_gas_price ON transaction_receipts;
DROP INDEX idx_transaction_receipts_blob_gas_used ON transaction_receipts;
DROP INDEX idx_transaction_receipts_blob_gas_price ON transaction_receipts;

-- 2. 移除添加的字段
ALTER TABLE transaction_receipts DROP COLUMN effective_gas_price;
ALTER TABLE transaction_receipts DROP COLUMN blob_gas_used;
ALTER TABLE transaction_receipts DROP COLUMN blob_gas_price;

-- 3. 恢复 cumulative_gas_used 字段
ALTER TABLE transaction_receipts ADD COLUMN cumulative_gas_used BIGINT UNSIGNED DEFAULT 0 COMMENT '累计Gas使用量';
