/**
 * 功能：SKU HTTP处理器
 * 创建时间：2026-04-18
 * 创建人：CodeArts Agent
 * 修改时间：2026-04-19
 * 修改内容：统一使用项目标准的响应格式和GetUserID调用方式
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

func ListSKUs(c *gin.Context) {
	var q request.SKUQuery
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

	list, total, err := service.ListSKUs(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func GetSKU(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的SKU ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	detail, err := service.GetSKU(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, detail)
}

func CreateSKU(c *gin.Context) {
	var req request.CreateSKUReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	id, err := service.CreateSKU(c.Request.Context(), &req, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func UpdateSKU(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的SKU ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	var req request.UpdateSKUReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	err = service.UpdateSKU(c.Request.Context(), id, &req, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func DeleteSKU(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的SKU ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	err = service.DeleteSKU(c.Request.Context(), id, userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ListSKUsByMaterial 获取指定物料下的SKU列表（用于采购/入库选择）
func ListSKUsByMaterial(c *gin.Context) {
	materialID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "无效的物料ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	list, err := service.ListSKUsByMaterial(c.Request.Context(), materialID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, list)
}
