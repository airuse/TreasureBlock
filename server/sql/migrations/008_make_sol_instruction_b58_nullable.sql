-- Make data_b58 nullable since we only store jsonParsed going forward
ALTER TABLE `sol_instruction`
  MODIFY COLUMN `data_b58` longtext NULL COMMENT '指令数据(Base58编码，已弃用)';


