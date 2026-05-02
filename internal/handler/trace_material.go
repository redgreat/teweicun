package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/service"
)

func TraceMaterialBySerial(c *gin.Context) {
	serialCode := c.Query("serial_code")
	if serialCode == "" {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	result, err := service.TraceMaterialBySerial(c.Request.Context(), serialCode)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
