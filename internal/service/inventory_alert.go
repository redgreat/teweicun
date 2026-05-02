/**
 * 功能：inventory_alert.go
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

func ListInventoryAlerts(ctx context.Context, q *request.InventoryAlertQuery) ([]response.InventoryAlertResp, int64, error) {
	return db.ListInventoryAlerts(ctx, q)
}

func CheckInventoryAlerts(ctx context.Context) error {
	return db.CheckInventoryAlerts(ctx)
}
