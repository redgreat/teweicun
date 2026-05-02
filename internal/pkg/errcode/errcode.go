/**
 * 功能：errcode.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

// Package errcode documentation
package errcode

import "net/http"

type AppError struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	HTTPCode int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Msg
}

// System level errors (1000 - 1999)
var (
	Success             = &AppError{Code: 0, Msg: "success", HTTPCode: http.StatusOK}
	ErrInternalServer   = &AppError{Code: 1000, Msg: "Internal server error", HTTPCode: http.StatusInternalServerError}
	ErrInvalidParam     = &AppError{Code: 1001, Msg: "Invalid parameter", HTTPCode: http.StatusBadRequest}
	ErrUnauthorized     = &AppError{Code: 1002, Msg: "Unauthorized", HTTPCode: http.StatusUnauthorized}
	ErrForbidden        = &AppError{Code: 1003, Msg: "Forbidden", HTTPCode: http.StatusForbidden}
	ErrNotFound         = &AppError{Code: 1004, Msg: "Not found", HTTPCode: http.StatusNotFound}
	ErrTooManyRequests  = &AppError{Code: 1005, Msg: "Too many requests", HTTPCode: http.StatusTooManyRequests}
)

// Business level errors (2000 - 9999)
var (
	ErrUserNotFound          = &AppError{Code: 2001, Msg: "User not found", HTTPCode: http.StatusNotFound}
	ErrInvalidPassword       = &AppError{Code: 2002, Msg: "Invalid password", HTTPCode: http.StatusBadRequest}
	ErrDataConcurrencyConflict = &AppError{Code: 2003, Msg: "Data concurrency conflict", HTTPCode: http.StatusConflict}
	
	// Add more business logic errors here
	ErrMaterialNotFound      = &AppError{Code: 3001, Msg: "Material not found", HTTPCode: http.StatusNotFound}
	ErrStockNotEnough        = &AppError{Code: 3002, Msg: "Stock not enough", HTTPCode: http.StatusBadRequest}
)

// NewAppError creates a new custom error
func NewAppError(code int, msg string, httpCode ...int) *AppError {
	hc := http.StatusInternalServerError
	if len(httpCode) > 0 {
		hc = httpCode[0]
	}
	return &AppError{
		Code:     code,
		Msg:      msg,
		HTTPCode: hc,
	}
}
