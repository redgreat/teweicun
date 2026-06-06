/**
 * 功能：warehouse.go
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

// ListWarehouses 分页查询仓库
func ListWarehouses(ctx context.Context, q *request.WarehouseQuery) ([]response.WarehouseResp, int64, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if q.WarehouseName != "" {
		where = append(where, fmt.Sprintf("warehouse_name ILIKE $%d", argID))
		args = append(args, "%"+q.WarehouseName+"%")
		argID++
	}
	if q.WarehouseCode != "" {
		where = append(where, fmt.Sprintf("warehouse_code ILIKE $%d", argID))
		args = append(args, "%"+q.WarehouseCode+"%")
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_warehouse_list WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, warehouse_code, warehouse_name, warehouse_type, warehouse_type_name,
		       manager_id, manager_name, status, status_name,
		       created_at, updated_at
		FROM v_warehouse_list
		WHERE %s
		ORDER BY id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.WarehouseResp
	for rows.Next() {
		var item response.WarehouseResp
		if err := rows.Scan(&item.ID, &item.WarehouseCode, &item.WarehouseName, &item.WarehouseType, &item.WarehouseTypeName,
			&item.ManagerID, &item.ManagerName, &item.Status, &item.StatusName,
			&item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// CreateWarehouse 创建仓库
func CreateWarehouse(ctx context.Context, req *request.CreateWarehouseReq, userID int64, username string) (int64, string, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, "", err
	}
	defer tx.Rollback(ctx)

	// 自动生成仓库编码
	var warehouseCode string
	err = tx.QueryRow(ctx, "SELECT fn_generate_base_code('W')").Scan(&warehouseCode)
	if err != nil {
		return 0, "", fmt.Errorf("生成仓库编码失败: %w", err)
	}

	var id int64
	query := `
		INSERT INTO warehouse (warehouse_code, warehouse_name, warehouse_type, manager_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	err = tx.QueryRow(ctx, query,
		warehouseCode, req.WarehouseName, req.WarehouseType, req.ManagerID).Scan(&id)
	if err != nil {
		return 0, "", err
	}

	// 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "CREATE", "WAREHOUSE", "warehouse", id, nil)
	if err != nil {
		return 0, "", err
	}

	return id, warehouseCode, tx.Commit(ctx)
}

// UpdateWarehouse 更新仓库
func UpdateWarehouse(ctx context.Context, id int64, req *request.UpdateWarehouseReq, userID int64, username string) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 禁用验证：检查是否有库存记录
	if req.Status == "disabled" {
		var refCount int64
		err = tx.QueryRow(ctx, `SELECT count(*) FROM inventory WHERE warehouse_id = $1 AND quantity > 0`, id).Scan(&refCount)
		if err != nil {
			return err
		}
		if refCount > 0 {
			return fmt.Errorf("该仓库有 %d 条库存记录，无法禁用", refCount)
		}
	}

	query := `
		UPDATE warehouse
		SET warehouse_name = $1, warehouse_type = $2, manager_id = $3, status = $4, updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL
	`
	res, err := tx.Exec(ctx, query,
		req.WarehouseName, req.WarehouseType, req.ManagerID, req.Status, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("warehouse not found")
	}

	// 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "UPDATE", "WAREHOUSE", "warehouse", id, nil)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// DeleteWarehouse 软删除仓库
func DeleteWarehouse(ctx context.Context, id int64, userID int64, username string) error {
	// 删除验证：检查是否有库存记录
	var invCount int64
	err := database.Pool.QueryRow(ctx, `SELECT count(*) FROM inventory WHERE warehouse_id = $1 AND quantity > 0`, id).Scan(&invCount)
	if err != nil {
		return err
	}
	if invCount > 0 {
		return fmt.Errorf("该仓库有 %d 条库存记录，无法删除", invCount)
	}

	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `UPDATE warehouse SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := tx.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("warehouse not found")
	}

	// 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "DELETE", "WAREHOUSE", "warehouse", id, nil)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
