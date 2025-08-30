-- 应用性能优化索引脚本
-- 执行此脚本前请确保数据库连接正常
-- 注意：如果索引已存在，请先手动删除再执行此脚本

-- 1. 为交易表的统计查询创建复合索引
CREATE INDEX idx_transaction_stats_chain_ctime_gas 
ON `transaction` (chain, ctime, gas_price);

-- 2. 为交易表的交易量统计创建索引
CREATE INDEX idx_transaction_volume_chain_ctime_amount 
ON `transaction` (chain, ctime, amount);

-- 3. 为区块表的出块时间统计创建索引
CREATE INDEX idx_block_time_chain_timestamp 
ON `blocks` (chain, timestamp);

-- 4. 为交易表的基础查询创建索引
CREATE INDEX idx_transaction_chain_deleted 
ON `transaction` (chain, deleted_at);

-- 5. 为区块表的基础查询创建索引
CREATE INDEX idx_block_chain_deleted 
ON `blocks` (chain, deleted_at);

-- 6. 为交易表的Gas价格查询创建专用索引
CREATE INDEX idx_transaction_gas_price_chain_ctime 
ON `transaction` (chain, ctime);

-- 7. 为首页统计查询优化的复合索引
CREATE INDEX idx_transaction_home_stats 
ON `transaction` (chain, ctime, amount, gas_price);

-- 8. 为区块验证状态查询优化
CREATE INDEX idx_block_verified_chain_height 
ON `blocks` (chain, is_verified, height);

-- 9. 为区块基础费用查询优化
CREATE INDEX idx_block_base_fee_chain_height 
ON `blocks` (chain, base_fee, height);

-- 10. 为交易表的时间范围查询优化
CREATE INDEX idx_transaction_chain_ctime_range 
ON `transaction` (chain, ctime);

-- 11. 为币种配置查询优化
CREATE INDEX idx_coin_config_chain_status 
ON `coin_config` (chain_name, status);

-- 12. 为交易表的高度查询优化
CREATE INDEX idx_transaction_chain_height_index 
ON `transaction` (chain, height, block_index);

-- 显示创建的索引
SELECT 
    index_name,
    table_name,
    column_name,
    seq_in_index
FROM information_schema.statistics 
WHERE table_schema = DATABASE() 
AND table_name IN ('transaction', 'blocks', 'coin_config')
ORDER BY table_name, index_name, seq_in_index;
