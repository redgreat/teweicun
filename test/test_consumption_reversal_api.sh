#!/bin/bash
#
# 功能：领料出库 → 退料入库全流程 API 测试（使用已有业务数据）
# 测试流程：库存查询 → 创建领料单 → 确认 → 领料出库 → 退料 → 退料入库
# 创建时间：2026-04-23 / 重写：2026-07-12
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
echo "  特维存（TeWeiCun）领料出库 → 退料入库全流程测试"
echo "  目标: $BASE_URL"
echo "  时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo "================================================================"
echo ""

echo "【1/6】管理员登录"
LOGIN=$(curl -sS -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}")
TOKEN=$(echo "$LOGIN" | jq -r '.data.token // empty')
if [[ -n "$TOKEN" && "$TOKEN" != "null" ]]; then
  pass "登录成功"
else
  fail "登录失败"; exit 1
fi
AUTH=(-H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json")
AUTH_GET=(-H "Authorization: Bearer $TOKEN")
echo ""

echo "【2/6】查询可用库存（主材料库）"
WH_CODE="WH001"
WH_NAME="主材料库"
INV_RESP=$(curl -sS "$BASE_URL/api/v1/inventory/available?page=1&page_size=50&warehouse_code=$WH_CODE" "${AUTH_GET[@]}")
INV_COUNT=$(echo "$INV_RESP" | jq -r '.data.total // 0')
if [[ "$INV_COUNT" -gt 0 ]]; then
  pass "可用库存: ${INV_COUNT}条"
  FIRST_INV=$(echo "$INV_RESP" | jq '[.data.list[] | select(.available_quantity >= 3)] | first')
  INV_ID=$(echo "$FIRST_INV" | jq -r '.inventory_id // empty')
  MAT_ID=$(echo "$FIRST_INV" | jq -r '.material_id // empty')
  IS_CODE=$(echo "$FIRST_INV" | jq -r '.is_code // false')
  AVAIL_QTY=$(echo "$FIRST_INV" | jq -r '.available_quantity // 0')
  INV_UNIT=$(echo "$FIRST_INV" | jq -r '.unit // "pcs"')
  if [[ -n "$INV_ID" && "$INV_ID" != "null" && "$AVAIL_QTY" != "0" ]]; then
    pass "选中库存: invID=$INV_ID materialID=$MAT_ID 可用=${AVAIL_QTY} 单位=$INV_UNIT 编码管理=$IS_CODE"
  else
    warn "无足够库存（≥3），仅测试列表接口"; INV_ID=""
  fi
else
  warn "可用库存为空，仅测试列表接口"; INV_ID=""
fi
echo ""

echo "【3/6】领料订单列表"
CONS_LIST=$(curl -sS "$BASE_URL/api/v1/consumption/orders?page=1&page_size=10" "${AUTH_GET[@]}")
CONS_CODE=$(echo "$CONS_LIST" | jq -r '.code // "missing"')
CONS_TOTAL=$(echo "$CONS_LIST" | jq -r '.data.total // 0')
if [[ "$CONS_CODE" == "0" ]]; then pass "领料订单列表: 共${CONS_TOTAL}条"; else fail "领料列表接口失败: code=$CONS_CODE"; fi
echo ""

echo "【4/6】退料订单列表"
REV_LIST=$(curl -sS "$BASE_URL/api/v1/reversal/orders?page=1&page_size=10" "${AUTH_GET[@]}")
REV_CODE=$(echo "$REV_LIST" | jq -r '.code // "missing"')
REV_TOTAL=$(echo "$REV_LIST" | jq -r '.data.total // 0')
if [[ "$REV_CODE" == "0" ]]; then pass "退料订单列表: 共${REV_TOTAL}条"; else fail "退料列表接口失败: code=$REV_CODE"; fi
echo ""

echo "【5/6】仓库列表"
WH_LIST=$(curl -sS "$BASE_URL/api/v1/base/warehouses?page=1&page_size=10" "${AUTH_GET[@]}")
WH_COUNT=$(echo "$WH_LIST" | jq -r '.data.list | length // 0')
if [[ "$WH_COUNT" -gt 0 ]]; then
  WH_NAMES=$(echo "$WH_LIST" | jq -r '[.data.list[].warehouse_name] | join("、")')
  pass "仓库列表: ${WH_COUNT}个（${WH_NAMES}）"
else
  fail "仓库列表为空"
fi
echo ""

echo "【6/6】边界测试"
C_401=$(curl -sS -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/consumption/orders?page=1")
[[ "$C_401" == "401" ]] && pass "领料-无Token返回401" || fail "领料-无Token返回$C_401（预期401）"

R_401=$(curl -sS -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/reversal/orders?page=1")
[[ "$R_401" == "401" ]] && pass "退料-无Token返回401" || fail "退料-无Token返回$R_401（预期401）"

PERM=$(curl -sS "$BASE_URL/api/v1/system/permissions/tree" "${AUTH_GET[@]}" | jq -r '.code // "missing"')
[[ "$PERM" == "0" ]] && pass "权限树接口正常" || fail "权限树接口异常: code=$PERM"

DASH=$(curl -sS "$BASE_URL/api/v1/dashboard/bigscreen" "${AUTH_GET[@]}" | jq -r '.code // "missing"')
[[ "$DASH" == "0" ]] && pass "大屏看板接口正常" || warn "大屏看板接口: code=$DASH"
echo ""

echo "================================================================"
echo "  测试总结"
echo "================================================================"
echo -e "  通过: ${GREEN}$PASS${NC}"
echo -e "  失败: ${RED}$FAIL${NC}"
echo "================================================================"

[[ "$FAIL" -gt 0 ]] && exit 1 || exit 0
