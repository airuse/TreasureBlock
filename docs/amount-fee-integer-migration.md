# 用户交易表金额字段整数化迁移文档

## 概述

将用户交易信息表（`user_transactions`）中的 `amount` 和 `fee` 字段从带小数的 `decimal` 类型改为整数类型，以符合数字货币没有小数概念的设计原则。

## 数据库变更

### 字段类型修改

- `amount`: `decimal(65,18)` → `decimal(65,0)`
- `fee`: `decimal(36,18)` → `decimal(65,0)`

### 迁移脚本

创建了迁移脚本：`server/sql/migrations/update_user_transactions_amount_fee_to_integer.sql`

```sql
-- 修改 amount 字段类型
ALTER TABLE `user_transactions` 
MODIFY COLUMN `amount` decimal(65,0) NOT NULL COMMENT '交易金额';

-- 修改 fee 字段类型  
ALTER TABLE `user_transactions` 
MODIFY COLUMN `fee` decimal(65,0) NOT NULL DEFAULT '0' COMMENT '手续费';
```

## 后端代码变更

### 1. 模型定义更新

**文件**: `server/internal/models/user_transaction.go`

```go
Amount string `json:"amount" gorm:"type:decimal(65,0);not null;comment:交易金额"`
Fee    string `json:"fee" gorm:"type:decimal(65,0);not null;default:0;comment:手续费"`
```

### 2. 服务层更新

**文件**: `server/internal/services/user_transaction_service.go`

- 添加了 `math/big` 包导入
- 更新了 `convertAmountToHex` 函数，使用大整数处理金额转换

```go
func (s *userTransactionService) convertAmountToHex(amount string) string {
    amountBig, ok := new(big.Int).SetString(amount, 10)
    if !ok {
        return "0x0"
    }
    return fmt.Sprintf("0x%s", amountBig.Text(16))
}
```

### 3. 处理器更新

**文件**: `server/internal/handlers/user_transaction_handler.go`

- 添加了 `math/big` 包导入
- 添加了 `validateAmountFormat` 函数，验证金额格式
- 在交易验证中添加了金额格式检查

```go
func (h *UserTransactionHandler) validateAmountFormat(amount string) error {
    if amount == "" {
        return fmt.Errorf("金额不能为空")
    }
    
    amountBig, ok := new(big.Int).SetString(amount, 10)
    if !ok {
        return fmt.Errorf("金额必须是有效的整数格式")
    }
    
    if amountBig.Sign() < 0 {
        return fmt.Errorf("金额不能为负数")
    }
    
    return nil
}
```

## 前端代码变更

### 1. 交易列表页面

**文件**: `vue/src/views/eth/personal/TransactionsView.vue`

#### 金额格式化函数更新

```typescript
// 格式化金额 - 处理整数金额显示
const formatAmount = (amount: string, symbol: string, decimals: number | undefined) => {
  const intAmount = BigInt(amount)
  if (intAmount === 0n) return '0'
  
  // 根据币种和精度进行转换
  if (decimals !== undefined && decimals >= 0) {
    const factor = BigInt(Math.pow(10, decimals).toString())
    const readableAmount = Number(intAmount) / Number(factor)
    return readableAmount.toFixed(Math.min(decimals, 8))
  }
  
  // 币种特定的精度处理
  if (symbol === 'ETH') {
    const factor = BigInt('1000000000000000000') // 10^18
    const readableAmount = Number(intAmount) / Number(factor)
    return readableAmount.toFixed(8)
  }
  // ... 其他币种处理
}
```

#### 十六进制转换函数更新

```typescript
// 转换金额为十六进制格式
const convertToHexString = (amount: string) => {
  const intAmount = BigInt(amount)
  const hexString = intAmount.toString(16)
  return '0x' + hexString
}
```

### 2. 创建交易模态框

**文件**: `vue/src/components/eth/personal/CreateTransactionModal.vue`

#### 金额转换函数更新

```typescript
// 转换为wei单位 - 处理整数金额
const formatToWei = (amount: string) => {
  if (!amount) return '0'
  const num = parseFloat(amount)
  if (isNaN(num)) return '0'
  const wei = Math.floor(num * Math.pow(10, 18))
  return wei.toString()
}

// 转换为代币最小单位 - 处理整数金额
const formatToTokenUnits = (amount: string, decimals: number) => {
  if (!amount) return '0'
  const num = parseFloat(amount)
  if (isNaN(num)) return '0'
  const units = Math.floor(num * Math.pow(10, decimals))
  return units.toString()
}
```

#### 编辑模式初始化更新

```typescript
// 编辑模式下初始化表单数据
const initEditForm = () => {
  if (props.isEditMode && props.transaction) {
    const tx = props.transaction
    
    // 将整数金额转换为显示格式
    let displayAmount = ''
    if (tx.amount) {
      if (tx.transaction_type === 'token' && tx.token_decimals !== undefined) {
        const intAmount = BigInt(tx.amount)
        const factor = BigInt(Math.pow(10, tx.token_decimals).toString())
        displayAmount = (Number(intAmount) / Number(factor)).toString()
      } else if (tx.symbol === 'ETH') {
        const intAmount = BigInt(tx.amount)
        const factor = BigInt('1000000000000000000') // 10^18
        displayAmount = (Number(intAmount) / Number(factor)).toString()
      }
    }
    // ... 其他初始化逻辑
  }
}
```

### 3. 发送交易模态框

**文件**: `vue/src/components/eth/personal/SendTransactionModal.vue`

#### 手续费计算更新

```typescript
// 计算属性 - 返回整数格式的手续费
const calculateAutoMinerFee = computed(() => {
  const gasPrice = autoFeeRates[autoFeeSpeed.value]
  const gasLimit = props.transaction.gas_limit || 21000
  const fee = (gasPrice * gasLimit) / 1e9
  return Math.floor(fee * 1e18).toString() // 转换为wei单位（整数）
})
```

## 数据迁移注意事项

### 1. 现有数据处理

- 现有数据库中的小数金额需要转换为整数
- 例如：1.5 ETH = 1500000000000000000 wei（整数）

### 2. 精度处理

- ETH: 18位精度（1 ETH = 10^18 wei）
- USDC/USDT: 6位精度（1 USDC = 10^6 最小单位）
- DAI: 18位精度（1 DAI = 10^18 最小单位）

### 3. 前端显示

- 用户界面仍然显示可读的金额格式（如 1.5 ETH）
- 后端存储和传输使用整数格式（如 1500000000000000000）

## 测试建议

### 1. 数据库测试

- 验证字段类型修改成功
- 测试大整数金额的存储和读取

### 2. 前端测试

- 测试金额输入和显示
- 测试不同币种的精度处理
- 测试编辑模式下的金额转换

### 3. 后端测试

- 测试金额验证逻辑
- 测试十六进制转换
- 测试大整数运算

## 部署步骤

1. **备份数据库**
   ```sql
   CREATE TABLE user_transactions_backup AS SELECT * FROM user_transactions;
   ```

2. **执行迁移脚本**
   ```bash
   mysql -u username -p database_name < server/sql/migrations/update_user_transactions_amount_fee_to_integer.sql
   ```

3. **部署后端代码**
   - 确保所有Go代码编译通过
   - 测试API接口

4. **部署前端代码**
   - 确保所有TypeScript代码编译通过
   - 测试用户界面

5. **验证功能**
   - 创建新交易
   - 编辑现有交易
   - 发送交易
   - 查看交易历史

## 回滚方案

如果出现问题，可以回滚到原始的小数格式：

```sql
-- 回滚 amount 字段
ALTER TABLE `user_transactions` 
MODIFY COLUMN `amount` decimal(65,18) NOT NULL COMMENT '交易金额';

-- 回滚 fee 字段
ALTER TABLE `user_transactions` 
MODIFY COLUMN `fee` decimal(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '手续费';
```

然后恢复之前的代码版本。
