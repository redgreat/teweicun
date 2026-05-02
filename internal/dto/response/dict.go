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

type DictTypeResp struct {
	ID        int64             `json:"id"`
	DictType  string            `json:"dict_type"`
	DictName  string            `json:"dict_name"`
	Remark    database.NullString `json:"remark"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type DictDataResp struct {
	ID        int64             `json:"id"`
	DictType  string            `json:"dict_type"`
	DictLabel string            `json:"dict_label"`
	DictValue string            `json:"dict_value"`
	SortOrder int               `json:"sort_order"`
	Remark    database.NullString `json:"remark"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
