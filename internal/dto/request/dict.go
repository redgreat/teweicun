/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type DictTypeQuery struct {
	PageQuery
	DictName string `form:"dict_name"`
	DictType string `form:"dict_type"`
}

type CreateDictTypeReq struct {
	DictType string `json:"dict_type" binding:"required"`
	DictName string `json:"dict_name" binding:"required"`
	Remark   string `json:"remark"`
}

type UpdateDictTypeReq struct {
	DictType string `json:"dict_type" binding:"required"`
	DictName string `json:"dict_name" binding:"required"`
	Remark   string `json:"remark"`
}

type DictDataQuery struct {
	PageQuery
	DictType  string `form:"dict_type" binding:"required"`
	DictLabel string `form:"dict_label"`
}

type CreateDictDataReq struct {
	DictType  string `json:"dict_type" binding:"required"`
	DictLabel string `json:"dict_label" binding:"required"`
	DictValue string `json:"dict_value" binding:"required"`
	SortOrder int    `json:"sort_order"`
	Remark    string `json:"remark"`
}

type UpdateDictDataReq struct {
	DictLabel string `json:"dict_label" binding:"required"`
	DictValue string `json:"dict_value" binding:"required"`
	SortOrder int    `json:"sort_order"`
	Remark    string `json:"remark"`
}
