-- Add program_id column to contracts table for associating Sol Program ID (e.g., TOKEN_2022_PROGRAM_ID)
-- Up
ALTER TABLE `contracts`
  ADD COLUMN `program_id` VARCHAR(120) NULL DEFAULT NULL COMMENT 'Associated Program ID (e.g., Sol Program ID)'
  AFTER `chain_name`;

-- Extend user_addresses for SOL ATA associations
ALTER TABLE `user_addresses`
  ADD COLUMN `ata_owner_address` VARCHAR(120) NULL DEFAULT '' AFTER `balance_height`,
  ADD COLUMN `ata_mint_address` VARCHAR(120) NULL DEFAULT '' AFTER `ata_owner_address`;

-- Down
-- ALTER TABLE `contracts` DROP COLUMN `program_id`;
-- ALTER TABLE `user_addresses` DROP COLUMN `ata_owner_address`;
-- ALTER TABLE `user_addresses` DROP COLUMN `ata_mint_address`;


