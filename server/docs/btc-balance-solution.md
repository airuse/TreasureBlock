# BTC余额解决方案

## 问题背景

BTC地址余额计算面临以下挑战：
1. 从创世区块开始扫描成本极高
2. 新增地址需要重新扫描所有历史数据
3. 实时性要求与成本之间的平衡

## 解决方案

### 1. 混合策略

```
本地UTXO数据 + 外部API故障转移 + 增量扫描
```

#### 优先级顺序：
1. **本地UTXO数据** - 如果数据新鲜（<1000个区块）
2. **外部API故障转移** - 当本地数据过时或缺失时
3. **增量扫描** - 持续更新本地数据

### 2. 架构设计

#### 参考ETH故障转移模式
- **位置**: `server/internal/utils/btc_failover.go`
- **模式**: 与 `eth_failover.go` 完全一致
- **配置**: 直接使用 `rpc_urls` 配置
- **故障转移**: 自动轮询多个API URL
- **统一处理**: 所有API使用相同的HTTP请求模式

### 3. 支持的API提供商

#### BlockCypher API
- **优点**: 稳定、功能完整、支持UTXO
- **缺点**: 有请求限制
- **URL**: `https://api.blockcypher.com/v1/btc/main/addrs/{address}/balance`

#### Blockstream API
- **优点**: 开源、免费、无限制
- **缺点**: 功能相对简单
- **URL**: `https://blockstream.info/api/address/{address}`

#### Mempool.space API
- **优点**: 快速、现代化界面
- **缺点**: 相对较新
- **URL**: `https://mempool.space/api/address/{address}`

### 4. 配置示例

```yaml
# config.yaml
blockchain:
  chains:
    btc:
      chain_id: 1
      name: "Bitcoin"
      symbol: "BTC"
      decimals: 8
      enabled: true
      rpc_urls:
        - "http://127.0.0.1:8332"  # 本地BTC节点
        - "https://blockstream.info/api"  # Blockstream API (支持verbosity=3)
        - "https://mempool.space/api"  # Mempool API (支持verbosity=3)
      username: "bitcoin"
      password: "bitcoin123"
```

#### 配置说明
- **统一格式**: 所有API都支持 `verbosity=3` 参数
- **统一接口**: 所有API都使用 `/address/{address}?verbosity=3` 格式
- **统一解析**: 所有API返回相同格式的JSON响应
- **完全统一**: 代码中不需要任何URL判断或特殊处理

### 5. 使用方式

```go
// 在用户地址服务中自动使用故障转移
// 无需手动初始化，系统会自动创建BTC故障转移管理器

// 刷新余额时会自动使用混合策略
response, err := userAddressService.RefreshAddressBalances(userID, addressID)
```

#### 核心代码
```go
// server/internal/services/user_address_service.go
func (s *userAddressService) getBTCBalanceFromUTXO(address string) (uint64, string, error) {
    // ... 本地UTXO计算逻辑 ...
    
    // 如果本地数据过时，使用故障转移API
    if totalSatoshi == 0 || s.isUTXODataStale(maxHeight) {
        btcFailover, err := utils.NewBTCFailoverFromChain("btc")
        if err == nil {
            ctx := context.Background()
            apiResponse, apiErr := btcFailover.GetBalance(ctx, address)
            if apiErr == nil && apiResponse.Height > maxHeight {
                return apiResponse.Height, apiResponse.Balance, nil
            }
        }
    }
    
    return maxHeight, balanceStr, nil
}
```

#### 完全统一的故障转移实现
```go
// server/internal/utils/btc_failover.go
func (m *BTCFailoverManager) GetBalance(ctx context.Context, address string) (*BTCBalanceResponse, error) {
    return m.GetBalanceAtHeight(ctx, address, 0) // 0表示获取最新余额
}

// 支持指定高度的余额查询
func (m *BTCFailoverManager) GetBalanceAtHeight(ctx context.Context, address string, height uint64) (*BTCBalanceResponse, error) {
    var lastErr error
    deadline := time.Now().Add(m.timeout)
    
    for time.Now().Before(deadline) {
        url := m.next()  // 轮询下一个URL
        balance, err := m.getBalanceFromURL(ctx, url, address, height)
        if err == nil {
            return balance, nil
        }
        lastErr = err
    }
    
    return nil, fmt.Errorf("所有BTC API都获取失败: %w", lastErr)
}

// 统一的API调用，支持高度参数
func (m *BTCFailoverManager) getBalanceFromURL(ctx context.Context, baseURL, address string, height uint64) (*BTCBalanceResponse, error) {
    var url string
    if height > 0 {
        // 如果指定了高度，使用height参数
        url = fmt.Sprintf("%s/address/%s?verbosity=3&height=%d", baseURL, address, height)
    } else {
        // 如果高度为0，获取最新余额
        url = fmt.Sprintf("%s/address/%s?verbosity=3", baseURL, address)
    }
    
    // 统一的HTTP请求处理
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    // ... 统一的请求和响应处理
}
```

### 6. 性能优化

#### 缓存策略
```go
// 添加Redis缓存
type cachedBTCBalanceService struct {
    btcBalanceService BTCBalanceService
    redisClient       *redis.Client
    cacheTTL          time.Duration
}

func (s *cachedBTCBalanceService) GetBalance(address string) (string, uint64, error) {
    // 1. 检查缓存
    if cached := s.getFromCache(address); cached != nil {
        return cached.Balance, cached.Height, nil
    }
    
    // 2. 调用API
    balance, height, err := s.btcBalanceService.GetBalance(address)
    if err != nil {
        return "0", 0, err
    }
    
    // 3. 写入缓存
    s.setCache(address, balance, height)
    
    return balance, height, nil
}
```

#### 批量处理
```go
// 批量获取多个地址余额
func (s *btcBalanceService) GetBalancesBatch(addresses []string) (map[string]BalanceInfo, error) {
    // 并行调用多个API
    // 合并结果
    // 返回批量数据
}
```

### 6. 监控和告警

```go
// 添加监控指标
type metrics struct {
    apiCallsTotal    prometheus.Counter
    apiErrorsTotal   prometheus.Counter
    cacheHitRate     prometheus.Gauge
    responseTime     prometheus.Histogram
}

// 监控API调用成功率
func (s *btcBalanceService) GetBalance(address string) (string, uint64, error) {
    start := time.Now()
    defer func() {
        s.metrics.responseTime.Observe(time.Since(start).Seconds())
    }()
    
    // API调用逻辑
}
```

### 7. 故障处理

#### 降级策略
1. **API1失败** → 尝试API2
2. **所有API失败** → 使用本地数据
3. **本地数据过时** → 返回错误或使用缓存

#### 重试机制
```go
func (s *btcBalanceService) getBalanceWithRetry(address string) (string, uint64, error) {
    for i := 0; i < 3; i++ {
        balance, height, err := s.getBalanceFromAPI(address)
        if err == nil {
            return balance, height, nil
        }
        
        if i < 2 {
            time.Sleep(time.Duration(i+1) * time.Second)
        }
    }
    
    return "0", 0, fmt.Errorf("重试3次后仍然失败")
}
```

## 总结

这个解决方案提供了：
- ✅ **成本效益**: 避免全量扫描历史数据
- ✅ **实时性**: 通过API获取最新余额
- ✅ **可靠性**: 多API故障转移
- ✅ **可扩展性**: 支持添加更多API提供商
- ✅ **性能**: 缓存和批量处理优化

适用于大多数BTC应用场景，平衡了成本、性能和可靠性。
