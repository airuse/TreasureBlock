-- 初始化权限类型到baseconfig字典表
-- 用于API密钥的权限范围配置
-- 只包含：上传区块、上传交易 两个权限类型

INSERT INTO base_config (`group`, no, config_type, config_name, config_key, config_value, description) VALUES
('api_permissions', 1, 1, '上传区块权限', 'blocks_write', 'blocks:write', '上传区块数据权限'),
('api_permissions', 2, 1, '上传交易权限', 'transactions_write', 'transactions:write', '上传交易数据权限');

-- 重新设计coin_config表结构（删除旧表并创建新表）
-- 删除旧的coin_config表（如果存在）
DROP TABLE IF EXISTS coin_config;

-- 创建新的coin_config表
CREATE TABLE coin_config (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    chain_name VARCHAR(20) NOT NULL COMMENT '链名称(eth, btc, polygon等)',
    symbol VARCHAR(20) NOT NULL COMMENT '币种符号(ETH, USDT, USDC等)',
    coin_type TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '币种类型: 0=原生币, 1=ERC20, 2=ERC223, 3=ERC721, 4=ERC1155',
    contract_addr VARCHAR(120) NOT NULL DEFAULT '' COMMENT '合约地址(原生币为空)',
    precision TINYINT UNSIGNED NOT NULL DEFAULT 18 COMMENT '精度(小数位数)',
    decimals TINYINT UNSIGNED NOT NULL DEFAULT 18 COMMENT '精度别名(兼容性)',
    name VARCHAR(100) NOT NULL COMMENT '币种全名',
    logo_url VARCHAR(255) DEFAULT '' COMMENT '币种Logo URL',
    website_url VARCHAR(255) DEFAULT '' COMMENT '官方网站',
    explorer_url VARCHAR(255) DEFAULT '' COMMENT '区块浏览器地址',
    description TEXT COMMENT '币种描述',
    market_cap_rank INT UNSIGNED DEFAULT 0 COMMENT '市值排名',
    is_stablecoin BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否为稳定币',
    is_verified BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否已验证',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 0=禁用, 1=启用',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    -- 索引
    INDEX idx_chain_symbol (chain_name, symbol),
    INDEX idx_contract_addr (contract_addr),
    INDEX idx_status (status),
    INDEX idx_chain_status (chain_name, status),
    UNIQUE KEY uk_chain_contract (chain_name, contract_addr),
    UNIQUE KEY uk_chain_symbol (chain_name, symbol)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='币种配置表';

-- 初始化主流代币配置到coin_config表
-- 包含市面上最常见的ERC-20代币（仅限以太坊主网）

INSERT INTO `coin_config` (`chain_name`, `symbol`, `coin_type`, `contract_addr`, `precision`, `decimals`, `name`, `logo_url`, `website_url`, `explorer_url`, `description`, `market_cap_rank`, `is_stablecoin`, `is_verified`, `status`) VALUES
('eth', 'ETH', 0, '', 18, 18, 'Ethereum', 'https://cryptologos.cc/logos/ethereum-eth-logo.png', 'https://ethereum.org', 'https://etherscan.io', '以太坊是去中心化的开源区块链平台，支持智能合约和去中心化应用', 2, FALSE, TRUE, 1),
('eth', 'USDT', 1, '0xdAC17F958D2ee523a2206206994597C13D831ec7', 6, 6, 'Tether USD', 'https://cryptologos.cc/logos/tether-usdt-logo.png', 'https://tether.to', 'https://etherscan.io/token/0xdac17f958d2ee523a2206206994597c13d831ec7', 'Tether USD是一种与美元挂钩的稳定币，由Tether Limited发行', 3, TRUE, TRUE, 1),
('eth', 'USDC', 1, '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48', 6, 6, 'USD Coin', 'https://cryptologos.cc/logos/usd-coin-usdc-logo.png', 'https://www.circle.com/en/usdc', 'https://etherscan.io/token/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48', 'USD Coin是由Circle和Coinbase共同发行的与美元挂钩的稳定币', 6, TRUE, TRUE, 1),
('eth', 'DAI', 1, '0x6B175474E89094C44Da98b954EedeAC495271d0F', 18, 18, 'Dai Stablecoin', 'https://cryptologos.cc/logos/multi-collateral-dai-dai-logo.png', 'https://makerdao.com', 'https://etherscan.io/token/0x6b175474e89094c44da98b954eedeac495271d0f', 'Dai是由MakerDAO发行的去中心化稳定币，通过超额抵押机制维持与美元的挂钩', 15, TRUE, TRUE, 1),
('eth', 'UNI', 1, '0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984', 18, 18, 'Uniswap', 'https://cryptologos.cc/logos/uniswap-uni-logo.png', 'https://uniswap.org', 'https://etherscan.io/token/0x1f9840a85d5af5bf1d1762f925bdaddc4201f984', 'Uniswap是以太坊上最大的去中心化交易所协议，UNI是其治理代币', 20, FALSE, TRUE, 1),
('eth', 'LINK', 1, '0x514910771AF9Ca656af840dff83E8264EcF986CA', 18, 18, 'Chainlink', 'https://cryptologos.cc/logos/chainlink-link-logo.png', 'https://chainlink.link', 'https://etherscan.io/token/0x514910771af9ca656af840dff83e8264ecf986ca', 'Chainlink是去中心化的预言机网络，为智能合约提供真实世界数据', 12, FALSE, TRUE, 1),
('eth', 'WETH', 1, '0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2', 18, 18, 'Wrapped Ether', 'https://cryptologos.cc/logos/weth-logo.png', 'https://weth.io', 'https://etherscan.io/token/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2', 'WETH是以太坊的ERC-20包装版本，用于DeFi协议中的流动性提供', 0, FALSE, TRUE, 1),
('eth', 'WBTC', 1, '0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599', 8, 8, 'Wrapped Bitcoin', 'https://cryptologos.cc/logos/wrapped-bitcoin-wbtc-logo.png', 'https://www.wbtc.network', 'https://etherscan.io/token/0x2260fac5e5542a773aa44fbcfedf7c193bc2c599', 'WBTC是以太坊上的比特币包装版本，1:1锚定比特币', 0, FALSE, TRUE, 1),
('eth', 'AAVE', 1, '0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DDaE9', 18, 18, 'Aave', 'https://cryptologos.cc/logos/aave-aave-logo.png', 'https://aave.com', 'https://etherscan.io/token/0x7fc66500c84a76ad7e9c93437bfc5ac33e2ddae9', 'Aave是以太坊上的去中心化借贷协议，AAVE是其治理代币', 25, FALSE, TRUE, 1),
('eth', 'CRV', 1, '0xD533a949740bb3306d119CC777fa900bA034cd52', 18, 18, 'Curve DAO Token', 'https://cryptologos.cc/logos/curve-dao-token-crv-logo.png', 'https://curve.fi', 'https://etherscan.io/token/0xd533a949740bb3306d119cc777fa900ba034cd52', 'Curve是专注于稳定币交易的去中心化交易所，CRV是其治理代币', 30, FALSE, TRUE, 1),
('eth', 'COMP', 1, '0xc00e94Cb662C3520282E6f5717214004A7f26888', 18, 18, 'Compound', 'https://cryptologos.cc/logos/compound-comp-logo.png', 'https://compound.finance', 'https://etherscan.io/token/0xc00e94cb663c68f0b0f0b0f0b0f0b0f0b0f0b0b', 'Compound是以太坊上的去中心化借贷协议，COMP是其治理代币', 35, FALSE, TRUE, 1),
('eth', 'MKR', 1, '0x9f8F72aA9304c8B593d555F12eF6589cC3A579A2', 18, 18, 'Maker', 'https://cryptologos.cc/logos/maker-mkr-logo.png', 'https://makerdao.com', 'https://etherscan.io/token/0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2', 'MakerDAO是以太坊上的去中心化自治组织，MKR是其治理代币', 40, FALSE, TRUE, 1),
('eth', 'SNX', 1, '0xC011a73ee8576Fb46F5E1c5751cA3B9Fe0af2a6F', 18, 18, 'Synthetix', 'https://cryptologos.cc/logos/synthetix-network-token-snx-logo.png', 'https://synthetix.io', 'https://etherscan.io/token/0xc011a73ee8576fb46f5e1c5751ca3b9fe0af2a6f', 'Synthetix是以太坊上的合成资产协议，SNX是其治理代币', 45, FALSE, TRUE, 1),
('eth', '1INCH', 1, '0x111111111117dC0aa78b770fA6A738034120C302', 18, 18, '1inch', 'https://cryptologos.cc/logos/1inch-1inch-logo.png', 'https://1inch.io', 'https://etherscan.io/token/0x111111111117dc0aa78b770fa6a738034120c302', '1inch是以太坊上的DEX聚合器，为用户提供最优的交易路径', 50, FALSE, TRUE, 1),
('eth', 'SUSHI', 1, '0x6B3595068778DD592e39A122f4f5a5cF09C90fE2', 18, 18, 'SushiSwap', 'https://cryptologos.cc/logos/sushiswap-sushi-logo.png', 'https://sushiswap.fi', 'https://etherscan.io/token/0x6b3595068778dd592e39a122f4f5a5cf09c90fe2', 'SushiSwap是以太坊上的去中心化交易所，SUSHI是其治理代币', 55, FALSE, TRUE, 1),
('eth', 'YFI', 1, '0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e', 18, 18, 'yearn.finance', 'https://cryptologos.cc/logos/yearn-finance-yfi-logo.png', 'https://yearn.finance', 'https://etherscan.io/token/0x0bc529c00c6401aef6d220be8c6ea1667f6ad93e', 'yearn.finance是以太坊上的收益聚合器，YFI是其治理代币', 60, FALSE, TRUE, 1),
('eth', 'BAL', 1, '0xba100000625a3754423978a60c9317c58a424e3D', 18, 18, 'Balancer', 'https://cryptologos.cc/logos/balancer-bal-logo.png', 'https://balancer.fi', 'https://etherscan.io/token/0xba100000625a3754423978a60c9317c58a424e3d', 'Balancer是以太坊上的自动做市商协议，BAL是其治理代币', 65, FALSE, TRUE, 1),
('eth', 'KNC', 1, '0xdd974D5C2e2928deA5F71b9825b8b646686BD200', 18, 18, 'Kyber Network', 'https://cryptologos.cc/logos/kyber-network-crystal-knc-logo.png', 'https://kyber.network', 'https://etherscan.io/token/0xdd974d5c2e2928dea5f71b9825b8b646686bd200', 'Kyber Network是以太坊上的去中心化交易所协议，KNC是其治理代币', 75, FALSE, TRUE, 1),
('eth', 'ZRX', 1, '0xE41d2489571d322189246DaFA5ebDe1F4699F498', 18, 18, '0x Protocol', 'https://cryptologos.cc/logos/0x-zrx-logo.png', 'https://0x.org', 'https://etherscan.io/token/0xe41d2489571d322189246dafa5ebde1f4699f498', '0x Protocol是以太坊上的去中心化交易所协议，ZRX是其治理代币', 80, FALSE, TRUE, 1),
('eth', 'BAND', 1, '0xBA11D00c5f74255f56a5E366F4F77f5A186d7f55', 18, 18, 'Band Protocol', 'https://cryptologos.cc/logos/band-protocol-band-logo.png', 'https://bandprotocol.com', 'https://etherscan.io/token/0xba11d00c5f74255f56a5e366f4f77f5a186d7f55', 'Band Protocol是以太坊上的预言机网络，BAND是其治理代币', 85, FALSE, TRUE, 1),
('eth', 'ENJ', 1, '0xF629cBd94d3791C9250152BD8dfBDF380E2a3B9c', 18, 18, 'Enjin Coin', 'https://cryptologos.cc/logos/enjin-coin-enj-logo.png', 'https://enjin.io', 'https://etherscan.io/token/0xf629cbd94d3791c9250152bd8dfbdf380e2a3b9c', 'Enjin是以太坊上的游戏平台，ENJ是其治理代币', 90, FALSE, TRUE, 1),
('eth', 'MANA', 1, '0x0F5D2fB29fb7d3CFeE444a200298f468908cC942', 18, 18, 'Decentraland', 'https://cryptologos.cc/logos/decentraland-mana-logo.png', 'https://decentraland.org', 'https://etherscan.io/token/0x0f5d2fb29fb7d3cfee444a200298f468908cc942', 'Decentraland是以太坊上的虚拟世界平台，MANA是其治理代币', 95, FALSE, TRUE, 1),
('eth', 'SAND', 1, '0x3845badAde8e6dFF049820680d1F14bD3903a5d0', 18, 18, 'The Sandbox', 'https://cryptologos.cc/logos/the-sandbox-sand-logo.png', 'https://www.sandbox.game', 'https://etherscan.io/token/0x3845badade8e6dff049820680d1f14bd3903a5d0', 'The Sandbox是以太坊上的虚拟世界平台，SAND是其治理代币', 100, FALSE, TRUE, 1),
('eth', 'AXS', 1, '0xBB0E17EF65F82Ab018d8EDd776e8DD940327B28b', 18, 18, 'Axie Infinity', 'https://cryptologos.cc/logos/axie-infinity-axs-logo.png', 'https://axieinfinity.com', 'https://etherscan.io/token/0xbb0e17ef65f82ab018d8edd776e8dd940327b28b', 'Axie Infinity是以太坊上的游戏平台，AXS是其治理代币', 105, FALSE, TRUE, 1),
('eth', 'CHZ', 1, '0x3506424F91fD33084466F402d5D97f05F8e3b4AF', 18, 18, 'Chiliz', 'https://cryptologos.cc/logos/chiliz-chz-logo.png', 'https://www.chiliz.com', 'https://etherscan.io/token/0x3506424f91fd33084466f402d5d97f05f8e3b4af', 'Chiliz是以太坊上的体育娱乐平台，CHZ是其治理代币', 110, FALSE, TRUE, 1);

