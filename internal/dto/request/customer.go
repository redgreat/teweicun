/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type CustomerQuery struct {
	PageQuery
	CustomerName string `form:"customer_name"`
	CustomerCode string `form:"customer_code"`
	Status       string `form:"status"`
}

type CreateCustomerReq struct {
	CustomerCode  string `json:"customer_code"`
	CustomerName  string `json:"customer_name" binding:"required"`
	CreditCode    string `json:"credit_code"`
	ContactPerson string `json:"contact_person" binding:"required"`
	ContactPhone  string `json:"contact_phone" binding:"required"`
	Address       string `json:"address"`
	Remark        string `json:"remark"`
}

type UpdateCustomerReq struct {
	CustomerName  string `json:"customer_name" binding:"required"`
	CreditCode    string `json:"credit_code"`
	ContactPerson string `json:"contact_person" binding:"required"`
	ContactPhone  string `json:"contact_phone" binding:"required"`
	Address       string `json:"address"`
	Remark        string `json:"remark"`
	Status        string `json:"status" binding:"oneof=enabled disabled blacklisted"`
}
