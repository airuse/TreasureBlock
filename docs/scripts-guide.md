# Scripts 目录说明

这个目录包含了区块链浏览器项目的各种实用脚本。

## 📁 脚本列表

### 🔐 `generate-tls-cert.sh` - TLS证书生成
**用途**: 生成开发环境用的TLS证书
**使用方法**: 
```bash
# 在项目根目录执行
./scripts/generate-tls-cert.sh
```
**生成文件**: `server/certs/localhost.crt` 和 `server/certs/localhost.key`

### 🌐 `generate-domain-cert.sh` - 域名证书生成
**用途**: 为特定域名生成TLS证书
**使用方法**: 
```bash
# 为 example.com 生成证书
./scripts/generate-domain-cert.sh example.com
```
**适用场景**: 生产环境部署

### 📦 `install-dependencies.sh` - 依赖安装
**用途**: 自动安装TLS证书生成所需的依赖
**使用方法**: 
```bash
# 在Linux服务器上执行
./scripts/install-dependencies.sh
```
**支持系统**: Ubuntu, CentOS, Fedora, Alpine等

## 🚀 快速开始

### 1. 生成开发证书
```bash
# 确保在项目根目录
cd /path/to/blockChainBrowser

# 生成证书
./scripts/generate-tls-cert.sh
```

### 2. 启动HTTPS服务器
```bash
cd server
go run main.go
```

### 3. 测试访问
```bash
# 测试HTTPS（开发环境需要忽略证书警告）
curl -k https://localhost:8443/health
```

## 📋 注意事项

- 所有脚本都应在**项目根目录**执行
- 生成的证书仅适用于**开发环境**
- 生产环境请使用Let's Encrypt等CA签发的证书
- 脚本会自动创建必要的目录结构

## 🔧 故障排除

### 权限问题
```bash
chmod +x scripts/*.sh
```

### OpenSSL未安装
```bash
./scripts/install-dependencies.sh
```

### 证书路径错误
确保 `server/config.yaml` 中的路径正确：
```yaml
cert_file: "./certs/localhost.crt"
key_file: "./certs/localhost.key"
```
