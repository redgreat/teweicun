/**
 * 功能：return_order.go
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

func ListReturnOrders(c *gin.Context) {
	var q request.ReturnOrderQuery
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

	list, total, err := service.ListReturnOrders(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func GetReturnOrderDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	detail, err := service.GetReturnOrderDetail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, detail)
}

func CreateReturnOrder(c *gin.Context) {
	var req request.CreateReturnOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	id, err := service.CreateReturnOrder(c.Request.Context(), &req, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"id": id})
}

func UpdateReturnOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	var req request.UpdateReturnOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	err = service.UpdateReturnOrder(c.Request.Context(), id, &req, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func ConfirmReturnOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	err = service.ConfirmReturnOrder(c.Request.Context(), id, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func DeleteReturnOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	err = service.DeleteReturnOrder(c.Request.Context(), id, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}
