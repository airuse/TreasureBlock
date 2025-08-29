# 合约地址功能说明

## 概述

本次更新为个人地址管理功能添加了合约关联支持，允许用户在添加地址时选择关联的智能合约。

## 新增功能

### 1. 数据库变更

- 在 `user_addresses` 表中添加了 `contract_id` 字段
- 创建了相应的索引以提高查询性能
- 添加了与 `contracts` 表的外键关联

### 2. 后端API更新

#### 模型更新
- `UserAddress` 模型新增 `ContractID` 字段
- 添加了与 `Contract` 模型的关联关系

#### DTO更新
- `CreateUserAddressRequest` 新增 `contract_id` 字段
- `UpdateUserAddressRequest` 新增 `contract_id` 字段
- `UserAddressResponse` 新增 `contract_id` 字段

#### 服务层更新
- 创建地址时支持保存合约ID
- 更新地址时支持修改合约ID
- 当地址类型从合约改为其他类型时，自动清空合约ID

### 3. 前端界面更新

#### 添加地址模态框
- 当选择地址类型为"合约"时，显示合约选择器
- 合约选择器显示合约名称、符号和地址
- 支持搜索和分页加载合约列表

#### 编辑地址模态框
- 支持修改已关联的合约
- 当类型改变时自动清空合约ID

#### 地址列表显示
- 新增"合约"列显示关联的合约信息
- 合约信息以标签形式展示，包含名称和符号

## 使用方法

### 1. 添加合约地址

1. 点击"添加地址"按钮
2. 填写地址信息
3. 选择类型为"合约"
4. 从下拉列表中选择要关联的合约
5. 填写标签和备注信息
6. 点击"添加"保存

### 2. 编辑合约地址

1. 在地址列表中点击"编辑"按钮
2. 修改标签、类型或关联的合约
3. 点击"更新"保存修改

### 3. 查看合约信息

- 在地址列表中，"合约"列会显示关联的合约名称
- 如果地址类型不是合约，则显示"-"

## 技术细节

### 数据验证

- 当类型为"合约"时，`contract_id` 为必填字段
- 当类型不是"合约"时，`contract_id` 自动清空
- 支持合约ID的更新和删除操作

### 性能优化

- 合约列表支持分页加载
- 添加了数据库索引以提高查询性能
- 前端实现了智能的合约选择器

### 错误处理

- 合约加载失败时显示友好的错误提示
- 表单验证确保数据的完整性
- 支持合约关联的撤销和重新关联

## 数据库迁移

### 执行迁移

```bash
# 应用迁移（按顺序执行）
mysql -u username -p database_name < server/sql/migrations/000007_add_contract_id_to_user_addresses.up.sql
mysql -u username -p database_name < server/sql/migrations/000008_add_notes_to_user_addresses.up.sql

# 回滚迁移（如果需要，按相反顺序执行）
mysql -u username -p database_name < server/sql/migrations/000008_add_notes_to_user_addresses.down.sql
mysql -u username -p database_name < server/sql/migrations/000007_add_contract_id_to_user_addresses.down.sql
```

### 验证迁移

```bash
# 运行测试脚本验证所有字段
mysql -u username -p database_name < server/scripts/test_all_fields.sql
```

## 注意事项

1. **数据一致性**: 确保关联的合约ID在 `contracts` 表中存在
2. **权限控制**: 用户只能关联自己有权限访问的合约
3. **性能考虑**: 大量合约数据时建议实现虚拟滚动或搜索功能
4. **向后兼容**: 现有地址数据不受影响，`contract_id` 字段为可选

## 未来扩展

1. **批量操作**: 支持批量关联/取消关联合约
2. **合约搜索**: 实现更智能的合约搜索和过滤
3. **合约详情**: 在地址列表中直接显示合约的详细信息
4. **多链支持**: 扩展到支持其他区块链的合约关联
