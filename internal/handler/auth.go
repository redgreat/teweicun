/**
 * 功能：auth.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

// Package handler documentation
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/service"
)

// Login handles the user login request
func Login(c *gin.Context) {
	var req request.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	ip := c.ClientIP()
	resp, err := service.Login(c.Request.Context(), &req, ip)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
