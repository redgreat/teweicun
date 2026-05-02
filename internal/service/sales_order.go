/**
 * 功能：sales_order.go
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

func ListSalesOrders(ctx context.Context, q *request.SalesOrderQuery) ([]response.SalesOrderResp, int64, error) {
	return db.ListSalesOrders(ctx, q)
}

func GetSalesOrderDetail(ctx context.Context, id int64) (*response.SalesOrderResp, error) {
	return db.GetSalesOrderDetail(ctx, id)
}

func CreateSalesOrder(ctx context.Context, req *request.CreateSalesOrderReq, userID int64, username string) (int64, error) {
	return db.CreateSalesOrder(ctx, req, userID, username)
}

func ConfirmSalesOrder(ctx context.Context, id int64, userID int64, username string) error {
	return db.ConfirmSalesOrder(ctx, id, userID, username)
}

func CancelSalesOrder(ctx context.Context, id int64, userID int64, username string) error {
	return db.CancelSalesOrder(ctx, id, userID, username)
}

func ShipSalesOrder(ctx context.Context, id int64, req *request.ShipSalesOrderReq, userID int64, username string) error {
	var items []struct {
		MaterialID int64
		Quantity   float64
	}
	for _, item := range req.Items {
		items = append(items, struct {
			MaterialID int64
			Quantity   float64
		}{
			MaterialID: item.MaterialID,
			Quantity:   item.Quantity,
		})
	}
	return db.ShipSalesOrder(ctx, id, items, userID, username)
}
