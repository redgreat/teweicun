package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
)

func GetMaterialSerialCodesByStockInItem(ctx context.Context, stockInItemID int64) ([]db.MaterialSerialCodeItem, error) {
	return db.QueryMaterialSerialCodesByStockInItem(ctx, stockInItemID)
}

func GetMaterialSerialCodesByStockOutItem(ctx context.Context, stockOutItemID int64) ([]db.MaterialSerialCodeItem, error) {
	return db.QueryMaterialSerialCodesByStockOutItem(ctx, stockOutItemID)
}

func GetAvailableMaterialSerialCodesByStockOutItem(ctx context.Context, stockOutItemID int64) ([]db.MaterialSerialCodeItem, error) {
	return db.QueryAvailableMaterialSerialCodesByStockOutItem(ctx, stockOutItemID)
}

func GetAvailableIssuedMaterialSerialCodesByStockInItem(ctx context.Context, stockInItemID int64) ([]db.MaterialSerialCodeItem, error) {
	return db.QueryAvailableIssuedMaterialSerialCodesByStockInItem(ctx, stockInItemID)
}
