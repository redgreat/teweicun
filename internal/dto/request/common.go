/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type PageQuery struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"page_size" binding:"required,min=1,max=100"`
}

// Offset calculates the standard SQL offset
func (q *PageQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}

type ReconciliationSummaryQuery struct {
	PageQuery
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	Keyword    string `form:"keyword"`
	CustomerID int64  `form:"customer_id"`
	SupplierID int64  `form:"supplier_id"`
}

type ProfitReportQuery struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

type PartnerDropdownQuery struct {
	Type    string `form:"type" binding:"required,oneof=customer supplier"`
	Keyword string `form:"keyword"`
	Status  string `form:"status"`
	Limit   int    `form:"limit"`
}
