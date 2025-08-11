# 🔐 安全防护指南

## 网络传输安全

### 1. HTTPS 配置（生产环境必须）

更新服务器配置支持 HTTPS：

```go
// server/internal/server/server.go
func (s *Server) StartTLS(certFile, keyFile string) error {
    log.Printf("Starting HTTPS server on %s", s.server.Addr)
    return s.server.ListenAndServeTLS(certFile, keyFile)
}
```

配置文件添加TLS支持：

```yaml
# server/config.yaml
server:
  host: "0.0.0.0"
  port: 8443
  tls_enabled: true
  cert_file: "/path/to/cert.pem"
  key_file: "/path/to/key.pem"
  read_timeout: 30s
  write_timeout: 30s
```

### 2. 强制HTTPS重定向

```go
func HTTPSRedirectMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        if c.Request.Header.Get("X-Forwarded-Proto") != "https" {
            if c.Request.Method != "GET" {
                c.AbortWithStatus(http.StatusBadRequest)
                return
            }
            
            target := "https://" + c.Request.Host + c.Request.URL.Path
            if len(c.Request.URL.RawQuery) > 0 {
                target += "?" + c.Request.URL.RawQuery
            }
            c.Redirect(http.StatusMovedPermanently, target)
            return
        }
        c.Next()
    })
}
```

## 密码安全最佳实践

### 1. 密码强度验证

```go
// server/internal/utils/password.go
package utils

import (
    "errors"
    "regexp"
    "unicode"
)

type PasswordStrength struct {
    MinLength    int
    RequireUpper bool
    RequireLower bool
    RequireDigit bool
    RequireSpecial bool
}

func ValidatePasswordStrength(password string, rules PasswordStrength) error {
    if len(password) < rules.MinLength {
        return errors.New(fmt.Sprintf("密码长度至少需要%d位", rules.MinLength))
    }
    
    if rules.RequireUpper && !hasUpperCase(password) {
        return errors.New("密码必须包含大写字母")
    }
    
    if rules.RequireLower && !hasLowerCase(password) {
        return errors.New("密码必须包含小写字母")
    }
    
    if rules.RequireDigit && !hasDigit(password) {
        return errors.New("密码必须包含数字")
    }
    
    if rules.RequireSpecial && !hasSpecialChar(password) {
        return errors.New("密码必须包含特殊字符")
    }
    
    // 检查常见弱密码
    if isCommonPassword(password) {
        return errors.New("密码过于简单，请使用更复杂的密码")
    }
    
    return nil
}

func hasUpperCase(s string) bool {
    for _, r := range s {
        if unicode.IsUpper(r) {
            return true
        }
    }
    return false
}

func hasLowerCase(s string) bool {
    for _, r := range s {
        if unicode.IsLower(r) {
            return true
        }
    }
    return false
}

func hasDigit(s string) bool {
    for _, r := range s {
        if unicode.IsDigit(r) {
            return true
        }
    }
    return false
}

func hasSpecialChar(s string) bool {
    specialChars := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`)
    return specialChars.MatchString(s)
}

func isCommonPassword(password string) bool {
    commonPasswords := []string{
        "password", "123456", "123456789", "12345678", "12345",
        "1234567", "password123", "admin", "qwerty", "abc123",
        "Password1", "password1", "123123", "000000",
    }
    
    for _, common := range commonPasswords {
        if password == common {
            return true
        }
    }
    return false
}
```

### 2. 增强的密码哈希

```go
// server/internal/utils/crypto.go
package utils

import (
    "crypto/rand"
    "crypto/subtle"
    "encoding/base64"
    "errors"
    "fmt"
    "strings"
    
    "golang.org/x/crypto/argon2"
    "golang.org/x/crypto/bcrypt"
)

type HashConfig struct {
    Memory      uint32
    Iterations  uint32
    Parallelism uint8
    SaltLength  uint32
    KeyLength   uint32
}

var DefaultHashConfig = HashConfig{
    Memory:      64 * 1024, // 64MB
    Iterations:  3,
    Parallelism: 2,
    SaltLength:  16,
    KeyLength:   32,
}

// Argon2id 哈希（推荐用于新项目）
func HashPasswordArgon2(password string) (string, error) {
    salt := make([]byte, DefaultHashConfig.SaltLength)
    if _, err := rand.Read(salt); err != nil {
        return "", err
    }
    
    hash := argon2.IDKey(
        []byte(password),
        salt,
        DefaultHashConfig.Iterations,
        DefaultHashConfig.Memory,
        DefaultHashConfig.Parallelism,
        DefaultHashConfig.KeyLength,
    )
    
    b64Salt := base64.RawStdEncoding.EncodeToString(salt)
    b64Hash := base64.RawStdEncoding.EncodeToString(hash)
    
    return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
        argon2.Version,
        DefaultHashConfig.Memory,
        DefaultHashConfig.Iterations,
        DefaultHashConfig.Parallelism,
        b64Salt,
        b64Hash,
    ), nil
}

func VerifyPasswordArgon2(password, hash string) bool {
    parts := strings.Split(hash, "$")
    if len(parts) != 6 {
        return false
    }
    
    var memory, iterations uint32
    var parallelism uint8
    
    _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
    if err != nil {
        return false
    }
    
    salt, err := base64.RawStdEncoding.DecodeString(parts[4])
    if err != nil {
        return false
    }
    
    decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
    if err != nil {
        return false
    }
    
    comparisonHash := argon2.IDKey(
        []byte(password),
        salt,
        iterations,
        memory,
        parallelism,
        uint32(len(decodedHash)),
    )
    
    return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1
}

// 保持 bcrypt 兼容性
func HashPasswordBcrypt(password string) (string, error) {
    cost := 12 // 更高的成本
    return bcrypt.GenerateFromPassword([]byte(password), cost)
}

func VerifyPasswordBcrypt(password, hash string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
```

## 防止常见攻击

### 1. 暴力破解防护

```go
// server/internal/middleware/security_middleware.go
package middleware

import (
    "fmt"
    "net/http"
    "sync"
    "time"
    
    "github.com/gin-gonic/gin"
)

type LoginAttempt struct {
    Count     int
    LastTry   time.Time
    BlockedAt *time.Time
}

type BruteForceProtector struct {
    attempts    map[string]*LoginAttempt
    mutex       sync.RWMutex
    maxAttempts int
    blockTime   time.Duration
    windowTime  time.Duration
}

func NewBruteForceProtector() *BruteForceProtector {
    return &BruteForceProtector{
        attempts:    make(map[string]*LoginAttempt),
        maxAttempts: 5,
        blockTime:   15 * time.Minute,
        windowTime:  1 * time.Hour,
    }
}

func (bf *BruteForceProtector) LoginAttemptMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.URL.Path == "/api/auth/login" && c.Request.Method == "POST" {
            clientIP := c.ClientIP()
            
            if bf.isBlocked(clientIP) {
                c.JSON(http.StatusTooManyRequests, gin.H{
                    "success": false,
                    "error":   "登录尝试过多，请稍后再试",
                })
                c.Abort()
                return
            }
        }
        
        c.Next()
        
        // 如果登录失败，记录尝试
        if c.Request.URL.Path == "/api/auth/login" && c.Writer.Status() == 401 {
            bf.recordFailedAttempt(c.ClientIP())
        } else if c.Request.URL.Path == "/api/auth/login" && c.Writer.Status() == 200 {
            bf.clearAttempts(c.ClientIP())
        }
    }
}

func (bf *BruteForceProtector) isBlocked(ip string) bool {
    bf.mutex.RLock()
    defer bf.mutex.RUnlock()
    
    attempt, exists := bf.attempts[ip]
    if !exists {
        return false
    }
    
    if attempt.BlockedAt != nil && time.Since(*attempt.BlockedAt) < bf.blockTime {
        return true
    }
    
    return false
}

func (bf *BruteForceProtector) recordFailedAttempt(ip string) {
    bf.mutex.Lock()
    defer bf.mutex.Unlock()
    
    now := time.Now()
    attempt, exists := bf.attempts[ip]
    
    if !exists {
        bf.attempts[ip] = &LoginAttempt{
            Count:   1,
            LastTry: now,
        }
        return
    }
    
    // 如果超过时间窗口，重置计数
    if now.Sub(attempt.LastTry) > bf.windowTime {
        attempt.Count = 1
        attempt.BlockedAt = nil
    } else {
        attempt.Count++
    }
    
    attempt.LastTry = now
    
    // 如果超过最大尝试次数，标记为阻止
    if attempt.Count >= bf.maxAttempts {
        blockTime := now
        attempt.BlockedAt = &blockTime
    }
}

func (bf *BruteForceProtector) clearAttempts(ip string) {
    bf.mutex.Lock()
    defer bf.mutex.Unlock()
    delete(bf.attempts, ip)
}
```

### 2. CSRF 防护

```go
func CSRFMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" {
            token := c.GetHeader("X-CSRF-Token")
            if token == "" {
                token = c.PostForm("_token")
            }
            
            if !isValidCSRFToken(token, c) {
                c.JSON(http.StatusForbidden, gin.H{
                    "success": false,
                    "error":   "CSRF token validation failed",
                })
                c.Abort()
                return
            }
        }
        c.Next()
    }
}
```

### 3. SQL 注入防护

我们已经使用了 GORM，它自动防止 SQL 注入，但还要注意：

```go
// ✅ 安全的做法
user, err := s.userRepo.GetByUsername(req.Username)

// ❌ 危险的做法（永远不要这样）
query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", req.Username)
```

## 配置安全头

```go
func SecurityHeadersMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 防止XSS攻击
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        
        // HTTPS相关
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        
        // 内容安全策略
        c.Header("Content-Security-Policy", "default-src 'self'")
        
        // 引用策略
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        
        c.Next()
    }
}
```

## 敏感信息保护

```go
// 配置敏感信息不出现在日志中
func SanitizeForLog(data interface{}) interface{} {
    if m, ok := data.(map[string]interface{}); ok {
        sanitized := make(map[string]interface{})
        for k, v := range m {
            if isSensitiveField(k) {
                sanitized[k] = "***"
            } else {
                sanitized[k] = v
            }
        }
        return sanitized
    }
    return data
}

func isSensitiveField(field string) bool {
    sensitiveFields := []string{
        "password", "secret", "token", "key", "auth",
        "credential", "private", "confidential",
    }
    
    fieldLower := strings.ToLower(field)
    for _, sensitive := range sensitiveFields {
        if strings.Contains(fieldLower, sensitive) {
            return true
        }
    }
    return false
}
```
