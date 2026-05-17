/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

import "time"

type StockOutItemResp struct {
	ID           int64   `json:"id"`
	MaterialID   int64   `json:"material_id"`
	MaterialCode string  `json:"material_code"`
	MaterialName string  `json:"material_name"`
	InventoryID  int64   `json:"inventory_id"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	UnitCost     float64 `json:"unit_cost,omitempty"`
	IsCode       bool    `json:"is_code"`
}

type StockOutResp struct {
	ID              int64              `json:"id"`
	StockOutNo      string             `json:"stock_out_no"`
	StockOutDate    time.Time          `json:"stock_out_date"`
	OutType         string             `json:"out_type"`
	RefDocType      string             `json:"ref_doc_type"`
	RefDocID        *int64             `json:"ref_doc_id"`
	BusinessDocType string             `json:"business_doc_type"`
	BusinessDocID   int64              `json:"business_doc_id"`
	BusinessDocNo   string             `json:"business_doc_no"`
	WarehouseCode   string             `json:"warehouse_code"`
	WarehouseName   string             `json:"warehouse_name"`
	CustomerCode    string             `json:"customer_code"`
	CustomerName    string             `json:"customer_name"`
	Receiver        string             `json:"receiver"`
	Status          string             `json:"status"`
	Remark          string             `json:"remark"`
	CreatedAt       time.Time          `json:"created_at"`
	ConfirmedAt     *time.Time         `json:"confirmed_at"`
	Items           []StockOutItemResp `json:"items,omitempty"`
	TotalAmount     float64            `json:"total_amount"`
}
