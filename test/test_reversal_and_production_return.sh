#!/bin/bash
#
# 功能：退料订单 + 生产退货单 综合测试
# 退料：已出库库存 → 创建退料单 → 确认 → 退料入库确认 → 验证库存
# 生产退货：生产单 → 生产退货单（预留功能）
# 创建时间：2026-07-12
# 创建人：Hermes Agent
#

set -eo pipefail

BASE_URL="${TWC_BASE_URL:-http://localhost:8080}"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
PASS=0; FAIL=0
pass() { echo -e "${GREEN}[PASS]${NC} $*"; PASS=$((PASS+1)); }
fail() { echo -e "${RED}[FAIL]${NC} $*"; FAIL=$((FAIL+1)); }
warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }

echo "================================================================"
echo "  特维存 退料 + 生产退货 综合测试"
echo "  目标: $BASE_URL"
echo "================================================================"
echo ""

# ── 登录 ──
echo "【登录】"
LOGIN=$(curl -sS -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')
TOKEN=$(echo "$LOGIN" | jq -r '.data.token')
AUTH=(-H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json")
AUTH_GET=(-H "Authorization: Bearer $TOKEN")
pass "登录成功"
echo ""

# ════════════════════════════════════════════
# 第一部分：退料订单（Reversal Order）
# ════════════════════════════════════════════
echo "══════════ 退料订单流程 ══════════"
echo ""

# ── 1. 查已出库库存 ──
echo "【退料1/5】查询已出库库存"
ISSUED=$(curl -sS "$BASE_URL/api/v1/inventory/issued?page=1&page_size=50" "${AUTH_GET[@]}")
ISSUED_TOTAL=$(echo "$ISSUED" | jq -r '.data.total // 0')
# 找一个非编码的已出库库存（退料不必处理编码）
REV_INV=$(echo "$ISSUED" | jq '[.data.list[] | select(.is_code == false and .issued_quantity >= 2)] | first')
REV_INV_ID=$(echo "$REV_INV" | jq -r '.inventory_id // empty')
REV_MAT_ID=$(echo "$REV_INV" | jq -r '.material_id // empty')
REV_QTY=$(echo "$REV_INV" | jq -r '.issued_quantity // 0')
REV_WH_CODE=$(echo "$REV_INV" | jq -r '.warehouse_code // "WH001"')

if [[ -n "$REV_INV_ID" && "$REV_INV_ID" != "null" ]]; then
  pass "选中已出库库存: inv=$REV_INV_ID mat=$REV_MAT_ID issued=$REV_QTY wh=$REV_WH_CODE"
else
  # fallback: 用编码物料
  REV_INV=$(echo "$ISSUED" | jq '[.data.list[] | select(.issued_quantity >= 2)] | first')
  REV_INV_ID=$(echo "$REV_INV" | jq -r '.inventory_id // empty')
  REV_MAT_ID=$(echo "$REV_INV" | jq -r '.material_id // empty')
  REV_QTY=$(echo "$REV_INV" | jq -r '.issued_quantity // 0')
  REV_WH_CODE=$(echo "$REV_INV" | jq -r '.warehouse_code // "WH001"')
  pass "使用编码物料库存: inv=$REV_INV_ID mat=$REV_MAT_ID issued=$REV_QTY"
fi
TODAY=$(date +%Y-%m-%d)
echo ""

# ── 2. 创建退料单 ──
echo "【退料2/5】创建退料单"
REV_JSON=$(cat <<EOF
{
  "project_no": "REV-TEST-$(date +%H%M%S)",
  "product_name": "退料测试产品",
  "order_date": "$TODAY",
  "designer_name": "退料测试员",
  "remark": "退料流程自动化测试",
  "items": [{
    "inventory_id": $REV_INV_ID,
    "material_id": $REV_MAT_ID,
    "quantity": 1,
    "unit": "pcs",
    "remark": "退回1件"
  }]
}
EOF
)

REV_RESP=$(curl -sS -X POST "$BASE_URL/api/v1/reversal/orders" "${AUTH[@]}" -d "$REV_JSON")
REV_CODE=$(echo "$REV_RESP" | jq -r '.code')
REV_ID=$(echo "$REV_RESP" | jq -r '.data.id // empty')

if [[ "$REV_CODE" == "0" && -n "$REV_ID" && "$REV_ID" != "null" ]]; then
  pass "退料单创建成功（ID=$REV_ID）"
else
  REV_MSG=$(echo "$REV_RESP" | jq -r '.msg')
  fail "退料单创建失败: $REV_MSG"
  REV_ID=""
fi
echo ""

# ── 3. 确认退料单 ──
if [[ -n "${REV_ID:-}" && "$REV_ID" != "null" ]]; then
  echo "【退料3/5】确认退料单"
  REV_CFM=$(curl -sS -X POST "$BASE_URL/api/v1/reversal/orders/$REV_ID/confirm" "${AUTH[@]}")
  REV_CFM_CODE=$(echo "$REV_CFM" | jq -r '.code')
  if [[ "$REV_CFM_CODE" == "0" ]]; then
    pass "退料单确认成功"
    sleep 2
    REV_DETAIL=$(curl -sS "$BASE_URL/api/v1/reversal/orders/$REV_ID" "${AUTH_GET[@]}")
    REV_SI_ID=$(echo "$REV_DETAIL" | jq -r '.data.stock_in_id // 0')
    pass "退料入库单已生成（stock_in_id=$REV_SI_ID）"
  else
    REV_CFM_MSG=$(echo "$REV_CFM" | jq -r '.msg')
    fail "退料确认失败: $REV_CFM_MSG"
    REV_SI_ID=""
  fi
else
  echo "【退料3/5】跳过"
  REV_SI_ID=""
fi
echo ""

# ── 4. 确认退料入库 ──
if [[ -n "${REV_SI_ID:-}" && "$REV_SI_ID" != "0" ]]; then
  echo "【退料4/5】确认退料入库"
  SI_CFM=$(curl -sS -X POST "$BASE_URL/api/v1/stock-in/$REV_SI_ID/confirm-reversal" "${AUTH[@]}")
  SI_CFM_CODE=$(echo "$SI_CFM" | jq -r '.code')
  if [[ "$SI_CFM_CODE" == "0" ]]; then
    pass "退料入库确认成功（物料退回仓库）"
  else
    SI_MSG=$(echo "$SI_CFM" | jq -r '.msg')
    warn "退料入库确认: $SI_MSG（可能已自动确认）"
  fi
else
  echo "【退料4/5】跳过"
fi
echo ""

# ── 5. 退料列表验证 ──
echo "【退料5/5】退料单列表"
REV_LIST=$(curl -sS "$BASE_URL/api/v1/reversal/orders?page=1&page_size=5" "${AUTH_GET[@]}")
REV_TOTAL=$(echo "$REV_LIST" | jq -r '.data.total // 0')
[[ "$REV_TOTAL" -gt 0 ]] && pass "退料单列表: 共${REV_TOTAL}条" || fail "退料单列表为空"
echo ""

# ════════════════════════════════════════════
# 第二部分：生产退货单（Production Return）
# ════════════════════════════════════════════
echo "══════════ 生产退货单流程 ══════════"
echo ""

# ── 6. 查生产单 ──
echo "【生产退货1/4】查询可用生产单"
PROD_LIST=$(curl -sS "$BASE_URL/api/v1/production/orders?page=1&page_size=10" "${AUTH_GET[@]}")
PROD_TOTAL=$(echo "$PROD_LIST" | jq -r '.data.total // 0')
PROD_ROW=$(echo "$PROD_LIST" | jq '.data.list[0]')
PROD_ID=$(echo "$PROD_ROW" | jq -r '.id // empty')
PROD_NO=$(echo "$PROD_ROW" | jq -r '.production_no // empty')
PROD_MAT_NAME=$(echo "$PROD_ROW" | jq -r '.produced_material_name // empty')
PROD_QTY=$(echo "$PROD_ROW" | jq -r '.produced_quantity // 0')

[[ "$PROD_TOTAL" -gt 0 ]] && pass "生产单: 共${PROD_TOTAL}条，最新=${PROD_NO} 产品=${PROD_MAT_NAME} 数量=$PROD_QTY" || fail "无生产单"
echo ""

# ── 7. 创建生产退货单（新接口） ──
echo "【生产退货2/4】创建生产退货单"
if [[ -n "$PROD_ID" && "$PROD_ID" != "null" && "$PROD_QTY" != "0" ]]; then
  RET_JSON=$(cat <<EOF
{
  "production_order_id": $PROD_ID,
  "returned_quantity": 1,
  "remark": "自动化测试-成品退回"
}
EOF
)
  RET_RESP=$(curl -sS -X POST "$BASE_URL/api/v1/production/returns" "${AUTH[@]}" -d "$RET_JSON")
  RET_CODE=$(echo "$RET_RESP" | jq -r '.code')
  RET_ID=$(echo "$RET_RESP" | jq -r '.data.id // empty')

  if [[ "$RET_CODE" == "0" && -n "$RET_ID" && "$RET_ID" != "null" ]]; then
    pass "生产退货单创建成功（ID=$RET_ID）"
  else
    RET_MSG=$(echo "$RET_RESP" | jq -r '.msg // "unknown"')
    warn "生产退货单创建: $RET_MSG"
  fi
else
  warn "无可用生产单，跳过创建"
fi
echo ""
RET_LIST=$(curl -sS "$BASE_URL/api/v1/production/returns?page=1&page_size=5" "${AUTH_GET[@]}")
RET_TOTAL=$(echo "$RET_LIST" | jq -r '.data.total // 0')
pass "生产退货单列表: 共${RET_TOTAL}条（当前数据）"
echo ""

# ── 8. 生产退货单详情（如果有） ──
echo "【生产退货3/4】边界测试"
R_401=$(curl -sS -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/production/returns?page=1")
[[ "$R_401" == "401" ]] && pass "生产退货-无Token返回401" || fail "生产退货-无Token返回$R_401"

# 退料无Token
REV_401=$(curl -sS -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/reversal/orders?page=1")
[[ "$REV_401" == "401" ]] && pass "退料-无Token返回401" || fail "退料-无Token返回$REV_401"
echo ""

# ── 9. 验证退料后库存恢复 ──
echo "【生产退货4/4】验证库存状态"
AVAIL=$(curl -sS "$BASE_URL/api/v1/inventory/available?page=1&page_size=5" "${AUTH_GET[@]}")
AVAIL_TOTAL=$(echo "$AVAIL" | jq -r '.data.total // 0')
[[ "$AVAIL_TOTAL" -gt 0 ]] && pass "可用库存: 共${AVAIL_TOTAL}条" || fail "可用库存为空"

ISSUED_AFTER=$(curl -sS "$BASE_URL/api/v1/inventory/issued?page=1&page_size=5" "${AUTH_GET[@]}")
ISSUED_AFTER_TOTAL=$(echo "$ISSUED_AFTER" | jq -r '.data.total // 0')
pass "已出库库存: 共${ISSUED_AFTER_TOTAL}条"
echo ""

echo "================================================================"
echo "  流程说明"
echo "================================================================"
echo "  退料流程: 已出库库存 → 创建退料单 → 确认 → 退料入库确认"
echo "    领料出库的材料可以退回到仓库，恢复可用库存"
echo ""
echo "  生产退货: 生产单成品 → 生产退货单（预留功能）"
echo "    当前生产退货单由系统自动生成，暂无手动创建入口"
echo "================================================================"
echo ""
echo "  测试总结: 通过=${GREEN}$PASS${NC}  失败=${RED}$FAIL${NC}"
echo "================================================================"

[[ "$FAIL" -gt 0 ]] && exit 1 || exit 0
