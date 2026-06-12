/**
 * 功能：全流程编码物料测试
 *   采购→入库(自动生成编码)→对账→领料(自动生产单+成品入库生成编码)→退料→销售出库(自动选编码)→对账→销售退库
 *   + 多线程并发出库审核→库存一致性校验
 * 创建时间：2026-06-06
 * 创建人：GPT-5.2
 */

package api_flow

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/redgreat/teweicun/test/testutil"
)

func TestFlow_PurchaseProductionSalesCycle(t *testing.T) {
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
	supplierCode := fixture.SupplierCode
	customerCode := fixture.CustomerCode
	rawWhCode := fixture.MainMaterialWarehouse
	fgWhCode := fixture.FinishedWarehouse
	fgWhID := mustFindWarehouseID(ctx, t, admin, fgWhCode)

	// 原料和成品都是编码物料
	rawMatID := mustCreateMaterial(ctx, t, admin, categoryID, uniqueChineseName("编码钢板"), true)
	fgMatID := mustCreateMaterial(ctx, t, admin, categoryID, uniqueChineseName("编码成品"), true)
	t.Logf("coded materials: rawMatID=%d fgMatID=%d", rawMatID, fgMatID)

	// ========== 1. 采购入库（入库确认自动生成编码）==========
	poID := mustCreatePurchaseOrderSingle(ctx, t, admin, supplierCode, rawMatID, 10)

	po := waitPurchaseOrderDetailExt(ctx, t, admin, poID, func(p PurchaseOrderDetailExt) bool { return true })
	if po.StockInID <= 0 {
		po.StockInID = mustCreateStockInForPurchaseExt(ctx, t, admin, poID, rawWhCode)
	}
	setStockInWarehouseIfEmpty(ctx, t, admin, po.StockInID, rawWhCode)
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-in/%d/confirm", po.StockInID))
	t.Logf("purchase flow done: poID=%d orderNo=%s stockInID=%d", poID, po.OrderNo, po.StockInID)

	// 验证编码自动生成
	rawSerials := verifyStockInSerials(ctx, t, admin, po.StockInID, rawMatID, 10)
	t.Logf("raw serial codes generated: count=%d sample=%s", len(rawSerials), rawSerials[0])

	// ========== 2. 采购对账 ==========
	supplierID := mustFindSupplierID(ctx, t, admin, supplierCode)
	poPayID := mustCreateFundPayment(ctx, t, admin, supplierID, po.OrderNo, po.OrderDate, po.TotalAmount, poID)
	poPay := mustGetFundPayment(ctx, t, admin, poPayID)
	if poPay.Status != "completed" {
		t.Fatalf("fund payment not completed: status=%s", poPay.Status)
	}
	t.Logf("purchase reconciled: paymentID=%d statementNo=%s amount=%.2f",
		poPayID, poPay.StatementNo, poPay.PaymentAmount)

	// ========== 3. 领料+自动生产（生产入库自动生成成品编码）==========
	rawInv := mustPickInventoryAvailable(ctx, t, admin, rawWhCode, rawMatID, true, 10)

	consID := mustCreateConsumptionOrderWithProduction(ctx, t, admin, rawInv, 6, fgMatID, fgWhID, 3)
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/consumption/orders/%d/confirm", consID))

	cons := waitConsumptionOrder(ctx, t, admin, consID, func(o ConsumptionOrder) bool { return o.StockOutID > 0 })

	// 领料出库前自动选择编码
	if err := tryAutoStockOutSerialSelections(ctx, admin, cons.StockOutID); err != nil {
		t.Fatalf("auto serial selection for consumption stock-out failed: %v", err)
	}
	t.Logf("consumption serials selected for stockOutID=%d", cons.StockOutID)

	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-out/%d/confirm", cons.StockOutID))

	var consDetail map[string]any
	if err := admin.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/consumption/orders/%d", consID), nil, nil, &consDetail); err != nil {
		t.Fatalf("get consumption detail failed: %v", err)
	}
	if as, ok := asInt64(consDetail["production_order_id"]); !ok || as <= 0 {
		t.Fatalf("expected production_order_id generated, got=%v", consDetail["production_order_id"])
	}
	prodStockInID, _ := asInt64(consDetail["production_stock_in_id"])
	t.Logf("production generated: consumptionID=%d productionOrderID=%v stockInNo=%v stockInID=%v",
		consID, consDetail["production_order_id"], consDetail["production_stock_in_no"], prodStockInID)

	// 验证生产入库也生成了成品编码
	if prodStockInID > 0 {
		fgSerials := verifyStockInSerials(ctx, t, admin, prodStockInID, fgMatID, 3)
		t.Logf("fg serial codes from production: count=%d sample=%s", len(fgSerials), fgSerials[0])
	}

	rawQtyAfterConsume := mustSumAvailableQty(ctx, t, admin, rawWhCode, rawMatID)
	if rawQtyAfterConsume != 4 {
		t.Fatalf("raw qty after consumption mismatch: got=%.3f expect=4.000", rawQtyAfterConsume)
	}
	fgQtyAfterProduce := mustSumAvailableQty(ctx, t, admin, fgWhCode, fgMatID)
	if fgQtyAfterProduce != 3 {
		t.Fatalf("fg qty after production mismatch: got=%.3f expect=3.000", fgQtyAfterProduce)
	}

	// ========== 4. 退料 ==========
	issued := mustPickInventoryIssued(ctx, t, admin, rawWhCode, rawMatID, true, 2)
	revID := mustCreateReversalOrderSingle(ctx, t, admin, issued, 2)
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/reversal/orders/%d/confirm", revID))
	rev := waitReversalOrder(ctx, t, admin, revID, func(o ReversalOrder) bool { return o.StockInID > 0 })

	// 退料入库需备货编码（自动选择已发编码）
	if err := tryAutoReversalStockInSerialSelections(ctx, admin, rev.StockInID); err != nil {
		t.Fatalf("auto serial selection for reversal stock-in failed: %v", err)
	}
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-in/%d/confirm-reversal", rev.StockInID))

	rawQtyAfterReversal := mustSumAvailableQty(ctx, t, admin, rawWhCode, rawMatID)
	if rawQtyAfterReversal != 6 {
		t.Fatalf("raw qty after reversal mismatch: got=%.3f expect=6.000", rawQtyAfterReversal)
	}

	// ========== 5. 销售出库（自动选编码）==========
	salesID := mustCreateSalesOrderSingle(ctx, t, admin, customerCode, fgMatID, 2, 100)
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/sales/orders/%d/confirm", salesID))
	shipReq := map[string]any{"stock_out_date": time.Now().Format("2006-01-02"), "remark": "e2e ship"}
	var shipOut struct {
		StockOutID int64 `json:"stock_out_id"`
	}
	if err := admin.DoJSON(ctx, http.MethodPost, fmt.Sprintf("/api/v1/sales/orders/%d/ship", salesID), nil, shipReq, &shipOut); err != nil {
		t.Fatalf("ship sales order failed: %v", err)
	}
	if shipOut.StockOutID <= 0 {
		t.Fatalf("ship sales order returned empty stock_out_id")
	}

	// 销售出库自动选编码
	if err := tryAutoStockOutSerialSelections(ctx, admin, shipOut.StockOutID); err != nil {
		t.Fatalf("auto serial selection for sales ship failed: %v", err)
	}
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-out/%d/confirm", shipOut.StockOutID))

	fgQtyAfterSale := mustSumAvailableQty(ctx, t, admin, fgWhCode, fgMatID)
	if fgQtyAfterSale != 1 {
		t.Fatalf("fg qty after sales stock-out mismatch: got=%.3f expect=1.000", fgQtyAfterSale)
	}

	// 验证销售出库的编码已变更为 in_stock→issued
	verifySalesOutSerials(ctx, t, admin, shipOut.StockOutID, fgMatID, 2)

	// ========== 6. 销售对账 ==========
	var soDetail struct {
		OrderNo     string  `json:"order_no"`
		OrderDate   string  `json:"order_date"`
		TotalAmount float64 `json:"total_amount"`
	}
	if err := admin.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/sales/orders/%d", salesID), nil, nil, &soDetail); err != nil {
		t.Fatalf("get sales order detail failed: %v", err)
	}
	customerID := mustFindCustomerID(ctx, t, admin, customerCode)
	soColID := mustCreateFundCollection(ctx, t, admin, customerID, soDetail.OrderNo, soDetail.OrderDate, soDetail.TotalAmount, salesID)
	soCol := mustGetFundCollection(ctx, t, admin, soColID)
	if soCol.Status != "completed" {
		t.Fatalf("fund collection not completed: status=%s", soCol.Status)
	}
	t.Logf("sales reconciled: collectionID=%d statementNo=%s amount=%.2f",
		soColID, soCol.StatementNo, soCol.CollectionAmount)

	// ========== 7. 销售退库 ==========
	srID := mustCreateSalesReturnOrderSingle(ctx, t, admin, customerCode, fgWhCode, fgMatID, 1)
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/returns/%d/confirm", srID))

	var ro struct {
		StockInID int64 `json:"stock_in_id"`
	}
	if err := admin.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/returns/%d", srID), nil, nil, &ro); err != nil {
		t.Fatalf("get return order detail failed: %v", err)
	}
	if ro.StockInID <= 0 {
		t.Fatalf("return order has no stock_in_id after confirm")
	}
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-in/%d/confirm", ro.StockInID))

	fgQtyAfterSalesReturn := mustSumAvailableQty(ctx, t, admin, fgWhCode, fgMatID)
	if fgQtyAfterSalesReturn != 2 {
		t.Fatalf("fg qty after sales return mismatch: got=%.3f expect=2.000", fgQtyAfterSalesReturn)
	}

	// ========== 8. 并发出库审核测试 ==========
	t.Log("========== concurrent stock-out confirm test ==========")
	testConcurrentStockOutConfirm(t, admin, adminLogin.UserID, categoryID, supplierCode, customerCode, rawWhCode)
}

// ========== 并发出库审核 ==========

func testConcurrentStockOutConfirm(t *testing.T, admin *testutil.Client, userID int64, categoryID int64, supplierCode, customerCode, warehouseCode string) {
	ctx := context.Background()

	// 创建编码物料并采购入库 20 个
	matID := mustCreateMaterial(ctx, t, admin, categoryID, uniqueChineseName("并发测试"), true)
	poID := mustCreatePurchaseOrderSingle(ctx, t, admin, supplierCode, matID, 20)

	po := waitPurchaseOrderDetailExt(ctx, t, admin, poID, func(p PurchaseOrderDetailExt) bool { return true })
	if po.StockInID <= 0 {
		po.StockInID = mustCreateStockInForPurchaseExt(ctx, t, admin, poID, warehouseCode)
	}
	setStockInWarehouseIfEmpty(ctx, t, admin, po.StockInID, warehouseCode)
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-in/%d/confirm", po.StockInID))

	serials := verifyStockInSerials(ctx, t, admin, po.StockInID, matID, 20)
	t.Logf("concurrent test: %d serial codes generated", len(serials))

	initialQty := mustSumAvailableQty(ctx, t, admin, warehouseCode, matID)
	if initialQty != 20 {
		t.Fatalf("concurrent initial stock mismatch: got=%.3f expect=20.000", initialQty)
	}

	// 创建 5 个销售订单，每个出库 2 个 → 共 10
	const orderCount = 5
	const qtyPerOrder = float64(2)
	const totalOut = orderCount * int(qtyPerOrder)

	var orders []struct {
		salesID    int64
		stockOutID int64
	}
	for i := 0; i < orderCount; i++ {
		sid := mustCreateSalesOrderSingle(ctx, t, admin, customerCode, matID, int64(qtyPerOrder), 50)
		mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/sales/orders/%d/confirm", sid))

		shipReq := map[string]any{"stock_out_date": time.Now().Format("2006-01-02"), "remark": fmt.Sprintf("concurrent-ship-%d", i)}
		var shipOut struct {
			StockOutID int64 `json:"stock_out_id"`
		}
		if err := admin.DoJSON(ctx, http.MethodPost, fmt.Sprintf("/api/v1/sales/orders/%d/ship", sid), nil, shipReq, &shipOut); err != nil {
			t.Fatalf("concurrent ship order %d failed: %v", i, err)
		}
		if shipOut.StockOutID <= 0 {
			t.Fatalf("concurrent ship order %d returned empty stock_out_id", i)
		}
		// 自动选编码
		if err := tryAutoStockOutSerialSelections(ctx, admin, shipOut.StockOutID); err != nil {
			t.Fatalf("concurrent auto serial selection for stockOutID=%d failed: %v", shipOut.StockOutID, err)
		}
		orders = append(orders, struct {
			salesID    int64
			stockOutID int64
		}{sid, shipOut.StockOutID})
		t.Logf("concurrent prep: order[%d] salesID=%d stockOutID=%d", i, sid, shipOut.StockOutID)
	}

	// 并发出库确认
	var wg sync.WaitGroup
	errCh := make(chan error, orderCount)
	for i, o := range orders {
		wg.Add(1)
		go func(idx int, stockOutID int64) {
			defer wg.Done()
			localCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// 每个 goroutine 用自己的 client
			client := testutil.NewClient(admin.BaseURL)
			client.Token = admin.Token
			if err := client.DoJSON(localCtx, http.MethodPost,
				fmt.Sprintf("/api/v1/stock-out/%d/confirm", stockOutID), nil, nil, nil); err != nil {
				errCh <- fmt.Errorf("goroutine %d stockOutID=%d: %v", idx, stockOutID, err)
				return
			}
		}(i, o.stockOutID)
	}
	wg.Wait()
	close(errCh)

	var errs []error
	for e := range errCh {
		errs = append(errs, e)
	}
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("concurrent error: %v", e)
		}
		t.Fatalf("concurrent stock-out confirm failed: %d errors", len(errs))
	}

	// 验证库存：20 - 10 = 10
	finalQty := mustSumAvailableQty(ctx, t, admin, warehouseCode, matID)
	if finalQty != float64(20-totalOut) {
		t.Fatalf("concurrent stock mismatch after %d confirms: got=%.3f expect=%.3f",
			orderCount, finalQty, float64(20-totalOut))
	}
	t.Logf("concurrent stock-out confirm PASS: initial=20 out=%d final=%.0f", totalOut, finalQty)
}

// ========== 编码验证辅助函数 ==========

// verifyStockInSerials 获取入库单中的编码物料条目并验证编码数量
func verifyStockInSerials(ctx context.Context, t *testing.T, c *testutil.Client, stockInID int64, materialID int64, expectCount int) []string {
	t.Helper()
	var si StockInDetail
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-in/%d", stockInID), nil, nil, &si); err != nil {
		t.Fatalf("get stock-in %d failed: %v", stockInID, err)
	}
	for _, item := range si.Items {
		if item.MaterialID == materialID && item.IsCode {
			var serials []SerialCodeItem
			if err := c.DoJSON(ctx, http.MethodGet,
				fmt.Sprintf("/api/v1/serial-codes/stock-in-item/%d", item.ID), nil, nil, &serials); err != nil {
				t.Fatalf("get serial codes for stock-in-item %d failed: %v", item.ID, err)
			}
			if len(serials) != expectCount {
				t.Fatalf("serial code count mismatch for stock-in %d: got=%d expect=%d",
					stockInID, len(serials), expectCount)
			}
			result := make([]string, len(serials))
			for i, s := range serials {
				result[i] = s.SerialCode
			}
			return result
		}
	}
	t.Fatalf("no coded stock-in item found for materialID=%d in stockInID=%d", materialID, stockInID)
	return nil
}

// verifySalesOutSerials 验证销售出库的编码已正确选择
func verifySalesOutSerials(ctx context.Context, t *testing.T, c *testutil.Client, stockOutID int64, materialID int64, expectCount int) {
	t.Helper()
	var so StockOutDetail
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-out/%d", stockOutID), nil, nil, &so); err != nil {
		t.Fatalf("get stock-out %d failed: %v", stockOutID, err)
	}
	for _, item := range so.Items {
		if item.MaterialID == materialID && item.IsCode {
			var serials []SerialCodeItem
			if err := c.DoJSON(ctx, http.MethodGet,
				fmt.Sprintf("/api/v1/serial-codes/stock-out-item/%d", item.ID), nil, nil, &serials); err != nil {
				t.Fatalf("get serial codes for stock-out-item %d failed: %v", item.ID, err)
			}
			if len(serials) != expectCount {
				t.Fatalf("sales out serial count mismatch: got=%d expect=%d", len(serials), expectCount)
			}
			t.Logf("sales out serials verified: count=%d", len(serials))
			return
		}
	}
	t.Fatalf("no coded stock-out item found for materialID=%d in stockOutID=%d", materialID, stockOutID)
}

// tryAutoReversalStockInSerialSelections 退料入库的编码备货：逐个获取 available-issued 编码并按需选择
func tryAutoReversalStockInSerialSelections(ctx context.Context, c *testutil.Client, stockInID int64) error {
	var si StockInDetail
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-in/%d", stockInID), nil, nil, &si); err != nil {
		return err
	}
	for _, item := range si.Items {
		if !item.IsCode {
			continue
		}
		needQty := int(item.AcceptedQuantity)
		if needQty <= 0 {
			continue
		}

		// 获取 available-issued 编码
		var serials []SerialCodeItem
		if err := c.DoJSON(ctx, http.MethodGet,
			fmt.Sprintf("/api/v1/serial-codes/stock-in-item/%d/available-issued", item.ID), nil, nil, &serials); err != nil {
			return fmt.Errorf("get available-issued serials for stock-in-item %d failed: %w", item.ID, err)
		}
		if len(serials) < needQty {
			return fmt.Errorf("insufficient available-issued serials for stock-in-item %d: have=%d need=%d", item.ID, len(serials), needQty)
		}

		// 只选需要的数量
		ids := make([]int64, 0, needQty)
		for i := 0; i < needQty && i < len(serials); i++ {
			if serials[i].ID > 0 {
				ids = append(ids, serials[i].ID)
			}
		}
		if len(ids) != needQty {
			return fmt.Errorf("failed to collect enough valid serial IDs for stock-in-item %d: got=%d need=%d", item.ID, len(ids), needQty)
		}

		req := map[string]any{"serial_code_ids": ids}
		putURL := fmt.Sprintf("/api/v1/stock-in-item/%d/serial-selections", item.ID)
		// 用带类型的目标接收 PUT 结果，确保捕获 API 错误
		var putResp struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		if err := c.DoJSON(ctx, http.MethodPut, putURL, nil, req, &putResp); err != nil {
			return fmt.Errorf("set serial selections for stock-in-item %d failed: %w", item.ID, err)
		}
		log.Printf("[reversal-serial] stockInID=%d itemID=%d selected %d/%d serials", stockInID, item.ID, len(ids), needQty)
	}
	return nil
}

// ========== 以下为复用的辅助函数 ==========

func mustCreateWarehouse(ctx context.Context, t *testing.T, c *testutil.Client, code, name string, managerID int64) string {
	t.Helper()
	req := map[string]any{
		"warehouse_code": code,
		"warehouse_name": name,
		"warehouse_type": "normal",
		"manager_id":     managerID,
	}
	var out struct {
		ID            int64  `json:"id"`
		WarehouseCode string `json:"warehouse_code"`
	}
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/warehouses", nil, req, &out); err != nil {
		t.Fatalf("create warehouse failed: %v", err)
	}
	if out.WarehouseCode == "" {
		out.WarehouseCode = mustFindLatestWarehouseCodeByName(ctx, t, c, name, code)
	}
	t.Logf("created warehouse: id=%d code=%s", out.ID, out.WarehouseCode)
	return out.WarehouseCode
}

func mustFindWarehouseID(ctx context.Context, t *testing.T, c *testutil.Client, code string) int64 {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	q.Set("warehouse_code", code)
	var list []Warehouse
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/base/warehouses", q, &list, nil); err != nil {
		t.Fatalf("list warehouses failed: %v", err)
	}
	for _, w := range list {
		if w.WarehouseCode == code {
			return w.ID
		}
	}
	t.Fatalf("cannot find warehouse id by code=%s", code)
	return 0
}

func mustCreateCustomer(ctx context.Context, t *testing.T, c *testutil.Client, prefix string) string {
	t.Helper()
	code := prefix + "_CUST"
	req := map[string]any{
		"customer_code":  code,
		"customer_name":  "测试客户",
		"contact_person": "E2E",
		"contact_phone":  "13900000000",
		"address":        "E2E Address",
		"remark":         "e2e",
	}
	var out struct {
		ID           int64  `json:"id"`
		CustomerCode string `json:"customer_code"`
	}
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/customers", nil, req, &out); err != nil {
		t.Fatalf("create customer failed: %v", err)
	}
	if out.CustomerCode == "" {
		out.CustomerCode = mustFindLatestCustomerCodeByName(ctx, t, c, "测试客户", code)
	}
	t.Logf("created customer: id=%d code=%s", out.ID, out.CustomerCode)
	return out.CustomerCode
}

func mustCreatePurchaseOrderSingle(ctx context.Context, t *testing.T, c *testutil.Client, supplierCode string, materialID int64, qty float64) int64 {
	t.Helper()
	today := time.Now().Format("2006-01-02")
	req := map[string]any{
		"order_type":    "purchase",
		"supplier_code": supplierCode,
		"order_date":    today,
		"remark":        "e2e purchase " + today,
		"items": []map[string]any{
			{"material_id": materialID, "quantity": qty, "unit_price": 10},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/purchase/orders", nil, req, &out); err != nil {
		t.Fatalf("create purchase order failed: %v", err)
	}
	return out.ID
}

func mustCreateConsumptionOrderWithProduction(ctx context.Context, t *testing.T, c *testutil.Client, inv InventoryAvailable, consumeQty float64, fgMatID, fgWhID int64, fgQty float64) int64 {
	t.Helper()
	today := time.Now().Format("2006-01-02")
	req := map[string]any{
		"project_no":            "E2E-" + today,
		"product_name":          "E2E Product " + today,
		"order_date":            today,
		"remark":                "e2e consumption with production",
		"produced_material_id":  fgMatID,
		"produced_warehouse_id": fgWhID,
		"produced_quantity":     fgQty,
		"items": []map[string]any{
			{"material_id": inv.MaterialID, "inventory_id": inv.InventoryID, "quantity": consumeQty, "unit": inv.Unit, "remark": "raw"},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/consumption/orders", nil, req, &out); err != nil {
		t.Fatalf("create consumption order failed: %v", err)
	}
	return out.ID
}

func mustCreateReversalOrderSingle(ctx context.Context, t *testing.T, c *testutil.Client, issued InventoryIssued, qty float64) int64 {
	t.Helper()
	today := time.Now().Format("2006-01-02")
	req := map[string]any{
		"project_no":   "E2E-" + today,
		"product_name": "E2E Product " + today,
		"order_date":   today,
		"remark":       "e2e reversal",
		"items": []map[string]any{
			{"inventory_id": issued.InventoryID, "material_id": issued.MaterialID, "quantity": qty, "unit": issued.Unit, "remark": "raw"},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/reversal/orders", nil, req, &out); err != nil {
		t.Fatalf("create reversal order failed: %v", err)
	}
	return out.ID
}

func mustSumAvailableQty(ctx context.Context, t *testing.T, c *testutil.Client, warehouseCode string, materialID int64) float64 {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	q.Set("warehouse_code", warehouseCode)
	var list []InventoryAvailable
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/inventory/available", q, &list, nil); err != nil {
		t.Fatalf("list inventory available failed: %v", err)
	}
	var sum float64
	for _, it := range list {
		if it.MaterialID == materialID {
			sum += it.AvailableQuantity
		}
	}
	return sum
}

func mustCreateSalesOrderSingle(ctx context.Context, t *testing.T, c *testutil.Client, customerCode string, materialID int64, qty int64, unitPrice float64) int64 {
	t.Helper()
	today := time.Now().Format("2006-01-02")
	req := map[string]any{
		"customer_code": customerCode,
		"order_date":    today,
		"remark":        "e2e sales",
		"items": []map[string]any{
			{"material_id": materialID, "quantity": float64(qty), "unit_price": unitPrice, "remark": "fg"},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/sales/orders", nil, req, &out); err != nil {
		t.Fatalf("create sales order failed: %v", err)
	}
	return out.ID
}

func mustCreateSalesReturnOrderSingle(ctx context.Context, t *testing.T, c *testutil.Client, customerCode, warehouseCode string, materialID int64, qty float64) int64 {
	t.Helper()
	today := time.Now().Format("2006-01-02")
	req := map[string]any{
		"return_date":    today,
		"return_type":    "sales_return",
		"customer_code":  customerCode,
		"warehouse_code": warehouseCode,
		"remark":         "e2e sales return",
		"items": []map[string]any{
			{"material_id": materialID, "quantity": qty},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/returns", nil, req, &out); err != nil {
		t.Fatalf("create sales return order failed: %v", err)
	}
	return out.ID
}

// ========== 对账相关 ==========

type PurchaseOrderDetailExt struct {
	ID          int64   `json:"id"`
	StockInID   int64   `json:"stock_in_id"`
	OrderNo     string  `json:"order_no"`
	OrderDate   string  `json:"order_date"`
	TotalAmount float64 `json:"total_amount"`
	SupplierID  int64   `json:"supplier_id"`
	Items       []struct {
		ID         int64   `json:"id"`
		MaterialID int64   `json:"material_id"`
		Quantity   float64 `json:"quantity"`
	} `json:"items"`
}

func waitPurchaseOrderDetailExt(ctx context.Context, t *testing.T, c *testutil.Client, id int64, ready func(PurchaseOrderDetailExt) bool) PurchaseOrderDetailExt {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	var last PurchaseOrderDetailExt
	var lastErr error
	for time.Now().Before(deadline) {
		var po PurchaseOrderDetailExt
		err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/purchase/orders/%d", id), nil, nil, &po)
		if err == nil {
			last = po
			if ready(po) {
				return po
			}
		} else {
			lastErr = err
		}
		time.Sleep(1 * time.Second)
	}
	if lastErr != nil {
		t.Fatalf("wait purchase order detail failed: %v", lastErr)
	}
	t.Fatalf("purchase order %d not ready within timeout", id)
	return last
}

func mustCreateStockInForPurchaseExt(ctx context.Context, t *testing.T, c *testutil.Client, poID int64, warehouseCode string) int64 {
	t.Helper()
	today := time.Now().Format("2006-01-02")
	var po PurchaseOrderDetailExt
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/purchase/orders/%d", poID), nil, nil, &po); err != nil {
		t.Fatalf("get purchase order detail failed: %v", err)
	}
	items := make([]map[string]any, 0, len(po.Items))
	for _, it := range po.Items {
		items = append(items, map[string]any{
			"material_id":       it.MaterialID,
			"purchase_item_id":  it.ID,
			"arrived_quantity":  it.Quantity,
			"accepted_quantity": it.Quantity,
			"unit_price":        10,
			"cert_id":           0,
		})
	}
	req := map[string]any{
		"stock_in_date":     today,
		"stock_in_type":     "purchase",
		"warehouse_code":    warehouseCode,
		"purchase_order_id": poID,
		"remark":            "e2e stock-in for purchase",
		"items":             items,
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/stock-in", nil, req, &out); err != nil {
		t.Fatalf("create stock-in failed: %v", err)
	}
	return out.ID
}

func mustFindSupplierID(ctx context.Context, t *testing.T, c *testutil.Client, code string) int64 {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	var list []Supplier
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/base/suppliers", q, &list, nil); err != nil {
		t.Fatalf("list suppliers failed: %v", err)
	}
	for _, s := range list {
		if s.SupplierCode == code {
			return s.ID
		}
	}
	t.Fatalf("cannot find supplier id by code=%s", code)
	return 0
}

func mustFindCustomerID(ctx context.Context, t *testing.T, c *testutil.Client, code string) int64 {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	q.Set("customer_code", code)
	type CustomerRow struct {
		ID int64 `json:"id"`
	}
	var list []CustomerRow
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/base/customers", q, &list, nil); err != nil {
		t.Fatalf("list customers failed: %v", err)
	}
	if len(list) == 0 {
		t.Fatalf("cannot find customer by code=%s", code)
	}
	return list[0].ID
}

type FundPaymentResp struct {
	ID            int64   `json:"id"`
	StatementNo   string  `json:"statement_no"`
	Status        string  `json:"status"`
	PaymentAmount float64 `json:"payment_amount"`
}

type FundCollectionResp struct {
	ID               int64   `json:"id"`
	StatementNo      string  `json:"statement_no"`
	Status           string  `json:"status"`
	CollectionAmount float64 `json:"collection_amount"`
}

func mustCreateFundPayment(ctx context.Context, t *testing.T, c *testutil.Client, supplierID int64, orderNo, orderDate string, orderAmount float64, poID int64) int64 {
	t.Helper()
	today := time.Now().Format("2006-01-02")
	req := map[string]any{
		"supplier_id":    supplierID,
		"statement_date": today,
		"bill_type":      "cash",
		"payment_amount": orderAmount,
		"invoice_amount": 0,
		"actual_amount":  orderAmount,
		"remark":         "e2e payment",
		"items": []map[string]any{
			{
				"source_doc_type":       "purchase_order",
				"source_order_id":       poID,
				"source_order_no":       orderNo,
				"business_type":         "采购订单",
				"order_date":            orderDate,
				"order_amount":          orderAmount,
				"verified_amount":       0,
				"unverified_amount":     orderAmount,
				"current_verify_amount": orderAmount,
				"custom_tax_amount":     0,
			},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/fund/payments", nil, req, &out); err != nil {
		t.Fatalf("create fund payment failed: %v", err)
	}
	return out.ID
}

func mustGetFundPayment(ctx context.Context, t *testing.T, c *testutil.Client, id int64) FundPaymentResp {
	t.Helper()
	var resp FundPaymentResp
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/fund/payments/%d", id), nil, nil, &resp); err != nil {
		t.Fatalf("get fund payment failed: %v", err)
	}
	return resp
}

func mustCreateFundCollection(ctx context.Context, t *testing.T, c *testutil.Client, customerID int64, orderNo, orderDate string, orderAmount float64, soID int64) int64 {
	t.Helper()
	today := time.Now().Format("2006-01-02")
	req := map[string]any{
		"customer_id":       customerID,
		"statement_date":    today,
		"bill_type":         "cash",
		"collection_amount": orderAmount,
		"invoice_amount":    0,
		"actual_amount":     orderAmount,
		"remark":            "e2e collection",
		"items": []map[string]any{
			{
				"source_doc_type":       "sales_order",
				"source_order_id":       soID,
				"source_order_no":       orderNo,
				"business_type":         "销售订单",
				"order_date":            orderDate,
				"order_amount":          orderAmount,
				"verified_amount":       0,
				"unverified_amount":     orderAmount,
				"current_verify_amount": orderAmount,
				"custom_tax_amount":     0,
			},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/fund/collections", nil, req, &out); err != nil {
		t.Fatalf("create fund collection failed: %v", err)
	}
	return out.ID
}

func mustGetFundCollection(ctx context.Context, t *testing.T, c *testutil.Client, id int64) FundCollectionResp {
	t.Helper()
	var resp FundCollectionResp
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/fund/collections/%d", id), nil, nil, &resp); err != nil {
		t.Fatalf("get fund collection failed: %v", err)
	}
	return resp
}
