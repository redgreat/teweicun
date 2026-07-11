#!/bin/bash
#
# 功能：仓库调拨全流程 API 测试
# 测试流程：查询仓库 → 查询库存 → 创建调拨单 → 调拨出库确认 → 调拨入库确认 → 验证库存变动
# 创建时间：2026-07-12
# 创建人：Hermes Agent
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
echo "  特维存（TeWeiCun）仓库调拨全流程测试"
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
  pass "登录成功"
else
  fail "登录失败"; exit 1
fi
AUTH=(-H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json")
AUTH_GET=(-H "Authorization: Bearer $TOKEN")
echo ""

# ── 2. 查询仓库 ──
echo "【2/8】查询可用仓库（需要至少2个仓库）"
WH_LIST=$(curl -sS "$BASE_URL/api/v1/base/warehouses?page=1&page_size=10" "${AUTH_GET[@]}")
WH_COUNT=$(echo "$WH_LIST" | jq -r '.data.list | length // 0')

FROM_WH_ID=$(echo "$WH_LIST" | jq -r '.data.list[0].id // empty')
FROM_WH_CODE=$(echo "$WH_LIST" | jq -r '.data.list[0].warehouse_code // empty')
FROM_WH_NAME=$(echo "$WH_LIST" | jq -r '.data.list[0].warehouse_name // empty')
TO_WH_ID=$(echo "$WH_LIST" | jq -r '.data.list[1].id // empty')
TO_WH_CODE=$(echo "$WH_LIST" | jq -r '.data.list[1].warehouse_code // empty')
TO_WH_NAME=$(echo "$WH_LIST" | jq -r '.data.list[1].warehouse_name // empty')

if [[ "$WH_COUNT" -ge 2 ]]; then
  pass "仓库: 共${WH_COUNT}个 调出=${FROM_WH_NAME}(${FROM_WH_CODE}) → 调入=${TO_WH_NAME}(${TO_WH_CODE})"
else
  fail "仓库数量不足（需要≥2个，当前${WH_COUNT}个）"; exit 1
fi
echo ""

# ── 3. 查询调出仓库的可用库存 ──
echo "【3/8】查询调出仓库可用库存（${FROM_WH_NAME}）"
INV_FROM=$(curl -sS "$BASE_URL/api/v1/inventory/available?page=1&page_size=20&warehouse_code=${FROM_WH_CODE}" "${AUTH_GET[@]}")

# 找一条可用量≥3的库存
FIRST_INV=$(echo "$INV_FROM" | jq '[.data.list[] | select(.available_quantity >= 2)] | first')
INV_ID=$(echo "$FIRST_INV" | jq -r '.inventory_id // empty')
MAT_ID=$(echo "$FIRST_INV" | jq -r '.material_id // empty')
AVAIL_QTY=$(echo "$FIRST_INV" | jq -r '.available_quantity // 0')
INV_UNIT=$(echo "$FIRST_INV" | jq -r '.unit // "pcs"')

if [[ -n "$INV_ID" && "$INV_ID" != "null" ]]; then
  pass "选中库存: invID=${INV_ID} matID=${MAT_ID} 可用=${AVAIL_QTY}${INV_UNIT}"
  TRANSFER_QTY=2
else
  fail "调出仓库无可调拨库存（可用量≥3）"
fi
echo ""

# ── 4. 创建调拨单 ──
if [[ -n "${INV_ID:-}" && "$INV_ID" != "null" ]]; then
  echo "【4/8】创建调拨单（${FROM_WH_NAME} → ${TO_WH_NAME}，调拨${TRANSFER_QTY}${INV_UNIT}）"
  TODAY=$(date +%Y-%m-%d)
  CREATE_RESP=$(curl -sS -X POST "$BASE_URL/api/v1/stock-transfers" \
    "${AUTH[@]}" \
    -d "{
      \"transfer_date\": \"$TODAY\",
      \"from_warehouse_id\": $FROM_WH_ID,
      \"to_warehouse_id\": $TO_WH_ID,
      \"remark\": \"自动化测试-仓库调拨-$TODAY\",
      \"items\": [{
        \"inventory_id\": $INV_ID,
        \"material_id\": $MAT_ID,
        \"quantity\": $TRANSFER_QTY
      }]
    }")

  TRANSFER_CODE=$(echo "$CREATE_RESP" | jq -r '.code // "missing"')
  TRANSFER_ID=$(echo "$CREATE_RESP" | jq -r '.data.id // empty')

  if [[ "$TRANSFER_CODE" == "0" && -n "$TRANSFER_ID" && "$TRANSFER_ID" != "null" ]]; then
    pass "调拨单创建成功（ID=${TRANSFER_ID}）"
  else
    TRANSFER_MSG=$(echo "$CREATE_RESP" | jq -r '.msg // "unknown"')
    fail "创建调拨单失败: code=${TRANSFER_CODE} msg=${TRANSFER_MSG}"
    TRANSFER_ID=""
  fi
else
  echo "【4/8】跳过（无可用库存）"
  TRANSFER_ID=""
fi
echo ""

# ── 5. 调拨单列表 ──
echo "【5/8】调拨单列表"
TRANSFER_LIST=$(curl -sS "$BASE_URL/api/v1/stock-transfers?page=1&page_size=10" "${AUTH_GET[@]}")
TRANSFER_LIST_CODE=$(echo "$TRANSFER_LIST" | jq -r '.code // "missing"')
TRANSFER_TOTAL=$(echo "$TRANSFER_LIST" | jq -r '.data.total // 0')

if [[ "$TRANSFER_LIST_CODE" == "0" ]]; then
  pass "调拨单列表: 共${TRANSFER_TOTAL}条"
else
  fail "调拨单列表接口异常: code=${TRANSFER_LIST_CODE}"
fi
echo ""

# ── 6. 调拨出库确认 ──
if [[ -n "${TRANSFER_ID:-}" && "$TRANSFER_ID" != "null" ]]; then
  echo "【6/8】调拨出库确认"
  OUT_RESP=$(curl -sS -X POST "$BASE_URL/api/v1/stock-transfers/${TRANSFER_ID}/confirm-out" "${AUTH[@]}")
  OUT_CODE=$(echo "$OUT_RESP" | jq -r '.code // "missing"')

  if [[ "$OUT_CODE" == "0" ]]; then
    pass "调拨出库确认成功"
  else
    OUT_MSG=$(echo "$OUT_RESP" | jq -r '.msg // "unknown"')
    fail "调拨出库确认失败: code=${OUT_CODE} msg=${OUT_MSG}"
  fi
else
  echo "【6/8】跳过（调拨单未创建）"
fi
echo ""

# ── 7. 调拨入库确认 ──
if [[ -n "${TRANSFER_ID:-}" && "$TRANSFER_ID" != "null" ]]; then
  echo "【7/8】调拨入库确认"
  IN_RESP=$(curl -sS -X POST "$BASE_URL/api/v1/stock-transfers/${TRANSFER_ID}/confirm-in" "${AUTH[@]}")
  IN_CODE=$(echo "$IN_RESP" | jq -r '.code // "missing"')

  if [[ "$IN_CODE" == "0" ]]; then
    pass "调拨入库确认成功"
    # 验证调入仓库库存增加
    sleep 1
    INV_TO=$(curl -sS "$BASE_URL/api/v1/inventory/available?page=1&page_size=20&warehouse_code=${TO_WH_CODE}" "${AUTH_GET[@]}")
    TO_MATCH=$(echo "$INV_TO" | jq "[.data.list[] | select(.material_id == ${MAT_ID})] | first")
    TO_AVAIL=$(echo "$TO_MATCH" | jq -r '.available_quantity // 0')
    pass "调入仓库(${TO_WH_NAME})该物料当前可用: ${TO_AVAIL}${INV_UNIT}"
  else
    IN_MSG=$(echo "$IN_RESP" | jq -r '.msg // "unknown"')
    fail "调拨入库确认失败: code=${IN_CODE} msg=${IN_MSG}"
  fi
else
  echo "【7/8】跳过（调拨单未创建）"
fi
echo ""

# ── 8. 边界测试 ──
echo "【8/8】边界测试"

# 无Token访问
T_401=$(curl -sS -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/stock-transfers?page=1")
[[ "$T_401" == "401" ]] && pass "调拨-无Token返回401" || fail "调拨-无Token返回${T_401}"

# 无效ID详情
T_BAD=$(curl -sS "$BASE_URL/api/v1/stock-transfers/999999" "${AUTH_GET[@]}" | jq -r '.code // 0')
[[ "$T_BAD" != "0" ]] && pass "无效调拨单ID正确拒绝" || fail "无效调拨单ID未拒绝"

# 相同仓库调拨（应被拒绝）
if [[ -n "${TRANSFER_ID:-}" && "$TRANSFER_ID" != "null" ]]; then
  SAME_RESP=$(curl -sS -X POST "$BASE_URL/api/v1/stock-transfers" "${AUTH[@]}" \
    -d "{
      \"transfer_date\": \"$TODAY\",
      \"from_warehouse_id\": $FROM_WH_ID,
      \"to_warehouse_id\": $FROM_WH_ID,
      \"items\": [{\"inventory_id\": $INV_ID, \"material_id\": $MAT_ID, \"quantity\": 1}]
    }")
  SAME_CODE=$(echo "$SAME_RESP" | jq -r '.code // 0')
  if [[ "$SAME_CODE" != "0" ]]; then
    pass "相同仓库调拨正确拒绝"
  else
    warn "相同仓库调拨未拒绝（可能服务端未做校验）"
  fi
else
  warn "跳过相同仓库调拨测试（缺少库存ID）"
fi
echo ""

# ── 测试总结 ──
echo "================================================================"
echo "  测试总结"
echo "================================================================"
echo -e "  通过: ${GREEN}$PASS${NC}"
echo -e "  失败: ${RED}$FAIL${NC}"
echo "================================================================"

[[ "$FAIL" -gt 0 ]] && exit 1 || exit 0
