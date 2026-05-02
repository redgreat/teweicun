/**
 * 功能：auth.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/pkg/utils"
)

// AuthJWT validates the JWT token provided in the Authorization header.
func AuthJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Error(c, errcode.NewAppError(errcode.ErrUnauthorized.Code, "Authorization header format must be Bearer {token}", errcode.ErrUnauthorized.HTTPCode))
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			response.Error(c, errcode.NewAppError(errcode.ErrUnauthorized.Code, "Invalid or expired token", errcode.ErrUnauthorized.HTTPCode))
			c.Abort()
			return
		}

		// Store user info in context for later use
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		
		c.Next()
	}
}

// GetUserID retrieves the user_id from the Gin context safely
func GetUserID(c *gin.Context) (int64, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := val.(int64)
	return id, ok
}
