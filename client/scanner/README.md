# 区块链扫块器 (Blockchain Scanner)

一个高性能的区块链数据扫描工具，支持多链并行扫描和批量数据处理。

## 主要特性

- **多链支持**: 支持 Bitcoin (BTC) 和 Ethereum (ETH) 等主流区块链
- **高性能扫描**: 优化的区块扫描算法，快速处理新区块
- **批量上传**: 支持交易数据的批量上传，显著提升性能
- **配置灵活**: 支持链级别的独立配置和动态调整
- **容错机制**: 内置重试、超时和错误处理机制

## 性能优化

### 批量上传功能

为了解决扫块程序跟不上最新区块数据的问题，我们实现了批量上传功能：

#### 问题分析
- **ETH出块时间**: 平均12秒
- **扫块处理时间**: 平均4秒
- **原有上传时间**: 10秒（并发轮询方式）
- **总耗时**: 14秒 > 12秒，无法跟上新区块

#### 解决方案
- **批量上传**: 将多个交易打包成一批，减少网络IO次数
- **配置控制**: 支持启用/禁用批量上传，灵活配置批量大小
- **性能提升**: 预计可将上传时间从10秒减少到2-3秒

#### 配置示例
```yaml
blockchain:
  chains:
    eth:
      scan:
        # 启用批量上传
        batch_upload: true           # 推荐启用
        batch_size: 1000             # 批量大小，默认1000条
        batch_timeout: 30s           # 上传超时时间
        # 其他配置...
        max_concurrent: 40           # 单个上传时的并发数
```

#### 性能对比
| 上传方式 | 网络IO次数 | 预计耗时 | 适用场景 |
|---------|------------|----------|----------|
| 单个上传 | N次（N=交易数） | 10秒 | 交易数量少，网络延迟低 |
| 批量上传 | 1次 | 2-3秒 | 交易数量多，推荐使用 |

## 快速开始

### 1. 安装依赖
```bash
go mod download
```

### 2. 配置
复制 `config.yaml.example` 到 `config.yaml` 并修改配置：
```yaml
server:
  host: "your-server-host"
  port: 8443
  protocol: "https"
  api_key: "your-api-key"
  secret_key: "your-secret-key"

blockchain:
  chains:
    eth:
      enabled: true
      scan:
        batch_upload: true  # 启用批量上传
        batch_size: 1000    # 设置批量大小
```

### 3. 运行
```bash
go run cmd/main.go
```

## 配置说明

### 批量上传配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `batch_upload` | bool | false | 是否启用批量上传 |
| `batch_size` | int | 1000 | 批量上传的交易数量 |
| `batch_timeout` | duration | 30s | 批量上传超时时间 |

### 扫描配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `interval` | duration | 1s | 扫描间隔 |
| `confirmations` | int | 6 | 确认数 |
| `max_concurrent` | int | 10 | 最大并发数 |
| `block_timeout` | duration | 30s | 区块超时时间 |

## 性能调优建议

### 1. 启用批量上传
```yaml
scan:
  batch_upload: true
  batch_size: 1000
```

### 2. 调整扫描间隔
```yaml
scan:
  interval: 1s  # 根据出块时间调整
```

### 3. 优化确认数
```yaml
scan:
  confirmations: 61  # ETH推荐61个确认
```

### 4. 监控性能指标
- 区块扫描时间
- 交易上传时间
- 内存使用情况
- 网络延迟

## 故障排除

### 常见问题

1. **上传超时**
   - 检查 `batch_timeout` 配置
   - 调整 `batch_size` 大小
   - 检查网络连接

2. **内存使用过高**
   - 减少 `batch_size`
   - 启用文件保存 `save_to_file: true`

3. **扫描速度慢**
   - 启用批量上传
   - 增加 `max_concurrent`
   - 检查RPC节点性能

## 开发说明

### 架构设计
- **模块化设计**: 清晰的职责分离
- **接口抽象**: 支持不同链的扩展
- **配置驱动**: 灵活的配置管理
- **错误处理**: 完善的错误处理机制

### 扩展新链
1. 实现 `Scanner` 接口
2. 添加链配置
3. 在 `initializeScanners` 中注册

## 许可证

MIT License

