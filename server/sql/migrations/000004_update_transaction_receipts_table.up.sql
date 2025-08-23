-- 更新 transaction_receipts 表结构
-- 添加缺失的字段，移除无用的字段

-- 1. 移除无用的 cumulative_gas_used 字段
ALTER TABLE transaction_receipts DROP COLUMN cumulative_gas_used;

-- 2. 添加 EffectiveGasPrice 字段
ALTER TABLE transaction_receipts ADD COLUMN effective_gas_price VARCHAR(100) COMMENT '有效Gas价格(wei)';

-- 3. 添加 BlobGasUsed 字段
ALTER TABLE transaction_receipts ADD COLUMN blob_gas_used BIGINT UNSIGNED DEFAULT 0 COMMENT 'Blob Gas使用量';

-- 4. 添加 BlobGasPrice 字段
ALTER TABLE transaction_receipts ADD COLUMN blob_gas_price VARCHAR(100) COMMENT 'Blob Gas价格(wei)';

-- 5. 确保 block_number 字段存在（如果不存在的话）
-- ALTER TABLE transaction_receipts ADD COLUMN block_number BIGINT UNSIGNED COMMENT '区块号';

-- 6. 添加索引以提高查询性能
CREATE INDEX idx_transaction_receipts_effective_gas_price ON transaction_receipts(effective_gas_price);
CREATE INDEX idx_transaction_receipts_blob_gas_used ON transaction_receipts(blob_gas_used);
CREATE INDEX idx_transaction_receipts_blob_gas_price ON transaction_receipts(blob_gas_price);
