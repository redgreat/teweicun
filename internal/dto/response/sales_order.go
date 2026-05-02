/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

import (
	"time"

	"github.com/redgreat/teweicun/pkg/database"
)

type SalesOrderItemResp struct {
	ID              int64             `json:"id"`
	MaterialID      int64             `json:"material_id"`
	MaterialCode    string            `json:"material_code"`
	MaterialName    string            `json:"material_name"`
	Specification   string            `json:"specification"`
	Quantity        float64           `json:"quantity"`
	UnitPrice       float64           `json:"unit_price"`
	Amount          float64           `json:"amount"`
	ShippedQuantity float64           `json:"shipped_quantity"`
	Unit            string            `json:"unit"`
	Remark          database.NullString `json:"remark"`
}

type SalesOrderResp struct {
	ID                int64                `json:"id"`
	OrderNo           string               `json:"order_no"`
	CustomerCode      string               `json:"customer_code"`
	CustomerName      database.NullString  `json:"customer_name"`
	SalesPersonID     int64                `json:"sales_person_id"`
	SalesPersonName   database.NullString  `json:"sales_person_name"`
	OrderDate         time.Time            `json:"order_date"`
	DeliveryDate      *time.Time           `json:"delivery_date"`
	OrderStatus       string               `json:"order_status"`
	OrderStatusName   string               `json:"order_status_name"`
	TotalAmount       float64              `json:"total_amount"`
	Remark            database.NullString  `json:"remark"`
	CreatedAt         time.Time            `json:"created_at"`
	Items             []SalesOrderItemResp `json:"items,omitempty"`
}
