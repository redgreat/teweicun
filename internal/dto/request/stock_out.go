/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type CreateStockOutItem struct {
	MaterialID  int64   `json:"material_id" binding:"required"`
	InventoryID int64   `json:"inventory_id"` // Specific lot, optional if auto-fifo
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
}

type CreateStockOutReq struct {
	StockOutDate string               `json:"stock_out_date" binding:"required"`
	OutType      string               `json:"out_type" binding:"required"` // 'sales','production','transfer'
	RefDocType   string               `json:"ref_doc_type"`
	RefDocID     int64                `json:"ref_doc_id"`
	Receiver     string               `json:"receiver"`
	Remark       string               `json:"remark"`
	Items        []CreateStockOutItem `json:"items" binding:"required,min=1,dive"`
}
