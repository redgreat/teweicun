/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type CreateCategoryReq struct {
	ParentID     int64  `json:"parent_id"`
	CategoryCode string `json:"category_code"`
	CategoryName string `json:"category_name" binding:"required"`
	SortOrder    int    `json:"sort_order"`
}

type UpdateCategoryReq struct {
	ParentID     int64  `json:"parent_id"`
	CategoryCode string `json:"category_code" binding:"required"`
	CategoryName string `json:"category_name" binding:"required"`
	SortOrder    int    `json:"sort_order"`
	Status       string `json:"status" binding:"oneof=enabled disabled"`
}
