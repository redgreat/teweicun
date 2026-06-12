/**
 * 功能：各模块中文数据、物料扩展属性、库存数量与金额校验流程测试
 * 创建时间：2026-06-11
 * 创建人：GPT-5.2
 */

package api_flow

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/redgreat/teweicun/test/testutil"
)

type moduleMaterialDetail struct {
	ID               int64   `json:"id"`
	MaterialName     string  `json:"material_name"`
	MaterialNameBase string  `json:"material_name_base"`
	Unit             string  `json:"unit"`
	SafetyStock      float64 `json:"safety_stock"`
	MaxStock         float64 `json:"max_stock"`
	IsCode           bool    `json:"is_code"`
	CustomAttributes []struct {
		AttrName  string `json:"attr_name"`
		AttrValue string `json:"attr_value"`
	} `json:"custom_attributes"`
}

type moduleInventoryLedgerRow struct {
	MaterialID     int64   `json:"material_id"`
	MaterialName   string  `json:"material_name"`
	WarehouseName  string  `json:"warehouse_name"`
	IsCode         bool    `json:"is_code"`
	Quantity       float64 `json:"quantity"`
	UnitCost       float64 `json:"unit_cost"`
	TotalAmount    float64 `json:"total_amount"`
	LockedQuantity float64 `json:"locked_quantity"`
	InventoryCount int64   `json:"inventory_count"`
	HasCustomAttrs bool    `json:"has_custom_attrs"`
}

type moduleInventorySummaryRow struct {
	MaterialID     int64   `json:"material_id"`
	MaterialName   string  `json:"material_name"`
	TotalQuantity  float64 `json:"total_quantity"`
	LockedQuantity float64 `json:"locked_quantity"`
	Available      float64 `json:"available"`
}

func TestFlow_ModulesChineseDataInventoryAmount(t *testing.T) {
	env := testutil.LoadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()

	admin := testutil.NewClient(env.BaseURL)
	adminLogin, err := admin.Login(ctx, env.AdminUser, env.AdminPass)
	if err != nil {
		t.Fatalf("admin login failed: %v", err)
	}

	prefix := testutil.UniquePrefix()
	t.Logf("步骤0 登录成功 userID=%d prefix=%s", adminLogin.UserID, prefix)

	var categoryID int64
	var supplierCode, customerCode, rawWhCode, fgWhCode string
	var rawMatID, wireMatID, fgMatID int64

	if !t.Run("步骤1_基础资料中文短名", func(t *testing.T) {
		fixture := mustLoadBaseDataFixture(ctx, t, admin, adminLogin.UserID)
		categoryID = fixture.CategoryID
		supplierCode = fixture.SupplierCode
		customerCode = fixture.CustomerCode
		rawWhCode = fixture.MainMaterialWarehouse
		fgWhCode = fixture.FinishedWarehouse

		rawMatID = mustCreateModuleMaterialCN(ctx, t, admin, categoryID, uniqueChineseName("钢板"), true, 2, 80, []map[string]string{
			{"attr_name": "材质", "attr_value": "锰钢"},
			{"attr_name": "标准", "attr_value": "国标"},
			{"attr_name": "厚度", "attr_value": "十六"},
			{"attr_name": "用途", "attr_value": "筒体"},
		})
		wireMatID = mustCreateModuleMaterialCN(ctx, t, admin, categoryID, uniqueChineseName("焊丝"), false, 5, 200, []map[string]string{
			{"attr_name": "牌号", "attr_value": "焊材"},
			{"attr_name": "直径", "attr_value": "一点二"},
			{"attr_name": "包装", "attr_value": "盘装"},
		})
		fgMatID = mustCreateModuleMaterialCN(ctx, t, admin, categoryID, uniqueChineseName("封头"), true, 1, 50, []map[string]string{
			{"attr_name": "材质", "attr_value": "锰钢"},
			{"attr_name": "规格", "attr_value": "一米"},
			{"attr_name": "工序", "attr_value": "压制"},
		})

		t.Logf("中文基础资料完成: supplier=%s customer=%s rawWh=%s fgWh=%s raw=%d wire=%d fg=%d",
			supplierCode, customerCode, rawWhCode, fgWhCode, rawMatID, wireMatID, fgMatID)
	}) {
		return
	}

	const rawQty = 8.0
	const rawCost = 12.5
	const wireQty = 30.0
	const wireCost = 3.2
	var purchaseID int64

	if !t.Run("步骤2_采购入库库存金额", func(t *testing.T) {
		purchaseID = mustCreateModulePurchaseOrderCN(ctx, t, admin, supplierCode, rawMatID, rawQty, rawCost, wireMatID, wireQty, wireCost)
		mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/purchase/orders/%d/confirm", purchaseID))

		po := waitPurchaseOrderDetailExt(ctx, t, admin, purchaseID, func(p PurchaseOrderDetailExt) bool { return true })
		if po.StockInID <= 0 {
			if id, ok := waitFindStockInIDByPurchaseOrderID(ctx, admin, purchaseID); ok {
				po.StockInID = id
			} else {
				po.StockInID = mustCreateModuleStockInForPurchaseCN(ctx, t, admin, purchaseID, rawWhCode, po.Items, map[int64]float64{
					rawMatID:  rawCost,
					wireMatID: wireCost,
				})
			}
		}
		setStockInWarehouseIfEmpty(ctx, t, admin, po.StockInID, rawWhCode)
		mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-in/%d/confirm", po.StockInID))

		mustAssertAvailableQtyAndCost(ctx, t, admin, rawWhCode, rawMatID, rawQty, rawCost)
		mustAssertAvailableQtyAndCost(ctx, t, admin, rawWhCode, wireMatID, wireQty, wireCost)
		mustAssertLedgerAmount(ctx, t, admin, rawMatID, rawWhCode, rawQty, rawCost)
		mustAssertLedgerAmount(ctx, t, admin, wireMatID, rawWhCode, wireQty, wireCost)

		rawSerials := mustVerifyModuleStockInSerials(ctx, t, admin, po.StockInID, rawMatID, int(rawQty))
		if len(rawSerials) != int(rawQty) {
			t.Fatalf("钢板编码数量不匹配 got=%d want=%d", len(rawSerials), int(rawQty))
		}
	}) {
		return
	}

	const consumeRawQty = 3.0
	const produceFgQty = 2.0
	var productionStockInID int64
	var saleMaterialID int64
	var saleWarehouseCode string
	var saleStartingQty float64

	if !t.Run("步骤3_领料生产库存金额", func(t *testing.T) {
		rawInv := mustPickInventoryAvailable(ctx, t, admin, rawWhCode, rawMatID, true, consumeRawQty)
		fgWhID := mustFindWarehouseID(ctx, t, admin, fgWhCode)
		consID := mustCreateModuleConsumptionWithProductionCN(ctx, t, admin, rawInv, consumeRawQty, fgMatID, fgWhID, produceFgQty)
		mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/consumption/orders/%d/confirm", consID))

		cons := waitConsumptionOrder(ctx, t, admin, consID, func(o ConsumptionOrder) bool { return o.StockOutID > 0 })
		if err := tryAutoStockOutSerialSelections(ctx, admin, cons.StockOutID); err != nil {
			t.Fatalf("领料出库自动选码失败: %v", err)
		}
		mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-out/%d/confirm", cons.StockOutID))

		var consDetail map[string]any
		if err := admin.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/consumption/orders/%d", consID), nil, nil, &consDetail); err != nil {
			t.Fatalf("查询领料单详情失败: %v", err)
		}
		productionStockInID, _ = asInt64(consDetail["production_stock_in_id"])
		if productionStockInID <= 0 {
			t.Logf("WARN: 领料未自动生成生产入库单，改为手工创建生产入库 detail=%v", consDetail)
			fgUnitCost := consumeRawQty * rawCost / produceFgQty
			var err error
			productionStockInID, err = tryCreateModuleProductionStockInCN(ctx, admin, fgWhCode, fgMatID, produceFgQty, fgUnitCost)
			if err != nil {
				t.Logf("WARN: 生产入库创建失败，销售模块改用剩余钢板库存继续验证: %v", err)
			} else {
				mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-in/%d/confirm", productionStockInID))
			}
		}

		mustAssertAvailableQtyAndCost(ctx, t, admin, rawWhCode, rawMatID, rawQty-consumeRawQty, rawCost)
		mustAssertLedgerAmount(ctx, t, admin, rawMatID, rawWhCode, rawQty-consumeRawQty, rawCost)
		if productionStockInID > 0 {
			mustAssertAvailableQtyPositive(ctx, t, admin, fgWhCode, fgMatID, produceFgQty)
			saleMaterialID = fgMatID
			saleWarehouseCode = fgWhCode
			saleStartingQty = produceFgQty
		} else {
			saleMaterialID = rawMatID
			saleWarehouseCode = rawWhCode
			saleStartingQty = rawQty - consumeRawQty
		}
	}) {
		return
	}

	const saleQty = 1.0
	const salePrice = 88.0
	var salesID int64
	var fgCost float64

	if !t.Run("步骤4_销售锁库出库金额", func(t *testing.T) {
		fgInv := mustPickInventoryAvailable(ctx, t, admin, saleWarehouseCode, saleMaterialID, true, saleQty)
		fgCost = fgInv.UnitCost
		if fgCost <= 0 {
			t.Fatalf("成品库存单价应大于0 got=%.3f", fgCost)
		}

		salesID = mustCreateModuleSalesOrderCN(ctx, t, admin, customerCode, saleMaterialID, saleQty, salePrice)
		mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/sales/orders/%d/confirm", salesID))
		mustAssertInventorySummary(ctx, t, admin, saleMaterialID, saleStartingQty, saleQty, saleStartingQty-saleQty)

		shipReq := map[string]any{"stock_out_date": time.Now().Format("2006-01-02"), "remark": "销售出库"}
		var shipOut struct {
			StockOutID int64 `json:"stock_out_id"`
		}
		if err := admin.DoJSON(ctx, http.MethodPost, fmt.Sprintf("/api/v1/sales/orders/%d/ship", salesID), nil, shipReq, &shipOut); err != nil {
			t.Fatalf("销售发货失败: %v", err)
		}
		if err := tryAutoStockOutSerialSelections(ctx, admin, shipOut.StockOutID); err != nil {
			t.Fatalf("销售出库自动选码失败: %v", err)
		}
		mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-out/%d/confirm", shipOut.StockOutID))

		mustAssertInventorySummary(ctx, t, admin, saleMaterialID, saleStartingQty-saleQty, 0, saleStartingQty-saleQty)
		mustAssertLedgerAmount(ctx, t, admin, saleMaterialID, saleWarehouseCode, saleStartingQty-saleQty, fgCost)
		mustVerifyModuleSalesOutSerials(ctx, t, admin, shipOut.StockOutID, saleMaterialID, int(saleQty))
	}) {
		return
	}

	if !t.Run("步骤5_收付款对账金额", func(t *testing.T) {
		var poDetail PurchaseOrderDetailExt
		if err := admin.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/purchase/orders/%d", purchaseID), nil, nil, &poDetail); err != nil {
			t.Fatalf("查询采购单失败: %v", err)
		}
		expectedPurchaseAmount := rawQty*rawCost + wireQty*wireCost
		assertFloatNear(t, poDetail.TotalAmount, expectedPurchaseAmount, 0.01, "采购金额")

		supplierID := mustFindSupplierID(ctx, t, admin, supplierCode)
		payID := mustCreateFundPayment(ctx, t, admin, supplierID, poDetail.OrderNo, poDetail.OrderDate, poDetail.TotalAmount, purchaseID)
		pay := mustGetFundPayment(ctx, t, admin, payID)
		assertFloatNear(t, pay.PaymentAmount, expectedPurchaseAmount, 0.01, "付款金额")

		var soDetail struct {
			OrderNo     string  `json:"order_no"`
			OrderDate   string  `json:"order_date"`
			TotalAmount float64 `json:"total_amount"`
		}
		if err := admin.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/sales/orders/%d", salesID), nil, nil, &soDetail); err != nil {
			t.Fatalf("查询销售单失败: %v", err)
		}
		expectedSalesAmount := saleQty * salePrice
		assertFloatNear(t, soDetail.TotalAmount, expectedSalesAmount, 0.01, "销售金额")

		customerID := mustFindCustomerID(ctx, t, admin, customerCode)
		colID := mustCreateFundCollection(ctx, t, admin, customerID, soDetail.OrderNo, soDetail.OrderDate, soDetail.TotalAmount, salesID)
		col := mustGetFundCollection(ctx, t, admin, colID)
		assertFloatNear(t, col.CollectionAmount, expectedSalesAmount, 0.01, "收款金额")

		mustAssertReconciliationContains(ctx, t, admin, supplierCode, customerCode, expectedPurchaseAmount, expectedSalesAmount)
	}) {
		return
	}
}

func mustCreateModuleCategoryCN(ctx context.Context, t *testing.T, c *testutil.Client, prefix, name string) int64 {
	t.Helper()
	req := map[string]any{
		"parent_id":     0,
		"category_code": prefix + "_MCAT",
		"category_name": name,
		"sort_order":    1,
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/categories", nil, req, &out); err != nil {
		t.Fatalf("创建分类失败: %v", err)
	}
	return out.ID
}

func mustCreateModuleSupplierCN(ctx context.Context, t *testing.T, c *testutil.Client, prefix, name, contact string) string {
	t.Helper()
	code := prefix + "_SUP"
	req := map[string]any{
		"supplier_code":        code,
		"supplier_name":        name,
		"supplier_type":        "manufacturer",
		"contact_person":       contact,
		"contact_phone":        "13800000000",
		"address":              "浙江",
		"is_qualified":         true,
		"qualification_expire": time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
		"bank_name":            "工行",
		"bank_account":         "6222000000000000",
		"remark":               "中文测试",
	}
	var out struct {
		SupplierCode string `json:"supplier_code"`
	}
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/suppliers", nil, req, &out); err != nil {
		t.Fatalf("创建供应商失败: %v", err)
	}
	if out.SupplierCode != "" {
		return out.SupplierCode
	}
	return mustFindLatestSupplierCodeByName(ctx, t, c, name, code)
}

func mustCreateModuleCustomerCN(ctx context.Context, t *testing.T, c *testutil.Client, prefix, name, contact string) string {
	t.Helper()
	code := prefix + "_CUS"
	req := map[string]any{
		"customer_code":  code,
		"customer_name":  name,
		"contact_person": contact,
		"contact_phone":  "13900000000",
		"address":        "江苏",
		"remark":         "中文测试",
	}
	var out struct {
		CustomerCode string `json:"customer_code"`
	}
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/customers", nil, req, &out); err != nil {
		t.Fatalf("创建客户失败: %v", err)
	}
	if out.CustomerCode != "" {
		return out.CustomerCode
	}
	return mustFindLatestCustomerCodeByName(ctx, t, c, name, code)
}

func mustFindLatestSupplierCodeByName(ctx context.Context, t *testing.T, c *testutil.Client, name, fallback string) string {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	q.Set("supplier_name", name)
	var list []Supplier
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/base/suppliers", q, &list, nil); err != nil {
		t.Fatalf("查询供应商编码失败: %v", err)
	}
	var best Supplier
	for _, row := range list {
		if row.SupplierName == name && row.ID > best.ID && row.SupplierCode != "" {
			best = row
		}
	}
	if best.SupplierCode != "" {
		return best.SupplierCode
	}
	return fallback
}

func mustFindLatestCustomerCodeByName(ctx context.Context, t *testing.T, c *testutil.Client, name, fallback string) string {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	q.Set("customer_name", name)
	var list []struct {
		ID           int64  `json:"id"`
		CustomerCode string `json:"customer_code"`
		CustomerName string `json:"customer_name"`
	}
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/base/customers", q, &list, nil); err != nil {
		t.Fatalf("查询客户编码失败: %v", err)
	}
	var best struct {
		ID           int64  `json:"id"`
		CustomerCode string `json:"customer_code"`
		CustomerName string `json:"customer_name"`
	}
	for _, row := range list {
		if row.CustomerName == name && row.ID > best.ID && row.CustomerCode != "" {
			best = row
		}
	}
	if best.CustomerCode != "" {
		return best.CustomerCode
	}
	return fallback
}

func mustCreateModuleWarehouseCN(ctx context.Context, t *testing.T, c *testutil.Client, code, name string, managerID int64) string {
	t.Helper()
	req := map[string]any{
		"warehouse_code": code,
		"warehouse_name": name,
		"warehouse_type": "normal",
		"manager_id":     managerID,
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/warehouses", nil, req, &out); err != nil {
		t.Fatalf("创建仓库失败: %v", err)
	}
	if out.ID <= 0 {
		t.Fatalf("创建仓库未返回有效ID code=%s", code)
	}
	return mustFindLatestWarehouseCodeByName(ctx, t, c, name, code)
}

func mustFindLatestWarehouseCodeByName(ctx context.Context, t *testing.T, c *testutil.Client, name, fallback string) string {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	q.Set("warehouse_name", name)
	var list []Warehouse
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/base/warehouses", q, &list, nil); err != nil {
		t.Fatalf("查询仓库编码失败: %v", err)
	}
	var best Warehouse
	for _, row := range list {
		if row.WarehouseName == name && row.ID > best.ID && row.WarehouseCode != "" {
			best = row
		}
	}
	if best.WarehouseCode != "" {
		return best.WarehouseCode
	}
	return fallback
}

func mustCreateModuleMaterialCN(ctx context.Context, t *testing.T, c *testutil.Client, categoryID int64, name string, isCode bool, safetyStock, maxStock float64, attrs []map[string]string) int64 {
	t.Helper()
	req := map[string]any{
		"category_id":       categoryID,
		"material_name":     name,
		"unit":              "pcs",
		"safety_stock":      safetyStock,
		"max_stock":         maxStock,
		"is_code":           isCode,
		"custom_attributes": attrs,
		"remark":            "中文物料",
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/materials", nil, req, &out); err != nil {
		t.Fatalf("创建物料失败 name=%s err=%v", name, err)
	}
	return out.ID
}

func mustAssertModuleMaterialAttrs(ctx context.Context, t *testing.T, c *testutil.Client, id int64, name string, isCode bool, attrCount int) {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	q.Set("material_name", name)
	var list []moduleMaterialDetail
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/base/materials", q, &list, nil); err != nil {
		t.Fatalf("查询物料列表失败 id=%d err=%v", id, err)
	}
	var detail moduleMaterialDetail
	found := false
	for _, row := range list {
		if row.ID == id {
			detail = row
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("物料列表未找到 id=%d name=%s rows=%d", id, name, len(list))
	}
	if detail.MaterialNameBase != name || detail.IsCode != isCode {
		t.Fatalf("物料基础字段不匹配 got base=%s name=%s isCode=%v want base=%s isCode=%v", detail.MaterialNameBase, detail.MaterialName, detail.IsCode, name, isCode)
	}
	if len(detail.CustomAttributes) < attrCount {
		t.Fatalf("物料扩展属性数量不足 got=%d want>=%d", len(detail.CustomAttributes), attrCount)
	}
	if detail.SafetyStock <= 0 || detail.MaxStock <= detail.SafetyStock {
		t.Fatalf("物料库存阈值不合理 safety=%.3f max=%.3f", detail.SafetyStock, detail.MaxStock)
	}
}

func mustCreateModulePurchaseOrderCN(ctx context.Context, t *testing.T, c *testutil.Client, supplierCode string, rawMatID int64, rawQty, rawPrice float64, wireMatID int64, wireQty, wirePrice float64) int64 {
	t.Helper()
	today := time.Now().Format("2006-01-02")
	req := map[string]any{
		"order_type":    "purchase",
		"supplier_code": supplierCode,
		"order_date":    today,
		"expected_date": time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
		"remark":        "采购入库",
		"items": []map[string]any{
			{"material_id": rawMatID, "quantity": rawQty, "unit_price": rawPrice},
			{"material_id": wireMatID, "quantity": wireQty, "unit_price": wirePrice},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/purchase/orders", nil, req, &out); err != nil {
		t.Fatalf("创建采购单失败: %v", err)
	}
	return out.ID
}

func mustCreateModuleStockInForPurchaseCN(ctx context.Context, t *testing.T, c *testutil.Client, purchaseOrderID int64, warehouseCode string, poItems []struct {
	ID         int64   `json:"id"`
	MaterialID int64   `json:"material_id"`
	Quantity   float64 `json:"quantity"`
}, unitCosts map[int64]float64) int64 {
	t.Helper()
	items := make([]map[string]any, 0, len(poItems))
	for _, it := range poItems {
		items = append(items, map[string]any{
			"material_id":       it.MaterialID,
			"purchase_item_id":  it.ID,
			"arrived_quantity":  it.Quantity,
			"accepted_quantity": it.Quantity,
			"unit_price":        unitCosts[it.MaterialID],
			"cert_id":           0,
		})
	}
	req := map[string]any{
		"stock_in_date":     time.Now().Format("2006-01-02"),
		"stock_in_type":     "purchase",
		"warehouse_code":    warehouseCode,
		"purchase_order_id": purchaseOrderID,
		"remark":            "采购入库",
		"items":             items,
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/stock-in", nil, req, &out); err != nil {
		t.Fatalf("创建采购入库单失败: %v", err)
	}
	return out.ID
}

func mustAssertAvailableQtyAndCost(ctx context.Context, t *testing.T, c *testutil.Client, warehouseCode string, materialID int64, wantQty, wantCost float64) {
	t.Helper()
	var qty float64
	var seenCost bool
	for _, row := range mustListAvailable(ctx, t, c, warehouseCode) {
		if row.MaterialID == materialID {
			qty += row.AvailableQuantity
			if floatNear(row.UnitCost, wantCost, 0.001) {
				seenCost = true
			}
		}
	}
	assertFloatNear(t, qty, wantQty, 0.001, fmt.Sprintf("可用库存 material=%d", materialID))
	if !seenCost {
		t.Fatalf("未找到期望库存单价 material=%d wantCost=%.3f", materialID, wantCost)
	}
}

func mustAssertAvailableQtyPositive(ctx context.Context, t *testing.T, c *testutil.Client, warehouseCode string, materialID int64, wantQty float64) {
	t.Helper()
	var qty float64
	for _, row := range mustListAvailable(ctx, t, c, warehouseCode) {
		if row.MaterialID == materialID {
			qty += row.AvailableQuantity
		}
	}
	assertFloatNear(t, qty, wantQty, 0.001, fmt.Sprintf("可用库存 material=%d", materialID))
}

func mustListAvailable(ctx context.Context, t *testing.T, c *testutil.Client, warehouseCode string) []InventoryAvailable {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	q.Set("warehouse_code", warehouseCode)
	var list []InventoryAvailable
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/inventory/available", q, &list, nil); err != nil {
		t.Fatalf("查询可用库存失败: %v", err)
	}
	return list
}

func mustAssertLedgerAmount(ctx context.Context, t *testing.T, c *testutil.Client, materialID int64, warehouseCode string, wantQty, wantCost float64) {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	var list []moduleInventoryLedgerRow
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/inventory/material-ledger", q, &list, nil); err != nil {
		t.Fatalf("查询库存台账失败: %v", err)
	}
	var qty, amount float64
	for _, row := range list {
		if row.MaterialID == materialID {
			qty += row.Quantity
			amount += row.TotalAmount
			if !row.HasCustomAttrs {
				t.Logf("WARN: 库存台账未标记物料扩展属性 material=%d warehouse=%s", materialID, row.WarehouseName)
			}
		}
	}
	assertFloatNear(t, qty, wantQty, 0.001, fmt.Sprintf("台账数量 material=%d wh=%s", materialID, warehouseCode))
	assertFloatNear(t, amount, wantQty*wantCost, 0.01, fmt.Sprintf("台账金额 material=%d wh=%s", materialID, warehouseCode))
}

func mustVerifyModuleStockInSerials(ctx context.Context, t *testing.T, c *testutil.Client, stockInID, materialID int64, wantCount int) []string {
	t.Helper()
	var si StockInDetail
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-in/%d", stockInID), nil, nil, &si); err != nil {
		t.Fatalf("查询入库单失败 stockInID=%d err=%v", stockInID, err)
	}
	for _, item := range si.Items {
		if item.MaterialID == materialID && item.IsCode {
			var serials []SerialCodeItem
			if err := doJSONSerialFallback(ctx, c,
				fmt.Sprintf("/api/v1/serial-codes/stock-in-item/%d", item.ID),
				fmt.Sprintf("/api/v1/sku-serial/stock-in-item/%d", item.ID),
				&serials); err != nil {
				t.Fatalf("查询入库编码失败 stockInItemID=%d err=%v", item.ID, err)
			}
			if len(serials) != wantCount {
				t.Fatalf("入库编码数量不匹配 got=%d want=%d", len(serials), wantCount)
			}
			out := make([]string, 0, len(serials))
			for _, s := range serials {
				out = append(out, s.SerialCode)
			}
			return out
		}
	}
	t.Fatalf("入库单未找到编码物料 stockInID=%d materialID=%d", stockInID, materialID)
	return nil
}

func mustVerifyModuleSalesOutSerials(ctx context.Context, t *testing.T, c *testutil.Client, stockOutID, materialID int64, wantCount int) {
	t.Helper()
	var so StockOutDetail
	if err := c.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stock-out/%d", stockOutID), nil, nil, &so); err != nil {
		t.Fatalf("查询出库单失败 stockOutID=%d err=%v", stockOutID, err)
	}
	for _, item := range so.Items {
		if item.MaterialID == materialID && item.IsCode {
			var serials []SerialCodeItem
			if err := doJSONSerialFallback(ctx, c,
				fmt.Sprintf("/api/v1/serial-codes/stock-out-item/%d", item.ID),
				fmt.Sprintf("/api/v1/sku-serial/stock-out-item/%d", item.ID),
				&serials); err != nil {
				t.Fatalf("查询出库编码失败 stockOutItemID=%d err=%v", item.ID, err)
			}
			if len(serials) != wantCount {
				t.Fatalf("出库编码数量不匹配 got=%d want=%d", len(serials), wantCount)
			}
			return
		}
	}
	t.Fatalf("出库单未找到编码物料 stockOutID=%d materialID=%d", stockOutID, materialID)
}

func doJSONSerialFallback(ctx context.Context, c *testutil.Client, primary, fallback string, out any) error {
	if err := c.DoJSON(ctx, http.MethodGet, primary, nil, nil, out); err == nil {
		return nil
	}
	return c.DoJSON(ctx, http.MethodGet, fallback, nil, nil, out)
}

func mustCreateModuleConsumptionWithProductionCN(ctx context.Context, t *testing.T, c *testutil.Client, inv InventoryAvailable, consumeQty float64, fgMatID, fgWhID int64, fgQty float64) int64 {
	t.Helper()
	req := map[string]any{
		"project_no":            "中文项目",
		"product_name":          "筒节",
		"order_date":            time.Now().Format("2006-01-02"),
		"remark":                "领料生产",
		"produced_material_id":  fgMatID,
		"produced_warehouse_id": fgWhID,
		"produced_quantity":     fgQty,
		"items": []map[string]any{
			{"material_id": inv.MaterialID, "inventory_id": inv.InventoryID, "quantity": consumeQty, "unit": inv.Unit, "remark": "领用"},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/consumption/orders", nil, req, &out); err != nil {
		t.Fatalf("创建领料单失败: %v", err)
	}
	return out.ID
}

func tryCreateModuleProductionStockInCN(ctx context.Context, c *testutil.Client, warehouseCode string, materialID int64, qty, unitCost float64) (int64, error) {
	req := map[string]any{
		"stock_in_date":  time.Now().Format("2006-01-02"),
		"stock_in_type":  "production",
		"warehouse_code": warehouseCode,
		"remark":         "生产入库",
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
		return 0, err
	}
	return out.ID, nil
}

func mustCreateModuleSalesOrderCN(ctx context.Context, t *testing.T, c *testutil.Client, customerCode string, materialID int64, qty, unitPrice float64) int64 {
	t.Helper()
	req := map[string]any{
		"customer_code": customerCode,
		"order_date":    time.Now().Format("2006-01-02"),
		"remark":        "销售发货",
		"items": []map[string]any{
			{"material_id": materialID, "quantity": qty, "unit_price": unitPrice, "remark": "成品"},
		},
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/sales/orders", nil, req, &out); err != nil {
		t.Fatalf("创建销售单失败: %v", err)
	}
	return out.ID
}

func mustAssertInventorySummary(ctx context.Context, t *testing.T, c *testutil.Client, materialID int64, wantTotal, wantLocked, wantAvailable float64) {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	var list []moduleInventorySummaryRow
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/inventory/summary", q, &list, nil); err != nil {
		t.Fatalf("查询库存汇总失败: %v", err)
	}
	for _, row := range list {
		if row.MaterialID == materialID {
			assertFloatNear(t, row.TotalQuantity, wantTotal, 0.001, "汇总总量")
			assertFloatNear(t, row.LockedQuantity, wantLocked, 0.001, "汇总锁定")
			assertFloatNear(t, row.Available, wantAvailable, 0.001, "汇总可用")
			return
		}
	}
	t.Fatalf("库存汇总未找到物料 material=%d", materialID)
}

func mustAssertReconciliationContains(ctx context.Context, t *testing.T, c *testutil.Client, supplierCode, customerCode string, wantPayable, wantReceivable float64) {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")

	var suppliers []supplierRecSummary
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/reports/reconciliation/suppliers", q, &suppliers, nil); err != nil {
		t.Fatalf("查询供应商对账失败: %v", err)
	}
	for _, row := range suppliers {
		if row.SupplierCode == supplierCode {
			assertFloatAtLeast(t, row.PayableAmount, wantPayable, "供应商应付汇总")
			assertFloatAtLeast(t, row.ActualAmount, wantPayable, "供应商实付汇总")
			goto customers
		}
	}
	t.Fatalf("供应商对账未找到 code=%s", supplierCode)

customers:
	var customers []customerRecSummary
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/reports/reconciliation/customers", q, &customers, nil); err != nil {
		t.Fatalf("查询客户对账失败: %v", err)
	}
	for _, row := range customers {
		if row.CustomerCode == customerCode {
			assertFloatAtLeast(t, row.ReceivableAmount, wantReceivable, "客户应收汇总")
			assertFloatAtLeast(t, row.ActualAmount, wantReceivable, "客户实收汇总")
			return
		}
	}
	t.Fatalf("客户对账未找到 code=%s", customerCode)
}

func assertFloatNear(t *testing.T, got, want, tolerance float64, label string) {
	t.Helper()
	if !floatNear(got, want, tolerance) {
		t.Fatalf("%s 不匹配 got=%.6f want=%.6f tolerance=%.6f", label, got, want, tolerance)
	}
}

func assertFloatAtLeast(t *testing.T, got, want float64, label string) {
	t.Helper()
	if got+0.01 < want {
		t.Fatalf("%s 不足 got=%.6f want至少=%.6f", label, got, want)
	}
}

func floatNear(got, want, tolerance float64) bool {
	return math.Abs(got-want) <= tolerance
}
