package constants

// ChainType 链类型
type ChainType string

const (
	// ChainBTC 比特币链
	ChainBTC ChainType = "btc"
	// ChainETH 以太坊链
	ChainETH ChainType = "eth"
	// ChainBSC BSC链
	ChainBSC ChainType = "bsc"
	// ChainSOL Solana链
	ChainSOL ChainType = "sol"
)

// IsValid 检查链类型是否有效
func (c ChainType) IsValid() bool {
	return c == ChainBTC || c == ChainETH || c == ChainBSC || c == ChainSOL
}

// String 返回链类型的字符串表示
func (c ChainType) String() string {
	return string(c)
}

// GetChainTypes 获取所有有效的链类型
func GetChainTypes() []ChainType {
	return []ChainType{ChainBTC, ChainETH, ChainBSC, ChainSOL}
}

// IsBTC 检查是否为BTC链
func (c ChainType) IsBTC() bool {
	return c == ChainBTC
}

// IsETH 检查是否为ETH链
func (c ChainType) IsETH() bool {
	return c == ChainETH
}

// IsBSC 检查是否为BSC链
func (c ChainType) IsBSC() bool {
	return c == ChainBSC
}

// IsSOL 检查是否为SOL链
func (c ChainType) IsSOL() bool {
	return c == ChainSOL
}
