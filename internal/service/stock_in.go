/**
 * 功能：入库单业务服务
 * 创建时间：2026-04-18
 * 创建人：CodeArts Agent
 */

package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
)

func ListStockIns(ctx context.Context, q *request.StockInQuery) ([]response.StockInResp, int64, error) {
	return db.ListStockIns(ctx, q)
}

func GetStockIn(ctx context.Context, id int64) (*response.StockInDetailResp, error) {
	return db.GetStockInByID(ctx, id)
}

func CreateStockIn(ctx context.Context, req *request.CreateStockInReq, userID int64) (int64, error) {
	return db.CreateStockIn(ctx, req, userID)
}

func ConfirmStockIn(ctx context.Context, stockInID, userID int64) error {
	return db.ConfirmStockIn(ctx, stockInID, userID)
}

func ConfirmReversalStockIn(ctx context.Context, stockInID, userID int64) error {
	return db.ConfirmReversalStockIn(ctx, stockInID, userID)
}

func ListStockInConfirmLogs(ctx context.Context, stockInID int64) ([]response.StockInConfirmLogResp, error) {
	return db.ListStockInConfirmLogs(ctx, stockInID)
}

func UpdateStockIn(ctx context.Context, id int64, warehouseCode string, remark string, items []request.UpdateStockInItem) error {
	return db.UpdateStockIn(ctx, id, warehouseCode, remark, items)
}
