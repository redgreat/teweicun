/**
 * 功能：物料属性相关请求DTO
 * 创建时间：2026-04-17
 * 创建人：wangcw
 */

package request

type MaterialAttributeDefQuery struct {
	PageQuery
	AttrCode string `form:"attr_code"`
	AttrName string `form:"attr_name"`
	Status   string `form:"status"`
}

type CreateMaterialAttributeDefReq struct {
	AttrCode       string `json:"attr_code"`
	AttrName       string `json:"attr_name" binding:"required"`
	AttrType       string `json:"attr_type" binding:"required"`
	AttrUnit       string `json:"attr_unit"`
	SelectOptions  string `json:"select_options"`
	IsRequired     bool   `json:"is_required"`
	SortOrder      int    `json:"sort_order"`
	Remark         string `json:"remark"`
}

type UpdateMaterialAttributeDefReq struct {
	AttrName       string `json:"attr_name" binding:"required"`
	AttrType       string `json:"attr_type" binding:"required"`
	AttrUnit       string `json:"attr_unit"`
	SelectOptions  string `json:"select_options"`
	IsRequired     bool   `json:"is_required"`
	SortOrder      int    `json:"sort_order"`
	Remark         string `json:"remark"`
	Status         string `json:"status" binding:"required"`
}

type MaterialAttributeValueReq struct {
	AttrID    int64  `json:"attr_id" binding:"required"`
	AttrValue string `json:"attr_value" binding:"required"`
}

type UpdateMaterialAttributesReq struct {
	Attributes []MaterialAttributeValueReq `json:"attributes" binding:"required"`
}
