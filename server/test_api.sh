#!/bin/bash

# 区块链浏览器API测试脚本

BASE_URL="http://localhost:8080"
API_BASE="$BASE_URL/api/v1"

echo "=== 区块链浏览器API测试 ==="
echo "基础URL: $BASE_URL"
echo ""

# 测试健康检查
echo "1. 测试健康检查..."
curl -s "$BASE_URL/health" | jq .
echo ""

# 测试获取区块列表
echo "2. 测试获取区块列表..."
curl -s "$API_BASE/blocks?page=1&page_size=5" | jq .
echo ""

# 测试获取最新区块
echo "3. 测试获取最新区块..."
curl -s "$API_BASE/blocks/latest?chain=btc" | jq .
echo ""

# 测试获取交易列表
echo "4. 测试获取交易列表..."
curl -s "$API_BASE/transactions?page=1&page_size=5" | jq .
echo ""

# 测试根据哈希获取区块（使用示例哈希）
echo "5. 测试根据哈希获取区块..."
curl -s "$API_BASE/blocks/hash/0000000000000000000000000000000000000000000000000000000000000000" | jq .
echo ""

# 测试根据高度获取区块
echo "6. 测试根据高度获取区块..."
curl -s "$API_BASE/blocks/height/123456" | jq .
echo ""

# 测试根据地址获取交易
echo "7. 测试根据地址获取交易..."
curl -s "$API_BASE/transactions/address/1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa?page=1&page_size=5" | jq .
echo ""

echo "=== 测试完成 ===" 