/**
 * 功能：notification.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/middleware"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/service"
)

func ListNotifications(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	list, err := service.ListNotifications(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, list)
}

func MarkNotificationRead(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	userID, _ := middleware.GetUserID(c)
	err = service.MarkNotificationRead(c.Request.Context(), id, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func ReportInventoryBalance(c *gin.Context) {
	start := c.Query("start_date")
	end := c.Query("end_date")
	if start == "" || end == "" {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	result, err := service.ReportInventoryBalance(c.Request.Context(), start, end)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func ReportInventoryTurnover(c *gin.Context) {
	start := c.Query("start_date")
	end := c.Query("end_date")
	if start == "" || end == "" {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	result, err := service.ReportInventoryTurnover(c.Request.Context(), start, end)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
