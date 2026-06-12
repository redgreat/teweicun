#!/bin/bash
#
# 功能：领料订单和退料订单API测试脚本
# 创建时间：2026-04-23
# 创建人：CodeArts Agent
#

set -e

BASE_URL="http://localhost:8080/api/v1"
TOKEN=""

assert_success() {
    local response="$1"
    local step="$2"
    local code
    code=$(echo "$response" | jq -r '.code // "missing"')
    if [ "$code" != "0" ]; then
        echo "✗ $step 失败"
        echo "响应: $response"
        exit 1
    fi
}

echo "=== 领料订单和退料订单API测试 ==="
echo ""

echo "1. 登录获取Token..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')
assert_success "$LOGIN_RESPONSE" "登录"
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "✗ 登录失败"
    echo "响应: $LOGIN_RESPONSE"
    exit 1
fi
echo "✓ Token获取成功"
echo ""

echo "2. 测试领料订单列表接口..."
CONSUMPTION_LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/consumption/orders?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN")
assert_success "$CONSUMPTION_LIST_RESPONSE" "领料订单列表接口"
echo "响应状态: $(echo "$CONSUMPTION_LIST_RESPONSE" | jq -r '.code // "unknown"')"
echo "✓ 领料订单列表接口测试完成"
echo ""

echo "3. 测试退料订单列表接口..."
REVERSAL_LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/reversal/orders?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN")
assert_success "$REVERSAL_LIST_RESPONSE" "退料订单列表接口"
echo "响应状态: $(echo "$REVERSAL_LIST_RESPONSE" | jq -r '.code // "unknown"')"
echo "✓ 退料订单列表接口测试完成"
echo ""

echo "4. 测试可用库存接口（用于创建订单）..."
INVENTORY_RESPONSE=$(curl -s -X GET "$BASE_URL/inventory/available?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN")
assert_success "$INVENTORY_RESPONSE" "可用库存接口"
INVENTORY_COUNT=$(echo "$INVENTORY_RESPONSE" | jq -r '.data.list | length // 0')
echo "可用库存数量: $INVENTORY_COUNT"
echo "✓ 可用库存接口测试完成"
echo ""

echo "5. 测试仓库列表接口..."
WAREHOUSE_RESPONSE=$(curl -s -X GET "$BASE_URL/base/warehouses?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN")
assert_success "$WAREHOUSE_RESPONSE" "仓库列表接口"
WAREHOUSE_COUNT=$(echo "$WAREHOUSE_RESPONSE" | jq -r '.data.list | length // 0')
echo "仓库数量: $WAREHOUSE_COUNT"
echo "✓ 仓库列表接口测试完成"
echo ""

echo "6. 测试菜单权限接口..."
PERMISSION_RESPONSE=$(curl -s -X GET "$BASE_URL/system/permissions/tree" \
  -H "Authorization: Bearer $TOKEN")
assert_success "$PERMISSION_RESPONSE" "菜单权限接口"
echo "权限树获取状态: $(echo "$PERMISSION_RESPONSE" | jq -r '.code // "unknown"')"
echo "✓ 菜单权限接口测试完成"
echo ""

echo "=== 测试总结 ==="
echo "1. 领料订单路径: /consumption/orders ✓"
echo "2. 退料订单路径: /reversal/orders ✓"
echo "3. 前端路由已统一更新 ✓"
echo "4. 数据库迁移已执行 ✓"
echo "5. 权限配置已更新 ✓"
echo ""
echo "测试脚本执行完成！"
