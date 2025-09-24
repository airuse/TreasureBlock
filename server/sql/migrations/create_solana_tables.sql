-- 创建 Solana 相关表
-- 执行时间: 2024-12-XX

-- 创建 Solana 交易明细表
CREATE TABLE IF NOT EXISTS `sol_tx_detail` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tx_id` varchar(120) NOT NULL COMMENT '交易签名',
  `slot` bigint(20) unsigned NOT NULL COMMENT 'slot高度',
  `blockhash` varchar(120) NOT NULL DEFAULT '' COMMENT '区块哈希',
  `recent_blockhash` varchar(120) NOT NULL DEFAULT '' COMMENT '最近区块哈希',
  `version` varchar(20) NOT NULL DEFAULT 'legacy' COMMENT '交易版本',
  `fee` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '交易费用(lamports)',
  `compute_units` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '消耗的计算单元',
  `cost_units` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '成本单位',
  `status_json` longtext NOT NULL COMMENT 'status结构JSON',
  `account_keys` longtext NOT NULL COMMENT 'accountKeys+loaded地址集合JSON',
  `loaded_writable` longtext NOT NULL COMMENT 'loaded writable地址JSON',
  `loaded_readonly` longtext NOT NULL COMMENT 'loaded readonly地址JSON',
  `pre_balances` longtext NOT NULL COMMENT 'preBalances JSON',
  `post_balances` longtext NOT NULL COMMENT 'postBalances JSON',
  `pre_token_balances` longtext NOT NULL COMMENT 'preTokenBalances JSON',
  `post_token_balances` longtext NOT NULL COMMENT 'postTokenBalances JSON',
  `rewards_json` longtext NOT NULL COMMENT '奖励JSON',
  `logs_json` longtext NOT NULL COMMENT '日志JSON',
  `raw_json` longtext NOT NULL COMMENT '完整RPC返回或交易序列化JSON',
  `ctime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `mtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sol_tx_detail_tx_id` (`tx_id`),
  KEY `idx_sol_tx_detail_slot` (`slot`),
  KEY `idx_sol_tx_detail_ctime` (`ctime`),
  KEY `idx_sol_tx_detail_mtime` (`mtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Solana交易明细表';

-- 创建 Solana 指令表
CREATE TABLE IF NOT EXISTS `sol_instruction` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tx_id` varchar(120) NOT NULL COMMENT '交易签名',
  `outer_index` int(11) NOT NULL DEFAULT '-1' COMMENT '外层指令索引;内层同外层索引',
  `inner_index` int(11) NOT NULL DEFAULT '-1' COMMENT '内层指令在该外层内的索引;外层为-1',
  `program_id` varchar(120) NOT NULL DEFAULT '' COMMENT '程序ID',
  `accounts_json` longtext NOT NULL COMMENT '索引或展开后的账户列表JSON',
  `data_b58` longtext NOT NULL COMMENT '指令数据(Base58编码)',
  `is_inner` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否为内层指令',
  PRIMARY KEY (`id`),
  KEY `idx_solins_txid` (`tx_id`,`outer_index`,`inner_index`),
  KEY `idx_solins_program_id` (`program_id`),
  KEY `idx_solins_is_inner` (`is_inner`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Solana指令表';
