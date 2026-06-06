/**
 * 功能：入库单数据库操作
 * 创建时间：2026-04-18
 * 创建人：CodeArts Agent
 */

package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

func ListStockIns(ctx context.Context, q *request.StockInQuery) ([]response.StockInResp, int64, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if q.StockInNo != "" {
		where = append(where, fmt.Sprintf("si.stock_in_no ILIKE $%d", argID))
		args = append(args, "%"+q.StockInNo+"%")
		argID++
	}
	if q.WarehouseCode != "" {
		where = append(where, fmt.Sprintf("si.warehouse_code = $%d", argID))
		args = append(args, q.WarehouseCode)
		argID++
	}
	if q.StockInType != "" {
		if q.StockInType == "return" {
			q.StockInType = "sales_return"
		}
		where = append(where, fmt.Sprintf("COALESCE(sbi.stock_in_type, 'purchase') = $%d", argID))
		args = append(args, q.StockInType)
		argID++
	}
	if q.SupplierCode != "" {
		where = append(where, fmt.Sprintf("si.supplier_code = $%d", argID))
		args = append(args, q.SupplierCode)
		argID++
	}
	if q.Status != "" {
		if q.Status == "partial" {
			where = append(where, "si.stock_in_status IN ('preparing', 'pending')")
		} else if q.Status == "preparing" {
			where = append(where, "si.stock_in_status = 'preparing'")
		} else if q.Status == "pending" {
			where = append(where, "si.stock_in_status = 'pending'")
		} else {
			where = append(where, fmt.Sprintf("si.stock_in_status = $%d", argID))
			args = append(args, q.Status)
			argID++
		}
	}
	if q.StartDate != "" {
		where = append(where, fmt.Sprintf("si.stock_in_date::date >= $%d::date", argID))
		args = append(args, q.StartDate)
		argID++
	}
	if q.EndDate != "" {
		where = append(where, fmt.Sprintf("si.stock_in_date::date <= $%d::date", argID))
		args = append(args, q.EndDate)
		argID++
	}
	whereClause := strings.Join(where, " AND ")

	const listFrom = `FROM v_stock_in_list si LEFT JOIN stock_in sbi ON sbi.id = si.id`

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) %s WHERE %s", listFrom, whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT si.id, si.stock_in_no, COALESCE(sbi.stock_in_type, 'purchase'),
		       CASE
		           WHEN sbi.purchase_order_id IS NOT NULL THEN 'purchase_order'
		           WHEN ro.id IS NOT NULL THEN 'reversal_order'
		           WHEN sro.id IS NOT NULL THEN 'sales_return'
		           ELSE ''
		       END AS business_doc_type,
		       COALESCE(sbi.purchase_order_id, ro.id, sro.id, 0) AS business_doc_id,
		       COALESCE(po.order_no, ro.order_no, sro.return_no, '') AS business_doc_no,
		       si.warehouse_code, si.warehouse_name,
		       si.stock_in_date,
		       si.supplier_code, si.supplier_name,
		       COALESCE(sbi.purchase_order_id, 0), si.purchase_order_no,
		       COALESCE(ro.id, 0), COALESCE(ro.order_no, ''),
		       COALESCE((
		           SELECT SUM(COALESCE(sii.accepted_quantity, 0) * COALESCE(sii.unit_cost, 0))
		           FROM stock_in_item sii
		           WHERE sii.stock_in_id = si.id
		       ), 0) AS total_amount,
		       si.stock_in_status,
		       COALESCE(NULLIF(TRIM(ro.remark), ''), si.remark) AS remark,
		       si.created_at
		%s
		LEFT JOIN reversal_order ro ON ro.stock_in_id = si.id AND ro.deleted_at IS NULL
		LEFT JOIN return_order sro ON sro.stock_in_id = si.id AND sro.return_type = 'sales_return' AND sro.deleted_at IS NULL
		LEFT JOIN purchase_order po ON po.id = sbi.purchase_order_id AND po.deleted_at IS NULL
		WHERE %s
		ORDER BY si.id DESC
		LIMIT $%d OFFSET $%d
	`, listFrom, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.StockInResp
	for rows.Next() {
		var item response.StockInResp
		if err := rows.Scan(&item.ID, &item.StockInNo, &item.StockInType,
			&item.BusinessDocType, &item.BusinessDocID, &item.BusinessDocNo,
			&item.WarehouseCode, &item.WarehouseName,
			&item.StockInDate,
			&item.SupplierCode, &item.SupplierName, &item.PurchaseOrderID, &item.PurchaseOrderNo,
			&item.ReversalOrderID, &item.ReversalOrderNo, &item.TotalAmount, &item.StockInStatus,
			&item.Remark, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

func GetStockInByID(ctx context.Context, id int64) (*response.StockInDetailResp, error) {
	var stockIn response.StockInDetailResp
	query := `
		SELECT si.id, si.stock_in_no, COALESCE(sbi.stock_in_type, 'purchase'),
		       CASE
		           WHEN sbi.purchase_order_id IS NOT NULL THEN 'purchase_order'
		           WHEN ro.id IS NOT NULL THEN 'reversal_order'
		           WHEN sro.id IS NOT NULL THEN 'sales_return'
		           ELSE ''
		       END AS business_doc_type,
		       COALESCE(sbi.purchase_order_id, ro.id, sro.id, 0) AS business_doc_id,
		       COALESCE(po.order_no, ro.order_no, sro.return_no, '') AS business_doc_no,
		       si.warehouse_code, si.warehouse_name,
		       si.stock_in_date,
		       si.supplier_code, si.supplier_name,
		       COALESCE(sbi.purchase_order_id, 0), si.purchase_order_no,
		       COALESCE(ro.id, 0), COALESCE(ro.order_no, ''),
		       COALESCE((
		           SELECT SUM(COALESCE(sii.accepted_quantity, 0) * COALESCE(sii.unit_cost, 0))
		           FROM stock_in_item sii
		           WHERE sii.stock_in_id = si.id
		       ), 0) AS total_amount,
		       si.stock_in_status,
		       COALESCE(NULLIF(TRIM(ro.remark), ''), si.remark) AS remark,
		       si.created_at
		FROM v_stock_in_list si
		LEFT JOIN stock_in sbi ON sbi.id = si.id
		LEFT JOIN reversal_order ro ON ro.stock_in_id = si.id AND ro.deleted_at IS NULL
		LEFT JOIN return_order sro ON sro.stock_in_id = si.id AND sro.return_type = 'sales_return' AND sro.deleted_at IS NULL
		LEFT JOIN purchase_order po ON po.id = sbi.purchase_order_id AND po.deleted_at IS NULL
		WHERE si.id = $1
	`
	err := database.Pool.QueryRow(ctx, query, id).Scan(
		&stockIn.ID, &stockIn.StockInNo, &stockIn.StockInType,
		&stockIn.BusinessDocType, &stockIn.BusinessDocID, &stockIn.BusinessDocNo,
		&stockIn.WarehouseCode, &stockIn.WarehouseName,
		&stockIn.StockInDate,
		&stockIn.SupplierCode, &stockIn.SupplierName, &stockIn.PurchaseOrderID, &stockIn.PurchaseOrderNo,
		&stockIn.ReversalOrderID, &stockIn.ReversalOrderNo, &stockIn.TotalAmount, &stockIn.StockInStatus,
		&stockIn.Remark, &stockIn.CreatedAt)
	if err != nil {
		return nil, err
	}

	err = database.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM inventory_transaction WHERE ref_doc_type = 'stock_in' AND ref_doc_id = $1)
	`, id).Scan(&stockIn.HasStockIn)
	if err != nil {
		return nil, err
	}

	itemQuery := `
		SELECT sii.id, sii.material_id, m.material_code, m.material_name, m.is_code,
		       COALESCE(poi.quantity, sii.arrived_quantity) AS purchase_quantity,
		       CASE
		           WHEN COALESCE(sbi.stock_in_type, '') = 'reversal' THEN sii.accepted_quantity
		           ELSE COALESCE(poi.received_quantity, 0)
		       END AS received_quantity,
		       sii.arrived_quantity, sii.accepted_quantity,
		       GREATEST(COALESCE(poi.quantity - poi.received_quantity, sii.accepted_quantity), 0) AS pending_quantity,
		       sii.unit_cost
		FROM stock_in_item sii
		INNER JOIN stock_in sbi ON sbi.id = sii.stock_in_id AND sbi.deleted_at IS NULL
		JOIN material m ON m.id = sii.material_id
		LEFT JOIN purchase_order_item poi ON poi.material_id = sii.material_id
			AND poi.order_id = (SELECT purchase_order_id FROM stock_in WHERE id = $1)
		WHERE sii.stock_in_id = $1
	`
	rows, err := database.Pool.Query(ctx, itemQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item response.StockInItemResp
		if err := rows.Scan(&item.ID, &item.MaterialID, &item.MaterialCode, &item.MaterialName, &item.IsCode,
			&item.PurchaseQuantity, &item.ReceivedQuantity,
			&item.ArrivedQuantity, &item.AcceptedQuantity,
			&item.PendingQuantity,
			&item.UnitCost); err != nil {
			return nil, err
		}
		stockIn.Items = append(stockIn.Items, item)
	}

	return &stockIn, rows.Err()
}

func CreateStockIn(ctx context.Context, req *request.CreateStockInReq, userID int64) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var warehouseID int64
	err = tx.QueryRow(ctx, `
		SELECT id FROM warehouse
		WHERE warehouse_code = $1 AND deleted_at IS NULL
	`, req.WarehouseCode).Scan(&warehouseID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("仓库不存在或已删除")
		}
		return 0, err
	}

	stockInNo, err := generateStockInNo(ctx, tx)
	if err != nil {
		return 0, err
	}

	var stockInID int64
	stockInQuery := `
		INSERT INTO stock_in (stock_in_no, warehouse_id, warehouse_code, stock_in_date, stock_in_type, purchase_order_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	err = tx.QueryRow(ctx, stockInQuery, stockInNo, warehouseID, req.WarehouseCode, req.StockInDate, req.StockInType, req.PurchaseOrderID, userID).Scan(&stockInID)
	if err != nil {
		return 0, err
	}

	itemQuery := `
		INSERT INTO stock_in_item (stock_in_id, material_id, arrived_quantity, accepted_quantity, unit_cost)
		VALUES ($1, $2, $3, $4, $5)
	`
	for _, item := range req.Items {
		_, err := tx.Exec(ctx, itemQuery, stockInID, item.MaterialID, item.ArrivedQuantity, item.AcceptedQuantity, item.UnitPrice)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return stockInID, nil
}

func ConfirmStockIn(ctx context.Context, stockInID, userID int64) error {
	_, err := database.Pool.Exec(ctx, `CALL sp_confirm_stock_in($1, $2)`, stockInID, userID)
	return err
}

// ConfirmReversalStockIn 退料入库确认：
// - 必须 stock_in_type = 'reversal'
// - 必须已备货编码（选中数量=accepted_quantity）
// - 将 material_serial_code 从 issued -> in_stock，并回补 inventory.quantity
// - 入库单直接置为 passed，同时把关联退料单置为已完成
func ConfirmReversalStockIn(ctx context.Context, stockInID, userID int64) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var stockInType string
	var stockInNo string
	var stockInStatus string
	var warehouseID int64
	if err := tx.QueryRow(ctx, `
		SELECT stock_in_type, stock_in_no, stock_in_status, COALESCE(warehouse_id, 0)
		FROM stock_in
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, stockInID).Scan(&stockInType, &stockInNo, &stockInStatus, &warehouseID); err != nil {
		return err
	}
	if stockInType != "reversal" {
		return fmt.Errorf("仅退料入库单支持该确认方式")
	}
	if warehouseID <= 0 {
		return fmt.Errorf("退料入库单未设置仓库，无法确认入库")
	}
	if stockInStatus == "passed" {
		return fmt.Errorf("入库单已完成，无需重复确认")
	}

	// 逐行校验并确认
	rows, err := tx.Query(ctx, `
		SELECT sii.id, sii.material_id, sii.accepted_quantity, m.is_code
		FROM stock_in_item sii
		INNER JOIN material m ON m.id = sii.material_id
		WHERE sii.stock_in_id = $1
		ORDER BY sii.id ASC
	`, stockInID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type line struct {
		ItemID     int64
		MaterialID int64
		Need       int64
		IsCode     bool
	}
	lines := make([]line, 0)
	for rows.Next() {
		var it line
		var accepted float64
		if err := rows.Scan(&it.ItemID, &it.MaterialID, &accepted, &it.IsCode); err != nil {
			return err
		}
		it.Need = int64(accepted)
		lines = append(lines, it)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, ln := range lines {
		if ln.Need <= 0 {
			continue
		}
		if !ln.IsCode {
			// 无编码物料：直接回补库存数量即可（不需要备货编码）
			// 采用与库存唯一键一致的 upsert（inspection_no 为空）
			var invID int64
			if err := tx.QueryRow(ctx, `
				INSERT INTO inventory (material_id, warehouse_id, quantity, unit, updated_at)
				VALUES (
					$1, $2, $3,
					COALESCE((SELECT unit FROM material WHERE id = $1), ''),
					NOW()
				)
				ON CONFLICT (material_id, warehouse_id, COALESCE(material_inspection_no, ''))
				DO UPDATE SET quantity = inventory.quantity + EXCLUDED.quantity, updated_at = NOW()
				RETURNING id
			`, ln.MaterialID, warehouseID, float64(ln.Need)).Scan(&invID); err != nil {
				return err
			}

			// 入库流水（按明细汇总）
			if _, err := tx.Exec(ctx, `
				INSERT INTO inventory_transaction (material_id, warehouse_id, trans_type, quantity, balance, ref_doc_type, ref_doc_no, ref_doc_id, operator_id, created_at)
				VALUES ($1, $2, 'in', $3,
				        (SELECT quantity FROM inventory WHERE id = $4),
				        'stock_in', $5, $6, $7, NOW())
			`, ln.MaterialID, warehouseID, float64(ln.Need), invID, stockInNo, stockInID, userID); err != nil {
				return err
			}
			continue
		}

		var selectedCnt int64
		if err := tx.QueryRow(ctx, `
			SELECT count(*)
			FROM stock_in_item_serial_selection
			WHERE stock_in_item_id = $1
		`, ln.ItemID).Scan(&selectedCnt); err != nil {
			return err
		}
		if selectedCnt != ln.Need {
			return fmt.Errorf("退料入库需备货编码：明细[id=%d] 需要%d个，已选%d个", ln.ItemID, ln.Need, selectedCnt)
		}

		// 先读完编码ID并关闭 rows，避免同连接并发查询导致 conn busy
		srows, err := tx.Query(ctx, `
			SELECT serial_code_id
			FROM stock_in_item_serial_selection
			WHERE stock_in_item_id = $1
			ORDER BY serial_code_id ASC
		`, ln.ItemID)
		if err != nil {
			return err
		}
		serialIDs := make([]int64, 0, ln.Need)
		for srows.Next() {
			var serialID int64
			if err := srows.Scan(&serialID); err != nil {
				srows.Close()
				return err
			}
			serialIDs = append(serialIDs, serialID)
		}
		if err := srows.Err(); err != nil {
			srows.Close()
			return err
		}
		srows.Close()

		// 再逐个锁定并处理编码
		for _, serialID := range serialIDs {
			var invID int64
			var status string
			var serialCode string
			if err := tx.QueryRow(ctx, `
				SELECT inventory_id, status, serial_code
				FROM material_serial_code
				WHERE id = $1
				FOR UPDATE
			`, serialID).Scan(&invID, &status, &serialCode); err != nil {
				return err
			}
			if status != "issued" {
				return fmt.Errorf("编码状态不允许退回（需要已出库/已领用）：%s", serialCode)
			}

			// 回补库存
			if _, err := tx.Exec(ctx, `UPDATE inventory SET quantity = quantity + 1, updated_at = NOW() WHERE id = $1`, invID); err != nil {
				return err
			}
			// 编码回库
			if _, err := tx.Exec(ctx, `
				UPDATE material_serial_code
				SET status = 'in_stock',
				    stock_in_id = $2,
				    stock_in_item_id = $3,
				    updated_at = NOW()
				WHERE id = $1
			`, serialID, stockInID, ln.ItemID); err != nil {
				return err
			}
			// 追溯记录
			if _, err := tx.Exec(ctx, `
				INSERT INTO material_serial_trace (serial_code_id, serial_code, action, ref_doc_type, ref_doc_no, ref_doc_id, operator_id, remark)
				VALUES ($1, $2, 'stock_in', 'stock_in', $3, $4, $5, '退料入库确认回库')
			`, serialID, serialCode, stockInNo, stockInID, userID); err != nil {
				return err
			}
		}

		// 入库流水（按明细汇总）
		if _, err := tx.Exec(ctx, `
			INSERT INTO inventory_transaction (material_id, warehouse_id, trans_type, quantity, balance, ref_doc_type, ref_doc_no, ref_doc_id, operator_id, created_at)
			SELECT $1, i.warehouse_id, 'in', $2,
			       (SELECT quantity FROM inventory WHERE id = i.id),
			       'stock_in', $3, $4, $5, NOW()
			FROM inventory i
			WHERE i.material_id = $1
			ORDER BY i.id ASC
			LIMIT 1
		`, ln.MaterialID, ln.Need, stockInNo, stockInID, userID); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `
		UPDATE stock_in si
		SET stock_in_status = 'passed',
		    remark = COALESCE(
		        NULLIF(TRIM((
		            SELECT remark FROM reversal_order
		            WHERE stock_in_id = si.id AND deleted_at IS NULL
		            ORDER BY id LIMIT 1
		        )), ''),
		        si.remark
		    ),
		    updated_by = $1,
		    updated_at = NOW()
		WHERE si.id = $2 AND si.deleted_at IS NULL
	`, userID, stockInID); err != nil {
		return err
	}
	// 关联退料单置为已完成
	if _, err := tx.Exec(ctx, `UPDATE reversal_order SET status = 'completed', updated_at = NOW() WHERE stock_in_id = $1 AND deleted_at IS NULL`, stockInID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func ListStockInConfirmLogs(ctx context.Context, stockInID int64) ([]response.StockInConfirmLogResp, error) {
	rows, err := database.Pool.Query(ctx, `
		SELECT l.id, l.stock_in_id, l.stock_in_item_id, l.material_id,
		       COALESCE(m.material_code, ''), COALESCE(m.material_name, ''),
		       l.purchase_quantity, l.before_received_quantity, l.current_received_quantity, l.after_received_quantity,
		       l.operator_id, COALESCE(u.real_name, u.username, ''),
		       l.created_at, COALESCE(l.remark, '')
		FROM stock_in_confirm_log l
		LEFT JOIN material m ON m.id = l.material_id
		LEFT JOIN sys_user u ON u.id = l.operator_id
		WHERE l.stock_in_id = $1
		ORDER BY l.created_at DESC, l.id DESC
	`, stockInID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]response.StockInConfirmLogResp, 0)
	for rows.Next() {
		var item response.StockInConfirmLogResp
		if err := rows.Scan(
			&item.ID, &item.StockInID, &item.StockInItemID, &item.MaterialID,
			&item.MaterialCode, &item.MaterialName,
			&item.PurchaseQuantity, &item.BeforeReceivedQuantity, &item.CurrentReceivedQuantity, &item.AfterReceivedQuantity,
			&item.OperatorID, &item.OperatorName,
			&item.CreatedAt, &item.Remark,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, rows.Err()
}

func UpdateStockIn(ctx context.Context, id int64, warehouseCode string, _ string, items []request.UpdateStockInItem) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var hasStockIn bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM inventory_transaction WHERE ref_doc_type = 'stock_in' AND ref_doc_id = $1)
	`, id).Scan(&hasStockIn)
	if err != nil {
		return err
	}

	if warehouseCode != "" {
		if hasStockIn {
			var currentCode database.NullString
			if err := tx.QueryRow(ctx, `SELECT warehouse_code FROM stock_in WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&currentCode); err != nil {
				return err
			}
			cc := currentCode.String()
			// 已发生入库流水后，允许重复提交同一仓库；但禁止变更仓库
			if cc != "" && cc != warehouseCode {
				return fmt.Errorf("已部分入库，仓库不可修改")
			}
			// 仓库一致则不再重复更新
			goto UPDATE_ITEMS
		}
		var warehouseID int64
		err = tx.QueryRow(ctx, `
			SELECT id FROM warehouse
			WHERE warehouse_code = $1 AND deleted_at IS NULL
		`, warehouseCode).Scan(&warehouseID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return fmt.Errorf("仓库不存在或已删除")
			}
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE stock_in SET warehouse_id = $1, warehouse_code = $2, updated_at = NOW() WHERE id = $3`, warehouseID, warehouseCode, id)
	}
	if err != nil {
		return err
	}

UPDATE_ITEMS:
	for _, item := range items {
		_, err := tx.Exec(ctx, `
			UPDATE stock_in_item
			SET arrived_quantity = $1, accepted_quantity = $2, unit_cost = $3
			WHERE id = $4 AND stock_in_id = $5
		`, item.ArrivedQuantity, item.AcceptedQuantity, item.UnitCost, item.ID, id)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func generateStockInNo(ctx context.Context, tx pgx.Tx) (string, error) {
	var stockInNo string
	query := `
		SELECT TO_CHAR(NOW(), 'YYYYMMDD') || LPAD(nextval('stock_in_no_seq')::TEXT, 3, '0')
	`
	err := tx.QueryRow(ctx, query).Scan(&stockInNo)
	if err != nil {
		return "", err
	}
	return "SI" + stockInNo, nil
}
