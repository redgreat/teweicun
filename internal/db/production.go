/**
 * 功能：生产单/生产退货单 数据访问
 * 创建时间：2026-06-06
 * 创建人：GPT-5.2
 */

package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/pkg/database"
)

func ListProductionOrders(ctx context.Context, q request.ProductionOrderQuery) (*response.ProductionOrderListResp, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, "TRUE")
	if strings.TrimSpace(q.ProductionNo) != "" {
		conditions = append(conditions, fmt.Sprintf("production_no ILIKE $%d", argIdx))
		args = append(args, "%"+strings.TrimSpace(q.ProductionNo)+"%")
		argIdx++
	}
	if strings.TrimSpace(q.Status) != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, strings.TrimSpace(q.Status))
		argIdx++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	var total int64
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM v_production_order_list %s`, whereClause)
	if err := database.Pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}

	args = append(args, q.PageSize, q.Offset())
	listSQL := fmt.Sprintf(`
		SELECT
			id, production_no, status,
			COALESCE(consumption_order_id, 0), COALESCE(consumption_order_no, ''),
			COALESCE(stock_out_id, 0), COALESCE(stock_out_no, ''),
			COALESCE(stock_in_id, 0), COALESCE(stock_in_no, ''),
			COALESCE(produced_material_id, 0), COALESCE(produced_material_code, ''), COALESCE(produced_material_name, ''),
			COALESCE(produced_warehouse_id, 0), COALESCE(produced_warehouse_code, ''), COALESCE(produced_warehouse_name, ''),
			COALESCE(produced_quantity, 0)::float8, COALESCE(produced_unit_cost, 0)::float8,
			COALESCE(cost_price, 0)::float8,
			COALESCE(remark, ''), created_at, updated_at
		FROM v_production_order_list
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	rows, err := database.Pool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]response.ProductionOrderResp, 0, q.PageSize)
	for rows.Next() {
		var r response.ProductionOrderResp
		if err := rows.Scan(
			&r.ID, &r.ProductionNo, &r.Status,
			&r.ConsumptionOrderID, &r.ConsumptionOrderNo,
			&r.StockOutID, &r.StockOutNo,
			&r.StockInID, &r.StockInNo,
			&r.ProducedMaterialID, &r.ProducedMaterialCode, &r.ProducedMaterialName,
			&r.ProducedWarehouseID, &r.ProducedWarehouseCode, &r.ProducedWarehouseName,
			&r.ProducedQuantity, &r.ProducedUnitCost,
			&r.CostPrice,
			&r.Remark, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		r.StatusName = productionOrderStatusName(r.Status)
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &response.ProductionOrderListResp{
		List:  out,
		Total: total,
		Page:  q.Page,
		Size:  q.PageSize,
	}, nil
}

func GetProductionOrderDetail(ctx context.Context, id int64) (*response.ProductionOrderResp, error) {
	sql := `
		SELECT
			id, production_no, status,
			COALESCE(consumption_order_id, 0), COALESCE(consumption_order_no, ''),
			COALESCE(stock_out_id, 0), COALESCE(stock_out_no, ''),
			COALESCE(stock_in_id, 0), COALESCE(stock_in_no, ''),
			COALESCE(produced_material_id, 0), COALESCE(produced_material_code, ''), COALESCE(produced_material_name, ''),
			COALESCE(produced_warehouse_id, 0), COALESCE(produced_warehouse_code, ''), COALESCE(produced_warehouse_name, ''),
			COALESCE(produced_quantity, 0)::float8, COALESCE(produced_unit_cost, 0)::float8,
			COALESCE(cost_price, 0)::float8,
			COALESCE(remark, ''), created_at, updated_at
		FROM v_production_order_list
		WHERE id = $1
	`
	var r response.ProductionOrderResp
	if err := database.Pool.QueryRow(ctx, sql, id).Scan(
		&r.ID, &r.ProductionNo, &r.Status,
		&r.ConsumptionOrderID, &r.ConsumptionOrderNo,
		&r.StockOutID, &r.StockOutNo,
		&r.StockInID, &r.StockInNo,
		&r.ProducedMaterialID, &r.ProducedMaterialCode, &r.ProducedMaterialName,
		&r.ProducedWarehouseID, &r.ProducedWarehouseCode, &r.ProducedWarehouseName,
		&r.ProducedQuantity, &r.ProducedUnitCost,
		&r.CostPrice,
		&r.Remark, &r.CreatedAt, &r.UpdatedAt,
	); err != nil {
		return nil, err
	}
	r.StatusName = productionOrderStatusName(r.Status)
	return &r, nil
}

func ListProductionReturnOrders(ctx context.Context, q request.ProductionReturnOrderQuery) (*response.ProductionReturnOrderListResp, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, "TRUE")
	if strings.TrimSpace(q.ReturnNo) != "" {
		conditions = append(conditions, fmt.Sprintf("return_no ILIKE $%d", argIdx))
		args = append(args, "%"+strings.TrimSpace(q.ReturnNo)+"%")
		argIdx++
	}
	if strings.TrimSpace(q.Status) != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, strings.TrimSpace(q.Status))
		argIdx++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	var total int64
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM v_production_return_order_list %s`, whereClause)
	if err := database.Pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}

	args = append(args, q.PageSize, q.Offset())
	listSQL := fmt.Sprintf(`
		SELECT
			id, return_no, status,
			COALESCE(production_order_id, 0), COALESCE(production_no, ''),
			COALESCE(consumption_order_id, 0), COALESCE(consumption_order_no, ''),
			COALESCE(stock_out_id, 0), COALESCE(stock_out_no, ''),
			COALESCE(produced_material_id, 0), COALESCE(produced_material_code, ''), COALESCE(produced_material_name, ''),
			COALESCE(produced_warehouse_id, 0), COALESCE(produced_warehouse_code, ''), COALESCE(produced_warehouse_name, ''),
			COALESCE(returned_quantity, 0)::float8,
			COALESCE(cost_price, 0)::float8,
			COALESCE(remark, ''), created_at, updated_at
		FROM v_production_return_order_list
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	rows, err := database.Pool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]response.ProductionReturnOrderResp, 0, q.PageSize)
	for rows.Next() {
		var r response.ProductionReturnOrderResp
		if err := rows.Scan(
			&r.ID, &r.ReturnNo, &r.Status,
			&r.ProductionOrderID, &r.ProductionNo,
			&r.ConsumptionOrderID, &r.ConsumptionOrderNo,
			&r.StockOutID, &r.StockOutNo,
			&r.ProducedMaterialID, &r.ProducedMaterialCode, &r.ProducedMaterialName,
			&r.ProducedWarehouseID, &r.ProducedWarehouseCode, &r.ProducedWarehouseName,
			&r.ReturnedQuantity,
			&r.CostPrice,
			&r.Remark, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		r.StatusName = productionReturnStatusName(r.Status)
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &response.ProductionReturnOrderListResp{
		List:  out,
		Total: total,
		Page:  q.Page,
		Size:  q.PageSize,
	}, nil
}

func GetProductionReturnOrderDetail(ctx context.Context, id int64) (*response.ProductionReturnOrderResp, error) {
	sql := `
		SELECT
			id, return_no, status,
			COALESCE(production_order_id, 0), COALESCE(production_no, ''),
			COALESCE(consumption_order_id, 0), COALESCE(consumption_order_no, ''),
			COALESCE(stock_out_id, 0), COALESCE(stock_out_no, ''),
			COALESCE(produced_material_id, 0), COALESCE(produced_material_code, ''), COALESCE(produced_material_name, ''),
			COALESCE(produced_warehouse_id, 0), COALESCE(produced_warehouse_code, ''), COALESCE(produced_warehouse_name, ''),
			COALESCE(returned_quantity, 0)::float8,
			COALESCE(cost_price, 0)::float8,
			COALESCE(remark, ''), created_at, updated_at
		FROM v_production_return_order_list
		WHERE id = $1
	`
	var r response.ProductionReturnOrderResp
	if err := database.Pool.QueryRow(ctx, sql, id).Scan(
		&r.ID, &r.ReturnNo, &r.Status,
		&r.ProductionOrderID, &r.ProductionNo,
		&r.ConsumptionOrderID, &r.ConsumptionOrderNo,
		&r.StockOutID, &r.StockOutNo,
		&r.ProducedMaterialID, &r.ProducedMaterialCode, &r.ProducedMaterialName,
		&r.ProducedWarehouseID, &r.ProducedWarehouseCode, &r.ProducedWarehouseName,
		&r.ReturnedQuantity,
		&r.CostPrice,
		&r.Remark, &r.CreatedAt, &r.UpdatedAt,
	); err != nil {
		return nil, err
	}
	r.StatusName = productionReturnStatusName(r.Status)
	return &r, nil
}

// UpdateProductionOrder 更新生产单（成本价格、备注）
func UpdateProductionOrder(ctx context.Context, id int64, req request.ProductionOrderUpdate, userID int64) error {
	sql := `
		UPDATE production_order
		SET cost_price = CASE WHEN $1 > 0 THEN $1 ELSE cost_price END,
		    remark = CASE WHEN $2 <> '' THEN $2 ELSE remark END,
		    updated_by = $3,
		    updated_at = NOW()
		WHERE id = $4
	`
	_, err := database.Pool.Exec(ctx, sql, req.CostPrice, req.Remark, userID, id)
	return err
}

// UpdateProductionReturnOrder 更新生产退货单（成本价格、备注）
func UpdateProductionReturnOrder(ctx context.Context, id int64, req request.ProductionReturnOrderUpdate, userID int64) error {
	sql := `
		UPDATE production_return_order
		SET cost_price = CASE WHEN $1 > 0 THEN $1 ELSE cost_price END,
		    remark = CASE WHEN $2 <> '' THEN $2 ELSE remark END,
		    updated_by = $3,
		    updated_at = NOW()
		WHERE id = $4
	`
	_, err := database.Pool.Exec(ctx, sql, req.CostPrice, req.Remark, userID, id)
	return err
}

// ListConsumptionOrdersByProduction 查询某个生产单关联的领料单列表
func ListConsumptionOrdersByProduction(ctx context.Context, productionOrderID int64) ([]response.ConsumptionOrderResp, error) {
	sql := `
		SELECT
			co.id, co.order_no, co.project_no, co.product_name,
			0, '', '', co.order_date,
			co.designer_id, COALESCE(co.designer_name, ''), co.status,
			COALESCE(co.stock_out_id, 0), COALESCE(so.stock_out_no, ''), COALESCE(co.remark, ''),
			co.created_at, co.updated_at,
			COALESCE((SELECT COUNT(1) FROM consumption_order_item coi WHERE coi.order_id = co.id), 0),
			COALESCE((SELECT COALESCE(SUM(coi.quantity), 0) FROM consumption_order_item coi WHERE coi.order_id = co.id), 0),
			COALESCE((SELECT SUM(coi2.quantity * COALESCE(i.unit_cost, 0)) FROM consumption_order_item coi2 LEFT JOIN inventory i ON i.id = coi2.inventory_id WHERE coi2.order_id = co.id), 0),
			COALESCE(co.production_order_id, 0), '',
			0, '',
			COALESCE(co.production_return_order_id, 0), ''
		FROM consumption_order co
		LEFT JOIN stock_out so ON so.id = co.stock_out_id AND so.deleted_at IS NULL
		WHERE co.production_order_id = $1 AND co.deleted_at IS NULL
		ORDER BY co.created_at DESC
	`
	rows, err := database.Pool.Query(ctx, sql, productionOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []response.ConsumptionOrderResp
	for rows.Next() {
		var r response.ConsumptionOrderResp
		if err := rows.Scan(
			&r.ID, &r.OrderNo, &r.ProjectNo, &r.ProductName,
			&r.WarehouseID, &r.WarehouseCode, &r.WarehouseName,
			&r.OrderDate, &r.DesignerID, &r.DesignerName, &r.Status,
			&r.StockOutID, &r.StockOutNo, &r.Remark,
			&r.CreatedAt, &r.UpdatedAt,
			&r.ItemCount, &r.TotalQuantity, &r.TotalAmount,
			&r.ProductionOrderID, &r.ProductionNo,
			&r.ProductionStockInID, &r.ProductionStockInNo,
			&r.ProductionReturnOrderID, &r.ProductionReturnNo,
		); err != nil {
			return nil, err
		}
		r.StatusName = getConsumptionOrderStatusName(r.Status)
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListReversalOrdersByProduction 查询某个生产单关联的退料单列表
func ListReversalOrdersByProduction(ctx context.Context, productionOrderID int64) ([]response.ReversalOrderResp, error) {
	sql := `
		SELECT
			ro.id, ro.order_no, ro.project_no, ro.product_name,
			ro.warehouse_id, COALESCE(ro.warehouse_code, ''), COALESCE(w.warehouse_name, ''),
			ro.order_date, ro.designer_id, COALESCE(ro.designer_name, ''), ro.status,
			COALESCE(ro.stock_in_id, 0), COALESCE(si.stock_in_no, ''), COALESCE(ro.remark, ''),
			ro.created_at, ro.updated_at,
			COALESCE((SELECT COUNT(1) FROM reversal_order_item roi WHERE roi.order_id = ro.id), 0),
			COALESCE((SELECT COALESCE(SUM(roi.quantity), 0) FROM reversal_order_item roi WHERE roi.order_id = ro.id), 0),
			COALESCE((SELECT SUM(roi.quantity * COALESCE(inv.unit_cost, 0)) FROM reversal_order_item roi LEFT JOIN inventory inv ON inv.id = roi.inventory_id WHERE roi.order_id = ro.id), 0),
			COALESCE(ro.production_order_id, 0), '',
			COALESCE(ro.production_return_order_id, 0), ''
		FROM reversal_order ro
		LEFT JOIN warehouse w ON w.id = ro.warehouse_id AND w.deleted_at IS NULL
		LEFT JOIN stock_in si ON si.id = ro.stock_in_id AND si.deleted_at IS NULL
		WHERE ro.production_order_id = $1 AND ro.deleted_at IS NULL
		ORDER BY ro.created_at DESC
	`
	rows, err := database.Pool.Query(ctx, sql, productionOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []response.ReversalOrderResp
	for rows.Next() {
		var r response.ReversalOrderResp
		if err := rows.Scan(
			&r.ID, &r.OrderNo, &r.ProjectNo, &r.ProductName,
			&r.WarehouseID, &r.WarehouseCode, &r.WarehouseName,
			&r.OrderDate, &r.DesignerID, &r.DesignerName, &r.Status,
			&r.StockInID, &r.StockInNo, &r.Remark,
			&r.CreatedAt, &r.UpdatedAt,
			&r.ItemCount, &r.TotalQuantity, &r.TotalAmount,
			&r.ProductionOrderID, &r.ProductionNo,
			&r.ProductionReturnOrderID, &r.ProductionReturnNo,
		); err != nil {
			return nil, err
		}
		r.StatusName = getReversalOrderStatusName(r.Status)
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListConsumptionOrdersByProductionReturn 查询某个生产退货单关联的领料单列表
func ListConsumptionOrdersByProductionReturn(ctx context.Context, returnOrderID int64) ([]response.ConsumptionOrderResp, error) {
	sql := `
		SELECT
			co.id, co.order_no, co.project_no, co.product_name,
			0, '', '', co.order_date,
			co.designer_id, COALESCE(co.designer_name, ''), co.status,
			COALESCE(co.stock_out_id, 0), COALESCE(so.stock_out_no, ''), COALESCE(co.remark, ''),
			co.created_at, co.updated_at,
			COALESCE((SELECT COUNT(1) FROM consumption_order_item coi WHERE coi.order_id = co.id), 0),
			COALESCE((SELECT COALESCE(SUM(coi.quantity), 0) FROM consumption_order_item coi WHERE coi.order_id = co.id), 0),
			COALESCE((SELECT SUM(coi2.quantity * COALESCE(i.unit_cost, 0)) FROM consumption_order_item coi2 LEFT JOIN inventory i ON i.id = coi2.inventory_id WHERE coi2.order_id = co.id), 0),
			COALESCE(co.production_order_id, 0), '',
			0, '',
			COALESCE(co.production_return_order_id, 0), ''
		FROM consumption_order co
		LEFT JOIN stock_out so ON so.id = co.stock_out_id AND so.deleted_at IS NULL
		WHERE co.production_return_order_id = $1 AND co.deleted_at IS NULL
		ORDER BY co.created_at DESC
	`
	rows, err := database.Pool.Query(ctx, sql, returnOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []response.ConsumptionOrderResp
	for rows.Next() {
		var r response.ConsumptionOrderResp
		if err := rows.Scan(
			&r.ID, &r.OrderNo, &r.ProjectNo, &r.ProductName,
			&r.WarehouseID, &r.WarehouseCode, &r.WarehouseName,
			&r.OrderDate, &r.DesignerID, &r.DesignerName, &r.Status,
			&r.StockOutID, &r.StockOutNo, &r.Remark,
			&r.CreatedAt, &r.UpdatedAt,
			&r.ItemCount, &r.TotalQuantity, &r.TotalAmount,
			&r.ProductionOrderID, &r.ProductionNo,
			&r.ProductionStockInID, &r.ProductionStockInNo,
			&r.ProductionReturnOrderID, &r.ProductionReturnNo,
		); err != nil {
			return nil, err
		}
		r.StatusName = getConsumptionOrderStatusName(r.Status)
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListReversalOrdersByProductionReturn 查询某个生产退货单关联的退料单列表
func ListReversalOrdersByProductionReturn(ctx context.Context, returnOrderID int64) ([]response.ReversalOrderResp, error) {
	sql := `
		SELECT
			ro.id, ro.order_no, ro.project_no, ro.product_name,
			ro.warehouse_id, COALESCE(ro.warehouse_code, ''), COALESCE(w.warehouse_name, ''),
			ro.order_date, ro.designer_id, COALESCE(ro.designer_name, ''), ro.status,
			COALESCE(ro.stock_in_id, 0), COALESCE(si.stock_in_no, ''), COALESCE(ro.remark, ''),
			ro.created_at, ro.updated_at,
			COALESCE((SELECT COUNT(1) FROM reversal_order_item roi WHERE roi.order_id = ro.id), 0),
			COALESCE((SELECT COALESCE(SUM(roi.quantity), 0) FROM reversal_order_item roi WHERE roi.order_id = ro.id), 0),
			COALESCE((SELECT SUM(roi.quantity * COALESCE(inv.unit_cost, 0)) FROM reversal_order_item roi LEFT JOIN inventory inv ON inv.id = roi.inventory_id WHERE roi.order_id = ro.id), 0),
			COALESCE(ro.production_order_id, 0), '',
			COALESCE(ro.production_return_order_id, 0), ''
		FROM reversal_order ro
		LEFT JOIN warehouse w ON w.id = ro.warehouse_id AND w.deleted_at IS NULL
		LEFT JOIN stock_in si ON si.id = ro.stock_in_id AND si.deleted_at IS NULL
		WHERE ro.production_return_order_id = $1 AND ro.deleted_at IS NULL
		ORDER BY ro.created_at DESC
	`
	rows, err := database.Pool.Query(ctx, sql, returnOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []response.ReversalOrderResp
	for rows.Next() {
		var r response.ReversalOrderResp
		if err := rows.Scan(
			&r.ID, &r.OrderNo, &r.ProjectNo, &r.ProductName,
			&r.WarehouseID, &r.WarehouseCode, &r.WarehouseName,
			&r.OrderDate, &r.DesignerID, &r.DesignerName, &r.Status,
			&r.StockInID, &r.StockInNo, &r.Remark,
			&r.CreatedAt, &r.UpdatedAt,
			&r.ItemCount, &r.TotalQuantity, &r.TotalAmount,
			&r.ProductionOrderID, &r.ProductionNo,
			&r.ProductionReturnOrderID, &r.ProductionReturnNo,
		); err != nil {
			return nil, err
		}
		r.StatusName = getReversalOrderStatusName(r.Status)
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListProductionOrdersForDropdown 下拉列表用的生产单简要查询
func ListProductionOrdersForDropdown(ctx context.Context, keyword string) ([]map[string]interface{}, error) {
	sql := `
		SELECT id, production_no, COALESCE(produced_material_name, '') AS material_name, COALESCE(cost_price, 0)::float8 AS cost_price
		FROM v_production_order_list
		WHERE ($1 = '' OR production_no ILIKE $1)
		ORDER BY created_at DESC
		LIMIT 50
	`
	kw := "%" + strings.TrimSpace(keyword) + "%"
	rows, err := database.Pool.Query(ctx, sql, kw)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]interface{}
	for rows.Next() {
		var id int64
		var no, matName string
		var costPrice float64
		if err := rows.Scan(&id, &no, &matName, &costPrice); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{
			"id":            id,
			"production_no": no,
			"material_name": matName,
			"cost_price":    costPrice,
		})
	}
	return out, rows.Err()
}

// ListProductionReturnOrdersForDropdown 下拉列表用的生产退货单简要查询
func ListProductionReturnOrdersForDropdown(ctx context.Context, keyword string) ([]map[string]interface{}, error) {
	sql := `
		SELECT id, return_no, COALESCE(produced_material_name, '') AS material_name, COALESCE(cost_price, 0)::float8 AS cost_price
		FROM v_production_return_order_list
		WHERE ($1 = '' OR return_no ILIKE $1)
		ORDER BY created_at DESC
		LIMIT 50
	`
	kw := "%" + strings.TrimSpace(keyword) + "%"
	rows, err := database.Pool.Query(ctx, sql, kw)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]interface{}
	for rows.Next() {
		var id int64
		var no, matName string
		var costPrice float64
		if err := rows.Scan(&id, &no, &matName, &costPrice); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{
			"id":            id,
			"return_no":     no,
			"material_name": matName,
			"cost_price":    costPrice,
		})
	}
	return out, rows.Err()
}

// productionOrderStatusName 生产单状态中文映射
func productionOrderStatusName(status string) string {
	switch status {
	case "completed": return "已完成"
	case "draft": return "待提交"
	default: return status
	}
}

// productionReturnStatusName 生产退货单状态中文映射
func productionReturnStatusName(status string) string {
	switch status {
	case "created": return "待处理"
	case "confirmed": return "已确认"
	case "cancelled": return "已取消"
	default: return status
	}
}

// CreateProductionReturnOrder 创建生产退货单
func CreateProductionReturnOrder(ctx context.Context, req request.CreateProductionReturnOrderReq, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil { return 0, err }
	defer tx.Rollback(ctx)

	var po struct {
		productionNo string; consumptionOrderID int64; producedMaterialID int64
		producedWarehouseID int64; producedQuantity float64; producedUnitCost float64
	}
	err = tx.QueryRow(ctx, `SELECT production_no, consumption_order_id, produced_material_id, produced_warehouse_id, produced_quantity, produced_unit_cost FROM production_order WHERE id = $1 FOR UPDATE`, req.ProductionOrderID).Scan(&po.productionNo, &po.consumptionOrderID, &po.producedMaterialID, &po.producedWarehouseID, &po.producedQuantity, &po.producedUnitCost)
	if err != nil {
		if err == pgx.ErrNoRows { return 0, errcode.NewAppError(errcode.ErrNotFound.Code, "生产单不存在", errcode.ErrNotFound.HTTPCode) }
		return 0, err
	}
	if req.ReturnedQuantity > po.producedQuantity {
		return 0, errcode.NewAppError(errcode.ErrInvalidParam.Code, fmt.Sprintf("DB_ERROR: 退货数量(%.3f)不能超过生产数量(%.3f)", req.ReturnedQuantity, po.producedQuantity), errcode.ErrInvalidParam.HTTPCode)
	}

	var returnNo string
	if err = tx.QueryRow(ctx, "SELECT fn_generate_serial_no('PRR')").Scan(&returnNo); err != nil { return 0, err }

	var stockOutID int64
	if err = tx.QueryRow(ctx, `INSERT INTO stock_out (stock_out_no, stock_out_date, out_type, ref_doc_type, status, remark, created_by, updated_by) VALUES ($1, CURRENT_DATE, 'production_return', 'production_return', 'draft', '生产退货-成品退回', $2, $3) RETURNING id`, "SO"+time.Now().Format("20060102150405"), userID, userID).Scan(&stockOutID); err != nil { return 0, err }

	var invID int64
	err = tx.QueryRow(ctx, `SELECT i.id FROM inventory i WHERE i.material_id = $1 AND i.warehouse_id = $2 AND i.quantity > 0 ORDER BY i.id LIMIT 1 FOR UPDATE`, po.producedMaterialID, po.producedWarehouseID).Scan(&invID)
	if err != nil { return 0, fmt.Errorf("DB_ERROR: 成品库存不足") }

	if _, err = tx.Exec(ctx, `INSERT INTO stock_out_item (stock_out_id, material_id, inventory_id, quantity, unit, remark) SELECT $1, $2, $3, $4, COALESCE(m.unit,'pcs'), '生产退货' FROM material m WHERE m.id=$2`, stockOutID, po.producedMaterialID, invID, req.ReturnedQuantity); err != nil { return 0, err }
	if _, err = tx.Exec(ctx, `CALL sp_confirm_stock_out($1, $2)`, stockOutID, userID); err != nil { return 0, fmt.Errorf("DB_ERROR: %s", err.Error()) }

	var returnID int64
	err = tx.QueryRow(ctx, `INSERT INTO production_return_order (return_no, production_order_id, stock_out_id, returned_quantity, status, remark, created_by, updated_by) VALUES ($1, $2, $3, $4, 'confirmed', $5, $6, $7) RETURNING id`, returnNo, req.ProductionOrderID, stockOutID, req.ReturnedQuantity, req.Remark, userID, userID).Scan(&returnID)
	if err != nil { return 0, err }

	_, _ = tx.Exec(ctx, `CALL sp_write_audit_log($1,$2,$3,$4,$5,$6,$7)`, userID, username, "CREATE", "PROD_RETURN", "production_return_order", returnID, nil)
	return returnID, tx.Commit(ctx)
}

// CreateProductionOrder 手动创建生产单
func CreateProductionOrder(ctx context.Context, req request.CreateProductionOrderReq, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil { return 0, err }
	defer tx.Rollback(ctx)

	var matCode, matName, matUnit string
	err = tx.QueryRow(ctx, `SELECT material_code, material_name, COALESCE(unit,'pcs') FROM material WHERE id=$1 AND deleted_at IS NULL`,
		req.ProducedMaterialID).Scan(&matCode, &matName, &matUnit)
	if err != nil { return 0, fmt.Errorf("DB_ERROR: 成品物料不存在") }

	var whCode, whName string
	err = tx.QueryRow(ctx, `SELECT warehouse_code, warehouse_name FROM warehouse WHERE id=$1 AND deleted_at IS NULL`,
		req.ProducedWarehouseID).Scan(&whCode, &whName)
	if err != nil { return 0, fmt.Errorf("DB_ERROR: 成品仓库不存在") }

	var prodNo string
	if err = tx.QueryRow(ctx, "SELECT fn_generate_serial_no('PR')").Scan(&prodNo); err != nil { return 0, err }

	unitCost := float64(0)
	if req.ProducedQuantity > 0 && req.CostPrice > 0 {
		unitCost = req.CostPrice / req.ProducedQuantity
	}

	// 创建入库单
	var siNo string
	if err = tx.QueryRow(ctx, "SELECT fn_generate_serial_no('SI')").Scan(&siNo); err != nil { return 0, err }
	var siID int64
	err = tx.QueryRow(ctx, `INSERT INTO stock_in (stock_in_no, warehouse_id, warehouse_code, stock_in_date, stock_in_status, stock_in_type, remark, created_by, updated_by)
		VALUES ($1,$2,$3,CURRENT_DATE,'preparing','production','手动创建生产单',$4,$5) RETURNING id`,
		siNo, req.ProducedWarehouseID, whCode, userID, userID).Scan(&siID)
	if err != nil { return 0, err }

	_, err = tx.Exec(ctx, `INSERT INTO stock_in_item (stock_in_id, material_id, arrived_quantity, accepted_quantity, unit, unit_cost, custom_attributes, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,'[]'::jsonb,NOW())`,
		siID, req.ProducedMaterialID, req.ProducedQuantity, req.ProducedQuantity, matUnit, unitCost)
	if err != nil { return 0, err }

	// 确认入库
	if _, err = tx.Exec(ctx, `CALL sp_confirm_stock_in($1,$2)`, siID, userID); err != nil {
		return 0, fmt.Errorf("DB_ERROR: %s", err.Error())
	}

	// 创建生产单
	var consumptionID interface{} = nil
	if req.ConsumptionOrderID > 0 { consumptionID = req.ConsumptionOrderID }
	var id int64
	err = tx.QueryRow(ctx, `INSERT INTO production_order (production_no, consumption_order_id, stock_in_id,
		produced_material_id, produced_warehouse_id, produced_quantity, produced_unit_cost, cost_price,
		status, remark, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'completed',$9,$10,$11) RETURNING id`,
		prodNo, consumptionID, siID, req.ProducedMaterialID, req.ProducedWarehouseID,
		req.ProducedQuantity, unitCost, req.CostPrice, req.Remark, userID, userID).Scan(&id)
	if err != nil { return 0, err }

	_, _ = tx.Exec(ctx, `CALL sp_write_audit_log($1,$2,$3,$4,$5,$6,$7)`, userID, username, "CREATE", "PRODUCTION", "production_order", id, nil)
	return id, tx.Commit(ctx)
}
