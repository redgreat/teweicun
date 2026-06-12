/**
 * 功能：采购订单请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：CodeArts Agent
 */

package request

type PurchaseOrderQuery struct {
	PageQuery
	OrderNo         string `form:"order_no"`
	SupplierCode    string `form:"supplier_code"`
	SupplierKeyword string `form:"supplier_keyword"`
	OrderStatus     string `form:"order_status"`
	OrderType       string `form:"order_type"`
	StartDate       string `form:"start_date"`
	EndDate         string `form:"end_date"`
}

type CreatePurchaseOrderReq struct {
	OrderType    string                       `json:"order_type"`
	SupplierCode string                       `json:"supplier_code" binding:"required"`
	OrderDate    string                       `json:"order_date" binding:"required"`
	ExpectedDate string                       `json:"expected_date"`
	Remark       string                       `json:"remark"`
	Items        []CreatePurchaseOrderItemReq `json:"items" binding:"required,min=1"`
}

type CreatePurchaseOrderItemReq struct {
	MaterialID int64   `json:"material_id" binding:"required"`
	Quantity   float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice  float64 `json:"unit_price"`
}

type UpdatePurchaseOrderReq struct {
	ExpectedDate string                       `json:"expected_date"`
	Remark       string                       `json:"remark"`
	Items        []UpdatePurchaseOrderItemReq `json:"items"`
}

type UpdatePurchaseOrderItemReq struct {
	ID         int64   `json:"id" binding:"required"`
	MaterialID int64   `json:"material_id" binding:"required"`
	Quantity   float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice  float64 `json:"unit_price"`
}
