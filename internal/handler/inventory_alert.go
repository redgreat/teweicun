/**
 * 功能：inventory_alert.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/service"
)

func ListInventoryAlerts(c *gin.Context) {
	var q request.InventoryAlertQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 10
	}

	list, total, err := service.ListInventoryAlerts(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func CheckInventoryAlerts(c *gin.Context) {
	err := service.CheckInventoryAlerts(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}
