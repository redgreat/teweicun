/**
 * 功能：单独验证采购模块主流程
 *   采购订单（含编码/非编码多物料）-> 自动生成入库单 -> 入库确认自动生成编码
 *   -> 采购付款与供应商对账 -> 采购退货单 -> 自动生成出库单 -> 出库确认
 *   -> 校验最终库存、库存台账、供应商对账
 * 创建时间：2026-06-11
 * 创建人：GPT-5.4
 */

package api_flow

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/redgreat/teweicun/test/testutil"
)

type purchaseModulePlan struct {
	MaterialID   int64
	MaterialName string
	IsCode       bool
	PurchaseQty  float64
	ReturnQty    float64
	UnitPrice    float64
}

func TestFlow_PurchaseModuleLifecycle(t *testing.T) {
	env := testutil.LoadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()

	admin := testutil.NewClient(env.BaseURL)
	adminLogin, err := admin.Login(ctx, env.AdminUser, env.AdminPass)
	if err != nil {
		t.Fatalf("admin login failed: %v", err)
	}

	fixture := mustLoadBaseDataFixture(ctx, t, admin, adminLogin.UserID)
	categoryID := fixture.CategoryID
	warehouseCode := fixture.MainMaterialWarehouse
	supplierCode := fixture.SupplierCode
	supplierRecBefore, hasSupplierRecBefore := mustGetSupplierReconciliationSnapshot(ctx, t, admin, supplierCode)

	manualPartialOrderAmount := mustExercisePurchasePartialReceiptManualSerialOneByOne(ctx, t, admin, categoryID, warehouseCode, supplierCode)
	autoPartialOrderAmount := mustExercisePurchasePartialReceiptAutoSerialOneByOne(ctx, t, admin, categoryID, warehouseCode, supplierCode)

	plans := mustCreatePurchaseModuleMaterials(ctx, t, admin, categoryID)
	if len(plans) < 6 {
		t.Fatalf("purchase module materials not enough: got=%d", len(plans))
	}

	orderGroups := [][]purchaseModulePlan{
		plans[0:2],
		plans[2:4],
		plans[4:6],
	}

	totalOrderAmount := manualPartialOrderAmount + autoPartialOrderAmount
	for _, group := range orderGroups {
		poID, orderAmount := mustCreatePurchaseOrderBatch(ctx, t, admin, supplierCode, group)
		mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/purchase/orders/%d/confirm", poID))

		po := waitPurchaseOrderDetailExt(ctx, t, admin, poID, func(p PurchaseOrderDetailExt) bool { return true })
		if po.StockInID <= 0 {
			if autoID, ok := waitFindStockInIDByPurchaseOrderID(ctx, admin, poID); ok {
				po.StockInID = autoID
			}
		}
		if po.StockInID <= 0 {
			t.Fatalf("purchase order did not auto-generate stock-in: poID=%d orderNo=%s", poID, po.OrderNo)
		}
		if len(po.Items) != len(group) {
			t.Fatalf("purchase order item count mismatch: got=%d want=%d", len(po.Items), len(group))
		}

		setStockInWarehouseIfEmpty(ctx, t, admin, po.StockInID, warehouseCode)
		mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-in/%d/confirm", po.StockInID))

		var stockIn StockInDetail
		if err := admin.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-in/%d", po.StockInID), nil, nil, &stockIn); err != nil {
			t.Fatalf("get stock-in failed: %v", err)
		}
		if stockIn.ID != po.StockInID || stockIn.StockInNo == "" {
			t.Fatalf("stock-in detail invalid: id=%d stockInNo=%s", stockIn.ID, stockIn.StockInNo)
		}
		if len(stockIn.Items) != len(group) {
			t.Fatalf("stock-in item count mismatch: got=%d want=%d", len(stockIn.Items), len(group))
		}

		for _, plan := range group {
			if !plan.IsCode {
				continue
			}
			mustWaitStockInSerials(ctx, t, admin, po.StockInID, plan.MaterialID, int(plan.PurchaseQty))
		}

		totalOrderAmount += orderAmount
	}

	mustAssertSupplierReconciliationDelta(ctx, t, admin, supplierCode, supplierRecBefore, hasSupplierRecBefore, totalOrderAmount)

	inventoryMap := make(map[int64]InventoryAvailable, len(plans))
	for _, plan := range plans {
		inventoryMap[plan.MaterialID] = mustPickInventoryAvailable(ctx, t, admin, warehouseCode, plan.MaterialID, plan.IsCode, plan.ReturnQty)
	}

	for _, group := range orderGroups {
		returnID := mustCreatePurchaseReturnOrderBatch(ctx, t, admin, supplierCode, warehouseCode, group, inventoryMap)
		mustConfirmReturnOrder(ctx, t, admin, returnID, warehouseCode)

		ret := waitReturnOrder(ctx, t, admin, returnID, func(o ReturnOrder) bool {
			return o.StockOutID != nil && *o.StockOutID > 0
		})
		if ret.StockOutID == nil || *ret.StockOutID <= 0 {
			t.Fatalf("purchase return did not auto-generate stock-out: returnID=%d returnNo=%s", ret.ID, ret.ReturnNo)
		}

		if err := tryAutoStockOutSerialSelections(ctx, admin, *ret.StockOutID); err != nil {
			t.Fatalf("auto serial selection for purchase return stock-out failed: %v", err)
		}
		mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-out/%d/confirm", *ret.StockOutID))

		var stockOut StockOutDetail
		if err := admin.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-out/%d", *ret.StockOutID), nil, nil, &stockOut); err != nil {
			t.Fatalf("get purchase return stock-out failed: %v", err)
		}
		if len(stockOut.Items) != len(group) {
			t.Fatalf("purchase return stock-out item count mismatch: got=%d want=%d", len(stockOut.Items), len(group))
		}

		mustAssertReturnStockOutSerials(ctx, t, admin, stockOut, group)
	}

	mustAssertPurchaseModuleInventory(ctx, t, admin, warehouseCode, plans)
	mustAssertPurchaseModuleLedger(ctx, t, admin, plans)
}

func mustCreatePurchaseModuleMaterials(ctx context.Context, t *testing.T, c *testutil.Client, categoryID int64) []purchaseModulePlan {
	t.Helper()

	rawPlans := []struct {
		name      string
		isCode    bool
		purchase  float64
		ret       float64
		unitPrice float64
	}{
		{"编码钢板A", true, 3, 1, 21.50},
		{"普通钢板B", false, 12, 2, 17.90},
		{"编码法兰C", true, 4, 1, 24.80},
		{"散装焊材D", false, 15, 3, 8.20},
		{"编码封头E", true, 5, 2, 48.60},
		{"垫片辅料F", false, 18, 4, 3.60},
	}

	plans := make([]purchaseModulePlan, 0, len(rawPlans))
	for _, item := range rawPlans {
		name := uniqueChineseName(item.name)
		materialID := mustCreateMaterial(ctx, t, c, categoryID, name, item.isCode)
		plans = append(plans, purchaseModulePlan{
			MaterialID:   materialID,
			MaterialName: name,
			IsCode:       item.isCode,
			PurchaseQty:  item.purchase,
			ReturnQty:    item.ret,
			UnitPrice:    item.unitPrice,
		})
	}
	return plans
}

func mustExercisePurchasePartialReceiptManualSerialOneByOne(ctx context.Context, t *testing.T, c *testutil.Client, categoryID int64, warehouseCode, supplierCode string) float64 {
	t.Helper()

	materialID := mustCreateMaterial(ctx, t, c, categoryID, uniqueChineseName("编码手动分批收货"), true)
	plan := purchaseModulePlan{
		MaterialID:   materialID,
		MaterialName: uniqueChineseName("编码手动分批收货物料"),
		IsCode:       true,
		PurchaseQty:  3,
		UnitPrice:    19.8,
	}

	poID, _ := mustCreatePurchaseOrderBatch(ctx, t, c, supplierCode, []purchaseModulePlan{plan})
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/purchase/orders/%d/confirm", poID))

	po := waitPurchaseOrderDetailExt(ctx, t, c, poID, func(p PurchaseOrderDetailExt) bool { return true })
	if po.StockInID <= 0 {
		if autoID, ok := waitFindStockInIDByPurchaseOrderID(ctx, c, poID); ok {
			po.StockInID = autoID
		}
	}
	if po.StockInID <= 0 {
		t.Fatalf("manual partial purchase receipt scenario missing stock-in: poID=%d", poID)
	}

	setStockInWarehouseIfEmpty(ctx, t, c, po.StockInID, warehouseCode)

	manualCodes := make([]string, 0, int(plan.PurchaseQty))
	for step := 1; step <= int(plan.PurchaseQty); step++ {
		manualCode := fmt.Sprintf("MANUAL%s%02d", testutil.UniquePrefix(), step)
		manualCodes = append(manualCodes, manualCode)
		customAttrs := map[string]any{
			"manual_serial_codes": []string{manualCode},
		}
		mustUpdatePurchaseStockInAcceptedQty(ctx, t, c, po.StockInID, warehouseCode, materialID, 1, customAttrs)
		mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/stock-in/%d/confirm", po.StockInID))

		stockIn := mustGetStockInDetail(ctx, t, c, po.StockInID)
		wantStatus := "pending"
		if step == int(plan.PurchaseQty) {
			wantStatus = "passed"
		}
		if stockIn.StockInStatus != wantStatus {
			t.Fatalf("manual partial purchase receipt status mismatch at step=%d got=%s want=%s", step, stockIn.StockInStatus, wantStatus)
		}

		serials := mustWaitAndFetchStockInSerials(ctx, t, c, po.StockInID, materialID, step)
		if len(serials) != step {
			t.Fatalf("manual partial purchase receipt serial count mismatch at step=%d got=%d want=%d", step, len(serials), step)
		}
		got := make(map[string]struct{}, len(serials))
		for _, serial := range serials {
			got[serial.SerialCode] = struct{}{}
		}
		for _, code := range manualCodes {
			if _, ok := got[code]; !ok {
				t.Fatalf("manual partial purchase receipt missing manual serial code at step=%d code=%s", step, code)
			}
		}
	}

	got := mustSumAvailableQty(ctx, t, c, warehouseCode, materialID)
	assertFloatNear(t, got, plan.PurchaseQty, 0.001, "手动分批收货最终库存")

	return plan.PurchaseQty * plan.UnitPrice
}

func mustExercisePurchasePartialReceiptAutoSerialOneByOne(ctx context.Context, t *testing.T, c *testutil.Client, categoryID int64, warehouseCode, supplierCode string) float64 {
	t.Helper()

	materialID := mustCreateMaterial(ctx, t, c, categoryID, uniqueChineseName("编码自动分批收货"), true)
	plan := purchaseModulePlan{
		MaterialID:   materialID,
		MaterialName: uniqueChineseName("编码自动分批收货物料"),
		IsCode:       true,
		PurchaseQty:  6,
		UnitPrice:    19.8,
	}

	poID, _ := mustCreatePurchaseOrderBatch(ctx, t, c, supplierCode, []purchaseModulePlan{plan})
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/purchase/orders/%d/confirm", poID))

	po := waitPurchaseOrderDetailExt(ctx, t, c, poID, func(p PurchaseOrderDetailExt) bool { return true })
	if po.StockInID <= 0 {
		if autoID, ok := waitFindStockInIDByPurchaseOrderID(ctx, c, poID); ok {
			po.StockInID = autoID
		}
	}
	if po.StockInID <= 0 {
		t.Fatalf("partial purchase receipt scenario missing stock-in: poID=%d", poID)
	}

	setStockInWarehouseIfEmpty(ctx, t, c, po.StockInID, warehouseCode)

	seenCodes := make(map[string]struct{}, int(plan.PurchaseQty))
	for step := 1; step <= int(plan.PurchaseQty); step++ {
		mustUpdatePurchaseStockInAcceptedQty(ctx, t, c, po.StockInID, warehouseCode, materialID, 1, nil)
		mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/stock-in/%d/confirm", po.StockInID))

		stockIn := mustGetStockInDetail(ctx, t, c, po.StockInID)
		wantStatus := "pending"
		if step == int(plan.PurchaseQty) {
			wantStatus = "passed"
		}
		if stockIn.StockInStatus != wantStatus {
			t.Fatalf("partial purchase receipt status mismatch at step=%d got=%s want=%s", step, stockIn.StockInStatus, wantStatus)
		}

		serials := mustWaitAndFetchStockInSerials(ctx, t, c, po.StockInID, materialID, step)
		if len(serials) != step {
			t.Fatalf("partial purchase receipt serial count mismatch at step=%d got=%d", step, len(serials))
		}
		for _, serial := range serials {
			if serial.SerialCode == "" {
				t.Fatalf("empty serial code generated at step=%d", step)
			}
			seenCodes[serial.SerialCode] = struct{}{}
		}
		if len(seenCodes) != step {
			t.Fatalf("partial purchase receipt serial uniqueness mismatch at step=%d unique=%d", step, len(seenCodes))
		}
	}

	got := mustSumAvailableQty(ctx, t, c, warehouseCode, materialID)
	assertFloatNear(t, got, plan.PurchaseQty, 0.001, "自动分批收货最终库存")

	return plan.PurchaseQty * plan.UnitPrice
}

func mustCreatePurchaseOrderBatch(ctx context.Context, t *testing.T, c *testutil.Client, supplierCode string, plans []purchaseModulePlan) (int64, float64) {
	t.Helper()

	today := time.Now().Format("2006-01-02")
	items := make([]map[string]any, 0, len(plans))
	var total float64
	for _, plan := range plans {
		items = append(items, map[string]any{
			"material_id": plan.MaterialID,
			"quantity":    plan.PurchaseQty,
			"unit_price":  plan.UnitPrice,
		})
		total += plan.PurchaseQty * plan.UnitPrice
	}

	req := map[string]any{
		"order_type":    "purchase",
		"supplier_code": supplierCode,
		"order_date":    today,
		"expected_date": time.Now().AddDate(0, 0, 5).Format("2006-01-02"),
		"remark":        "采购模块专项自动化",
		"items":         items,
	}

	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/purchase/orders", nil, req, &out); err != nil {
		t.Fatalf("create batch purchase order failed: %v", err)
	}
	return out.ID, total
}

func mustCreatePurchaseReturnOrderBatch(ctx context.Context, t *testing.T, c *testutil.Client, supplierCode, warehouseCode string, plans []purchaseModulePlan, inventoryMap map[int64]InventoryAvailable) int64 {
	t.Helper()

	today := time.Now().Format("2006-01-02")
	items := make([]map[string]any, 0, len(plans))
	for _, plan := range plans {
		inv, ok := inventoryMap[plan.MaterialID]
		if !ok {
			t.Fatalf("missing inventory for material_id=%d", plan.MaterialID)
		}
		items = append(items, map[string]any{
			"inventory_id": inv.InventoryID,
			"material_id":  plan.MaterialID,
			"quantity":     plan.ReturnQty,
		})
	}

	req := map[string]any{
		"return_date":    today,
		"return_type":    "purchase_return",
		"supplier_code":  supplierCode,
		"warehouse_code": warehouseCode,
		"remark":         "采购模块专项退货",
		"items":          items,
	}

	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/returns", nil, req, &out); err != nil {
		t.Fatalf("create purchase return order failed: %v", err)
	}
	return out.ID
}

func mustGetStockInDetail(ctx context.Context, t *testing.T, c *testutil.Client, stockInID int64) StockInDetail {
	t.Helper()

	var stockIn StockInDetail
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-in/%d", stockInID), nil, nil, &stockIn); err != nil {
		t.Fatalf("get stock-in detail failed: %v", err)
	}
	return stockIn
}

func mustUpdatePurchaseStockInAcceptedQty(ctx context.Context, t *testing.T, c *testutil.Client, stockInID int64, warehouseCode string, materialID int64, qty float64, customAttrs any) {
	t.Helper()

	stockIn := mustGetStockInDetail(ctx, t, c, stockInID)
	items := make([]map[string]any, 0, len(stockIn.Items))
	found := false
	for _, it := range stockIn.Items {
		unitCost := 0.0
		if it.UnitCost != nil {
			unitCost = *it.UnitCost
		}
		arrivedQty := it.ArrivedQuantity
		acceptedQty := it.AcceptedQuantity
		itemCustomAttrs := any(nil)
		if it.MaterialID == materialID {
			arrivedQty = qty
			acceptedQty = qty
			itemCustomAttrs = customAttrs
			found = true
		}
		items = append(items, map[string]any{
			"id":                it.ID,
			"material_id":       it.MaterialID,
			"arrived_quantity":  arrivedQty,
			"accepted_quantity": acceptedQty,
			"unit_cost":         unitCost,
			"cert_id":           0,
			"custom_attributes": itemCustomAttrs,
		})
	}
	if !found {
		t.Fatalf("stock-in item not found for material=%d stockInID=%d", materialID, stockInID)
	}

	req := map[string]any{
		"warehouse_code": warehouseCode,
		"remark":         "partial receipt one by one",
		"items":          items,
	}
	if err := c.DoJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stock-in/%d", stockInID), nil, req, nil); err != nil {
		t.Fatalf("update stock-in accepted qty failed: %v", err)
	}
}

func mustAssertReturnStockOutSerials(ctx context.Context, t *testing.T, c *testutil.Client, stockOut StockOutDetail, plans []purchaseModulePlan) {
	t.Helper()

	for _, plan := range plans {
		for _, item := range stockOut.Items {
			if item.MaterialID != plan.MaterialID || !plan.IsCode {
				continue
			}
			serials, err := getSerialCodesByStockOutItem(ctx, c, item.ID)
			if err != nil {
				t.Fatalf("get purchase return serial codes failed: material=%d err=%v", plan.MaterialID, err)
			}
			if len(serials) != int(plan.ReturnQty) {
				t.Fatalf("purchase return serial count mismatch material=%d got=%d want=%d", plan.MaterialID, len(serials), int(plan.ReturnQty))
			}
			goto nextPlan
		}
		if plan.IsCode {
			t.Fatalf("purchase return stock-out missing coded material item: material=%d", plan.MaterialID)
		}
	nextPlan:
	}
}

func mustWaitStockInSerials(ctx context.Context, t *testing.T, c *testutil.Client, stockInID, materialID int64, expectCount int) {
	t.Helper()

	deadline := time.Now().Add(20 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		var si StockInDetail
		if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-in/%d", stockInID), nil, nil, &si); err != nil {
			lastErr = err
			time.Sleep(1 * time.Second)
			continue
		}
		for _, item := range si.Items {
			if item.MaterialID != materialID || !item.IsCode {
				continue
			}
			serials, err := getSerialCodesByStockInItem(ctx, c, item.ID)
			if err == nil && len(serials) == expectCount {
				return
			}
			lastErr = err
		}
		time.Sleep(1 * time.Second)
	}
	if lastErr != nil {
		t.Fatalf("wait stock-in serials failed: stockInID=%d materialID=%d err=%v", stockInID, materialID, lastErr)
	}
	t.Fatalf("wait stock-in serials timeout: stockInID=%d materialID=%d expect=%d", stockInID, materialID, expectCount)
}

func mustWaitAndFetchStockInSerials(ctx context.Context, t *testing.T, c *testutil.Client, stockInID, materialID int64, expectCount int) []SerialCodeItem {
	t.Helper()

	deadline := time.Now().Add(20 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		var si StockInDetail
		if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-in/%d", stockInID), nil, nil, &si); err != nil {
			lastErr = err
			time.Sleep(1 * time.Second)
			continue
		}
		for _, item := range si.Items {
			if item.MaterialID != materialID || !item.IsCode {
				continue
			}
			serials, err := getSerialCodesByStockInItem(ctx, c, item.ID)
			if err == nil && len(serials) == expectCount {
				return serials
			}
			lastErr = err
		}
		time.Sleep(1 * time.Second)
	}
	if lastErr != nil {
		t.Fatalf("wait and fetch stock-in serials failed: stockInID=%d materialID=%d err=%v", stockInID, materialID, lastErr)
	}
	t.Fatalf("wait and fetch stock-in serials timeout: stockInID=%d materialID=%d expect=%d", stockInID, materialID, expectCount)
	return nil
}

func getSerialCodesByStockInItem(ctx context.Context, c *testutil.Client, itemID int64) ([]SerialCodeItem, error) {
	paths := []string{
		fmt.Sprintf("/api/v1/serial-codes/stock-in-item/%d", itemID),
		fmt.Sprintf("/api/v1/sku-serial/stock-in-item/%d", itemID),
	}
	return getSerialCodesByPaths(ctx, c, paths)
}

func getSerialCodesByStockOutItem(ctx context.Context, c *testutil.Client, itemID int64) ([]SerialCodeItem, error) {
	paths := []string{
		fmt.Sprintf("/api/v1/serial-codes/stock-out-item/%d", itemID),
		fmt.Sprintf("/api/v1/sku-serial/stock-out-item/%d", itemID),
	}
	return getSerialCodesByPaths(ctx, c, paths)
}

func getSerialCodesByPaths(ctx context.Context, c *testutil.Client, paths []string) ([]SerialCodeItem, error) {
	var lastErr error
	for _, path := range paths {
		var serials []SerialCodeItem
		err := c.DoJSON(ctx, http.MethodGet, path, nil, nil, &serials)
		if err == nil {
			return serials, nil
		}
		if apiErr, ok := err.(*testutil.APIError); ok && apiErr.HTTPStatus == http.StatusNotFound {
			lastErr = err
			continue
		}
		return nil, err
	}
	return nil, lastErr
}

func mustAssertPurchaseModuleInventory(ctx context.Context, t *testing.T, c *testutil.Client, warehouseCode string, plans []purchaseModulePlan) {
	t.Helper()

	for _, plan := range plans {
		got := mustSumAvailableQty(ctx, t, c, warehouseCode, plan.MaterialID)
		want := plan.PurchaseQty - plan.ReturnQty
		assertFloatNear(t, got, want, 0.001, "最终库存")
	}
}

func mustAssertPurchaseModuleLedger(ctx context.Context, t *testing.T, c *testutil.Client, plans []purchaseModulePlan) {
	t.Helper()

	rows := mustFetchLedger(ctx, t, c)
	for _, plan := range plans {
		got := sumLedgerQty(rows, plan.MaterialID)
		want := plan.PurchaseQty - plan.ReturnQty
		assertFloatNear(t, got, want, 0.001, "库存台账")
	}
}

func sumLedgerQty(rows []InventoryMaterialLedgerRow, materialID int64) float64 {
	var sum float64
	for _, row := range rows {
		if row.MaterialID == materialID {
			sum += row.Quantity
		}
	}
	return sum
}

func mustGetSupplierReconciliationSnapshot(ctx context.Context, t *testing.T, c *testutil.Client, supplierCode string) (supplierRecSummary, bool) {
	t.Helper()

	for page := 1; page <= 5; page++ {
		q := url.Values{}
		q.Set("page", fmt.Sprintf("%d", page))
		q.Set("page_size", "100")

		var suppliers []supplierRecSummary
		if err := c.DoPage(ctx, http.MethodGet, "/api/v1/reports/reconciliation/suppliers", q, &suppliers, nil); err != nil {
			if apiErr, ok := err.(*testutil.APIError); ok && apiErr.HTTPStatus == http.StatusNotFound {
				t.Logf("skip supplier reconciliation assertion: route unavailable in current environment")
				return supplierRecSummary{}, false
			}
			t.Fatalf("query supplier reconciliation failed: %v", err)
		}

		for _, row := range suppliers {
			if row.SupplierCode == supplierCode {
				return row, true
			}
		}

		if len(suppliers) < 100 {
			break
		}
	}

	return supplierRecSummary{}, false
}

func mustAssertSupplierReconciliationDelta(ctx context.Context, t *testing.T, c *testutil.Client, supplierCode string, before supplierRecSummary, hasBefore bool, wantDelta float64) {
	t.Helper()

	after, ok := mustGetSupplierReconciliationSnapshot(ctx, t, c, supplierCode)
	if !ok {
		return
	}

	wantPayable := wantDelta
	wantActual := 0.0
	wantBalance := wantDelta
	if hasBefore {
		wantPayable += before.PayableAmount
		wantActual += before.ActualAmount
		wantBalance += before.BalanceAmount
	}

	assertFloatNear(t, after.PayableAmount, wantPayable, 0.01, "供应商应付")
	assertFloatNear(t, after.ActualAmount, wantActual, 0.01, "供应商实付")
	assertFloatNear(t, after.BalanceAmount, wantBalance, 0.01, "供应商余额")
}
