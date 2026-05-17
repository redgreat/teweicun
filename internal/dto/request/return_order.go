/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type ReturnOrderQuery struct {
	PageQuery
	ReturnNo        string `form:"return_no"`
	ReturnType      string `form:"return_type"` // purchase_return, sales_return
	SupplierCode    string `form:"supplier_code"`
	CustomerCode    string `form:"customer_code"`
	CustomerKeyword string `form:"customer_keyword"`
	Status          string `form:"status"`
	StartDate       string `form:"start_date"` // return_date >=
	EndDate         string `form:"end_date"`   // return_date <=
}

type CreateReturnOrderItem struct {
	InventoryID int64   `json:"inventory_id"`
	MaterialID  int64   `json:"material_id"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
}

type CreateReturnOrderReq struct {
	ReturnDate    string                  `json:"return_date" binding:"required"`
	ReturnType    string                  `json:"return_type" binding:"required"` // purchase_return, sales_return
	RefDocType    string                  `json:"ref_doc_type"`
	RefDocID      int64                   `json:"ref_doc_id"`
	WarehouseCode string                  `json:"warehouse_code"`
	SupplierCode  string                  `json:"supplier_code"`
	CustomerCode  string                  `json:"customer_code"`
	Remark        string                  `json:"remark"`
	Items         []CreateReturnOrderItem `json:"items" binding:"required,min=1,dive"`
}

// UpdateReturnOrderReq 更新采购退货单（支持草稿和待出库状态）
type UpdateReturnOrderReq struct {
	ReturnDate   string                  `json:"return_date" binding:"required"`
	SupplierCode string                  `json:"supplier_code" binding:"required"`
	Remark       string                  `json:"remark"`
	Items        []CreateReturnOrderItem `json:"items" binding:"required,min=1,dive"`
}

// UpdateSalesReturnOrderReq 更新销售退货单
type UpdateSalesReturnOrderReq struct {
	ReturnDate    string                  `json:"return_date" binding:"required"`
	CustomerCode  string                  `json:"customer_code" binding:"required"`
	WarehouseCode string                  `json:"warehouse_code" binding:"required"`
	Remark        string                  `json:"remark"`
	Items         []CreateReturnOrderItem `json:"items" binding:"required,min=1,dive"`
}
