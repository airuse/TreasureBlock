# 🔐 生产环境安全配置指南

## 关于密码安全的说明

### ❓ 为什么不在前端加密密码？

你提出的问题非常好！让我解释现代Web应用的密码安全最佳实践：

#### 🚫 前端加密的问题

1. **哈希值变成新密码**
   ```javascript
   // 如果前端这样做：
   const hash = bcrypt.hash(password); // 前端加密
   // 这个哈希值就成了"真正的密码"
   ```

2. **重放攻击风险**
   - 攻击者截获哈希值后，可以直接用哈希值登录
   - 无法区分是用户还是攻击者在使用哈希值

3. **无法使用盐值**
   - 前端无法安全存储随机盐值
   - 相同密码在不同网站产生相同哈希
   - 容易受到彩虹表攻击

#### ✅ 正确的安全模型

```
用户密码 → HTTPS传输 → 服务器加盐哈希 → 数据库存储
    ↓                      ↓
   明文              随机盐+强哈希算法
                    (Argon2/bcrypt)
```

### 🛡️ 现代安全防护层次

## 1. 传输层安全 (TLS/HTTPS)

### 启用HTTPS

```bash
# 生成开发证书
./scripts/generate-ssl-cert.sh

# 配置HTTPS
# server/config.yaml
server:
  tls_enabled: true
  tls_port: 8443
  cert_file: "./certs/localhost.crt"
  key_file: "./certs/localhost.key"
```

### 获取生产证书

```bash
# 使用 Let's Encrypt (免费)
sudo apt-get install certbot
sudo certbot certonly --standalone -d yourdomain.com

# 或使用 acme.sh
curl https://get.acme.sh | sh
acme.sh --issue -d yourdomain.com --standalone
```

### HTTPS配置最佳实践

```yaml
# server/config.yaml
server:
  tls_enabled: true
  tls_port: 443
  cert_file: "/etc/ssl/certs/yourdomain.com.crt"
  key_file: "/etc/ssl/private/yourdomain.com.key"
  
  # 强制HTTPS
  force_https: true
  
  # HSTS配置
  hsts_max_age: 31536000  # 1年
  hsts_include_subdomains: true
  hsts_preload: true
```

## 2. 密码安全强化

### 密码强度要求

```go
// 密码策略配置
security:
  password_policy:
    min_length: 12
    require_uppercase: true
    require_lowercase: true
    require_digits: true
    require_special: true
    max_age_days: 90
    history_count: 5  # 记住最近5个密码
```

### 使用Argon2id算法

```go
// server/internal/services/auth_service.go
func (s *authService) hashPassword(password string) (string, error) {
    // 使用Argon2id替代bcrypt（更安全）
    return utils.HashPasswordArgon2(password)
}
```

### 密码复杂度验证

```go
func validatePassword(password string) error {
    if len(password) < 12 {
        return errors.New("密码长度至少12位")
    }
    
    checks := []struct {
        condition bool
        message   string
    }{
        {hasUpperCase(password), "必须包含大写字母"},
        {hasLowerCase(password), "必须包含小写字母"},
        {hasDigit(password), "必须包含数字"},
        {hasSpecialChar(password), "必须包含特殊字符"},
        {!isCommonPassword(password), "密码过于简单"},
        {!hasSequentialChars(password), "不能包含连续字符"},
    }
    
    for _, check := range checks {
        if !check.condition {
            return errors.New(check.message)
        }
    }
    
    return nil
}
```

## 3. 访问控制和监控

### 多因素认证 (2FA)

```go
// 添加TOTP支持
type User struct {
    // ... 其他字段
    TotpSecret   string `json:"-" gorm:"type:varchar(32)"`
    TotpEnabled  bool   `json:"totp_enabled" gorm:"default:false"`
    BackupCodes  string `json:"-" gorm:"type:text"` // JSON数组
}

func (s *authService) EnableTOTP(userID uint) (*TOTPSetup, error) {
    secret := generateTOTPSecret()
    qrCode := generateQRCode(secret, user.Email)
    backupCodes := generateBackupCodes()
    
    return &TOTPSetup{
        Secret:      secret,
        QRCode:      qrCode,
        BackupCodes: backupCodes,
    }, nil
}
```

### 会话管理

```go
type Session struct {
    ID          string    `json:"id" gorm:"primaryKey"`
    UserID      uint      `json:"user_id" gorm:"not null;index"`
    DeviceInfo  string    `json:"device_info" gorm:"type:varchar(500)"`
    IPAddress   string    `json:"ip_address" gorm:"type:varchar(45)"`
    UserAgent   string    `json:"user_agent" gorm:"type:varchar(500)"`
    LastActive  time.Time `json:"last_active"`
    ExpiresAt   time.Time `json:"expires_at"`
    IsRevoked   bool      `json:"is_revoked" gorm:"default:false"`
    CreatedAt   time.Time `json:"created_at"`
}

// 会话安全策略
func (s *authService) CreateSession(userID uint, req *http.Request) (*Session, error) {
    session := &Session{
        ID:         generateSessionID(),
        UserID:     userID,
        DeviceInfo: getDeviceInfo(req),
        IPAddress:  getClientIP(req),
        UserAgent:  req.UserAgent(),
        LastActive: time.Now(),
        ExpiresAt:  time.Now().Add(24 * time.Hour),
    }
    
    // 限制同时活跃会话数
    if err := s.limitActiveSessions(userID, 5); err != nil {
        return nil, err
    }
    
    return session, s.sessionRepo.Create(session)
}
```

### IP地址白名单

```go
// 配置IP白名单
security:
  ip_whitelist:
    admin_endpoints: 
      - "192.168.1.0/24"
      - "10.0.0.0/8"
    api_endpoints:
      enabled: false
      ips: []
```

### 地理位置检测

```go
func (s *authService) checkSuspiciousLogin(userID uint, ip string) error {
    location := getIPLocation(ip)
    lastLocation := s.getLastLoginLocation(userID)
    
    if location.Country != lastLocation.Country {
        // 发送安全警告邮件
        s.sendSecurityAlert(userID, "异地登录检测", location)
        
        // 要求额外验证
        return errors.New("检测到异地登录，请进行额外验证")
    }
    
    return nil
}
```

## 4. API安全防护

### 请求签名验证

```go
func (s *authService) ValidateAPISignature(req *http.Request, apiKey, secretKey string) error {
    timestamp := req.Header.Get("X-Timestamp")
    signature := req.Header.Get("X-Signature")
    
    // 检查时间戳防重放
    reqTime, _ := strconv.ParseInt(timestamp, 10, 64)
    if time.Now().Unix()-reqTime > 300 { // 5分钟窗口
        return errors.New("请求已过期")
    }
    
    // 验证签名
    expectedSig := s.generateSignature(req.Method, req.URL.Path, timestamp, secretKey)
    if signature != expectedSig {
        return errors.New("签名验证失败")
    }
    
    return nil
}

func (s *authService) generateSignature(method, path, timestamp, secret string) string {
    data := fmt.Sprintf("%s\n%s\n%s", method, path, timestamp)
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(data))
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
```

### 智能限流

```go
type SmartRateLimit struct {
    normalLimit    int           // 正常用户限制
    suspiciousLimit int          // 可疑用户限制
    windowSize     time.Duration
    redis          *redis.Client
}

func (rl *SmartRateLimit) CheckLimit(userID uint, ip string) error {
    // 基于用户行为的动态限流
    riskScore := rl.calculateRiskScore(userID, ip)
    
    var limit int
    if riskScore > 0.7 {
        limit = rl.suspiciousLimit
    } else {
        limit = rl.normalLimit
    }
    
    return rl.checkRate(fmt.Sprintf("user:%d", userID), limit)
}

func (rl *SmartRateLimit) calculateRiskScore(userID uint, ip string) float64 {
    score := 0.0
    
    // 因素1: 登录失败历史
    failCount := rl.getRecentFailCount(userID)
    score += float64(failCount) * 0.1
    
    // 因素2: 新IP地址
    if rl.isNewIP(userID, ip) {
        score += 0.3
    }
    
    // 因素3: 请求模式异常
    if rl.hasAbnormalPattern(userID) {
        score += 0.4
    }
    
    return math.Min(score, 1.0)
}
```

## 5. 数据保护

### 敏感数据加密

```go
type EncryptedField struct {
    Value     string `json:"-"`           // 加密值
    PlainText string `json:"-" gorm:"-"`  // 明文值（不存储）
}

func (ef *EncryptedField) Encrypt(plaintext string) error {
    encrypted, err := encrypt(plaintext, getEncryptionKey())
    if err != nil {
        return err
    }
    ef.Value = encrypted
    ef.PlainText = plaintext
    return nil
}

func (ef *EncryptedField) Decrypt() (string, error) {
    if ef.PlainText != "" {
        return ef.PlainText, nil
    }
    
    decrypted, err := decrypt(ef.Value, getEncryptionKey())
    if err != nil {
        return "", err
    }
    
    ef.PlainText = decrypted
    return decrypted, nil
}
```

### 数据脱敏

```go
func (u *User) Sanitize() *User {
    sanitized := *u
    sanitized.Password = ""
    
    // 脱敏邮箱
    if len(u.Email) > 0 {
        parts := strings.Split(u.Email, "@")
        if len(parts) == 2 {
            username := parts[0]
            if len(username) > 2 {
                sanitized.Email = username[:2] + "***@" + parts[1]
            }
        }
    }
    
    return &sanitized
}
```

## 6. 安全监控和告警

### 安全事件日志

```go
type SecurityEvent struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    EventType   string    `json:"event_type" gorm:"type:varchar(50);not null"`
    UserID      *uint     `json:"user_id,omitempty" gorm:"index"`
    IPAddress   string    `json:"ip_address" gorm:"type:varchar(45)"`
    UserAgent   string    `json:"user_agent" gorm:"type:varchar(500)"`
    Severity    string    `json:"severity" gorm:"type:varchar(20)"` // LOW, MEDIUM, HIGH, CRITICAL
    Message     string    `json:"message" gorm:"type:text"`
    Metadata    string    `json:"metadata" gorm:"type:json"`
    CreatedAt   time.Time `json:"created_at"`
}

func (s *securityService) LogEvent(eventType string, severity string, userID *uint, req *http.Request, message string, metadata map[string]interface{}) {
    metadataJSON, _ := json.Marshal(metadata)
    
    event := &SecurityEvent{
        EventType: eventType,
        UserID:    userID,
        IPAddress: getClientIP(req),
        UserAgent: req.UserAgent(),
        Severity:  severity,
        Message:   message,
        Metadata:  string(metadataJSON),
    }
    
    s.eventRepo.Create(event)
    
    // 高危事件立即告警
    if severity == "CRITICAL" {
        s.sendAlert(event)
    }
}
```

### 实时威胁检测

```go
func (s *securityService) AnalyzeThreat(req *http.Request) ThreatLevel {
    score := 0
    
    // 检查1: SQL注入尝试
    if s.detectSQLInjection(req) {
        score += 50
    }
    
    // 检查2: XSS尝试
    if s.detectXSS(req) {
        score += 30
    }
    
    // 检查3: 暴力破解
    if s.detectBruteForce(req) {
        score += 40
    }
    
    // 检查4: 异常User-Agent
    if s.detectBotUserAgent(req) {
        score += 20
    }
    
    if score >= 80 {
        return ThreatCritical
    } else if score >= 50 {
        return ThreatHigh
    } else if score >= 30 {
        return ThreatMedium
    }
    
    return ThreatLow
}
```

## 7. 生产环境部署

### 环境变量配置

```bash
# 生产环境变量
export JWT_SECRET="your-super-secret-jwt-key-256-bits-long"
export DB_PASSWORD="your-database-password"
export ENCRYPTION_KEY="your-encryption-key"
export TLS_CERT_FILE="/etc/ssl/certs/yourdomain.com.crt"
export TLS_KEY_FILE="/etc/ssl/private/yourdomain.com.key"

# 安全配置
export BCRYPT_COST=14
export SESSION_TIMEOUT=3600
export MAX_LOGIN_ATTEMPTS=3
export LOCKOUT_DURATION=1800
```

### Docker安全配置

```dockerfile
# 使用非root用户
FROM alpine:latest
RUN adduser -D -s /bin/sh appuser
USER appuser

# 最小权限
COPY --chown=appuser:appuser ./app /app
WORKDIR /app

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ./healthcheck || exit 1

CMD ["./app"]
```

### Nginx反向代理配置

```nginx
server {
    listen 443 ssl http2;
    server_name yourdomain.com;
    
    # SSL配置
    ssl_certificate /etc/ssl/certs/yourdomain.com.crt;
    ssl_certificate_key /etc/ssl/private/yourdomain.com.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
    ssl_prefer_server_ciphers off;
    
    # 安全头
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    
    # 限流
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=login:10m rate=1r/s;
    
    location /api/auth/login {
        limit_req zone=login burst=3 nodelay;
        proxy_pass http://localhost:8080;
    }
    
    location /api/ {
        limit_req zone=api burst=20 nodelay;
        proxy_pass http://localhost:8080;
    }
}
```

## 8. 安全检查清单

### ✅ 部署前检查

- [ ] 所有密码使用强哈希算法（Argon2id/bcrypt cost≥12）
- [ ] 启用HTTPS并配置HSTS
- [ ] 设置安全响应头
- [ ] 配置请求限流
- [ ] 启用暴力破解防护
- [ ] 敏感数据加密存储
- [ ] 错误信息不泄露敏感信息
- [ ] 会话管理安全
- [ ] 输入验证和输出编码
- [ ] 安全日志记录

### 🔍 定期安全审计

```bash
# 依赖安全扫描
go list -json -m all | nancy sleuth

# 代码安全扫描
gosec ./...

# 证书有效期检查
openssl x509 -in cert.pem -noout -dates

# 数据库安全检查
mysql_secure_installation
```

---

## 🎯 总结

密码安全的核心是：**在传输层使用HTTPS保护明文密码，在服务器端使用强哈希算法存储**。前端加密反而会降低安全性。

现代安全防护是多层次的：
1. **传输安全** - HTTPS/TLS
2. **认证安全** - 强密码策略 + 多因素认证
3. **会话安全** - 安全的令牌管理
4. **访问控制** - 限流 + 监控
5. **数据保护** - 加密 + 脱敏
6. **威胁检测** - 实时监控 + 告警

这样构建的系统才能真正抵御现代网络攻击！🛡️
