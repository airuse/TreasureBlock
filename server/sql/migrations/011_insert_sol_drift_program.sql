-- Insert Drift V2 Program metadata into sol_program
-- Reference: https://solscan.io/account/dRiftyHA39MWEi3m9aunc5MzRF1JYuBsbn6VPcn33UH

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
  'dRiftyHA39MWEi3m9aunc5MzRF1JYuBsbn6VPcn33UH',
  'Drift V2',
  'Drift',
  'DeFi',
  'Protocol',
  0,
  'v2',
  'active',
  'Inserted by migration 011: Drift perpetuals protocol program metadata.',
  '{
    "schema_version": 1,
    "program_id": "dRiftyHA39MWEi3m9aunc5MzRF1JYuBsbn6VPcn33UH",
    "instructions": [
      {
        "instruction_type": "initializeUser",
        "selector": { "parsed.key": "initializeUser" },
        "output": {
          "type": "INSTRUCTION",
          "instruction_type": "initialize_user",
          "fields": [
            { "name": "authority", "path": "accounts[0]" },
            { "name": "user", "path": "accounts[1]" }
          ]
        }
      },
      {
        "instruction_type": "deposit",
        "selector": { "parsed.key": "deposit" },
        "output": {
          "type": "EVENT",
          "event_type": "deposit",
          "fields": [
            { "name": "user", "path": "accounts[0]" },
            { "name": "amount", "path": "parsed.value.amount" },
            { "name": "mint", "path": "parsed.value.mint" },
            { "name": "decimals", "path": "parsed.value.decimals", "default": 6 }
          ]
        }
      },
      {
        "instruction_type": "withdraw",
        "selector": { "parsed.key": "withdraw" },
        "output": {
          "type": "EVENT",
          "event_type": "withdraw",
          "fields": [
            { "name": "user", "path": "accounts[0]" },
            { "name": "amount", "path": "parsed.value.amount" },
            { "name": "mint", "path": "parsed.value.mint" },
            { "name": "decimals", "path": "parsed.value.decimals", "default": 6 }
          ]
        }
      },
      {
        "instruction_type": "placePerpOrder",
        "selector": { "parsed.key": "placePerpOrder" },
        "output": {
          "type": "INSTRUCTION",
          "instruction_type": "place_perp_order",
          "fields": [
            { "name": "marketIndex", "path": "parsed.value.marketIndex" },
            { "name": "direction", "path": "parsed.value.direction" },
            { "name": "baseAssetAmount", "path": "parsed.value.baseAssetAmount" }
          ]
        }
      }
    ]
  }',
  '{
    "schema_version": 1,
    "events": [
      {
        "event_type": "deposit",
        "mapping": {
          "ProgramID": "context.programId",
          "FromAddress": "accounts[0]",
          "ToAddress": "accounts[1]",
          "Amount": "parsed.value.amount",
          "Mint": "parsed.value.mint",
          "Decimals": "parsed.value.decimals"
        }
      },
      {
        "event_type": "withdraw",
        "mapping": {
          "ProgramID": "context.programId",
          "FromAddress": "accounts[1]",
          "ToAddress": "accounts[0]",
          "Amount": "parsed.value.amount",
          "Mint": "parsed.value.mint",
          "Decimals": "parsed.value.decimals"
        }
      }
    ]
  }',
  '{
    "tx": {
      "tx_id": "ExAmPlEtxId111",
      "slot": 100,
      "instructions": [
        {
          "programId": "dRiftyHA39MWEi3m9aunc5MzRF1JYuBsbn6VPcn33UH",
          "parsed": { "key": "deposit", "value": { "amount": "1000000", "mint": "USDCmint111", "decimals": 6 } },
          "accounts": ["User111", "Vault111"]
        }
      ]
    },
    "expected": {
      "events": [
        { "event_type": "deposit", "amount": "1000000", "mint": "USDCmint111", "decimals": 6 }
      ],
      "instructions": [
        { "instruction_type": "place_perp_order", "marketIndex": 0, "direction": "long" }
      ]
    }
  }',
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


