/**
 * 功能：生产单/生产退货单 HTTP Handler
 * 创建时间：2026-06-06
 * 创建人：GPT-5.2
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

func ListProductionOrders(c *gin.Context) {
	var q request.ProductionOrderQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if _, ok := middleware.GetUserID(c); !ok {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	out, err := service.ListProductionOrders(c.Request.Context(), q)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}
	response.Success(c, out)
}

func GetProductionOrderDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if _, ok := middleware.GetUserID(c); !ok {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	out, err := service.GetProductionOrderDetail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}
	response.Success(c, out)
}

// UpdateProductionOrder 更新生产单（成本价格、备注）
func UpdateProductionOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	var req request.ProductionOrderUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	if err := service.UpdateProductionOrder(c.Request.Context(), id, req, userID); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}
	response.Success(c, nil)
}

// ListConsumptionOrdersByProduction 查询生产单关联的领料单
func ListConsumptionOrdersByProduction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if _, ok := middleware.GetUserID(c); !ok {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	out, err := service.ListConsumptionOrdersByProduction(c.Request.Context(), id)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}
	response.Success(c, out)
}

// ListReversalOrdersByProduction 查询生产单关联的退料单
func ListReversalOrdersByProduction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if _, ok := middleware.GetUserID(c); !ok {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	out, err := service.ListReversalOrdersByProduction(c.Request.Context(), id)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}
	response.Success(c, out)
}

func ListProductionReturnOrders(c *gin.Context) {
	var q request.ProductionReturnOrderQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if _, ok := middleware.GetUserID(c); !ok {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	out, err := service.ListProductionReturnOrders(c.Request.Context(), q)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}
	response.Success(c, out)
}

func GetProductionReturnOrderDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if _, ok := middleware.GetUserID(c); !ok {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	out, err := service.GetProductionReturnOrderDetail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}
	response.Success(c, out)
}

// UpdateProductionReturnOrder 更新生产退货单（成本价格、备注）
func UpdateProductionReturnOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	var req request.ProductionReturnOrderUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	if err := service.UpdateProductionReturnOrder(c.Request.Context(), id, req, userID); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}
	response.Success(c, nil)
}

// ListConsumptionOrdersByProductionReturn 查询生产退货单关联的领料单
func ListConsumptionOrdersByProductionReturn(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if _, ok := middleware.GetUserID(c); !ok {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	out, err := service.ListConsumptionOrdersByProductionReturn(c.Request.Context(), id)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}
	response.Success(c, out)
}

// ListReversalOrdersByProductionReturn 查询生产退货单关联的退料单
func ListReversalOrdersByProductionReturn(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if _, ok := middleware.GetUserID(c); !ok {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	out, err := service.ListReversalOrdersByProductionReturn(c.Request.Context(), id)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}
	response.Success(c, out)
}

// ListProductionOrdersForDropdown 下拉列表：生产单
func ListProductionOrdersForDropdown(c *gin.Context) {
	if _, ok := middleware.GetUserID(c); !ok {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}
	keyword := c.Query("keyword")
	out, err := service.ListProductionOrdersForDropdown(c.Request.Context(), keyword)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}
	response.Success(c, out)
}

// ListProductionReturnOrdersForDropdown 下拉列表：生产退货单
func ListProductionReturnOrdersForDropdown(c *gin.Context) {
	if _, ok := middleware.GetUserID(c); !ok {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}
	keyword := c.Query("keyword")
	out, err := service.ListProductionReturnOrdersForDropdown(c.Request.Context(), keyword)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}
	response.Success(c, out)
}

// CreateProductionReturnOrder 创建生产退货单（成品退回，出库减少库存）
func CreateProductionReturnOrder(c *gin.Context) {
	var req request.CreateProductionReturnOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}
	usernameVal, _ := c.Get("username")
	username, _ := usernameVal.(string)
	id, err := service.CreateProductionReturnOrder(c.Request.Context(), req, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}
