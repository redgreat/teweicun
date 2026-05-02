/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type UserQuery struct {
	PageQuery
	Username   string `form:"username"`
	RealName   string `form:"real_name"`
	Department string `form:"department"`
	Status     string `form:"status"`
}

type CreateUserReq struct {
	Username   string `json:"username" binding:"required,min=3,max=50"`
	Password   string `json:"password" binding:"required,min=6,max=50"`
	RealName   string `json:"real_name" binding:"required,min=1,max=50"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Department string `json:"department"`
	RoleIDs    []int64 `json:"role_ids"`
}

type UpdateUserReq struct {
	RealName   string `json:"real_name" binding:"required,min=1,max=50"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Status     string `json:"status" binding:"oneof=enabled disabled"`
	RoleIDs    []int64 `json:"role_ids"`
}

type UpdatePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=50"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=50"`
}

type AssignRolesReq struct {
	RoleIDs []int64 `json:"role_ids" binding:"required"`
}
