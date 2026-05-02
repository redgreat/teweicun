/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type RoleQuery struct {
	PageQuery
	RoleCode string `form:"role_code"`
	RoleName string `form:"role_name"`
	Status   string `form:"status"`
}

type CreateRoleReq struct {
	RoleCode   string `json:"role_code" binding:"required,min=1,max=50"`
	RoleName   string `json:"role_name" binding:"required,min=1,max=100"`
	Description string `json:"description"`
}

type UpdateRoleReq struct {
	RoleName    string `json:"role_name" binding:"required,min=1,max=100"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"oneof=enabled disabled"`
}

type SetRolePermissionsReq struct {
	PermissionIDs []int64 `json:"permission_ids" binding:"required"`
}
