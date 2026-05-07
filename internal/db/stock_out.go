/**
 * 功能：stock_out.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package db

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

// ListStockOuts 分页查询出库单
func ListStockOuts(ctx context.Context, q *request.StockOutQuery) ([]response.StockOutResp, int64, error) {
	where := []string{"so.deleted_at IS NULL"}
	var args []interface{}
	argID := 1

	if q.StockOutNo != "" {
		where = append(where, fmt.Sprintf("so.stock_out_no ILIKE $%d", argID))
		args = append(args, "%"+q.StockOutNo+"%")
		argID++
	}
	if q.OutType != "" {
		where = append(where, fmt.Sprintf("so.out_type = $%d", argID))
		args = append(args, q.OutType)
		argID++
	}
	if q.RefDocType != "" {
		where = append(where, fmt.Sprintf("so.ref_doc_type = $%d", argID))
		args = append(args, q.RefDocType)
		argID++
	}
	if q.WarehouseCode != "" {
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM stock_out_item soi
			INNER JOIN inventory i ON i.id = soi.inventory_id AND COALESCE(soi.inventory_id, 0) <> 0
			INNER JOIN warehouse w ON w.id = i.warehouse_id AND w.deleted_at IS NULL
			WHERE soi.stock_out_id = so.id AND w.warehouse_code = $%d
		)`, argID))
		args = append(args, q.WarehouseCode)
		argID++
	}
	if q.Status != "" {
		if q.Status == "pending" {
			where = append(where, "so.status IN ('draft', 'pending')")
		} else {
			where = append(where, fmt.Sprintf("so.status = $%d", argID))
			args = append(args, q.Status)
			argID++
		}
	}
	if strings.TrimSpace(q.Receiver) != "" {
		where = append(where, fmt.Sprintf("so.receiver ILIKE $%d", argID))
		args = append(args, "%"+strings.TrimSpace(q.Receiver)+"%")
		argID++
	}
	if q.StartDate != "" {
		where = append(where, fmt.Sprintf("so.stock_out_date::date >= $%d::date", argID))
		args = append(args, q.StartDate)
		argID++
	}
	if q.EndDate != "" {
		where = append(where, fmt.Sprintf("so.stock_out_date::date <= $%d::date", argID))
		args = append(args, q.EndDate)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM stock_out so WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT so.id, so.stock_out_no, so.stock_out_date, so.out_type,
		       COALESCE(wh.warehouse_code, ''),
		       COALESCE(wh.warehouse_name, ''),
		       COALESCE(so.ref_doc_type, ''), so.ref_doc_id,
		       COALESCE(so.ref_doc_type, '') AS business_doc_type,
		       COALESCE(so.ref_doc_id, 0) AS business_doc_id,
		       COALESCE(ro.return_no, co.order_no, '') AS business_doc_no,
		       COALESCE(so.receiver, ''), so.status, COALESCE(so.remark, ''), so.created_at, so.confirmed_at,
		       COALESCE((
		           SELECT SUM(soi2.quantity * COALESCE(inv2.unit_cost, 0))
		           FROM stock_out_item soi2
		           LEFT JOIN inventory inv2 ON inv2.id = soi2.inventory_id
		           WHERE soi2.stock_out_id = so.id
		       ), 0) AS total_amount
		FROM stock_out so
		LEFT JOIN LATERAL (
			SELECT w.warehouse_code, w.warehouse_name
			FROM stock_out_item soi
			INNER JOIN inventory i ON i.id = soi.inventory_id AND COALESCE(soi.inventory_id, 0) <> 0
			INNER JOIN warehouse w ON w.id = i.warehouse_id AND w.deleted_at IS NULL
			WHERE soi.stock_out_id = so.id
			ORDER BY soi.id
			LIMIT 1
		) wh ON true
		LEFT JOIN return_order ro ON ro.id = so.ref_doc_id AND so.ref_doc_type = 'purchase_return' AND ro.deleted_at IS NULL
		LEFT JOIN consumption_order co ON co.id = so.ref_doc_id AND so.ref_doc_type = 'consumption_order' AND co.deleted_at IS NULL
		WHERE %s
		ORDER BY so.id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.StockOutResp
	for rows.Next() {
		var item response.StockOutResp
		if err := rows.Scan(&item.ID, &item.StockOutNo, &item.StockOutDate, &item.OutType, &item.WarehouseCode,
			&item.WarehouseName, &item.RefDocType, &item.RefDocID,
			&item.BusinessDocType, &item.BusinessDocID, &item.BusinessDocNo,
			&item.Receiver, &item.Status,
			&item.Remark, &item.CreatedAt, &item.ConfirmedAt, &item.TotalAmount); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// GetStockOutDetail 获取出库单详情
func GetStockOutDetail(ctx context.Context, id int64) (*response.StockOutResp, error) {
	query := `
		SELECT so.id, so.stock_out_no, so.stock_out_date, so.out_type,
		       COALESCE(wh.warehouse_code, ''),
		       COALESCE(wh.warehouse_name, ''),
		       COALESCE(so.ref_doc_type, ''), so.ref_doc_id,
		       COALESCE(so.ref_doc_type, '') AS business_doc_type,
		       COALESCE(so.ref_doc_id, 0) AS business_doc_id,
		       COALESCE(ro.return_no, co.order_no, '') AS business_doc_no,
		       COALESCE(so.receiver, ''), so.status, COALESCE(so.remark, ''), so.created_at, so.confirmed_at,
		       COALESCE((
		           SELECT SUM(soi2.quantity * COALESCE(inv2.unit_cost, 0))
		           FROM stock_out_item soi2
		           LEFT JOIN inventory inv2 ON inv2.id = soi2.inventory_id
		           WHERE soi2.stock_out_id = so.id
		       ), 0) AS total_amount
		FROM stock_out so
		LEFT JOIN LATERAL (
			SELECT w.warehouse_code, w.warehouse_name
			FROM stock_out_item soi
			INNER JOIN inventory i ON i.id = soi.inventory_id AND COALESCE(soi.inventory_id, 0) <> 0
			INNER JOIN warehouse w ON w.id = i.warehouse_id AND w.deleted_at IS NULL
			WHERE soi.stock_out_id = so.id
			ORDER BY soi.id
			LIMIT 1
		) wh ON true
		LEFT JOIN return_order ro ON ro.id = so.ref_doc_id AND so.ref_doc_type = 'purchase_return' AND ro.deleted_at IS NULL
		LEFT JOIN consumption_order co ON co.id = so.ref_doc_id AND so.ref_doc_type = 'consumption_order' AND co.deleted_at IS NULL
		WHERE so.id = $1 AND so.deleted_at IS NULL
	`
	var item response.StockOutResp
	err := database.Pool.QueryRow(ctx, query, id).Scan(&item.ID, &item.StockOutNo, &item.StockOutDate, &item.OutType, &item.WarehouseCode,
		&item.WarehouseName, &item.RefDocType, &item.RefDocID,
		&item.BusinessDocType, &item.BusinessDocID, &item.BusinessDocNo,
		&item.Receiver, &item.Status,
		&item.Remark, &item.CreatedAt, &item.ConfirmedAt, &item.TotalAmount)
	if err != nil {
		return nil, err
	}

	itemQuery := `
		SELECT soi.id, soi.material_id, m.material_code, m.material_name,
		       0::bigint AS sku_id, ''::varchar AS sku_code, ''::varchar AS sku_name,
		       soi.inventory_id, soi.quantity,
		       soi.unit, COALESCE(i.unit_cost, 0),
		       m.is_code
		FROM stock_out_item soi
		INNER JOIN material m ON m.id = soi.material_id
		LEFT JOIN inventory i ON i.id = soi.inventory_id
		WHERE soi.stock_out_id = $1
	`
	rows, err := database.Pool.Query(ctx, itemQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sub response.StockOutItemResp
		if err := rows.Scan(&sub.ID, &sub.MaterialID, &sub.MaterialCode, &sub.MaterialName,
			&sub.SKUID, &sub.SKUCode, &sub.SKUName,
			&sub.InventoryID, &sub.Quantity, &sub.Unit, &sub.UnitCost, &sub.IsCode); err != nil {
			return nil, err
		}
		item.Items = append(item.Items, sub)
	}

	return &item, rows.Err()
}

// CreateStockOut 创建出库单（草稿）
func CreateStockOut(ctx context.Context, req *request.CreateStockOutReq, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// 1. 生成单号
	var stockOutNo string
	err = tx.QueryRow(ctx, "SELECT fn_generate_serial_no('SO')").Scan(&stockOutNo)
	if err != nil {
		return 0, err
	}

	// 2. 插入主表
	var id int64
	mainQuery := `
		INSERT INTO stock_out (stock_out_no, stock_out_date, out_type, ref_doc_type, ref_doc_id,
		                       receiver, status, remark, created_by)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, 0), $6, 'draft', $7, $8)
		RETURNING id
	`
	err = tx.QueryRow(ctx, mainQuery, stockOutNo, req.StockOutDate, req.OutType, req.RefDocType,
		req.RefDocID, req.Receiver, req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
	}

	// 3. 插入明细表
	for _, item := range req.Items {
		itemQuery := `
			INSERT INTO stock_out_item (stock_out_id, material_id, inventory_id,
			                           quantity, unit)
			VALUES ($1, $2, NULLIF($3, 0), $4,
			       (SELECT unit FROM material WHERE id = $2))
		`
		_, err = tx.Exec(ctx, itemQuery, id, item.MaterialID, item.InventoryID,
			item.Quantity)
		if err != nil {
			return 0, err
		}
	}

	// 4. 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "CREATE", "STOCK_OUT", "stock_out", id, nil)
	if err != nil {
		return 0, err
	}

	return id, tx.Commit(ctx)
}

// ConfirmStockOut 确认出库（调用存储过程）
func ConfirmStockOut(ctx context.Context, stockOutID int64, userID int64, username string) error {
	var outType string
	if err := database.Pool.QueryRow(ctx, `SELECT out_type FROM stock_out WHERE id = $1 AND deleted_at IS NULL`, stockOutID).Scan(&outType); err != nil {
		return err
	}
	if outType == "purchase_return" {
		if err := confirmPurchaseReturnStockOut(ctx, stockOutID, userID); err != nil {
			return err
		}
	} else if outType == "consumption" {
		// 领料出库现在也支持手动备货，直接调用存储过程即可，
		// 因为 sp_confirm_stock_out 已被我们重构为支持从暂存表读取编码。
		query := `CALL sp_confirm_stock_out($1, $2)`
		_, err := database.Pool.Exec(ctx, query, stockOutID, userID)
		if err != nil {
			return err
		}
	} else {
		query := `CALL sp_confirm_stock_out($1, $2)`
		_, err := database.Pool.Exec(ctx, query, stockOutID, userID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				return fmt.Errorf("DB_ERROR: %s", pgErr.Message)
			}
			return err
		}
	}

	// 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, _ = database.Pool.Exec(ctx, auditQuery, userID, username, "CONFIRM", "STOCK_OUT", "stock_out", stockOutID, nil)

	return nil
}

func confirmPurchaseReturnStockOut(ctx context.Context, stockOutID, userID int64) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var stockOutNo, status string
	var refDocID *int64
	err = tx.QueryRow(ctx, `
		SELECT stock_out_no, status, ref_doc_id
		FROM stock_out
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, stockOutID).Scan(&stockOutNo, &status, &refDocID)
	if err != nil {
		return err
	}
	if status == "confirmed" {
		return fmt.Errorf("出库单已完成，无需重复确认")
	}

	rows, err := tx.Query(ctx, `
		SELECT soi.id, soi.material_id, soi.inventory_id, soi.quantity, soi.unit, m.is_code
		FROM stock_out_item soi
		JOIN material m ON m.id = soi.material_id
		WHERE soi.stock_out_id = $1
	`, stockOutID)
	if err != nil {
		return err
	}
	type stockOutRow struct {
		itemID      int64
		materialID  int64
		inventoryID int64
		quantity    float64
		unit        string
		isCode      bool
	}
	items := make([]stockOutRow, 0)
	for rows.Next() {
		var r stockOutRow
		if err := rows.Scan(&r.itemID, &r.materialID, &r.inventoryID, &r.quantity, &r.unit, &r.isCode); err != nil {
			rows.Close()
			return err
		}
		items = append(items, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, r := range items {
		itemID, materialID, inventoryID := r.itemID, r.materialID, r.inventoryID
		quantity, isCode := r.quantity, r.isCode
		if inventoryID == 0 {
			return fmt.Errorf("采购退货出库明细必须绑定库存")
		}

		var warehouseID int64
		var invQty, lockedQty, inTransitQty float64
		err = tx.QueryRow(ctx, `
			SELECT warehouse_id, quantity, locked_quantity, COALESCE(in_transit_quantity, 0)
			FROM inventory
			WHERE id = $1
			FOR UPDATE
		`, inventoryID).Scan(&warehouseID, &invQty, &lockedQty, &inTransitQty)
		if err != nil {
			return err
		}

		if (invQty - lockedQty) < quantity {
			return fmt.Errorf("库存不足 [inventory_id=%d]", inventoryID)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE inventory
			SET quantity = quantity - $1,
			    in_transit_quantity = GREATEST(COALESCE(in_transit_quantity, 0) - $1, 0),
			    updated_at = NOW()
			WHERE id = $2
		`, quantity, inventoryID); err != nil {
			return err
		}

		var balance float64
		if err := tx.QueryRow(ctx, `SELECT quantity FROM inventory WHERE id = $1`, inventoryID).Scan(&balance); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO inventory_transaction (
				material_id, warehouse_id, trans_type, quantity, balance,
				ref_doc_type, ref_doc_no, ref_doc_id, operator_id, remark, created_at
			) VALUES ($1, $2, 'out', $3, $4, 'stock_out', $5, $6, $7, '出库确认', NOW())
		`, materialID, warehouseID, -quantity, balance, stockOutNo, stockOutID, userID); err != nil {
			return err
		}

		if isCode {
			need := int(math.Floor(quantity))
			var selectedIDs []int64
			selectionRows, err := tx.Query(ctx, `
				SELECT serial_code_id
				FROM stock_out_item_serial_selection
				WHERE stock_out_item_id = $1
				ORDER BY id
			`, itemID)
			if err != nil {
				return err
			}
			for selectionRows.Next() {
				var id int64
				if err := selectionRows.Scan(&id); err != nil {
					selectionRows.Close()
					return err
				}
				selectedIDs = append(selectedIDs, id)
			}
			selectionRows.Close()

			if len(selectedIDs) != need {
				return fmt.Errorf("编码物料未备齐，需先备货 %d 个编码 [stock_out_item_id=%d]", need, itemID)
			}

			tag, err := tx.Exec(ctx, `
				UPDATE sku_serial_code
				SET status = 'issued', updated_at = NOW()
				WHERE id = ANY($1)
				  AND inventory_id = $2
				  AND status = 'in_stock'
			`, selectedIDs, inventoryID)
			if err != nil {
				return err
			}
			if int(tag.RowsAffected()) != need {
				return fmt.Errorf("部分编码状态已变化，请刷新后重试 [stock_out_item_id=%d]", itemID)
			}

			if _, err := tx.Exec(ctx, `
				INSERT INTO sku_serial_trace (
					serial_code_id, serial_code, action, ref_doc_type, ref_doc_no, ref_doc_id,
					from_warehouse_id, operator_id, remark
				)
				SELECT id, serial_code, 'stock_out', 'stock_out', $1, $2, warehouse_id, $3, '出库确认-编码已领用'
				FROM sku_serial_code
				WHERE id = ANY($4)
			`, stockOutNo, stockOutID, userID, selectedIDs); err != nil {
				return err
			}
		}
	}

	if _, err := tx.Exec(ctx, `
		UPDATE stock_out
		SET status = 'confirmed',
		    confirmed_at = NOW(),
		    updated_by = $2,
		    updated_at = NOW()
		WHERE id = $1
	`, stockOutID, userID); err != nil {
		return err
	}

	if refDocID != nil {
		if _, err := tx.Exec(ctx, `
			UPDATE return_order
			SET return_status = 'completed',
			    updated_by = $2,
			    updated_at = NOW()
			WHERE id = $1
		`, *refDocID, userID); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM stock_out_item_serial_selection
		WHERE stock_out_item_id IN (SELECT id FROM stock_out_item WHERE stock_out_id = $1)
	`, stockOutID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func UpdateStockOutSerialSelections(ctx context.Context, stockOutID, userID int64, req *request.StockOutSerialSelectionReq) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM stock_out WHERE id = $1 AND deleted_at IS NULL`, stockOutID).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("出库单不存在")
		}
		return err
	}
	if status == "confirmed" {
		return fmt.Errorf("出库单已完成，不能再调整编码")
	}

	rows, err := tx.Query(ctx, `
		SELECT soi.id, soi.inventory_id, soi.quantity, m.is_code
		FROM stock_out_item soi
		JOIN material m ON m.id = soi.material_id
		WHERE soi.stock_out_id = $1
	`, stockOutID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type codedItem struct {
		inventoryID int64
		needCount   int
	}
	coded := map[int64]codedItem{}
	for rows.Next() {
		var itemID, inventoryID int64
		var qty float64
		var isCode bool
		if err := rows.Scan(&itemID, &inventoryID, &qty, &isCode); err != nil {
			return err
		}
		if isCode {
			coded[itemID] = codedItem{
				inventoryID: inventoryID,
				needCount:   int(math.Floor(qty)),
			}
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if req.Mode == "auto_fifo" {
		if _, err := tx.Exec(ctx, `
			DELETE FROM stock_out_item_serial_selection
			WHERE stock_out_item_id IN (
				SELECT id FROM stock_out_item WHERE stock_out_id = $1
			)
		`, stockOutID); err != nil {
			return err
		}

		for stockOutItemID, meta := range coded {
			if meta.needCount <= 0 {
				continue
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO stock_out_item_serial_selection (stock_out_item_id, serial_code_id, created_by)
				SELECT $1, sc.id, $2
				FROM sku_serial_code sc
				WHERE sc.inventory_id = $3
				  AND sc.status = 'in_stock'
				  AND NOT EXISTS (
				      SELECT 1
				      FROM stock_out_item_serial_selection other_sel
				      INNER JOIN stock_out_item other_soi ON other_soi.id = other_sel.stock_out_item_id
				      INNER JOIN stock_out other_so ON other_so.id = other_soi.stock_out_id
				      WHERE other_sel.serial_code_id = sc.id
				        AND other_so.id <> $5
				        AND other_so.deleted_at IS NULL
				        AND other_so.status IN ('draft', 'pending')
				  )
				ORDER BY sc.created_at ASC, sc.id ASC
				LIMIT $4
			`, stockOutItemID, userID, meta.inventoryID, meta.needCount, stockOutID); err != nil {
				return err
			}
		}
		return tx.Commit(ctx)
	}

	manualMap := map[int64][]int64{}
	for _, item := range req.Items {
		manualMap[item.StockOutItemID] = item.SerialCodeIDs
	}

	for stockOutItemID, meta := range coded {
		ids := manualMap[stockOutItemID]
		if len(ids) != meta.needCount {
			return fmt.Errorf("编码物料需要选择 %d 个编码 [stock_out_item_id=%d]", meta.needCount, stockOutItemID)
		}

		var validCount int
		validRows, err := tx.Query(ctx, `
			SELECT id
			FROM sku_serial_code
			WHERE id = ANY($1)
			  AND inventory_id = $2
			  AND status = 'in_stock'
			FOR UPDATE
		`, ids, meta.inventoryID)
		if err != nil {
			return err
		}
		for validRows.Next() {
			validCount++
		}
		if err := validRows.Err(); err != nil {
			validRows.Close()
			return err
		}
		validRows.Close()
		if validCount != meta.needCount {
			return fmt.Errorf("所选编码包含无效项（非在库或不属于当前库存）[stock_out_item_id=%d]", stockOutItemID)
		}

		var lockedByOthers int
		if err := tx.QueryRow(ctx, `
			SELECT count(*)
			FROM stock_out_item_serial_selection other_sel
			INNER JOIN stock_out_item other_soi ON other_soi.id = other_sel.stock_out_item_id
			INNER JOIN stock_out other_so ON other_so.id = other_soi.stock_out_id
			WHERE other_sel.serial_code_id = ANY($1)
			  AND other_so.id <> $2
			  AND other_so.deleted_at IS NULL
			  AND other_so.status IN ('draft', 'pending')
		`, ids, stockOutID).Scan(&lockedByOthers); err != nil {
			return err
		}
		if lockedByOthers > 0 {
			return fmt.Errorf("所选编码已被其他待出库单备货占用，请刷新后重试 [stock_out_item_id=%d]", stockOutItemID)
		}
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM stock_out_item_serial_selection
		WHERE stock_out_item_id IN (
			SELECT id FROM stock_out_item WHERE stock_out_id = $1
		)
	`, stockOutID); err != nil {
		return err
	}

	for stockOutItemID, ids := range manualMap {
		if _, ok := coded[stockOutItemID]; !ok {
			continue
		}
		if len(ids) == 0 {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO stock_out_item_serial_selection (stock_out_item_id, serial_code_id, created_by)
			SELECT $1, unnest($2::bigint[]), $3
		`, stockOutItemID, ids, userID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
