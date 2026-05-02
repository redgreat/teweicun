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

type CustomerResp struct {
	ID              int64             `json:"id"`
	CustomerCode    string            `json:"customer_code"`
	CustomerName    string            `json:"customer_name"`
	CreditCode      database.NullString `json:"credit_code"`
	CustomerType    database.NullString `json:"customer_type"`
	CustomerTypeName string           `json:"customer_type_name"`
	ContactPerson   string            `json:"contact_person"`
	ContactPhone    string            `json:"contact_phone"`
	Address         database.NullString `json:"address"`
	SalesPersonName database.NullString `json:"sales_person_name"`
	Remark          database.NullString `json:"remark"`
	Status          string            `json:"status"`
	StatusName      string            `json:"status_name"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}
