/**
 * 功能：采购订单数据库操作
 * 创建时间：2026-04-18
 * 创建人：CodeArts Agent
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

func ensurePurchaseOrderNotReceived(ctx context.Context, id int64) error {
	var status string
	err := database.Pool.QueryRow(ctx, `
		SELECT po.order_status
		FROM purchase_order po
		WHERE po.id = $1 AND po.deleted_at IS NULL
	`, id).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errcode.ErrNotFound
		}
		return err
	}
	if status != "draft" && status != "ordered" {
		return errcode.NewAppError(errcode.ErrForbidden.Code, "仅待提交或已下单且未入库的单据才允许编辑或删除", errcode.ErrForbidden.HTTPCode)
	}

	if status == "ordered" {
		var hasReceived bool
		err = database.Pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM purchase_order_item
				WHERE order_id = $1 AND COALESCE(received_quantity, 0) > 0
			)
		`, id).Scan(&hasReceived)
		if err != nil {
			return err
		}
		if hasReceived {
			return errcode.NewAppError(errcode.ErrForbidden.Code, "已有入库记录，不可编辑或删除", errcode.ErrForbidden.HTTPCode)
		}
	}
	return nil
}

func ListPurchaseOrders(ctx context.Context, q *request.PurchaseOrderQuery) ([]response.PurchaseOrderResp, int64, error) {
	where := []string{"po.deleted_at IS NULL"}
	var args []interface{}
	argID := 1

	if q.OrderNo != "" {
		where = append(where, fmt.Sprintf("po.order_no ILIKE $%d", argID))
		args = append(args, "%"+q.OrderNo+"%")
		argID++
	}
	if q.SupplierCode != "" {
		where = append(where, fmt.Sprintf("po.supplier_code = $%d", argID))
		args = append(args, q.SupplierCode)
		argID++
	}
	if strings.TrimSpace(q.SupplierKeyword) != "" {
		where = append(where, fmt.Sprintf(`(
			po.supplier_code ILIKE $%d OR EXISTS (
				SELECT 1
				FROM supplier s_kw
				WHERE s_kw.supplier_code = po.supplier_code
				  AND s_kw.deleted_at IS NULL
				  AND s_kw.supplier_name ILIKE $%d
			)
		)`, argID, argID))
		args = append(args, "%"+strings.TrimSpace(q.SupplierKeyword)+"%")
		argID++
	}
	if q.OrderStatus != "" {
		where = append(where, fmt.Sprintf("po.order_status = $%d", argID))
		args = append(args, q.OrderStatus)
		argID++
	}
	if q.OrderType != "" {
		where = append(where, fmt.Sprintf("po.order_type = $%d", argID))
		args = append(args, q.OrderType)
		argID++
	}
	if q.StartDate != "" {
		where = append(where, fmt.Sprintf("po.order_date >= $%d", argID))
		args = append(args, q.StartDate)
		argID++
	}
	if q.EndDate != "" {
		where = append(where, fmt.Sprintf("po.order_date <= $%d", argID))
		args = append(args, q.EndDate)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM purchase_order po WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT po.id, po.order_no,
		       po.supplier_code, COALESCE(s.supplier_name, ''),
		       po.order_date, po.expected_date, po.order_status, po.total_amount,
		       po.remark, po.created_at, po.updated_at,
		       COALESCE((
		           SELECT si.id
		           FROM stock_in si
		           WHERE si.purchase_order_id = po.id
		             AND si.deleted_at IS NULL
		           ORDER BY si.id DESC
		           LIMIT 1
		       ), 0) AS stock_in_id,
		       COALESCE((
		           SELECT si.stock_in_no
		           FROM stock_in si
		           WHERE si.purchase_order_id = po.id
		             AND si.deleted_at IS NULL
		           ORDER BY si.id DESC
		           LIMIT 1
		       ), '') AS stock_in_no
		FROM purchase_order po
		LEFT JOIN supplier s ON s.supplier_code = po.supplier_code
		WHERE %s
		ORDER BY po.id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.PurchaseOrderResp
	for rows.Next() {
		var item response.PurchaseOrderResp
		if err := rows.Scan(&item.ID, &item.OrderNo,
			&item.SupplierCode, &item.SupplierName,
			&item.OrderDate, &item.ExpectedDate, &item.OrderStatus, &item.TotalAmount,
			&item.Remark, &item.CreatedAt, &item.UpdatedAt, &item.StockInID, &item.StockInNo); err != nil {
			return nil, 0, err
		}
		item.WarehouseCode = ""
		item.WarehouseName = ""
		item.OrderStatusName = getOrderStatusName(item.OrderStatus)
		result = append(result, item)
	}

	return result, total, rows.Err()
}

func GetPurchaseOrderByID(ctx context.Context, id int64) (*response.PurchaseOrderDetailResp, error) {
	var order response.PurchaseOrderDetailResp
	query := `
		SELECT po.id, po.order_no,
		       po.supplier_code, COALESCE(s.supplier_name, ''),
		       po.order_date, po.expected_date, po.order_status, po.total_amount,
		       po.remark, po.created_at, po.updated_at,
		       COALESCE((
		           SELECT si.id FROM stock_in si
		           WHERE si.purchase_order_id = po.id AND si.deleted_at IS NULL
		           ORDER BY si.id DESC LIMIT 1
		       ), 0),
		       COALESCE((
		           SELECT si.stock_in_no FROM stock_in si
		           WHERE si.purchase_order_id = po.id AND si.deleted_at IS NULL
		           ORDER BY si.id DESC LIMIT 1
		       ), '')
		FROM purchase_order po
		LEFT JOIN supplier s ON s.supplier_code = po.supplier_code
		WHERE po.id = $1 AND po.deleted_at IS NULL
	`
	err := database.Pool.QueryRow(ctx, query, id).Scan(
		&order.ID, &order.OrderNo,
		&order.SupplierCode, &order.SupplierName,
		&order.OrderDate, &order.ExpectedDate, &order.OrderStatus, &order.TotalAmount,
		&order.Remark, &order.CreatedAt, &order.UpdatedAt,
		&order.StockInID, &order.StockInNo,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	order.WarehouseCode = ""
	order.WarehouseName = ""
	order.OrderStatusName = getOrderStatusName(order.OrderStatus)

	itemsQuery := `
		SELECT poi.id, poi.material_id,
		       m.material_code, m.material_name,
		       poi.quantity, COALESCE(m.unit, ''), poi.unit_price, poi.amount, poi.received_quantity
		FROM purchase_order_item poi
		LEFT JOIN material m ON m.id = poi.material_id
		WHERE poi.order_id = $1
		ORDER BY poi.id
	`
	rows, err := database.Pool.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item response.PurchaseOrderItemResp
		if err := rows.Scan(&item.ID, &item.MaterialID,
			&item.MaterialCode, &item.MaterialName,
			&item.Quantity, &item.Unit, &item.UnitPrice, &item.Amount,
			&item.ReceivedQuantity); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	// 查询关联的入库单
	stockInQuery := `
		SELECT id, stock_in_no
		FROM stock_in
		WHERE purchase_order_id = $1 AND deleted_at IS NULL
		ORDER BY id DESC
	`
	stockInRows, err := database.Pool.Query(ctx, stockInQuery, id)
	if err == nil {
		defer stockInRows.Close()
		for stockInRows.Next() {
			var si response.RelatedStockIn
			if err := stockInRows.Scan(&si.ID, &si.OrderNo); err == nil {
				order.StockInRecords = append(order.StockInRecords, si)
			}
		}
	}

	return &order, rows.Err()
}

func CreatePurchaseOrder(ctx context.Context, req *request.CreatePurchaseOrderReq, userID int64) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var supplierID int64
	err = tx.QueryRow(ctx, `
		SELECT id FROM supplier
		WHERE supplier_code = $1 AND deleted_at IS NULL
	`, req.SupplierCode).Scan(&supplierID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("DB_ERROR: 供应商不存在或已删除")
		}
		return 0, err
	}

	orderNo, err := generateOrderNo(ctx, tx, "PO")
	if err != nil {
		return 0, err
	}

	var orderID int64
	orderType := req.OrderType
	if orderType == "" {
		orderType = "purchase"
	}

	var expectedDate interface{}
	if strings.TrimSpace(req.ExpectedDate) != "" {
		expectedDate = strings.TrimSpace(req.ExpectedDate)
	}

	orderQuery := `
		INSERT INTO purchase_order (order_no, supplier_id, supplier_code, buyer_id, order_date, expected_date, order_type, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	err = tx.QueryRow(ctx, orderQuery, orderNo, supplierID, req.SupplierCode, userID, req.OrderDate, expectedDate, orderType, userID).Scan(&orderID)
	if err != nil {
		return 0, err
	}

	itemQuery := `
		INSERT INTO purchase_order_item (order_id, material_id, quantity, unit_price, custom_attributes)
		VALUES ($1, $2, $3, $4, $5)
	`
	for _, item := range req.Items {
		materialID := item.MaterialID
		customAttributes := []byte("[]")
		if materialID == 0 {
			return 0, fmt.Errorf("DB_ERROR: 采购订货明细必须选择物料")
		}
		if orderType == "purchase" {
			err := tx.QueryRow(ctx, `
				SELECT COALESCE(custom_attributes, '[]'::jsonb)
				FROM material WHERE id = $1 AND deleted_at IS NULL
			`, materialID).Scan(&customAttributes)
			if err != nil {
				if err == pgx.ErrNoRows {
					return 0, fmt.Errorf("DB_ERROR: 物料不存在或已删除")
				}
				return 0, err
			}
		}
		_, err := tx.Exec(ctx, itemQuery, orderID, materialID, item.Quantity, item.UnitPrice, customAttributes)
		if err != nil {
			return 0, err
		}
	}

	if _, err := tx.Exec(ctx, `UPDATE purchase_order SET total_amount = (SELECT SUM(quantity * unit_price) FROM purchase_order_item WHERE order_id = $1) WHERE id = $1`, orderID); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	if err := ConfirmPurchaseOrder(ctx, orderID, userID); err != nil {
		return 0, fmt.Errorf("订单创建成功但自动确认失败: %w", err)
	}

	return orderID, nil
}

func UpdatePurchaseOrder(ctx context.Context, id int64, req *request.UpdatePurchaseOrderReq, userID int64) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var orderStatus string
	var orderType string
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(order_status, 'draft'), COALESCE(order_type, 'purchase')
		FROM purchase_order WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, id).Scan(&orderStatus, &orderType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errcode.ErrNotFound
		}
		return err
	}

	if orderStatus != "draft" && orderStatus != "ordered" {
		return errcode.NewAppError(errcode.ErrForbidden.Code, "仅待提交或已下单且未入库的单据允许编辑", errcode.ErrForbidden.HTTPCode)
	}

	var hasReceived bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM purchase_order_item
			WHERE order_id = $1 AND COALESCE(received_quantity, 0) > 0
		)
	`, id).Scan(&hasReceived)
	if err != nil {
		return err
	}
	if hasReceived {
		return errcode.NewAppError(errcode.ErrForbidden.Code, "已有入库记录，不可编辑", errcode.ErrForbidden.HTTPCode)
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM stock_in_item WHERE stock_in_id IN (
			SELECT id FROM stock_in WHERE purchase_order_id = $1 AND stock_in_status = 'preparing'
		)
	`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM stock_in WHERE purchase_order_id = $1 AND stock_in_status = 'preparing'
	`, id); err != nil {
		return err
	}

	var expectedDate interface{}
	if strings.TrimSpace(req.ExpectedDate) != "" {
		expectedDate = strings.TrimSpace(req.ExpectedDate)
	}

	if _, err = tx.Exec(ctx, `
		UPDATE purchase_order
		SET expected_date = $1, remark = $2, updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`, expectedDate, req.Remark, id); err != nil {
		return err
	}

	if len(req.Items) > 0 {
		if _, err = tx.Exec(ctx, "DELETE FROM purchase_order_item WHERE order_id = $1", id); err != nil {
			return err
		}

		for _, item := range req.Items {
			materialID := item.MaterialID
			customAttributes := []byte("[]")
			if materialID == 0 {
							return fmt.Errorf("DB_ERROR: 采购订货明细必须选择物料")
			}
			if orderType == "purchase" {
				if err := tx.QueryRow(ctx, `
					SELECT COALESCE(custom_attributes, '[]'::jsonb)
					FROM material WHERE id = $1 AND deleted_at IS NULL
				`, materialID).Scan(&customAttributes); err != nil {
					if err == pgx.ErrNoRows {
											return fmt.Errorf("DB_ERROR: 物料不存在或已删除")
					}
					return err
				}
			}
			if _, err = tx.Exec(ctx, `
				INSERT INTO purchase_order_item (order_id, material_id, quantity, unit_price, custom_attributes)
				VALUES ($1, $2, $3, $4, $5)
			`, id, materialID, item.Quantity, item.UnitPrice, customAttributes); err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	if orderStatus == "ordered" {
		return nil
	}

	return ConfirmPurchaseOrder(ctx, id, userID)
}

func DeletePurchaseOrder(ctx context.Context, id int64) error {
	if err := ensurePurchaseOrderNotReceived(ctx, id); err != nil {
		return err
	}
	query := `UPDATE purchase_order SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := database.Pool.Exec(ctx, query, id)
	return err
}

func ConfirmPurchaseOrder(ctx context.Context, orderID, userID int64) error {
	var status string
	err := database.Pool.QueryRow(ctx, `
		SELECT COALESCE(order_status, 'draft')
		FROM purchase_order
		WHERE id = $1 AND deleted_at IS NULL
	`, orderID).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errcode.ErrNotFound
		}
		return err
	}
	if status == "ordered" {
		return nil
	}
	if status != "draft" {
		return errcode.NewAppError(errcode.ErrForbidden.Code, fmt.Sprintf("采购订单当前状态[%s]不允许确认", status), errcode.ErrForbidden.HTTPCode)
	}

	_, err = database.Pool.Exec(ctx, `CALL sp_confirm_purchase_order($1, $2)`, orderID, userID)
	if err != nil {
		return fmt.Errorf("DB_ERROR: %s", strings.TrimSpace(err.Error()))
	}
	return nil
}

func generateOrderNo(ctx context.Context, tx pgx.Tx, prefix string) (string, error) {
	dateStr := time.Now().Format("20060102")

	var seq int
	seqQuery := `
		INSERT INTO sys_serial_number (prefix, date_str, current_seq)
		VALUES ($1, $2, 1)
		ON CONFLICT (prefix, date_str) 
		DO UPDATE SET current_seq = sys_serial_number.current_seq + 1
		RETURNING current_seq
	`
	err := tx.QueryRow(ctx, seqQuery, prefix, dateStr).Scan(&seq)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s%03d", prefix, dateStr, seq), nil
}

func getOrderStatusName(status string) string {
	statusMap := map[string]string{
		"draft":            "待提交",
		"ordered":          "已下单",
		"partial_received": "部分到货",
		"full_received":    "已完成",
		"closed":           "已关闭",
	}
	if name, ok := statusMap[status]; ok {
		return name
	}
	return status
}
