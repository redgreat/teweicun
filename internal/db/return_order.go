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
	if q.CustomerCode != "" {
		where = append(where, fmt.Sprintf("ro.customer_code = $%d", argID))
		args = append(args, q.CustomerCode)
		argID++
	}
	if strings.TrimSpace(q.CustomerKeyword) != "" {
		where = append(where, fmt.Sprintf("(COALESCE(ro.customer_name, c.customer_name, '') ILIKE $%d OR ro.customer_code ILIKE $%d)", argID, argID))
		args = append(args, "%"+strings.TrimSpace(q.CustomerKeyword)+"%")
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
	countQuery := fmt.Sprintf(`
		SELECT count(*)
		FROM return_order ro
		LEFT JOIN customer c ON c.customer_code = ro.customer_code AND c.deleted_at IS NULL
		WHERE %s
	`, whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT ro.id, ro.return_no, ro.return_type, COALESCE(ro.ref_doc_type, ''), 
		       ro.ref_doc_id, ro.warehouse_code, w.warehouse_name,
		       COALESCE(ro.supplier_code, ''), COALESCE(s.supplier_name, ''),
		       COALESCE(ro.customer_code, ''), COALESCE(ro.customer_name, c.customer_name, ''),
		       COALESCE(ro.stock_out_id, 0), COALESCE(so.stock_out_no, ''),
		       COALESCE(ro.stock_in_id, 0), COALESCE(si.stock_in_no, ''),
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
		LEFT JOIN customer c ON c.customer_code = ro.customer_code AND c.deleted_at IS NULL
		LEFT JOIN stock_out so ON so.id = ro.stock_out_id AND so.deleted_at IS NULL
		LEFT JOIN stock_in si ON si.id = ro.stock_in_id AND si.deleted_at IS NULL
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
			&item.CustomerCode, &item.CustomerName,
			&item.StockOutID, &item.StockOutNo,
			&item.StockInID, &item.StockInNo,
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
		       COALESCE(ro.customer_code, ''), COALESCE(ro.customer_name, c.customer_name, ''),
		       COALESCE(ro.stock_out_id, 0), COALESCE(so.stock_out_no, ''),
		       COALESCE(ro.stock_in_id, 0), COALESCE(si.stock_in_no, ''),
		       ro.return_date, ro.return_status, COALESCE(ro.remark, ''), ro.created_at
		FROM return_order ro
		LEFT JOIN warehouse w ON w.warehouse_code = ro.warehouse_code
		LEFT JOIN supplier s ON s.supplier_code = ro.supplier_code
		LEFT JOIN customer c ON c.customer_code = ro.customer_code AND c.deleted_at IS NULL
		LEFT JOIN stock_out so ON so.id = ro.stock_out_id AND so.deleted_at IS NULL
		LEFT JOIN stock_in si ON si.id = ro.stock_in_id AND si.deleted_at IS NULL
		WHERE ro.id = $1 AND ro.deleted_at IS NULL
	`
	var item response.ReturnOrderResp
	err := database.Pool.QueryRow(ctx, query, id).Scan(&item.ID, &item.ReturnNo, &item.ReturnType, &item.RefDocType, &item.RefDocID,
		&item.WarehouseCode, &item.WarehouseName,
		&item.SupplierCode, &item.SupplierName,
		&item.CustomerCode, &item.CustomerName,
		&item.StockOutID, &item.StockOutNo,
		&item.StockInID, &item.StockInNo,
		&item.ReturnDate, &item.Status, &item.Remark, &item.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 查明细（含当前可用量、单价，供编辑页校验与展示）
	itemQuery := `
		SELECT roi.id, roi.material_id, m.material_code, m.material_name,
		       COALESCE(roi.inventory_id, 0),
		       COALESCE(roi.warehouse_code, ''), COALESCE(wh_line.warehouse_name, ''),
		       roi.quantity, roi.unit,
		       COALESCE(inv.quantity - inv.locked_quantity - COALESCE(inv.in_transit_quantity, 0), 0),
		       COALESCE(inv.unit_cost, 0)
		FROM return_order_item roi
		INNER JOIN material m ON m.id = roi.material_id
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
	if req.ReturnType == "sales_return" && strings.TrimSpace(req.CustomerCode) == "" {
		return 0, fmt.Errorf("销售退货必须选择客户")
	}
	var warehouseID int64
	warehouseCode := strings.TrimSpace(req.WarehouseCode)
	customerCode := strings.TrimSpace(req.CustomerCode)
	customerName := ""

	if req.ReturnType == "sales_return" {
		if warehouseCode == "" {
			return 0, fmt.Errorf("销售退货必须选择入库仓库")
		}
		if err := tx.QueryRow(ctx, `
			SELECT id
			FROM warehouse
			WHERE warehouse_code = $1
			  AND deleted_at IS NULL
			  AND status = 'enabled'
		`, warehouseCode).Scan(&warehouseID); err != nil {
			if err == pgx.ErrNoRows {
				return 0, fmt.Errorf("退货仓库不存在或已停用")
			}
			return 0, err
		}
		if err := tx.QueryRow(ctx, `
			SELECT customer_name
			FROM customer
			WHERE customer_code = $1
			  AND deleted_at IS NULL
			  AND status = 'enabled'
		`, customerCode).Scan(&customerName); err != nil {
			if err == pgx.ErrNoRows {
				return 0, fmt.Errorf("客户不存在或已停用")
			}
			return 0, err
		}
	}

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
		                         warehouse_id, warehouse_code, supplier_code, customer_code, customer_name,
		                         return_date, return_status, remark, created_by)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, 0), $5, $6, NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''),
		        $10, 'draft', $11, $12)
		RETURNING id
	`
	err = tx.QueryRow(ctx, mainQuery, returnNo, req.ReturnType, req.RefDocType, req.RefDocID,
		warehouseID, warehouseCode, req.SupplierCode, customerCode, customerName,
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
			var unit string
			var invWarehouseID int64
			var invWarehouseCode string

			err = tx.QueryRow(ctx, `
				SELECT i.material_id, i.unit,
				       i.warehouse_id, w.warehouse_code
				FROM inventory i
				INNER JOIN warehouse w ON w.id = i.warehouse_id
				WHERE i.id = $1
			`, item.InventoryID).Scan(&materialID, &unit, &invWarehouseID, &invWarehouseCode)
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
				INSERT INTO return_order_item (return_id, inventory_id, material_id, quantity, unit, warehouse_id, warehouse_code)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`
			_, err = tx.Exec(ctx, itemQuery, id, item.InventoryID, materialID, item.Quantity, unit, invWarehouseID, invWarehouseCode)
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

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	if req.ReturnType == "purchase_return" {
		if err := ConfirmReturnOrder(ctx, id, userID, username); err != nil {
			return 0, fmt.Errorf("退货单创建成功但自动提交失败: %w", err)
		}
	}

	return id, nil
}

// UpdateReturnOrder 更新采购退货单（待出库状态可编辑）
func UpdateReturnOrder(ctx context.Context, id int64, req *request.UpdateReturnOrderReq, userID int64, username string) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var rt string
	var st string
	var stockOutID int64
	err = tx.QueryRow(ctx, `
		SELECT return_type, return_status, COALESCE(stock_out_id, 0)
		FROM return_order WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, id).Scan(&rt, &st, &stockOutID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("退货单不存在")
		}
		return err
	}
	if rt != "purchase_return" {
		return fmt.Errorf("仅支持编辑采购退货单")
	}
	if st != "pending_out" {
		return fmt.Errorf("仅待出库状态可编辑")
	}
	if strings.TrimSpace(req.SupplierCode) == "" {
		return fmt.Errorf("采购退货必须选择供应商")
	}

	if stockOutID > 0 {
		var stockOutStatus string
		err = tx.QueryRow(ctx, `SELECT COALESCE(status, 'pending') FROM stock_out WHERE id = $1`, stockOutID).Scan(&stockOutStatus)
		if err != nil && err != pgx.ErrNoRows {
			return err
		}
		if stockOutStatus != "" && stockOutStatus != "pending" {
			return fmt.Errorf("出库单已操作，不可编辑")
		}

		if _, err := tx.Exec(ctx, `
			UPDATE inventory
			SET in_transit_quantity = GREATEST(COALESCE(in_transit_quantity, 0) - src.qty, 0),
			    updated_at = NOW()
			FROM (
				SELECT soi.inventory_id, SUM(soi.quantity) AS qty
				FROM stock_out_item soi
				WHERE soi.stock_out_id = $1 AND COALESCE(soi.inventory_id, 0) <> 0
				GROUP BY soi.inventory_id
			) src
			WHERE inventory.id = src.inventory_id
		`, stockOutID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `DELETE FROM stock_out_item WHERE stock_out_id = $1`, stockOutID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM stock_out WHERE id = $1`, stockOutID); err != nil {
			return err
		}
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
		var unit string
		var invWarehouseID int64
		var invWarehouseCode string

		err = tx.QueryRow(ctx, `
			SELECT i.material_id, i.unit, i.warehouse_id, w.warehouse_code
			FROM inventory i
			INNER JOIN warehouse w ON w.id = i.warehouse_id
			WHERE i.id = $1
		`, item.InventoryID).Scan(&materialID, &unit, &invWarehouseID, &invWarehouseCode)
		if err != nil {
			if err == pgx.ErrNoRows {
				return fmt.Errorf("库存不存在")
			}
			return err
		}

		if warehouseID == 0 {
			warehouseID = invWarehouseID
			warehouseCode = invWarehouseCode
		}

		if _, err = tx.Exec(ctx, `
			INSERT INTO return_order_item (return_id, inventory_id, material_id, quantity, unit, warehouse_id, warehouse_code)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, id, item.InventoryID, materialID, item.Quantity, unit, invWarehouseID, invWarehouseCode); err != nil {
			return err
		}
	}

	if _, err = tx.Exec(ctx, `
		UPDATE return_order
		SET return_date = $1, remark = $2, supplier_code = NULLIF($3, ''),
		    warehouse_id = $4, warehouse_code = $5, stock_out_id = NULL,
		    updated_at = NOW(), updated_by = $6
		WHERE id = $7
	`, req.ReturnDate, req.Remark, req.SupplierCode, warehouseID, warehouseCode, userID, id); err != nil {
		return err
	}

	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, _ = tx.Exec(ctx, auditQuery, userID, username, "UPDATE", "RETURN", "return_order", id, nil)

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return ConfirmReturnOrder(ctx, id, userID, username)
}

// UpdateSalesReturnOrder 更新销售退货单（待入库状态可编辑）
func UpdateSalesReturnOrder(ctx context.Context, id int64, req *request.UpdateSalesReturnOrderReq, userID int64, username string) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var rt string
	var st string
	var stockInID int64
	err = tx.QueryRow(ctx, `
		SELECT return_type, return_status, COALESCE(stock_in_id, 0)
		FROM return_order WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, id).Scan(&rt, &st, &stockInID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("退货单不存在")
		}
		return err
	}
	if rt != "sales_return" {
		return fmt.Errorf("仅支持编辑销售退货单")
	}
	if st != "draft" && st != "confirmed" {
		return fmt.Errorf("仅待提交或待入库状态可编辑")
	}

	if stockInID > 0 {
		if _, err := tx.Exec(ctx, `DELETE FROM stock_in_item WHERE stock_in_id = $1`, stockInID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM stock_in WHERE id = $1`, stockInID); err != nil {
			return err
		}
	}

	if _, err = tx.Exec(ctx, `DELETE FROM return_order_item WHERE return_id = $1`, id); err != nil {
		return err
	}

	warehouseCode := strings.TrimSpace(req.WarehouseCode)
	var warehouseID int64
	if err := tx.QueryRow(ctx, `
		SELECT id FROM warehouse
		WHERE warehouse_code = $1 AND deleted_at IS NULL AND status = 'enabled'
	`, warehouseCode).Scan(&warehouseID); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("退货仓库不存在或已停用")
		}
		return err
	}

	customerCode := strings.TrimSpace(req.CustomerCode)
	var customerName string
	if err := tx.QueryRow(ctx, `
		SELECT customer_name FROM customer
		WHERE customer_code = $1 AND deleted_at IS NULL AND status = 'enabled'
	`, customerCode).Scan(&customerName); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("客户不存在或已停用")
		}
		return err
	}

	for _, item := range req.Items {
		if item.MaterialID == 0 {
			return fmt.Errorf("销售退货明细必须填写物料")
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO return_order_item (return_id, material_id, quantity, unit)
			VALUES ($1, $2, $3, (SELECT unit FROM material WHERE id = $2))
		`, id, item.MaterialID, item.Quantity); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `
		UPDATE return_order
		SET return_date = $1, remark = $2,
		    customer_code = NULLIF($3, ''), customer_name = NULLIF($4, ''),
		    warehouse_id = $5, warehouse_code = $6, stock_in_id = NULL,
		    updated_at = NOW(), updated_by = $7
		WHERE id = $8
	`, req.ReturnDate, req.Remark, customerCode, customerName, warehouseID, warehouseCode, userID, id); err != nil {
		return err
	}

	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, _ = tx.Exec(ctx, auditQuery, userID, username, "UPDATE", "RETURN", "return_order", id, nil)

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return ConfirmReturnOrder(ctx, id, userID, username)
}

// ConfirmReturnOrder 确认退货（调用 SP）实现库存扣减或回库
func ConfirmReturnOrder(ctx context.Context, returnID int64, userID int64, username string) error {
	var returnType string
	var returnStatus string
	var stockOutID int64
	var stockInID int64
	err := database.Pool.QueryRow(ctx, `
		SELECT return_type, COALESCE(return_status, 'draft'),
		       COALESCE(stock_out_id, 0), COALESCE(stock_in_id, 0)
		FROM return_order
		WHERE id = $1 AND deleted_at IS NULL
	`, returnID).Scan(&returnType, &returnStatus, &stockOutID, &stockInID)
	if err != nil {
		return err
	}

	query := `CALL sp_confirm_return_order($1, $2)`
	if returnType == "purchase_return" {
		if returnStatus == "pending_out" || stockOutID > 0 {
			return nil
		}
		query = `CALL sp_submit_purchase_return($1, $2)`
	} else if returnType == "sales_return" && (returnStatus == "confirmed" || returnStatus == "completed" || stockInID > 0) {
		return nil
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
