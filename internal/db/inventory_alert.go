/**
 * 功能：inventory_alert.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

// ListInventoryAlerts 查询库存预警 (直接查询视图)
func ListInventoryAlerts(ctx context.Context, q *request.InventoryAlertQuery) ([]response.InventoryAlertResp, int64, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if q.MaterialName != "" {
		where = append(where, fmt.Sprintf("material_name ILIKE $%d", argID))
		args = append(args, "%"+q.MaterialName+"%")
		argID++
	}
	if q.AlertType != "" {
		where = append(where, fmt.Sprintf("alert_type = $%d", argID))
		args = append(args, q.AlertType)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_inventory_alert WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT material_id, material_code, material_name, unit, 
		       total_quantity, safety_stock, max_stock, alert_type, alert_quantity
		FROM v_inventory_alert
		WHERE %s
		ORDER BY material_code ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.InventoryAlertResp
	for rows.Next() {
		var item response.InventoryAlertResp
		if err := rows.Scan(&item.MaterialID, &item.MaterialCode, &item.MaterialName, &item.Unit,
			&item.TotalQuantity, &item.SafetyStock, &item.MaxStock, &item.AlertType, &item.AlertQuantity); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// CheckInventoryAlerts 触发预警检查 (生成通知)
func CheckInventoryAlerts(ctx context.Context) error {
	_, err := database.Pool.Exec(ctx, "CALL sp_check_stock_alerts()")
	return err
}
