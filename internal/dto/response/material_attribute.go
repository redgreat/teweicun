/**
 * 功能：物料属性相关响应DTO
 * 创建时间：2026-04-17
 * 创建人：wangcw
 */

package response

import (
	"time"

	"github.com/redgreat/teweicun/pkg/database"
)

type MaterialAttributeDefResp struct {
	ID            int64             `json:"id"`
	AttrCode      string            `json:"attr_code"`
	AttrName      string            `json:"attr_name"`
	AttrType      string            `json:"attr_type"`
	AttrUnit      database.NullString `json:"attr_unit"`
	SelectOptions database.NullString `json:"select_options"`
	IsRequired    bool              `json:"is_required"`
	SortOrder     int               `json:"sort_order"`
	Remark        database.NullString `json:"remark"`
	Status        string            `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type MaterialAttributeValueResp struct {
	ID         int64             `json:"id"`
	MaterialID int64             `json:"material_id"`
	AttrID     int64             `json:"attr_id"`
	AttrCode   string            `json:"attr_code"`
	AttrName   string            `json:"attr_name"`
	AttrType   string            `json:"attr_type"`
	AttrUnit   database.NullString `json:"attr_unit"`
	AttrValue  string            `json:"attr_value"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}
