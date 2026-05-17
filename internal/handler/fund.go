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

func ListFundPayments(c *gin.Context) {
	var q request.FundPaymentQuery
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

	list, total, err := service.ListFundPayments(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func GetFundPayment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	order, err := service.GetFundPayment(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	if order == nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.Success(c, order)
}

func CreateFundPayment(c *gin.Context) {
	var req request.CreateFundPaymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	username, _ := c.Get("username")
	id, err := service.CreateFundPayment(c.Request.Context(), &req, userID, username.(string))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func ListFundCollections(c *gin.Context) {
	var q request.FundCollectionQuery
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

	list, total, err := service.ListFundCollections(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func GetFundCollection(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	order, err := service.GetFundCollection(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	if order == nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.Success(c, order)
}

func CreateFundCollection(c *gin.Context) {
	var req request.CreateFundCollectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	username, _ := c.Get("username")
	id, err := service.CreateFundCollection(c.Request.Context(), &req, userID, username.(string))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}
