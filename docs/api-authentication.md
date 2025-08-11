# API 认证使用指南

本文档介绍如何使用新的API认证系统来保护区块链浏览器的API接口。

## 🔐 认证流程概述

1. **用户注册** - 创建用户账户
2. **用户登录** - 获取登录令牌
3. **创建API密钥** - 生成API Key和Secret Key
4. **获取访问令牌** - 使用API密钥换取访问令牌
5. **调用API** - 使用访问令牌调用受保护的API

## 📋 API 接口说明

### 1. 用户注册

```bash
POST /api/auth/register
Content-Type: application/json

{
  "username": "your_username",
  "email": "your_email@example.com",
  "password": "your_password"
}
```

**响应：**
```json
{
  "success": true,
  "message": "注册成功",
  "data": {
    "id": 1,
    "username": "your_username",
    "email": "your_email@example.com",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 2. 用户登录

```bash
POST /api/auth/login
Content-Type: application/json

{
  "username": "your_username",
  "password": "your_password"
}
```

**响应：**
```json
{
  "success": true,
  "message": "登录成功",
  "data": {
    "user_id": 1,
    "username": "your_username",
    "email": "your_email@example.com",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1672531200
  }
}
```

### 3. 创建API密钥

```bash
POST /api/user/api-keys
Authorization: Bearer <login_token>
Content-Type: application/json

{
  "name": "生产环境API密钥",
  "rate_limit": 1000,
  "expires_at": "2024-12-31T23:59:59Z"
}
```

**响应：**
```json
{
  "success": true,
  "message": "API密钥创建成功",
  "data": {
    "id": 1,
    "name": "生产环境API密钥",
    "api_key": "ak_1234567890abcdef1234567890abcdef",
    "secret_key": "sk_abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
    "rate_limit": 1000,
    "expires_at": "2024-12-31T23:59:59Z",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

> ⚠️ **重要提示**：`secret_key` 只在创建时返回一次，请妥善保存！

### 4. 获取访问令牌

```bash
POST /api/auth/token
Content-Type: application/json

{
  "api_key": "ak_1234567890abcdef1234567890abcdef",
  "secret_key": "sk_abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
}
```

**响应：**
```json
{
  "success": true,
  "message": "访问令牌获取成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "expires_at": 1672534800
  }
}
```

### 5. 调用受保护的API

```bash
GET /api/v1/blocks
Authorization: Bearer <access_token>
```

**响应：**
```json
{
  "success": true,
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "hash": "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
      "height": 0,
      "timestamp": "2009-01-03T18:15:05Z"
    }
  ]
}
```

## 🔧 扫描器配置

更新扫描器配置文件 `client/scanner/config.yaml`：

```yaml
server:
  host: "localhost"
  port: 8080
  protocol: "http"
  api_key: "ak_1234567890abcdef1234567890abcdef"
  secret_key: "sk_abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
```

## 📊 API使用统计

获取API使用统计信息：

```bash
GET /api/user/api-keys/{api_key_id}/stats
Authorization: Bearer <login_token>
```

**响应：**
```json
{
  "success": true,
  "message": "获取成功",
  "data": {
    "total_requests": 1000,
    "today_requests": 50,
    "this_hour_requests": 5,
    "avg_response_time": 120.5
  }
}
```

## 🛡️ 安全特性

### 限流保护
- 每个API密钥都有独立的限流设置
- 默认限制：1000请求/小时
- 可通过API密钥管理界面调整

### 访问令牌管理
- 访问令牌有效期：24小时（可配置）
- 自动刷新机制，提前5分钟刷新
- 令牌撤销功能

### 请求日志
- 记录所有API请求
- 包含IP地址、用户代理、响应时间等信息
- 支持审计和分析

## 💻 代码示例

### Go 客户端示例

```go
package main

import (
    "fmt"
    "log"
    
    "blockChainBrowser/client/scanner/pkg"
)

func main() {
    // 创建客户端
    client := pkg.NewClient(
        "http://localhost:8080",
        "ak_1234567890abcdef1234567890abcdef",
        "sk_abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
    )
    
    // 调用API
    var blocks []Block
    err := client.GET("/api/v1/blocks", &blocks)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("获取到 %d 个区块\\n", len(blocks))
}
```

### cURL 示例

```bash
#!/bin/bash

# 设置变量
API_BASE="http://localhost:8080"
API_KEY="ak_1234567890abcdef1234567890abcdef"
SECRET_KEY="sk_abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

# 获取访问令牌
ACCESS_TOKEN=$(curl -s -X POST "${API_BASE}/api/auth/token" \
  -H "Content-Type: application/json" \
  -d "{\"api_key\":\"${API_KEY}\",\"secret_key\":\"${SECRET_KEY}\"}" \
  | jq -r '.data.access_token')

echo "获取到访问令牌: ${ACCESS_TOKEN}"

# 调用API
curl -s -X GET "${API_BASE}/api/v1/blocks" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  | jq '.'
```

### Python 示例

```python
import requests
import json

class BlockchainAPIClient:
    def __init__(self, base_url, api_key, secret_key):
        self.base_url = base_url
        self.api_key = api_key
        self.secret_key = secret_key
        self.access_token = None
    
    def get_access_token(self):
        url = f"{self.base_url}/api/auth/token"
        payload = {
            "api_key": self.api_key,
            "secret_key": self.secret_key
        }
        
        response = requests.post(url, json=payload)
        response.raise_for_status()
        
        data = response.json()
        self.access_token = data['data']['access_token']
        return self.access_token
    
    def get_blocks(self):
        if not self.access_token:
            self.get_access_token()
        
        url = f"{self.base_url}/api/v1/blocks"
        headers = {"Authorization": f"Bearer {self.access_token}"}
        
        response = requests.get(url, headers=headers)
        response.raise_for_status()
        
        return response.json()

# 使用示例
client = BlockchainAPIClient(
    "http://localhost:8080",
    "ak_1234567890abcdef1234567890abcdef",
    "sk_abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
)

blocks = client.get_blocks()
print(f"获取到 {len(blocks['data'])} 个区块")
```

## 🚀 快速开始

1. **启动服务器**
   ```bash
   cd server
   go run main.go
   ```

2. **注册用户并创建API密钥**
   ```bash
   # 注册用户
   curl -X POST http://localhost:8080/api/auth/register \
     -H "Content-Type: application/json" \
     -d '{"username":"test","email":"test@example.com","password":"password123"}'
   
   # 登录获取令牌
   LOGIN_TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"test","password":"password123"}' \
     | jq -r '.data.token')
   
   # 创建API密钥
   curl -X POST http://localhost:8080/api/user/api-keys \
     -H "Authorization: Bearer $LOGIN_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"name":"测试密钥","rate_limit":1000}'
   ```

3. **配置扫描器**
   更新 `client/scanner/config.yaml` 中的 API 密钥

4. **启动扫描器**
   ```bash
   cd client/scanner
   go run cmd/main.go
   ```

## ❓ 常见问题

### Q: 如何重置API密钥？
A: 删除现有密钥并创建新的密钥。Secret Key不能重置，只能重新生成。

### Q: 访问令牌过期了怎么办？
A: 客户端会自动刷新令牌。如果手动调用，需要重新调用 `/api/auth/token` 接口。

### Q: 如何提高API调用限制？
A: 联系管理员或通过API密钥管理界面调整 `rate_limit` 参数。

### Q: 为什么要使用两层认证（API密钥 + 访问令牌）？
A: API密钥用于身份验证，访问令牌用于会话管理。这样可以：
- 提高安全性（令牌会过期）
- 防止API密钥泄露
- 支持令牌撤销
- 实现更好的审计追踪

## 🔗 相关链接

- [项目主页](https://gitee.com/airuse/treasure-block)
- [API文档](http://localhost:8080/docs)
- [WebSocket使用指南](./vue/src/utils/README_WebSocket.md)

---

**注意**：请在生产环境中使用强密码和HTTPS连接以确保安全性。
