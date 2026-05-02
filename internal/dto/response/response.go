/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

// Package response documentation
package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Success returns a success response with data
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: errcode.Success.Code,
		Msg:  errcode.Success.Msg,
		Data: data,
	})
}

// SuccessWithMessage returns a success response with a custom message
func SuccessWithMessage(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: errcode.Success.Code,
		Msg:  msg,
		Data: data,
	})
}

// Error returns an error response
func Error(c *gin.Context, err error) {
	if appErr, ok := err.(*errcode.AppError); ok {
		c.JSON(appErr.HTTPCode, Response{
			Code: appErr.Code,
			Msg:  appErr.Msg,
		})
		return
	}

	// For unexpected errors, return generic internal server error and log it
	if err != nil {
		fmt.Printf("UNEXPECTED ERROR: %v\n", err)
	}

	// In production, we don't expose raw server errors to the client
	c.JSON(http.StatusInternalServerError, Response{
		Code: errcode.ErrInternalServer.Code,
		Msg:  errcode.ErrInternalServer.Msg,
	})
}

// CustomError returns a response with specific code and message
func CustomError(c *gin.Context, code int, msg string, httpCode int) {
	c.JSON(httpCode, Response{
		Code: code,
		Msg:  msg,
	})
}
