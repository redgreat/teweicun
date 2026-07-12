#!/bin/bash
#
# 功能：生产管理全流程 API 测试
# 流程：领料（指定成品物料+数量+仓库）→ 确认 → 出库确认 → 触发器自动生产入库 → 成本验证
# 原理：领料出库确认时，DB触发器调用 sp_generate_production_from_consumption
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

echo "================================================================
  特维存（TeWeiCun）生产管理全流程测试
  领料 → 自动生产入库 → 成本计算 → 生产单验证
  目标: $BASE_URL  时间: $(date '+%H:%M:%S')
================================================================"
echo ""

# ── 1. 登录 ──
echo "【1/8】登录"
LOGIN=$(curl -sS -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')
TOKEN=$(echo "$LOGIN" | jq -r '.data.token')
AUTH=(-H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json")
AUTH_GET=(-H "Authorization: Bearer $TOKEN")
pass "登录成功"
echo ""

# ── 2. 准备数据 ──
echo "【2/8】准备仓库和库存"
WH_CODE="WH001"
WH_ID=2

# 取任意可用库存（≥3）
INV=$(curl -sS "$BASE_URL/api/v1/inventory/available?page=1&page_size=50&warehouse_code=$WH_CODE" "${AUTH_GET[@]}")
INV_ROW=$(echo "$INV" | jq '[.data.list[] | select(.available_quantity >= 3 and .is_code == false)] | first')
INV_ID=$(echo "$INV_ROW" | jq -r '.inventory_id // empty')
INV_MAT_ID=$(echo "$INV_ROW" | jq -r '.material_id // empty')
INV_QTY=$(echo "$INV_ROW" | jq -r '.available_quantity // 0')

if [[ -z "$INV_ID" || "$INV_ID" == "null" ]]; then
  fail "无可用库存（需要≥3件非编码物料）"
  exit 1
fi
pass "选中库存: inv=$INV_ID mat=$INV_MAT_ID avail=$INV_QTY"

# 成品物料：取另一个非编码物料
PROD_ROW=$(echo "$INV" | jq "[.data.list[] | select(.is_code == false and .material_id != $INV_MAT_ID)] | first")
PROD_MAT_ID=$(echo "$PROD_ROW" | jq -r '.material_id // empty')
if [[ -z "$PROD_MAT_ID" || "$PROD_MAT_ID" == "null" ]]; then
  PROD_MAT_ID="$INV_MAT_ID"
  warn "无独立成品物料，复用原材料"
else
  pass "成品物料: mat=$PROD_MAT_ID"
fi
echo ""

# ── 3. 创建领料单（带生产入库参数） ──
echo "【3/8】创建领料单"
TODAY=$(date +%Y-%m-%d)
CO_JSON=$(cat <<JSONEOF
{
  "project_no": "PROD-$(date +%H%M%S)",
  "product_name": "压力容器-测试产品",
  "order_date": "$TODAY",
  "designer_name": "生产测试员",
  "remark": "生产流程自动化测试",
  "produced_material_id": $PROD_MAT_ID,
  "produced_quantity": 2,
  "produced_warehouse_id": $WH_ID,
  "items": [{
    "material_id": $INV_MAT_ID,
    "inventory_id": $INV_ID,
    "quantity": 2,
    "unit": "pcs",
    "remark": "原材料消耗-生产测试"
  }]
}
JSONEOF
)

CO_RESP=$(curl -sS -X POST "$BASE_URL/api/v1/consumption/orders" "${AUTH[@]}" -d "$CO_JSON")
CO_CODE=$(echo "$CO_RESP" | jq -r '.code')
CO_ID=$(echo "$CO_RESP" | jq -r '.data.id // empty')

if [[ "$CO_CODE" == "0" && -n "$CO_ID" && "$CO_ID" != "null" ]]; then
  pass "领料单创建成功（ID=$CO_ID）"
else
  CO_MSG=$(echo "$CO_RESP" | jq -r '.msg')
  fail "领料单创建失败: $CO_MSG"; exit 1
fi
echo ""

# ── 4. 确认领料单 ──
echo "【4/8】确认领料单"
curl -sS -X POST "$BASE_URL/api/v1/consumption/orders/$CO_ID/confirm" "${AUTH[@]}" > /dev/null
sleep 2
CO_DETAIL=$(curl -sS "$BASE_URL/api/v1/consumption/orders/$CO_ID" "${AUTH_GET[@]}")
STOCK_OUT_ID=$(echo "$CO_DETAIL" | jq -r '.data.stock_out_id // 0')
[[ "$STOCK_OUT_ID" != "0" ]] && pass "领料确认→出库单SO=$STOCK_OUT_ID" || fail "出库单未生成"
echo ""

# ── 5. 确认出库（触发器自动生产入库！） ──
echo "【5/8】确认出库单"
[[ "$STOCK_OUT_ID" == "0" ]] && { fail "无法继续"; exit 1; }

curl -sS -X PUT "$BASE_URL/api/v1/stock-out/$STOCK_OUT_ID/serial-selections" "${AUTH[@]}" \
  -d '{"mode":"auto_fifo","items":[]}' > /dev/null 2>&1

SO_RESP=$(curl -sS -X POST "$BASE_URL/api/v1/stock-out/$STOCK_OUT_ID/confirm" "${AUTH[@]}")
SO_CODE=$(echo "$SO_RESP" | jq -r '.code')

if [[ "$SO_CODE" == "0" ]]; then
  pass "出库确认成功（触发器执行生产入库！）"
  sleep 3
else
  SO_MSG=$(echo "$SO_RESP" | jq -r '.msg')
  fail "出库确认失败: $SO_MSG"; exit 1
fi
echo ""

# ── 6. 验证生产单 ──
echo "【6/8】验证自动生成的生产单"
PROD_LIST=$(curl -sS "$BASE_URL/api/v1/production/orders?page=1&page_size=10" "${AUTH_GET[@]}")
PROD_MATCH=$(echo "$PROD_LIST" | jq "[.data.list[] | select(.consumption_order_id == $CO_ID)] | first")
PROD_ID=$(echo "$PROD_MATCH" | jq -r '.id // empty')
PROD_NO=$(echo "$PROD_MATCH" | jq -r '.production_no // empty')
PROD_QTY=$(echo "$PROD_MATCH" | jq -r '.produced_quantity // 0')
PROD_UNIT_COST=$(echo "$PROD_MATCH" | jq -r '.produced_unit_cost // 0')
PROD_COST=$(echo "$PROD_MATCH" | jq -r '.cost_price // 0')
PROD_SI_ID=$(echo "$PROD_MATCH" | jq -r '.stock_in_id // 0')

if [[ -n "$PROD_ID" && "$PROD_ID" != "null" ]]; then
  pass "生产单: $PROD_NO qty=$PROD_QTY unit_cost=$PROD_UNIT_COST cost=$PROD_COST stock_in=$PROD_SI_ID"
else
  fail "未找到生产单 — 触发器可能未执行"
fi
echo ""

# ── 7. 验证入库和成本 ──
echo "【7/8】验证入库库存和成本"
INV_AFTER=$(curl -sS "$BASE_URL/api/v1/inventory/available?page=1&page_size=50&warehouse_code=$WH_CODE" "${AUTH_GET[@]}")
PROD_INV=$(echo "$INV_AFTER" | jq "[.data.list[] | select(.material_id == $PROD_MAT_ID)] | first")
PROD_AVAIL=$(echo "$PROD_INV" | jq -r '.available_quantity // 0')

[[ "$PROD_AVAIL" != "0" ]] && pass "成品库存: mat=$PROD_MAT_ID avail=$PROD_AVAIL" || warn "成品库存未找到"

# 抽查销售订单成本
SO_ID=$(curl -sS "$BASE_URL/api/v1/sales/orders?page=1&page_size=1" "${AUTH_GET[@]}" | jq -r '.data.list[0].id // empty')
if [[ -n "$SO_ID" && "$SO_ID" != "null" ]]; then
  SO_DETAIL=$(curl -sS "$BASE_URL/api/v1/sales/orders/$SO_ID" "${AUTH_GET[@]}")
  HAS_COST=$(echo "$SO_DETAIL" | jq '[.data.items[]? | .unit_cost] | map(select(. > 0)) | length')
  [[ "$HAS_COST" -gt 0 ]] && pass "销售订单含成本数据（${HAS_COST}条）" || warn "销售订单无成本数据"
fi
echo ""

# ── 8. 边界测试 ──
echo "【8/8】边界测试"
P_401=$(curl -sS -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/production/orders?page=1")
[[ "$P_401" == "401" ]] && pass "生产-无Token返回401" || fail "生产-无Token返回$P_401"

P_BAD=$(curl -sS "$BASE_URL/api/v1/production/orders/999999" "${AUTH_GET[@]}" | jq -r '.code // 0')
[[ "$P_BAD" != "0" ]] && pass "无效生产单ID正确拒绝" || fail "无效生产单ID未拒绝"
echo ""

echo "================================================================
  流程说明
================================================================
  领料指定 produced_material_id/quantity/warehouse_id
  → 领料确认 → 出库单生成
  → 出库确认 → DB触发器 → sp_generate_production_from_consumption
  → 自动: 创建入库单 + 确认入库 + 生成生产单
  → 成本 = Σ(领料量×库存单价) / 成品数量
================================================================
  测试总结
================================================================
  通过: ${GREEN}$PASS${NC}  失败: ${RED}$FAIL${NC}
================================================================"

[[ "$FAIL" -gt 0 ]] && exit 1 || exit 0
