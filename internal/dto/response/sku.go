/**
 * 功能：SKU响应DTO定义
 * 创建时间：2026-04-18
 * 修改时间：2026-04-19
 * 修改内容：增加物料规格、分类、属性摘要等字段；新增SKUSelectItem
 */

package response

import "time"

// SKUListItem SKU列表项
type SKUListItem struct {
	ID               int64                 `json:"id"`
	MaterialID       int64                 `json:"material_id"`
	MaterialCode     string                `json:"material_code"`
	MaterialName     string                `json:"material_name"`
	Unit             string                `json:"unit"`
	ReferencePrice   float64               `json:"reference_price"`
	CategoryName     string                `json:"category_name"`
	SKUCode          string                `json:"sku_code"`
	SKUName          string                `json:"sku_name"`
	CustomAttributes []CustomAttributeItem `json:"custom_attributes"`
	AttrSummary      *string               `json:"attr_summary"`
	Status           string                `json:"status"`
	StatusName       string                `json:"status_name"`
	Remark           *string               `json:"remark"`
	CreatedAt        time.Time             `json:"created_at"`
}

// SKUDetail SKU详情
type SKUDetail struct {
	ID               int64                 `json:"id"`
	MaterialID       int64                 `json:"material_id"`
	MaterialCode     string                `json:"material_code"`
	MaterialName     string                `json:"material_name"`
	Unit             string                `json:"unit"`
	ReferencePrice   float64               `json:"reference_price"`
	CategoryName     string                `json:"category_name"`
	SKUCode          string                `json:"sku_code"`
	SKUName          string                `json:"sku_name"`
	CustomAttributes []CustomAttributeItem `json:"custom_attributes"`
	AttrSummary      *string               `json:"attr_summary"`
	Status           string                `json:"status"`
	StatusName       string                `json:"status_name"`
	Remark           *string               `json:"remark"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
}

// SKUSelectItem 采购/入库时选择SKU的简要信息
type SKUSelectItem struct {
	ID               int64                 `json:"id"`
	SKUCode          string                `json:"sku_code"`
	SKUName          string                `json:"sku_name"`
	ReferencePrice   float64               `json:"reference_price"`
	CustomAttributes []CustomAttributeItem `json:"custom_attributes"`
	AttrSummary      *string               `json:"attr_summary"`
}
