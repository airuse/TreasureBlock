package constants

// ChainType 链类型
type ChainType string

const (
	// ChainBTC 比特币链
	ChainBTC ChainType = "btc"
	// ChainETH 以太坊链
	ChainETH ChainType = "eth"
)

// IsValid 检查链类型是否有效
func (c ChainType) IsValid() bool {
	return c == ChainBTC || c == ChainETH
}

// String 返回链类型的字符串表示
func (c ChainType) String() string {
	return string(c)
}

// GetChainTypes 获取所有有效的链类型
func GetChainTypes() []ChainType {
	return []ChainType{ChainBTC, ChainETH}
}

// IsBTC 检查是否为BTC链
func (c ChainType) IsBTC() bool {
	return c == ChainBTC
}

// IsETH 检查是否为ETH链
func (c ChainType) IsETH() bool {
	return c == ChainETH
}
