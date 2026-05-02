/**
 * 功能：SKU请求DTO定义
 * 创建时间：2026-04-18
 * 修改时间：2026-04-19
 * 修改内容：统一使用CustomAttributeItem类型
 */

package request

// SKUQuery SKU查询参数
type SKUQuery struct {
	PageQuery
	MaterialID int64  `form:"material_id"`
	SKUCode    string `form:"sku_code"`
	SKUName    string `form:"sku_name"`
	Status     string `form:"status"`
}

// CreateSKUReq 创建SKU请求
type CreateSKUReq struct {
	MaterialID       int64                 `json:"material_id" binding:"required"`
	SKUName          string                `json:"sku_name"`
	ReferencePrice   float64               `json:"reference_price"`
	CustomAttributes []CustomAttributeItem `json:"custom_attributes" binding:"required,min=1"`
	Remark           string                `json:"remark"`
}

// UpdateSKUReq 更新SKU请求
type UpdateSKUReq struct {
	SKUName          string                `json:"sku_name"`
	ReferencePrice   float64               `json:"reference_price"`
	CustomAttributes []CustomAttributeItem `json:"custom_attributes"`
	Remark           string                `json:"remark"`
	Status           string                `json:"status" binding:"oneof=enabled disabled"`
}
