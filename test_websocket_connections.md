# WebSocket连接数量测试

## 测试目标
验证前端是否只创建一个WebSocket连接，避免连接泄漏。

## 最新修复内容

### 1. 修复了BlocksView.vue的问题
- 移除了重复的WebSocket订阅
- 修复了组件卸载逻辑错误
- 避免了重复调用unsubscribeChainEvent

### 2. 修复了useChainWebSocket的问题
- 使用全局WebSocket管理器，避免重复创建
- 修复了类型错误
- 确保只有一个连接实例

## 测试步骤

### 1. 重启服务
```bash
# 重启后端服务
cd server
go run main.go

# 重启前端服务
cd vue
npm run dev
```

### 2. 检查初始状态
```bash
curl http://localhost:8443/ws/status
# 应该显示 total_clients: 0
```

### 3. 打开前端页面
- 打开浏览器，访问前端页面
- 等待页面完全加载
- 检查WebSocket连接状态

### 4. 验证连接数量
```bash
curl http://localhost:8443/ws/status
# 应该显示 total_clients: 1
```

### 5. 模拟用户操作
- 刷新页面
- 切换路由（如从首页切换到交易页面）
- 再次检查连接数量

### 6. 预期结果
- 连接数量应该始终为1
- 不应该出现多个连接
- 每个连接都应该有正确的订阅信息

## 修复后的架构

### 应用级别管理
- 在 `main.ts` 中创建全局WebSocket管理器
- 所有组件共享同一个连接实例
- 避免重复创建连接

### 组件级别使用
- 组件通过 `useChainWebSocket` 获取全局管理器
- 只管理订阅，不管理连接
- 组件卸载时只清理订阅，不关闭连接

### 连接生命周期
1. 应用启动时创建连接
2. 页面可见性变化时管理连接
3. 应用关闭时清理连接
4. 组件只负责订阅/取消订阅

## 如果仍有问题

### 检查点
1. 浏览器开发者工具中的WebSocket连接
2. 后端日志中的连接信息
3. 前端控制台的连接状态

### 调试命令
```bash
# 实时监控连接数量
watch -n 2 'curl -s http://localhost:8443/ws/status | jq ".data.total_clients"'

# 强制关闭所有连接
curl -X POST http://localhost:8443/ws/close-all
```

### 常见问题排查
1. **多个空订阅连接**：可能是组件重复挂载
2. **连接数量不断增加**：可能是组件没有正确卸载
3. **订阅信息丢失**：可能是WebSocket重连后没有重新订阅
