/**
 * 功能：inventory.go
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

// ListInventoryDetail 查询库存明细（基于视图 v_inventory_detail）
func ListInventoryDetail(ctx context.Context, q *request.InventoryQuery) ([]response.InventoryDetailResp, int64, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if q.MaterialCode != "" {
		where = append(where, fmt.Sprintf("material_code ILIKE $%d", argID))
		args = append(args, "%"+q.MaterialCode+"%")
		argID++
	}
	if q.MaterialName != "" {
		where = append(where, fmt.Sprintf("material_name ILIKE $%d", argID))
		args = append(args, "%"+q.MaterialName+"%")
		argID++
	}
	if q.WarehouseID != 0 {
		where = append(where, fmt.Sprintf("warehouse_id = $%d", argID))
		args = append(args, q.WarehouseID)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_inventory_detail WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, material_id, material_code, material_name, unit, 
		       warehouse_id, warehouse_name, quantity, locked_quantity, 
		       available_quantity as available, stock_in_date, 
		       certificate_id, COALESCE(certificate_no, '')
		FROM v_inventory_detail
		WHERE %s
		ORDER BY stock_in_date ASC, id ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.InventoryDetailResp
	for rows.Next() {
		var item response.InventoryDetailResp
		if err := rows.Scan(&item.ID, &item.MaterialID, &item.MaterialCode, &item.MaterialName,
			&item.Unit, &item.WarehouseID, &item.WarehouseName, &item.Quantity, &item.LockedQuantity,
			&item.Available, &item.StockInDate, &item.CertificateID, &item.CertificateNo); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// ListInventorySummary 查询库存汇总（基于视图 v_inventory_summary）
func ListInventorySummary(ctx context.Context, q *request.InventoryQuery) ([]response.InventorySummaryResp, int64, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if q.MaterialCode != "" {
		where = append(where, fmt.Sprintf("material_code ILIKE $%d", argID))
		args = append(args, "%"+q.MaterialCode+"%")
		argID++
	}
	if q.MaterialName != "" {
		where = append(where, fmt.Sprintf("material_name ILIKE $%d", argID))
		args = append(args, "%"+q.MaterialName+"%")
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_inventory_summary WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT material_id, material_code, material_name, unit, 
		       total_quantity, locked_quantity, available_quantity as available
		FROM v_inventory_summary
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

	var result []response.InventorySummaryResp
	for rows.Next() {
		var item response.InventorySummaryResp
		if err := rows.Scan(&item.MaterialID, &item.MaterialCode, &item.MaterialName,
			&item.Unit, &item.TotalQuantity, &item.LockedQuantity, &item.Available); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

func ListInventoryAvailable(ctx context.Context, q *request.InventoryAvailableQuery) ([]response.InventoryAvailableResp, int64, error) {
	where := []string{"i.quantity > 0"}
	var args []interface{}
	argID := 1

	if strings.TrimSpace(q.WarehouseCode) != "" {
		where = append(where, fmt.Sprintf("w.warehouse_code = $%d", argID))
		args = append(args, q.WarehouseCode)
		argID++
	}

	if strings.TrimSpace(q.SupplierCode) != "" {
		where = append(where, fmt.Sprintf("si.supplier_code = $%d", argID))
		args = append(args, q.SupplierCode)
		argID++
	}

	if strings.TrimSpace(q.Q) != "" {
		where = append(where, fmt.Sprintf("(m.material_code ILIKE $%d OR m.material_name ILIKE $%d OR v.sku_code ILIKE $%d OR v.sku_name ILIKE $%d)", argID, argID, argID, argID))
		args = append(args, "%"+q.Q+"%")
		argID++
	}

	where = append(where, "(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) > 0")
	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`
		SELECT count(*)
		FROM inventory i
		INNER JOIN material m ON m.id = i.material_id
		INNER JOIN warehouse w ON w.id = i.warehouse_id
		LEFT JOIN LATERAL (
		  SELECT it.ref_doc_id
		  FROM inventory_transaction it
		  WHERE it.material_id = i.material_id
		    AND it.warehouse_id = i.warehouse_id
		    AND it.trans_type = 'in'
		    AND it.ref_doc_type = 'stock_in'
		  ORDER BY it.created_at DESC
		  LIMIT 1
		) itx ON true
		LEFT JOIN stock_in si ON si.id = itx.ref_doc_id
		LEFT JOIN v_sku_list v ON v.id = i.sku_id
		WHERE %s
	`, whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT
			i.id AS inventory_id,
			i.material_id, m.material_code, m.material_name, m.is_code,
			COALESCE(i.sku_id, 0) AS sku_id,
			COALESCE(v.sku_code, '') AS sku_code,
			COALESCE(v.sku_name, '') AS sku_name,
			i.warehouse_id, w.warehouse_code, w.warehouse_name,
			COALESCE(v.unit, m.unit, i.unit) AS unit, COALESCE(i.unit_cost, 0),
			i.quantity, i.locked_quantity, COALESCE(i.in_transit_quantity, 0) AS in_transit_quantity,
			(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) AS available_quantity
		FROM inventory i
		INNER JOIN material m ON m.id = i.material_id
		INNER JOIN warehouse w ON w.id = i.warehouse_id
		LEFT JOIN LATERAL (
		  SELECT it.ref_doc_id
		  FROM inventory_transaction it
		  WHERE it.material_id = i.material_id
		    AND it.warehouse_id = i.warehouse_id
		    AND it.trans_type = 'in'
		    AND it.ref_doc_type = 'stock_in'
		  ORDER BY it.created_at DESC
		  LIMIT 1
		) itx ON true
		LEFT JOIN stock_in si ON si.id = itx.ref_doc_id
		LEFT JOIN v_sku_list v ON v.id = i.sku_id
		WHERE %s
		ORDER BY w.warehouse_code, m.material_code, i.stock_in_date ASC NULLS LAST, i.id ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())
	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.InventoryAvailableResp
	for rows.Next() {
		var item response.InventoryAvailableResp
		if err := rows.Scan(
			&item.InventoryID,
			&item.MaterialID, &item.MaterialCode, &item.MaterialName, &item.IsCode,
			&item.SKUID, &item.SKUCode, &item.SKUName,
			&item.WarehouseID, &item.WarehouseCode, &item.WarehouseName,
			&item.Unit, &item.UnitCost,
			&item.Quantity, &item.LockedQuantity, &item.InTransitQuantity,
			&item.AvailableQuantity,
		); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// ListInventoryIssued 查询“已出库(可退料)”库存批次（同一列表合并有编码 / 无编码）
// - 有编码：sku_serial_code.status=issued 且追溯为领料出库完成；issued_quantity=该批次仍 issued 的件数
// - 无编码：按 stock_out_item（consumption_order）汇总出库量，减去 reversal_order 已完成退料（净额）
// 两者均返回 available_quantity（当前批次可用量）；前端上限取 min(issued, available)
func ListInventoryIssued(ctx context.Context, q *request.InventoryIssuedQuery) ([]response.InventoryIssuedResp, int64, error) {
	var args []interface{}
	whSQL := "TRUE"
	qSQL := "TRUE"
	if strings.TrimSpace(q.WarehouseCode) != "" {
		whSQL = fmt.Sprintf("w.warehouse_code = $%d", len(args)+1)
		args = append(args, strings.TrimSpace(q.WarehouseCode))
	}
	if strings.TrimSpace(q.Q) != "" {
		qSQL = fmt.Sprintf("(m.material_code ILIKE $%d OR m.material_name ILIKE $%d OR COALESCE(v.sku_code, '') ILIKE $%d OR COALESCE(v.sku_name, '') ILIKE $%d)", len(args)+1, len(args)+1, len(args)+1, len(args)+1)
		args = append(args, "%"+strings.TrimSpace(q.Q)+"%")
	}

	codedWhere := strings.Join([]string{
		"sc.status = 'issued'",
		`EXISTS (
			SELECT 1
			FROM sku_serial_trace t
			INNER JOIN stock_out so ON so.id = t.ref_doc_id AND so.deleted_at IS NULL
			WHERE t.serial_code_id = sc.id
			  AND t.action = 'stock_out'
			  AND t.ref_doc_type = 'stock_out'
			  AND so.status = 'confirmed'
			  AND so.ref_doc_type = 'consumption_order'
		)`,
		whSQL,
		qSQL,
		"(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) > 0",
	}, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT count(*)
		FROM (
			SELECT i.id
			FROM sku_serial_code sc
			INNER JOIN inventory i ON i.id = sc.inventory_id
			INNER JOIN material m ON m.id = sc.material_id
			INNER JOIN warehouse w ON w.id = i.warehouse_id AND w.deleted_at IS NULL
			LEFT JOIN v_sku_list v ON v.id = i.sku_id
			WHERE %s
			GROUP BY i.id, sc.material_id, m.material_code, m.material_name, m.is_code,
			         COALESCE(i.sku_id, 0), COALESCE(v.sku_code, ''), COALESCE(v.sku_name, ''),
			         i.warehouse_id, w.warehouse_code, w.warehouse_name,
			         COALESCE(v.unit, m.unit, i.unit), COALESCE(i.unit_cost, 0),
			         i.quantity, i.locked_quantity, i.in_transit_quantity
			HAVING COUNT(*) > 0
			UNION ALL
			SELECT i.id
			FROM inventory i
			INNER JOIN material m ON m.id = i.material_id AND COALESCE(m.is_code, false) = false
			INNER JOIN warehouse w ON w.id = i.warehouse_id AND w.deleted_at IS NULL
			LEFT JOIN v_sku_list v ON v.id = i.sku_id
			INNER JOIN (
				SELECT soi.inventory_id, SUM(soi.quantity) AS out_qty
				FROM stock_out_item soi
				INNER JOIN stock_out so ON so.id = soi.stock_out_id AND so.deleted_at IS NULL
				WHERE so.status = 'confirmed'
				  AND so.ref_doc_type = 'consumption_order'
				  AND COALESCE(soi.inventory_id, 0) <> 0
				GROUP BY soi.inventory_id
			) so_agg ON so_agg.inventory_id = i.id
			LEFT JOIN (
				SELECT roi.inventory_id, SUM(roi.quantity) AS ret_qty
				FROM reversal_order_item roi
				INNER JOIN reversal_order ro ON ro.id = roi.order_id AND ro.deleted_at IS NULL
				WHERE ro.status = 'completed'
				  AND COALESCE(roi.inventory_id, 0) <> 0
				GROUP BY roi.inventory_id
			) rev_agg ON rev_agg.inventory_id = i.id
			WHERE GREATEST(COALESCE(so_agg.out_qty, 0) - COALESCE(rev_agg.ret_qty, 0), 0) > 0
			  AND (i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) > 0
			  AND %s
			  AND %s
		) t
	`, codedWhere, whSQL, qSQL)

	var total int64
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limitArg := len(args) + 1
	offArg := len(args) + 2
	query := fmt.Sprintf(`
		SELECT * FROM (
			SELECT
				i.id AS inventory_id,
				sc.material_id, m.material_code, m.material_name, COALESCE(m.is_code, false),
				COALESCE(i.sku_id, 0) AS sku_id,
				COALESCE(v.sku_code, '') AS sku_code, COALESCE(v.sku_name, '') AS sku_name,
				i.warehouse_id, w.warehouse_code, w.warehouse_name,
				COALESCE(v.unit, m.unit, i.unit) AS unit,
				COALESCE(i.unit_cost, 0) AS unit_cost,
				COUNT(*)::double precision AS issued_quantity,
				(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0))::double precision AS available_quantity
			FROM sku_serial_code sc
			INNER JOIN inventory i ON i.id = sc.inventory_id
			INNER JOIN material m ON m.id = sc.material_id
			INNER JOIN warehouse w ON w.id = i.warehouse_id AND w.deleted_at IS NULL
			LEFT JOIN v_sku_list v ON v.id = i.sku_id
			WHERE %s
			GROUP BY i.id, sc.material_id, m.material_code, m.material_name, m.is_code,
			         COALESCE(i.sku_id, 0), COALESCE(v.sku_code, ''), COALESCE(v.sku_name, ''),
			         i.warehouse_id, w.warehouse_code, w.warehouse_name,
			         COALESCE(v.unit, m.unit, i.unit), COALESCE(i.unit_cost, 0),
			         i.quantity, i.locked_quantity, i.in_transit_quantity
			HAVING COUNT(*) > 0
			UNION ALL
			SELECT
				i.id AS inventory_id,
				i.material_id, m.material_code, m.material_name, COALESCE(m.is_code, false),
				COALESCE(i.sku_id, 0) AS sku_id,
				COALESCE(v.sku_code, '') AS sku_code, COALESCE(v.sku_name, '') AS sku_name,
				i.warehouse_id, w.warehouse_code, w.warehouse_name,
				COALESCE(v.unit, m.unit, i.unit) AS unit,
				COALESCE(i.unit_cost, 0) AS unit_cost,
				GREATEST(COALESCE(so_agg.out_qty, 0) - COALESCE(rev_agg.ret_qty, 0), 0)::double precision AS issued_quantity,
				(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0))::double precision AS available_quantity
			FROM inventory i
			INNER JOIN material m ON m.id = i.material_id AND COALESCE(m.is_code, false) = false
			INNER JOIN warehouse w ON w.id = i.warehouse_id AND w.deleted_at IS NULL
			LEFT JOIN v_sku_list v ON v.id = i.sku_id
			INNER JOIN (
				SELECT soi.inventory_id, SUM(soi.quantity) AS out_qty
				FROM stock_out_item soi
				INNER JOIN stock_out so ON so.id = soi.stock_out_id AND so.deleted_at IS NULL
				WHERE so.status = 'confirmed'
				  AND so.ref_doc_type = 'consumption_order'
				  AND COALESCE(soi.inventory_id, 0) <> 0
				GROUP BY soi.inventory_id
			) so_agg ON so_agg.inventory_id = i.id
			LEFT JOIN (
				SELECT roi.inventory_id, SUM(roi.quantity) AS ret_qty
				FROM reversal_order_item roi
				INNER JOIN reversal_order ro ON ro.id = roi.order_id AND ro.deleted_at IS NULL
				WHERE ro.status = 'completed'
				  AND COALESCE(roi.inventory_id, 0) <> 0
				GROUP BY roi.inventory_id
			) rev_agg ON rev_agg.inventory_id = i.id
			WHERE GREATEST(COALESCE(so_agg.out_qty, 0) - COALESCE(rev_agg.ret_qty, 0), 0) > 0
			  AND (i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) > 0
			  AND %s
			  AND %s
		) merged
		ORDER BY merged.warehouse_code, merged.material_code, merged.sku_code, merged.inventory_id
		LIMIT $%d OFFSET $%d
	`, codedWhere, whSQL, qSQL, limitArg, offArg)

	args = append(args, q.PageSize, q.Offset())
	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]response.InventoryIssuedResp, 0)
	for rows.Next() {
		var r response.InventoryIssuedResp
		if err := rows.Scan(
			&r.InventoryID,
			&r.MaterialID, &r.MaterialCode, &r.MaterialName, &r.IsCode,
			&r.SKUID, &r.SKUCode, &r.SKUName,
			&r.WarehouseID, &r.WarehouseCode, &r.WarehouseName,
			&r.Unit, &r.UnitCost,
			&r.IssuedQuantity,
			&r.AvailableQuantity,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

func ListInventorySKULedger(ctx context.Context, q *request.InventorySKULedgerQuery) ([]response.InventorySKULedgerResp, int64, *response.InventorySKULedgerStatsResp, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if strings.TrimSpace(q.SKUName) != "" {
		where = append(where, fmt.Sprintf("(COALESCE(v.sku_name, '') ILIKE $%d OR COALESCE(v.sku_code, '') ILIKE $%d)", argID, argID))
		args = append(args, "%"+strings.TrimSpace(q.SKUName)+"%")
		argID++
	}
	if strings.TrimSpace(q.WarehouseName) != "" {
		where = append(where, fmt.Sprintf("w.warehouse_name ILIKE $%d", argID))
		args = append(args, "%"+strings.TrimSpace(q.WarehouseName)+"%")
		argID++
	}
	if q.PriceMin > 0 {
		where = append(where, fmt.Sprintf("COALESCE(i.unit_cost, 0) >= $%d", argID))
		args = append(args, q.PriceMin)
		argID++
	}
	if q.PriceMax > 0 {
		where = append(where, fmt.Sprintf("COALESCE(i.unit_cost, 0) <= $%d", argID))
		args = append(args, q.PriceMax)
		argID++
	}

	where = append(where, "(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) > 0")
	whereClause := strings.Join(where, " AND ")

	groupBase := fmt.Sprintf(`
		FROM inventory i
		INNER JOIN material m ON m.id = i.material_id
		INNER JOIN warehouse w ON w.id = i.warehouse_id AND w.deleted_at IS NULL
		LEFT JOIN v_sku_list v ON v.id = i.sku_id
		WHERE %s
	`, whereClause)

	var total int64
	countQuery := fmt.Sprintf(`
		SELECT count(*)
		FROM (
			SELECT i.material_id, COALESCE(i.sku_id, 0), i.warehouse_id, COALESCE(i.unit_cost, 0)
			%s
			GROUP BY i.material_id, COALESCE(i.sku_id, 0), i.warehouse_id, COALESCE(i.unit_cost, 0)
		) t
	`, groupBase)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, nil, err
	}

	statsQuery := fmt.Sprintf(`
		SELECT
			COALESCE(SUM((i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) * COALESCE(i.unit_cost, 0)), 0),
			COALESCE(SUM(CASE WHEN m.is_code THEN (i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) * COALESCE(i.unit_cost, 0) ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN NOT m.is_code THEN (i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) * COALESCE(i.unit_cost, 0) ELSE 0 END), 0),
			COALESCE(SUM(i.locked_quantity), 0)
		%s
	`, groupBase)
	stats := &response.InventorySKULedgerStatsResp{}
	if err := database.Pool.QueryRow(ctx, statsQuery, args...).Scan(
		&stats.TotalAmount, &stats.CodeTotalAmount, &stats.NoCodeTotalAmount, &stats.TotalLockedQty,
	); err != nil {
		return nil, 0, nil, err
	}

	query := fmt.Sprintf(`
		SELECT
			i.material_id,
			COALESCE(m.material_name, ''),
			COALESCE(i.sku_id, 0),
			COALESCE(v.sku_name, ''),
			i.warehouse_id,
			COALESCE(w.warehouse_name, ''),
			COALESCE(m.is_code, false),
			COALESCE(SUM(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)), 0) AS qty,
			COALESCE(i.unit_cost, 0) AS unit_cost,
			COALESCE(SUM((i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) * COALESCE(i.unit_cost, 0)), 0) AS total_amount,
			COALESCE(SUM(i.locked_quantity), 0) AS locked_quantity,
			COUNT(*) AS inventory_count,
			MAX(CASE WHEN i.sku_id IS NOT NULL THEN 1 ELSE 0 END) > 0 AS has_custom_attrs
		%s
		GROUP BY i.material_id, m.material_name, COALESCE(i.sku_id, 0), COALESCE(v.sku_name, ''), i.warehouse_id, w.warehouse_name, COALESCE(m.is_code, false), COALESCE(i.unit_cost, 0)
		ORDER BY m.material_name ASC, COALESCE(v.sku_name, '') ASC, w.warehouse_name ASC, COALESCE(i.unit_cost, 0) ASC
		LIMIT $%d OFFSET $%d
	`, groupBase, argID, argID+1)

	queryArgs := append(args, q.PageSize, q.Offset())
	rows, err := database.Pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, nil, err
	}
	defer rows.Close()

	result := make([]response.InventorySKULedgerResp, 0)
	for rows.Next() {
		var item response.InventorySKULedgerResp
		if err := rows.Scan(
			&item.MaterialID, &item.MaterialName, &item.SKUID, &item.SKUName, &item.WarehouseID, &item.WarehouseName,
			&item.IsCode, &item.Quantity, &item.UnitCost, &item.TotalAmount, &item.LockedQuantity, &item.InventoryCount, &item.HasCustomAttrs,
		); err != nil {
			return nil, 0, nil, err
		}
		result = append(result, item)
	}

	return result, total, stats, rows.Err()
}

func ListInventorySKUSerials(ctx context.Context, q *request.InventorySKUSerialQuery) ([]response.InventorySKUSerialResp, error) {
	sql := `
		SELECT
			sc.serial_code,
			sc.status,
			CASE sc.status
				WHEN 'in_stock' THEN '在库'
				WHEN 'issued' THEN '已领用'
				WHEN 'returned' THEN '已退回'
				WHEN 'scrapped' THEN '已报废'
				ELSE sc.status
			END AS status_name
		FROM sku_serial_code sc
		INNER JOIN inventory i ON i.id = sc.inventory_id
		WHERE i.material_id = $1
		  AND i.warehouse_id = $2
		  AND COALESCE(i.sku_id, 0) = COALESCE($3, 0)
		  AND COALESCE(i.unit_cost, 0) = $4
		ORDER BY sc.serial_code ASC
	`
	rows, err := database.Pool.Query(ctx, sql, q.MaterialID, q.WarehouseID, q.SKUID, q.UnitCost)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]response.InventorySKUSerialResp, 0)
	for rows.Next() {
		var item response.InventorySKUSerialResp
		if err := rows.Scan(&item.SerialCode, &item.Status, &item.StatusName); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func ExportInventorySKULedger(ctx context.Context, q *request.InventorySKULedgerQuery) ([]response.InventorySKULedgerResp, *response.InventorySKULedgerStatsResp, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if strings.TrimSpace(q.SKUName) != "" {
		where = append(where, fmt.Sprintf("(COALESCE(v.sku_name, '') ILIKE $%d OR COALESCE(v.sku_code, '') ILIKE $%d)", argID, argID))
		args = append(args, "%"+strings.TrimSpace(q.SKUName)+"%")
		argID++
	}
	if strings.TrimSpace(q.WarehouseName) != "" {
		where = append(where, fmt.Sprintf("w.warehouse_name ILIKE $%d", argID))
		args = append(args, "%"+strings.TrimSpace(q.WarehouseName)+"%")
		argID++
	}
	if q.PriceMin > 0 {
		where = append(where, fmt.Sprintf("COALESCE(i.unit_cost, 0) >= $%d", argID))
		args = append(args, q.PriceMin)
		argID++
	}
	if q.PriceMax > 0 {
		where = append(where, fmt.Sprintf("COALESCE(i.unit_cost, 0) <= $%d", argID))
		args = append(args, q.PriceMax)
		argID++
	}

	where = append(where, "(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) > 0")
	whereClause := strings.Join(where, " AND ")

	groupBase := fmt.Sprintf(`
		FROM inventory i
		INNER JOIN material m ON m.id = i.material_id
		INNER JOIN warehouse w ON w.id = i.warehouse_id AND w.deleted_at IS NULL
		LEFT JOIN v_sku_list v ON v.id = i.sku_id
		WHERE %s
	`, whereClause)

	statsQuery := fmt.Sprintf(`
		SELECT
			COALESCE(SUM((i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) * COALESCE(i.unit_cost, 0)), 0),
			COALESCE(SUM(CASE WHEN m.is_code THEN (i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) * COALESCE(i.unit_cost, 0) ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN NOT m.is_code THEN (i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) * COALESCE(i.unit_cost, 0) ELSE 0 END), 0),
			COALESCE(SUM(i.locked_quantity), 0)
		%s
	`, groupBase)
	stats := &response.InventorySKULedgerStatsResp{}
	if err := database.Pool.QueryRow(ctx, statsQuery, args...).Scan(
		&stats.TotalAmount, &stats.CodeTotalAmount, &stats.NoCodeTotalAmount, &stats.TotalLockedQty,
	); err != nil {
		return nil, nil, err
	}

	query := fmt.Sprintf(`
		SELECT
			i.material_id,
			COALESCE(m.material_name, ''),
			COALESCE(i.sku_id, 0),
			COALESCE(v.sku_name, ''),
			i.warehouse_id,
			COALESCE(w.warehouse_name, ''),
			COALESCE(m.is_code, false),
			COALESCE(SUM(i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)), 0) AS qty,
			COALESCE(i.unit_cost, 0) AS unit_cost,
			COALESCE(SUM((i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0)) * COALESCE(i.unit_cost, 0)), 0) AS total_amount,
			COALESCE(SUM(i.locked_quantity), 0) AS locked_quantity,
			COUNT(*) AS inventory_count,
			MAX(CASE WHEN i.sku_id IS NOT NULL THEN 1 ELSE 0 END) > 0 AS has_custom_attrs
		%s
		GROUP BY i.material_id, m.material_name, COALESCE(i.sku_id, 0), COALESCE(v.sku_name, ''), i.warehouse_id, w.warehouse_name, COALESCE(m.is_code, false), COALESCE(i.unit_cost, 0)
		ORDER BY m.material_name ASC, COALESCE(v.sku_name, '') ASC, w.warehouse_name ASC, COALESCE(i.unit_cost, 0) ASC
	`, groupBase)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	result := make([]response.InventorySKULedgerResp, 0)
	for rows.Next() {
		var item response.InventorySKULedgerResp
		if err := rows.Scan(
			&item.MaterialID, &item.MaterialName, &item.SKUID, &item.SKUName, &item.WarehouseID, &item.WarehouseName,
			&item.IsCode, &item.Quantity, &item.UnitCost, &item.TotalAmount, &item.LockedQuantity, &item.InventoryCount, &item.HasCustomAttrs,
		); err != nil {
			return nil, nil, err
		}
		result = append(result, item)
	}

	return result, stats, rows.Err()
}
