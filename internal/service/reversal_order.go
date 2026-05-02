package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
)

func CreateReversalOrder(ctx context.Context, req request.ReversalOrderCreate, userID int64, username string) (int64, error) {
	return db.CreateReversalOrder(ctx, req, userID, username)
}

func ListReversalOrders(ctx context.Context, query request.ReversalOrderQuery) (*response.ReversalOrderListResp, error) {
	return db.ListReversalOrders(ctx, query)
}

func GetReversalOrderDetail(ctx context.Context, id int64) (*response.ReversalOrderResp, error) {
	return db.GetReversalOrderDetail(ctx, id)
}

func UpdateReversalOrderStatus(ctx context.Context, id int64, status string, userID int64) error {
	return db.UpdateReversalOrderStatus(ctx, id, status, userID)
}

func ConfirmReversalOrder(ctx context.Context, id int64, userID int64) error {
	return db.ConfirmReversalOrder(ctx, id, userID)
}

func DeleteReversalOrder(ctx context.Context, id int64, userID int64) error {
	return db.DeleteReversalOrder(ctx, id, userID)
}