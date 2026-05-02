/**
 * 功能：return_order.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
)

func ListReturnOrders(ctx context.Context, q *request.ReturnOrderQuery) ([]response.ReturnOrderResp, int64, error) {
	return db.ListReturnOrders(ctx, q)
}

func GetReturnOrderDetail(ctx context.Context, id int64) (*response.ReturnOrderResp, error) {
	return db.GetReturnOrderDetail(ctx, id)
}

func CreateReturnOrder(ctx context.Context, req *request.CreateReturnOrderReq, userID int64, username string) (int64, error) {
	return db.CreateReturnOrder(ctx, req, userID, username)
}

func UpdateReturnOrder(ctx context.Context, id int64, req *request.UpdateReturnOrderReq, userID int64, username string) error {
	return db.UpdateReturnOrder(ctx, id, req, userID, username)
}

func ConfirmReturnOrder(ctx context.Context, id int64, userID int64, username string) error {
	return db.ConfirmReturnOrder(ctx, id, userID, username)
}

func DeleteReturnOrder(ctx context.Context, id int64, userID int64, username string) error {
	return db.DeleteReturnOrder(ctx, id, userID, username)
}
