/**
 * 功能：采购订单响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：CodeArts Agent
 */

package response

import (
	"time"

	"github.com/redgreat/teweicun/pkg/database"
)

type PurchaseOrderResp struct {
	ID              int64               `json:"id"`
	OrderNo         string              `json:"order_no"`
	StockInID       int64               `json:"stock_in_id"`
	StockInNo       string              `json:"stock_in_no"`
	WarehouseCode   string              `json:"warehouse_code"`
	WarehouseName   string              `json:"warehouse_name"`
	SupplierCode    string              `json:"supplier_code"`
	SupplierName    string              `json:"supplier_name"`
	OrderDate       time.Time           `json:"order_date"`
	ExpectedDate    *time.Time          `json:"expected_date"`
	OrderStatus     string              `json:"order_status"`
	OrderStatusName string              `json:"order_status_name"`
	TotalAmount     float64             `json:"total_amount"`
	Remark          database.NullString `json:"remark"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

type PurchaseOrderDetailResp struct {
	PurchaseOrderResp
	Items          []PurchaseOrderItemResp `json:"items"`
	StockInRecords []RelatedStockIn        `json:"stock_in_records"`
}

type RelatedStockIn struct {
	ID      int64  `json:"id"`
	OrderNo string `json:"order_no"`
}

type PurchaseOrderItemResp struct {
	ID               int64               `json:"id"`
	MaterialID       int64               `json:"material_id"`
	SKUID            *int64              `json:"sku_id"`
	SKUCode          *string             `json:"sku_code"`
	SKUName          *string             `json:"sku_name"`
	MaterialCode     string              `json:"material_code"`
	MaterialName     string              `json:"material_name"`
	Quantity         float64             `json:"quantity"`
	Unit             string              `json:"unit"`
	UnitPrice        *float64            `json:"unit_price"`
	Amount           *float64            `json:"amount"`
	ReceivedQuantity float64             `json:"received_quantity"`
}
