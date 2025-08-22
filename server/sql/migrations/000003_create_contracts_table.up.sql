-- 创建合约表
CREATE TABLE IF NOT EXISTS contracts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    address VARCHAR(42) NOT NULL UNIQUE COMMENT '合约地址',
    chain_name VARCHAR(20) NOT NULL COMMENT '链名称',
    contract_type VARCHAR(50) NOT NULL COMMENT '合约类型',
    name VARCHAR(100) COMMENT '合约名称',
    symbol VARCHAR(20) COMMENT '合约符号',
    decimals TINYINT UNSIGNED DEFAULT 0 COMMENT '小数位数（ERC-20）',
    total_supply VARCHAR(100) COMMENT '总供应量',
    is_erc20 BOOLEAN DEFAULT FALSE COMMENT '是否为ERC-20代币',
    interfaces TEXT COMMENT '支持的接口（JSON格式）',
    methods TEXT COMMENT '可调用的方法（JSON格式）',
    events TEXT COMMENT '支持的事件（JSON格式）',
    metadata TEXT COMMENT '其他元数据（JSON格式）',
    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，0-禁用，2-暂停，3-升级中',
    verified BOOLEAN DEFAULT FALSE COMMENT '是否已验证',
    creator VARCHAR(42) COMMENT '创建者地址',
    creation_tx VARCHAR(66) COMMENT '创建交易哈希',
    creation_block BIGINT UNSIGNED DEFAULT 0 COMMENT '创建区块高度',
    ctime TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    mtime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    -- 索引
    INDEX idx_chain_name (chain_name),
    INDEX idx_contract_type (contract_type),
    INDEX idx_is_erc20 (is_erc20),
    INDEX idx_status (status),
    INDEX idx_verified (verified),
    INDEX idx_creation_block (creation_block),
    INDEX idx_ctime (ctime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='智能合约信息表';
