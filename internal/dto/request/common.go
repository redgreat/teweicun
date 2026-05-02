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
