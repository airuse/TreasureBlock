# Treasure Block - 区块链浏览器

一个现代化的区块链浏览器项目，支持比特币和以太坊网络的实时数据查询、统计展示与离线签名。

## 📖 项目简介

Treasure Block 是一个全栈区块链浏览器解决方案，包含：
- **区块链扫描器** (Go) - 实时扫描和收集区块链数据
- **API 服务端** (Go) - 提供 RESTful API 和 WebSocket 服务
- **前端界面** (Vue.js) - 现代化的用户界面，支持实时数据展示

## 🏗️ 项目架构

```
blockChainBrowser/
├── client/
│   ├── scanner/      # 区块链数据扫描器（Go）
│   └── signer/       # 离线交易签名器（Go，支持 ETH/BTC）
├── server/           # API 服务端（Go + Gin + GORM）
├── vue/              # 前端应用（Vue 3 + TS + Vite + Tailwind）
├── contract/         # 合约源码（Solidity）
├── scripts/          # 证书脚本与辅助脚本
└── docs/             # 详细文档
```

## ✨ 主要功能

### 🔍 扫描器功能
- ✅ 比特币 (Bitcoin) 网络扫描
- ✅ 以太坊 (Ethereum) 网络扫描
- ✅ 实时区块数据收集
- ✅ 交易信息解析
- ✅ 地址余额追踪
 - ✅ 批量上传交易数据（提升吞吐）

### 🚀 API 服务
- ✅ RESTful API 接口
- ✅ WebSocket 实时推送
- ✅ 区块数据查询
- ✅ 交易记录查询
- ✅ 地址信息查询
- ✅ 资产统计分析
- ✅ 用户认证和API密钥管理
 - ✅ 多链支持（BTC、ETH）

### 💎 前端界面
- ✅ 响应式设计
- ✅ 实时数据展示
- ✅ 区块浏览
- ✅ 交易查询
- ✅ 地址搜索
- ✅ 统计图表

### ✍️ 离线签名器（可选）
- ✅ 离线签名 ETH/BTC 交易
- ✅ 私钥加密存储与密码保护
- ✅ 支持 QR 码导入/导出

## 🛠️ 技术栈

### 后端
- **Go** - 高性能后端开发
- **Gin** - Web 框架
- **GORM** - ORM 数据库操作
- **WebSocket** - 实时通信
- **MySQL/PostgreSQL** - 数据存储
- **JWT** - 用户认证
- **Bcrypt** - 密码加密

### 前端
- **Vue.js 3** - 前端框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **Tailwind CSS** - 样式框架
- **WebSocket** - 实时数据

## 🚀 快速开始

### 环境要求
- Go 1.20+
- Node.js 18+
- MySQL 8.0+ 或 PostgreSQL 13+

### 1. 克隆项目
```bash
git clone <your_repo_url>
cd blockChainBrowser
```

### 2. 启动数据库服务
确保 MySQL 或 PostgreSQL 服务正在运行

### 3. 配置扫描器（client/scanner）
```bash
cd client/scanner
cp config.yaml.example config.yaml
# 编辑 config.yaml 配置文件
go mod tidy
```

### 4. 启动扫描器
```bash
cd client/scanner
go run cmd/main.go
# 或者
make build && ./main
```

### 5. 配置 API 服务（server）
```bash
cd server
cp config.yaml.example config.yaml
# 编辑 config.yaml 配置文件
go mod tidy
```

### 6. 启动 API 服务
```bash
cd server
go run main.go
```

可选：启用 TLS 服务（推荐生产环境）
```bash
# 生成本地域名证书（自签）
cd server/scripts
bash generate-domain-cert.sh
# 或生成通用 TLS 证书
bash generate-tls-cert.sh
```

### 7. 启动前端应用（vue）
```bash
cd vue
npm install
npm run dev
```

### 8. 访问应用
- 前端界面: http://localhost:5173
- 健康检查: http://localhost:8080/health
- WebSocket: ws://localhost:8080/ws

## 📝 配置说明

### 扫描器配置 (client/scanner/config.yaml)
```yaml
scanner:
  interval: 10s
  bitcoin:
    rpc_url: "http://localhost:8332"
    rpc_user: "bitcoin"
    rpc_password: "password"
  ethereum:
    rpc_url: "http://localhost:8545"
    
database:
  host: localhost
  port: 3306
  user: root
  password: password
  dbname: blockchain_browser
```

### 服务端配置 (server/config.yaml)
```yaml
server:
  port: 8080
  mode: debug
  tls_enabled: true
  tls_port: 8443

database:
  host: localhost
  port: 3306
  user: root
  password: password
  dbname: blockchain_browser
```

### 签名器（可选，client/signer）
```bash
cd client/signer
go mod tidy
go build -o signer cmd/main.go
./signer
```
默认密码为 "hello"。支持 QR 码导入、ETH/BTC 交易离线签名与结果导出。

## 🔧 开发指南

### 代码规范
- 遵循 Go 官方代码规范
- 使用 gofmt 格式化代码
- 编写单元测试
- 遵循 SOLID 原则

### 测试
```bash
# 后端测试
cd server
go test ./...

# 前端测试
cd vue
npm run test
```

### 构建部署
```bash
# 构建扫描器
cd client/scanner
make build

# 构建服务端
cd server
go build -o main main.go

# 构建前端
cd vue
npm run build
```

## 📊 API 文档

### 主要接口（示例，真实以 /api/v1 为前缀）

#### 区块相关
- `GET /api/v1/blocks` - 获取区块列表
- `GET /api/v1/blocks/hash/:hash` - 获取区块详情
- `GET /api/v1/blocks/latest` - 获取最新区块

#### 交易相关
- `GET /api/v1/transactions` - 获取交易列表
- `GET /api/v1/transactions/hash/:hash` - 获取交易详情

#### 地址相关
- `GET /api/v1/addresses/:address` - 获取地址信息
- `GET /api/v1/addresses/:address/transactions` - 获取地址交易记录

#### WebSocket
- `ws://localhost:8080/ws` - 实时数据推送

## 📚 详细文档

- **[📖 完整文档](./docs/INDEX.md)** - 详细的使用指南和API文档
- **[🔐 安全配置](./docs/security-configuration.md)** - 生产环境安全设置
- **[🧰 脚本指南](./docs/scripts-guide.md)** - 证书与脚本使用
- **[📈 统计服务实现](./server/docs/stats-service-implementation.md)**
- **[⚙️ 性能优化记录](./server/docs/performance-optimization.md)**

## 🤝 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 开源协议

本项目采用 MIT 协议 - 查看 [LICENSE](LICENSE) 文件了解详情

## 👥 团队

- **开发者**: [airuse](https://gitee.com/airuse)

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者们！

## 📞 联系我们

如果你有任何问题或建议，请通过以下方式联系我们：
- 提交 Issue
- 发送邮件 viruse123@outlook.com
- 创建 Pull Request

---

⭐ 如果这个项目对你有帮助，请给我们一个 Star！
