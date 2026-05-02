/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type InventoryCheckQuery struct {
	PageQuery
	CheckNo string `form:"check_no"`
	Status  string `form:"status"`
}

type CreateCheckItem struct {
	MaterialID     int64   `json:"material_id" binding:"required"`
	BookQuantity   float64 `json:"book_quantity" binding:"required,min=0"`
	ActualQuantity float64 `json:"actual_quantity" binding:"required,min=0"`
}

type CreateInventoryCheckReq struct {
	WarehouseID int64             `json:"warehouse_id" binding:"required"`
	CheckDate   string            `json:"check_date" binding:"required"`
	Remark      string            `json:"remark"`
	Items       []CreateCheckItem `json:"items" binding:"required,dive,min=1"`
}
