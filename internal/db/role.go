/**
 * 功能：role.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

// ListRoles fetches roles with pagination and filters
func ListRoles(ctx context.Context, q *request.RoleQuery) ([]response.RoleResp, int64, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if q.RoleCode != "" {
		where = append(where, fmt.Sprintf("role_code ILIKE $%d", argID))
		args = append(args, "%"+q.RoleCode+"%")
		argID++
	}
	if q.RoleName != "" {
		where = append(where, fmt.Sprintf("role_name ILIKE $%d", argID))
		args = append(args, "%"+q.RoleName+"%")
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_role_list WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, role_code, role_name, description, status, status_name,
		       user_count, created_at, updated_at
		FROM v_role_list
		WHERE %s
		ORDER BY id ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.RoleResp
	for rows.Next() {
		var item response.RoleResp
		if err := rows.Scan(&item.ID, &item.RoleCode, &item.RoleName, &item.Description,
			&item.Status, &item.StatusName, &item.UserCount, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Fetch permission IDs for each role
	for i := range result {
		permIDs, err := GetRolePermissionIDs(ctx, result[i].ID)
		if err != nil {
			return nil, 0, err
		}
		result[i].PermissionIDs = permIDs
	}

	return result, total, nil
}

// GetRoleByID retrieves a role by ID
func GetRoleByID(ctx context.Context, id int64) (*response.RoleResp, error) {
	query := `
		SELECT r.id, r.role_code, r.role_name, r.description, r.status, r.created_at, r.updated_at,
		       (SELECT count(*) FROM sys_user_role ur WHERE ur.role_id = r.id) AS user_count
		FROM sys_role r
		WHERE r.id = $1 AND r.deleted_at IS NULL
	`
	var item response.RoleResp
	err := database.Pool.QueryRow(ctx, query, id).Scan(
		&item.ID, &item.RoleCode, &item.RoleName, &item.Description,
		&item.Status, &item.CreatedAt, &item.UpdatedAt, &item.UserCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// GetRolePermissionIDs retrieves permission IDs for a role
func GetRolePermissionIDs(ctx context.Context, roleID int64) ([]int64, error) {
	query := `
		SELECT permission_id
		FROM sys_role_permission
		WHERE role_id = $1
		ORDER BY permission_id
	`
	rows, err := database.Pool.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// CreateRole inserts a new role
func CreateRole(ctx context.Context, roleCode, roleName, description string) (int64, error) {
	query := `
		INSERT INTO sys_role (role_code, role_name, description, status)
		VALUES ($1, $2, $3, 'enabled')
		RETURNING id
	`
	var id int64
	err := database.Pool.QueryRow(ctx, query, roleCode, roleName, description).Scan(&id)
	return id, err
}

// UpdateRole updates an existing role
func UpdateRole(ctx context.Context, id int64, roleName, description, status string) error {
	query := `
		UPDATE sys_role
		SET role_name = $1, description = $2, status = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
	`
	_, err := database.Pool.Exec(ctx, query, roleName, description, status, id)
	return err
}

// DeleteRole soft deletes a role
func DeleteRole(ctx context.Context, id int64) error {
	query := `
		UPDATE sys_role
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := database.Pool.Exec(ctx, query, id)
	return err
}

// SetRolePermissions replaces all permissions for a role
func SetRolePermissions(ctx context.Context, roleID int64, permissionIDs []int64) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete existing permissions
	_, err = tx.Exec(ctx, "DELETE FROM sys_role_permission WHERE role_id = $1", roleID)
	if err != nil {
		return err
	}

	// Insert new permissions
	for _, permID := range permissionIDs {
		_, err = tx.Exec(ctx, "INSERT INTO sys_role_permission (role_id, permission_id) VALUES ($1, $2)", roleID, permID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// ListPermissions fetches all permissions as a flat list
func ListPermissions(ctx context.Context) ([]response.PermissionResp, error) {
	query := `
		SELECT id, parent_id, perm_code, perm_name, perm_type,
		       path, icon, sort_order, status
		FROM sys_permission
		ORDER BY sort_order ASC, id ASC
	`
	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []response.PermissionResp
	for rows.Next() {
		var item response.PermissionResp
		if err := rows.Scan(&item.ID, &item.ParentID, &item.PermCode, &item.PermName,
			&item.PermType, &item.Path, &item.Icon, &item.SortOrder, &item.Status); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// GetPermissionTree builds the permission tree from flat list
func GetPermissionTree(ctx context.Context) ([]response.PermissionResp, error) {
	flatList, err := ListPermissions(ctx)
	if err != nil {
		return nil, err
	}

	// Build a map for quick lookup
	nodeMap := make(map[int64]*response.PermissionResp)
	for i := range flatList {
		nodeMap[flatList[i].ID] = &flatList[i]
	}

	// Build tree
	var tree []response.PermissionResp
	for i := range flatList {
		if flatList[i].ParentID == 0 {
			// Root node, build children recursively
			buildPermissionChildren(&flatList[i], nodeMap)
			tree = append(tree, flatList[i])
		}
	}

	return tree, nil
}

// buildPermissionChildren recursively builds the children of a permission node
func buildPermissionChildren(node *response.PermissionResp, nodeMap map[int64]*response.PermissionResp) {
	for id, child := range nodeMap {
		if child.ParentID == node.ID {
			buildPermissionChildren(child, nodeMap)
			node.Children = append(node.Children, *child)
			delete(nodeMap, id) // Remove to avoid re-processing
		}
	}
}
