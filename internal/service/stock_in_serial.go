package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
)

func UpdateStockInItemSerialSelections(ctx context.Context, stockInItemID int64, serialIDs []int64, userID int64) error {
	return db.UpdateStockInItemSerialSelections(ctx, stockInItemID, serialIDs, userID)
}

