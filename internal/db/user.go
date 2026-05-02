/**
 * 功能：user.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

// Package db documentation
package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

type SysUserRow struct {
	ID           int64
	Username     string
	PasswordHash string
	RealName     string
	Phone        *string
	Email        *string
	Department   *string
	Status       string
	LastLoginAt  *time.Time
	LastLoginIP  *string
	CreatedBy    int64
	CreatedAt    time.Time
	UpdatedBy    *int64
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type UserListRow struct {
	ID         int64
	Username   string
	RealName   string
	Phone      *string
	Email      *string
	Department *string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(ctx context.Context, username string) (*SysUserRow, error) {
	query := `
		SELECT id, username, password_hash, real_name, phone, email, department,
		       status, last_login_at, last_login_ip, created_by, created_at,
		       updated_by, updated_at, deleted_at
		FROM sys_user
		WHERE username = $1 AND deleted_at IS NULL
	`

	var user SysUserRow
	err := database.Pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.RealName,
		&user.Phone,
		&user.Email,
		&user.Department,
		&user.Status,
		&user.LastLoginAt,
		&user.LastLoginIP,
		&user.CreatedBy,
		&user.CreatedAt,
		&user.UpdatedBy,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil, nil if not found
		}
		return nil, err
	}

	return &user, nil
}

// GetUserRoles retrieves the roles associated with a user
func GetUserRoles(ctx context.Context, userID int64) ([]string, error) {
	query := `
		SELECT r.role_code
		FROM sys_role r
		INNER JOIN sys_user_role ur ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.status = 'enabled' AND r.deleted_at IS NULL
	`

	rows, err := database.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

// UpdateUserLoginInfo updates the last login time and IP
func UpdateUserLoginInfo(ctx context.Context, userID int64, ip string) error {
	query := `
		UPDATE sys_user
		SET last_login_at = NOW(), last_login_ip = $1
		WHERE id = $2
	`
	_, err := database.Pool.Exec(ctx, query, ip, userID)
	return err
}

// GetUserByID retrieves a user by ID
func GetUserByID(ctx context.Context, id int64) (*SysUserRow, error) {
	query := `
		SELECT id, username, password_hash, real_name, phone, email, department,
		       status, last_login_at, last_login_ip, created_by, created_at,
		       updated_by, updated_at, deleted_at
		FROM sys_user
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user SysUserRow
	err := database.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.RealName,
		&user.Phone,
		&user.Email,
		&user.Department,
		&user.Status,
		&user.LastLoginAt,
		&user.LastLoginIP,
		&user.CreatedBy,
		&user.CreatedAt,
		&user.UpdatedBy,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// ListUsers fetches users with pagination and filters
func ListUsers(ctx context.Context, q *request.UserQuery) ([]response.UserResp, int64, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if q.Username != "" {
		where = append(where, fmt.Sprintf("username ILIKE $%d", argID))
		args = append(args, "%"+q.Username+"%")
		argID++
	}
	if q.RealName != "" {
		where = append(where, fmt.Sprintf("real_name ILIKE $%d", argID))
		args = append(args, "%"+q.RealName+"%")
		argID++
	}
	if q.Department != "" {
		where = append(where, fmt.Sprintf("department ILIKE $%d", argID))
		args = append(args, "%"+q.Department+"%")
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_user_list WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, username, real_name, phone, email,
		       department, status, status_name, last_login_at, created_at, updated_at,
		       role_names, role_ids
		FROM v_user_list
		WHERE %s
		ORDER BY id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.UserResp
	for rows.Next() {
		var item response.UserResp
		if err := rows.Scan(&item.ID, &item.Username, &item.RealName, &item.Phone, &item.Email,
			&item.Department, &item.Status, &item.StatusName, &item.LastLoginAt, &item.CreatedAt, &item.UpdatedAt,
			&item.RoleNames, &item.RoleIDs); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

// GetUserRoleDetails retrieves role IDs and names for a user
func GetUserRoleDetails(ctx context.Context, userID int64) ([]int64, []string, error) {
	query := `
		SELECT r.id, r.role_name
		FROM sys_role r
		INNER JOIN sys_user_role ur ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.id
	`
	rows, err := database.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var ids []int64
	var names []string
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, nil, err
		}
		ids = append(ids, id)
		names = append(names, name)
	}

	return ids, names, rows.Err()
}

// CreateUser inserts a new user with password hash
func CreateUser(ctx context.Context, username, passwordHash, realName, phone, email, department string, createdBy int64) (int64, error) {
	query := `
		INSERT INTO sys_user (username, password_hash, real_name, phone, email, department, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, 'enabled', $7)
		RETURNING id
	`
	var id int64
	err := database.Pool.QueryRow(ctx, query, username, passwordHash, realName, phone, email, department, createdBy).Scan(&id)
	return id, err
}

// UpdateUser updates an existing user
func UpdateUser(ctx context.Context, id int64, realName, phone, email, department, status string, updatedBy int64) error {
	query := `
		UPDATE sys_user
		SET real_name = $1, phone = $2, email = $3, department = $4, status = $5,
		    updated_by = $6, updated_at = NOW()
		WHERE id = $7 AND deleted_at IS NULL
	`
	_, err := database.Pool.Exec(ctx, query, realName, phone, email, department, status, updatedBy, id)
	return err
}

// DeleteUser soft deletes a user
func DeleteUser(ctx context.Context, id int64, updatedBy int64) error {
	query := `
		UPDATE sys_user
		SET deleted_at = NOW(), updated_by = $1
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err := database.Pool.Exec(ctx, query, updatedBy, id)
	return err
}

// UpdateUserPassword updates the user's password hash
func UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error {
	query := `
		UPDATE sys_user
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err := database.Pool.Exec(ctx, query, passwordHash, id)
	return err
}

// AssignUserRoles replaces all roles for a user
func AssignUserRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete existing roles
	_, err = tx.Exec(ctx, "DELETE FROM sys_user_role WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	// Insert new roles
	for _, roleID := range roleIDs {
		_, err = tx.Exec(ctx, "INSERT INTO sys_user_role (user_id, role_id) VALUES ($1, $2)", userID, roleID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
