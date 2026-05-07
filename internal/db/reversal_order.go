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

type reversalResolvedLine struct {
	MaterialID  int64
	Unit        string
	Quantity    float64
	InventoryID int64
	Remark      string
}

func CreateReversalOrder(ctx context.Context, req request.ReversalOrderCreate, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	lines := make([]reversalResolvedLine, 0, len(req.Items))
	var headerWhID int64
	var headerWhCode string

	for i, item := range req.Items {
		if item.InventoryID <= 0 {
			return 0, fmt.Errorf("第%d行须选择在库 SKU/库存批次", i+1)
		}

		var whID int64
		var whCode string
		var matID int64
		var unit string
		var avail float64
		err = tx.QueryRow(ctx, `
			SELECT i.warehouse_id, w.warehouse_code, i.material_id,
			       COALESCE(NULLIF(TRIM(i.unit), ''), m.unit),
			       (i.quantity - i.locked_quantity - COALESCE(i.in_transit_quantity, 0))
			FROM inventory i
			INNER JOIN warehouse w ON w.id = i.warehouse_id AND w.deleted_at IS NULL
			INNER JOIN material m ON m.id = i.material_id
			WHERE i.id = $1
		`, item.InventoryID).Scan(&whID, &whCode, &matID, &unit, &avail)
		if err != nil {
			if err == pgx.ErrNoRows {
				return 0, fmt.Errorf("第%d行：库存批次不存在", i+1)
			}
			return 0, err
		}

		if item.MaterialID > 0 && item.MaterialID != matID {
			return 0, fmt.Errorf("第%d行：物料与所选库存不一致", i+1)
		}

		if i == 0 {
			headerWhID, headerWhCode = whID, whCode
		} else if whID != headerWhID {
			return 0, fmt.Errorf("退料明细涉及多个仓库，请拆单后再提交")
		}

		if req.WarehouseID > 0 && req.WarehouseID != headerWhID {
			return 0, fmt.Errorf("表头仓库与所选库存所在仓库不一致")
		}
		if strings.TrimSpace(req.WarehouseCode) != "" && strings.TrimSpace(req.WarehouseCode) != strings.TrimSpace(headerWhCode) {
			return 0, fmt.Errorf("表头仓库编码与所选库存所在仓库不一致")
		}

		if item.Quantity > avail+1e-9 {
			return 0, fmt.Errorf("第%d行：退料数量超过可用库存（可用 %.3f）", i+1, avail)
		}

		u := strings.TrimSpace(item.Unit)
		if u == "" {
			u = strings.TrimSpace(unit)
		}
		if u == "" {
			u = "件"
		}

		lines = append(lines, reversalResolvedLine{
			MaterialID:  matID,
			Unit:        u,
			Quantity:    item.Quantity,
			InventoryID: item.InventoryID,
			Remark:      item.Remark,
		})
	}

	if headerWhID == 0 {
		return 0, fmt.Errorf("无法解析退料仓库")
	}

	var orderID int64
	orderQuery := `
		INSERT INTO reversal_order (
			order_no, project_no, product_name, warehouse_id, warehouse_code,
			order_date, designer_id, designer_name, status, remark, created_by
		) VALUES (
			fn_generate_reversal_order_no(), $1, $2, $3, $4, $5, $6, $7, 'pending', $8, $9
		) RETURNING id
	`
	err = tx.QueryRow(ctx, orderQuery,
		req.ProjectNo, req.ProductName, headerWhID, headerWhCode,
		req.OrderDate, req.DesignerID, req.DesignerName, req.Remark, userID,
	).Scan(&orderID)
	if err != nil {
		return 0, err
	}

	itemQuery := `
		INSERT INTO reversal_order_item (
			order_id, material_id, quantity, unit, remark, inventory_id
		) VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, ln := range lines {
		_, err = tx.Exec(ctx, itemQuery,
			orderID, ln.MaterialID, ln.Quantity, ln.Unit,
			ln.Remark, ln.InventoryID,
		)
		if err != nil {
			return 0, err
		}
	}

	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "CREATE", "REVERSAL_ORDER", "reversal_order", orderID, nil)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(ctx, `CALL sp_confirm_reversal_order($1, $2)`, orderID, userID)
	if err != nil {
		return 0, err
	}

	return orderID, tx.Commit(ctx)
}

func ListReversalOrders(ctx context.Context, query request.ReversalOrderQuery) (*response.ReversalOrderListResp, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if query.OrderNo != "" {
		conditions = append(conditions, fmt.Sprintf("ro.order_no ILIKE $%d", argIdx))
		args = append(args, "%"+query.OrderNo+"%")
		argIdx++
	}
	if query.ProjectNo != "" {
		conditions = append(conditions, fmt.Sprintf("ro.project_no ILIKE $%d", argIdx))
		args = append(args, "%"+query.ProjectNo+"%")
		argIdx++
	}
	if query.ProductName != "" {
		conditions = append(conditions, fmt.Sprintf("ro.product_name ILIKE $%d", argIdx))
		args = append(args, "%"+query.ProductName+"%")
		argIdx++
	}
	if query.WarehouseCode != "" {
		conditions = append(conditions, fmt.Sprintf("ro.warehouse_code = $%d", argIdx))
		args = append(args, query.WarehouseCode)
		argIdx++
	}
	if query.DesignerID > 0 {
		conditions = append(conditions, fmt.Sprintf("ro.designer_id = $%d", argIdx))
		args = append(args, query.DesignerID)
		argIdx++
	}
	if query.Status != "" {
		conditions = append(conditions, fmt.Sprintf("ro.status = $%d", argIdx))
		args = append(args, query.Status)
		argIdx++
	}
	if !query.StartDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("ro.order_date >= $%d", argIdx))
		args = append(args, query.StartDate)
		argIdx++
	}
	if !query.EndDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("ro.order_date <= $%d", argIdx))
		args = append(args, query.EndDate)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM v_reversal_order_list ro %s
	`, whereClause)

	var total int64
	err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (query.Page - 1) * query.PageSize
	args = append(args, query.PageSize, offset)

	listQuery := fmt.Sprintf(`
		SELECT 
			id, order_no, project_no, product_name, warehouse_id, COALESCE(warehouse_code, ''),
			COALESCE(warehouse_name, ''), order_date, designer_id, COALESCE(designer_name, ''), status,
			COALESCE(stock_in_id, 0), COALESCE(stock_in_no, ''), COALESCE(remark, ''), created_at, updated_at,
			item_count, total_quantity,
			COALESCE((
				SELECT SUM(roi.quantity * COALESCE(inv.unit_cost, 0))
				FROM reversal_order_item roi
				LEFT JOIN inventory inv ON inv.id = roi.inventory_id
				WHERE roi.order_id = ro.id
			), 0) AS total_amount
		FROM v_reversal_order_list ro
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	rows, err := database.Pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []response.ReversalOrderResp
	for rows.Next() {
		var order response.ReversalOrderResp
		err := rows.Scan(
			&order.ID, &order.OrderNo, &order.ProjectNo, &order.ProductName,
			&order.WarehouseID, &order.WarehouseCode, &order.WarehouseName,
			&order.OrderDate, &order.DesignerID, &order.DesignerName, &order.Status,
			&order.StockInID, &order.StockInNo, &order.Remark,
			&order.CreatedAt, &order.UpdatedAt,
			&order.ItemCount, &order.TotalQuantity, &order.TotalAmount,
		)
		if err != nil {
			return nil, err
		}
		order.StatusName = getReversalOrderStatusName(order.Status)
		orders = append(orders, order)
	}

	return &response.ReversalOrderListResp{
		List:  orders,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func GetReversalOrderDetail(ctx context.Context, id int64) (*response.ReversalOrderResp, error) {
	orderQuery := `
		SELECT 
			ro.id, ro.order_no, ro.project_no, ro.product_name, ro.warehouse_id, COALESCE(ro.warehouse_code, ''),
			COALESCE(w.warehouse_name, ''), ro.order_date, ro.designer_id, COALESCE(ro.designer_name, ''), ro.status,
			COALESCE(ro.stock_in_id, 0), COALESCE(si.stock_in_no, ''), COALESCE(ro.remark, ''), ro.created_at, ro.updated_at
		FROM reversal_order ro
		LEFT JOIN warehouse w ON w.id = ro.warehouse_id
		LEFT JOIN stock_in si ON si.id = ro.stock_in_id
		WHERE ro.id = $1 AND ro.deleted_at IS NULL
	`
	var order response.ReversalOrderResp
	err := database.Pool.QueryRow(ctx, orderQuery, id).Scan(
		&order.ID, &order.OrderNo, &order.ProjectNo, &order.ProductName,
		&order.WarehouseID, &order.WarehouseCode, &order.WarehouseName,
		&order.OrderDate, &order.DesignerID, &order.DesignerName, &order.Status,
		&order.StockInID, &order.StockInNo, &order.Remark,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	order.StatusName = getReversalOrderStatusName(order.Status)

	itemQuery := `
		SELECT 
			roi.id, roi.order_id, roi.material_id, m.material_code, m.material_name,
			0::bigint AS sku_id, ''::varchar AS sku_code, ''::varchar AS sku_name,
			roi.quantity, roi.unit, COALESCE(inv.unit_cost, 0),
			COALESCE(roi.remark, ''), COALESCE(w.warehouse_name, '')
		FROM reversal_order_item roi
		INNER JOIN material m ON m.id = roi.material_id
		LEFT JOIN inventory inv ON inv.id = roi.inventory_id
		LEFT JOIN warehouse w ON w.id = inv.warehouse_id
		WHERE roi.order_id = $1
	`
	rows, err := database.Pool.Query(ctx, itemQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item response.ReversalOrderItemResp
		if err := rows.Scan(
			&item.ID, &item.OrderID, &item.MaterialID, &item.MaterialCode, &item.MaterialName,
			&item.SKUID, &item.SKUCode, &item.SKUName,
			&item.Quantity, &item.Unit, &item.UnitCost,
			&item.Remark, &item.WarehouseName,
		); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	order.ItemCount = len(order.Items)
	for _, item := range order.Items {
		order.TotalQuantity += item.Quantity
	}

	return &order, rows.Err()
}

func UpdateReversalOrderStatus(ctx context.Context, id int64, status string, userID int64) error {
	query := `
		UPDATE reversal_order
		SET status = $1, updated_by = $2, updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`
	result, err := database.Pool.Exec(ctx, query, status, userID, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func ConfirmReversalOrder(ctx context.Context, id int64, userID int64) error {
	query := `CALL sp_confirm_reversal_order($1, $2)`
	_, err := database.Pool.Exec(ctx, query, id, userID)
	return err
}

func DeleteReversalOrder(ctx context.Context, id int64, userID int64) error {
	query := `
		UPDATE reversal_order
		SET deleted_at = NOW(), updated_by = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL AND status = 'draft'
	`
	result, err := database.Pool.Exec(ctx, query, userID, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("只能删除草稿状态的退料订单")
	}
	return nil
}

func getReversalOrderStatusName(status string) string {
	statusMap := map[string]string{
		"draft":     "待提交",
		"pending":   "待入库",
		"confirmed": "待入库",
		"completed": "已完成",
		"cancelled": "已取消",
	}
	if name, ok := statusMap[status]; ok {
		return name
	}
	return status
}
