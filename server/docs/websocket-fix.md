# WebSocket Ping消息处理修复

## 问题描述

在WebSocket连接中，客户端会定期发送 `ping` 消息来检测连接状态，但服务器端代码存在以下问题：

```
{"level":"info","msg":"Failed to unmarshal message: invalid character 'p' looking for beginning of value","time":"2025-08-27T11:10:20+08:00"}
```

### 错误原因

1. **JSON解析错误**: 服务器期望接收JSON格式的消息，但 `ping` 是纯文本
2. **消息类型不匹配**: `ping` 消息被错误地传递给JSON解析器
3. **重复处理**: `ping` 消息在多个地方被处理，导致逻辑混乱

## 解决方案

### 1. 消息预处理

在JSON解析之前，先检查是否为特殊消息类型：

```go
// 检查是否为ping消息（特殊处理）
messageStr := string(message)
if messageStr == "ping" {
    // 直接响应pong，不需要JSON解析
    response := map[string]interface{}{
        "type": "pong",
        "data": "pong",
    }
    h.sendMessage(client, response)
    continue
}
```

### 2. 优化心跳检测

改进服务器端的心跳检测机制：

```go
case <-ticker.C:
    // 发送心跳ping消息
    pingMessage := map[string]interface{}{
        "type": "ping",
        "data": "ping",
        "timestamp": time.Now().UnixMilli(),
    }
    
    pingData, err := json.Marshal(pingMessage)
    if err != nil {
        log.Printf("Failed to marshal ping message: %v", err)
        return
    }
    
    if err := client.conn.WriteMessage(websocket.TextMessage, pingData); err != nil {
        log.Printf("Failed to send ping message: %v", err)
        return
    }
```

### 3. 连接参数优化

添加WebSocket连接配置：

```go
// 设置连接参数
conn.SetReadLimit(512 * 1024) // 限制消息大小为512KB
conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // 设置读取超时
conn.SetPongHandler(func(string) error {
    conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // 重置读取超时
    return nil
})
```

### 4. 错误处理改进

区分不同类型的连接错误：

```go
if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
    log.Printf("WebSocket unexpected close error: %v", err)
} else {
    log.Printf("WebSocket read error: %v", err)
}
```

## 修复后的消息流程

### Ping消息处理流程

```
客户端发送 "ping" 
    ↓
服务器检测到纯文本消息
    ↓
直接响应 {"type": "pong", "data": "pong"}
    ↓
不进行JSON解析
    ↓
避免错误日志
```

### 正常消息处理流程

```
客户端发送JSON消息
    ↓
服务器检测到非ping消息
    ↓
进行JSON解析
    ↓
处理业务逻辑
    ↓
发送响应
```

## 配置参数

### WebSocket升级器配置

```go
upgrader := websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // 允许所有来源，生产环境应该限制
    },
    EnableCompression: true,    // 启用压缩
    ReadBufferSize:    1024,    // 读取缓冲区大小
    WriteBufferSize:   1024,    // 写入缓冲区大小
}
```

### 连接参数配置

```go
conn.SetReadLimit(512 * 1024)                    // 消息大小限制
conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // 读取超时
conn.SetPongHandler(func(string) error { ... })  // Pong处理器
```

## 测试验证

### 运行测试

```bash
cd server
go test ./internal/handlers -v -run TestWebSocket
```

### 测试覆盖

- ✅ WebSocket处理器创建
- ✅ 消息结构验证
- ✅ 订阅消息处理
- ✅ 消息类型常量
- ✅ 消息分类常量
- ✅ 区块链类型常量

## 性能优化

### 1. 消息缓冲区

```go
send: make(chan []byte, 256) // 256条消息的缓冲区
```

### 2. 心跳间隔

```go
ticker := time.NewTicker(30 * time.Second) // 30秒心跳间隔
```

### 3. 连接超时

```go
conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // 60秒读取超时
```

## 监控和日志

### 连接日志

```go
log.Printf("New WebSocket client connected from %s", conn.RemoteAddr().String())
log.Printf("WebSocket client disconnected from %s", client.conn.RemoteAddr().String())
```

### 错误日志

```go
log.Printf("Failed to unmarshal message: %v, message: %s", err, messageStr)
log.Printf("WebSocket unexpected close error: %v", err)
log.Printf("WebSocket read error: %v", err)
```

## 生产环境建议

### 1. 安全配置

```go
CheckOrigin: func(r *http.Request) bool {
    // 生产环境应该限制来源
    origin := r.Header.Get("Origin")
    allowedOrigins := []string{"https://yourdomain.com"}
    for _, allowed := range allowedOrigins {
        if origin == allowed {
            return true
        }
    }
    return false
}
```

### 2. 速率限制

```go
// 添加消息速率限制
// 防止客户端发送过多消息
```

### 3. 连接池管理

```go
// 限制最大连接数
// 监控连接状态
// 自动清理僵尸连接
```

## 总结

通过这次修复，WebSocket连接现在能够：

1. **正确处理ping消息**: 不再产生JSON解析错误
2. **稳定心跳检测**: 30秒间隔的心跳机制
3. **更好的错误处理**: 区分不同类型的连接错误
4. **连接参数优化**: 合理的超时和缓冲区设置
5. **完整的测试覆盖**: 确保功能正确性

这些改进使得WebSocket连接更加稳定可靠，减少了错误日志的产生，提升了用户体验。
