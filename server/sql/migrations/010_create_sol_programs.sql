-- +goose Up
CREATE TABLE IF NOT EXISTS `sol_program` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `program_id` varchar(120) NOT NULL COMMENT 'Sol 程序ID(地址)',
  `name` varchar(120) NOT NULL DEFAULT '' COMMENT '程序名称',
  `alias` varchar(120) NOT NULL DEFAULT '' COMMENT '别名',
  `category` varchar(50) NOT NULL DEFAULT '' COMMENT '分类',
  `type` varchar(50) NOT NULL DEFAULT '' COMMENT '类型',
  `is_system` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否系统程序',
  `version` varchar(20) NOT NULL DEFAULT '' COMMENT '版本',
  `status` varchar(20) NOT NULL DEFAULT 'active' COMMENT '状态',
  `description` text COMMENT '描述',
  `instruction_rules` json NULL COMMENT '指令解析规则',
  `event_rules` json NULL COMMENT '事件解析规则',
  `sample_data` json NULL COMMENT '样例',
  `ctime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `mtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_program_id` (`program_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `sol_parsed_extra` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tx_id` varchar(120) NOT NULL COMMENT '交易签名',
  `block_id` bigint unsigned NULL,
  `slot` bigint unsigned NOT NULL,
  `program_id` varchar(120) NOT NULL DEFAULT '' COMMENT '程序ID',
  `is_inner` tinyint(1) NOT NULL DEFAULT 0,
  `data` json NULL,
  `ctime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_txid` (`tx_id`),
  KEY `idx_slot` (`slot`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS `sol_parsed_extra`;
DROP TABLE IF EXISTS `sol_program`;


