package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/service"
)

func GetMaterialSerialCodesByStockInItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	result, err := service.GetMaterialSerialCodesByStockInItem(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func GetMaterialSerialCodesByStockOutItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	result, err := service.GetMaterialSerialCodesByStockOutItem(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func GetAvailableMaterialSerialCodesByStockOutItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	result, err := service.GetAvailableMaterialSerialCodesByStockOutItem(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func GetAvailableIssuedMaterialSerialCodesByStockInItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	result, err := service.GetAvailableIssuedMaterialSerialCodesByStockInItem(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
