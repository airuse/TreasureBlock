# 🔄 JWT令牌刷新指南

本文档介绍如何使用JWT令牌刷新功能来延长用户会话。

## 🎯 功能概述

JWT令牌刷新功能允许用户：
- 在令牌即将过期时获取新令牌
- 无需重新登录即可延长会话
- 保持应用程序的连续使用体验

## 🔐 刷新令牌流程

### 1. 用户登录获取初始令牌
```bash
POST /api/auth/login
Content-Type: application/json

{
  "username": "your_username",
  "password": "your_password"
}
```

**响应示例：**
```json
{
  "success": true,
  "message": "登录成功",
  "data": {
    "user_id": 1,
    "username": "your_username",
    "email": "your@email.com",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1672531200
  }
}
```

### 2. 使用令牌刷新接口
```bash
POST /api/auth/refresh
Authorization: Bearer <current_token>
Content-Type: application/json
```

**响应示例：**
```json
{
  "success": true,
  "message": "令牌刷新成功",
  "data": {
    "user_id": 1,
    "username": "your_username",
    "email": "your@email.com",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", // 新令牌
    "expires_at": 1672534800  // 新的过期时间
  }
}
```

## 🚀 使用场景

### 1. 前端自动刷新
```javascript
class TokenManager {
    constructor() {
        this.token = localStorage.getItem('token');
        this.expiresAt = localStorage.getItem('expiresAt');
        this.setupAutoRefresh();
    }

    setupAutoRefresh() {
        // 在令牌过期前5分钟自动刷新
        const refreshTime = this.expiresAt * 1000 - 5 * 60 * 1000;
        const now = Date.now();
        
        if (refreshTime > now) {
            setTimeout(() => this.refreshToken(), refreshTime - now);
        }
    }

    async refreshToken() {
        try {
            const response = await fetch('/api/auth/refresh', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (response.ok) {
                const data = await response.json();
                this.token = data.data.token;
                this.expiresAt = data.data.expires_at;
                
                // 更新本地存储
                localStorage.setItem('token', this.token);
                localStorage.setItem('expiresAt', this.expiresAt);
                
                // 设置下次刷新
                this.setupAutoRefresh();
            }
        } catch (error) {
            console.error('令牌刷新失败:', error);
            // 重定向到登录页面
            window.location.href = '/login';
        }
    }
}
```

### 2. 拦截器自动处理
```javascript
// Axios拦截器
axios.interceptors.response.use(
    response => response,
    async error => {
        if (error.response?.status === 401) {
            // 尝试刷新令牌
            try {
                const refreshResponse = await axios.post('/api/auth/refresh', {}, {
                    headers: {
                        'Authorization': `Bearer ${getToken()}`
                    }
                });
                
                // 更新令牌
                setToken(refreshResponse.data.data.token);
                
                // 重试原始请求
                error.config.headers['Authorization'] = `Bearer ${refreshResponse.data.data.token}`;
                return axios.request(error.config);
            } catch (refreshError) {
                // 刷新失败，重定向到登录
                window.location.href = '/login';
                return Promise.reject(refreshError);
            }
        }
        return Promise.reject(error);
    }
);
```

### 3. 命令行测试
```bash
# 使用提供的测试脚本
./test_refresh_token.sh

# 或手动测试
# 1. 登录获取令牌
TOKEN=$(curl -s -X POST "http://localhost:8080/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"password"}' \
  | jq -r '.data.token')

# 2. 刷新令牌
curl -X POST "http://localhost:8080/api/auth/refresh" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

## 🔒 安全特性

### 1. 令牌验证
- 刷新请求必须携带有效的JWT令牌
- 令牌必须未过期且签名有效
- 用户账户必须处于激活状态

### 2. 用户状态检查
- 验证用户是否仍然存在
- 检查用户账户是否被禁用
- 更新最后登录时间

### 3. 新令牌生成
- 每次刷新都生成全新的JWT令牌
- 新令牌有新的过期时间
- 旧令牌仍然有效直到过期

## ⚠️ 注意事项

### 1. 刷新时机
- 建议在令牌过期前5-10分钟刷新
- 避免在每次请求时都刷新令牌
- 考虑实现指数退避策略

### 2. 错误处理
- 刷新失败时重定向到登录页面
- 记录刷新失败的原因
- 提供用户友好的错误信息

### 3. 并发处理
- 避免多个并发刷新请求
- 实现请求去重机制
- 使用适当的锁机制

## 🧪 测试验证

### 1. 功能测试
```bash
# 运行完整测试
./test_refresh_token.sh

# 测试令牌过期场景
# 等待令牌过期后尝试刷新
```

### 2. 安全测试
```bash
# 测试无效令牌
curl -X POST "http://localhost:8080/api/auth/refresh" \
  -H "Authorization: Bearer invalid_token"

# 测试过期令牌
curl -X POST "http://localhost:8080/api/auth/refresh" \
  -H "Authorization: Bearer expired_token"
```

### 3. 性能测试
```bash
# 测试并发刷新
for i in {1..10}; do
    curl -X POST "http://localhost:8080/api/auth/refresh" \
      -H "Authorization: Bearer $TOKEN" &
done
wait
```

## 📋 配置选项

### 1. 令牌过期时间
```yaml
# server/config.yaml
security:
  jwt_expiration: 24h  # 登录令牌过期时间
```

### 2. 刷新策略
```go
// 可以在服务中配置刷新策略
type RefreshConfig struct {
    MinRefreshInterval time.Duration // 最小刷新间隔
    MaxRefreshAttempts int           // 最大刷新尝试次数
    RefreshWindow      time.Duration // 刷新窗口时间
}
```

## 🔗 相关接口

- **登录**: `POST /api/auth/login`
- **刷新**: `POST /api/auth/refresh`
- **用户资料**: `GET /api/user/profile`
- **修改密码**: `POST /api/user/change-password`

## 📞 故障排除

### 常见问题

1. **刷新失败 401**
   - 检查令牌是否有效
   - 确认令牌未过期
   - 验证用户账户状态

2. **刷新后仍显示过期**
   - 检查前端是否正确更新令牌
   - 确认本地存储已更新
   - 验证请求头中的令牌

3. **频繁刷新请求**
   - 实现刷新去重机制
   - 调整刷新时机
   - 检查前端逻辑

---

**提示**: 刷新令牌功能提供了更好的用户体验，但要注意合理使用，避免过度刷新！
