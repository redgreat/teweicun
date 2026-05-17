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
	ID                   int64                 `json:"id"`
	CategoryID           int64                 `json:"category_id"`
	CategoryName         database.NullString   `json:"category_name"`
	MaterialCode         string                `json:"material_code"`
	MaterialName         string                `json:"material_name"`
	MaterialNameBase     string                `json:"material_name_base"`
	MaterialDisplayName  string                `json:"material_display_name"`
	Unit                 string                `json:"unit"`
	UnitName             string                `json:"unit_name"`
	SafetyStock          float64               `json:"safety_stock"`
	MaxStock             float64               `json:"max_stock"`
	IsCode               bool                  `json:"is_code"`
	CustomAttributes     []CustomAttributeItem `json:"custom_attributes"`
	DefaultWarehouseID   database.NullString   `json:"default_warehouse_id"`
	DefaultWarehouseName database.NullString   `json:"default_warehouse_name"`
	Status               string                `json:"status"`
	StatusName           string                `json:"status_name"`
	Remark               database.NullString   `json:"remark"`
	CreatedAt            time.Time             `json:"created_at"`
	UpdatedAt            time.Time             `json:"updated_at"`
}

type CustomAttributeItem struct {
	AttrName  string `json:"attr_name"`
	AttrValue string `json:"attr_value"`
}
