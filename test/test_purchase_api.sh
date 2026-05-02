#!/bin/bash
#
# 功能：采购订单API测试脚本
# 创建时间：2026-04-18
# 创建人：CodeArts Agent
#

set -e

BASE_URL="http://localhost:8080/api/v1"
TOKEN=""

echo "=== 采购订单API测试 ==="
echo ""

echo "1. 登录获取Token..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')
echo "✓ Token获取成功"
echo ""

echo "2. 创建采购订单..."
echo "   获取一个可用SKU..."
SKU_RESPONSE=$(curl -s -X GET "$BASE_URL/base/skus?page=1&page_size=1&status=enabled" \
  -H "Authorization: Bearer $TOKEN")
SKU_ID=$(echo "$SKU_RESPONSE" | jq -r '.data.list[0].id')
MATERIAL_ID=$(echo "$SKU_RESPONSE" | jq -r '.data.list[0].material_id')
SKU_NAME=$(echo "$SKU_RESPONSE" | jq -r '.data.list[0].sku_name')
SKU_CODE=$(echo "$SKU_RESPONSE" | jq -r '.data.list[0].sku_code')
echo "✓ SKU: $SKU_NAME [$SKU_CODE], sku_id=$SKU_ID, material_id=$MATERIAL_ID"
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/purchase/orders" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "supplier_id": 1,
    "order_date": "2026-04-18",
    "expected_date": "2026-04-25",
    "remark": "API测试采购订单",
    "items": [
      {
        "material_id": '"$MATERIAL_ID"',
        "sku_id": '"$SKU_ID"',
        "quantity": 50,
        "unit": "件",
        "unit_price": 100.00,
        "remark": "行备注"
      }
    ]
  }')
ORDER_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.id')
echo "✓ 采购订单创建成功, ID: $ORDER_ID"
echo ""

echo "3. 查询采购订单列表..."
LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/purchase/orders?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN")
TOTAL=$(echo "$LIST_RESPONSE" | jq -r '.data.total')
echo "✓ 查询成功, 共 $TOTAL 条记录"
echo ""

echo "4. 查询采购订单详情..."
DETAIL_RESPONSE=$(curl -s -X GET "$BASE_URL/purchase/orders/$ORDER_ID" \
  -H "Authorization: Bearer $TOKEN")
ORDER_NO=$(echo "$DETAIL_RESPONSE" | jq -r '.data.order_no')
TOTAL_AMOUNT=$(echo "$DETAIL_RESPONSE" | jq -r '.data.total_amount')
echo "✓ 订单详情: $ORDER_NO, 总金额: ¥$TOTAL_AMOUNT"
echo ""

echo "5. 更新采购订单..."
curl -s -X PUT "$BASE_URL/purchase/orders/$ORDER_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "expected_date": "2026-04-30",
    "remark": "已更新备注"
  }' > /dev/null
echo "✓ 采购订单更新成功"
echo ""

echo "6. 确认采购订单..."
curl -s -X POST "$BASE_URL/purchase/orders/$ORDER_ID/confirm" \
  -H "Authorization: Bearer $TOKEN" > /dev/null
echo "✓ 采购订单确认成功"
echo ""

echo "=== 测试完成 ==="
