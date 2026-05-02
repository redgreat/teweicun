/**
 * 功能：stock_transfer.go
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

// ListStockTransfers 分页查询调拨单
func ListStockTransfers(ctx context.Context, q *request.StockTransferQuery) ([]response.StockTransferResp, int64, error) {
	where := []string{"tr.deleted_at IS NULL"}
	var args []interface{}
	argID := 1

	if q.TransferNo != "" {
		where = append(where, fmt.Sprintf("tr.transfer_no ILIKE $%d", argID))
		args = append(args, "%"+q.TransferNo+"%")
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("tr.transfer_status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM stock_transfer tr WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT tr.id, tr.transfer_no, tr.from_warehouse_id, fw.warehouse_name, 
		       tr.to_warehouse_id, tw.warehouse_name, tr.transfer_date, 
		       tr.transfer_status, COALESCE(tr.remark, ''), tr.created_at
		FROM stock_transfer tr
		LEFT JOIN warehouse fw ON fw.id = tr.from_warehouse_id
		LEFT JOIN warehouse tw ON tw.id = tr.to_warehouse_id
		WHERE %s
		ORDER BY tr.id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.StockTransferResp
	for rows.Next() {
		var item response.StockTransferResp
		if err := rows.Scan(&item.ID, &item.TransferNo, &item.FromWarehouseID, &item.FromWarehouseName,
			&item.ToWarehouseID, &item.ToWarehouseName, &item.TransferDate, &item.Status, &item.Remark, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// GetStockTransferDetail 获取调拨单详情
func GetStockTransferDetail(ctx context.Context, id int64) (*response.StockTransferResp, error) {
	query := `
		SELECT tr.id, tr.transfer_no, tr.from_warehouse_id, fw.warehouse_name, 
		       tr.to_warehouse_id, tw.warehouse_name, tr.transfer_date, 
		       tr.transfer_status, COALESCE(tr.remark, ''), tr.created_at
		FROM stock_transfer tr
		LEFT JOIN warehouse fw ON fw.id = tr.from_warehouse_id
		LEFT JOIN warehouse tw ON tw.id = tr.to_warehouse_id
		WHERE tr.id = $1 AND tr.deleted_at IS NULL
	`
	var item response.StockTransferResp
	err := database.Pool.QueryRow(ctx, query, id).Scan(&item.ID, &item.TransferNo, &item.FromWarehouseID, &item.FromWarehouseName,
		&item.ToWarehouseID, &item.ToWarehouseName, &item.TransferDate, &item.Status, &item.Remark, &item.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 查明细
	itemQuery := `
		SELECT sti.id, sti.material_id, m.material_code, m.material_name, 
		       sti.quantity, m.unit, COALESCE(sti.remark, '')
		FROM stock_transfer_item sti
		INNER JOIN material m ON m.id = sti.material_id
		WHERE sti.transfer_id = $1
	`
	rows, err := database.Pool.Query(ctx, itemQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sub response.StockTransferItemResp
		if err := rows.Scan(&sub.ID, &sub.MaterialID, &sub.MaterialCode, &sub.MaterialName,
			&sub.Quantity, &sub.Unit, &sub.Remark); err != nil {
			return nil, err
		}
		item.Items = append(item.Items, sub)
	}

	return &item, rows.Err()
}

// CreateStockTransfer 创建调拨单
func CreateStockTransfer(ctx context.Context, req *request.CreateStockTransferReq, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// 1. 生成单号
	var transferNo string
	err = tx.QueryRow(ctx, "SELECT fn_generate_serial_no('TR')").Scan(&transferNo)
	if err != nil {
		return 0, err
	}

	// 2. 插入主表
	var id int64
	mainQuery := `
		INSERT INTO stock_transfer (transfer_no, from_warehouse_id, to_warehouse_id, transfer_date, 
		                           transfer_status, remark, created_by)
		VALUES ($1, $2, $3, $4, 'draft', $5, $6)
		RETURNING id
	`
	err = tx.QueryRow(ctx, mainQuery, transferNo, req.FromWarehouseID, req.ToWarehouseID,
		req.TransferDate, req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
	}

	// 3. 插入明细表
	for _, item := range req.Items {
		itemQuery := `
			INSERT INTO stock_transfer_item (transfer_id, material_id, quantity, unit, remark)
			VALUES ($1, $2, $3, (SELECT unit FROM material WHERE id = $2), $4)
		`
		_, err = tx.Exec(ctx, itemQuery, id, item.MaterialID, item.Quantity, item.Remark)
		if err != nil {
			return 0, err
		}
	}

	// 4. 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "CREATE", "TRANSFER", "stock_transfer", id, nil)
	if err != nil {
		return 0, err
	}

	return id, tx.Commit(ctx)
}

// ConfirmTransferOut 确认调出
func ConfirmTransferOut(ctx context.Context, transferID int64, userID int64, username string) error {
	query := `CALL sp_confirm_transfer_out($1, $2)`
	_, err := database.Pool.Exec(ctx, query, transferID, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return fmt.Errorf("DB_ERROR: %s", pgErr.Message)
		}
		return err
	}
	return nil
}

// ConfirmTransferIn 确认调入
func ConfirmTransferIn(ctx context.Context, transferID int64, userID int64, username string) error {
	query := `CALL sp_confirm_transfer_in($1, $2)`
	_, err := database.Pool.Exec(ctx, query, transferID, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return fmt.Errorf("DB_ERROR: %s", pgErr.Message)
		}
		return err
	}
	return nil
}
