# SOL ATA Token余额更新功能

## 概述

本功能为SOL链的ATA（Associated Token Account）账户添加了Token余额自动更新功能。当用户创建或刷新SOL ATA地址时，系统会自动获取并更新对应的SPL Token余额。

## 功能特性

1. **自动获取Token余额**：在创建SOL ATA地址时，系统会自动获取对应Token的余额
2. **余额刷新**：通过`RefreshAddressBalances`接口可以刷新SOL ATA的Token余额
3. **故障转移**：使用多个RPC节点确保高可用性
4. **精确匹配**：根据`AtaMintAddress`精确匹配对应的Token账户

## 使用方法

### 1. 创建SOL ATA地址

```json
{
  "address": "ATA账户地址",
  "chain": "sol",
  "label": "USDC ATA",
  "type": "ata",
  "ata_owner_address": "主钱包地址",
  "ata_mint_address": "USDC的mint地址"
}
```

### 2. 刷新余额

调用 `RefreshAddressBalances` 接口，系统会：
- 获取原生SOL余额（lamports）
- 获取指定mint的Token余额
- 更新`ContractBalance`字段

## 技术实现

### 新增方法

1. **SolFailoverManager.GetTokenAccountsByOwner()**
   - 调用Solana RPC的`getTokenAccountsByOwner`方法
   - 支持按mint地址过滤
   - 返回Token账户的详细信息

2. **TokenAccountInfo结构体**
   ```go
   type TokenAccountInfo struct {
       Address  string   `json:"address"`
       Mint     string   `json:"mint"`
       Owner    string   `json:"owner"`
       Amount   string   `json:"amount"`
       Decimals int      `json:"decimals"`
       UIAmount *float64 `json:"uiAmount,omitempty"`
   }
   ```

### 更新逻辑

1. **CreateAddress方法**
   - 检测SOL链和ATA类型
   - 自动获取Token余额
   - 存储到`ContractBalance`字段

2. **RefreshAddressBalances方法**
   - 刷新原生SOL余额
   - 刷新Token余额
   - 更新数据库记录

## 数据存储

- **Balance字段**：存储原生SOL余额（lamports）
- **ContractBalance字段**：存储SPL Token余额（最小单位）
- **AtaOwnerAddress字段**：存储主钱包地址
- **AtaMintAddress字段**：存储Token的mint地址

## 注意事项

1. 确保RPC节点支持`getTokenAccountsByOwner`方法
2. Token余额以最小单位存储，需要根据decimals进行转换
3. 如果Token账户不存在，`ContractBalance`将为空
4. 系统会自动处理RPC故障转移

## 示例响应

```json
{
  "id": 123,
  "address": "ATA账户地址",
  "chain": "sol",
  "label": "USDC ATA",
  "type": "ata",
  "balance": "5000000",
  "contract_balance": "1000000000",
  "ata_owner_address": "主钱包地址",
  "ata_mint_address": "USDC的mint地址"
}
```

其中：
- `balance`：原生SOL余额（lamports）
- `contract_balance`：USDC Token余额（最小单位）
