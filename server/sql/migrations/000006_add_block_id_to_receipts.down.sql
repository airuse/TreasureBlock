-- 回滚：移除 block_id 字段
ALTER TABLE `transaction_receipts` 
DROP INDEX `idx_block_id`;

ALTER TABLE `transaction_receipts` 
DROP COLUMN `block_id`;
