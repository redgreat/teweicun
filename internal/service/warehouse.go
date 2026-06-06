/**
 * 功能：warehouse.go
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

func ListWarehouses(ctx context.Context, q *request.WarehouseQuery) ([]response.WarehouseResp, int64, error) {
	return db.ListWarehouses(ctx, q)
}

func CreateWarehouse(ctx context.Context, req *request.CreateWarehouseReq, userID int64, username string) (int64, string, error) {
	return db.CreateWarehouse(ctx, req, userID, username)
}

func UpdateWarehouse(ctx context.Context, id int64, req *request.UpdateWarehouseReq, userID int64, username string) error {
	return db.UpdateWarehouse(ctx, id, req, userID, username)
}

func DeleteWarehouse(ctx context.Context, id int64, userID int64, username string) error {
	return db.DeleteWarehouse(ctx, id, userID, username)
}
