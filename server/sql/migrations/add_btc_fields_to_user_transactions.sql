-- 为user_transactions表添加BTC特有字段
-- 执行时间: 2024-01-XX

-- 添加BTC原始交易字段
ALTER TABLE user_transactions 
ADD COLUMN btc_version INT COMMENT 'BTC原始交易Version',
ADD COLUMN btc_lock_time INT UNSIGNED COMMENT 'BTC原始交易LockTime',
ADD COLUMN btc_tx_in_json LONGTEXT COMMENT 'BTC TxIn数组(JSON)',
ADD COLUMN btc_tx_out_json LONGTEXT COMMENT 'BTC TxOut数组(JSON)';

-- 添加索引以提高查询性能
CREATE INDEX idx_user_transactions_btc_version ON user_transactions(btc_version);
CREATE INDEX idx_user_transactions_btc_lock_time ON user_transactions(btc_lock_time);

-- 添加注释
ALTER TABLE user_transactions COMMENT = '用户交易表 - 存储用户创建的待签名交易，支持ETH和BTC';
