-- 添加 allowance_address 字段到 user_transactions 表
-- 用于存储代币持有者地址（transferFrom 操作中的 from 参数）

ALTER TABLE user_transactions 
ADD COLUMN allowance_address VARCHAR(120) DEFAULT NULL 
COMMENT '授权地址（代币持有者地址）' 
AFTER token_contract_address;

-- 添加索引以提高查询性能
CREATE INDEX idx_user_transactions_allowance_address ON user_transactions(allowance_address);
