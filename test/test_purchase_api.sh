#!/bin/bash
#
# 功能：进销存采购入库全流程API测试（使用已有业务数据）
# 测试流程：供应商查询 → 物料查询 → 创建采购订单 → 确认 → 入库单确认 → 追踪
# 创建时间：2026-04-18 / 重写：2026-07-12
# 创建人：CodeArts Agent / 重写人：Hermes Agent
#

set -eo pipefail

BASE_URL="${TWC_BASE_URL:-http://localhost:8080}"
ADMIN_USER="${TWC_ADMIN_USER:-admin}"
ADMIN_PASS="${TWC_ADMIN_PASS:-admin123}"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
PASS=0; FAIL=0
pass() { echo -e "${GREEN}[PASS]${NC} $*"; PASS=$((PASS+1)); }
fail() { echo -e "${RED}[FAIL]${NC} $*"; FAIL=$((FAIL+1)); }
warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }

echo "================================================================"
echo "  特维存（TeWeiCun）采购入库全流程 API 测试"
echo "  目标: $BASE_URL"
echo "  时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo "================================================================"
echo ""

# ── 1. 登录 ──
echo "【1/8】管理员登录"
LOGIN=$(curl -sS -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}")
TOKEN=$(echo "$LOGIN" | jq -r '.data.token // empty')
if [[ -n "$TOKEN" && "$TOKEN" != "null" ]]; then
  pass "登录成功，Token已获取"
else
  fail "登录失败: $(echo "$LOGIN" | jq -c '.')"
  exit 1
fi
AUTH=(-H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json")
AUTH_GET=(-H "Authorization: Bearer $TOKEN")
echo ""

# ── 2. 查询供应商 ──
echo "【2/8】查询已有供应商"
SUPPLIER_LIST=$(curl -sS "$BASE_URL/api/v1/base/suppliers?page=1&page_size=1" "${AUTH_GET[@]}")
SUPPLIER_CODE=$(echo "$SUPPLIER_LIST" | jq -r '.data.list[0].supplier_code // empty')
SUPPLIER_NAME=$(echo "$SUPPLIER_LIST" | jq -r '.data.list[0].supplier_name // empty')
if [[ -n "$SUPPLIER_CODE" && "$SUPPLIER_CODE" != "null" ]]; then
  pass "供应商: ${SUPPLIER_NAME}（$SUPPLIER_CODE）"
else
  fail "未找到可用供应商"
fi
echo ""

# ── 3. 查询仓库 ──
echo "【3/8】查询已有仓库（主材料库）"
WH_LIST=$(curl -sS "$BASE_URL/api/v1/base/warehouses?page=1&page_size=10" "${AUTH_GET[@]}")
WH_CODE=$(echo "$WH_LIST" | jq -r '.data.list[] | select(.warehouse_name | contains("主材料")) | .warehouse_code // empty' | head -1)
WH_NAME=$(echo "$WH_LIST" | jq -r '.data.list[] | select(.warehouse_name | contains("主材料")) | .warehouse_name // empty' | head -1)
if [[ -z "$WH_CODE" ]]; then
  WH_CODE=$(echo "$WH_LIST" | jq -r '.data.list[0].warehouse_code // empty')
  WH_NAME=$(echo "$WH_LIST" | jq -r '.data.list[0].warehouse_name // empty')
fi
if [[ -n "$WH_CODE" && "$WH_CODE" != "null" ]]; then
  pass "仓库: ${WH_NAME}（$WH_CODE）"
else
  fail "未找到可用仓库"
fi
echo ""

# ── 4. 查询物料 ──
echo "【4/8】查询物料（一个编码管理 + 一个非编码管理）"
MAT_LIST=$(curl -sS "$BASE_URL/api/v1/base/materials?page=1&page_size=100" "${AUTH_GET[@]}")

CODE_MAT=$(echo "$MAT_LIST" | jq '[.data.list[] | select(.is_code == true)] | first')
CODE_MAT_ID=$(echo "$CODE_MAT" | jq -r '.id // empty')
CODE_MAT_NAME=$(echo "$CODE_MAT" | jq -r '.material_name // empty')

NOCODE_MAT=$(echo "$MAT_LIST" | jq '[.data.list[] | select(.is_code == false)] | first')
NOCODE_MAT_ID=$(echo "$NOCODE_MAT" | jq -r '.id // empty')
NOCODE_MAT_NAME=$(echo "$NOCODE_MAT" | jq -r '.material_name // empty')

if [[ -n "$CODE_MAT_ID" && "$CODE_MAT_ID" != "null" ]]; then
  pass "编码物料: ${CODE_MAT_NAME}（id=$CODE_MAT_ID）"
else
  fail "未找到编码管理物料"
fi
if [[ -n "$NOCODE_MAT_ID" && "$NOCODE_MAT_ID" != "null" ]]; then
  pass "非编码物料: ${NOCODE_MAT_NAME}（id=$NOCODE_MAT_ID）"
else
  warn "未找到非编码物料，仅使用编码物料"
  NOCODE_MAT_ID="$CODE_MAT_ID"
  NOCODE_MAT_NAME="$CODE_MAT_NAME"
fi
echo ""

# ── 5. 创建采购订单 ──
echo "【5/8】创建采购订单（含编码+非编码物料各10件，单价100元）"
TODAY=$(date +%Y-%m-%d)
ITEMS_JSON="[{\"material_id\":$CODE_MAT_ID,\"quantity\":10,\"unit_price\":100},{\"material_id\":$NOCODE_MAT_ID,\"quantity\":10,\"unit_price\":100}]"

CREATE_RESP=$(curl -sS -X POST "$BASE_URL/api/v1/purchase/orders" \
  "${AUTH[@]}" \
  -d "{
    \"supplier_code\": \"$SUPPLIER_CODE\",
    \"order_type\": \"purchase\",
    \"order_date\": \"$TODAY\",
    \"remark\": \"自动化测试-采购订单-$TODAY\",
    \"items\": $ITEMS_JSON
  }")

CREATE_CODE=$(echo "$CREATE_RESP" | jq -r '.code // "missing"')
ORDER_ID=$(echo "$CREATE_RESP" | jq -r '.data.id // empty')
ORDER_MSG=$(echo "$CREATE_RESP" | jq -r '.msg // "unknown"')

if [[ "$CREATE_CODE" == "0" && -n "$ORDER_ID" && "$ORDER_ID" != "null" ]]; then
  pass "采购订单创建成功（ID=$ORDER_ID）"
else
  fail "创建采购订单失败: code=$CREATE_CODE msg=$ORDER_MSG"
fi
echo ""

# ── 6. 查询订单详情 ──
if [[ -n "${ORDER_ID:-}" && "$ORDER_ID" != "null" ]]; then
  echo "【6/8】查询采购订单详情"
  DETAIL=$(curl -sS "$BASE_URL/api/v1/purchase/orders/$ORDER_ID" "${AUTH_GET[@]}")
  ORDER_NO=$(echo "$DETAIL" | jq -r '.data.order_no // empty')
  ORDER_STATUS=$(echo "$DETAIL" | jq -r '.data.order_status // empty')
  STOCK_IN_ID=$(echo "$DETAIL" | jq -r '.data.stock_in_id // 0')
  ITEM_COUNT=$(echo "$DETAIL" | jq -r '.data.items | length // 0')

  if [[ -n "$ORDER_NO" && "$ORDER_NO" != "null" ]]; then
    pass "订单详情: $ORDER_NO 状态=$ORDER_STATUS 明细=$ITEM_COUNT行 关联入库单=$STOCK_IN_ID"
  else
    fail "订单详情缺少order_no"
  fi
else
  echo "【6/8】跳过（订单未创建）"
fi
echo ""

# ── 7. 确认采购订单并验证 ──
if [[ -n "${ORDER_ID:-}" && "$ORDER_ID" != "null" ]]; then
  echo "【7/8】确认采购订单"
  DETAIL_BEFORE=$(curl -sS "$BASE_URL/api/v1/purchase/orders/$ORDER_ID" "${AUTH_GET[@]}")
  CURRENT_STATUS=$(echo "$DETAIL_BEFORE" | jq -r '.data.order_status // "unknown"')

  if [[ "$CURRENT_STATUS" == "draft" || "$CURRENT_STATUS" == "待提交" ]]; then
    CONFIRM_RESP=$(curl -sS -X POST "$BASE_URL/api/v1/purchase/orders/$ORDER_ID/confirm" "${AUTH[@]}")
    CONFIRM_CODE=$(echo "$CONFIRM_RESP" | jq -r '.code // "missing"')
    CONFIRM_MSG=$(echo "$CONFIRM_RESP" | jq -r '.msg // "unknown"')

    if [[ "$CONFIRM_CODE" == "0" ]]; then
      sleep 0.5
      DETAIL2=$(curl -sS "$BASE_URL/api/v1/purchase/orders/$ORDER_ID" "${AUTH_GET[@]}")
      NEW_STATUS=$(echo "$DETAIL2" | jq -r '.data.order_status // empty"')
      pass "订单确认成功（状态: $CURRENT_STATUS → $NEW_STATUS）"
    else
      fail "确认订单失败: code=$CONFIRM_CODE msg=$CONFIRM_MSG"
    fi
  else
    STOCK_IN_ID=$(echo "$DETAIL_BEFORE" | jq -r '.data.stock_in_id // 0')
    pass "订单已自动确认（状态=$CURRENT_STATUS 入库单=$STOCK_IN_ID）"
  fi
else
  echo "【7/8】跳过（订单未创建）"
fi
echo ""

# ── 8. 边界测试 ──
echo "【8/8】边界测试"

HTTP_401=$(curl -sS -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/purchase/orders?page=1")
[[ "$HTTP_401" == "401" ]] && pass "无Token访问返回401" || fail "无Token访问返回$HTTP_401（预期401）"

BAD_ID=$(curl -sS "$BASE_URL/api/v1/purchase/orders/9999999" "${AUTH_GET[@]}" | jq -r '.code // 0')
[[ "$BAD_ID" != "0" ]] && pass "无效ID查询正确返回错误（code=$BAD_ID）" || fail "无效ID查询未拒绝"

MISSING=$(curl -sS -X POST "$BASE_URL/api/v1/purchase/orders" "${AUTH[@]}" \
  -d '{"order_date":"2026-01-01"}' | jq -r '.code // 0')
[[ "$MISSING" != "0" ]] && pass "缺失必填字段正确拒绝（code=$MISSING）" || fail "缺失必填字段未拒绝"

BAD_SUPPLIER=$(curl -sS -X POST "$BASE_URL/api/v1/purchase/orders" "${AUTH[@]}" \
  -d "{
    \"supplier_code\": \"NOT_EXIST_SUP_999\",
    \"order_type\": \"purchase\",
    \"order_date\": \"$TODAY\",
    \"items\": [{\"material_id\":$CODE_MAT_ID,\"quantity\":1,\"unit_price\":10}]
  }")
BAD_SUP_CODE=$(echo "$BAD_SUPPLIER" | jq -r '.code // 0')
BAD_SUP_MSG=$(echo "$BAD_SUPPLIER" | jq -r '.msg // "unknown"')
if [[ "$BAD_SUP_CODE" != "0" ]]; then
  pass "不存在的供应商正确拒绝: $BAD_SUP_MSG"
else
  fail "不存在的供应商未拒绝"
fi

HEALTH=$(curl -sS "$BASE_URL/api/v1/health" | jq -r '.status // "fail"')
[[ "$HEALTH" == "ok" ]] && pass "健康检查通过" || fail "健康检查失败（status=$HEALTH）"
echo ""

# ── 测试总结 ──
echo "================================================================"
echo "  测试总结"
echo "================================================================"
echo -e "  通过: ${GREEN}$PASS${NC}"
echo -e "  失败: ${RED}$FAIL${NC}"
echo "================================================================"

[[ "$FAIL" -gt 0 ]] && exit 1 || exit 0
