/**
 * 功能：E2E 流程测试共享辅助函数
 * 将原巨型测试中的辅助函数提取到此文件，供各流程测试复用
 * 创建时间：2026-04-28 / 重构时间：2026-07-12
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

// ============================================================================
// 类型定义（从原测试迁移）
// ============================================================================

type PermissionNode struct {
	ID       int64            `json:"id"`
	Children []PermissionNode `json:"children"`
}

type Supplier struct {
	SupplierCode string `json:"supplier_code"`
	SupplierName string `json:"supplier_name"`
}

type Warehouse struct {
	WarehouseCode string `json:"warehouse_code"`
	WarehouseName string `json:"warehouse_name"`
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

// ============================================================================
// 环境初始化
// ============================================================================

var (
	e2eEnv   testutil.Env
	e2eAdmin *testutil.Client
	e2eCtx   context.Context
)

func initE2E(t *testing.T) (context.Context, context.CancelFunc, *testutil.Client) {
	t.Helper()
	if e2eAdmin == nil {
		e2eEnv = testutil.LoadEnv()
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
		e2eCtx = ctx
		e2eAdmin = testutil.NewClient(e2eEnv.BaseURL)

		login, err := e2eAdmin.Login(ctx, e2eEnv.AdminUser, e2eEnv.AdminPass)
		if err != nil {
			t.Fatalf("admin login failed: %v", err)
		}
		t.Logf("admin login ok: userID=%d username=%s", login.UserID, login.Username)

		_ = e2eAdmin.DoJSON(ctx, http.MethodGet, "/api/v1/health", nil, nil, nil)
		return ctx, cancel, e2eAdmin
	}
	return e2eCtx, func() {}, e2eAdmin
}

// ============================================================================
// 权限/角色/用户
// ============================================================================

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
	return dedup(ids), nil
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
	return out.ID, true
}

func trySetRolePerms(ctx context.Context, t *testing.T, c *testutil.Client, roleID int64, permIDs []int64) error {
	t.Helper()
	req := map[string]any{"permission_ids": permIDs}
	return c.DoJSON(ctx, http.MethodPost, fmt.Sprintf("/api/v1/system/roles/%d/permissions", roleID), nil, req, nil)
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
	return out.ID, true
}

// ============================================================================
// 基础数据
// ============================================================================

func mustEnsureSupplier(ctx context.Context, t *testing.T, c *testutil.Client, prefix string) string {
	t.Helper()
	var list []Supplier
	_ = c.DoPage(ctx, http.MethodGet, "/api/v1/base/suppliers", nil, &list, nil)
	if len(list) > 0 && list[0].SupplierCode != "" {
		return list[0].SupplierCode
	}
	req := map[string]any{
		"supplier_code":  prefix + "_SUP",
		"supplier_name":  "E2E Supplier " + prefix,
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

// ============================================================================
// 采购订单
// ============================================================================

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

func mustConfirm(ctx context.Context, t *testing.T, c *testutil.Client, path string) {
	t.Helper()
	if err := c.DoJSON(ctx, http.MethodPost, path, nil, nil, nil); err != nil {
		t.Fatalf("confirm %s failed: %v", path, err)
	}
}

func waitPurchaseOrderDetail(ctx context.Context, t *testing.T, c *testutil.Client, id int64, ready func(PurchaseOrderDetail) bool) PurchaseOrderDetail {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	var last PurchaseOrderDetail
	for time.Now().Before(deadline) {
		var po PurchaseOrderDetail
		err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/purchase/orders/%d", id), nil, nil, &po)
		if err == nil && ready(po) {
			return po
		}
		if err == nil {
			last = po
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("purchase order not ready: id=%d stockInID=%d", id, last.StockInID)
	return last
}

func findStockInItemIDByMaterial(stockIn StockInDetail, materialID int64) int64 {
	for _, it := range stockIn.Items {
		if it.MaterialID == materialID {
			return it.ID
		}
	}
	return 0
}

// ============================================================================
// 领料订单
// ============================================================================

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
	t.Fatalf("no inventory available for material=%d qty>=%.2f", materialID, qty)
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

func waitConsumptionOrder(ctx context.Context, t *testing.T, c *testutil.Client, id int64, ready func(ConsumptionOrder) bool) ConsumptionOrder {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	var last ConsumptionOrder
	for time.Now().Before(deadline) {
		var o ConsumptionOrder
		err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/consumption/orders/%d", id), nil, nil, &o)
		if err == nil && ready(o) {
			return o
		}
		if err == nil {
			last = o
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("consumption order not ready: id=%d stockOutID=%d", id, last.StockOutID)
	return last
}

// ============================================================================
// 出库
// ============================================================================

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

// ============================================================================
// 退料
// ============================================================================

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
	t.Fatalf("no inventory issued for material=%d qty>=%.2f", materialID, qty)
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

func waitReversalOrder(ctx context.Context, t *testing.T, c *testutil.Client, id int64, ready func(ReversalOrder) bool) ReversalOrder {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	var last ReversalOrder
	for time.Now().Before(deadline) {
		var o ReversalOrder
		err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/reversal/orders/%d", id), nil, nil, &o)
		if err == nil && ready(o) {
			return o
		}
		if err == nil {
			last = o
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("reversal order not ready: id=%d stockInID=%d", id, last.StockInID)
	return last
}

// ============================================================================
// 退货
// ============================================================================

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

func waitReturnOrder(ctx context.Context, t *testing.T, c *testutil.Client, id int64, ready func(ReturnOrder) bool) ReturnOrder {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	var last ReturnOrder
	for time.Now().Before(deadline) {
		var o ReturnOrder
		err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/returns/%d", id), nil, nil, &o)
		if err == nil && ready(o) {
			return o
		}
		if err == nil {
			last = o
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("return order not ready: id=%d stockOutID=%v", id, last.StockOutID)
	return last
}

// ============================================================================
// 追踪/台账/大屏
// ============================================================================

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

// ============================================================================
// 通用工具
// ============================================================================

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
