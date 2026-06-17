package api_flow

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redgreat/teweicun/test/testutil"
)

type materialLedgerSerialRow struct {
	SerialCode        string `json:"serial_code"`
	Status            string `json:"status"`
	StatusName        string `json:"status_name"`
	DisplayStatus     string `json:"display_status"`
	DisplayStatusName string `json:"display_status_name"`
}

func TestFlow_ManualStockInLifecycle(t *testing.T) {
	env := testutil.LoadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	admin, fixture := mustLoginAndLoadFixture(ctx, t, env)
	materialID := mustCreateMaterial(ctx, t, admin, fixture.CategoryID, uniqueChineseName("手动入库编码件"), true)
	manualCodes := []string{
		"MIN" + testutil.UniquePrefix() + "01",
		"MIN" + testutil.UniquePrefix() + "02",
	}

	stockInID := mustCreateManualStockIn(ctx, t, admin, fixture.MainMaterialWarehouse, materialID, float64(len(manualCodes)), 23.5, "purchase")
	mustSetStockInItemManualSerialCodes(ctx, t, admin, stockInID, materialID, manualCodes)
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-in/%d/confirm", stockInID))

	stockIn := mustGetStockInDetail(ctx, t, admin, stockInID)
	if stockIn.StockInStatus != "passed" {
		t.Fatalf("manual stock-in status mismatch: got=%s want=passed", stockIn.StockInStatus)
	}

	serials := mustWaitAndFetchStockInSerials(ctx, t, admin, stockInID, materialID, len(manualCodes))
	gotCodes := map[string]struct{}{}
	for _, serial := range serials {
		gotCodes[serial.SerialCode] = struct{}{}
	}
	for _, code := range manualCodes {
		if _, ok := gotCodes[code]; !ok {
			t.Fatalf("manual stock-in missing serial code: %s", code)
		}
	}

	gotQty := mustSumAvailableQty(ctx, t, admin, fixture.MainMaterialWarehouse, materialID)
	assertFloatNear(t, gotQty, float64(len(manualCodes)), 0.001, "手动入库库存")

	pool := openOptionalDBPool(ctx, t)
	mustAssertDBManualStockIn(ctx, t, pool, stockInID, materialID, len(manualCodes))
}

func TestFlow_ManualStockOutSerialSelectionLifecycle(t *testing.T) {
	env := testutil.LoadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	admin, fixture := mustLoginAndLoadFixture(ctx, t, env)
	materialID, inv := mustSeedManualStockInInventory(ctx, t, admin, fixture, "手动出库编码件", 4, 31.2)

	stockOutID := mustCreateManualStockOut(ctx, t, admin, materialID, inv.InventoryID, 2, "other", "", 0)
	itemID := mustFindStockOutItemID(ctx, t, admin, stockOutID, materialID)
	selectedIDs := mustPickAvailableStockOutSerialIDs(ctx, t, admin, itemID, 2)

	if err := admin.DoJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stock-out-item/%d/serial-selections", itemID), nil, map[string]any{"serial_code_ids": selectedIDs}, nil); err != nil {
		t.Fatalf("manual stock-out serial selection failed: %v", err)
	}
	mustAssertStockOutSelectedSerialCount(ctx, t, admin, itemID, 2)

	if err := admin.DoJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stock-out-item/%d/serial-selections", itemID), nil, map[string]any{"serial_code_ids": selectedIDs[:1]}, nil); err == nil {
		t.Fatalf("manual stock-out under-selection should fail")
	}
	mustAssertStockOutSelectedSerialCount(ctx, t, admin, itemID, 2)

	if err := admin.DoJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stock-out-item/%d/serial-selections", itemID), nil, map[string]any{"serial_code_ids": []int64{}}, nil); err != nil {
		t.Fatalf("clear manual stock-out serial selection failed: %v", err)
	}
	mustAssertStockOutSelectedSerialCount(ctx, t, admin, itemID, 0)

	if err := admin.DoJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stock-out-item/%d/serial-selections", itemID), nil, map[string]any{"serial_code_ids": selectedIDs}, nil); err != nil {
		t.Fatalf("reselect manual stock-out serials failed: %v", err)
	}
	mustAssertStockOutSelectedSerialCount(ctx, t, admin, itemID, 2)

	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-out/%d/confirm", stockOutID))
	verifyStockOutSerialCount(ctx, t, admin, stockOutID, materialID, 2)

	afterQty := mustSumAvailableQty(ctx, t, admin, fixture.MainMaterialWarehouse, materialID)
	assertFloatNear(t, afterQty, 2, 0.001, "手动出库后库存")
}

func TestFlow_SerialSelectionReserveLedgerState(t *testing.T) {
	env := testutil.LoadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	admin, fixture := mustLoginAndLoadFixture(ctx, t, env)
	pool := openOptionalDBPool(ctx, t)
	materialID, inv := mustSeedManualStockInInventory(ctx, t, admin, fixture, "备货台账编码件", 5, 41.5)

	stockOutID := mustCreateManualStockOut(ctx, t, admin, materialID, inv.InventoryID, 2, "other", "", 0)
	stockOutItemID := mustFindStockOutItemID(ctx, t, admin, stockOutID, materialID)
	stockOutSerialIDs := mustPickAvailableStockOutSerialIDs(ctx, t, admin, stockOutItemID, 2)
	if err := admin.DoJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stock-out-item/%d/serial-selections", stockOutItemID), nil, map[string]any{"serial_code_ids": stockOutSerialIDs}, nil); err != nil {
		t.Fatalf("reserve stock-out serials failed: %v", err)
	}

	mustAssertLedgerReserve(ctx, t, admin, materialID, fixture.MainMaterialWarehouse, 2, "stock_out_reserved", "出库备货中")
	mustAssertDBStockOutReservation(ctx, t, pool, stockOutItemID, stockOutSerialIDs)

	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-out/%d/confirm", stockOutID))
	mustAssertLedgerReserve(ctx, t, admin, materialID, fixture.MainMaterialWarehouse, 0, "", "")
	mustAssertDBStockOutConfirmed(ctx, t, pool, stockOutID, stockOutItemID, stockOutSerialIDs)

	consumptionStockOutID := mustCreateManualStockOut(ctx, t, admin, materialID, inv.InventoryID, 1, "consumption", "consumption_order", 0)
	consumptionItemID := mustFindStockOutItemID(ctx, t, admin, consumptionStockOutID, materialID)
	consumptionSerialIDs := mustPickAvailableStockOutSerialIDs(ctx, t, admin, consumptionItemID, 1)
	if err := admin.DoJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stock-out-item/%d/serial-selections", consumptionItemID), nil, map[string]any{"serial_code_ids": consumptionSerialIDs}, nil); err != nil {
		t.Fatalf("reserve consumption serial failed: %v", err)
	}
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-out/%d/confirm", consumptionStockOutID))

	reversalStockInID := mustCreateManualStockIn(ctx, t, admin, fixture.MainMaterialWarehouse, materialID, 1, 41.5, "reversal")
	reversalItemID := findStockInItemIDByMaterial(mustGetStockInDetail(ctx, t, admin, reversalStockInID), materialID)
	if reversalItemID <= 0 {
		t.Fatalf("reversal stock-in item not found: stockInID=%d materialID=%d", reversalStockInID, materialID)
	}
	reversalSerialIDs := mustPickAvailableIssuedStockInSerialIDs(ctx, t, admin, reversalItemID, 1)
	if err := admin.DoJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stock-in-item/%d/serial-selections", reversalItemID), nil, map[string]any{"serial_code_ids": reversalSerialIDs}, nil); err != nil {
		t.Fatalf("reserve reversal stock-in serial failed: %v", err)
	}

	mustAssertLedgerReserve(ctx, t, admin, materialID, fixture.MainMaterialWarehouse, 1, "stock_in_reserved", "退料备货中")
	mustAssertDBStockInReservation(ctx, t, pool, reversalItemID, reversalSerialIDs)
}

func TestFlow_ConcurrentSerialReserveSameSerialConflict(t *testing.T) {
	env := testutil.LoadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	admin, fixture := mustLoginAndLoadFixture(ctx, t, env)
	pool := openOptionalDBPool(ctx, t)
	materialID, inv := mustSeedManualStockInInventory(ctx, t, admin, fixture, "并发抢编码备货", 2, 35.8)

	stockOutA := mustCreateManualStockOut(ctx, t, admin, materialID, inv.InventoryID, 1, "other", "", 0)
	stockOutB := mustCreateManualStockOut(ctx, t, admin, materialID, inv.InventoryID, 1, "other", "", 0)
	itemA := mustFindStockOutItemID(ctx, t, admin, stockOutA, materialID)
	itemB := mustFindStockOutItemID(ctx, t, admin, stockOutB, materialID)
	serialIDs := mustPickAvailableStockOutSerialIDs(ctx, t, admin, itemA, 1)

	type reserveResult struct {
		itemID int64
		err    error
	}
	results := make(chan reserveResult, 2)
	var wg sync.WaitGroup
	for _, itemID := range []int64{itemA, itemB} {
		wg.Add(1)
		go func(itemID int64) {
			defer wg.Done()
			localCtx, localCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer localCancel()
			c := testutil.NewClient(admin.BaseURL)
			c.Token = admin.Token
			err := c.DoJSON(localCtx, http.MethodPut, fmt.Sprintf("/api/v1/stock-out-item/%d/serial-selections", itemID), nil, map[string]any{"serial_code_ids": serialIDs}, nil)
			results <- reserveResult{itemID: itemID, err: err}
		}(itemID)
	}
	wg.Wait()
	close(results)

	successItems := make([]int64, 0, 1)
	failures := 0
	for result := range results {
		if result.err == nil {
			successItems = append(successItems, result.itemID)
			continue
		}
		failures++
		t.Logf("expected concurrent reserve conflict: itemID=%d err=%v", result.itemID, result.err)
	}
	if len(successItems) != 1 || failures != 1 {
		t.Fatalf("same serial concurrent reserve should allow exactly one winner: successes=%d failures=%d", len(successItems), failures)
	}

	mustAssertStockOutSelectedSerialCount(ctx, t, admin, successItems[0], 1)
	mustAssertDBSingleStockOutReservationForSerial(ctx, t, pool, serialIDs[0])
}

func TestFlow_ReturnAndReversalFailureRollback(t *testing.T) {
	env := testutil.LoadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	admin, fixture := mustLoginAndLoadFixture(ctx, t, env)
	pool := openOptionalDBPool(ctx, t)
	materialID, inv := mustSeedManualStockInInventory(ctx, t, admin, fixture, "退货退料失败回滚", 3, 52.6)

	returnID := mustCreatePurchaseReturnOrderBatch(ctx, t, admin, fixture.SupplierCode, fixture.MainMaterialWarehouse, []purchaseModulePlan{{
		MaterialID:  materialID,
		IsCode:      true,
		ReturnQty:   1,
		PurchaseQty: 3,
		UnitPrice:   52.6,
	}}, map[int64]InventoryAvailable{materialID: inv})
	mustConfirmReturnOrder(ctx, t, admin, returnID, fixture.MainMaterialWarehouse)
	ret := waitReturnOrder(ctx, t, admin, returnID, func(o ReturnOrder) bool {
		return o.StockOutID != nil && *o.StockOutID > 0
	})
	if ret.StockOutID == nil {
		t.Fatalf("purchase return did not generate stock-out")
	}
	beforePurchaseReturnConfirmQty := mustSumAvailableQty(ctx, t, admin, fixture.MainMaterialWarehouse, materialID)
	purchaseReturnItemID := mustFindStockOutItemID(ctx, t, admin, *ret.StockOutID, materialID)
	if err := admin.DoJSON(ctx, http.MethodPost, fmt.Sprintf("/api/v1/stock-out/%d/confirm", *ret.StockOutID), nil, nil, nil); err == nil {
		t.Fatalf("purchase return stock-out confirm without serial selection should fail")
	}
	afterPurchaseReturnQty := mustSumAvailableQty(ctx, t, admin, fixture.MainMaterialWarehouse, materialID)
	assertFloatNear(t, afterPurchaseReturnQty, beforePurchaseReturnConfirmQty, 0.001, "采购退货失败回滚库存")
	mustAssertStockOutPending(ctx, t, admin, *ret.StockOutID)
	mustAssertDBNoStockOutSideEffects(ctx, t, pool, *ret.StockOutID, purchaseReturnItemID, materialID)

	consumptionStockOutID := mustCreateManualStockOut(ctx, t, admin, materialID, inv.InventoryID, 1, "consumption", "consumption_order", 0)
	consumptionItemID := mustFindStockOutItemID(ctx, t, admin, consumptionStockOutID, materialID)
	consumptionSerialIDs := mustPickAvailableStockOutSerialIDs(ctx, t, admin, consumptionItemID, 1)
	if err := admin.DoJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stock-out-item/%d/serial-selections", consumptionItemID), nil, map[string]any{"serial_code_ids": consumptionSerialIDs}, nil); err != nil {
		t.Fatalf("reserve consumption serial failed: %v", err)
	}
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-out/%d/confirm", consumptionStockOutID))

	beforeReversalQty := mustSumAvailableQty(ctx, t, admin, fixture.MainMaterialWarehouse, materialID)
	reversalStockInID := mustCreateManualStockIn(ctx, t, admin, fixture.MainMaterialWarehouse, materialID, 1, 52.6, "reversal")
	reversalBefore := mustGetStockInDetail(ctx, t, admin, reversalStockInID)
	reversalItemID := findStockInItemIDByMaterial(reversalBefore, materialID)
	if reversalItemID <= 0 {
		t.Fatalf("reversal stock-in item not found: stockInID=%d materialID=%d", reversalStockInID, materialID)
	}
	if err := admin.DoJSON(ctx, http.MethodPost, fmt.Sprintf("/api/v1/stock-in/%d/confirm-reversal", reversalStockInID), nil, nil, nil); err == nil {
		t.Fatalf("reversal stock-in confirm without serial selection should fail")
	}
	afterReversalQty := mustSumAvailableQty(ctx, t, admin, fixture.MainMaterialWarehouse, materialID)
	assertFloatNear(t, afterReversalQty, beforeReversalQty, 0.001, "退料失败回滚库存")
	mustAssertStockInStatus(ctx, t, admin, reversalStockInID, reversalBefore.StockInStatus)
	mustAssertDBNoStockInSideEffects(ctx, t, pool, reversalStockInID, reversalItemID, consumptionSerialIDs)
}

func mustLoginAndLoadFixture(ctx context.Context, t *testing.T, env testutil.Env) (*testutil.Client, BaseDataFixture) {
	t.Helper()
	admin := testutil.NewClient(env.BaseURL)
	login, err := admin.Login(ctx, env.AdminUser, env.AdminPass)
	if err != nil {
		t.Fatalf("admin login failed: %v", err)
	}
	return admin, mustLoadBaseDataFixture(ctx, t, admin, login.UserID)
}

func mustSeedManualStockInInventory(ctx context.Context, t *testing.T, c *testutil.Client, fixture BaseDataFixture, name string, qty int, unitCost float64) (int64, InventoryAvailable) {
	t.Helper()
	materialID := mustCreateMaterial(ctx, t, c, fixture.CategoryID, uniqueChineseName(name), true)
	stockInID := mustCreateManualStockIn(ctx, t, c, fixture.MainMaterialWarehouse, materialID, float64(qty), unitCost, "purchase")
	mustConfirm(ctx, t, c, fmt.Sprintf("/api/v1/stock-in/%d/confirm", stockInID))
	mustWaitStockInSerials(ctx, t, c, stockInID, materialID, qty)
	inv := mustPickInventoryAvailable(ctx, t, c, fixture.MainMaterialWarehouse, materialID, true, float64(qty))
	return materialID, inv
}

func mustCreateManualStockIn(ctx context.Context, t *testing.T, c *testutil.Client, warehouseCode string, materialID int64, qty float64, unitCost float64, stockInType string) int64 {
	t.Helper()
	req := map[string]any{
		"stock_in_date":  time.Now().Format("2006-01-02"),
		"stock_in_type":  stockInType,
		"warehouse_code": warehouseCode,
		"remark":         "api flow manual stock-in",
		"items": []map[string]any{
			{
				"material_id":       materialID,
				"arrived_quantity":  qty,
				"accepted_quantity": qty,
				"unit_price":        unitCost,
				"cert_id":           0,
			},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/stock-in", nil, req, &out); err != nil {
		t.Fatalf("create manual stock-in failed: %v", err)
	}
	if out.ID <= 0 {
		t.Fatalf("create manual stock-in returned empty id")
	}
	return out.ID
}

func mustSetStockInItemManualSerialCodes(ctx context.Context, t *testing.T, c *testutil.Client, stockInID, materialID int64, serialCodes []string) {
	t.Helper()
	customAttrs := map[string]any{"manual_serial_codes": serialCodes}
	mustUpdatePurchaseStockInAcceptedQty(ctx, t, c, stockInID, "", materialID, float64(len(serialCodes)), customAttrs)
}

func mustCreateManualStockOut(ctx context.Context, t *testing.T, c *testutil.Client, materialID, inventoryID int64, qty float64, outType, refDocType string, refDocID int64) int64 {
	t.Helper()
	req := map[string]any{
		"stock_out_date": time.Now().Format("2006-01-02"),
		"out_type":       outType,
		"ref_doc_type":   refDocType,
		"ref_doc_id":     refDocID,
		"receiver":       "api-flow",
		"remark":         "api flow manual stock-out",
		"items": []map[string]any{
			{
				"material_id":  materialID,
				"inventory_id": inventoryID,
				"quantity":     qty,
			},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/stock-out", nil, req, &out); err != nil {
		t.Fatalf("create manual stock-out failed: %v", err)
	}
	if out.ID <= 0 {
		t.Fatalf("create manual stock-out returned empty id")
	}
	return out.ID
}

func mustFindStockOutItemID(ctx context.Context, t *testing.T, c *testutil.Client, stockOutID, materialID int64) int64 {
	t.Helper()
	var so StockOutDetail
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-out/%d", stockOutID), nil, nil, &so); err != nil {
		t.Fatalf("get stock-out failed: %v", err)
	}
	for _, item := range so.Items {
		if item.MaterialID == materialID {
			return item.ID
		}
	}
	t.Fatalf("stock-out item not found: stockOutID=%d materialID=%d", stockOutID, materialID)
	return 0
}

func mustPickAvailableStockOutSerialIDs(ctx context.Context, t *testing.T, c *testutil.Client, stockOutItemID int64, need int) []int64 {
	t.Helper()
	serials, err := getAvailableSerialsByStockOutItem(ctx, c, stockOutItemID)
	if err != nil {
		t.Fatalf("get available stock-out serials failed: %v", err)
	}
	if len(serials) < need {
		t.Fatalf("not enough available stock-out serials: have=%d need=%d", len(serials), need)
	}
	ids := make([]int64, 0, need)
	for i := 0; i < need; i++ {
		ids = append(ids, serials[i].ID)
	}
	return ids
}

func mustPickAvailableIssuedStockInSerialIDs(ctx context.Context, t *testing.T, c *testutil.Client, stockInItemID int64, need int) []int64 {
	t.Helper()
	serials, err := getAvailableIssuedSerialsByStockInItem(ctx, c, stockInItemID)
	if err != nil {
		t.Fatalf("get available-issued stock-in serials failed: %v", err)
	}
	if len(serials) < need {
		t.Fatalf("not enough available-issued serials: have=%d need=%d", len(serials), need)
	}
	ids := make([]int64, 0, need)
	for i := 0; i < need; i++ {
		ids = append(ids, serials[i].ID)
	}
	return ids
}

func mustAssertStockOutSelectedSerialCount(ctx context.Context, t *testing.T, c *testutil.Client, stockOutItemID int64, want int) {
	t.Helper()
	serials, err := getSerialCodesByStockOutItem(ctx, c, stockOutItemID)
	if err != nil {
		t.Fatalf("get selected stock-out serials failed: %v", err)
	}
	if len(serials) != want {
		t.Fatalf("selected stock-out serial count mismatch: got=%d want=%d", len(serials), want)
	}
}

func mustAssertLedgerReserve(ctx context.Context, t *testing.T, c *testutil.Client, materialID int64, warehouseCode string, wantReserved float64, wantDisplayStatus, wantDisplayStatusName string) {
	t.Helper()
	row := mustFindMaterialLedgerRow(ctx, t, c, materialID, warehouseCode)
	assertFloatNear(t, row.BookQuantity-row.LockedQuantity-row.InTransitQuantity, row.Quantity, 0.001, "台账可用数量公式")
	assertFloatNear(t, row.SerialReservedQuantity, wantReserved, 0.001, "台账编码备货数")

	if wantReserved <= 0 {
		return
	}
	serials := mustFetchMaterialLedgerSerials(ctx, t, c, row.MaterialID, row.WarehouseID, row.UnitCost)
	var matched int
	for _, serial := range serials {
		if serial.DisplayStatus == wantDisplayStatus && serial.DisplayStatusName == wantDisplayStatusName {
			matched++
		}
	}
	if matched != int(wantReserved) {
		t.Fatalf("ledger serial reserved status mismatch: got=%d want=%.0f status=%s name=%s", matched, wantReserved, wantDisplayStatus, wantDisplayStatusName)
	}
}

func mustFindMaterialLedgerRow(ctx context.Context, t *testing.T, c *testutil.Client, materialID int64, warehouseCode string) InventoryMaterialLedgerRow {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	var rows []InventoryMaterialLedgerRow
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/inventory/material-ledger", q, &rows, nil); err != nil {
		t.Fatalf("fetch material ledger failed: %v", err)
	}
	for _, row := range rows {
		if row.MaterialID == materialID && (warehouseCode == "" || row.WarehouseName != "") {
			return row
		}
	}
	t.Fatalf("material ledger row not found: materialID=%d warehouseCode=%s", materialID, warehouseCode)
	return InventoryMaterialLedgerRow{}
}

func mustFetchMaterialLedgerSerials(ctx context.Context, t *testing.T, c *testutil.Client, materialID, warehouseID int64, unitCost float64) []materialLedgerSerialRow {
	t.Helper()
	q := url.Values{}
	q.Set("material_id", fmt.Sprintf("%d", materialID))
	q.Set("warehouse_id", fmt.Sprintf("%d", warehouseID))
	q.Set("unit_cost", fmt.Sprintf("%.6f", unitCost))
	var rows []materialLedgerSerialRow
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/inventory/material-ledger/serials", q, nil, &rows); err != nil {
		t.Fatalf("fetch material ledger serials failed: %v", err)
	}
	return rows
}

func openOptionalDBPool(ctx context.Context, t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("TWC_DB_DSN"))
	if dsn == "" {
		t.Logf("skip db-level assertions: TWC_DB_DSN is not set")
		return nil
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Logf("skip db-level assertions: connect test database failed: %v", err)
		return nil
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Logf("skip db-level assertions: ping test database failed: %v", err)
		return nil
	}
	t.Cleanup(pool.Close)
	return pool
}

func mustAssertDBManualStockIn(ctx context.Context, t *testing.T, pool *pgxpool.Pool, stockInID, materialID int64, wantSerials int) {
	t.Helper()
	if pool == nil {
		return
	}

	var itemUnit, materialUnit string
	var itemCount, serialCount, inStockCount int
	if err := pool.QueryRow(ctx, `
		SELECT sii.unit, m.unit, COUNT(*) OVER()::int
		FROM stock_in_item sii
		INNER JOIN material m ON m.id = sii.material_id
		WHERE sii.stock_in_id = $1 AND sii.material_id = $2
		LIMIT 1
	`, stockInID, materialID).Scan(&itemUnit, &materialUnit, &itemCount); err != nil {
		t.Fatalf("db stock-in item unit query failed: %v", err)
	}
	if itemUnit == "" || itemUnit != materialUnit {
		t.Fatalf("db stock-in item unit mismatch: item=%q material=%q", itemUnit, materialUnit)
	}
	if itemCount != 1 {
		t.Fatalf("db stock-in item count mismatch: got=%d want=1", itemCount)
	}

	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int,
		       COUNT(*) FILTER (WHERE status = 'in_stock')::int
		FROM material_serial_code
		WHERE stock_in_id = $1 AND material_id = $2
	`, stockInID, materialID).Scan(&serialCount, &inStockCount); err != nil {
		t.Fatalf("db stock-in serial query failed: %v", err)
	}
	if serialCount != wantSerials || inStockCount != wantSerials {
		t.Fatalf("db stock-in serial count mismatch: total=%d in_stock=%d want=%d", serialCount, inStockCount, wantSerials)
	}
}

func mustAssertDBStockOutReservation(ctx context.Context, t *testing.T, pool *pgxpool.Pool, stockOutItemID int64, serialIDs []int64) {
	t.Helper()
	if pool == nil {
		return
	}

	var selectionCount, inStockCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM stock_out_item_serial_selection
		WHERE stock_out_item_id = $1 AND serial_code_id = ANY($2)
	`, stockOutItemID, serialIDs).Scan(&selectionCount); err != nil {
		t.Fatalf("db stock-out reservation query failed: %v", err)
	}
	if selectionCount != len(serialIDs) {
		t.Fatalf("db stock-out reservation count mismatch: got=%d want=%d", selectionCount, len(serialIDs))
	}

	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM material_serial_code
		WHERE id = ANY($1) AND status = 'in_stock'
	`, serialIDs).Scan(&inStockCount); err != nil {
		t.Fatalf("db stock-out reserved serial status query failed: %v", err)
	}
	if inStockCount != len(serialIDs) {
		t.Fatalf("db reserved serials should remain in_stock before confirm: got=%d want=%d", inStockCount, len(serialIDs))
	}
}

func mustAssertDBStockOutConfirmed(ctx context.Context, t *testing.T, pool *pgxpool.Pool, stockOutID, stockOutItemID int64, serialIDs []int64) {
	t.Helper()
	if pool == nil {
		return
	}

	var selectionCount, issuedCount, traceCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM stock_out_item_serial_selection
		WHERE stock_out_item_id = $1
	`, stockOutItemID).Scan(&selectionCount); err != nil {
		t.Fatalf("db stock-out selection cleanup query failed: %v", err)
	}
	if selectionCount != 0 {
		t.Fatalf("db stock-out selections should be cleared after confirm: got=%d", selectionCount)
	}

	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM material_serial_code
		WHERE id = ANY($1) AND status = 'issued'
	`, serialIDs).Scan(&issuedCount); err != nil {
		t.Fatalf("db stock-out issued serial query failed: %v", err)
	}
	if issuedCount != len(serialIDs) {
		t.Fatalf("db stock-out issued serial count mismatch: got=%d want=%d", issuedCount, len(serialIDs))
	}

	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM material_serial_trace
		WHERE serial_code_id = ANY($1)
		  AND action = 'stock_out'
		  AND ref_doc_type = 'stock_out'
		  AND ref_doc_id = $2
	`, serialIDs, stockOutID).Scan(&traceCount); err != nil {
		t.Fatalf("db stock-out trace query failed: %v", err)
	}
	if traceCount != len(serialIDs) {
		t.Fatalf("db stock-out trace count mismatch: got=%d want=%d", traceCount, len(serialIDs))
	}
}

func mustAssertDBStockInReservation(ctx context.Context, t *testing.T, pool *pgxpool.Pool, stockInItemID int64, serialIDs []int64) {
	t.Helper()
	if pool == nil {
		return
	}

	var selectionCount, issuedCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM stock_in_item_serial_selection
		WHERE stock_in_item_id = $1 AND serial_code_id = ANY($2)
	`, stockInItemID, serialIDs).Scan(&selectionCount); err != nil {
		t.Fatalf("db stock-in reservation query failed: %v", err)
	}
	if selectionCount != len(serialIDs) {
		t.Fatalf("db stock-in reservation count mismatch: got=%d want=%d", selectionCount, len(serialIDs))
	}

	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM material_serial_code
		WHERE id = ANY($1) AND status = 'issued'
	`, serialIDs).Scan(&issuedCount); err != nil {
		t.Fatalf("db stock-in reserved serial status query failed: %v", err)
	}
	if issuedCount != len(serialIDs) {
		t.Fatalf("db stock-in reserved serials should remain issued before confirm: got=%d want=%d", issuedCount, len(serialIDs))
	}
}

func mustAssertDBSingleStockOutReservationForSerial(ctx context.Context, t *testing.T, pool *pgxpool.Pool, serialID int64) {
	t.Helper()
	if pool == nil {
		return
	}

	var selectionCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM stock_out_item_serial_selection
		WHERE serial_code_id = $1
	`, serialID).Scan(&selectionCount); err != nil {
		t.Fatalf("db concurrent reservation query failed: %v", err)
	}
	if selectionCount != 1 {
		t.Fatalf("db same serial reservation uniqueness mismatch: got=%d want=1 serialID=%d", selectionCount, serialID)
	}
}

func mustAssertStockOutPending(ctx context.Context, t *testing.T, c *testutil.Client, stockOutID int64) {
	t.Helper()
	var so StockOutDetail
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-out/%d", stockOutID), nil, nil, &so); err != nil {
		t.Fatalf("get stock-out after failed confirm failed: %v", err)
	}
	if so.Status != "pending" {
		t.Fatalf("stock-out status after failed confirm mismatch: got=%s want=pending", so.Status)
	}
}

func mustAssertStockInStatus(ctx context.Context, t *testing.T, c *testutil.Client, stockInID int64, want string) {
	t.Helper()
	stockIn := mustGetStockInDetail(ctx, t, c, stockInID)
	if stockIn.StockInStatus != want {
		t.Fatalf("stock-in status mismatch: got=%s want=%s", stockIn.StockInStatus, want)
	}
}

func mustAssertDBNoStockOutSideEffects(ctx context.Context, t *testing.T, pool *pgxpool.Pool, stockOutID, stockOutItemID, materialID int64) {
	t.Helper()
	if pool == nil {
		return
	}

	var selectionCount, traceCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM stock_out_item_serial_selection
		WHERE stock_out_item_id = $1
	`, stockOutItemID).Scan(&selectionCount); err != nil {
		t.Fatalf("db failed stock-out selection query failed: %v", err)
	}
	if selectionCount != 0 {
		t.Fatalf("db failed stock-out should not create selections: got=%d", selectionCount)
	}
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM material_serial_trace
		WHERE action = 'stock_out'
		  AND ref_doc_type = 'stock_out'
		  AND ref_doc_id = $1
		  AND serial_code_id IN (
		      SELECT id FROM material_serial_code WHERE material_id = $2
		  )
	`, stockOutID, materialID).Scan(&traceCount); err != nil {
		t.Fatalf("db failed stock-out trace query failed: %v", err)
	}
	if traceCount != 0 {
		t.Fatalf("db failed stock-out should not write trace: got=%d", traceCount)
	}
}

func mustAssertDBNoStockInSideEffects(ctx context.Context, t *testing.T, pool *pgxpool.Pool, stockInID, stockInItemID int64, serialIDs []int64) {
	t.Helper()
	if pool == nil {
		return
	}

	var selectionCount, inStockCount, traceCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM stock_in_item_serial_selection
		WHERE stock_in_item_id = $1
	`, stockInItemID).Scan(&selectionCount); err != nil {
		t.Fatalf("db failed stock-in selection query failed: %v", err)
	}
	if selectionCount != 0 {
		t.Fatalf("db failed stock-in should not create selections: got=%d", selectionCount)
	}
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM material_serial_code
		WHERE id = ANY($1) AND status = 'in_stock' AND stock_in_id = $2
	`, serialIDs, stockInID).Scan(&inStockCount); err != nil {
		t.Fatalf("db failed stock-in serial status query failed: %v", err)
	}
	if inStockCount != 0 {
		t.Fatalf("db failed stock-in should not return serials to stock: got=%d", inStockCount)
	}
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM material_serial_trace
		WHERE serial_code_id = ANY($1)
		  AND action = 'stock_in'
		  AND ref_doc_type = 'stock_in'
		  AND ref_doc_id = $2
	`, serialIDs, stockInID).Scan(&traceCount); err != nil {
		t.Fatalf("db failed stock-in trace query failed: %v", err)
	}
	if traceCount != 0 {
		t.Fatalf("db failed stock-in should not write trace: got=%d", traceCount)
	}
}
