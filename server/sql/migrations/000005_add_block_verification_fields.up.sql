
ALTER TABLE transaction DROP INDEX uk_tx_block;

-- 回滚：删除block_id索引
ALTER TABLE transaction DROP INDEX idx_block_id;

-- 回滚：删除block_id字段
ALTER TABLE transaction DROP COLUMN block_id;

-- 回滚：删除验证相关字段
ALTER TABLE blocks DROP COLUMN is_verified;
ALTER TABLE blocks DROP COLUMN verification_deadline;
ALTER TABLE blocks ADD COLUMN verification_deadline TIMESTAMP NULL COMMENT '最晚验证时间';
ALTER TABLE blocks ADD COLUMN is_verified TINYINT(1) NOT NULL DEFAULT 0 COMMENT '验证是否通过 0:未验证 1:验证通过 2:验证失败';

-- 为transaction表添加block_id字段
ALTER TABLE transaction ADD COLUMN block_id BIGINT UNSIGNED NULL COMMENT '关联的区块ID';
ALTER TABLE transaction ADD COLUMN INDEX idx_block_id (block_id);

-- 添加复合唯一约束：tx_id + block_id
ALTER TABLE transaction ADD UNIQUE INDEX uk_tx_block (tx_id, block_id);


CumulativeGasUsed uint64 `json:"cumulative_gas_used" gorm:"comment:累计Gas使用量"`
	ReceiptType       uint8  `json:"receipt_type" gorm:"comment:收据类型(0:legacy,1:accesslist,2:dynamicfee)"`

-- 为transaction_receipt表添加cumulative_gas_used和receipt_type字段
ALTER TABLE transaction_receipt ADD COLUMN cumulative_gas_used BIGINT NULL COMMENT '累计Gas使用量';
ALTER TABLE transaction_receipt ADD COLUMN receipt_type TINYINT(1) NULL DEFAULT 0 COMMENT '收据类型(0:legacy,1:accesslist,2:dynamicfee)';



