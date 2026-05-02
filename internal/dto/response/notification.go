/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

import "time"

type InventoryBalanceReport struct {
	TransDate   string  `json:"trans_date"`
	WarehouseID int64   `json:"warehouse_id"`
	MaterialID  int64   `json:"material_id"`
	InQty       float64 `json:"in_qty"`
	OutQty      float64 `json:"out_qty"`
}

type InventoryTurnoverReport struct {
	MaterialID   int64   `json:"material_id"`
	MaterialCode string  `json:"material_code"`
	MaterialName string  `json:"material_name"`
	OutTotal     float64 `json:"out_total"`
	AvgInventory float64 `json:"avg_inventory"`
	TurnoverRate float64 `json:"turnover_rate"`
}

type NotificationResp struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	NotifyType string    `json:"notify_type"`
	RefType    string    `json:"ref_type"`
	RefID      int64     `json:"ref_id"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
	ReadAt     *time.Time `json:"read_at"`
}
