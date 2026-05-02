/**
 * 功能：category.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/service"
)

// GetCategoryTree handles the category list api which returns the entire category tree
func GetCategoryTree(c *gin.Context) {
	rawJsonStr, err := service.GetCategoryTree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	var data interface{}
	if len(rawJsonStr) > 0 {
		_ = json.Unmarshal(rawJsonStr, &data)
	}

	response.Success(c, data)
}

// CreateCategory handles creating a new material category
func CreateCategory(c *gin.Context) {
	var req request.CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	id, err := service.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"id": id})
}

// UpdateCategory handles updating an existing material category
func UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid category ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	var req request.UpdateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	err = service.UpdateCategory(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteCategory handles deleting a material category
func DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid category ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	err = service.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}
