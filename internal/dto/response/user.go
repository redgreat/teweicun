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

type UserResp struct {
	ID         int64             `json:"id"`
	Username   string            `json:"username"`
	RealName   string            `json:"real_name"`
	Phone      database.NullString `json:"phone"`
	Email      database.NullString `json:"email"`
	Department database.NullString `json:"department"`
	Status     string            `json:"status"`
	StatusName string            `json:"status_name"`
	RoleIDs    []int64           `json:"role_ids"`
	RoleNames  []string          `json:"role_names"`
	LastLoginAt *time.Time       `json:"last_login_at"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}
