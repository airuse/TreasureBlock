-- 在 transaction_receipts 表中添加 block_id 字段
ALTER TABLE `transaction_receipts` 
ADD COLUMN `block_id` BIGINT UNSIGNED COMMENT '关联的区块ID' AFTER `BlobGasPrice`;

-- 为 block_id 字段添加索引
ALTER TABLE `transaction_receipts` 
ADD INDEX `idx_block_id` (`block_id`);

-- 更新现有数据：根据 block_hash 查找对应的 block_id 并设置
UPDATE `transaction_receipts` tr
JOIN `blocks` b ON tr.block_hash = b.hash
SET tr.block_id = b.id
WHERE tr.block_id IS NULL AND tr.block_hash IS NOT NULL;
