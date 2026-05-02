/**
 * 功能：logger.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

// Package middleware documentation
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"github.com/redgreat/teweicun/pkg/logger"
)

// RequestLogger is a gin middleware to log all requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		// 跳过健康检查接口的日志记录
		if path == "/api/v1/health" {
			return
		}

		cost := time.Since(start)

		if logger.Log != nil {
			logger.Log.Info(path,
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
				zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
				zap.Duration("cost", cost),
			)
		}
	}
}
