package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
)

func CreateConsumptionOrder(ctx context.Context, req request.ConsumptionOrderCreate, userID int64, username string) (int64, error) {
	return db.CreateConsumptionOrder(ctx, req, userID, username)
}

func ListConsumptionOrders(ctx context.Context, query request.ConsumptionOrderQuery) (*response.ConsumptionOrderListResp, error) {
	return db.ListConsumptionOrders(ctx, query)
}

func GetConsumptionOrderDetail(ctx context.Context, id int64) (*response.ConsumptionOrderResp, error) {
	return db.GetConsumptionOrderDetail(ctx, id)
}

func UpdateConsumptionOrderStatus(ctx context.Context, id int64, status string, userID int64) error {
	return db.UpdateConsumptionOrderStatus(ctx, id, status, userID)
}

func ConfirmConsumptionOrder(ctx context.Context, id int64, userID int64) error {
	return db.ConfirmConsumptionOrder(ctx, id, userID)
}

func DeleteConsumptionOrder(ctx context.Context, id int64, userID int64) error {
	return db.DeleteConsumptionOrder(ctx, id, userID)
}