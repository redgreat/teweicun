/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

import (
	"time"

	"github.com/redgreat/teweicun/pkg/database"
)

type StockInResp struct {
	ID              int64               `json:"id"`
	StockInNo       string              `json:"stock_in_no"`
	StockInType     string              `json:"stock_in_type"`
	BusinessDocType string              `json:"business_doc_type"`
	BusinessDocID   int64               `json:"business_doc_id"`
	BusinessDocNo   string              `json:"business_doc_no"`
	StockInDate     time.Time           `json:"stock_in_date"`
	WarehouseCode   database.NullString `json:"warehouse_code"`
	WarehouseName   database.NullString `json:"warehouse_name"`
	SupplierCode    database.NullString `json:"supplier_code"`
	SupplierName    database.NullString `json:"supplier_name"`
	PurchaseOrderID int64               `json:"purchase_order_id"`
	PurchaseOrderNo database.NullString `json:"purchase_order_no"`
	ReversalOrderID int64               `json:"reversal_order_id"`
	ReversalOrderNo database.NullString `json:"reversal_order_no"`
	TotalAmount     float64             `json:"total_amount"`
	StockInStatus   string              `json:"stock_in_status"`
	Remark          database.NullString `json:"remark"`
	CreatedAt       time.Time           `json:"created_at"`
}

type StockInDetailResp struct {
	StockInResp
	HasStockIn bool              `json:"has_stock_in"`
	Items      []StockInItemResp `json:"items"`
}

type StockInItemResp struct {
	ID               int64               `json:"id"`
	MaterialID       int64               `json:"material_id"`
	MaterialCode     string              `json:"material_code"`
	MaterialName     string              `json:"material_name"`
	IsCode           bool                `json:"is_code"`
	PurchaseQuantity float64             `json:"purchase_quantity"`
	ReceivedQuantity float64             `json:"received_quantity"`
	ArrivedQuantity  float64             `json:"arrived_quantity"`
	AcceptedQuantity float64             `json:"accepted_quantity"`
	PendingQuantity  float64             `json:"pending_quantity"`
	UnitCost         *float64            `json:"unit_cost"`
}

type InventoryValueStatsResp struct {
	MaterialCount int64   `json:"material_count"`
	TotalQuantity float64 `json:"total_quantity"`
	TotalValue    float64 `json:"total_value"`
}

type StockInConfirmLogResp struct {
	ID                      int64     `json:"id"`
	StockInID               int64     `json:"stock_in_id"`
	StockInItemID           int64     `json:"stock_in_item_id"`
	MaterialID              int64     `json:"material_id"`
	MaterialCode            string    `json:"material_code"`
	MaterialName            string    `json:"material_name"`
	PurchaseQuantity        float64   `json:"purchase_quantity"`
	BeforeReceivedQuantity  float64   `json:"before_received_quantity"`
	CurrentReceivedQuantity float64   `json:"current_received_quantity"`
	AfterReceivedQuantity   float64   `json:"after_received_quantity"`
	OperatorID              int64     `json:"operator_id"`
	OperatorName            string    `json:"operator_name"`
	CreatedAt               time.Time `json:"created_at"`
	Remark                  string    `json:"remark"`
}
