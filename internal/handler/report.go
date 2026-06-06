/**
 * 功能：report.go
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

func ReportCustomerReconciliationSummary(c *gin.Context) {
	var q request.ReconciliationSummaryQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 50
	}

	list, total, err := service.ReportCustomerReconciliationSummary(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func ReportSupplierReconciliationSummary(c *gin.Context) {
	var q request.ReconciliationSummaryQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 50
	}

	list, total, err := service.ReportSupplierReconciliationSummary(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func ReportProfit(c *gin.Context) {
	var q request.ProfitReportQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	result, err := service.ReportProfit(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
