/**
 * 功能：采购订单HTTP处理器
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

func ListPurchaseOrders(c *gin.Context) {
	var q request.PurchaseOrderQuery
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

	list, total, err := service.ListPurchaseOrders(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func GetPurchaseOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid order ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	order, err := service.GetPurchaseOrder(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	if order == nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.Success(c, order)
}

func CreatePurchaseOrder(c *gin.Context) {
	var req request.CreatePurchaseOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	id, err := service.CreatePurchaseOrder(c.Request.Context(), &req, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"id": id})
}

func UpdatePurchaseOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid order ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	var req request.UpdatePurchaseOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	err = service.UpdatePurchaseOrder(c.Request.Context(), id, &req, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func DeletePurchaseOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid order ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	err = service.DeletePurchaseOrder(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func ConfirmPurchaseOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid order ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	err = service.ConfirmPurchaseOrder(c.Request.Context(), id, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}
