/**
 * 功能：sales_order.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/pkg/database"
)

func salesOrderStatusName(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "draft":
		return "待提交"
	case "confirmed":
		return "待出库"
	case "preparing":
		return "出库中"
	case "shipped":
		return "已完成"
	case "cancelled":
		return "已取消"
	default:
		return status
	}
}

// ListSalesOrders 分页查询销售订单
func ListSalesOrders(ctx context.Context, q *request.SalesOrderQuery) ([]response.SalesOrderResp, int64, error) {
	where := []string{"so.deleted_at IS NULL"}
	var args []interface{}
	argID := 1

	if strings.TrimSpace(q.OrderNo) != "" {
		where = append(where, fmt.Sprintf("so.order_no ILIKE $%d", argID))
		args = append(args, "%"+q.OrderNo+"%")
		argID++
	}
	if strings.TrimSpace(q.CustomerCode) != "" {
		where = append(where, fmt.Sprintf("so.customer_code = $%d", argID))
		args = append(args, q.CustomerCode)
		argID++
	}
	if strings.TrimSpace(q.CustomerKeyword) != "" {
		where = append(where, fmt.Sprintf("(COALESCE(so.customer_name, c.customer_name, '') ILIKE $%d OR so.customer_code ILIKE $%d)", argID, argID))
		args = append(args, "%"+strings.TrimSpace(q.CustomerKeyword)+"%")
		argID++
	}
	if strings.TrimSpace(q.Status) != "" {
		where = append(where, fmt.Sprintf("so.order_status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}
	if strings.TrimSpace(q.StartDate) != "" {
		where = append(where, fmt.Sprintf("so.order_date::date >= $%d::date", argID))
		args = append(args, q.StartDate)
		argID++
	}
	if strings.TrimSpace(q.EndDate) != "" {
		where = append(where, fmt.Sprintf("so.order_date::date <= $%d::date", argID))
		args = append(args, q.EndDate)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`
		SELECT count(*)
		FROM sales_order so
		LEFT JOIN customer c ON c.customer_code = so.customer_code AND c.deleted_at IS NULL
		WHERE %s
	`, whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT so.id, so.order_no, so.customer_code,
		       COALESCE(so.customer_name, c.customer_name, ''),
		       COALESCE(so.sales_person_id, 0),
		       COALESCE(u.real_name, ''),
		       so.order_date, so.delivery_date,
		       COALESCE(so.contract_no, ''),
		       COALESCE(so.payment_method, ''),
		       COALESCE(so.receiver_name, ''),
		       COALESCE(so.receiver_phone, ''),
		       COALESCE(so.receiver_address, ''),
		       so.order_status,
		       COALESCE(so.total_amount, 0),
		       COALESCE(so.remark, ''),
		       so.created_at,
		       COALESCE(so.stock_out_id, 0),
		       COALESCE(sout.stock_out_no, '')
		FROM sales_order so
		LEFT JOIN customer c ON c.customer_code = so.customer_code AND c.deleted_at IS NULL
		LEFT JOIN sys_user u ON u.id = so.sales_person_id
		LEFT JOIN stock_out sout ON sout.id = so.stock_out_id AND sout.deleted_at IS NULL
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

	var result []response.SalesOrderResp
	for rows.Next() {
		var item response.SalesOrderResp
		if err := rows.Scan(&item.ID, &item.OrderNo, &item.CustomerCode, &item.CustomerName, &item.SalesPersonID,
			&item.SalesPersonName, &item.OrderDate, &item.DeliveryDate,
			&item.ContractNo, &item.PaymentMethod, &item.ReceiverName, &item.ReceiverPhone, &item.ReceiverAddress,
			&item.OrderStatus, &item.TotalAmount, &item.Remark, &item.CreatedAt,
			&item.StockOutID, &item.StockOutNo); err != nil {
			return nil, 0, err
		}
		item.OrderStatusName = salesOrderStatusName(item.OrderStatus)
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// GetSalesOrderDetail 获取销售订单详情
func GetSalesOrderDetail(ctx context.Context, id int64) (*response.SalesOrderResp, error) {
	query := `
		SELECT so.id, so.order_no, so.customer_code,
		       COALESCE(so.customer_name, c.customer_name, ''),
		       COALESCE(so.sales_person_id, 0),
		       COALESCE(u.real_name, ''),
		       so.order_date, so.delivery_date,
		       COALESCE(so.contract_no, ''),
		       COALESCE(so.payment_method, ''),
		       COALESCE(so.receiver_name, ''),
		       COALESCE(so.receiver_phone, ''),
		       COALESCE(so.receiver_address, ''),
		       so.order_status, COALESCE(so.total_amount, 0), COALESCE(so.remark, ''), so.created_at,
		       COALESCE(so.stock_out_id, 0), COALESCE(sout.stock_out_no, '')
		FROM sales_order so
		LEFT JOIN customer c ON c.customer_code = so.customer_code AND c.deleted_at IS NULL
		LEFT JOIN sys_user u ON u.id = so.sales_person_id
		LEFT JOIN stock_out sout ON sout.id = so.stock_out_id AND sout.deleted_at IS NULL
		WHERE so.id = $1 AND so.deleted_at IS NULL
	`
	var item response.SalesOrderResp
	err := database.Pool.QueryRow(ctx, query, id).Scan(&item.ID, &item.OrderNo, &item.CustomerCode, &item.CustomerName, &item.SalesPersonID,
		&item.SalesPersonName, &item.OrderDate, &item.DeliveryDate,
		&item.ContractNo, &item.PaymentMethod, &item.ReceiverName, &item.ReceiverPhone, &item.ReceiverAddress,
		&item.OrderStatus, &item.TotalAmount, &item.Remark, &item.CreatedAt,
		&item.StockOutID, &item.StockOutNo)
	if err != nil {
		return nil, err
	}
	item.OrderStatusName = salesOrderStatusName(item.OrderStatus)

	itemQuery := `
		SELECT soi.id, soi.material_id,
		       COALESCE(soi.material_code, m.material_code, ''),
		       COALESCE(soi.material_name, m.material_name, ''),
		       soi.quantity, soi.unit_price, soi.amount, soi.shipped_quantity, 
		       COALESCE(soi.unit, m.unit, ''), soi.remark
		FROM sales_order_item soi
		LEFT JOIN material m ON m.id = soi.material_id AND m.deleted_at IS NULL
		WHERE soi.order_id = $1
	`
	rows, err := database.Pool.Query(ctx, itemQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sub response.SalesOrderItemResp
		if err := rows.Scan(&sub.ID, &sub.MaterialID, &sub.MaterialCode, &sub.MaterialName,
			&sub.Quantity, &sub.UnitPrice, &sub.Amount, &sub.ShippedQuantity, &sub.Unit, &sub.Remark); err != nil {
			return nil, err
		}
		item.Items = append(item.Items, sub)
	}

	return &item, rows.Err()
}

// UpdateSalesOrder 修改待出库状态的销售订单并重新生成出库单
func UpdateSalesOrder(ctx context.Context, id int64, req *request.UpdateSalesOrderReq, userID int64, username string) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var orderStatus string
	var stockOutID int64
	err = tx.QueryRow(ctx, `
		SELECT order_status, COALESCE(stock_out_id, 0)
		FROM sales_order
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, id).Scan(&orderStatus, &stockOutID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errcode.ErrNotFound
		}
		return err
	}

	if orderStatus != "confirmed" && orderStatus != "preparing" {
		return errcode.NewAppError(errcode.ErrForbidden.Code, "仅待出库/出库中状态可编辑", errcode.ErrForbidden.HTTPCode)
	}

	if stockOutID > 0 {
		if _, err := tx.Exec(ctx, `
			UPDATE inventory
			SET locked_quantity = GREATEST(locked_quantity - src.qty, 0),
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

	if _, err := tx.Exec(ctx, `DELETE FROM sales_order_item WHERE order_id = $1`, id); err != nil {
		return err
	}

	customerCode := strings.TrimSpace(req.CustomerCode)
	var customerName string
	if err := tx.QueryRow(ctx, `
		SELECT customer_name FROM customer
		WHERE customer_code = $1 AND deleted_at IS NULL AND status = 'enabled'
	`, customerCode).Scan(&customerName); err != nil {
		if err == pgx.ErrNoRows {
			return errcode.NewAppError(errcode.ErrInvalidParam.Code, "客户不存在或已停用", errcode.ErrInvalidParam.HTTPCode)
		}
		return err
	}

	orderDate := strings.TrimSpace(req.OrderDate)
	deliveryDate := strings.TrimSpace(req.DeliveryDate)
	if deliveryDate == "" {
		if parsed, parseErr := time.Parse("2006-01-02", orderDate); parseErr == nil {
			deliveryDate = parsed.AddDate(0, 0, 7).Format("2006-01-02")
		}
	}

	if _, err := tx.Exec(ctx, `
		UPDATE sales_order
		SET customer_code = $1, customer_name = $2,
		    sales_person_id = NULLIF($3, 0),
		    order_date = $4::date, delivery_date = NULLIF($5, '')::date,
		    contract_no = NULLIF($6, ''), payment_method = NULLIF($7, ''),
		    receiver_name = NULLIF($8, ''), receiver_phone = NULLIF($9, ''),
		    receiver_address = NULLIF($10, ''), remark = $11,
		    stock_out_id = NULL,
		    updated_by = $12, updated_at = NOW()
		WHERE id = $13
	`, customerCode, customerName, req.SalesPersonID, orderDate, deliveryDate,
		req.ContractNo, req.PaymentMethod, req.ReceiverName, req.ReceiverPhone,
		req.ReceiverAddress, req.Remark, userID, id); err != nil {
		return err
	}

	for _, item := range req.Items {
		var materialCode string
		var materialName string
		var unit string
		var materialStatus string
		if err := tx.QueryRow(ctx, `
			SELECT material_code, material_name, unit, status
			FROM material WHERE id = $1 AND deleted_at IS NULL
		`, item.MaterialID).Scan(&materialCode, &materialName, &unit, &materialStatus); err != nil {
			if err == pgx.ErrNoRows {
				return errcode.NewAppError(errcode.ErrInvalidParam.Code, fmt.Sprintf("物料不存在 [id=%d]", item.MaterialID), errcode.ErrInvalidParam.HTTPCode)
			}
			return err
		}
		if materialStatus == "disabled" {
			return errcode.NewAppError(errcode.ErrInvalidParam.Code, fmt.Sprintf("物料已停用 [%s]", materialName), errcode.ErrInvalidParam.HTTPCode)
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO sales_order_item (order_id, material_id, material_code, material_name, quantity, unit, unit_price, amount, remark)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $5::numeric * $7::numeric, $8)
		`, id, item.MaterialID, materialCode, materialName, item.Quantity, unit, item.UnitPrice, item.Remark); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, "UPDATE sales_order SET total_amount = (SELECT SUM(amount) FROM sales_order_item WHERE order_id = $1) WHERE id = $1", id); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `UPDATE sales_order SET order_status = 'draft' WHERE id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `CALL sp_confirm_sales_order($1, $2)`, id, userID); err != nil {
		return err
	}

	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, _ = tx.Exec(ctx, auditQuery, userID, username, "UPDATE", "SALES", "sales_order", id, nil)

	return tx.Commit(ctx)
}

// CreateSalesOrder 创建销售订单
func CreateSalesOrder(ctx context.Context, req *request.CreateSalesOrderReq, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var customerID int64
	var customerName string
	var customerStatus string
	var contactPerson string
	var contactPhone string
	var customerAddress string
	err = tx.QueryRow(ctx, `
		SELECT id, customer_name, status, COALESCE(contact_person, ''), COALESCE(contact_phone, ''), COALESCE(address, '')
		FROM customer
		WHERE customer_code = $1 AND deleted_at IS NULL
	`, req.CustomerCode).Scan(&customerID, &customerName, &customerStatus, &contactPerson, &contactPhone, &customerAddress)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, errcode.NewAppError(errcode.ErrInvalidParam.Code, "客户不存在或已删除", errcode.ErrInvalidParam.HTTPCode)
		}
		return 0, err
	}
	if customerStatus == "disabled" {
		return 0, errcode.NewAppError(errcode.ErrInvalidParam.Code, "客户已停用，不能创建销售订单", errcode.ErrInvalidParam.HTTPCode)
	}

	// 1. 生成单号
	var orderNo string
	err = tx.QueryRow(ctx, "SELECT fn_generate_serial_no('SO')").Scan(&orderNo)
	if err != nil {
		return 0, err
	}

	// 2. 插入主表
	var id int64
	receiverName := strings.TrimSpace(req.ReceiverName)
	if receiverName == "" {
		receiverName = contactPerson
	}
	receiverPhone := strings.TrimSpace(req.ReceiverPhone)
	if receiverPhone == "" {
		receiverPhone = contactPhone
	}
	receiverAddress := strings.TrimSpace(req.ReceiverAddress)
	if receiverAddress == "" {
		receiverAddress = customerAddress
	}
	orderDate := strings.TrimSpace(req.OrderDate)
	deliveryDate := strings.TrimSpace(req.DeliveryDate)
	if deliveryDate == "" {
		parsedOrderDate, parseErr := time.Parse("2006-01-02", orderDate)
		if parseErr == nil {
			deliveryDate = parsedOrderDate.AddDate(0, 0, 7).Format("2006-01-02")
		}
	}
	mainQuery := `
		INSERT INTO sales_order (
			order_no, customer_id, customer_code, customer_name, sales_person_id, order_date, delivery_date,
			contract_no, payment_method, receiver_name, receiver_phone, receiver_address,
			order_status, remark, created_by
		)
		VALUES ($1, $2, $3, $4, NULLIF($5, 0), $6::date, NULLIF($7, '')::date, NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''), NULLIF($12, ''), 'draft', $13, $14)
		RETURNING id
	`
	err = tx.QueryRow(ctx, mainQuery, orderNo, customerID, req.CustomerCode, customerName, req.SalesPersonID, orderDate,
		deliveryDate, strings.TrimSpace(req.ContractNo), strings.TrimSpace(req.PaymentMethod),
		receiverName, receiverPhone, receiverAddress, req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
	}

	// 3. 插入明细表
	for _, item := range req.Items {
		var materialCode string
		var materialName string
		var unit string
		var materialStatus string
		err = tx.QueryRow(ctx, `
			SELECT material_code, material_name, unit, status
			FROM material
			WHERE id = $1 AND deleted_at IS NULL
		`, item.MaterialID).Scan(&materialCode, &materialName, &unit, &materialStatus)
		if err != nil {
			if err == pgx.ErrNoRows {
				return 0, errcode.NewAppError(errcode.ErrInvalidParam.Code, fmt.Sprintf("物料不存在 [id=%d]", item.MaterialID), errcode.ErrInvalidParam.HTTPCode)
			}
			return 0, err
		}
		if materialStatus == "disabled" {
			return 0, errcode.NewAppError(errcode.ErrInvalidParam.Code, fmt.Sprintf("物料已停用，不能下单 [%s]", materialName), errcode.ErrInvalidParam.HTTPCode)
		}
		itemQuery := `
			INSERT INTO sales_order_item (order_id, material_id, material_code, material_name, quantity, unit, unit_price, amount, remark)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $5::numeric * $7::numeric, $8)
		`
		_, err = tx.Exec(ctx, itemQuery, id, item.MaterialID, materialCode, materialName, item.Quantity, unit, item.UnitPrice, item.Remark)
		if err != nil {
			return 0, err
		}
	}

	// 触发器/存储过程通常处理 total_amount 汇总。
	// 为简单起见，这里可以再手动刷一下汇总值（或者库里有触发器就不用）
	_, _ = tx.Exec(ctx, "UPDATE sales_order SET total_amount = (SELECT SUM(amount) FROM sales_order_item WHERE order_id = $1) WHERE id = $1", id)

	// 4. 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "CREATE", "SALES", "sales_order", id, nil)
	if err != nil {
		return 0, err
	}

	return id, tx.Commit(ctx)
}

// ConfirmSalesOrder 确认销售订单（锁定库存）
func ConfirmSalesOrder(ctx context.Context, orderID int64, userID int64, username string) error {
	query := `CALL sp_confirm_sales_order($1, $2)`
	_, err := database.Pool.Exec(ctx, query, orderID, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return fmt.Errorf("DB_ERROR: %s", pgErr.Message)
		}
		return err
	}
	return nil
}

// CancelSalesOrder 取消销售订单（释放库存）
func CancelSalesOrder(ctx context.Context, orderID int64, userID int64, username string) error {
	query := `CALL sp_cancel_sales_order($1, $2)`
	_, err := database.Pool.Exec(ctx, query, orderID, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return fmt.Errorf("DB_ERROR: %s", pgErr.Message)
		}
		return err
	}
	return nil
}

// ShipSalesOrder 发货
func ShipSalesOrder(ctx context.Context, orderID int64, req *request.ShipSalesOrderReq, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var existingStockOutID int64
	err = tx.QueryRow(ctx, `
		SELECT id
		FROM stock_out
		WHERE ref_doc_type = 'sales_order'
		  AND ref_doc_id = $1
		  AND deleted_at IS NULL
		  AND status IN ('draft', 'pending')
		ORDER BY id DESC
		LIMIT 1
	`, orderID).Scan(&existingStockOutID)
	if err == nil {
		return existingStockOutID, tx.Commit(ctx)
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	var orderNo string
	var orderStatus string
	var customerCode string
	var customerName string
	var receiverName string
	err = tx.QueryRow(ctx, `
		SELECT order_no, order_status, customer_code, COALESCE(customer_name, ''), COALESCE(receiver_name, '')
		FROM sales_order
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, orderID).Scan(&orderNo, &orderStatus, &customerCode, &customerName, &receiverName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errcode.NewAppError(errcode.ErrInvalidParam.Code, "销售订单不存在", errcode.ErrInvalidParam.HTTPCode)
		}
		return 0, err
	}
	if orderStatus != "confirmed" && orderStatus != "preparing" {
		return 0, errcode.NewAppError(errcode.ErrInvalidParam.Code, "当前状态不允许生成销售出库单", errcode.ErrInvalidParam.HTTPCode)
	}

	rows, err := tx.Query(ctx, `
		SELECT material_id, quantity - shipped_quantity AS remain_qty
		FROM sales_order_item
		WHERE order_id = $1
		  AND quantity > shipped_quantity
		ORDER BY id
	`, orderID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type pendingItem struct {
		materialID int64
		quantity   float64
	}
	var items []pendingItem
	for rows.Next() {
		var item pendingItem
		if err := rows.Scan(&item.materialID, &item.quantity); err != nil {
			return 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, errcode.NewAppError(errcode.ErrInvalidParam.Code, "销售订单已全部出库，无需重复生成", errcode.ErrInvalidParam.HTTPCode)
	}

	var stockOutNo string
	err = tx.QueryRow(ctx, "SELECT fn_generate_serial_no('SO')").Scan(&stockOutNo)
	if err != nil {
		return 0, err
	}

	stockOutDate := strings.TrimSpace(req.StockOutDate)
	if stockOutDate == "" {
		stockOutDate = time.Now().Format("2006-01-02")
	}
	receiver := strings.TrimSpace(receiverName)
	if receiver == "" {
		receiver = strings.TrimSpace(customerName)
	}

	var stockOutID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO stock_out (
			stock_out_no, stock_out_date, out_type, ref_doc_type, ref_doc_id,
			customer_code, customer_name, receiver, status, remark, created_by
		)
		VALUES ($1, $2, 'sales', 'sales_order', $3, NULLIF($4, ''), NULLIF($5, ''), $6, 'draft', $7, $8)
		RETURNING id
	`, stockOutNo, stockOutDate, orderID, customerCode, customerName, receiver, strings.TrimSpace(req.Remark), userID).Scan(&stockOutID)
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		_, err = tx.Exec(ctx, `
			INSERT INTO stock_out_item (stock_out_id, material_id, quantity, unit)
			SELECT $1, $2, $3, COALESCE(soi.unit, m.unit, '')
			FROM sales_order_item soi
			LEFT JOIN material m ON m.id = soi.material_id
			WHERE soi.order_id = $4 AND soi.material_id = $2
			ORDER BY soi.id
			LIMIT 1
		`, stockOutID, item.materialID, item.quantity, orderID)
		if err != nil {
			return 0, err
		}
	}

	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, _ = tx.Exec(ctx, auditQuery, userID, username, "CREATE_STOCK_OUT", "SALES", "stock_out", stockOutID, nil)

	return stockOutID, tx.Commit(ctx)
}
