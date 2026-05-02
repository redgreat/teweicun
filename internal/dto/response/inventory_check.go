/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

import "time"

type InventoryCheckItemResp struct {
	ID             int64   `json:"id"`
	MaterialID     int64   `json:"material_id"`
	MaterialCode   string  `json:"material_code"`
	MaterialName   string  `json:"material_name"`
	BookQuantity   float64 `json:"book_quantity"`
	ActualQuantity float64 `json:"actual_quantity"`
	DiffQuantity   float64 `json:"diff_quantity"`
	DiffReason     string  `json:"diff_reason"`
}

type InventoryCheckResp struct {
	ID            int64                    `json:"id"`
	CheckNo       string                   `json:"check_no"`
	WarehouseID   int64                    `json:"warehouse_id"`
	WarehouseName string                   `json:"warehouse_name"`
	CheckDate     time.Time                `json:"check_date"`
	Status        string                   `json:"status"`
	CheckerID     int64                    `json:"checker_id"`
	Remark        string                   `json:"remark"`
	CreatedAt     time.Time                `json:"created_at"`
	ApprovedAt    *time.Time               `json:"approved_at"`
	Items         []InventoryCheckItemResp `json:"items,omitempty"`
}
