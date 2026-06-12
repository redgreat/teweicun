/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

import "time"

type InventoryDetailResp struct {
	ID             int64      `json:"id"`
	MaterialID     int64      `json:"material_id"`
	MaterialCode   string     `json:"material_code"`
	MaterialName   string     `json:"material_name"`
	Unit           string     `json:"unit"`
	WarehouseID    int64      `json:"warehouse_id"`
	WarehouseName  string     `json:"warehouse_name"`
	Quantity       float64    `json:"quantity"`
	LockedQuantity float64    `json:"locked_quantity"`
	Available      float64    `json:"available"`
	StockInDate    *time.Time `json:"stock_in_date"`
	CertificateID  *int64     `json:"certificate_id"`
	CertificateNo  string     `json:"certificate_no"`
}

type InventorySummaryResp struct {
	MaterialID     int64   `json:"material_id"`
	MaterialCode   string  `json:"material_code"`
	MaterialName   string  `json:"material_name"`
	Unit           string  `json:"unit"`
	TotalQuantity  float64 `json:"total_quantity"`
	LockedQuantity float64 `json:"locked_quantity"`
	Available      float64 `json:"available"`
}

type InventoryAvailableResp struct {
	InventoryID       int64   `json:"inventory_id"`
	MaterialID        int64   `json:"material_id"`
	MaterialCode      string  `json:"material_code"`
	MaterialName      string  `json:"material_name"`
	IsCode            bool    `json:"is_code"`
	WarehouseID       int64   `json:"warehouse_id"`
	WarehouseCode     string  `json:"warehouse_code"`
	WarehouseName     string  `json:"warehouse_name"`
	Unit              string  `json:"unit"`
	UnitCost          float64 `json:"unit_cost"`
	Quantity          float64 `json:"quantity"`
	LockedQuantity    float64 `json:"locked_quantity"`
	InTransitQuantity float64 `json:"in_transit_quantity"`
	AvailableQuantity float64 `json:"available_quantity"`
}

// InventoryIssuedResp 查询“已出库(已领用)”可退回库存（按库存批次聚合）
// 有编码：issued_quantity 为仍处 issued 状态的序列件数（已含序列状态语义）
// 无编码：issued_quantity 为领料出库累计 − 已完成退料累计（净可退），与 available_quantity 取小作为实际上限
type InventoryIssuedResp struct {
	InventoryID       int64   `json:"inventory_id"`
	MaterialID        int64   `json:"material_id"`
	MaterialCode      string  `json:"material_code"`
	MaterialName      string  `json:"material_name"`
	IsCode            bool    `json:"is_code"`
	WarehouseID       int64   `json:"warehouse_id"`
	WarehouseCode     string  `json:"warehouse_code"`
	WarehouseName     string  `json:"warehouse_name"`
	Unit              string  `json:"unit"`
	UnitCost          float64 `json:"unit_cost"`
	IssuedQuantity    float64 `json:"issued_quantity"`
	AvailableQuantity float64 `json:"available_quantity"`
}

type InventoryMaterialLedgerResp struct {
	MaterialID             int64   `json:"material_id"`
	MaterialName           string  `json:"material_name"`
	WarehouseID            int64   `json:"warehouse_id"`
	WarehouseName          string  `json:"warehouse_name"`
	IsCode                 bool    `json:"is_code"`
	BookQuantity           float64 `json:"book_quantity"`
	Quantity               float64 `json:"quantity"`
	UnitCost               float64 `json:"unit_cost"`
	TotalAmount            float64 `json:"total_amount"`
	LockedQuantity         float64 `json:"locked_quantity"`
	InTransitQuantity      float64 `json:"in_transit_quantity"`
	SerialReservedQuantity float64 `json:"serial_reserved_quantity"`
	InventoryCount         int64   `json:"inventory_count"`
	HasCustomAttrs         bool    `json:"has_custom_attrs"`
}

type InventoryMaterialLedgerStatsResp struct {
	TotalAmount       float64 `json:"total_amount"`
	CodeTotalAmount   float64 `json:"code_total_amount"`
	NoCodeTotalAmount float64 `json:"no_code_total_amount"`
	TotalLockedQty    float64 `json:"total_locked_qty"`
}

type InventoryMaterialLedgerSerialResp struct {
	SerialCode        string `json:"serial_code"`
	Status            string `json:"status"`
	StatusName        string `json:"status_name"`
	DisplayStatus     string `json:"display_status"`
	DisplayStatusName string `json:"display_status_name"`
}
