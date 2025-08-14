# Blockchain Block Scanner

一个强大的区块链区块扫描工具，支持多种区块链网络。

## 功能特性

- **多链支持**: 支持比特币、以太坊等多种区块链
- **可配置扫描**: 可配置扫描间隔、批量大小、重试机制
- **自动重试**: 网络错误时自动重试
- **进度跟踪**: 实时跟踪扫描进度
- **文件输出**: 支持将区块数据保存到文件
- **服务器提交**: 自动提交区块数据到服务器
- **并发扫描**: 支持多链并发扫描

## 项目结构

```
scanner/
├── cmd/
│   └── main.go                    # 主程序入口
├── config/
│   ├── config.go                  # 配置管理
│   └── config.yaml                # 配置文件
├── internal/
│   ├── models/
│   │   └── block.go               # 区块模型
│   ├── scanners/
│   │   ├── bitcoin_scanner.go     # 比特币扫块器
│   │   └── ethereum_scanner.go    # 以太坊扫块器
│   └── scanner/
│       └── block_scanner.go       # 主扫块器
├── go.mod                         # Go模块文件
└── README.md                      # 说明文档
```

## 安装和运行

### 1. 安装依赖

```bash
cd scanner
go mod tidy
```

### 2. 配置

复制并修改配置文件：

```bash
cp config/config.yaml config/config.local.yaml
```

编辑 `config/config.local.yaml` 文件，设置相关配置：

```yaml
server:
  host: "localhost"
  port: 8080
  protocol: "http"
  api_key: "your-api-key"

blockchain:
  chains:
    btc:
      name: "Bitcoin"
      symbol: "BTC"
      decimals: 8
      enabled: true
      rpc_url: "http://localhost:8332"
      explorer_api_url: "https://blockstream.info/api"
      api_key: ""
    eth:
      name: "Ethereum"
      symbol: "ETH"
      decimals: 18
      enabled: true
      rpc_url: "http://localhost:8545"
      explorer_api_url: "https://api.etherscan.io/api"
      api_key: "your-etherscan-api-key"

scan:
  enabled: true
  interval: 10s
  batch_size: 100
  max_retries: 3
  retry_delay: 5s
  confirmations: 6
  start_block_height: 0
  end_block_height: 0
  auto_start: true
  save_to_file: false
  output_dir: "./output"
```

### 3. 运行

```bash
# 使用默认配置运行
go run cmd/main.go

# 指定配置文件
go run cmd/main.go --config config/config.local.yaml

# 指定扫描参数
go run cmd/main.go --start 1000000 --end 1000100 --chain btc

# 指定扫描间隔
go run cmd/main.go --interval 30s
```

## 命令行参数

- `--config, -c`: 配置文件路径
- `--start, -s`: 起始区块高度
- `--end, -e`: 结束区块高度
- `--chain, -n`: 指定链名称 (btc, eth)
- `--interval, -i`: 扫描间隔

## 配置说明

### 服务器配置

- `host`: 目标服务器地址
- `port`: 目标服务器端口
- `protocol`: 协议类型 (http/https)
- `api_key`: API密钥

### 区块链配置

- `name`: 链名称
- `symbol`: 链符号
- `decimals`: 小数位数
- `enabled`: 是否启用
- `rpc_url`: RPC节点地址
- `explorer_api_url`: 区块浏览器API地址
- `api_key`: API密钥

### 扫描配置

- `enabled`: 是否启用扫描
- `interval`: 扫描间隔
- `batch_size`: 批量大小
- `max_retries`: 最大重试次数
- `retry_delay`: 重试延迟
- `confirmations`: 确认数
- `start_block_height`: 起始区块高度
- `end_block_height`: 结束区块高度
- `auto_start`: 是否自动启动
- `save_to_file`: 是否保存到文件
- `output_dir`: 输出目录

## 使用示例

### 1. 扫描比特币区块

```bash
# 扫描最新的100个区块
go run cmd/main.go --chain btc --start 800000 --end 800100

# 持续扫描新区块
go run cmd/main.go --chain btc --interval 30s
```

### 2. 扫描以太坊区块

```bash
# 扫描指定范围的区块
go run cmd/main.go --chain eth --start 18000000 --end 18000100

# 保存到文件
go run cmd/main.go --chain eth --save-to-file --output-dir ./eth-blocks
```

### 3. 多链扫描

```bash
# 同时扫描比特币和以太坊
go run cmd/main.go --interval 60s
```

## 输出格式

### 控制台输出

```json
{
  "level": "info",
  "msg": "Successfully processed block 800000 for chain btc",
  "time": "2024-01-01T00:00:00Z"
}
```

### 文件输出

如果启用了文件输出，区块数据会保存为JSON格式：

```json
{
  "hash": "0000000000000000000000000000000000000000000000000000000000000000",
  "height": 800000,
  "previous_hash": "0000000000000000000000000000000000000000000000000000000000000000",
  "merkle_root": "0000000000000000000000000000000000000000000000000000000000000000",
  "timestamp": "2024-01-01T00:00:00Z",
  "difficulty": 1.0,
  "nonce": 0,
  "size": 1000,
  "transaction_count": 100,
  "total_amount": 1000.0,
  "fee": 0.001,
  "confirmations": 6,
  "is_orphan": false,
  "chain": "btc",
  "version": 1,
  "bits": "1d00ffff",
  "weight": 4000000,
  "stripped_size": 1000
}
```

## 开发

### 构建

```bash
go build -o block-scanner cmd/main.go
```

### 测试

```bash
go test ./...
```

### 代码格式化

```bash
go fmt ./...
```

## 注意事项

1. **API限制**: 请注意各种API的调用频率限制
2. **网络连接**: 确保能够访问区块链节点和区块浏览器API
3. **存储空间**: 如果启用文件输出，请确保有足够的存储空间
4. **资源消耗**: 扫描会消耗一定的网络和计算资源

## 故障排除

### 常见问题

1. **API调用失败**
   - 检查网络连接
   - 验证API密钥是否正确
   - 确认API调用频率是否超限

2. **扫描速度慢**
   - 调整扫描间隔
   - 减少批量大小
   - 检查网络延迟

3. **内存使用过高**
   - 减少批量大小
   - 调整扫描间隔
   - 检查是否有内存泄漏

## 许可证

本项目采用 MIT 许可证。


## btc/eth启动命令
bitcoind.exe -server=1 -rpcuser=bitcoin -rpcpassword=bitcoin123 -rpcallowip=0.0.0.0/0 -rpcbind=0.0.0.0 -rpcport=8332 -listen=1 -bind=0.0.0.0 -port=8333 -prune=550 -maxmempool=100 -dbcache=1000 -maxconnections=10 -blocksonly=1 -datadir="%APPDATA%\Bitcoin"

geth --syncmode=snap --cache=1024 --datadir=C:\ethereum-data --http --http.addr=192.168.1.70 --http.port=8545 --http.api="eth,net,web3" --maxpeers=15 --maxpendpeers=5

