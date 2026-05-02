/**
 * 功能：user.go
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
	"golang.org/x/crypto/bcrypt"
)

// ListUsers fetches users with pagination
func ListUsers(ctx context.Context, q *request.UserQuery) ([]response.UserResp, int64, error) {
	return db.ListUsers(ctx, q)
}

// CreateUser creates a new user with hashed password and assigns roles
func CreateUser(ctx context.Context, req *request.CreateUserReq, operatorID int64) (int64, error) {
	// Check if username already exists
	existing, err := db.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return 0, errcode.NewAppError(errcode.ErrDataConcurrencyConflict.Code, "Username already exists", errcode.ErrDataConcurrencyConflict.HTTPCode)
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// Create user
	id, err := db.CreateUser(ctx, req.Username, string(hash), req.RealName, req.Phone, req.Email, req.Department, operatorID)
	if err != nil {
		return 0, err
	}

	// Assign roles if provided
	if len(req.RoleIDs) > 0 {
		if err := db.AssignUserRoles(ctx, id, req.RoleIDs); err != nil {
			return 0, err
		}
	}

	return id, nil
}

// UpdateUser updates an existing user and reassigns roles
func UpdateUser(ctx context.Context, id int64, req *request.UpdateUserReq, operatorID int64) error {
	// Check user exists
	existing, err := db.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errcode.ErrUserNotFound
	}

	// Update user info
	err = db.UpdateUser(ctx, id, req.RealName, req.Phone, req.Email, req.Department, req.Status, operatorID)
	if err != nil {
		return err
	}

	// Reassign roles if provided
	if req.RoleIDs != nil {
		if err := db.AssignUserRoles(ctx, id, req.RoleIDs); err != nil {
			return err
		}
	}

	return nil
}

// DeleteUser soft deletes a user
func DeleteUser(ctx context.Context, id int64, operatorID int64) error {
	existing, err := db.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errcode.ErrUserNotFound
	}
	return db.DeleteUser(ctx, id, operatorID)
}

// UpdatePassword updates the user's password
func UpdatePassword(ctx context.Context, id int64, oldPassword, newPassword string) error {
	user, err := db.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errcode.ErrUserNotFound
	}

	// Verify old password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return errcode.ErrInvalidPassword
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return db.UpdateUserPassword(ctx, id, string(hash))
}

// AssignRoles assigns roles to a user
func AssignRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	existing, err := db.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errcode.ErrUserNotFound
	}
	return db.AssignUserRoles(ctx, userID, roleIDs)
}
