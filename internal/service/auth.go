/**
 * 功能：auth.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

// Package service documentation
package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

// Login authenticates a user and returns a JWT token
func Login(ctx context.Context, req *request.LoginReq, ip string) (*response.LoginResp, error) {
	// 1. Get user from DB
	user, err := db.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errcode.ErrUserNotFound
	}

	// 2. Check status
	if user.Status != "enabled" {
		return nil, errcode.NewAppError(errcode.ErrForbidden.Code, "User account is disabled", errcode.ErrForbidden.HTTPCode)
	}

	// 3. Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errcode.ErrInvalidPassword
	}

	// 4. Generate token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	// 5. Get roles
	roles, err := db.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// 6. Update login info asynchronously
	go func() {
		// Use a detached context for the async task since the request ctx will be cancelled
		_ = db.UpdateUserLoginInfo(context.Background(), user.ID, ip)
	}()

	return &response.LoginResp{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
		RealName: user.RealName,
		Roles:    roles,
	}, nil
}
