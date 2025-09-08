-- 删除不必要的合约余额高度与索引字段
ALTER TABLE user_addresses
  DROP COLUMN IF EXISTS contract_balance_height,
  DROP COLUMN IF EXISTS contract_balance_in_tx_index;


