/**
 * 功能：role.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/service"
)

// ListRoles handles GET /system/roles
func ListRoles(c *gin.Context) {
	var q request.RoleQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}

	list, total, err := service.ListRoles(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

// GetRole handles GET /system/roles/:id
func GetRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid role ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	role, err := service.GetRole(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, role)
}

// CreateRole handles POST /system/roles
func CreateRole(c *gin.Context) {
	var req request.CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	id, err := service.CreateRole(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"id": id})
}

// UpdateRole handles PUT /system/roles/:id
func UpdateRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid role ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	var req request.UpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	err = service.UpdateRole(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteRole handles DELETE /system/roles/:id
func DeleteRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid role ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	err = service.DeleteRole(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// SetRolePermissions handles POST /system/roles/:id/permissions
func SetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid role ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	var req request.SetRolePermissionsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	err = service.SetRolePermissions(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// GetPermissionTree handles GET /system/permissions/tree
func GetPermissionTree(c *gin.Context) {
	tree, err := service.GetPermissionTree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, tree)
}
