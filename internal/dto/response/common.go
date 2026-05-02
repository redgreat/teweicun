/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
)

type PageResult struct {
	Total int64       `json:"total"`
	List  interface{} `json:"list"`
}

// SuccessPage returns a paginated success response
func SuccessPage(c *gin.Context, total int64, list interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: errcode.Success.Code,
		Msg:  errcode.Success.Msg,
		Data: PageResult{
			Total: total,
			List:  list,
		},
	})
}
