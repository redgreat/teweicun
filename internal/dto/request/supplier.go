/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type SupplierQuery struct {
	PageQuery
	SupplierName string `form:"supplier_name"`
	SupplierCode string `form:"supplier_code"`
	SupplierType string `form:"supplier_type"`
	IsQualified  *bool  `form:"is_qualified"`
}

type CreateSupplierReq struct {
	SupplierCode        string `json:"supplier_code"`
	SupplierName        string `json:"supplier_name" binding:"required"`
	CreditCode          string `json:"credit_code"`
	SupplierType        string `json:"supplier_type" binding:"required"`
	ContactPerson       string `json:"contact_person" binding:"required"`
	ContactPhone        string `json:"contact_phone" binding:"required"`
	Address             string `json:"address"`
	IsQualified         bool   `json:"is_qualified"`
	QualificationExpire string `json:"qualification_expire"`
	BankName            string `json:"bank_name"`
	BankAccount         string `json:"bank_account"`
	Remark              string `json:"remark"`
}

type UpdateSupplierReq struct {
	SupplierName        string `json:"supplier_name" binding:"required"`
	CreditCode          string `json:"credit_code"`
	SupplierType        string `json:"supplier_type" binding:"required"`
	ContactPerson       string `json:"contact_person" binding:"required"`
	ContactPhone        string `json:"contact_phone" binding:"required"`
	Address             string `json:"address"`
	IsQualified         bool   `json:"is_qualified"`
	QualificationExpire string `json:"qualification_expire"`
	BankName            string `json:"bank_name"`
	BankAccount         string `json:"bank_account"`
	Remark              string `json:"remark"`
	Status              string `json:"status" binding:"oneof=enabled disabled blacklisted"`
}
