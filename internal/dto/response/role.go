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

type RoleResp struct {
	ID            int64             `json:"id"`
	RoleCode      string            `json:"role_code"`
	RoleName      string            `json:"role_name"`
	Description   database.NullString `json:"description"`
	Status        string            `json:"status"`
	StatusName    string            `json:"status_name"`
	PermissionIDs []int64          `json:"permission_ids"`
	UserCount     int               `json:"user_count"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type PermissionResp struct {
	ID        int64             `json:"id"`
	ParentID  int64             `json:"parent_id"`
	PermCode  string            `json:"perm_code"`
	PermName  string            `json:"perm_name"`
	PermType  string            `json:"perm_type"`
	Path      database.NullString `json:"path"`
	Icon      database.NullString `json:"icon"`
	SortOrder int               `json:"sort_order"`
	Status    string            `json:"status"`
	Children  []PermissionResp  `json:"children,omitempty"`
}
