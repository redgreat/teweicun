package db

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

type consumptionInvRow struct {
	ID             int64
	MaterialID     int64
	WarehouseID    int64
	WarehouseCode  string
	Quantity       float64
	LockedQuantity float64
	InTransit      float64
}

func sumPendingConsumptionByInventory(ctx context.Context, tx pgx.Tx, invIDs []int64) (map[int64]float64, error) {
	out := make(map[int64]float64)
	if len(invIDs) == 0 {
		return out, nil
	}
	rows, err := tx.Query(ctx, `
		SELECT coi.inventory_id, COALESCE(SUM(coi.quantity), 0)::float8
		FROM consumption_order_item coi
		INNER JOIN consumption_order co ON co.id = coi.order_id AND co.deleted_at IS NULL
		WHERE co.status IN ('draft', 'pending')
		  AND co.stock_out_id IS NULL
		  AND coi.inventory_id = ANY($1)
		GROUP BY coi.inventory_id
	`, invIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var qty float64
		if err := rows.Scan(&id, &qty); err != nil {
			return nil, err
		}
		out[id] = qty
	}
	return out, rows.Err()
}

func loadConsumptionInventoryRows(ctx context.Context, tx pgx.Tx, invIDs []int64) (map[int64]consumptionInvRow, error) {
	if len(invIDs) == 0 {
		return nil, fmt.Errorf("领料明细缺少库存批次")
	}
	rows, err := tx.Query(ctx, `
		SELECT i.id, i.material_id, i.warehouse_id, w.warehouse_code,
		       i.quantity, i.locked_quantity, COALESCE(i.in_transit_quantity, 0)
		FROM inventory i
		INNER JOIN warehouse w ON w.id = i.warehouse_id AND w.deleted_at IS NULL
		WHERE i.id = ANY($1)
		FOR UPDATE OF i
	`, invIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int64]consumptionInvRow, len(invIDs))
	for rows.Next() {
		var r consumptionInvRow
		if err := rows.Scan(
			&r.ID, &r.MaterialID, &r.WarehouseID, &r.WarehouseCode,
			&r.Quantity, &r.LockedQuantity, &r.InTransit,
		); err != nil {
			return nil, err
		}
		out[r.ID] = r
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) != len(invIDs) {
		return nil, fmt.Errorf("部分库存批次不存在或已删除")
	}
	return out, nil
}

func loadMaterialIsCode(ctx context.Context, tx pgx.Tx, materialIDs []int64) (map[int64]bool, error) {
	out := make(map[int64]bool)
	if len(materialIDs) == 0 {
		return out, nil
	}
	rows, err := tx.Query(ctx, `SELECT id, is_code FROM material WHERE id = ANY($1)`, materialIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var isCode bool
		if err := rows.Scan(&id, &isCode); err != nil {
			return nil, err
		}
		out[id] = isCode
	}
	return out, rows.Err()
}

func CreateConsumptionOrder(ctx context.Context, req request.ConsumptionOrderCreate, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	wantQty := make(map[int64]float64)
	invSeen := make(map[int64]struct{})
	invIDs := make([]int64, 0, len(req.Items))
	for _, it := range req.Items {
		if it.InventoryID <= 0 {
			return 0, fmt.Errorf("明细必须选择有效库存批次")
		}
		wantQty[it.InventoryID] += it.Quantity
		if _, ok := invSeen[it.InventoryID]; !ok {
			invSeen[it.InventoryID] = struct{}{}
			invIDs = append(invIDs, it.InventoryID)
		}
	}

	invRows, err := loadConsumptionInventoryRows(ctx, tx, invIDs)
	if err != nil {
		return 0, err
	}

	pendingReserved, err := sumPendingConsumptionByInventory(ctx, tx, invIDs)
	if err != nil {
		return 0, err
	}

	matSet := make(map[int64]struct{})
	for _, it := range req.Items {
		matSet[it.MaterialID] = struct{}{}
	}
	matIDs := make([]int64, 0, len(matSet))
	for id := range matSet {
		matIDs = append(matIDs, id)
	}
	isCodeByMat, err := loadMaterialIsCode(ctx, tx, matIDs)
	if err != nil {
		return 0, err
	}

	for i, it := range req.Items {
		row, ok := invRows[it.InventoryID]
		if !ok {
			return 0, fmt.Errorf("第%d行：库存批次无效", i+1)
		}
		if row.MaterialID != it.MaterialID {
			return 0, fmt.Errorf("第%d行：物料与所选库存批次不一致", i+1)
		}
		if isCode, ok := isCodeByMat[it.MaterialID]; ok && isCode {
			if math.Abs(it.Quantity-math.Trunc(it.Quantity)) > 1e-9 {
				return 0, fmt.Errorf("第%d行：编码管理物料领料数量须为整数", i+1)
			}
		}
	}

	for invID, need := range wantQty {
		row := invRows[invID]
		avail := row.Quantity - row.LockedQuantity - row.InTransit - pendingReserved[invID]
		if need > avail+1e-9 {
			return 0, fmt.Errorf("库存批次 ID %d 可用量不足（可用 %.3f，含其他草稿/待确认领料占用 %.3f），本次需求 %.3f",
				row.ID, row.Quantity-row.LockedQuantity-row.InTransit, pendingReserved[invID], need)
		}
	}

	var orderID int64
	orderQuery := `
		INSERT INTO consumption_order (
			order_no, project_no, product_name,
			order_date, designer_id, designer_name, status, remark, created_by
		) VALUES (
			fn_generate_consumption_order_no(), $1, $2, $3, $4, $5, 'pending', $6, $7
		) RETURNING id
	`
	err = tx.QueryRow(ctx, orderQuery,
		req.ProjectNo, req.ProductName,
		req.OrderDateTime, req.DesignerID, req.DesignerName, req.Remark, userID,
	).Scan(&orderID)
	if err != nil {
		return 0, err
	}

	itemQuery := `
		INSERT INTO consumption_order_item (
			order_id, material_id, inventory_id, quantity,
			unit, remark, warehouse_id, warehouse_code
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	for _, item := range req.Items {
		row := invRows[item.InventoryID]
		_, err = tx.Exec(ctx, itemQuery,
			orderID, item.MaterialID, item.InventoryID, item.Quantity,
			item.Unit, item.Remark, row.WarehouseID, row.WarehouseCode,
		)
		if err != nil {
			return 0, err
		}

	}

	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "CREATE", "CONSUMPTION_ORDER", "consumption_order", orderID, nil)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(ctx, `CALL sp_confirm_consumption_order($1, $2)`, orderID, userID)
	if err != nil {
		return 0, err
	}

	return orderID, tx.Commit(ctx)
}

func ListConsumptionOrders(ctx context.Context, query request.ConsumptionOrderQuery) (*response.ConsumptionOrderListResp, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if query.OrderNo != "" {
		conditions = append(conditions, fmt.Sprintf("co.order_no ILIKE $%d", argIdx))
		args = append(args, "%"+query.OrderNo+"%")
		argIdx++
	}
	if query.ProjectNo != "" {
		conditions = append(conditions, fmt.Sprintf("co.project_no ILIKE $%d", argIdx))
		args = append(args, "%"+query.ProjectNo+"%")
		argIdx++
	}
	if query.ProductName != "" {
		conditions = append(conditions, fmt.Sprintf("co.product_name ILIKE $%d", argIdx))
		args = append(args, "%"+query.ProductName+"%")
		argIdx++
	}
	if query.WarehouseCode != "" {
		conditions = append(conditions, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM consumption_order_item coi
			INNER JOIN inventory i ON i.id = coi.inventory_id
			INNER JOIN warehouse w ON w.id = i.warehouse_id AND w.deleted_at IS NULL
			WHERE coi.order_id = co.id AND w.warehouse_code = $%d
		)`, argIdx))
		args = append(args, query.WarehouseCode)
		argIdx++
	}
	if query.DesignerID > 0 {
		conditions = append(conditions, fmt.Sprintf("co.designer_id = $%d", argIdx))
		args = append(args, query.DesignerID)
		argIdx++
	}
	if query.Status != "" {
		conditions = append(conditions, fmt.Sprintf("co.status = $%d", argIdx))
		args = append(args, query.Status)
		argIdx++
	}
	if !query.StartDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("co.order_date >= $%d", argIdx))
		args = append(args, query.StartDate)
		argIdx++
	}
	if !query.EndDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("co.order_date <= $%d", argIdx))
		args = append(args, query.EndDate)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM v_consumption_order_list co %s
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
			COALESCE(stock_out_id, 0), COALESCE(stock_out_no, ''), COALESCE(remark, ''), created_at, updated_at,
			item_count, total_quantity,
			COALESCE((SELECT SUM(coi2.quantity * i.unit_cost) FROM consumption_order_item coi2 JOIN inventory i ON i.id = coi2.inventory_id WHERE coi2.order_id = co.id), 0) AS total_amount
		FROM v_consumption_order_list co
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	rows, err := database.Pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []response.ConsumptionOrderResp
	for rows.Next() {
		var order response.ConsumptionOrderResp
		err := rows.Scan(
			&order.ID, &order.OrderNo, &order.ProjectNo, &order.ProductName,
			&order.WarehouseID, &order.WarehouseCode, &order.WarehouseName,
			&order.OrderDate, &order.DesignerID, &order.DesignerName, &order.Status,
			&order.StockOutID, &order.StockOutNo, &order.Remark,
			&order.CreatedAt, &order.UpdatedAt,
			&order.ItemCount, &order.TotalQuantity, &order.TotalAmount,
		)
		if err != nil {
			return nil, err
		}
		order.StatusName = getConsumptionOrderStatusName(order.Status)
		orders = append(orders, order)
	}

	return &response.ConsumptionOrderListResp{
		List:  orders,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func GetConsumptionOrderDetail(ctx context.Context, id int64) (*response.ConsumptionOrderResp, error) {
	orderQuery := `
		SELECT
			co.id, co.order_no, co.project_no, co.product_name,
			COALESCE(
				(SELECT coi2.warehouse_id FROM consumption_order_item coi2
				 WHERE coi2.order_id = co.id AND coi2.warehouse_id IS NOT NULL
				 ORDER BY coi2.id DESC LIMIT 1),
				(SELECT i.warehouse_id FROM consumption_order_item coi2
				 INNER JOIN inventory i ON i.id = coi2.inventory_id
				 WHERE coi2.order_id = co.id ORDER BY coi2.id LIMIT 1),
				0
			),
			COALESCE(
				(SELECT coi2.warehouse_code FROM consumption_order_item coi2
				 WHERE coi2.order_id = co.id AND coi2.warehouse_code IS NOT NULL
				   AND btrim(coi2.warehouse_code::text) <> ''
				 ORDER BY coi2.id DESC LIMIT 1),
				(SELECT w2.warehouse_code FROM consumption_order_item coi2
				 INNER JOIN inventory i ON i.id = coi2.inventory_id
				 INNER JOIN warehouse w2 ON w2.id = i.warehouse_id AND w2.deleted_at IS NULL
				 WHERE coi2.order_id = co.id ORDER BY coi2.id LIMIT 1),
				''
			),
			COALESCE(
				(SELECT w3.warehouse_name FROM consumption_order_item coi2
				 INNER JOIN warehouse w3 ON w3.id = coi2.warehouse_id AND w3.deleted_at IS NULL
				 WHERE coi2.order_id = co.id AND coi2.warehouse_id IS NOT NULL
				 ORDER BY coi2.id DESC LIMIT 1),
				(SELECT w4.warehouse_name FROM consumption_order_item coi2
				 INNER JOIN inventory i ON i.id = coi2.inventory_id
				 INNER JOIN warehouse w4 ON w4.id = i.warehouse_id AND w4.deleted_at IS NULL
				 WHERE coi2.order_id = co.id ORDER BY coi2.id LIMIT 1),
				''
			),
			co.order_date, co.designer_id, COALESCE(co.designer_name, ''), co.status,
			COALESCE(co.stock_out_id, 0), COALESCE(so.stock_out_no, ''), COALESCE(co.remark, ''), co.created_at, co.updated_at
		FROM consumption_order co
		LEFT JOIN stock_out so ON so.id = co.stock_out_id
		WHERE co.id = $1 AND co.deleted_at IS NULL
	`
	var order response.ConsumptionOrderResp
	err := database.Pool.QueryRow(ctx, orderQuery, id).Scan(
		&order.ID, &order.OrderNo, &order.ProjectNo, &order.ProductName,
		&order.WarehouseID, &order.WarehouseCode, &order.WarehouseName,
		&order.OrderDate, &order.DesignerID, &order.DesignerName, &order.Status,
		&order.StockOutID, &order.StockOutNo, &order.Remark,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	order.StatusName = getConsumptionOrderStatusName(order.Status)

	itemQuery := `
		SELECT 
			coi.id, coi.order_id, coi.material_id, m.material_code, m.material_name,
			coi.inventory_id, COALESCE(i.sku_id, 0), COALESCE(v.sku_code, ''), COALESCE(v.sku_name, ''),
			coi.quantity, coi.unit, COALESCE(i.unit_cost, 0),
			COALESCE(coi.remark, '')
		FROM consumption_order_item coi
		INNER JOIN material m ON m.id = coi.material_id
		LEFT JOIN inventory i ON i.id = coi.inventory_id
		LEFT JOIN v_sku_list v ON v.id = i.sku_id
		WHERE coi.order_id = $1
	`
	rows, err := database.Pool.Query(ctx, itemQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item response.ConsumptionOrderItemResp
		if err := rows.Scan(
			&item.ID, &item.OrderID, &item.MaterialID, &item.MaterialCode, &item.MaterialName,
			&item.InventoryID, &item.SKUID, &item.SKUCode, &item.SKUName,
			&item.Quantity, &item.Unit, &item.UnitCost,
			&item.Remark,
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

func UpdateConsumptionOrderStatus(ctx context.Context, id int64, status string, userID int64) error {
	query := `
		UPDATE consumption_order
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

func ConfirmConsumptionOrder(ctx context.Context, id int64, userID int64) error {
	query := `CALL sp_confirm_consumption_order($1, $2)`
	_, err := database.Pool.Exec(ctx, query, id, userID)
	return err
}

func DeleteConsumptionOrder(ctx context.Context, id int64, userID int64) error {
	query := `
		UPDATE consumption_order
		SET deleted_at = NOW(), updated_by = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL AND status = 'draft'
	`
	result, err := database.Pool.Exec(ctx, query, userID, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("只能删除草稿状态的领料订单")
	}
	return nil
}

func getConsumptionOrderStatusName(status string) string {
	statusMap := map[string]string{
		"draft":     "待提交",
		"pending":   "待出库",
		"confirmed": "待出库",
		"completed": "已完成",
		"cancelled": "已取消",
	}
	if name, ok := statusMap[status]; ok {
		return name
	}
	return status
}
