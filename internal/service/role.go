/**
 * 功能：role.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
)

// ListRoles fetches roles with pagination
func ListRoles(ctx context.Context, q *request.RoleQuery) ([]response.RoleResp, int64, error) {
	return db.ListRoles(ctx, q)
}

// GetRole retrieves a single role by ID
func GetRole(ctx context.Context, id int64) (*response.RoleResp, error) {
	role, err := db.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errcode.NewAppError(errcode.ErrNotFound.Code, "Role not found", errcode.ErrNotFound.HTTPCode)
	}

	// Fetch permission IDs
	permIDs, err := db.GetRolePermissionIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	role.PermissionIDs = permIDs

	return role, nil
}

// CreateRole creates a new role
func CreateRole(ctx context.Context, req *request.CreateRoleReq) (int64, error) {
	return db.CreateRole(ctx, req.RoleCode, req.RoleName, req.Description)
}

// UpdateRole updates an existing role
func UpdateRole(ctx context.Context, id int64, req *request.UpdateRoleReq) error {
	existing, err := db.GetRoleByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errcode.NewAppError(errcode.ErrNotFound.Code, "Role not found", errcode.ErrNotFound.HTTPCode)
	}
	return db.UpdateRole(ctx, id, req.RoleName, req.Description, req.Status)
}

// DeleteRole soft deletes a role
func DeleteRole(ctx context.Context, id int64) error {
	existing, err := db.GetRoleByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errcode.NewAppError(errcode.ErrNotFound.Code, "Role not found", errcode.ErrNotFound.HTTPCode)
	}
	return db.DeleteRole(ctx, id)
}

// SetRolePermissions sets the permissions for a role
func SetRolePermissions(ctx context.Context, roleID int64, req *request.SetRolePermissionsReq) error {
	existing, err := db.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errcode.NewAppError(errcode.ErrNotFound.Code, "Role not found", errcode.ErrNotFound.HTTPCode)
	}
	return db.SetRolePermissions(ctx, roleID, req.PermissionIDs)
}

// GetPermissionTree returns the full permission tree
func GetPermissionTree(ctx context.Context) ([]response.PermissionResp, error) {
	return db.GetPermissionTree(ctx)
}
