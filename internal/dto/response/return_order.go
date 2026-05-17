/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

import "time"

type ReturnOrderItemResp struct {
	ID                int64   `json:"id"`
	MaterialID        int64   `json:"material_id"`
	MaterialCode      string  `json:"material_code"`
	MaterialName      string  `json:"material_name"`
	InventoryID       int64   `json:"inventory_id"`
	WarehouseCode     string  `json:"warehouse_code"`
	WarehouseName     string  `json:"warehouse_name"`
	Quantity          float64 `json:"quantity"`
	Unit              string  `json:"unit"`
	AvailableQuantity float64 `json:"available_quantity,omitempty"`
	UnitCost          float64 `json:"unit_cost,omitempty"`
}

type ReturnOrderResp struct {
	ID            int64                 `json:"id"`
	ReturnNo      string                `json:"return_no"`
	ReturnType    string                `json:"return_type"`
	RefDocType    string                `json:"ref_doc_type"`
	RefDocID      *int64                `json:"ref_doc_id"`
	WarehouseCode string                `json:"warehouse_code"`
	WarehouseName string                `json:"warehouse_name"`
	SupplierCode  string                `json:"supplier_code"`
	SupplierName  string                `json:"supplier_name"`
	CustomerCode  string                `json:"customer_code"`
	CustomerName  string                `json:"customer_name"`
	StockOutID    *int64                `json:"stock_out_id"`
	StockOutNo    string                `json:"stock_out_no"`
	StockInID     int64                 `json:"stock_in_id"`
	StockInNo     string                `json:"stock_in_no"`
	ReturnDate    time.Time             `json:"return_date"`
	Status        string                `json:"status"`
	TotalAmount   float64               `json:"total_amount"`
	Remark        string                `json:"remark"`
	CreatedAt     time.Time             `json:"created_at"`
	Items         []ReturnOrderItemResp `json:"items,omitempty"`
}
