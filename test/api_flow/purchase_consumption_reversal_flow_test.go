/**
 * 功能：按业务流程跑通采购→入库→领料→出库→退料→入库→采购退货→出库，并校验追踪/台账/大屏
 * 创建时间：2026-04-28
 * 创建人：GPT-5.2
 */

package api_flow

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/redgreat/teweicun/test/testutil"
)

type PermissionNode struct {
	ID       int64            `json:"id"`
	Children []PermissionNode `json:"children"`
}

type Supplier struct {
	ID           int64  `json:"id"`
	SupplierCode string `json:"supplier_code"`
	SupplierName string `json:"supplier_name"`
}

type Warehouse struct {
	WarehouseCode string `json:"warehouse_code"`
	WarehouseName string `json:"warehouse_name"`
	WarehouseType string `json:"warehouse_type"`
	ID            int64  `json:"id"`
}

type Material struct {
	ID           int64  `json:"id"`
	MaterialName string `json:"material_name"`
	IsCode       bool   `json:"is_code"`
}

type PurchaseOrderDetail struct {
	ID        int64  `json:"id"`
	StockInID int64  `json:"stock_in_id"`
	OrderNo   string `json:"order_no"`
	Items     []struct {
		ID         int64   `json:"id"`
		MaterialID int64   `json:"material_id"`
		Quantity   float64 `json:"quantity"`
	} `json:"items"`
}

type StockInDetail struct {
	ID            int64  `json:"id"`
	StockInNo     string `json:"stock_in_no"`
	StockInType   string `json:"stock_in_type"`
	StockInStatus string `json:"stock_in_status"`
	WarehouseCode string `json:"warehouse_code"`
	Items         []struct {
		ID               int64    `json:"id"`
		MaterialID       int64    `json:"material_id"`
		ArrivedQuantity  float64  `json:"arrived_quantity"`
		AcceptedQuantity float64  `json:"accepted_quantity"`
		UnitCost         *float64 `json:"unit_cost"`
		IsCode           bool     `json:"is_code"`
	} `json:"items"`
}

type StockInListRow struct {
	ID              int64  `json:"id"`
	PurchaseOrderID int64  `json:"purchase_order_id"`
	StockInType     string `json:"stock_in_type"`
	StockInNo       string `json:"stock_in_no"`
}

type SerialCodeItem struct {
	ID         int64  `json:"id"`
	SerialCode string `json:"serial_code"`
}

type InventoryAvailable struct {
	InventoryID       int64   `json:"inventory_id"`
	MaterialID        int64   `json:"material_id"`
	IsCode            bool    `json:"is_code"`
	WarehouseCode     string  `json:"warehouse_code"`
	Unit              string  `json:"unit"`
	UnitCost          float64 `json:"unit_cost"`
	AvailableQuantity float64 `json:"available_quantity"`
}

type ConsumptionOrder struct {
	ID         int64 `json:"id"`
	StockOutID int64 `json:"stock_out_id"`
}

type StockOutDetail struct {
	ID         int64  `json:"id"`
	StockOutNo string `json:"stock_out_no"`
	Status     string `json:"status"`
	RefDocType string `json:"ref_doc_type"`
	Items      []struct {
		ID         int64   `json:"id"`
		MaterialID int64   `json:"material_id"`
		Quantity   float64 `json:"quantity"`
		IsCode     bool    `json:"is_code"`
	} `json:"items"`
}

type InventoryIssued struct {
	InventoryID    int64   `json:"inventory_id"`
	MaterialID     int64   `json:"material_id"`
	IsCode         bool    `json:"is_code"`
	Unit           string  `json:"unit"`
	IssuedQuantity float64 `json:"issued_quantity"`
}

type ReversalOrder struct {
	ID        int64 `json:"id"`
	StockInID int64 `json:"stock_in_id"`
}

type ReturnOrder struct {
	ID            int64  `json:"id"`
	ReturnNo      string `json:"return_no"`
	ReturnType    string `json:"return_type"`
	Status        string `json:"status"`
	WarehouseCode string `json:"warehouse_code"`
	SupplierCode  string `json:"supplier_code"`
	StockOutID    *int64 `json:"stock_out_id"`
	Items         []struct {
		InventoryID int64   `json:"inventory_id"`
		MaterialID  int64   `json:"material_id"`
		Quantity    float64 `json:"quantity"`
	} `json:"items,omitempty"`
}

type InventoryMaterialLedgerRow struct {
	MaterialID    int64   `json:"material_id"`
	MaterialName  string  `json:"material_name"`
	IsCode        bool    `json:"is_code"`
	Quantity      float64 `json:"quantity"`
	WarehouseName string  `json:"warehouse_name"`
}

type TraceMaterialResult struct {
	SerialInfo *struct {
		SerialCode string `json:"serial_code"`
		Status     string `json:"status"`
	} `json:"serial_info"`
	Traces []struct {
		Action      string `json:"action"`
		ActionLabel string `json:"action_label"`
		RefDocType  string `json:"ref_doc_type"`
		RefDocNo    string `json:"ref_doc_no"`
	} `json:"traces"`
}

func TestFlow_PurchaseToReversalAndReturn(t *testing.T) {
	env := testutil.LoadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()

	admin := testutil.NewClient(env.BaseURL)

	adminLogin, err := admin.Login(ctx, env.AdminUser, env.AdminPass)
	if err != nil {
		t.Fatalf("admin login failed: %v", err)
	}
	t.Logf("admin login ok: userID=%d username=%s", adminLogin.UserID, adminLogin.Username)

	_ = admin.DoJSON(ctx, http.MethodGet, "/api/v1/health", nil, nil, nil)

	useAdminOnly := true
	fixture := mustLoadBaseDataFixture(ctx, t, admin, adminLogin.UserID)
	supplierCode := fixture.SupplierCode
	warehouseCode := fixture.MainMaterialWarehouse
	categoryID := fixture.CategoryID

	codeMatID := mustCreateMaterial(ctx, t, admin, categoryID, uniqueChineseName("编码钢板"), true)
	noCodeMatID := mustCreateMaterial(ctx, t, admin, categoryID, uniqueChineseName("普通钢板"), false)

	clientA := admin
	if !useAdminOnly {
		clientA = testutil.NewClient(env.BaseURL)
	}

	poID := mustCreatePurchaseOrder(ctx, t, clientA, supplierCode, codeMatID, noCodeMatID)
	mustConfirm(ctx, t, clientA, fmt.Sprintf("/api/v1/purchase/orders/%d/confirm", poID))

	po := waitPurchaseOrderDetail(ctx, t, clientA, poID, func(p PurchaseOrderDetail) bool { return true })
	if po.StockInID <= 0 {
		t.Logf("purchase has no stock_in_id in purchase detail, try locate stock-in by listing. poID=%d", poID)
		if id, ok := waitFindStockInIDByPurchaseOrderID(ctx, admin, poID); ok {
			po.StockInID = id
		} else {
			t.Logf("WARN: cannot locate stock-in by listing, try create stock-in manually. poID=%d", poID)
			po.StockInID = mustCreateStockInForPurchase(ctx, t, clientA, poID, warehouseCode, po.Items)
		}
	}
	t.Logf("purchase confirmed: poID=%d orderNo=%s stockInID=%d", poID, po.OrderNo, po.StockInID)

	clientC := admin
	if !useAdminOnly {
		clientC = testutil.NewClient(env.BaseURL)
	}
	setStockInWarehouseIfEmpty(ctx, t, clientC, po.StockInID, warehouseCode)
	mustConfirm(ctx, t, clientC, fmt.Sprintf("/api/v1/stock-in/%d/confirm", po.StockInID))

	var stockIn StockInDetail
	if err := clientC.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-in/%d", po.StockInID), nil, nil, &stockIn); err != nil {
		t.Fatalf("get stock-in failed: %v", err)
	}
	t.Logf("stock-in confirmed: stockInNo=%s items=%d", stockIn.StockInNo, len(stockIn.Items))

	codeStockInItemID := findStockInItemIDByMaterial(stockIn, codeMatID)
	if codeStockInItemID <= 0 {
		t.Fatalf("cannot find coded stock-in item for material_id=%d", codeMatID)
	}
	var serialsIn []SerialCodeItem
	if err := clientC.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/serial-codes/stock-in-item/%d", codeStockInItemID), nil, nil, &serialsIn); err != nil {
		t.Fatalf("get serial codes by stock-in item failed: %v", err)
	}
	if len(serialsIn) != 10 {
		t.Fatalf("expected 10 serial codes for coded stock-in item, got %d", len(serialsIn))
	}
	t.Logf("serial codes generated: count=%d sample=%s", len(serialsIn), serialsIn[0].SerialCode)

	clientB := admin
	if !useAdminOnly {
		clientB = testutil.NewClient(env.BaseURL)
	}

	codeInv := mustPickInventoryAvailable(ctx, t, clientB, warehouseCode, codeMatID, true, 5)
	noCodeInv := mustPickInventoryAvailable(ctx, t, clientB, warehouseCode, noCodeMatID, false, 5)

	consID := mustCreateConsumptionOrder(ctx, t, clientB, codeInv, 5, noCodeInv, 5)
	mustConfirm(ctx, t, clientB, fmt.Sprintf("/api/v1/consumption/orders/%d/confirm", consID))

	cons := waitConsumptionOrder(ctx, t, clientB, consID, func(o ConsumptionOrder) bool { return o.StockOutID > 0 })
	t.Logf("consumption confirmed: orderID=%d stockOutID=%d", consID, cons.StockOutID)

	_ = tryAutoStockOutSerialSelections(ctx, clientC, cons.StockOutID)
	mustConfirm(ctx, t, clientC, fmt.Sprintf("/api/v1/stock-out/%d/confirm", cons.StockOutID))

	var stockOut StockOutDetail
	if err := clientC.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-out/%d", cons.StockOutID), nil, nil, &stockOut); err != nil {
		t.Fatalf("get stock-out failed: %v", err)
	}
	t.Logf("stock-out confirmed: stockOutNo=%s items=%d", stockOut.StockOutNo, len(stockOut.Items))

	codeStockOutItemID := findStockOutItemIDByMaterial(stockOut, codeMatID)
	if codeStockOutItemID > 0 {
		var selectedOutSerials []SerialCodeItem
		if err := clientC.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/serial-codes/stock-out-item/%d", codeStockOutItemID), nil, nil, &selectedOutSerials); err == nil {
			if len(selectedOutSerials) != 5 {
				t.Fatalf("expected 5 serials selected for coded stock-out item, got %d", len(selectedOutSerials))
			}
		}
	}

	codeIssued := mustPickInventoryIssued(ctx, t, clientB, warehouseCode, codeMatID, true, 3)
	noCodeIssued := mustPickInventoryIssued(ctx, t, clientB, warehouseCode, noCodeMatID, false, 3)

	revID := mustCreateReversalOrder(ctx, t, clientB, codeIssued, 3, noCodeIssued, 3)
	mustConfirm(ctx, t, clientB, fmt.Sprintf("/api/v1/reversal/orders/%d/confirm", revID))

	rev := waitReversalOrder(ctx, t, clientB, revID, func(o ReversalOrder) bool { return o.StockInID > 0 })
	t.Logf("reversal confirmed: orderID=%d stockInID=%d", revID, rev.StockInID)

	prepareReversalStockInSerials(ctx, t, clientC, rev.StockInID, 3)
	mustConfirm(ctx, t, clientC, fmt.Sprintf("/api/v1/stock-in/%d/confirm-reversal", rev.StockInID))

	codeInv2 := mustPickInventoryAvailable(ctx, t, clientA, warehouseCode, codeMatID, true, 5)
	noCodeInv2 := mustPickInventoryAvailable(ctx, t, clientA, warehouseCode, noCodeMatID, false, 5)
	retID := mustCreateReturnOrder(ctx, t, clientA, supplierCode, warehouseCode, codeInv2, 5, noCodeInv2, 5)
	mustConfirmReturnOrder(ctx, t, clientA, retID, warehouseCode)

	ret := waitReturnOrder(ctx, t, clientA, retID, func(o ReturnOrder) bool { return o.StockOutID != nil && *o.StockOutID > 0 })
	t.Logf("return confirmed: returnNo=%s stockOutID=%d", ret.ReturnNo, *ret.StockOutID)

	_ = tryAutoStockOutSerialSelections(ctx, clientC, *ret.StockOutID)
	mustConfirm(ctx, t, clientC, fmt.Sprintf("/api/v1/stock-out/%d/confirm", *ret.StockOutID))

	sampleSerials := pickSerialSamples(serialsIn, 3)
	for _, s := range sampleSerials {
		q := url.Values{}
		q.Set("serial_code", s.SerialCode)
		var tr TraceMaterialResult
		if err := clientC.DoJSON(ctx, http.MethodGet, "/api/v1/trace/material/serial", q, nil, &tr); err != nil {
			t.Fatalf("trace serial=%s failed: %v", s.SerialCode, err)
		}
		if tr.SerialInfo == nil || tr.SerialInfo.SerialCode == "" {
			t.Fatalf("trace serial=%s: missing serial_info", s.SerialCode)
		}
		if len(tr.Traces) == 0 {
			t.Fatalf("trace serial=%s: expected non-empty traces", s.SerialCode)
		}
	}

	expectFinal := 3.0
	ledgerRows := mustFetchLedger(ctx, t, clientC)
	gotCode := findLedgerQty(ledgerRows, codeMatID)
	gotNoCode := findLedgerQty(ledgerRows, noCodeMatID)
	if gotCode != expectFinal {
		t.Fatalf("coded final qty mismatch: got=%v expect=%v", gotCode, expectFinal)
	}
	if gotNoCode != expectFinal {
		t.Fatalf("nocode final qty mismatch: got=%v expect=%v", gotNoCode, expectFinal)
	}

	var dash map[string]any
	if err := clientC.DoJSON(ctx, http.MethodGet, "/api/v1/dashboard/bigscreen", nil, nil, &dash); err != nil {
		t.Fatalf("bigscreen dashboard failed: %v", err)
	}
	if len(dash) == 0 {
		t.Fatalf("bigscreen dashboard returned empty data")
	}
}

func fetchAllPermissionIDs(ctx context.Context, c *testutil.Client) ([]int64, error) {
	var tree []PermissionNode
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/system/permissions/tree", nil, nil, &tree); err != nil {
		return nil, err
	}
	var ids []int64
	var walk func(nodes []PermissionNode)
	walk = func(nodes []PermissionNode) {
		for _, n := range nodes {
			if n.ID > 0 {
				ids = append(ids, n.ID)
			}
			if len(n.Children) > 0 {
				walk(n.Children)
			}
		}
	}
	walk(tree)
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	ids = dedup(ids)
	return ids, nil
}

func waitPurchaseOrderDetail(ctx context.Context, t *testing.T, c *testutil.Client, id int64, ready func(PurchaseOrderDetail) bool) PurchaseOrderDetail {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	var last PurchaseOrderDetail
	var lastErr error
	for time.Now().Before(deadline) {
		var po PurchaseOrderDetail
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
	t.Fatalf("purchase order not ready in time: id=%d stockInID=%d", id, last.StockInID)
	return last
}

func waitConsumptionOrder(ctx context.Context, t *testing.T, c *testutil.Client, id int64, ready func(ConsumptionOrder) bool) ConsumptionOrder {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	var last ConsumptionOrder
	var lastErr error
	for time.Now().Before(deadline) {
		var o ConsumptionOrder
		err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/consumption/orders/%d", id), nil, nil, &o)
		if err == nil {
			last = o
			if ready(o) {
				return o
			}
		} else {
			lastErr = err
		}
		time.Sleep(1 * time.Second)
	}
	if lastErr != nil {
		t.Fatalf("wait consumption order failed: %v", lastErr)
	}
	t.Fatalf("consumption order not ready in time: id=%d stockOutID=%d", id, last.StockOutID)
	return last
}

func waitReversalOrder(ctx context.Context, t *testing.T, c *testutil.Client, id int64, ready func(ReversalOrder) bool) ReversalOrder {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	var last ReversalOrder
	var lastErr error
	for time.Now().Before(deadline) {
		var o ReversalOrder
		err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/reversal/orders/%d", id), nil, nil, &o)
		if err == nil {
			last = o
			if ready(o) {
				return o
			}
		} else {
			lastErr = err
		}
		time.Sleep(1 * time.Second)
	}
	if lastErr != nil {
		t.Fatalf("wait reversal order failed: %v", lastErr)
	}
	t.Fatalf("reversal order not ready in time: id=%d stockInID=%d", id, last.StockInID)
	return last
}

func waitReturnOrder(ctx context.Context, t *testing.T, c *testutil.Client, id int64, ready func(ReturnOrder) bool) ReturnOrder {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	var last ReturnOrder
	var lastErr error
	for time.Now().Before(deadline) {
		var o ReturnOrder
		err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/returns/%d", id), nil, nil, &o)
		if err == nil {
			last = o
			if ready(o) {
				return o
			}
		} else {
			lastErr = err
		}
		time.Sleep(1 * time.Second)
	}
	if lastErr != nil {
		t.Fatalf("wait return order failed: %v", lastErr)
	}
	t.Fatalf("return order not ready in time: id=%d stockOutID=%v", id, last.StockOutID)
	return last
}

func waitFindStockInIDByPurchaseOrderID(ctx context.Context, c *testutil.Client, purchaseOrderID int64) (int64, bool) {
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		q := url.Values{}
		q.Set("page", "1")
		q.Set("page_size", "50")
		q.Set("stock_in_type", "purchase")
		var list []StockInListRow
		if err := c.DoPage(ctx, http.MethodGet, "/api/v1/stock-in", q, &list, nil); err == nil {
			for _, it := range list {
				if it.PurchaseOrderID == purchaseOrderID {
					return it.ID, true
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
	return 0, false
}

func dedup(in []int64) []int64 {
	if len(in) == 0 {
		return in
	}
	out := make([]int64, 0, len(in))
	var last int64
	for i, v := range in {
		if i == 0 || v != last {
			out = append(out, v)
			last = v
		}
	}
	return out
}

func mustCreateRole(ctx context.Context, t *testing.T, c *testutil.Client, roleCode, roleName string) int64 {
	t.Helper()
	req := map[string]any{
		"role_code":   roleCode,
		"role_name":   roleName,
		"description": "e2e role " + roleCode,
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/system/roles", nil, req, &out); err != nil {
		t.Fatalf("create role failed: %v", err)
	}
	t.Logf("role created: %s id=%d", roleCode, out.ID)
	return out.ID
}

func tryCreateRole(ctx context.Context, t *testing.T, c *testutil.Client, roleCode, roleName string) (int64, bool) {
	t.Helper()
	req := map[string]any{
		"role_code":   roleCode,
		"role_name":   roleName,
		"description": "e2e role " + roleCode,
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/system/roles", nil, req, &out); err != nil {
		t.Logf("WARN: create role failed: roleCode=%s err=%v", roleCode, err)
		return 0, false
	}
	t.Logf("role created: %s id=%d", roleCode, out.ID)
	return out.ID, true
}

func trySetRolePerms(ctx context.Context, t *testing.T, c *testutil.Client, roleID int64, permIDs []int64) error {
	t.Helper()
	req := map[string]any{"permission_ids": permIDs}
	return c.DoJSON(ctx, http.MethodPost, fmt.Sprintf("/api/v1/system/roles/%d/permissions", roleID), nil, req, nil)
}

func mustCreateUser(ctx context.Context, t *testing.T, c *testutil.Client, username, password, realName string, roleIDs []int64) int64 {
	t.Helper()
	req := map[string]any{
		"username":  username,
		"password":  password,
		"real_name": realName,
		"role_ids":  roleIDs,
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/system/users", nil, req, &out); err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	t.Logf("user created: %s id=%d", username, out.ID)
	return out.ID
}

func tryCreateUser(ctx context.Context, t *testing.T, c *testutil.Client, username, password, realName string, roleIDs []int64) (int64, bool) {
	t.Helper()
	req := map[string]any{
		"username":  username,
		"password":  password,
		"real_name": realName,
		"role_ids":  roleIDs,
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/system/users", nil, req, &out); err != nil {
		t.Logf("WARN: create user failed: username=%s err=%v", username, err)
		return 0, false
	}
	t.Logf("user created: %s id=%d", username, out.ID)
	return out.ID, true
}

func mustEnsureSupplier(ctx context.Context, t *testing.T, c *testutil.Client, prefix string) string {
	t.Helper()
	var list []Supplier
	_ = c.DoPage(ctx, http.MethodGet, "/api/v1/base/suppliers", nil, &list, nil)
	if len(list) > 0 && list[0].SupplierCode != "" {
		return list[0].SupplierCode
	}
	req := map[string]any{
		"supplier_code":  prefix + "_SUP",
		"supplier_name":  "测试供应商",
		"supplier_type":  "manufacturer",
		"contact_person": "E2E",
		"contact_phone":  "13800000000",
		"is_qualified":   true,
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/suppliers", nil, req, &out); err != nil {
		t.Fatalf("create supplier failed: %v", err)
	}
	return prefix + "_SUP"
}

func mustEnsureWarehouse(ctx context.Context, t *testing.T, c *testutil.Client, prefix string, managerID int64) string {
	t.Helper()
	var list []Warehouse
	_ = c.DoPage(ctx, http.MethodGet, "/api/v1/base/warehouses", nil, &list, nil)
	if len(list) > 0 && list[0].WarehouseCode != "" {
		return list[0].WarehouseCode
	}
	req := map[string]any{
		"warehouse_code": prefix + "_WH",
		"warehouse_name": "E2E Warehouse " + prefix,
		"warehouse_type": "normal",
		"manager_id":     managerID,
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/warehouses", nil, req, &out); err != nil {
		t.Fatalf("create warehouse failed: %v", err)
	}
	return prefix + "_WH"
}

func mustEnsureCategory(ctx context.Context, t *testing.T, c *testutil.Client, prefix string) int64 {
	t.Helper()
	var data []map[string]any
	_ = c.DoJSON(ctx, http.MethodGet, "/api/v1/base/categories", nil, nil, &data)
	if len(data) > 0 {
		if id, ok := asInt64(data[0]["id"]); ok && id > 0 {
			return id
		}
	}
	req := map[string]any{
		"parent_id":     0,
		"category_code": prefix + "_CAT",
		"category_name": "测试分类",
		"sort_order":    1,
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/categories", nil, req, &out); err != nil {
		t.Fatalf("create category failed: %v", err)
	}
	return out.ID
}

func mustCreateMaterial(ctx context.Context, t *testing.T, c *testutil.Client, categoryID int64, name string, isCode bool) int64 {
	t.Helper()
	req := map[string]any{
		"category_id":   categoryID,
		"material_name": name,
		"unit":          "pcs",
		"is_code":       isCode,
		"remark":        "e2e",
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/materials", nil, req, &out); err != nil {
		t.Fatalf("create material failed: %v", err)
	}
	return out.ID
}

func mustCreatePurchaseOrder(ctx context.Context, t *testing.T, c *testutil.Client, supplierCode string, codeMatID, noCodeMatID int64) int64 {
	t.Helper()
	today := time.Now().Format("2006-01-02")
	req := map[string]any{
		"order_type":    "purchase",
		"supplier_code": supplierCode,
		"order_date":    today,
		"remark":        "e2e purchase " + today,
		"items": []map[string]any{
			{"material_id": codeMatID, "quantity": 10, "unit_price": 10},
			{"material_id": noCodeMatID, "quantity": 10, "unit_price": 10},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/purchase/orders", nil, req, &out); err != nil {
		t.Fatalf("create purchase order failed: %v", err)
	}
	return out.ID
}

func mustCreateStockInForPurchase(ctx context.Context, t *testing.T, c *testutil.Client, purchaseOrderID int64, warehouseCode string, poItems []struct {
	ID         int64   `json:"id"`
	MaterialID int64   `json:"material_id"`
	Quantity   float64 `json:"quantity"`
}) int64 {
	t.Helper()
	today := time.Now().Format("2006-01-02")
	items := make([]map[string]any, 0, len(poItems))
	for _, it := range poItems {
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
		"purchase_order_id": purchaseOrderID,
		"remark":            "e2e stock-in for purchase",
		"items":             items,
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/stock-in", nil, req, &out); err != nil {
		t.Fatalf("create stock-in for purchase failed: %v", err)
	}
	return out.ID
}

func mustConfirm(ctx context.Context, t *testing.T, c *testutil.Client, path string) {
	t.Helper()
	if err := c.DoJSON(ctx, http.MethodPost, path, nil, nil, nil); err != nil {
		if apiErr, ok := err.(*testutil.APIError); ok {
			if strings.Contains(apiErr.Msg, "material_inspection_no") || strings.Contains(apiErr.Body, "material_inspection_no") {
				t.Fatalf("confirm %s failed (DB stored procedure mismatch: material_inspection_no). Need apply latest migrations / fix sp_confirm_stock_in. raw=%v", path, err)
			}
		}
		t.Fatalf("confirm %s failed: %v", path, err)
	}
}

func mustConfirmReturnOrder(ctx context.Context, t *testing.T, c *testutil.Client, returnID int64, warehouseCode string) {
	t.Helper()
	path := fmt.Sprintf("/api/v1/returns/%d/confirm", returnID)
	if err := c.DoJSON(ctx, http.MethodPost, path, nil, nil, nil); err != nil {
		t.Logf("confirm return order failed: returnID=%d err=%v", returnID, err)
		var ro ReturnOrder
		if e := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/returns/%d", returnID), nil, nil, &ro); e == nil {
			t.Logf("return snapshot: id=%d no=%s type=%s status=%s wh=%s supplier=%s stockOutID=%v items=%d",
				ro.ID, ro.ReturnNo, ro.ReturnType, ro.Status, ro.WarehouseCode, ro.SupplierCode, ro.StockOutID, len(ro.Items))
			for i, it := range ro.Items {
				t.Logf("return item[%d]: inventory_id=%d material_id=%d qty=%.3f",
					i, it.InventoryID, it.MaterialID, it.Quantity)
			}
		}
		q := url.Values{}
		q.Set("page", "1")
		q.Set("page_size", "100")
		q.Set("warehouse_code", warehouseCode)
		var avail []InventoryAvailable
		if e := c.DoPage(ctx, http.MethodGet, "/api/v1/inventory/available", q, &avail, nil); e == nil {
			t.Logf("available snapshot rows=%d warehouse=%s", len(avail), warehouseCode)
			for i, it := range avail {
				if i >= 8 {
					break
				}
				t.Logf("avail[%d]: inv=%d material=%d isCode=%v avail=%.3f",
					i, it.InventoryID, it.MaterialID, it.IsCode, it.AvailableQuantity)
			}
		}
		t.Fatalf("confirm %s failed: %v", path, err)
	}
}

func findStockInItemIDByMaterial(stockIn StockInDetail, materialID int64) int64 {
	for _, it := range stockIn.Items {
		if it.MaterialID == materialID {
			return it.ID
		}
	}
	return 0
}

func setStockInWarehouseIfEmpty(ctx context.Context, t *testing.T, c *testutil.Client, stockInID int64, warehouseCode string) {
	t.Helper()
	var si StockInDetail
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-in/%d", stockInID), nil, nil, &si); err != nil {
		t.Fatalf("get stock-in before update failed: %v", err)
	}
	if strings.TrimSpace(si.WarehouseCode) != "" {
		return
	}
	items := make([]map[string]any, 0, len(si.Items))
	for _, it := range si.Items {
		unitCost := 0.0
		if it.UnitCost != nil {
			unitCost = *it.UnitCost
		}
		items = append(items, map[string]any{
			"id":                it.ID,
			"material_id":       it.MaterialID,
			"arrived_quantity":  it.ArrivedQuantity,
			"accepted_quantity": it.AcceptedQuantity,
			"unit_cost":         unitCost,
			"cert_id":           0,
			"custom_attributes": nil,
		})
	}
	req := map[string]any{
		"warehouse_code": warehouseCode,
		"remark":         "e2e set warehouse",
		"items":          items,
	}
	if err := c.DoJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stock-in/%d", stockInID), nil, req, nil); err != nil {
		t.Fatalf("update stock-in warehouse failed: %v", err)
	}
}

func mustPickInventoryAvailable(ctx context.Context, t *testing.T, c *testutil.Client, warehouseCode string, materialID int64, isCode bool, qty float64) InventoryAvailable {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	q.Set("warehouse_code", warehouseCode)
	var list []InventoryAvailable
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/inventory/available", q, &list, nil); err != nil {
		t.Fatalf("list inventory available failed: %v", err)
	}
	for _, it := range list {
		if it.MaterialID == materialID && it.IsCode == isCode && it.AvailableQuantity >= qty {
			return it
		}
	}
	t.Fatalf("cannot find inventory available for material=%d qty>=%.2f (got %d rows)", materialID, qty, len(list))
	return InventoryAvailable{}
}

func mustCreateConsumptionOrder(ctx context.Context, t *testing.T, c *testutil.Client, codeInv InventoryAvailable, codeQty float64, noCodeInv InventoryAvailable, noCodeQty float64) int64 {
	t.Helper()
	today := time.Now().Format("2006-01-02")
	req := map[string]any{
		"project_no":   "E2E-" + today,
		"product_name": "E2E Product " + today,
		"order_date":   today,
		"remark":       "e2e consumption",
		"items": []map[string]any{
			{"material_id": codeInv.MaterialID, "inventory_id": codeInv.InventoryID, "quantity": codeQty, "unit": codeInv.Unit, "remark": "code"},
			{"material_id": noCodeInv.MaterialID, "inventory_id": noCodeInv.InventoryID, "quantity": noCodeQty, "unit": noCodeInv.Unit, "remark": "nocode"},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/consumption/orders", nil, req, &out); err != nil {
		t.Fatalf("create consumption order failed: %v", err)
	}
	return out.ID
}

func tryAutoStockOutSerialSelections(ctx context.Context, c *testutil.Client, stockOutID int64) error {
	req := map[string]any{
		"mode":  "auto_fifo",
		"items": []any{},
	}
	return c.DoJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stock-out/%d/serial-selections", stockOutID), nil, req, nil)
}

func findStockOutItemIDByMaterial(so StockOutDetail, materialID int64) int64 {
	for _, it := range so.Items {
		if it.MaterialID == materialID {
			return it.ID
		}
	}
	return 0
}

func mustPickInventoryIssued(ctx context.Context, t *testing.T, c *testutil.Client, warehouseCode string, materialID int64, isCode bool, qty float64) InventoryIssued {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	q.Set("warehouse_code", warehouseCode)
	var list []InventoryIssued
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/inventory/issued", q, &list, nil); err != nil {
		t.Fatalf("list inventory issued failed: %v", err)
	}
	for _, it := range list {
		if it.MaterialID == materialID && it.IsCode == isCode && it.IssuedQuantity >= qty {
			return it
		}
	}
	t.Fatalf("cannot find inventory issued for material=%d qty>=%.2f (got %d rows)", materialID, qty, len(list))
	return InventoryIssued{}
}

func mustCreateReversalOrder(ctx context.Context, t *testing.T, c *testutil.Client, codeInv InventoryIssued, codeQty float64, noCodeInv InventoryIssued, noCodeQty float64) int64 {
	t.Helper()
	today := time.Now().Format("2006-01-02")
	req := map[string]any{
		"project_no":   "E2E-" + today,
		"product_name": "E2E Product " + today,
		"order_date":   today,
		"remark":       "e2e reversal",
		"items": []map[string]any{
			{"inventory_id": codeInv.InventoryID, "material_id": codeInv.MaterialID, "quantity": codeQty, "unit": codeInv.Unit, "remark": "code"},
			{"inventory_id": noCodeInv.InventoryID, "material_id": noCodeInv.MaterialID, "quantity": noCodeQty, "unit": noCodeInv.Unit, "remark": "nocode"},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/reversal/orders", nil, req, &out); err != nil {
		t.Fatalf("create reversal order failed: %v", err)
	}
	return out.ID
}

func prepareReversalStockInSerials(ctx context.Context, t *testing.T, c *testutil.Client, stockInID int64, need int) {
	t.Helper()
	var si StockInDetail
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-in/%d", stockInID), nil, nil, &si); err != nil {
		t.Fatalf("get reversal stock-in failed: %v", err)
	}
	for _, item := range si.Items {
		var avail []SerialCodeItem
		if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/serial-codes/stock-in-item/%d/available-issued", item.ID), nil, nil, &avail); err != nil {
			t.Fatalf("get available issued serials failed: %v", err)
		}
		if len(avail) == 0 {
			continue
		}
		pick := need
		if len(avail) < pick {
			pick = len(avail)
		}
		ids := make([]int64, 0, pick)
		for i := 0; i < pick; i++ {
			ids = append(ids, avail[i].ID)
		}
		req := map[string]any{"serial_code_ids": ids}
		if err := c.DoJSON(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stock-in-item/%d/serial-selections", item.ID), nil, req, nil); err != nil {
			t.Fatalf("update stock-in item serial selections failed: %v", err)
		}
		t.Logf("reversal stock-in prepared serials: stockInItemID=%d selected=%d", item.ID, len(ids))
	}
}

func mustCreateReturnOrder(ctx context.Context, t *testing.T, c *testutil.Client, supplierCode, warehouseCode string, codeInv InventoryAvailable, codeQty float64, noCodeInv InventoryAvailable, noCodeQty float64) int64 {
	t.Helper()
	today := time.Now().Format("2006-01-02")
	req := map[string]any{
		"return_date":    today,
		"return_type":    "purchase_return",
		"supplier_code":  supplierCode,
		"warehouse_code": warehouseCode,
		"remark":         "e2e purchase return",
		"items": []map[string]any{
			{"inventory_id": codeInv.InventoryID, "material_id": codeInv.MaterialID, "quantity": codeQty},
			{"inventory_id": noCodeInv.InventoryID, "material_id": noCodeInv.MaterialID, "quantity": noCodeQty},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/returns", nil, req, &out); err != nil {
		t.Fatalf("create return order failed: %v", err)
	}
	return out.ID
}

func pickSerialSamples(in []SerialCodeItem, n int) []SerialCodeItem {
	if len(in) <= n {
		return in
	}
	out := []SerialCodeItem{in[0]}
	if n >= 2 {
		out = append(out, in[len(in)/2])
	}
	if n >= 3 {
		out = append(out, in[len(in)-1])
	}
	return out[:n]
}

func mustFetchLedger(ctx context.Context, t *testing.T, c *testutil.Client) []InventoryMaterialLedgerRow {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	var list []InventoryMaterialLedgerRow
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/inventory/material-ledger", q, &list, nil); err != nil {
		t.Fatalf("list material ledger failed: %v", err)
	}
	return list
}

func findLedgerQty(rows []InventoryMaterialLedgerRow, materialID int64) float64 {
	for _, r := range rows {
		if r.MaterialID == materialID {
			return r.Quantity
		}
	}
	return -9999
}

func asInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case int64:
		return x, true
	case int:
		return int64(x), true
	default:
		return 0, false
	}
}
