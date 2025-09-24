-- Insert system programs into sol_program table
-- This enables parsing of basic Solana system and token transactions

-- System Program (11111111111111111111111111111112)
INSERT INTO `sol_program` (
  `program_id`,
  `name`,
  `alias`,
  `category`,
  `type`,
  `is_system`,
  `version`,
  `status`,
  `description`,
  `instruction_rules`,
  `event_rules`,
  `sample_data`,
  `ctime`,
  `mtime`
) VALUES (
  '11111111111111111111111111111112',
  'System Program',
  'System',
  'System',
  'system',
  1,
  'v1',
  'active',
  'Solana native system program for basic operations like transfers and account creation.',
  '{}',
  '{}',
  '{}',
  NOW(),
  NOW()
)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `alias` = VALUES(`alias`),
  `category` = VALUES(`category`),
  `type` = VALUES(`type`),
  `is_system` = VALUES(`is_system`),
  `version` = VALUES(`version`),
  `status` = VALUES(`status`),
  `description` = VALUES(`description`),
  `mtime` = VALUES(`mtime`);

-- SPL Token Program (TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA)
INSERT INTO `sol_program` (
  `program_id`,
  `name`,
  `alias`,
  `category`,
  `type`,
  `is_system`,
  `version`,
  `status`,
  `description`,
  `instruction_rules`,
  `event_rules`,
  `sample_data`,
  `ctime`,
  `mtime`
) VALUES (
  'TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA',
  'SPL Token Program',
  'SPL Token',
  'System',
  'token',
  1,
  'v1',
  'active',
  'Solana Program Library Token program for token operations.',
  '{}',
  '{}',
  '{}',
  NOW(),
  NOW()
)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `alias` = VALUES(`alias`),
  `category` = VALUES(`category`),
  `type` = VALUES(`type`),
  `is_system` = VALUES(`is_system`),
  `version` = VALUES(`version`),
  `status` = VALUES(`status`),
  `description` = VALUES(`description`),
  `mtime` = VALUES(`mtime`);

-- Associated Token Account Program (ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL)
INSERT INTO `sol_program` (
  `program_id`,
  `name`,
  `alias`,
  `category`,
  `type`,
  `is_system`,
  `version`,
  `status`,
  `description`,
  `instruction_rules`,
  `event_rules`,
  `sample_data`,
  `ctime`,
  `mtime`
) VALUES (
  'ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL',
  'Associated Token Account Program',
  'ATA',
  'System',
  'spl-token',
  1,
  'v1',
  'active',
  'Associated Token Account program for managing token accounts.',
  '{}',
  '{}',
  '{}',
  NOW(),
  NOW()
)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `alias` = VALUES(`alias`),
  `category` = VALUES(`category`),
  `type` = VALUES(`type`),
  `is_system` = VALUES(`is_system`),
  `version` = VALUES(`version`),
  `status` = VALUES(`status`),
  `description` = VALUES(`description`),
  `mtime` = VALUES(`mtime`);
