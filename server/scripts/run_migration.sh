#!/bin/bash

# Gas费用单位迁移脚本
# 将数据库中的max_priority_fee_per_gas和max_fee_per_gas从Gwei转换为Wei

echo "🚀 开始执行Gas费用单位迁移..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到Go环境，请先安装Go"
    exit 1
fi

# 进入脚本目录
cd "$(dirname "$0")"

# 编译迁移程序
echo "📦 编译迁移程序..."
go mod init migrate_gas_fees 2>/dev/null || true
go mod tidy
go build -o migrate_gas_fees migrate_gas_fees.go

if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

# 执行迁移
echo "🔄 执行数据迁移..."
./migrate_gas_fees

# 清理编译文件
rm -f migrate_gas_fees migrate_gas_fees.go go.mod go.sum

echo "✅ 迁移完成！"
