# 关联数据清理机制

## 概述

新的数据清理机制实现了基于受保护地址的关联性数据清理，确保相关表之间的数据一致性。**基于受保护地址找到相关的交易哈希和高度，这些数据都不会被删除**，从而保证数据的完整性和一致性。

## 核心改进

### 1. 基于地址的关联性保护

- **地址级保护**：用户地址相关的交易数据会被保护
- **交易哈希保护**：基于受保护地址找到相关的交易哈希，这些交易哈希的所有相关表数据都会被保护
- **高度级保护**：基于受保护地址找到相关的高度，这些高度的所有相关表数据都会被保护
- **精确保护**：只保护真正需要的数据，避免过度保护

### 2. 清理策略

#### 原始策略（问题）
```go
// 分别清理每张表，可能导致数据不一致
cleanupTransactions()      // 可能清理了高度100的交易
cleanupTransactionReceipts() // 但高度100的收据可能被保留
cleanupContractParseResults() // 高度100的解析结果可能被保留
```

#### 新策略（解决方案）
```go
// 1. 基于受保护地址获取保护数据范围
protectedData := getProtectedData() // 基于受保护地址获取保护数据
// protectedData.Addresses: 用户地址
// protectedData.TxHashes: 基于受保护地址找到的交易哈希
// protectedData.Heights: 基于受保护地址找到的高度

// 2. 基于精确保护范围进行关联清理
cleanupTransactionsWithProtection()      // 排除受保护的高度、交易哈希和地址
cleanupTransactionReceiptsWithProtection() // 排除受保护的高度、交易哈希和地址  
cleanupContractParseResultsWithProtection() // 排除受保护的高度、交易哈希和地址
cleanupBlocksWithProtection()            // 排除受保护的高度
```

## 实现细节

### 1. 基于地址的保护高度获取

```go
func (s *DataCleanupScheduler) getProtectedHeights(tx *gorm.DB, chain string, cleanupHeight uint64) ([]uint64, error) {
    // 1. 获取受保护的地址
    addresses, err := s.getProtectedAddressesFromDB(tx)
    if err != nil {
        return nil, err
    }

    // 2. 基于受保护地址找到相关的高度
    heights, err := s.getProtectedHeightsByAddresses(tx, chain, addresses, cleanupHeight)
    if err != nil {
        return nil, err
    }

    return heights, nil
}
```

### 2. 关联清理实现

每个表的清理方法都会：
1. 排除受保护的高度范围
2. 排除受保护地址相关的数据
3. 确保数据一致性

```go
func (s *DataCleanupScheduler) cleanupTransactionsWithProtection(tx *gorm.DB, chain string, cleanupHeight uint64, protectedHeights []uint64, protectedAddresses []string, config *DataCleanupConfig) error {
    conditions := []string{"chain = ?", "height < ?"}
    args := []interface{}{chain, cleanupHeight}

    // 排除受保护的高度
    if len(protectedHeights) > 0 {
        conditions = append(conditions, "height NOT IN ?")
        args = append(args, protectedHeights)
    }

    // 排除受保护地址相关的交易
    if len(protectedAddresses) > 0 {
        conditions = append(conditions, "id NOT IN (SELECT DISTINCT t.id FROM transactions t WHERE t.chain = ? AND t.height < ? AND (t.from_address IN ? OR t.to_address IN ?))")
        args = append(args, chain, cleanupHeight, protectedAddresses, protectedAddresses)
    }

    whereClause := strings.Join(conditions, " AND ")
    return s.batchDelete(tx, "transactions", whereClause, args, config.BatchSize)
}
```

## 数据一致性保证

### 表关联关系

```
blocks.height = transactions.height = transaction_receipts.block_number = contract_parse_results.block_number
transactions.tx_id = transaction_receipts.tx_hash = contract_parse_results.tx_hash
```

### 保护规则

1. **地址保护**：如果交易涉及受保护的用户地址，则该交易及其相关数据都会被保护
2. **高度保护**：基于受保护地址找到的高度，这些高度的所有相关表数据都会被保护
3. **级联保护**：同一高度的数据要么全部保留，要么全部清理
4. **精确保护**：只保护真正需要的数据，避免过度保护

## 使用示例

```go
// 创建清理调度器
scheduler := NewDataCleanupScheduler(db, userAddressRepo, blockRepo, txRepo, receiptRepo, contractParseRepo)

// 设置清理配置
config := &DataCleanupConfig{
    Chain:            "eth",
    MaxBlocks:        50000,
    CleanupThreshold: 50000,
    BatchSize:        1000,
    Interval:         60,
}
scheduler.SetConfig("eth", config)

// 启动清理调度器
ctx := context.Background()
scheduler.Start(ctx)
```

## 优势

1. **数据一致性**：确保相关表之间的数据保持一致
2. **全面保护**：基于所有表的数据进行保护，避免数据丢失
3. **智能识别**：自动识别需要保护的高度、交易哈希和地址范围
4. **灵活配置**：支持高度保护、交易哈希保护和地址保护三重机制
5. **性能优化**：批量删除，避免长时间锁表
6. **事务安全**：所有清理操作在事务中执行，确保原子性

## 注意事项

1. 清理操作会跳过受保护的高度、交易哈希和地址
2. 建议在低峰期执行清理任务
3. 监控清理日志，确保清理效果符合预期
4. 定期检查受保护数据的大小，避免数据过度积累
5. 保护机制基于所有表的数据，确保不会意外删除重要数据
