/**
 * 功能：入库单HTTP处理器
 * 创建时间：2026-04-18
 * 创建人：CodeArts Agent
 */

package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/middleware"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/service"
)

func ListStockIns(c *gin.Context) {
	var q request.StockInQuery
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

	list, total, err := service.ListStockIns(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func GetStockIn(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid stock in ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	stockIn, err := service.GetStockIn(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	if stockIn == nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.Success(c, stockIn)
}

func CreateStockIn(c *gin.Context) {
	var req request.CreateStockInReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	id, err := service.CreateStockIn(c.Request.Context(), &req, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"id": id})
}

func ConfirmStockIn(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid stock in ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	err = service.ConfirmStockIn(c.Request.Context(), id, userID)
	if err != nil {
		// 将存储过程/业务错误回传给前端展示
		if _, ok := err.(*errcode.AppError); !ok {
			response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
			return
		}
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// ConfirmReversalStockIn 退料入库：基于“备货编码”确认入库
func ConfirmReversalStockIn(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid stock in ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := service.ConfirmReversalStockIn(c.Request.Context(), id, userID); err != nil {
		if _, ok := err.(*errcode.AppError); !ok {
			response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
			return
		}
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func ListStockInConfirmLogs(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid stock in ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	logs, err := service.ListStockInConfirmLogs(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, logs)
}

func UpdateStockIn(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid stock in ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	var req request.UpdateStockInReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	err = service.UpdateStockIn(c.Request.Context(), id, req.WarehouseCode, req.Remark, req.Items)
	if err != nil {
		// 将业务校验错误回传给前端展示
		if _, ok := err.(*errcode.AppError); !ok {
			response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
			return
		}
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}
