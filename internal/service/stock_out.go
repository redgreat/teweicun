/**
 * 功能：stock_out.go
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

func ListStockOuts(ctx context.Context, q *request.StockOutQuery) ([]response.StockOutResp, int64, error) {
	return db.ListStockOuts(ctx, q)
}

func GetStockOutDetail(ctx context.Context, id int64) (*response.StockOutResp, error) {
	return db.GetStockOutDetail(ctx, id)
}

func CreateStockOut(ctx context.Context, req *request.CreateStockOutReq, userID int64, username string) (int64, error) {
	return db.CreateStockOut(ctx, req, userID, username)
}

func ConfirmStockOut(ctx context.Context, id int64, userID int64, username string) error {
	return db.ConfirmStockOut(ctx, id, userID, username)
}

func UpdateStockOutSerialSelections(ctx context.Context, stockOutID, userID int64, req *request.StockOutSerialSelectionReq) error {
	return db.UpdateStockOutSerialSelections(ctx, stockOutID, userID, req)
}

func UpdateStockOutItemSerialSelections(ctx context.Context, stockOutItemID int64, req *request.UpdateStockOutItemSerialSelectionsReq, userID int64) error {
	return db.UpdateStockOutItemSerialSelections(ctx, stockOutItemID, req.SerialCodeIDs, userID)
}
