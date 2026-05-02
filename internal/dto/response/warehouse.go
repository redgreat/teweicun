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

type WarehouseResp struct {
	ID               int64             `json:"id"`
	WarehouseCode    string            `json:"warehouse_code"`
	WarehouseName    string            `json:"warehouse_name"`
	WarehouseType    string            `json:"warehouse_type"`
	WarehouseTypeName string           `json:"warehouse_type_name"`
	ManagerID        int64             `json:"manager_id"`
	ManagerName      database.NullString `json:"manager_name"`
	Status           string            `json:"status"`
	StatusName       string            `json:"status_name"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}
