/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type MaterialQuery struct {
	PageQuery
	MaterialCode string `form:"material_code"`
	MaterialName string `form:"material_name"`
	CategoryID   int64  `form:"category_id"`
	Unit         string `form:"unit"`
	Status       string `form:"status"`
}

type CreateMaterialReq struct {
	CategoryID       int64                 `json:"category_id" binding:"required"`
	MaterialCode     string                `json:"material_code"`
	MaterialName     string                `json:"material_name" binding:"required"`
	Unit             string                `json:"unit" binding:"required"`
	SafetyStock      float64               `json:"safety_stock"`
	MaxStock         float64               `json:"max_stock"`
	IsCode           bool                  `json:"is_code"`
	SkuManaged       bool                  `json:"sku_managed"`
	CustomAttributes []CustomAttributeItem `json:"custom_attributes"`
	Remark           string                `json:"remark"`
}

type UpdateMaterialReq struct {
	CategoryID       int64                 `json:"category_id" binding:"required"`
	MaterialCode     string                `json:"material_code"`
	MaterialName     string                `json:"material_name" binding:"required"`
	Unit             string                `json:"unit" binding:"required"`
	SafetyStock      float64               `json:"safety_stock"`
	MaxStock         float64               `json:"max_stock"`
	IsCode           bool                  `json:"is_code"`
	SkuManaged       bool                  `json:"sku_managed"`
	CustomAttributes []CustomAttributeItem `json:"custom_attributes"`
	Remark           string                `json:"remark"`
	Status           string                `json:"status" binding:"oneof=enabled disabled"`
}

type CustomAttributeItem struct {
	AttrName  string `json:"attr_name"`
	AttrValue string `json:"attr_value"`
}
