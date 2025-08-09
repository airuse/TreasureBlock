# Blockchain Browser Backend Server

这是一个用Go语言实现的区块链浏览器后端服务，提供RESTful API和WebSocket支持。

## 项目结构

```
server/
├── main.go                 # 主程序入口
├── go.mod                  # Go模块文件
├── config.yaml             # 主配置文件
├── config.yaml.example     # 配置文件示例
├── config/                 # 配置管理
│   └── config.go
├── internal/               # 内部包
│   ├── database/           # 数据库连接
│   │   └── database.go
│   ├── models/             # 数据模型
│   │   ├── block.go
│   │   └── transaction.go
│   ├── repository/         # 数据访问层
│   │   ├── block_repository.go
│   │   └── transaction_repository.go
│   ├── services/           # 业务逻辑层
│   │   ├── block_service.go
│   │   └── transaction_service.go
│   ├── handlers/           # HTTP处理器
│   │   ├── block_handler.go
│   │   ├── transaction_handler.go
│   │   └── websocket_handler.go
│   ├── routes/             # 路由配置
│   │   └── routes.go
│   ├── middleware/         # 中间件
│   │   └── middleware.go
│   ├── utils/              # 工具函数
│   │   ├── response.go
│   │   └── mock_data.go
│   └── server/             # 服务器主文件
│       └── server.go
└── README.md
```

## 功能特性

### 已实现功能
- ✅ 查询区块接口
- ✅ 查询交易接口
- ✅ WebSocket实时通信
- ✅ 分页查询支持
- ✅ 多链支持（BTC、ETH等）
- ✅ 统一错误处理
- ✅ CORS支持
- ✅ YAML配置文件支持

### 待实现功能
- 🔄 更新/新增区块接口
- 🔄 更新/新增交易接口
- 🔄 组建未签名交易接口
- 🔄 发送交易接口
- 🔄 文件上传处理

## API接口

### 区块相关接口

#### 1. 获取区块列表
```
GET /api/v1/blocks?page=1&page_size=20&chain=btc
```

#### 2. 获取最新区块
```
GET /api/v1/blocks/latest?chain=btc
```

#### 3. 根据哈希获取区块
```
GET /api/v1/blocks/hash/{hash}
```

#### 4. 根据高度获取区块
```
GET /api/v1/blocks/height/{height}
```

### 交易相关接口

#### 1. 获取交易列表
```
GET /api/v1/transactions?page=1&page_size=20&chain=btc
```

#### 2. 根据哈希获取交易
```
GET /api/v1/transactions/hash/{hash}
```

#### 3. 根据地址获取交易
```
GET /api/v1/transactions/address/{address}?page=1&page_size=20
```

#### 4. 根据区块哈希获取交易
```
GET /api/v1/transactions/block/{blockHash}
```

### WebSocket接口

#### WebSocket连接
```
WS /ws
```

支持的消息类型：
- `ping`: 心跳检测
- `subscribe`: 订阅频道

## 配置

项目使用YAML配置文件进行配置管理，支持以下配置项：

### 配置文件位置
项目会按以下优先级查找配置文件：
1. `config/config.yaml`
2. `config/config.yml`
3. `./config/config.yaml`
4. `./config/config.yml`
5. `../config/config.yaml`
6. `../config/config.yml`

### 配置示例
```yaml
server:
  host: "localhost"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  max_connections: 1000

database:
  driver: "sqlite"
  host: "localhost"
  port: 3306
  username: ""
  password: ""
  dbname: "blockchain.db"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600s

log:
  level: "info"
  format: "json"
  output: "stdout"

websocket:
  enabled: true
  path: "/ws"
  ping_interval: 30s
  pong_wait: 60s

cors:
  allowed_origins: ["*"]
  allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  allow_credentials: true

api:
  version: "v1"
  prefix: "/api"
  rate_limit:
    enabled: true
    requests_per_minute: 100
    burst: 20

blockchain:
  chains:
    btc:
      name: "Bitcoin"
      symbol: "BTC"
      decimals: 8
      enabled: true
    eth:
      name: "Ethereum"
      symbol: "ETH"
      decimals: 18
      enabled: true
```

### 环境变量回退
如果YAML配置文件加载失败，系统会自动回退到环境变量配置，支持的环境变量包括：
- `SERVER_HOST`, `SERVER_PORT`
- `DB_DRIVER`, `DB_HOST`, `DB_PORT`, `DB_NAME`
- `LOG_LEVEL`
- `WS_ENABLED`, `WS_PATH`
- 等等

## 运行

### 1. 安装依赖
```bash
go mod tidy
```

### 2. 配置
复制配置文件示例并修改：
```bash
cp config.yaml.example config.yaml
# 根据需要修改 config.yaml
```

### 3. 运行服务器
```bash
go run main.go
```

### 4. 访问API
- 健康检查: http://localhost:8080/health
- API文档: http://localhost:8080/api/v1/blocks

## 数据库

项目使用SQLite作为默认数据库，支持自动迁移。数据模型包括：

### Block表
- id: 主键
- hash: 区块哈希（唯一索引）
- height: 区块高度（唯一索引）
- previous_hash: 前一个区块哈希
- merkle_root: Merkle根
- timestamp: 时间戳
- difficulty: 难度
- nonce: 随机数
- size: 区块大小
- transaction_count: 交易数量
- total_amount: 总金额
- fee: 手续费
- confirmations: 确认数
- is_orphan: 是否孤块
- chain: 链类型
- created_at: 创建时间
- updated_at: 更新时间

### Transaction表
- id: 主键
- hash: 交易哈希（唯一索引）
- block_hash: 区块哈希
- block_height: 区块高度
- from_address: 发送地址
- to_address: 接收地址
- amount: 金额
- fee: 手续费
- gas_price: Gas价格
- gas_limit: Gas限制
- gas_used: 使用的Gas
- nonce: 随机数
- status: 状态
- confirmations: 确认数
- timestamp: 时间戳
- input_data: 输入数据
- contract_address: 合约地址
- chain: 链类型
- created_at: 创建时间
- updated_at: 更新时间

## 开发规范

### 代码结构
- 遵循SOLID原则
- 使用依赖注入
- 接口与实现分离
- 统一的错误处理
- 完整的日志记录

### 配置管理
- 使用YAML配置文件
- 支持环境变量回退
- 配置验证和默认值
- 多环境配置支持

### 测试
- 单元测试覆盖核心业务逻辑
- 集成测试验证API接口
- Mock数据用于测试

### 性能优化
- 数据库索引优化
- 分页查询
- 缓存策略（待实现）
- 连接池管理

## 扩展计划

1. **缓存层**: 添加Redis缓存
2. **监控**: 集成Prometheus监控
3. **日志**: 结构化日志和ELK集成
4. **安全**: JWT认证和权限控制
5. **文档**: Swagger API文档
6. **测试**: 完整的测试套件
7. **部署**: Docker容器化
8. **CI/CD**: 自动化部署流程 