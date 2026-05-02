/**
 * 功能：inventory_check.go
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

func ListInventoryChecks(ctx context.Context, q *request.InventoryCheckQuery) ([]response.InventoryCheckResp, int64, error) {
	return db.ListInventoryChecks(ctx, q)
}

func GetInventoryCheckDetail(ctx context.Context, id int64) (*response.InventoryCheckResp, error) {
	return db.GetInventoryCheckDetail(ctx, id)
}

func CreateInventoryCheck(ctx context.Context, req *request.CreateInventoryCheckReq, userID int64, username string) (int64, error) {
	return db.CreateInventoryCheck(ctx, req, userID, username)
}

func ConfirmInventoryCheck(ctx context.Context, id int64, userID int64, username string) error {
	return db.ConfirmInventoryCheck(ctx, id, userID, username)
}
