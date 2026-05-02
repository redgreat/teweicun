/**
 * 功能：dict.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/service"
	"github.com/redgreat/teweicun/internal/dto/response"
)

// ListDictTypes handles fetching a page of dictionary types
func ListDictTypes(c *gin.Context) {
	var q request.DictTypeQuery
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

	list, total, err := service.ListDictTypes(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

// CreateDictType handles creating a dictionary type
func CreateDictType(c *gin.Context) {
	var req request.CreateDictTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}
	id, err := service.CreateDictType(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

// UpdateDictType handles updating a dictionary type
func UpdateDictType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "invalid id", errcode.ErrInvalidParam.HTTPCode))
		return
	}
	var req request.UpdateDictTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if err := service.UpdateDictType(c.Request.Context(), id, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// DeleteDictType handles deleting a dictionary type
func DeleteDictType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "invalid id", errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if err := service.DeleteDictType(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ListDictData handles fetching dictionary data items by type
func ListDictData(c *gin.Context) {
	dictType := c.Param("type")
	if dictType == "" {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Dict type is required", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	list, err := service.ListDictDataByDictType(c.Request.Context(), dictType)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, list)
}

// CreateDictData handles creating a dictionary data item
func CreateDictData(c *gin.Context) {
	var req request.CreateDictDataReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}
	id, err := service.CreateDictData(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

// UpdateDictData handles updating a dictionary data item
func UpdateDictData(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "invalid id", errcode.ErrInvalidParam.HTTPCode))
		return
	}
	var req request.UpdateDictDataReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if err := service.UpdateDictData(c.Request.Context(), id, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// DeleteDictData handles deleting a dictionary data item
func DeleteDictData(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "invalid id", errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if err := service.DeleteDictData(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}
