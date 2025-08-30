-- 性能优化索引
-- 为首页统计查询创建复合索引

-- 1. 为交易表的统计查询创建复合索引
-- 这个索引将优化：chain + ctime + gas_price 的组合查询
CREATE INDEX IF NOT EXISTS idx_transaction_stats_chain_ctime_gas 
ON `transaction` (chain, ctime, gas_price) 
WHERE gas_price IS NOT NULL AND gas_price != '' AND gas_price != '0';

-- 2. 为交易表的交易量统计创建索引
-- 这个索引将优化：chain + ctime + amount 的组合查询
CREATE INDEX IF NOT EXISTS idx_transaction_volume_chain_ctime_amount 
ON `transaction` (chain, ctime, amount) 
WHERE amount IS NOT NULL AND amount != '';

-- 3. 为区块表的出块时间统计创建索引
-- 这个索引将优化：chain + timestamp 的组合查询
CREATE INDEX IF NOT EXISTS idx_block_time_chain_timestamp 
ON `blocks` (chain, timestamp);

-- 4. 为交易表的基础查询创建索引
CREATE INDEX IF NOT EXISTS idx_transaction_chain_deleted 
ON `transaction` (chain, deleted_at);

-- 5. 为区块表的基础查询创建索引
CREATE INDEX IF NOT EXISTS idx_block_chain_deleted 
ON `blocks` (chain, deleted_at);

-- 6. 为交易表的Gas价格查询创建专用索引（如果上面的复合索引不够快）
CREATE INDEX IF NOT EXISTS idx_transaction_gas_price_chain_ctime 
ON `transaction` (chain, ctime) 
WHERE gas_price IS NOT NULL AND gas_price != '' AND gas_price != '0';

-- 7. 新增：为首页统计查询优化的复合索引
-- 优化 chain + ctime + amount + gas_price 的组合查询
CREATE INDEX IF NOT EXISTS idx_transaction_home_stats 
ON `transaction` (chain, ctime, amount, gas_price) 
WHERE amount IS NOT NULL AND amount != '' AND gas_price IS NOT NULL AND gas_price != '' AND gas_price != '0';

-- 8. 新增：为区块验证状态查询优化
-- 优化 chain + is_verified + height 的组合查询
CREATE INDEX IF NOT EXISTS idx_block_verified_chain_height 
ON `blocks` (chain, is_verified, height) 
WHERE is_verified = 1;

-- 9. 新增：为区块基础费用查询优化
-- 优化 chain + base_fee + height 的组合查询
CREATE INDEX IF NOT EXISTS idx_block_base_fee_chain_height 
ON `blocks` (chain, base_fee, height) 
WHERE base_fee IS NOT NULL AND base_fee != '';

-- 10. 新增：为交易表的时间范围查询优化
-- 优化 chain + ctime 的范围查询
CREATE INDEX IF NOT EXISTS idx_transaction_chain_ctime_range 
ON `transaction` (chain, ctime);

-- 11. 新增：为币种配置查询优化
-- 优化 chain_name + status 的组合查询
CREATE INDEX IF NOT EXISTS idx_coin_config_chain_status 
ON `coin_config` (chain_name, status);

-- 12. 新增：为交易表的高度查询优化
-- 优化 chain + height + block_index 的组合查询
CREATE INDEX IF NOT EXISTS idx_transaction_chain_height_index 
ON `transaction` (chain, height, block_index);
