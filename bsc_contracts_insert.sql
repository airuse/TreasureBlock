-- BSC链主要代币合约数据插入脚本
-- 基于BSCscan tokens页面数据

INSERT INTO `contracts` (
  `id`, 
  `address`, 
  `chain_name`, 
  `contract_type`, 
  `name`, 
  `symbol`, 
  `decimals`, 
  `total_supply`, 
  `is_erc20`, 
  `interfaces`, 
  `methods`, 
  `events`, 
  `metadata`, 
  `status`, 
  `verified`, 
  `creator`, 
  `creation_tx`, 
  `creation_block`, 
  `contract_logo`, 
  `c_time`, 
  `m_time`
) VALUES

-- 1. Wrapped BNB (WBNB) - BSC原生代币包装合约
(101, '0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c', 'bsc', 'BEP20', 'Wrapped BNB', 'WBNB', 18, '0', 1, '["IERC20", "IERC20Metadata"]', '["transfer", "transferFrom", "approve", "allowance", "balanceOf", "totalSupply", "name", "symbol", "decimals", "deposit", "withdraw"]', '["Transfer", "Approval", "Deposit", "Withdrawal"]', '{"name": "Wrapped BNB", "symbol": "WBNB", "decimals": 18}', 1, 1, '0x0000000000000000000000000000000000000000', '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 1, '', '2024-01-01 00:00:00.000', '2025-01-24 11:07:18.897'),

-- 2. Binance-Peg USDT (USDT)
(102, '0x55d398326f99059fF775485246999027B3197955', 'bsc', 'BEP20', 'Tether USD', 'USDT', 18, '0', 1, '["IERC20", "IERC20Metadata"]', '["transfer", "transferFrom", "approve", "allowance", "balanceOf", "totalSupply", "name", "symbol", "decimals"]', '["Transfer", "Approval"]', '{"name": "Tether USD", "symbol": "USDT", "decimals": 18}', 1, 1, '0x0000000000000000000000000000000000000000', '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 1, '', '2024-01-01 00:00:00.000', '2025-01-24 11:07:18.897'),

-- 3. Binance-Peg USDC (USDC)
(103, '0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d', 'bsc', 'BEP20', 'USD Coin', 'USDC', 18, '0', 1, '["IERC20", "IERC20Metadata"]', '["transfer", "transferFrom", "approve", "allowance", "balanceOf", "totalSupply", "name", "symbol", "decimals"]', '["Transfer", "Approval"]', '{"name": "USD Coin", "symbol": "USDC", "decimals": 18}', 1, 1, '0x0000000000000000000000000000000000000000', '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 1, '', '2024-01-01 00:00:00.000', '2025-01-24 11:07:18.897'),

-- 4. Binance-Peg Ethereum Token (ETH)
(104, '0x2170Ed0880ac9A755fd29B2688956BD959F933F8', 'bsc', 'BEP20', 'Binance-Peg Ethereum Token', 'ETH', 18, '0', 1, '["IERC20", "IERC20Metadata"]', '["transfer", "transferFrom", "approve", "allowance", "balanceOf", "totalSupply", "name", "symbol", "decimals"]', '["Transfer", "Approval"]', '{"name": "Binance-Peg Ethereum Token", "symbol": "ETH", "decimals": 18}', 1, 1, '0x0000000000000000000000000000000000000000', '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 1, '', '2024-01-01 00:00:00.000', '2025-01-24 11:07:18.897'),

-- 5. Binance-Peg Bitcoin Token (BTCB)
(105, '0x7130d2A12B9BCbFAe4f2634d864A1Ee1Ce3Ead9c', 'bsc', 'BEP20', 'Binance-Peg Bitcoin Token', 'BTCB', 18, '0', 1, '["IERC20", "IERC20Metadata"]', '["transfer", "transferFrom", "approve", "allowance", "balanceOf", "totalSupply", "name", "symbol", "decimals"]', '["Transfer", "Approval"]', '{"name": "Binance-Peg Bitcoin Token", "symbol": "BTCB", "decimals": 18}', 1, 1, '0x0000000000000000000000000000000000000000', '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 1, '', '2024-01-01 00:00:00.000', '2025-01-24 11:07:18.897'),

-- 6. PancakeSwap Token (CAKE)
(106, '0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82', 'bsc', 'BEP20', 'PancakeSwap Token', 'CAKE', 18, '0', 1, '["IERC20", "IERC20Metadata"]', '["transfer", "transferFrom", "approve", "allowance", "balanceOf", "totalSupply", "name", "symbol", "decimals"]', '["Transfer", "Approval"]', '{"name": "PancakeSwap Token", "symbol": "CAKE", "decimals": 18}', 1, 1, '0x0000000000000000000000000000000000000000', '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 1, '', '2024-01-01 00:00:00.000', '2025-01-24 11:07:18.897'),

-- 7. Binance USD (BUSD)
(107, '0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56', 'bsc', 'BEP20', 'Binance USD', 'BUSD', 18, '0', 1, '["IERC20", "IERC20Metadata"]', '["transfer", "transferFrom", "approve", "allowance", "balanceOf", "totalSupply", "name", "symbol", "decimals"]', '["Transfer", "Approval"]', '{"name": "Binance USD", "symbol": "BUSD", "decimals": 18}', 1, 1, '0x0000000000000000000000000000000000000000', '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 1, '', '2024-01-01 00:00:00.000', '2025-01-24 11:07:18.897'),

-- 8. Binance-Peg XRP Token (XRP)
(108, '0x1D2F0da169ceB9Fc7C8C45F4FfE7A3bC8D4E7C3', 'bsc', 'BEP20', 'Binance-Peg XRP Token', 'XRP', 18, '0', 1, '["IERC20", "IERC20Metadata"]', '["transfer", "transferFrom", "approve", "allowance", "balanceOf", "totalSupply", "name", "symbol", "decimals"]', '["Transfer", "Approval"]', '{"name": "Binance-Peg XRP Token", "symbol": "XRP", "decimals": 18}', 1, 1, '0x0000000000000000000000000000000000000000', '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 1, '', '2024-01-01 00:00:00.000', '2025-01-24 11:07:18.897'),

-- 9. Binance-Peg Cardano Token (ADA)
(109, '0x3EE2200Efb3400fAbB9AacF31297cBdD1d435D47', 'bsc', 'BEP20', 'Binance-Peg Cardano Token', 'ADA', 18, '0', 1, '["IERC20", "IERC20Metadata"]', '["transfer", "transferFrom", "approve", "allowance", "balanceOf", "totalSupply", "name", "symbol", "decimals"]', '["Transfer", "Approval"]', '{"name": "Binance-Peg Cardano Token", "symbol": "ADA", "decimals": 18}', 1, 1, '0x0000000000000000000000000000000000000000', '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 1, '', '2024-01-01 00:00:00.000', '2025-01-24 11:07:18.897'),

-- 10. Binance-Peg Polkadot Token (DOT)
(110, '0x7083609fCE4d1d8Dc0C979AAb8c869Ea2C873402', 'bsc', 'BEP20', 'Binance-Peg Polkadot Token', 'DOT', 18, '0', 1, '["IERC20", "IERC20Metadata"]', '["transfer", "transferFrom", "approve", "allowance", "balanceOf", "totalSupply", "name", "symbol", "decimals"]', '["Transfer", "Approval"]', '{"name": "Binance-Peg Polkadot Token", "symbol": "DOT", "decimals": 18}', 1, 1, '0x0000000000000000000000000000000000000000', '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 1, '', '2024-01-01 00:00:00.000', '2025-01-24 11:07:18.897'),

-- 11. Binance-Peg Chainlink Token (LINK)
(111, '0xF8A0BF9cF54Bb92F17374d9e9A321E6a111a51bD', 'bsc', 'BEP20', 'Binance-Peg Chainlink Token', 'LINK', 18, '0', 1, '["IERC20", "IERC20Metadata"]', '["transfer", "transferFrom", "approve", "allowance", "balanceOf", "totalSupply", "name", "symbol", "decimals"]', '["Transfer", "Approval"]', '{"name": "Binance-Peg Chainlink Token", "symbol": "LINK", "decimals": 18}', 1, 1, '0x0000000000000000000000000000000000000000', '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 1, '', '2024-01-01 00:00:00.000', '2025-01-24 11:07:18.897'),

-- 12. Binance-Peg Litecoin Token (LTC)
(112, '0x4338665CBB7B2485A8855A139b75D5e34AB0DB94', 'bsc', 'BEP20', 'Binance-Peg Litecoin Token', 'LTC', 18, '0', 1, '["IERC20", "IERC20Metadata"]', '["transfer", "transferFrom", "approve", "allowance", "balanceOf", "totalSupply", "name", "symbol", "decimals"]', '["Transfer", "Approval"]', '{"name": "Binance-Peg Litecoin Token", "symbol": "LTC", "decimals": 18}', 1, 1, '0x0000000000000000000000000000000000000000', '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 1, '', '2024-01-01 00:00:00.000', '2025-01-24 11:07:18.897'),

-- 13. Binance-Peg Dogecoin Token (DOGE)
(113, '0xbA2aE424d960c26247Dd6c32edC70B295c744C43', 'bsc', 'BEP20', 'Binance-Peg Dogecoin Token', 'DOGE', 8, '0', 1, '["IERC20", "IERC20Metadata"]', '["transfer", "transferFrom", "approve", "allowance", "balanceOf", "totalSupply", "name", "symbol", "decimals"]', '["Transfer", "Approval"]', '{"name": "Binance-Peg Dogecoin Token", "symbol": "DOGE", "decimals": 8}', 1, 1, '0x0000000000000000000000000000000000000000', '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 1, '', '2024-01-01 00:00:00.000', '2025-01-24 11:07:18.897'),

-- 14. Binance-Peg Solana Token (SOL)
(114, '0x570A5D26f7765Ecb712C0924E4De545B89fD43dF', 'bsc', 'BEP20', 'Binance-Peg Solana Token', 'SOL', 18, '0', 1, '["IERC20", "IERC20Metadata"]', '["transfer", "transferFrom", "approve", "allowance", "balanceOf", "totalSupply", "name", "symbol", "decimals"]', '["Transfer", "Approval"]', '{"name": "Binance-Peg Solana Token", "symbol": "SOL", "decimals": 18}', 1, 1, '0x0000000000000000000000000000000000000000', '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', 1, '', '2024-01-01 00:00:00.000', '2025-01-24 11:07:18.897');
