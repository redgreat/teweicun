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

func ListReversalOrders(c *gin.Context) {
	var query request.ReversalOrderQuery
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

	result, err := service.ListReversalOrders(c.Request.Context(), query)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}

	response.Success(c, result)
}

func GetReversalOrderDetail(c *gin.Context) {
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

	detail, err := service.GetReversalOrderDetail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}

	response.Success(c, detail)
}

func CreateReversalOrder(c *gin.Context) {
	var req request.ReversalOrderCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
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

	orderID, err := service.CreateReversalOrder(c.Request.Context(), req, userID, username)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}

	response.Success(c, gin.H{"id": orderID})
}

func UpdateReversalOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的订单ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	var req request.ReversalOrderUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
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

	if len(req.Items) > 0 {
		err = service.UpdateReversalOrder(c.Request.Context(), id, req, userID, username)
	} else {
		err = service.UpdateReversalOrderStatus(c.Request.Context(), id, "pending", userID)
	}
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}

	response.Success(c, nil)
}

func ConfirmReversalOrder(c *gin.Context) {
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

	err = service.ConfirmReversalOrder(c.Request.Context(), id, userID)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}

	response.Success(c, nil)
}

func DeleteReversalOrder(c *gin.Context) {
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

	err = service.DeleteReversalOrder(c.Request.Context(), id, userID)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}

	response.Success(c, nil)
}