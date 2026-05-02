/**
 * 功能：物料属性HTTP处理器
 * 创建时间：2026-04-17
 * 创建人：wangcw
 */

package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/service"
)

func ListMaterialAttributeDefs(c *gin.Context) {
	var q request.MaterialAttributeDefQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	result, err := service.ListMaterialAttributeDefs(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

func CreateMaterialAttributeDef(c *gin.Context) {
	var req request.CreateMaterialAttributeDefReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID := c.GetInt64("userID")
	id, err := service.CreateMaterialAttributeDef(c.Request.Context(), &req, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"id": id})
}

func UpdateMaterialAttributeDef(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	var req request.UpdateMaterialAttributeDefReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID := c.GetInt64("userID")
	if err := service.UpdateMaterialAttributeDef(c.Request.Context(), id, &req, userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func DeleteMaterialAttributeDef(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID := c.GetInt64("userID")
	if err := service.DeleteMaterialAttributeDef(c.Request.Context(), id, userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func GetMaterialAttributes(c *gin.Context) {
	materialID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的物料ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	attrs, err := service.GetMaterialAttributes(c.Request.Context(), materialID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, attrs)
}

func UpdateMaterialAttributes(c *gin.Context) {
	materialID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的物料ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	var req request.UpdateMaterialAttributesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID := c.GetInt64("userID")
	if err := service.UpdateMaterialAttributes(c.Request.Context(), materialID, &req, userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}
