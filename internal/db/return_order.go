/**
 * 功能：return_order.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

// ListReturnOrders 分页查询退货单
func ListReturnOrders(ctx context.Context, q *request.ReturnOrderQuery) ([]response.ReturnOrderResp, int64, error) {
	where := []string{"ro.deleted_at IS NULL"}
	var args []interface{}
	argID := 1

	if q.ReturnNo != "" {
		where = append(where, fmt.Sprintf("ro.return_no ILIKE $%d", argID))
		args = append(args, "%"+q.ReturnNo+"%")
		argID++
	}
	if q.ReturnType != "" {
		where = append(where, fmt.Sprintf("ro.return_type = $%d", argID))
		args = append(args, q.ReturnType)
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("ro.return_status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}
	if q.SupplierCode != "" {
		where = append(where, fmt.Sprintf("ro.supplier_code = $%d", argID))
		args = append(args, q.SupplierCode)
		argID++
	}
	if q.StartDate != "" {
		where = append(where, fmt.Sprintf("ro.return_date >= $%d", argID))
		args = append(args, q.StartDate)
		argID++
	}
	if q.EndDate != "" {
		where = append(where, fmt.Sprintf("ro.return_date <= $%d", argID))
		args = append(args, q.EndDate)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM return_order ro WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT ro.id, ro.return_no, ro.return_type, COALESCE(ro.ref_doc_type, ''), 
		       ro.ref_doc_id, ro.warehouse_code, w.warehouse_name,
		       COALESCE(ro.supplier_code, ''), COALESCE(s.supplier_name, ''),
		       ro.stock_out_id, COALESCE(so.stock_out_no, ''),
		       ro.return_date, ro.return_status,
		       COALESCE((
		         SELECT SUM(roi.quantity * COALESCE(inv.unit_cost, 0))
		         FROM return_order_item roi
		         LEFT JOIN inventory inv ON inv.id = roi.inventory_id
		         WHERE roi.return_id = ro.id
		       ), 0),
		       COALESCE(ro.remark, ''), ro.created_at
		FROM return_order ro
		LEFT JOIN warehouse w ON w.warehouse_code = ro.warehouse_code
		LEFT JOIN supplier s ON s.supplier_code = ro.supplier_code
		LEFT JOIN stock_out so ON so.id = ro.stock_out_id AND so.deleted_at IS NULL
		WHERE %s
		ORDER BY ro.id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.ReturnOrderResp
	for rows.Next() {
		var item response.ReturnOrderResp
		if err := rows.Scan(&item.ID, &item.ReturnNo, &item.ReturnType, &item.RefDocType, &item.RefDocID,
			&item.WarehouseCode, &item.WarehouseName,
			&item.SupplierCode, &item.SupplierName,
			&item.StockOutID, &item.StockOutNo,
			&item.ReturnDate, &item.Status, &item.TotalAmount, &item.Remark, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// GetReturnOrderDetail 获取退货单详情
func GetReturnOrderDetail(ctx context.Context, id int64) (*response.ReturnOrderResp, error) {
	query := `
		SELECT ro.id, ro.return_no, ro.return_type, COALESCE(ro.ref_doc_type, ''), 
		       ro.ref_doc_id, ro.warehouse_code, w.warehouse_name,
		       COALESCE(ro.supplier_code, ''), COALESCE(s.supplier_name, ''),
		       ro.stock_out_id, COALESCE(so.stock_out_no, ''),
		       ro.return_date, ro.return_status, COALESCE(ro.remark, ''), ro.created_at
		FROM return_order ro
		LEFT JOIN warehouse w ON w.warehouse_code = ro.warehouse_code
		LEFT JOIN supplier s ON s.supplier_code = ro.supplier_code
		LEFT JOIN stock_out so ON so.id = ro.stock_out_id AND so.deleted_at IS NULL
		WHERE ro.id = $1 AND ro.deleted_at IS NULL
	`
	var item response.ReturnOrderResp
	err := database.Pool.QueryRow(ctx, query, id).Scan(&item.ID, &item.ReturnNo, &item.ReturnType, &item.RefDocType, &item.RefDocID,
		&item.WarehouseCode, &item.WarehouseName,
		&item.SupplierCode, &item.SupplierName,
		&item.StockOutID, &item.StockOutNo,
		&item.ReturnDate, &item.Status, &item.Remark, &item.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 查明细（含当前可用量、单价，供编辑页校验与展示）
	itemQuery := `
		SELECT roi.id, roi.material_id, m.material_code, m.material_name,
		       COALESCE(roi.sku_id, 0), COALESCE(v.sku_code, ''), COALESCE(v.sku_name, ''),
		       COALESCE(roi.inventory_id, 0),
		       COALESCE(roi.warehouse_code, ''), COALESCE(wh_line.warehouse_name, ''),
		       roi.quantity, roi.unit,
		       COALESCE(inv.quantity - inv.locked_quantity - COALESCE(inv.in_transit_quantity, 0), 0),
		       COALESCE(inv.unit_cost, 0)
		FROM return_order_item roi
		INNER JOIN material m ON m.id = roi.material_id
		LEFT JOIN v_sku_list v ON v.id = roi.sku_id
		LEFT JOIN warehouse wh_line ON wh_line.warehouse_code = roi.warehouse_code AND wh_line.deleted_at IS NULL
		LEFT JOIN inventory inv ON inv.id = roi.inventory_id
		WHERE roi.return_id = $1
	`
	rows, err := database.Pool.Query(ctx, itemQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sub response.ReturnOrderItemResp
		if err := rows.Scan(&sub.ID, &sub.MaterialID, &sub.MaterialCode, &sub.MaterialName,
			&sub.SKUID, &sub.SKUCode, &sub.SKUName,
			&sub.InventoryID,
			&sub.WarehouseCode, &sub.WarehouseName,
			&sub.Quantity, &sub.Unit,
			&sub.AvailableQuantity, &sub.UnitCost); err != nil {
			return nil, err
		}
		item.Items = append(item.Items, sub)
	}

	return &item, rows.Err()
}

// CreateReturnOrder 创建退货单（待提交）
func CreateReturnOrder(ctx context.Context, req *request.CreateReturnOrderReq, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	if req.ReturnType == "purchase_return" && strings.TrimSpace(req.SupplierCode) == "" {
		return 0, fmt.Errorf("采购退货必须选择供应商")
	}
	var warehouseID int64
	warehouseCode := strings.TrimSpace(req.WarehouseCode)

	// 1. 生成单号 (RT)
	var returnNo string
	err = tx.QueryRow(ctx, "SELECT fn_generate_serial_no('RT')").Scan(&returnNo)
	if err != nil {
		return 0, err
	}

	// 2. 插入主表
	var id int64
	mainQuery := `
		INSERT INTO return_order (return_no, return_type, ref_doc_type, ref_doc_id, 
		                         warehouse_id, warehouse_code, supplier_code,
		                         return_date, return_status, remark, created_by)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, 0), $5, $6, NULLIF($7, ''),
		        $8, 'draft', $9, $10)
		RETURNING id
	`
	err = tx.QueryRow(ctx, mainQuery, returnNo, req.ReturnType, req.RefDocType, req.RefDocID,
		warehouseID, warehouseCode, req.SupplierCode,
		req.ReturnDate, req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
	}

	// 3. 插入明细表
	for _, item := range req.Items {
		if req.ReturnType == "purchase_return" {
			if item.InventoryID == 0 {
				return 0, fmt.Errorf("采购退货明细必须选择在库SKU")
			}

			var materialID int64
			var skuID *int64
			var unit string
			var invWarehouseID int64
			var invWarehouseCode string

			err = tx.QueryRow(ctx, `
				SELECT i.material_id, i.sku_id, i.unit,
				       i.warehouse_id, w.warehouse_code
				FROM inventory i
				INNER JOIN warehouse w ON w.id = i.warehouse_id
				WHERE i.id = $1
			`, item.InventoryID).Scan(&materialID, &skuID, &unit, &invWarehouseID, &invWarehouseCode)
			if err != nil {
				if err == pgx.ErrNoRows {
					return 0, fmt.Errorf("库存不存在")
				}
				return 0, err
			}

			// 采购退货不再要求前端选择仓库：自动从第一条明细带出到主表（兼容旧结构）
			if warehouseID == 0 {
				warehouseID = invWarehouseID
				warehouseCode = invWarehouseCode
				if _, err := tx.Exec(ctx, `UPDATE return_order SET warehouse_id = $1, warehouse_code = $2, updated_at = NOW() WHERE id = $3`, warehouseID, warehouseCode, id); err != nil {
					return 0, err
				}
			}

			itemQuery := `
				INSERT INTO return_order_item (return_id, inventory_id, material_id, sku_id, quantity, unit, warehouse_id, warehouse_code)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`
			_, err = tx.Exec(ctx, itemQuery, id, item.InventoryID, materialID, skuID, item.Quantity, unit, invWarehouseID, invWarehouseCode)
			if err != nil {
				return 0, err
			}
		} else {
			if item.MaterialID == 0 {
				return 0, fmt.Errorf("销售退货明细必须填写物料")
			}
			itemQuery := `
				INSERT INTO return_order_item (return_id, material_id, quantity, unit)
				VALUES ($1, $2, $3, (SELECT unit FROM material WHERE id = $2))
			`
			_, err = tx.Exec(ctx, itemQuery, id, item.MaterialID, item.Quantity)
			if err != nil {
				return 0, err
			}
		}
	}

	// 4. 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "CREATE", "RETURN", "return_order", id, nil)
	if err != nil {
		return 0, err
	}

	return id, tx.Commit(ctx)
}

// UpdateReturnOrder 更新采购退货待提交单（替换明细，逻辑与创建时采购退货分支一致）
func UpdateReturnOrder(ctx context.Context, id int64, req *request.UpdateReturnOrderReq, userID int64, username string) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var rt string
	var st string
	err = tx.QueryRow(ctx, `SELECT return_type, return_status FROM return_order WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&rt, &st)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("退货单不存在")
		}
		return err
	}
	if st != "draft" {
		return fmt.Errorf("仅待提交状态可编辑")
	}
	if rt != "purchase_return" {
		return fmt.Errorf("仅支持编辑采购退货待提交单")
	}
	if strings.TrimSpace(req.SupplierCode) == "" {
		return fmt.Errorf("采购退货必须选择供应商")
	}

	if _, err = tx.Exec(ctx, `DELETE FROM return_order_item WHERE return_id = $1`, id); err != nil {
		return err
	}

	var warehouseID int64
	warehouseCode := ""

	for _, item := range req.Items {
		if item.InventoryID == 0 {
			return fmt.Errorf("采购退货明细必须选择在库SKU")
		}

		var materialID int64
		var skuID *int64
		var unit string
		var invWarehouseID int64
		var invWarehouseCode string

		err = tx.QueryRow(ctx, `
				SELECT i.material_id, i.sku_id, i.unit,
				       i.warehouse_id, w.warehouse_code
				FROM inventory i
				INNER JOIN warehouse w ON w.id = i.warehouse_id
				WHERE i.id = $1
			`, item.InventoryID).Scan(&materialID, &skuID, &unit, &invWarehouseID, &invWarehouseCode)
		if err != nil {
			if err == pgx.ErrNoRows {
				return fmt.Errorf("库存不存在")
			}
			return err
		}

		if warehouseID == 0 {
			warehouseID = invWarehouseID
			warehouseCode = invWarehouseCode
			if _, err = tx.Exec(ctx, `UPDATE return_order SET warehouse_id = $1, warehouse_code = $2, updated_at = NOW(), updated_by = $3 WHERE id = $4`, warehouseID, warehouseCode, userID, id); err != nil {
				return err
			}
		}

		itemQuery := `
				INSERT INTO return_order_item (return_id, inventory_id, material_id, sku_id, quantity, unit, warehouse_id, warehouse_code)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`
		if _, err = tx.Exec(ctx, itemQuery, id, item.InventoryID, materialID, skuID, item.Quantity, unit, invWarehouseID, invWarehouseCode); err != nil {
			return err
		}
	}

	cmdTag, err := tx.Exec(ctx, `
		UPDATE return_order
		SET return_date = $1,
		    remark = $2,
		    supplier_code = NULLIF($3, ''),
		    warehouse_id = $4,
		    warehouse_code = $5,
		    updated_at = NOW(),
		    updated_by = $6
		WHERE id = $7
		  AND deleted_at IS NULL
		  AND return_status = 'draft'
		  AND return_type = 'purchase_return'
	`, req.ReturnDate, req.Remark, req.SupplierCode, warehouseID, warehouseCode, userID, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("更新失败：单据状态已变更或不存在")
	}

	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	if _, err = tx.Exec(ctx, auditQuery, userID, username, "UPDATE", "RETURN", "return_order", id, nil); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ConfirmReturnOrder 确认退货（调用 SP）实现库存扣减或回库
func ConfirmReturnOrder(ctx context.Context, returnID int64, userID int64, username string) error {
	var returnType string
	err := database.Pool.QueryRow(ctx, `SELECT return_type FROM return_order WHERE id = $1 AND deleted_at IS NULL`, returnID).Scan(&returnType)
	if err != nil {
		return err
	}

	query := `CALL sp_confirm_return_order($1, $2)`
	if returnType == "purchase_return" {
		query = `CALL sp_submit_purchase_return($1, $2)`
	}

	_, err = database.Pool.Exec(ctx, query, returnID, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return fmt.Errorf("DB_ERROR: %s", pgErr.Message)
		}
		return err
	}

	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, _ = database.Pool.Exec(ctx, auditQuery, userID, username, "CONFIRM", "RETURN", "return_order", returnID, nil)

	return nil
}

func DeleteReturnOrder(ctx context.Context, id int64, userID int64, username string) error {
	cmdTag, err := database.Pool.Exec(ctx, `
		UPDATE return_order
		SET deleted_at = NOW(),
		    updated_at = NOW(),
		    updated_by = $2
		WHERE id = $1
		  AND deleted_at IS NULL
		  AND return_status = 'draft'
	`, id, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("仅允许删除待提交状态的退货单")
	}

	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, _ = database.Pool.Exec(ctx, auditQuery, userID, username, "DELETE", "RETURN", "return_order", id, nil)
	return nil
}
