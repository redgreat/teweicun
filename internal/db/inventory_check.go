/**
 * 功能：inventory_check.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

// ListInventoryChecks 分页查询盘点单
func ListInventoryChecks(ctx context.Context, q *request.InventoryCheckQuery) ([]response.InventoryCheckResp, int64, error) {
	where := []string{"ic.created_at IS NOT NULL"} // Placeholder for simple status filter
	var args []interface{}
	argID := 1

	if q.CheckNo != "" {
		where = append(where, fmt.Sprintf("ic.check_no ILIKE $%d", argID))
		args = append(args, "%"+q.CheckNo+"%")
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("ic.check_status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM inventory_check ic WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT ic.id, ic.check_no, ic.warehouse_id, w.warehouse_name, ic.check_date, 
		       ic.check_status, ic.checker_id, COALESCE(ic.remark, ''), ic.created_at, ic.approved_at
		FROM inventory_check ic
		LEFT JOIN warehouse w ON w.id = ic.warehouse_id
		WHERE %s
		ORDER BY ic.id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.InventoryCheckResp
	for rows.Next() {
		var item response.InventoryCheckResp
		if err := rows.Scan(&item.ID, &item.CheckNo, &item.WarehouseID, &item.WarehouseName, &item.CheckDate,
			&item.Status, &item.CheckerID, &item.Remark, &item.CreatedAt, &item.ApprovedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// GetInventoryCheckDetail 获取盘点单详情
func GetInventoryCheckDetail(ctx context.Context, id int64) (*response.InventoryCheckResp, error) {
	query := `
		SELECT ic.id, ic.check_no, ic.warehouse_id, w.warehouse_name, ic.check_date, 
		       ic.check_status, ic.checker_id, COALESCE(ic.remark, ''), ic.created_at, ic.approved_at
		FROM inventory_check ic
		LEFT JOIN warehouse w ON w.id = ic.warehouse_id
		WHERE ic.id = $1
	`
	var item response.InventoryCheckResp
	err := database.Pool.QueryRow(ctx, query, id).Scan(&item.ID, &item.CheckNo, &item.WarehouseID, &item.WarehouseName, &item.CheckDate,
		&item.Status, &item.CheckerID, &item.Remark, &item.CreatedAt, &item.ApprovedAt)
	if err != nil {
		return nil, err
	}

	// 查明细
	itemQuery := `
		SELECT ici.id, ici.material_id, m.material_code, m.material_name, 
		       ici.book_quantity, ici.actual_quantity, ici.diff_quantity, COALESCE(ici.diff_reason, '')
		FROM inventory_check_item ici
		INNER JOIN material m ON m.id = ici.material_id
		WHERE ici.check_id = $1
	`
	rows, err := database.Pool.Query(ctx, itemQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sub response.InventoryCheckItemResp
		if err := rows.Scan(&sub.ID, &sub.MaterialID, &sub.MaterialCode, &sub.MaterialName,
			&sub.BookQuantity, &sub.ActualQuantity, &sub.DiffQuantity, &sub.DiffReason); err != nil {
			return nil, err
		}
		item.Items = append(item.Items, sub)
	}

	return &item, rows.Err()
}

// CreateInventoryCheck 创建盘点单（直接进入 counting 状态）
func CreateInventoryCheck(ctx context.Context, req *request.CreateInventoryCheckReq, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// 1. 生成单号
	var checkNo string
	err = tx.QueryRow(ctx, "SELECT fn_generate_serial_no('IC')").Scan(&checkNo)
	if err != nil {
		return 0, err
	}

	// 2. 插入主表
	var id int64
	mainQuery := `
		INSERT INTO inventory_check (check_no, warehouse_id, check_date, check_status, checker_id, remark, created_by)
		VALUES ($1, $2, $3, 'counting', $4, $5, $4)
		RETURNING id
	`
	err = tx.QueryRow(ctx, mainQuery, checkNo, req.WarehouseID, req.CheckDate, userID, req.Remark).Scan(&id)
	if err != nil {
		return 0, err
	}

	// 3. 插入明细表
	for _, item := range req.Items {
		diffQty := item.ActualQuantity - item.BookQuantity
		itemQuery := `
			INSERT INTO inventory_check_item (check_id, material_id, book_quantity, actual_quantity, diff_quantity)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, err = tx.Exec(ctx, itemQuery, id, item.MaterialID, item.BookQuantity, item.ActualQuantity, diffQty)
		if err != nil {
			return 0, err
		}
	}

	// 4. 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "CREATE", "CHECK", "inventory_check", id, nil)
	if err != nil {
		return 0, err
	}

	return id, tx.Commit(ctx)
}

// ConfirmInventoryCheck 确认盘点并调整库存
func ConfirmInventoryCheck(ctx context.Context, checkID int64, userID int64, username string) error {
	query := `CALL sp_confirm_inventory_check($1, $2)`
	_, err := database.Pool.Exec(ctx, query, checkID, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return fmt.Errorf("DB_ERROR: %s", pgErr.Message)
		}
		return err
	}

	// 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, _ = database.Pool.Exec(ctx, auditQuery, userID, username, "CONFIRM", "CHECK", "inventory_check", checkID, nil)

	return nil
}
