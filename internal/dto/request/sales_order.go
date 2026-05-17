/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type SalesOrderQuery struct {
	PageQuery
	OrderNo         string `form:"order_no"`
	CustomerCode    string `form:"customer_code"`
	CustomerKeyword string `form:"customer_keyword"`
	Status          string `form:"status"`
	StartDate       string `form:"start_date"`
	EndDate         string `form:"end_date"`
}

type CreateSalesOrderItem struct {
	MaterialID int64   `json:"material_id" binding:"required"`
	Quantity   float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice  float64 `json:"unit_price" binding:"required,gt=0"`
	Remark     string  `json:"remark"`
}

type CreateSalesOrderReq struct {
	CustomerCode    string                 `json:"customer_code" binding:"required"`
	SalesPersonID   int64                  `json:"sales_person_id"`
	OrderDate       string                 `json:"order_date" binding:"required"`
	DeliveryDate    string                 `json:"delivery_date"`
	ContractNo      string                 `json:"contract_no"`
	PaymentMethod   string                 `json:"payment_method"`
	ReceiverName    string                 `json:"receiver_name"`
	ReceiverPhone   string                 `json:"receiver_phone"`
	ReceiverAddress string                 `json:"receiver_address"`
	Remark          string                 `json:"remark"`
	Items           []CreateSalesOrderItem `json:"items" binding:"required,min=1,dive"`
}

type UpdateSalesOrderItem struct {
	MaterialID int64   `json:"material_id" binding:"required"`
	Quantity   float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice  float64 `json:"unit_price" binding:"required,gt=0"`
	Remark     string  `json:"remark"`
}

type UpdateSalesOrderReq struct {
	CustomerCode    string                 `json:"customer_code" binding:"required"`
	SalesPersonID   int64                  `json:"sales_person_id"`
	OrderDate       string                 `json:"order_date" binding:"required"`
	DeliveryDate    string                 `json:"delivery_date"`
	ContractNo      string                 `json:"contract_no"`
	PaymentMethod   string                 `json:"payment_method"`
	ReceiverName    string                 `json:"receiver_name"`
	ReceiverPhone   string                 `json:"receiver_phone"`
	ReceiverAddress string                 `json:"receiver_address"`
	Remark          string                 `json:"remark"`
	Items           []UpdateSalesOrderItem `json:"items" binding:"required,min=1,dive"`
}

type ShipSalesOrderReq struct {
	StockOutDate string `json:"stock_out_date"`
	Remark       string `json:"remark"`
}
