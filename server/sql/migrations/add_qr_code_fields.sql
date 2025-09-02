-- 添加QR码导出相关字段和签名组件字段
-- 执行时间: 2024-01-01

-- 添加QR码导出相关字段
ALTER TABLE user_transactions 
ADD COLUMN chain_id VARCHAR(10) COMMENT '链ID',
ADD COLUMN tx_data LONGTEXT COMMENT '交易数据(十六进制)',
ADD COLUMN access_list LONGTEXT COMMENT '访问列表(JSON格式)';

-- 添加签名组件字段
ALTER TABLE user_transactions 
ADD COLUMN v VARCHAR(100) COMMENT '签名V组件',
ADD COLUMN r VARCHAR(100) COMMENT '签名R组件',
ADD COLUMN s VARCHAR(100) COMMENT '签名S组件';

-- 添加索引以提高查询性能
CREATE INDEX idx_user_transactions_chain_id ON user_transactions(chain_id);
CREATE INDEX idx_user_transactions_v ON user_transactions(v);
CREATE INDEX idx_user_transactions_r ON user_transactions(r);
CREATE INDEX idx_user_transactions_s ON user_transactions(s);
