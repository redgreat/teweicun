/**
 * 功能：notification.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package db

import (
	"context"

	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

// ListNotifications 获取用户通知列表
func ListNotifications(ctx context.Context, userID int64) ([]response.NotificationResp, error) {
	query := `
		SELECT id, title, content, notify_type, COALESCE(ref_type, ''), COALESCE(ref_id, 0), is_read, created_at, read_at
		FROM sys_notification
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := database.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []response.NotificationResp
	for rows.Next() {
		var item response.NotificationResp
		if err := rows.Scan(&item.ID, &item.Title, &item.Content, &item.NotifyType, &item.RefType, &item.RefID,
			&item.IsRead, &item.CreatedAt, &item.ReadAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

// MarkNotificationRead 标记通知已读
func MarkNotificationRead(ctx context.Context, id int64, userID int64) error {
	query := `UPDATE sys_notification SET is_read = TRUE, read_at = NOW() WHERE id = $1 AND user_id = $2`
	_, err := database.Pool.Exec(ctx, query, id, userID)
	return err
}

// ReportInventoryBalance 进出存变动报表
func ReportInventoryBalance(ctx context.Context, startDate, endDate string) ([]response.InventoryBalanceReport, error) {
	query := `
		SELECT trans_date::text, warehouse_id, material_id, in_qty, out_qty
		FROM v_in_out_summary
		WHERE trans_date BETWEEN $1 AND $2
		ORDER BY trans_date DESC
	`
	rows, err := database.Pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []response.InventoryBalanceReport
	for rows.Next() {
		var item response.InventoryBalanceReport
		if err := rows.Scan(&item.TransDate, &item.WarehouseID, &item.MaterialID, &item.InQty, &item.OutQty); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

// ReportInventoryTurnover 库存周转率报表
func ReportInventoryTurnover(ctx context.Context, startDate, endDate string) ([]response.InventoryTurnoverReport, error) {
	query := `
		SELECT material_id, material_code, material_name, out_total, avg_inventory, turnover_rate
		FROM fn_calc_inventory_turnover($1, $2)
		ORDER BY turnover_rate DESC
	`
	rows, err := database.Pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []response.InventoryTurnoverReport
	for rows.Next() {
		var item response.InventoryTurnoverReport
		if err := rows.Scan(&item.MaterialID, &item.MaterialCode, &item.MaterialName, &item.OutTotal, &item.AvgInventory, &item.TurnoverRate); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}
