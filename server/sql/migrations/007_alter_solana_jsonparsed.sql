-- Alter Solana tables to support jsonParsed structure and add redundant block_id

-- Add block_id to sol_tx_detail and new json fields if needed later
ALTER TABLE `sol_tx_detail`
  ADD COLUMN `block_id` bigint(20) unsigned NULL COMMENT '冗余 blocks 表的ID' AFTER `slot`,
  ADD KEY `idx_sol_tx_detail_block_id` (`block_id`);

-- Extend sol_instruction for jsonParsed
ALTER TABLE `sol_instruction`
  ADD COLUMN `block_id` bigint(20) unsigned NULL COMMENT '冗余 blocks 表的ID' AFTER `tx_id`,
  ADD COLUMN `slot` bigint(20) unsigned NULL COMMENT 'slot高度 (便于查询)' AFTER `block_id`,
  ADD COLUMN `program_id_index` int(11) NULL AFTER `program_id`,
  ADD COLUMN `stack_height` int(11) NULL AFTER `program_id_index`,
  ADD COLUMN `accounts_idx_json` longtext NULL COMMENT 'accounts索引数组JSON(来自message.instructions.accounts)' AFTER `accounts_json`,
  ADD COLUMN `accounts_keys_json` longtext NULL COMMENT '展开后的账户pubkey数组JSON(可选)' AFTER `accounts_idx_json`,
  ADD COLUMN `data_json` longtext NULL COMMENT 'jsonParsed结构化数据(如有)' AFTER `data_b58`;

ALTER TABLE `sol_instruction`
  ADD KEY `idx_solins_block_id` (`block_id`),
  ADD KEY `idx_solins_slot` (`slot`),
  ADD KEY `idx_solins_progidx` (`program_id_index`);


