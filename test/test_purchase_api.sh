#!/bin/bash
#
# 功能：采购订单API测试脚本
# 创建时间：2026-04-18
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

echo "=== 采购订单API测试 ==="
echo ""

echo "1. 登录获取Token..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')
assert_success "$LOGIN_RESPONSE" "登录"
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "✗ 登录失败：Token为空"
  echo "响应: $LOGIN_RESPONSE"
  exit 1
fi
echo "✓ Token获取成功"
echo ""

echo "2. 创建采购订单..."
echo "   获取一个可用物料..."
MATERIAL_RESPONSE=$(curl -s -X GET "$BASE_URL/base/materials?page=1&page_size=1" \
  -H "Authorization: Bearer $TOKEN")
assert_success "$MATERIAL_RESPONSE" "获取物料"
MATERIAL_ID=$(echo "$MATERIAL_RESPONSE" | jq -r '.data.list[0].id')
MATERIAL_NAME=$(echo "$MATERIAL_RESPONSE" | jq -r '.data.list[0].material_name')
if [ -z "$MATERIAL_ID" ] || [ "$MATERIAL_ID" = "null" ]; then
  echo "✗ 未找到可用物料"
  echo "响应: $MATERIAL_RESPONSE"
  exit 1
fi
echo "✓ 物料: $MATERIAL_NAME, material_id=$MATERIAL_ID"
echo "   获取一个可用供应商..."
SUPPLIER_RESPONSE=$(curl -s -X GET "$BASE_URL/base/suppliers?page=1&page_size=1" \
  -H "Authorization: Bearer $TOKEN")
assert_success "$SUPPLIER_RESPONSE" "获取供应商"
SUPPLIER_CODE=$(echo "$SUPPLIER_RESPONSE" | jq -r '.data.list[0].supplier_code')
SUPPLIER_NAME=$(echo "$SUPPLIER_RESPONSE" | jq -r '.data.list[0].supplier_name')
if [ -z "$SUPPLIER_CODE" ] || [ "$SUPPLIER_CODE" = "null" ]; then
  echo "✗ 未找到可用供应商"
  echo "响应: $SUPPLIER_RESPONSE"
  exit 1
fi
echo "✓ 供应商: $SUPPLIER_NAME, supplier_code=$SUPPLIER_CODE"
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/purchase/orders" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "supplier_code": "'"$SUPPLIER_CODE"'",
    "order_date": "2026-04-18",
    "expected_date": "2026-04-25",
    "remark": "API测试采购订单",
    "items": [
      {
        "material_id": '"$MATERIAL_ID"',
        "quantity": 50,
        "unit": "件",
        "unit_price": 100.00,
        "remark": "行备注"
      }
    ]
  }')
assert_success "$CREATE_RESPONSE" "创建采购订单"
ORDER_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.id')
if [ -z "$ORDER_ID" ] || [ "$ORDER_ID" = "null" ]; then
  echo "✗ 创建采购订单失败：ID为空"
  echo "响应: $CREATE_RESPONSE"
  exit 1
fi
echo "✓ 采购订单创建成功, ID: $ORDER_ID"
echo ""

echo "3. 查询采购订单列表..."
LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/purchase/orders?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN")
assert_success "$LIST_RESPONSE" "查询采购订单列表"
TOTAL=$(echo "$LIST_RESPONSE" | jq -r '.data.total')
echo "✓ 查询成功, 共 $TOTAL 条记录"
echo ""

echo "4. 查询采购订单详情..."
DETAIL_RESPONSE=$(curl -s -X GET "$BASE_URL/purchase/orders/$ORDER_ID" \
  -H "Authorization: Bearer $TOKEN")
assert_success "$DETAIL_RESPONSE" "查询采购订单详情"
ORDER_NO=$(echo "$DETAIL_RESPONSE" | jq -r '.data.order_no')
TOTAL_AMOUNT=$(echo "$DETAIL_RESPONSE" | jq -r '.data.total_amount')
if [ -z "$ORDER_NO" ] || [ "$ORDER_NO" = "null" ]; then
  echo "✗ 采购订单详情缺少订单号"
  echo "响应: $DETAIL_RESPONSE"
  exit 1
fi
echo "✓ 订单详情: $ORDER_NO, 总金额: ¥$TOTAL_AMOUNT"
echo ""

echo "5. 更新采购订单..."
UPDATE_RESPONSE=$(curl -s -X PUT "$BASE_URL/purchase/orders/$ORDER_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "expected_date": "2026-04-30",
    "remark": "已更新备注"
  }')
assert_success "$UPDATE_RESPONSE" "更新采购订单"
echo "✓ 采购订单更新成功"
echo ""

echo "6. 确认采购订单..."
CONFIRM_RESPONSE=$(curl -s -X POST "$BASE_URL/purchase/orders/$ORDER_ID/confirm" \
  -H "Authorization: Bearer $TOKEN")
assert_success "$CONFIRM_RESPONSE" "确认采购订单"
echo "✓ 采购订单确认成功"
echo ""

echo "=== 测试完成 ==="
