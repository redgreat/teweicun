/**
 * 功能：sales_order.go
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

func ListSalesOrders(c *gin.Context) {
	var q request.SalesOrderQuery
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

	list, total, err := service.ListSalesOrders(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func GetSalesOrderDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	detail, err := service.GetSalesOrderDetail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, detail)
}

func CreateSalesOrder(c *gin.Context) {
	var req request.CreateSalesOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	id, err := service.CreateSalesOrder(c.Request.Context(), &req, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"id": id})
}

func ConfirmSalesOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	err = service.ConfirmSalesOrder(c.Request.Context(), id, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func UpdateSalesOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	var req request.UpdateSalesOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	err = service.UpdateSalesOrder(c.Request.Context(), id, &req, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func CancelSalesOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	err = service.CancelSalesOrder(c.Request.Context(), id, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func ShipSalesOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	var req request.ShipSalesOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	stockOutID, err := service.ShipSalesOrder(c.Request.Context(), id, &req, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"stock_out_id": stockOutID})
}
