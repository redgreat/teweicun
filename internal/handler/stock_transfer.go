/**
 * 功能：stock_transfer.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
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

func ListStockTransfers(c *gin.Context) {
	var q request.StockTransferQuery
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

	list, total, err := service.ListStockTransfers(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func GetStockTransferDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	detail, err := service.GetStockTransferDetail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	if detail == nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.Success(c, detail)
}

func CreateStockTransfer(c *gin.Context) {
	var req request.CreateStockTransferReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	id, err := service.CreateStockTransfer(c.Request.Context(), &req, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"id": id})
}

func ConfirmTransferOut(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	err = service.ConfirmTransferOut(c.Request.Context(), id, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func ConfirmTransferIn(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	err = service.ConfirmTransferIn(c.Request.Context(), id, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}
