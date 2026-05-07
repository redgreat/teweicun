/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type InventoryQuery struct {
	PageQuery
	MaterialCode string `form:"material_code"`
	MaterialName string `form:"material_name"`
	WarehouseID  int64  `form:"warehouse_id"`
	CategoryPath string `form:"category_path"`
}

type InventoryAvailableQuery struct {
	PageQuery
	WarehouseCode string `form:"warehouse_code"`
	SupplierCode  string `form:"supplier_code"`
	Q             string `form:"q"`
}

// InventoryIssuedQuery 查询“已出库(已领用)”可退回库存（按库存批次聚合）
type InventoryIssuedQuery struct {
	PageQuery
	WarehouseCode string `form:"warehouse_code"`
	Q             string `form:"q"`
}

type InventoryMaterialLedgerQuery struct {
	PageQuery
	MaterialName  string  `form:"material_name"`
	SKUName       string  `form:"sku_name"`
	WarehouseName string  `form:"warehouse_name"`
	PriceMin      float64 `form:"price_min"`
	PriceMax      float64 `form:"price_max"`
}

type InventoryMaterialLedgerSerialQuery struct {
	MaterialID  int64   `form:"material_id" binding:"required,gt=0"`
	WarehouseID int64   `form:"warehouse_id" binding:"required,gt=0"`
	SKUID       int64   `form:"sku_id"`
	UnitCost    float64 `form:"unit_cost"`
}

type InventorySKULedgerQuery = InventoryMaterialLedgerQuery
type InventorySKUSerialQuery = InventoryMaterialLedgerSerialQuery

type StockOutQuery struct {
	PageQuery
	StockOutNo    string `form:"stock_out_no"`
	OrderNo       string `form:"order_no"`
	Status        string `form:"status"`
	OutType       string `form:"out_type"`
	RefDocType    string `form:"ref_doc_type"`
	WarehouseCode string `form:"warehouse_code"`
	Receiver      string `form:"receiver"`     // 仅 stock_out.receiver 模糊
	StartDate     string `form:"start_date"`   // 出库日期起
	EndDate       string `form:"end_date"`     // 出库日期止
}
