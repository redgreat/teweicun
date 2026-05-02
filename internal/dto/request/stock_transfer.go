/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type StockTransferQuery struct {
	PageQuery
	TransferNo string `form:"transfer_no"`
	Status     string `form:"status"`
}

type CreateStockTransferItem struct {
	MaterialID int64   `json:"material_id" binding:"required"`
	Quantity   float64 `json:"quantity" binding:"required,gt=0"`
	Remark     string  `json:"remark"`
}

type CreateStockTransferReq struct {
	FromWarehouseID int64                     `json:"from_warehouse_id" binding:"required"`
	ToWarehouseID   int64                     `json:"to_warehouse_id" binding:"required"`
	TransferDate    string                    `json:"transfer_date" binding:"required"`
	Remark          string                    `json:"remark"`
	Items           []CreateStockTransferItem `json:"items" binding:"required,dive,min=1"`
}
