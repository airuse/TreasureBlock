-- 创建用户交易表
CREATE TABLE IF NOT EXISTS `user_transactions` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint(20) unsigned NOT NULL COMMENT '用户ID',
  `chain` varchar(20) NOT NULL COMMENT '链类型(btc,eth)',
  `symbol` varchar(20) NOT NULL COMMENT '币种',
  `from_address` varchar(120) NOT NULL COMMENT '发送地址',
  `to_address` varchar(120) NOT NULL COMMENT '接收地址',
  `amount` decimal(65,18) NOT NULL COMMENT '交易金额',
  `fee` decimal(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '手续费',
  `gas_limit` int(11) unsigned DEFAULT NULL COMMENT 'Gas限制',
  `gas_price` varchar(100) DEFAULT NULL COMMENT 'Gas价格',
  `nonce` bigint(20) unsigned DEFAULT NULL COMMENT '交易序号',
  `status` varchar(20) NOT NULL DEFAULT 'draft' COMMENT '状态:draft,unsigned,unsent,in_progress,packed,confirmed,failed',
  `tx_hash` varchar(120) DEFAULT NULL COMMENT '交易哈希',
  `unsigned_tx` longtext DEFAULT NULL COMMENT '未签名交易数据',
  `signed_tx` longtext DEFAULT NULL COMMENT '已签名交易数据',
  `block_height` bigint(20) unsigned DEFAULT NULL COMMENT '区块高度',
  `confirmations` int(11) unsigned DEFAULT '0' COMMENT '确认数',
  `error_msg` text DEFAULT NULL COMMENT '错误信息',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_chain` (`chain`),
  KEY `idx_status` (`status`),
  KEY `idx_from_address` (`from_address`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户交易表';

-- 插入一些测试数据（可选）
INSERT INTO `user_transactions` (
  `user_id`, `chain`, `symbol`, `from_address`, `to_address`, 
  `amount`, `fee`, `gas_limit`, `gas_price`, `nonce`, 
  `status`, `remark`
) VALUES 
(1, 'eth', 'ETH', '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6', '0x8ba1f109551bD432803012645Hac136c22C177e9', '0.1', '0.00042', 21000, '20', 0, 'draft', '测试交易1'),
(1, 'eth', 'ETH', '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6', '0x1234567890123456789012345678901234567890', '0.05', '0.00042', 21000, '20', 1, 'unsigned', '测试交易2'),
(1, 'eth', 'ETH', '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6', '0x8ba1f109551bD432803012645Hac136c22C177e9', '0.2', '0.00042', 21000, '20', 2, 'unsent', '测试交易3'),
(1, 'eth', 'ETH', '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6', '0x1234567890123456789012345678901234567890', '0.15', '0.00042', 21000, '20', 3, 'in_progress', '测试交易4'),
(1, 'eth', 'ETH', '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6', '0x8ba1f109551bD432803012645Hac136c22C177e9', '0.3', '0.00042', 21000, '20', 4, 'confirmed', '测试交易5');
