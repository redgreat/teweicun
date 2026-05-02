/**
 * 功能：report.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/service"
)

func ReportStockInSummary(c *gin.Context) {
	month := c.Query("month")
	result, err := service.ReportStockInSummary(c.Request.Context(), month)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func ReportStockOutSummary(c *gin.Context) {
	month := c.Query("month")
	result, err := service.ReportStockOutSummary(c.Request.Context(), month)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func ReportInventoryStatus(c *gin.Context) {
	result, err := service.ReportInventoryStatus(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
