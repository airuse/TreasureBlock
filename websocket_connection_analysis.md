# WebSocket连接泄漏分析

## 问题描述
后端出现多个WebSocket连接，正常情况下一个浏览器应用应该只与后端建立一个WebSocket连接。

## 可能的原因

### 1. 前端连接泄漏
- **页面刷新/导航**：页面刷新时WebSocket连接没有正确关闭
- **组件重复挂载**：Vue组件重复挂载导致多个WebSocket实例
- **路由切换**：路由切换时没有清理WebSocket连接
- **开发环境热重载**：Vue开发环境热重载导致连接累积

### 2. 后端连接管理问题
- **连接没有正确清理**：客户端断开连接后，后端没有及时清理
- **goroutine泄漏**：WebSocket处理goroutine没有正确退出
- **资源没有释放**：连接相关的资源没有正确释放

### 3. 网络问题
- **连接中断重连**：网络中断导致自动重连，旧连接没有清理
- **TLS握手失败**：TLS证书问题导致连接异常

## 诊断方法

### 1. 查看当前连接状态
```bash
# 查看WebSocket连接状态
curl http://localhost:8443/ws/status

# 响应示例
{
  "success": true,
  "data": {
    "total_clients": 1,
    "active_connections": [
      {
        "remote_addr": "[::1]:12345",
        "subscribed_to": {"block:eth": true},
        "connected_at": 1756429189
      }
    ],
    "subscription_stats": {
      "block:eth": 1
    },
    "last_updated": 1756429189
  }
}
```

### 2. 强制关闭所有连接
```bash
# 关闭所有WebSocket连接（用于调试）
curl -X POST http://localhost:8443/ws/close-all
```

### 3. 监控连接变化
```bash
# 每5秒查看一次连接状态
watch -n 5 'curl -s http://localhost:8443/ws/status | jq ".data.total_clients"'
```

## 前端修复建议

### 1. 确保连接唯一性
```typescript
// 在useWebSocket中确保只有一个连接
let globalWebSocketManager: WebSocketManager | null = null

export function createWebSocketManager(options: WebSocketOptions): WebSocketManager {
  if (globalWebSocketManager) {
    globalWebSocketManager.destroy()
  }
  
  globalWebSocketManager = new WebSocketManager(options)
  return globalWebSocketManager
}
```

### 2. 页面可见性处理
```typescript
// 页面隐藏时断开连接，显示时重连
document.addEventListener('visibilitychange', () => {
  if (document.hidden) {
    // 页面隐藏，可以选择断开连接
    manager.value?.disconnect()
  } else {
    // 页面显示，重新连接
    manager.value?.connect()
  }
})
```

### 3. 路由切换时清理
```typescript
// 在Vue Router的beforeEach中清理连接
router.beforeEach((to, from, next) => {
  // 清理WebSocket连接
  if (manager.value) {
    manager.value.disconnect()
  }
  next()
})
```

## 后端修复建议

### 1. 连接超时管理
```go
// 设置合理的连接超时
conn.SetReadDeadline(time.Now().Add(60 * time.Second))
conn.SetPongHandler(func(string) error {
    conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    return nil
})
```

### 2. 连接清理
```go
// 在连接断开时及时清理
defer func() {
    log.Printf("WebSocket client disconnected from %s", client.conn.RemoteAddr().String())
    h.unregister <- client
    client.conn.Close()
    close(client.send)
}()
```

### 3. 定期清理僵尸连接
```go
// 定期检查并清理僵尸连接
func (h *WebSocketHandler) cleanupZombieConnections() {
    ticker := time.NewTicker(5 * time.Minute)
    go func() {
        for range ticker.C {
            h.mutex.Lock()
            for client := range h.clients {
                // 检查连接是否还活着
                if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                    delete(h.clients, client)
                    client.conn.Close()
                    close(client.send)
                }
            }
            h.mutex.Unlock()
        }
    }()
}
```

## 测试步骤

### 1. 重启服务
```bash
cd server
go run main.go
```

### 2. 检查初始状态
```bash
curl http://localhost:8443/ws/status
# 应该显示 total_clients: 0
```

### 3. 打开前端页面
- 打开浏览器，访问前端页面
- 检查WebSocket连接状态
- 应该显示 total_clients: 1

### 4. 模拟连接泄漏
- 刷新页面多次
- 切换路由
- 检查连接数量是否增加

### 5. 验证修复
- 连接数量应该保持为1
- 不再出现连接泄漏

## 预期结果
- 每个浏览器应用只维持一个WebSocket连接
- 连接断开后及时清理
- 连接数量稳定，不出现泄漏
- ping/pong机制正常工作
