/**
 * 功能：trace.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/service"
)

func TraceForward(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	result, err := service.TraceForward(c.Request.Context(), key)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func TraceBackward(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	result, err := service.TraceBackward(c.Request.Context(), keyword)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
