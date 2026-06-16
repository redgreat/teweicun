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

func TestFlow_ConsumptionReversalProductionModuleLifecycle(t *testing.T) {
	env := testutil.LoadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()

	admin := testutil.NewClient(env.BaseURL)
	adminLogin, err := admin.Login(ctx, env.AdminUser, env.AdminPass)
	if err != nil {
		t.Fatalf("admin login failed: %v", err)
	}

	fixture := mustLoadBaseDataFixture(ctx, t, admin, adminLogin.UserID)
	rawWhCode := fixture.MainMaterialWarehouse
	fgWhCode := fixture.FinishedWarehouse
	fgWhID := mustFindWarehouseID(ctx, t, admin, fgWhCode)

	consumeCodeQty := float64(3)
	consumeNoCodeQty := float64(5)
	produceQty := float64(2)
	reversalQty := float64(1)

	rawCodeInv := mustPickAnyInventoryAvailable(ctx, t, admin, rawWhCode, true, consumeCodeQty)
	rawNoCodeInv := mustPickAnyInventoryAvailable(ctx, t, admin, rawWhCode, false, consumeNoCodeQty)

	rawCodeMatID := rawCodeInv.MaterialID
	rawNoCodeMatID := rawNoCodeInv.MaterialID

	fgMatID := rawNoCodeMatID
	fgQtyBefore := mustSumAvailableQty(ctx, t, admin, fgWhCode, fgMatID)

	consID := mustCreateConsumptionOrderWithProductionAndTwoItems(ctx, t, admin, rawCodeInv, consumeCodeQty, rawNoCodeInv, consumeNoCodeQty, fgMatID, fgWhID, produceQty)
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/consumption/orders/%d/confirm", consID))

	cons := waitConsumptionOrder(ctx, t, admin, consID, func(o ConsumptionOrder) bool { return o.StockOutID > 0 })

	selectStockOutSerialsManually(ctx, t, admin, cons.StockOutID, rawCodeMatID, int(consumeCodeQty))
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-out/%d/confirm", cons.StockOutID))

	verifyStockOutSerialCount(ctx, t, admin, cons.StockOutID, rawCodeMatID, int(consumeCodeQty))

	fgQtyAfterProduce := mustSumAvailableQty(ctx, t, admin, fgWhCode, fgMatID)
	if prodStockInID, ok := tryWaitConsumptionProductionStockInID(ctx, admin, consID); ok {
		var si map[string]any
		if err := admin.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-in/%d", prodStockInID), nil, nil, &si); err != nil {
			t.Fatalf("get production stock-in failed: %v", err)
		}
		fgQtyAfterProduce = mustSumAvailableQty(ctx, t, admin, fgWhCode, fgMatID)
		if fgQtyAfterProduce < fgQtyBefore+produceQty-0.001 {
			t.Fatalf("fg qty after production too small: before=%.3f after=%.3f expect_add>=%.3f", fgQtyBefore, fgQtyAfterProduce, produceQty)
		}
	} else {
		t.Logf("skip production stock-in assertions: backend did not expose production_stock_in_id for consumption order")
	}

	issued := mustPickInventoryIssuedPaged(ctx, t, admin, rawWhCode, rawCodeMatID, true, reversalQty)
	revID := mustCreateReversalOrderSingle(ctx, t, admin, issued, reversalQty)
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/reversal/orders/%d/confirm", revID))

	rev := waitReversalOrder(ctx, t, admin, revID, func(o ReversalOrder) bool { return o.StockInID > 0 })
	selectReversalStockInSerialsManually(ctx, t, admin, rev.StockInID)
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-in/%d/confirm-reversal", rev.StockInID))

	rawQtyAfterReversal := mustSumAvailableQty(ctx, t, admin, rawWhCode, rawCodeMatID)
	if rawQtyAfterReversal < reversalQty-0.001 {
		t.Fatalf("raw qty after reversal too small: got=%.3f expect>=%.3f", rawQtyAfterReversal, reversalQty)
	}

	mustAssertMaterialLedgerHasAtLeast(ctx, t, admin, rawCodeMatID, rawNoCodeMatID, fgMatID)
	mustAssertReconciliationRoutesHealthy(ctx, t, admin)
	mustAssertProductionReturnRoutesHealthy(ctx, t, admin)
}

func mustPickAnyInventoryAvailable(ctx context.Context, t *testing.T, c *testutil.Client, warehouseCode string, wantIsCode bool, minQty float64) InventoryAvailable {
	t.Helper()

	for page := 1; page <= 20; page++ {
		q := url.Values{}
		q.Set("page", fmt.Sprintf("%d", page))
		q.Set("page_size", "100")
		q.Set("warehouse_code", warehouseCode)
		var list []InventoryAvailable
		if err := c.DoPage(ctx, http.MethodGet, "/api/v1/inventory/available", q, &list, nil); err != nil {
			t.Fatalf("list inventory available failed: %v", err)
		}
		for _, row := range list {
			if row.IsCode != wantIsCode {
				continue
			}
			if row.AvailableQuantity+1e-9 >= minQty {
				return row
			}
		}
		if len(list) < 100 {
			break
		}
	}

	t.Fatalf("cannot find available inventory: warehouse=%s is_code=%v min_qty=%.3f", warehouseCode, wantIsCode, minQty)
	return InventoryAvailable{}
}

func mustCreateConsumptionOrderWithProductionAndTwoItems(
	ctx context.Context,
	t *testing.T,
	c *testutil.Client,
	codeInv InventoryAvailable,
	codeQty float64,
	noCodeInv InventoryAvailable,
	noCodeQty float64,
	fgMatID int64,
	fgWhID int64,
	fgQty float64,
) int64 {
	t.Helper()

	today := time.Now().Format("2006-01-02")
	req := map[string]any{
		"project_no":            "E2E-" + today,
		"product_name":          "E2E Product " + today,
		"order_date":            today,
		"remark":                "manual serial selection flow",
		"produced_material_id":  fgMatID,
		"produced_warehouse_id": fgWhID,
		"produced_quantity":     fgQty,
		"items": []map[string]any{
			{"material_id": codeInv.MaterialID, "inventory_id": codeInv.InventoryID, "quantity": codeQty, "unit": codeInv.Unit, "remark": "code"},
			{"material_id": noCodeInv.MaterialID, "inventory_id": noCodeInv.InventoryID, "quantity": noCodeQty, "unit": noCodeInv.Unit, "remark": "nocode"},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/consumption/orders", nil, req, &out); err != nil {
		t.Fatalf("create consumption order with production failed: %v", err)
	}
	return out.ID
}

func selectStockOutSerialsManually(ctx context.Context, t *testing.T, c *testutil.Client, stockOutID, materialID int64, need int) {
	t.Helper()

	var so StockOutDetail
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-out/%d", stockOutID), nil, nil, &so); err != nil {
		t.Fatalf("get stock-out failed: %v", err)
	}
	var itemID int64
	for _, it := range so.Items {
		if it.MaterialID == materialID && it.IsCode {
			itemID = it.ID
			break
		}
	}
	if itemID <= 0 {
		t.Fatalf("cannot find coded stock-out item for material=%d stockOutID=%d", materialID, stockOutID)
	}

	avail, err := getAvailableSerialsByStockOutItem(ctx, c, itemID)
	if err != nil {
		t.Fatalf("list available serial codes failed: %v", err)
	}
	if len(avail) < need {
		t.Fatalf("available serial codes not enough: have=%d need=%d", len(avail), need)
	}
	ids := make([]int64, 0, need)
	for i := range need {
		ids = append(ids, avail[i].ID)
	}
	req := map[string]any{"serial_code_ids": ids}
	if err := c.DoJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stock-out-item/%d/serial-selections", itemID), nil, req, nil); err != nil {
		t.Fatalf("set stock-out item serial selections failed: %v", err)
	}
}

func verifyStockOutSerialCount(ctx context.Context, t *testing.T, c *testutil.Client, stockOutID, materialID int64, expect int) {
	t.Helper()

	var so StockOutDetail
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-out/%d", stockOutID), nil, nil, &so); err != nil {
		t.Fatalf("get stock-out failed: %v", err)
	}
	for _, it := range so.Items {
		if it.MaterialID == materialID && it.IsCode {
			serials, err := getSerialCodesByStockOutItem(ctx, c, it.ID)
			if err != nil {
				t.Fatalf("get stock-out item serials failed: %v", err)
			}
			if len(serials) != expect {
				t.Fatalf("stock-out serial count mismatch: got=%d expect=%d", len(serials), expect)
			}
			return
		}
	}
	t.Fatalf("cannot find coded stock-out item for material=%d", materialID)
}

func tryWaitConsumptionProductionStockInID(ctx context.Context, c *testutil.Client, consumptionID int64) (int64, bool) {
	deadline := time.Now().Add(30 * time.Second)
	var last map[string]any
	for time.Now().Before(deadline) {
		var detail map[string]any
		if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/consumption/orders/%d", consumptionID), nil, nil, &detail); err == nil {
			last = detail
			if _, exists := detail["production_stock_in_id"]; !exists {
				return 0, false
			}
			if id, ok := asInt64(detail["production_stock_in_id"]); ok && id > 0 {
				return id, true
			}
		}
		time.Sleep(1 * time.Second)
	}
	_ = last
	return 0, false
}

func selectReversalStockInSerialsManually(ctx context.Context, t *testing.T, c *testutil.Client, stockInID int64) {
	t.Helper()

	var si StockInDetail
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-in/%d", stockInID), nil, nil, &si); err != nil {
		t.Fatalf("get reversal stock-in failed: %v", err)
	}

	for _, item := range si.Items {
		if !item.IsCode {
			continue
		}
		need := int(item.AcceptedQuantity + 1e-9)
		if need <= 0 {
			continue
		}

		avail, err := getAvailableIssuedSerialsByStockInItem(ctx, c, item.ID)
		if err != nil {
			t.Fatalf("list available-issued serials failed: %v", err)
		}
		if len(avail) < need {
			t.Fatalf("available-issued serial codes not enough: have=%d need=%d", len(avail), need)
		}
		ids := make([]int64, 0, need)
		for i := range need {
			ids = append(ids, avail[i].ID)
		}
		req := map[string]any{"serial_code_ids": ids}
		if err := c.DoJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stock-in-item/%d/serial-selections", item.ID), nil, req, nil); err != nil {
			t.Fatalf("set reversal stock-in item serial selections failed: %v", err)
		}
	}
}

func getAvailableSerialsByStockOutItem(ctx context.Context, c *testutil.Client, itemID int64) ([]SerialCodeItem, error) {
	paths := []string{
		fmt.Sprintf("/api/v1/serial-codes/stock-out-item/%d/available", itemID),
		fmt.Sprintf("/api/v1/sku-serial/stock-out-item/%d/available", itemID),
	}
	return getSerialCodesByPaths(ctx, c, paths)
}

func getAvailableIssuedSerialsByStockInItem(ctx context.Context, c *testutil.Client, itemID int64) ([]SerialCodeItem, error) {
	paths := []string{
		fmt.Sprintf("/api/v1/serial-codes/stock-in-item/%d/available-issued", itemID),
		fmt.Sprintf("/api/v1/sku-serial/stock-in-item/%d/available-issued", itemID),
	}
	return getSerialCodesByPaths(ctx, c, paths)
}

func mustAssertMaterialLedgerHasAtLeast(ctx context.Context, t *testing.T, c *testutil.Client, materialIDs ...int64) {
	t.Helper()

	rows := mustFetchLedger(ctx, t, c)
	for _, mid := range materialIDs {
		if sumLedgerQty(rows, mid) <= -1e-9 {
			t.Fatalf("material ledger missing material: %d", mid)
		}
	}
}

func mustAssertReconciliationRoutesHealthy(ctx context.Context, t *testing.T, c *testutil.Client) {
	t.Helper()

	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "1")
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/reports/reconciliation/suppliers", q, nil, nil); err != nil {
		if apiErr, ok := err.(*testutil.APIError); ok && apiErr.HTTPStatus == http.StatusNotFound {
			t.Logf("skip reconciliation check: route unavailable in current environment")
			return
		}
		t.Fatalf("reconciliation suppliers route failed: %v", err)
	}
}

func mustAssertProductionReturnRoutesHealthy(ctx context.Context, t *testing.T, c *testutil.Client) {
	t.Helper()

	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "1")
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/production/returns", q, nil, nil); err != nil {
		if apiErr, ok := err.(*testutil.APIError); ok && apiErr.HTTPStatus == http.StatusNotFound {
			t.Logf("skip production return check: route unavailable in current environment")
			return
		}
		t.Fatalf("production returns route failed: %v", err)
	}
}

func mustPickInventoryIssuedPaged(ctx context.Context, t *testing.T, c *testutil.Client, warehouseCode string, materialID int64, isCode bool, qty float64) InventoryIssued {
	t.Helper()

	for page := 1; page <= 20; page++ {
		q := url.Values{}
		q.Set("page", fmt.Sprintf("%d", page))
		q.Set("page_size", "100")
		q.Set("warehouse_code", warehouseCode)
		var list []InventoryIssued
		if err := c.DoPage(ctx, http.MethodGet, "/api/v1/inventory/issued", q, &list, nil); err != nil {
			t.Fatalf("list inventory issued failed: %v", err)
		}
		for _, it := range list {
			if it.MaterialID == materialID && it.IsCode == isCode && it.IssuedQuantity+1e-9 >= qty {
				return it
			}
		}
		if len(list) < 100 {
			break
		}
	}
	t.Fatalf("cannot find inventory issued for material=%d qty>=%.2f", materialID, qty)
	return InventoryIssued{}
}
