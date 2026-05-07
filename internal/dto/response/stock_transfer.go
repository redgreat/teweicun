/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

import "time"

type StockTransferItemResp struct {
	ID           int64   `json:"id"`
	MaterialID   int64   `json:"material_id"`
	MaterialCode string  `json:"material_code"`
	MaterialName string  `json:"material_name"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	Remark       string  `json:"remark"`
}

type StockTransferResp struct {
	ID              int64                   `json:"id"`
	TransferNo      string                  `json:"transfer_no"`
	FromWarehouseID int64                   `json:"from_warehouse_id"`
	FromWarehouseName string                `json:"from_warehouse_name"`
	ToWarehouseID   int64                   `json:"to_warehouse_id"`
	ToWarehouseName string                  `json:"to_warehouse_name"`
	TransferDate    time.Time               `json:"transfer_date"`
	Status          string                  `json:"status"`
	Remark          string                  `json:"remark"`
	CreatedAt       time.Time               `json:"created_at"`
	Items           []StockTransferItemResp `json:"items,omitempty"`
}
