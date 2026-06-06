/**
 * 功能：warehouse.go
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

func ListWarehouses(c *gin.Context) {
	var q request.WarehouseQuery
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

	list, total, err := service.ListWarehouses(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func CreateWarehouse(c *gin.Context) {
	var req request.CreateWarehouseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	id, code, err := service.CreateWarehouse(c.Request.Context(), &req, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"id": id, "warehouse_code": code})
}

func UpdateWarehouse(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid warehouse ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	var req request.UpdateWarehouseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	err = service.UpdateWarehouse(c.Request.Context(), id, &req, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func DeleteWarehouse(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid warehouse ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	err = service.DeleteWarehouse(c.Request.Context(), id, userID, username)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}
