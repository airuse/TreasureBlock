# 用户交易系统

## 概述

用户交易系统是一个完整的离线签名交易解决方案，允许用户在网页上创建交易，然后导出到离线电脑进行签名，最后导入签名数据并发送交易。这种方式确保了私钥的安全性，因为签名过程完全在离线环境中进行。

## 功能特性

### 1. 交易创建
- 支持ETH和BTC链
- 可配置Gas参数（ETH）
- 支持交易备注
- 自动生成未签名交易数据

### 2. 交易状态管理
- **草稿 (draft)**: 刚创建的交易
- **未签名 (unsigned)**: 已导出但未签名的交易
- **未发送 (unsent)**: 已导入签名但未发送的交易
- **在途 (in_progress)**: 已发送到网络的交易
- **已打包 (packed)**: 已被打包到区块的交易
- **已确认 (confirmed)**: 已确认的交易
- **失败 (failed)**: 失败的交易

### 3. 离线签名流程
1. 用户在网页上创建交易
2. 系统生成未签名交易数据
3. 用户导出交易数据到离线电脑
4. 离线电脑使用私钥签名交易
5. 用户导入签名数据
6. 系统发送已签名的交易

## 技术架构

### 后端架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Handler       │    │    Service      │    │   Repository    │
│                 │    │                 │    │                 │
│ UserTransaction │───▶│ UserTransaction │───▶│ UserTransaction │
│    Handler      │    │    Service      │    │   Repository    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│      DTO        │    │     Model       │    │    Database     │
│                 │    │                 │    │                 │
│ Request/Response│    │ UserTransaction │    │ user_transactions│
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 前端架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Component     │    │      API        │    │     Types       │
│                 │    │                 │    │                 │
│TransactionsView │───▶│ userTransactions│───▶│ UserTransaction │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ CreateTransaction│    │     Request     │    │   Response      │
│     Modal       │    │      Utils      │    │      Utils      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## API接口

### 1. 创建交易
```http
POST /api/user/transactions
Content-Type: application/json

{
  "chain": "eth",
  "symbol": "ETH",
  "from_address": "0x...",
  "to_address": "0x...",
  "amount": "0.1",
  "fee": "0.00042",
  "gas_limit": 21000,
  "gas_price": "20",
  "nonce": 0,
  "remark": "测试交易"
}
```

### 2. 获取交易列表
```http
GET /api/user/transactions?page=1&page_size=10&status=unsigned
```

### 3. 导出交易
```http
POST /api/user/transactions/{id}/export
```

### 4. 导入签名
```http
POST /api/user/transactions/{id}/import-signature
Content-Type: application/json

{
  "id": 1,
  "signed_tx": "0x..."
}
```

### 5. 发送交易
```http
POST /api/user/transactions/{id}/send
```

### 6. 获取交易统计
```http
GET /api/user/transactions/stats
```

## 数据库设计

### 用户交易表 (user_transactions)
```sql
CREATE TABLE `user_transactions` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) unsigned NOT NULL COMMENT '用户ID',
  `chain` varchar(20) NOT NULL COMMENT '链类型(btc,eth)',
  `symbol` varchar(20) NOT NULL COMMENT '币种',
  `from_address` varchar(120) NOT NULL COMMENT '发送地址',
  `to_address` varchar(120) NOT NULL COMMENT '接收地址',
  `amount` decimal(65,0) NOT NULL COMMENT '交易金额',
  `fee` decimal(65,0) NOT NULL DEFAULT '0' COMMENT '手续费',
  `gas_limit` int(11) unsigned DEFAULT NULL COMMENT 'Gas限制',
  `gas_price` varchar(100) DEFAULT NULL COMMENT 'Gas价格',
  `nonce` bigint(20) unsigned DEFAULT NULL COMMENT '交易序号',
  `status` varchar(20) NOT NULL DEFAULT 'draft' COMMENT '状态',
  `tx_hash` varchar(120) DEFAULT NULL COMMENT '交易哈希',
  `unsigned_tx` longtext DEFAULT NULL COMMENT '未签名交易数据',
  `signed_tx` longtext DEFAULT NULL COMMENT '已签名交易数据',
  `block_height` bigint(20) unsigned DEFAULT NULL COMMENT '区块高度',
  `confirmations` int(11) unsigned DEFAULT '0' COMMENT '确认数',
  `error_msg` text DEFAULT NULL COMMENT '错误信息',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_chain` (`chain`),
  KEY `idx_status` (`status`),
  KEY `idx_from_address` (`from_address`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户交易表';
```

## 使用流程

### 1. 创建交易
1. 在交易历史页面点击"新建交易"按钮
2. 填写交易信息（发送地址、接收地址、金额、手续费等）
3. 选择链类型（ETH/BTC）
4. 点击"创建交易"按钮

### 2. 导出交易
1. 在交易列表中找到刚创建的交易
2. 点击"导出交易"按钮
3. 系统会生成未签名交易数据
4. 复制或下载交易数据

### 3. 离线签名
1. 将交易数据传输到离线电脑
2. 使用离线钱包或签名工具进行签名
3. 获得已签名的交易数据

### 4. 导入签名
1. 在交易列表中找到对应的交易
2. 点击"导入签名"按钮
3. 粘贴已签名的交易数据
4. 点击"导入签名"按钮

### 5. 发送交易
1. 导入签名成功后，交易状态变为"未发送"
2. 点击"发送交易"按钮
3. 系统将交易发送到区块链网络
4. 交易状态变为"在途"

## 安全考虑

### 1. 私钥安全
- 私钥永远不会离开用户的离线环境
- 网页端只处理未签名和已签名的交易数据
- 签名过程完全在离线环境中进行

### 2. 数据验证
- 所有输入数据都经过验证
- 交易参数在发送前进行二次确认
- 支持交易失败的错误处理

### 3. 权限控制
- 只有交易创建者才能操作交易
- 用户ID验证确保数据隔离
- JWT认证保护所有API接口

## 扩展功能

### 1. 批量操作
- 支持批量创建交易
- 支持批量导出交易
- 支持批量导入签名

### 2. 交易模板
- 保存常用交易参数
- 快速创建相似交易
- 支持交易参数预设

### 3. 多链支持
- 支持更多区块链网络
- 统一的交易管理界面
- 链特定的参数配置

### 4. 交易监控
- 实时交易状态更新
- 交易确认数监控
- 失败交易自动重试

## 部署说明

### 1. 数据库迁移
```bash
# 执行数据库迁移脚本
mysql -u username -p database_name < server/sql/migrations/create_user_transactions_table.sql
```

### 2. 后端启动
```bash
cd server
go mod tidy
go run main.go
```

### 3. 前端启动
```bash
cd vue
npm install
npm run dev
```

## 故障排除

### 1. 常见问题
- **交易创建失败**: 检查必填字段和参数格式
- **导出失败**: 确认交易状态是否为草稿或未签名
- **导入失败**: 检查签名数据格式是否正确
- **发送失败**: 确认网络连接和Gas费用设置

### 2. 日志查看
- 后端日志: 查看控制台输出
- 前端日志: 查看浏览器开发者工具
- 数据库日志: 查看MySQL错误日志

### 3. 性能优化
- 添加数据库索引
- 实现分页查询
- 使用缓存减少数据库查询
- 异步处理长时间操作

## 贡献指南

### 1. 代码规范
- 遵循Go语言编码规范
- 使用TypeScript进行前端开发
- 添加适当的注释和文档
- 编写单元测试

### 2. 提交规范
- 使用清晰的提交信息
- 一个提交只包含一个功能
- 在提交前运行测试
- 更新相关文档

### 3. 问题反馈
- 使用GitHub Issues报告问题
- 提供详细的错误信息和复现步骤
- 包含系统环境信息
- 描述期望的行为

## 许可证

本项目采用MIT许可证，详见LICENSE文件。
