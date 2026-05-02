/**
 * 功能：stock_transfer.go
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

func ListStockTransfers(ctx context.Context, q *request.StockTransferQuery) ([]response.StockTransferResp, int64, error) {
	return db.ListStockTransfers(ctx, q)
}

func GetStockTransferDetail(ctx context.Context, id int64) (*response.StockTransferResp, error) {
	return db.GetStockTransferDetail(ctx, id)
}

func CreateStockTransfer(ctx context.Context, req *request.CreateStockTransferReq, userID int64, username string) (int64, error) {
	return db.CreateStockTransfer(ctx, req, userID, username)
}

func ConfirmTransferOut(ctx context.Context, id int64, userID int64, username string) error {
	return db.ConfirmTransferOut(ctx, id, userID, username)
}

func ConfirmTransferIn(ctx context.Context, id int64, userID int64, username string) error {
	return db.ConfirmTransferIn(ctx, id, userID, username)
}
