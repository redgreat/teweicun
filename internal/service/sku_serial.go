package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
)

func GetSerialCodesByStockInItem(ctx context.Context, stockInItemID int64) ([]db.SkuSerialCodeItem, error) {
	return db.QuerySerialCodesByStockInItem(ctx, stockInItemID)
}

func GetSerialCodesByStockOutItem(ctx context.Context, stockOutItemID int64) ([]db.SkuSerialCodeItem, error) {
	return db.QuerySerialCodesByStockOutItem(ctx, stockOutItemID)
}

func GetAvailableSerialCodesByStockOutItem(ctx context.Context, stockOutItemID int64) ([]db.SkuSerialCodeItem, error) {
	return db.QueryAvailableSerialCodesByStockOutItem(ctx, stockOutItemID)
}

func GetAvailableIssuedSerialCodesByStockInItem(ctx context.Context, stockInItemID int64) ([]db.SkuSerialCodeItem, error) {
	return db.QueryAvailableIssuedSerialCodesByStockInItem(ctx, stockInItemID)
}
