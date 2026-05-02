/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type StockInQuery struct {
	PageQuery
	StockInNo     string `form:"stock_in_no"`
	StockInType   string `form:"stock_in_type"`
	SupplierCode  string `form:"supplier_code"`
	Status        string `form:"status"` // 空；partial=部分入库(preparing|pending)；或 preparing/pending/passed/failed
	WarehouseCode string `form:"warehouse_code"`
	StartDate     string `form:"start_date"` // 入库日期起 YYYY-MM-DD
	EndDate       string `form:"end_date"`   // 入库日期止 YYYY-MM-DD
}

type CreateStockInItem struct {
	MaterialID       int64   `json:"material_id" binding:"required"`
	PurchaseItemID   int64   `json:"purchase_item_id"` // Optional
	ArrivedQuantity  float64 `json:"arrived_quantity" binding:"required,gt=0"`
	AcceptedQuantity float64 `json:"accepted_quantity" binding:"required,gt=0"`
	UnitPrice        float64 `json:"unit_price" binding:"required,gt=0"`
	CertID           int64   `json:"cert_id"` // Matches migration cert_id
}

type CreateStockInReq struct {
	StockInDate     string              `json:"stock_in_date" binding:"required"`
	StockInType     string              `json:"stock_in_type" binding:"required"` // e.g., 'purchase'
	WarehouseCode   string              `json:"warehouse_code" binding:"required"`
	PurchaseOrderID int64               `json:"purchase_order_id"`
	Remark          string              `json:"remark"`
	Items           []CreateStockInItem `json:"items" binding:"required,min=1,dive"`
}

type UpdateStockInItem struct {
	ID               int64   `json:"id" binding:"required"`
	MaterialID       int64   `json:"material_id" binding:"required"`
	ArrivedQuantity  float64 `json:"arrived_quantity" binding:"required,gt=0"`
	AcceptedQuantity float64 `json:"accepted_quantity" binding:"required,gt=0"`
	UnitCost         float64 `json:"unit_cost"`
	CertID           int64   `json:"cert_id"`
	SKUID            int64   `json:"sku_id"`
	CustomAttributes any     `json:"custom_attributes"`
}

type UpdateStockInReq struct {
	WarehouseCode string              `json:"warehouse_code"`
	Remark        string              `json:"remark"`
	Items         []UpdateStockInItem `json:"items" binding:"required,min=1,dive"`
}
