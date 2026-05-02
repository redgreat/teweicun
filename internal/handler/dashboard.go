package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/service"
)

func GetBigscreenDashboard(c *gin.Context) {
	rangeKey := strings.TrimSpace(c.Query("range"))
	result, err := service.GetBigscreenDashboard(c.Request.Context(), rangeKey)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

