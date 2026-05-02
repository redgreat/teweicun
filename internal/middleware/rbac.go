/**
 * 功能：rbac.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/pkg/cache"
	"github.com/redgreat/teweicun/pkg/database"
)

const (
	UserPermsCacheKey = "twc:perms:%d"
	PermsCacheTTL     = 10 * time.Minute
)

// RequirePermission checks if the user has the required permission code
func RequirePermission(permCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		// 1. Check if user is admin (admin bypasses permission check)
		// We can get roles from context if we store them during AuthJWT, 
		// but let's assume we fetch or check a special flag.
		// For now, let's fetch perm codes.
		
		perms, err := GetUserPermissions(c.Request.Context(), userID)
		if err != nil {
			response.Error(c, errcode.ErrInternalServer)
			c.Abort()
			return
		}

		// admin role logic: if user has 'admin' in perms (or we check roles)
		// Let's check if permCode is in the list
		hasPerm := false
		for _, p := range perms {
			if p == permCode || p == "*" || p == "admin" {
				hasPerm = true
				break
			}
		}

		if !hasPerm {
			response.Error(c, errcode.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserPermissions retrieves permissions from cache or DB
func GetUserPermissions(ctx context.Context, userID int64) ([]string, error) {
	cacheKey := fmt.Sprintf(UserPermsCacheKey, userID)
	
	// Try Redis first
	if cache.Client != nil {
		perms, err := cache.Client.SMembers(ctx, cacheKey).Result()
		if err == nil && len(perms) > 0 {
			return perms, nil
		}
	}

	// Fetch from DB
	query := `SELECT perm_code FROM fn_get_user_permissions($1)`
	rows, err := database.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}

	// Cache in Redis
	if cache.Client != nil && len(perms) > 0 {
		cache.Client.SAdd(ctx, cacheKey, perms)
		cache.Client.Expire(ctx, cacheKey, PermsCacheTTL)
	}

	return perms, nil
}
