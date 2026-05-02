/**
 * 功能：recovery.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/pkg/logger"
	"go.uber.org/zap"
)

// CustomRecovery is a custom recovery middleware to handle panics gracefully
// and return a strictly typed JSON using standard response format instead of raw string.
func CustomRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Record the panic stack trace to log
				if logger.Log != nil {
					logger.Log.Error("Panic Recovered",
						zap.Any("error", err),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					fmt.Printf("Panic Recovered: %v\n%s\n", err, debug.Stack())
				}

				// Abort and return standard 500 error response
				c.Abort()
				response.CustomError(c, errcode.ErrInternalServer.Code, "System error, please try again later.", http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
