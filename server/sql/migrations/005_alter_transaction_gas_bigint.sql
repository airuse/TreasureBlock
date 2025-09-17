-- Expand gas columns to support large EVM chains (BSC/ETH, etc.)
-- idempotent-ish: check existence/compatibility is up to migration runner

ALTER TABLE `transaction`
  MODIFY COLUMN `gas_limit` BIGINT UNSIGNED NOT NULL COMMENT '燃油限制',
  MODIFY COLUMN `gas_used`  BIGINT UNSIGNED NOT NULL COMMENT '实际使用燃油';


