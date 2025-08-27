-- 回滚：删除外键约束（如果存在）
-- ALTER TABLE transaction DROP FOREIGN KEY fk_transaction_block;

-- 回滚：删除复合唯一约束
ALTER TABLE transaction DROP INDEX uk_tx_block;

-- 回滚：删除block_id索引
ALTER TABLE transaction DROP INDEX idx_block_id;

-- 回滚：删除block_id字段
ALTER TABLE transaction DROP COLUMN block_id;

-- 回滚：删除验证相关字段
ALTER TABLE blocks DROP COLUMN is_verified;
ALTER TABLE blocks DROP COLUMN verification_deadline;
