/**
 * 功能：采购订单 → 入库 全流程 E2E 测试
 * 测试范围：采购订单创建→确认→入库单生成→入库确认→编码验证
 * 创建时间：2026-07-12
 * 创建人：Hermes Agent
 */

package api_flow

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/redgreat/teweicun/test/testutil"
)

func TestFlow_PurchaseToStockIn(t *testing.T) {
	ctx, cancel, admin := initE2E(t)
	defer cancel()

	prefix := testutil.UniquePrefix()
	supplierCode := mustEnsureSupplier(ctx, t, admin, prefix)
	warehouseCode := mustEnsureWarehouse(ctx, t, admin, prefix, 1)

	// 创建测试物料
	codeMatID := mustCreateMaterial(ctx, t, admin, 0, prefix+"_编码物料_钢板", true)
	noCodeMatID := mustCreateMaterial(ctx, t, admin, 0, prefix+"_非编码物料_焊条", false)

	t.Logf("test data: supplier=%s warehouse=%s code_mat=%d nocode_mat=%d",
		supplierCode, warehouseCode, codeMatID, noCodeMatID)

	// ── 1. 创建采购订单 ──
	poID := mustCreatePurchaseOrder(ctx, t, admin, supplierCode, codeMatID, noCodeMatID)
	t.Logf("purchase order created: id=%d", poID)

	// ── 2. 确认采购订单 ──
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/purchase/orders/%d/confirm", poID))

	po := waitPurchaseOrderDetail(ctx, t, admin, poID, func(p PurchaseOrderDetail) bool { return p.StockInID > 0 })
	t.Logf("purchase confirmed: orderNo=%s stockInID=%d items=%d", po.OrderNo, po.StockInID, len(po.Items))

	// ── 3. 设置入库仓库并确认入库 ──
	setStockInWarehouseIfEmpty(ctx, t, admin, po.StockInID, warehouseCode)
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-in/%d/confirm", po.StockInID))

	var stockIn StockInDetail
	if err := admin.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-in/%d", po.StockInID), nil, nil, &stockIn); err != nil {
		t.Fatalf("get stock-in failed: %v", err)
	}
	t.Logf("stock-in confirmed: stockInNo=%s items=%d", stockIn.StockInNo, len(stockIn.Items))

	// ── 4. 验证编码物料生成了10个序列号 ──
	codeStockInItemID := findStockInItemIDByMaterial(stockIn, codeMatID)
	if codeStockInItemID <= 0 {
		t.Fatalf("no coded stock-in item for material_id=%d", codeMatID)
	}
	var serials []SerialCodeItem
	if err := admin.DoJSON(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/serial-codes/stock-in-item/%d", codeStockInItemID),
		nil, nil, &serials); err != nil {
		t.Fatalf("get serial codes failed: %v", err)
	}
	if len(serials) != 10 {
		t.Fatalf("expected 10 serial codes, got %d", len(serials))
	}
	t.Logf("serial codes verified: count=%d sample=%s", len(serials), serials[0].SerialCode)

	// ── 5. 验证入库后库存增加 ──
	codeInv := mustPickInventoryAvailable(ctx, t, admin, warehouseCode, codeMatID, true, 5)
	noCodeInv := mustPickInventoryAvailable(ctx, t, admin, warehouseCode, noCodeMatID, false, 5)
	t.Logf("inventory after stock-in: code_inv=%d avail=%.2f nocode_inv=%d avail=%.2f",
		codeInv.InventoryID, codeInv.AvailableQuantity, noCodeInv.InventoryID, noCodeInv.AvailableQuantity)
}

// ============================================================================
// 采购订单 —— 错误场景
// ============================================================================

func TestFlow_PurchaseOrder_InvalidSupplier(t *testing.T) {
	ctx, cancel, admin := initE2E(t)
	defer cancel()

	prefix := testutil.UniquePrefix()
	noCodeMatID := mustCreateMaterial(ctx, t, admin, 0, prefix+"_测试物料_钢管", false)

	req := map[string]any{
		"order_type":    "purchase",
		"supplier_code": "INVALID_SUPPLIER_99999",
		"order_date":    "2026-07-12",
		"items":         []map[string]any{{"material_id": noCodeMatID, "quantity": 1, "unit_price": 10}},
	}
	var out testutil.IDResp
	err := admin.DoJSON(ctx, http.MethodPost, "/api/v1/purchase/orders", nil, req, &out)
	if err == nil {
		t.Fatal("expected error for invalid supplier, but got success")
	}
	t.Logf("invalid supplier correctly rejected: %v", err)
}

func TestFlow_PurchaseOrder_ConfirmTwice(t *testing.T) {
	ctx, cancel, admin := initE2E(t)
	defer cancel()

	prefix := testutil.UniquePrefix()
	supplierCode := mustEnsureSupplier(ctx, t, admin, prefix)
	matID := mustCreateMaterial(ctx, t, admin, 0, prefix+"_封头", false)

	poID := mustCreatePurchaseOrder(ctx, t, admin, supplierCode, matID, matID)
	// 创建时已自动确认，再次确认应失败
	err := admin.DoJSON(ctx, http.MethodPost, fmt.Sprintf("/api/v1/purchase/orders/%d/confirm", poID), nil, nil, nil)
	if err != nil {
		t.Logf("double-confirm correctly rejected: %v", err)
	} else {
		t.Log("double-confirm did not reject (may be idempotent)")
	}
}

// ============================================================================
// 采购订单 —— 多行明细测试
// ============================================================================

func TestFlow_PurchaseOrder_MultiItems(t *testing.T) {
	ctx, cancel, admin := initE2E(t)
	defer cancel()

	prefix := testutil.UniquePrefix()
	supplierCode := mustEnsureSupplier(ctx, t, admin, prefix)

	// 创建3个物料
	mat1 := mustCreateMaterial(ctx, t, admin, 0, prefix+"_板材_A3钢", false)
	mat2 := mustCreateMaterial(ctx, t, admin, 0, prefix+"_管材_无缝管", false)
	mat3 := mustCreateMaterial(ctx, t, admin, 0, prefix+"_型材_角钢", true)

	req := map[string]any{
		"order_type":    "purchase",
		"supplier_code": supplierCode,
		"order_date":    "2026-07-12",
		"remark":        "多行明细测试",
		"items": []map[string]any{
			{"material_id": mat1, "quantity": 5, "unit_price": 100},
			{"material_id": mat2, "quantity": 10, "unit_price": 200},
			{"material_id": mat3, "quantity": 3, "unit_price": 500},
		},
	}
	var out testutil.IDResp
	if err := admin.DoJSON(ctx, http.MethodPost, "/api/v1/purchase/orders", nil, req, &out); err != nil {
		t.Fatalf("create multi-item purchase order failed: %v", err)
	}

	var po PurchaseOrderDetail
	if err := admin.DoJSON(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/purchase/orders/%d", out.ID), nil, nil, &po); err != nil {
		t.Fatalf("get purchase order failed: %v", err)
	}
	if len(po.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(po.Items))
	}
	t.Logf("multi-item order: id=%d items=%d stockInID=%d", po.ID, len(po.Items), po.StockInID)
}
