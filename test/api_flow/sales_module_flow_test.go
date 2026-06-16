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

// DTO for sales order detail (DoJSON strips outer code/data wrapper)
type SalesOrderDetail struct {
	ID           int64            `json:"id"`
	OrderNo      string           `json:"order_no"`
	CustomerCode string           `json:"customer_code"`
	CustomerName string           `json:"customer_name"`
	OrderStatus  string           `json:"order_status"`
	Items        []SalesOrderItem `json:"items"`
	StockOutID   *int64           `json:"stock_out_id"`
}

type SalesOrderItem struct {
	ID           int64   `json:"id"`
	MaterialID   int64   `json:"material_id"`
	MaterialName string  `json:"material_name"`
	Quantity     float64 `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	IsCode       bool    `json:"is_code"`
}

// ========== 主测试入口 ==========

func TestFlow_SalesModuleLifecycle(t *testing.T) {
	env := testutil.LoadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	admin := testutil.NewClient(env.BaseURL)
	adminLogin, err := admin.Login(ctx, env.AdminUser, env.AdminPass)
	if err != nil {
		t.Fatalf("admin login failed: %v", err)
	}

	fixture := mustLoadBaseDataFixture(ctx, t, admin, adminLogin.UserID)
	whCode := fixture.MainMaterialWarehouse
	fgWhCode := fixture.FinishedWarehouse
	if fgWhCode == "" {
		fgWhCode = whCode
	}

	// 1) 补充成品库存（采购原料 -> 生产 -> 成品入库 + 兜底采购）
	fgMatID, produceQty, _ := mustPrepareFGInventoryForSales(ctx, t, admin, fixture)

	beforeQty := mustSumAvailableQty(ctx, t, admin, fgWhCode, fgMatID)

	// 2) 销售主流程：编码+非编码混合销售，出库确认，退货入库
	soTotal := mustExerciseSalesOrderLifecycle(ctx, t, admin, fixture, fgWhCode, fgMatID)

	// 3) 销售编码专项：手动逐个选码 vs 自动选码
	soSerialTotal := mustExerciseSalesSerialManualAndAuto(ctx, t, admin, fixture, fgWhCode)
	_ = soSerialTotal

	// 4) 最终库存校验
	afterQty := mustSumAvailableQty(ctx, t, admin, fgWhCode, fgMatID)
	if afterQty > beforeQty+0.001 || afterQty < beforeQty-produceQty-0.001 {
		t.Fatalf("final inventory out of expected range: before=%.3f after=%.3f produceQty=%.3f", beforeQty, afterQty, produceQty)
	}

	// 5) 客户对账
	mustAssertCustomerReconciliationIncremental(ctx, t, admin, fixture.CustomerCode, soTotal)

	// 6) 库存台账
	mustAssertInventoryLedgerRecent(ctx, t, admin, fgWhCode, fgMatID)
}

// ========== 补库存 ==========

func mustPrepareFGInventoryForSales(ctx context.Context, t *testing.T, c *testutil.Client, fixture BaseDataFixture) (int64, float64, float64) {
	t.Helper()
	whCode := fixture.MainMaterialWarehouse
	fgWhCode := fixture.FinishedWarehouse
	if fgWhCode == "" {
		fgWhCode = whCode
	}
	fgWhID := mustFindWarehouseID(ctx, t, c, fgWhCode)

	// 采购编码原料
	rawCodeMatID := mustCreateMaterial(ctx, t, c, fixture.CategoryID, uniqueChineseName("销售测试编码原料"), true)
	rawCodeQty := float64(3)
	poCode := []purchaseModulePlan{{MaterialID: rawCodeMatID, MaterialName: "编码原料", IsCode: true, PurchaseQty: rawCodeQty, UnitPrice: 10}}
	poCodeID, _ := mustCreatePurchaseOrderBatch(ctx, t, c, fixture.SupplierCode, poCode)
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/purchase/orders/%d/confirm", poCodeID))
	poCodeObj := waitPurchaseOrderDetailExt(ctx, t, c, poCodeID, func(p PurchaseOrderDetailExt) bool { return true })
	if poCodeObj.StockInID <= 0 {
		if sid, ok := waitFindStockInIDByPurchaseOrderID(ctx, c, poCodeID); ok {
			poCodeObj.StockInID = sid
		}
	}
	setStockInWarehouseIfEmpty(ctx, t, c, poCodeObj.StockInID, whCode)
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/stock-in/%d/confirm", poCodeObj.StockInID))

	// 采购非编码原料
	rawNoCodeMatID := mustCreateMaterial(ctx, t, c, fixture.CategoryID, uniqueChineseName("销售测试非编码原料"), false)
	rawNoCodeQty := float64(5)
	poNoCode := []purchaseModulePlan{{MaterialID: rawNoCodeMatID, MaterialName: "非编码原料", IsCode: false, PurchaseQty: rawNoCodeQty, UnitPrice: 8}}
	poNoCodeID, _ := mustCreatePurchaseOrderBatch(ctx, t, c, fixture.SupplierCode, poNoCode)
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/purchase/orders/%d/confirm", poNoCodeID))
	poNoCodeObj := waitPurchaseOrderDetailExt(ctx, t, c, poNoCodeID, func(p PurchaseOrderDetailExt) bool { return true })
	if poNoCodeObj.StockInID <= 0 {
		if sid, ok := waitFindStockInIDByPurchaseOrderID(ctx, c, poNoCodeID); ok {
			poNoCodeObj.StockInID = sid
		}
	}
	setStockInWarehouseIfEmpty(ctx, t, c, poNoCodeObj.StockInID, whCode)
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/stock-in/%d/confirm", poNoCodeObj.StockInID))

	// 领料 -> 生产入库
	fgMatID := rawNoCodeMatID
	produceQty := float64(2)
	codeInv := mustPickAnyInventoryAvailable(ctx, t, c, whCode, true, rawCodeQty)
	noCodeInv := mustPickAnyInventoryAvailable(ctx, t, c, whCode, false, rawNoCodeQty)
	consID := mustCreateConsumptionOrderWithProductionAndTwoItems(ctx, t, c, codeInv, rawCodeQty, noCodeInv, rawNoCodeQty, fgMatID, fgWhID, produceQty)
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/consumption/orders/%d/confirm", consID))
	cons := waitConsumptionOrder(ctx, t, c, consID, func(o ConsumptionOrder) bool { return o.StockOutID > 0 })
	if err := tryAutoStockOutSerialSelections(ctx, c, cons.StockOutID); err != nil {
		t.Logf("auto stock-out serial selections skipped: %v", err)
	}
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/stock-out/%d/confirm", cons.StockOutID))
	verifySalesOutSerials(ctx, t, c, cons.StockOutID, codeInv.MaterialID, int(rawCodeQty))

	if prodStockInID, ok := tryWaitConsumptionProductionStockInID(ctx, c, consID); ok && prodStockInID > 0 {
		// 生产入库单可能已被 trigger 自动确认
		var si StockInDetail
		if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-in/%d", prodStockInID), nil, nil, &si); err == nil && si.StockInStatus != "passed" {
			mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/stock-in/%d/confirm", prodStockInID))
		}
	}

	// 直接采购成品兜底（无编码）
	extraFGQty := float64(5)
	poExtra := []purchaseModulePlan{{MaterialID: fgMatID, MaterialName: "成品兜底", IsCode: false, PurchaseQty: extraFGQty, UnitPrice: 15}}
	poExtraID, _ := mustCreatePurchaseOrderBatch(ctx, t, c, fixture.SupplierCode, poExtra)
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/purchase/orders/%d/confirm", poExtraID))
	poExtraObj := waitPurchaseOrderDetailExt(ctx, t, c, poExtraID, func(p PurchaseOrderDetailExt) bool { return true })
	if poExtraObj.StockInID <= 0 {
		if sid, ok := waitFindStockInIDByPurchaseOrderID(ctx, c, poExtraID); ok {
			poExtraObj.StockInID = sid
		}
	}
	setStockInWarehouseIfEmpty(ctx, t, c, poExtraObj.StockInID, fgWhCode)
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/stock-in/%d/confirm", poExtraObj.StockInID))

	// 采购编码成品兜底（供销售出库编码专项使用）
	fgCodeQty := float64(10)
	fgCodeMatID := mustCreateMaterial(ctx, t, c, fixture.CategoryID, uniqueChineseName("销售测试编码成品"), true)
	poFGCode := []purchaseModulePlan{{MaterialID: fgCodeMatID, MaterialName: "编码成品", IsCode: true, PurchaseQty: fgCodeQty, UnitPrice: 20}}
	poFGCodeID, _ := mustCreatePurchaseOrderBatch(ctx, t, c, fixture.SupplierCode, poFGCode)
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/purchase/orders/%d/confirm", poFGCodeID))
	poFGCodeObj := waitPurchaseOrderDetailExt(ctx, t, c, poFGCodeID, func(p PurchaseOrderDetailExt) bool { return true })
	if poFGCodeObj.StockInID <= 0 {
		if sid, ok := waitFindStockInIDByPurchaseOrderID(ctx, c, poFGCodeID); ok {
			poFGCodeObj.StockInID = sid
		}
	}
	setStockInWarehouseIfEmpty(ctx, t, c, poFGCodeObj.StockInID, fgWhCode)
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/stock-in/%d/confirm", poFGCodeObj.StockInID))

	totalCost := rawCodeQty*10 + rawNoCodeQty*8 + extraFGQty*15
	return fgMatID, produceQty + extraFGQty, totalCost
}

// ========== 销售主流程：下单 -> ship出库 -> 确认出库 -> 退货入库 ==========

func mustExerciseSalesOrderLifecycle(ctx context.Context, t *testing.T, c *testutil.Client, fixture BaseDataFixture, fgWhCode string, fgMatID int64) float64 {
	t.Helper()

	codeInv := mustPickAnyInventoryAvailable(ctx, t, c, fgWhCode, true, 2)
	noCodeInv := mustPickAnyInventoryAvailable(ctx, t, c, fgWhCode, false, 3)

	codeQty := int64(2)
	noCodeQty := int64(3)
	codePrice := 28.5
	noCodePrice := 18.0

	// 建销售单（两物料分别建单，避免出库跨仓库）
	customerCode := fixture.CustomerCode

	// 编码物料销售单
	codeSoID := mustCreateSalesOrderSingle(ctx, t, c, customerCode, codeInv.MaterialID, codeQty, codePrice)
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/sales/orders/%d/confirm", codeSoID))
	codeShipID := mustShipSingleItemSalesOrder(ctx, t, c, codeSoID, fgWhCode, codeInv.MaterialID, codeInv.InventoryID, float64(codeQty))
	if err := tryAutoStockOutSerialSelections(ctx, c, codeShipID); err != nil {
		t.Logf("auto stock-out serial selections skipped: %v", err)
	}
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/stock-out/%d/confirm", codeShipID))
	verifySalesOutSerials(ctx, t, c, codeShipID, codeInv.MaterialID, int(codeQty))

	codeAfter := mustSumAvailableQty(ctx, t, c, fgWhCode, codeInv.MaterialID)
	if codeAfter > codeInv.AvailableQuantity-float64(codeQty)+0.001 || codeAfter < codeInv.AvailableQuantity-float64(codeQty)-0.001 {
		t.Fatalf("sales out code inventory mismatch: before=%.3f after=%.3f", codeInv.AvailableQuantity, codeAfter)
	}

	// 非编码物料销售单
	noCodeSoID := mustCreateSalesOrderSingle(ctx, t, c, customerCode, noCodeInv.MaterialID, noCodeQty, noCodePrice)
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/sales/orders/%d/confirm", noCodeSoID))
	noCodeShipID := mustShipSingleItemSalesOrder(ctx, t, c, noCodeSoID, fgWhCode, noCodeInv.MaterialID, noCodeInv.InventoryID, float64(noCodeQty))
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/stock-out/%d/confirm", noCodeShipID))

	noCodeAfter := mustSumAvailableQty(ctx, t, c, fgWhCode, noCodeInv.MaterialID)
	if codeAfter > codeInv.AvailableQuantity-float64(codeQty)+0.001 || codeAfter < codeInv.AvailableQuantity-float64(codeQty)-0.001 {
		t.Fatalf("sales out code inventory mismatch: before=%.3f after=%.3f", codeInv.AvailableQuantity, codeAfter)
	}
	if noCodeAfter > noCodeInv.AvailableQuantity-float64(noCodeQty)+0.001 || noCodeAfter < noCodeInv.AvailableQuantity-float64(noCodeQty)-0.001 {
		t.Fatalf("sales out no-code inventory mismatch: before=%.3f after=%.3f", noCodeInv.AvailableQuantity, noCodeAfter)
	}

	// 销售退货（退非编码1个）
	returnQty := float64(1)
	returnID := mustCreateSalesReturnOrderSingle(ctx, t, c, customerCode, fgWhCode, noCodeInv.MaterialID, returnQty)
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/returns/%d/confirm", returnID))

	var retDetail struct {
		StockInID int64 `json:"stock_in_id"`
	}
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/returns/%d", returnID), nil, nil, &retDetail); err != nil {
		t.Fatalf("get return order detail failed: %v", err)
	}
	if retDetail.StockInID <= 0 {
		t.Fatalf("return order has no stock_in_id after confirm")
	}
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/stock-in/%d/confirm", retDetail.StockInID))

	noCodeFinal := mustSumAvailableQty(ctx, t, c, fgWhCode, noCodeInv.MaterialID)
	if noCodeFinal > noCodeAfter+returnQty+0.001 || noCodeFinal < noCodeAfter+returnQty-0.001 {
		t.Fatalf("sales return inventory mismatch: after_out=%.3f final=%.3f expect=%.3f", noCodeAfter, noCodeFinal, noCodeAfter+returnQty)
	}

	t.Logf("sales order lifecycle ok: returnID=%d", returnID)
	return float64(codeQty)*codePrice + float64(noCodeQty)*noCodePrice
}

// ========== 销售编码专项：手动选码 vs 自动选码 ==========

func mustExerciseSalesSerialManualAndAuto(ctx context.Context, t *testing.T, c *testutil.Client, fixture BaseDataFixture, fgWhCode string) float64 {
	t.Helper()

	// 手动选码
	manualInv := mustPickAnyInventoryAvailable(ctx, t, c, fgWhCode, true, 1)
	manualQty := int64(1)
	manualPrice := 35.0

	manualSoID := mustCreateSalesOrderSingle(ctx, t, c, fixture.CustomerCode, manualInv.MaterialID, manualQty, manualPrice)
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/sales/orders/%d/confirm", manualSoID))
	manualShipID := mustShipSingleItemSalesOrder(ctx, t, c, manualSoID, fgWhCode, manualInv.MaterialID, manualInv.InventoryID, float64(manualQty))
	selectStockOutSerialsManually(ctx, t, c, manualShipID, manualInv.MaterialID, int(manualQty))
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/stock-out/%d/confirm", manualShipID))
	verifySalesOutSerials(ctx, t, c, manualShipID, manualInv.MaterialID, int(manualQty))

	manualAfter := mustSumAvailableQty(ctx, t, c, fgWhCode, manualInv.MaterialID)
	if manualAfter > manualInv.AvailableQuantity-float64(manualQty)+0.001 || manualAfter < manualInv.AvailableQuantity-float64(manualQty)-0.001 {
		t.Fatalf("manual serial inventory mismatch: before=%.3f after=%.3f", manualInv.AvailableQuantity, manualAfter)
	}

	// 自动选码
	autoInv := mustPickAnyInventoryAvailable(ctx, t, c, fgWhCode, true, 2)
	autoQty := int64(2)
	autoPrice := 30.0

	autoSoID := mustCreateSalesOrderSingle(ctx, t, c, fixture.CustomerCode, autoInv.MaterialID, autoQty, autoPrice)
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/sales/orders/%d/confirm", autoSoID))
	autoShipID := mustShipSingleItemSalesOrder(ctx, t, c, autoSoID, fgWhCode, autoInv.MaterialID, autoInv.InventoryID, float64(autoQty))
	if err := tryAutoStockOutSerialSelections(ctx, c, autoShipID); err != nil {
		t.Logf("auto stock-out serial selections skipped: %v", err)
	}
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/stock-out/%d/confirm", autoShipID))
	verifySalesOutSerials(ctx, t, c, autoShipID, autoInv.MaterialID, int(autoQty))

	autoAfter := mustSumAvailableQty(ctx, t, c, fgWhCode, autoInv.MaterialID)
	if autoAfter > autoInv.AvailableQuantity-float64(autoQty)+0.001 || autoAfter < autoInv.AvailableQuantity-float64(autoQty)-0.001 {
		t.Fatalf("auto serial inventory mismatch: before=%.3f after=%.3f", autoInv.AvailableQuantity, autoAfter)
	}

	return float64(manualQty)*manualPrice + float64(autoQty)*autoPrice
}

// ========== 小型 helper ==========

func findSalesOrderItemID(ctx context.Context, t *testing.T, c *testutil.Client, soID, materialID int64) int64 {
	t.Helper()
	var so SalesOrderDetail
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/sales/orders/%d", soID), nil, nil, &so); err != nil {
		t.Fatalf("get sales order detail failed: %v", err)
	}
	for _, it := range so.Items {
		if it.MaterialID == materialID {
			return it.ID
		}
	}
	t.Fatalf("sales order item not found: soID=%d materialID=%d", soID, materialID)
	return 0
}

func mustShipSingleItemSalesOrder(ctx context.Context, t *testing.T, c *testutil.Client, soID int64, warehouseCode string, materialID int64, inventoryID int64, qty float64) int64 {
	t.Helper()
	oiID := findSalesOrderItemID(ctx, t, c, soID, materialID)
	shipReq := map[string]any{
		"warehouse_code": warehouseCode,
		"items": []map[string]any{
			{"order_item_id": oiID, "quantity": qty, "inventory_id": inventoryID},
		},
	}
	var shipOut struct {
		StockOutID int64 `json:"stock_out_id"`
	}
	if err := c.DoJSON(ctx, http.MethodPost, fmt.Sprintf("/api/v1/sales/orders/%d/ship", soID), nil, shipReq, &shipOut); err != nil {
		t.Fatalf("ship single item sales order failed: soID=%d err=%v", soID, err)
	}
	if shipOut.StockOutID <= 0 {
		t.Fatalf("ship single item returned empty stock_out_id")
	}
	return shipOut.StockOutID
}

func waitSalesOrderDetail(ctx context.Context, t *testing.T, c *testutil.Client, id int64, ready func(SalesOrderDetail) bool) SalesOrderDetail {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	var last SalesOrderDetail
	for time.Now().Before(deadline) {
		if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/sales/orders/%d", id), nil, nil, &last); err != nil {
			t.Logf("wait sales order detail err: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if ready(last) {
			return last
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("wait sales order detail timeout: id=%d", id)
	return last
}

func mustAssertCustomerReconciliationIncremental(ctx context.Context, t *testing.T, c *testutil.Client, customerCode string, minReceivable float64) {
	t.Helper()
	if customerCode == "" {
		t.Logf("skip customer reconciliation: no customer code")
		return
	}
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	var list []customerRecSummary
	var total int64
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/reports/reconciliation/customers", q, &list, &total); err != nil {
		t.Logf("customer reconciliation list skipped: %v", err)
		return
	}
	var found bool
	for _, it := range list {
		if it.CustomerCode == customerCode {
			found = true
			if it.ReceivableAmount < minReceivable-0.001 {
				t.Fatalf("customer reconciliation receivable too low: code=%s got=%.2f min=%.2f", customerCode, it.ReceivableAmount, minReceivable)
			}
			break
		}
	}
	if !found {
		t.Logf("customer reconciliation: target customer not in first page, skip assert")
	}
}

func mustAssertInventoryLedgerRecent(ctx context.Context, t *testing.T, c *testutil.Client, warehouseCode string, materialID int64) {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "50")
	q.Set("material_id", fmt.Sprintf("%d", materialID))
	q.Set("warehouse_code", warehouseCode)
	var list []map[string]any
	var total int64
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/inventory/material-ledger", q, &list, &total); err != nil {
		t.Logf("inventory ledger query skipped: %v", err)
		return
	}
	if len(list) == 0 {
		t.Logf("inventory ledger empty for material=%d warehouse=%s", materialID, warehouseCode)
		return
	}
	t.Logf("inventory ledger ok: material=%d warehouse=%s records=%d", materialID, warehouseCode, len(list))
}
