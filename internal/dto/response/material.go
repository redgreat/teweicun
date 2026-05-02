/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

import (
	"time"

	"github.com/redgreat/teweicun/pkg/database"
)

type MaterialResp struct {
	ID                   int64               `json:"id"`
	CategoryID           int64               `json:"category_id"`
	CategoryName         database.NullString `json:"category_name"`
	MaterialCode         string              `json:"material_code"`
	MaterialName         string              `json:"material_name"`
	Unit                 string              `json:"unit"`
	UnitName             string              `json:"unit_name"`
	SafetyStock          float64             `json:"safety_stock"`
	MaxStock             float64             `json:"max_stock"`
	IsCode               bool                `json:"is_code"`
	SkuManaged           bool                `json:"sku_managed"`
	CustomAttributes     []CustomAttributeItem `json:"custom_attributes"`
	DefaultWarehouseID   database.NullString `json:"default_warehouse_id"`
	DefaultWarehouseName database.NullString `json:"default_warehouse_name"`
	Status               string              `json:"status"`
	StatusName           string              `json:"status_name"`
	Remark               database.NullString `json:"remark"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
	SkuCount             int64               `json:"sku_count"`
}

type CustomAttributeItem struct {
	AttrID     int64  `json:"attr_id"`
	AttrCode   string `json:"attr_code"`
	AttrName   string `json:"attr_name"`
	AttrType   string `json:"attr_type"`
	AttrUnit   string `json:"attr_unit"`
	AttrValue  string `json:"attr_value"`
	IsRequired bool   `json:"is_required"`
}
