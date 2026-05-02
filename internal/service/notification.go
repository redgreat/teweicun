/**
 * 功能：notification.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/response"
)

func ListNotifications(ctx context.Context, userID int64) ([]response.NotificationResp, error) {
	return db.ListNotifications(ctx, userID)
}

func MarkNotificationRead(ctx context.Context, id int64, userID int64) error {
	return db.MarkNotificationRead(ctx, id, userID)
}

func ReportInventoryBalance(ctx context.Context, start, end string) ([]response.InventoryBalanceReport, error) {
	return db.ReportInventoryBalance(ctx, start, end)
}

func ReportInventoryTurnover(ctx context.Context, start, end string) ([]response.InventoryTurnoverReport, error) {
	return db.ReportInventoryTurnover(ctx, start, end)
}
