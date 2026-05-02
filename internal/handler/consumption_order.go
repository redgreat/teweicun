package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/middleware"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/service"
)

func ListConsumptionOrders(c *gin.Context) {
	var query request.ConsumptionOrderQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errcode.NewAppError(errcode.ErrUnauthorized.Code, errcode.ErrUnauthorized.Msg, errcode.ErrUnauthorized.HTTPCode))
		return
	}

	_ = userID // 暂时不使用，但保留用于权限检查

	result, err := service.ListConsumptionOrders(c.Request.Context(), query)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}

	response.Success(c, result)
}

func GetConsumptionOrderDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的订单ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errcode.NewAppError(errcode.ErrUnauthorized.Code, errcode.ErrUnauthorized.Msg, errcode.ErrUnauthorized.HTTPCode))
		return
	}

	_ = userID // 暂时不使用，但保留用于权限检查

	detail, err := service.GetConsumptionOrderDetail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}

	response.Success(c, detail)
}

func CreateConsumptionOrder(c *gin.Context) {
	var req request.ConsumptionOrderCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	// 兼容前端传 YYYY-MM-DD
	if t, err := time.ParseInLocation("2006-01-02", req.OrderDate, time.Local); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "领料日期格式错误，应为 YYYY-MM-DD", errcode.ErrInvalidParam.HTTPCode))
		return
	} else {
		req.OrderDateTime = t
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errcode.NewAppError(errcode.ErrUnauthorized.Code, errcode.ErrUnauthorized.Msg, errcode.ErrUnauthorized.HTTPCode))
		return
	}

	username := c.GetString("username")
	if username == "" {
		username = "未知用户"
	}

	req.DesignerID = userID
	req.DesignerName = username

	orderID, err := service.CreateConsumptionOrder(c.Request.Context(), req, userID, username)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}

	response.Success(c, gin.H{"id": orderID})
}

func UpdateConsumptionOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的订单ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	var req request.ConsumptionOrderUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errcode.NewAppError(errcode.ErrUnauthorized.Code, errcode.ErrUnauthorized.Msg, errcode.ErrUnauthorized.HTTPCode))
		return
	}

	err = service.UpdateConsumptionOrderStatus(c.Request.Context(), id, "pending", userID)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}

	response.Success(c, nil)
}

func ConfirmConsumptionOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的订单ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errcode.NewAppError(errcode.ErrUnauthorized.Code, errcode.ErrUnauthorized.Msg, errcode.ErrUnauthorized.HTTPCode))
		return
	}

	err = service.ConfirmConsumptionOrder(c.Request.Context(), id, userID)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}

	response.Success(c, nil)
}

func DeleteConsumptionOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的订单ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, errcode.NewAppError(errcode.ErrUnauthorized.Code, errcode.ErrUnauthorized.Msg, errcode.ErrUnauthorized.HTTPCode))
		return
	}

	err = service.DeleteConsumptionOrder(c.Request.Context(), id, userID)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}

	response.Success(c, nil)
}
