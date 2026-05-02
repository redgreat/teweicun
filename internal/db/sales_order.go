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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

// ListSalesOrders 分页查询销售订单
func ListSalesOrders(ctx context.Context, q *request.SalesOrderQuery) ([]response.SalesOrderResp, int64, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if q.OrderNo != "" {
		where = append(where, fmt.Sprintf("order_no ILIKE $%d", argID))
		args = append(args, "%"+q.OrderNo+"%")
		argID++
	}
	if q.CustomerCode != "" {
		where = append(where, fmt.Sprintf("customer_code = $%d", argID))
		args = append(args, q.CustomerCode)
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("order_status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_sales_order_list WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, order_no, customer_code, customer_name, sales_person_id,
		       sales_person_name, order_date, delivery_date, order_status, order_status_name,
		       total_amount, remark, created_at
		FROM v_sales_order_list
		WHERE %s
		ORDER BY id DESC
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
			&item.SalesPersonName, &item.OrderDate, &item.DeliveryDate, &item.OrderStatus, &item.OrderStatusName,
			&item.TotalAmount, &item.Remark, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// GetSalesOrderDetail 获取销售订单详情
func GetSalesOrderDetail(ctx context.Context, id int64) (*response.SalesOrderResp, error) {
	query := `
		SELECT so.id, so.order_no, so.customer_code, c.customer_name, so.sales_person_id, 
		       u.real_name, so.order_date, so.delivery_date, so.order_status, 
		       so.total_amount, so.remark, so.created_at
		FROM sales_order so
		LEFT JOIN customer c ON c.customer_code = so.customer_code
		LEFT JOIN sys_user u ON u.id = so.sales_person_id
		WHERE so.id = $1 AND so.deleted_at IS NULL
	`
	var item response.SalesOrderResp
	err := database.Pool.QueryRow(ctx, query, id).Scan(&item.ID, &item.OrderNo, &item.CustomerCode, &item.CustomerName, &item.SalesPersonID,
		&item.SalesPersonName, &item.OrderDate, &item.DeliveryDate, &item.OrderStatus, &item.TotalAmount, &item.Remark, &item.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 查明细
	itemQuery := `
		SELECT soi.id, soi.material_id, m.material_code, m.material_name, 
		       soi.quantity, soi.unit_price, soi.amount, soi.shipped_quantity, 
		       m.unit, soi.remark
		FROM sales_order_item soi
		INNER JOIN material m ON m.id = soi.material_id
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

// CreateSalesOrder 创建销售订单
func CreateSalesOrder(ctx context.Context, req *request.CreateSalesOrderReq, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var customerID int64
	err = tx.QueryRow(ctx, `
		SELECT id FROM customer
		WHERE customer_code = $1 AND deleted_at IS NULL
	`, req.CustomerCode).Scan(&customerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("客户不存在或已删除")
		}
		return 0, err
	}

	// 1. 生成单号
	var orderNo string
	err = tx.QueryRow(ctx, "SELECT fn_generate_serial_no('SO')").Scan(&orderNo)
	if err != nil {
		return 0, err
	}

	// 2. 插入主表
	var id int64
	mainQuery := `
		INSERT INTO sales_order (order_no, customer_id, customer_code, sales_person_id, order_date, delivery_date, 
		                        order_status, remark, created_by)
		VALUES ($1, $2, $3, NULLIF($4, 0), $5, $6, 'draft', $7, $8)
		RETURNING id
	`
	err = tx.QueryRow(ctx, mainQuery, orderNo, customerID, req.CustomerCode, req.SalesPersonID, req.OrderDate,
		req.DeliveryDate, req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
	}

	// 3. 插入明细表
	for _, item := range req.Items {
		itemQuery := `
			INSERT INTO sales_order_item (order_id, material_id, quantity, unit_price, amount, remark)
			VALUES ($1, $2, $3, $4, $3 * $4, $5)
		`
		_, err = tx.Exec(ctx, itemQuery, id, item.MaterialID, item.Quantity, item.UnitPrice, item.Remark)
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
func ShipSalesOrder(ctx context.Context, orderID int64, items []struct {
	MaterialID int64
	Quantity   float64
}, userID int64, username string) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		query := `CALL sp_ship_sales_order_item($1, $2, $3, $4)`
		_, err := tx.Exec(ctx, query, orderID, item.MaterialID, item.Quantity, userID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				return fmt.Errorf("DB_ERROR (MaterialID %d): %s", item.MaterialID, pgErr.Message)
			}
			return err
		}
	}

	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, _ = tx.Exec(ctx, auditQuery, userID, username, "SHIP", "SALES", "sales_order", orderID, nil)

	return tx.Commit(ctx)
}
