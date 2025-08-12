-- 初始化权限类型到baseconfig字典表
-- 用于API密钥的权限范围配置
-- 只包含：上传区块、上传交易 两个权限类型

INSERT INTO base_config (`group`, no, config_type, config_name, config_key, config_value, description) VALUES
('api_permissions', 1, 1, '上传区块权限', 'blocks_write', 'blocks:write', '上传区块数据权限'),
('api_permissions', 2, 1, '上传交易权限', 'transactions_write', 'transactions:write', '上传交易数据权限');

