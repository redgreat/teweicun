/**
 * 功能：customer.go
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

// ListCustomers 分页查询客户
func ListCustomers(ctx context.Context, q *request.CustomerQuery) ([]response.CustomerResp, int64, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if q.CustomerName != "" {
		where = append(where, fmt.Sprintf("customer_name ILIKE $%d", argID))
		args = append(args, "%"+q.CustomerName+"%")
		argID++
	}
	if q.CustomerCode != "" {
		where = append(where, fmt.Sprintf("customer_code ILIKE $%d", argID))
		args = append(args, "%"+q.CustomerCode+"%")
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_customer_list WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, customer_code, customer_name, credit_code, customer_type, customer_type_name,
		       contact_person, contact_phone, address, sales_person_name, remark,
		       status, status_name, created_at, updated_at
		FROM v_customer_list
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

	var result []response.CustomerResp
	for rows.Next() {
		var item response.CustomerResp
		if err := rows.Scan(&item.ID, &item.CustomerCode, &item.CustomerName, &item.CreditCode,
			&item.CustomerType, &item.CustomerTypeName, &item.ContactPerson, &item.ContactPhone,
			&item.Address, &item.SalesPersonName, &item.Remark,
			&item.Status, &item.StatusName, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// CreateCustomer 创建客户
func CreateCustomer(ctx context.Context, req *request.CreateCustomerReq, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// 自动生成客户编码
	var customerCode string
	err = tx.QueryRow(ctx, "SELECT fn_generate_base_code('C')").Scan(&customerCode)
	if err != nil {
		return 0, fmt.Errorf("生成客户编码失败: %w", err)
	}

	var id int64
	query := `
		INSERT INTO customer (customer_code, customer_name, credit_code, contact_person, contact_phone, 
		                     address, remark, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	err = tx.QueryRow(ctx, query, customerCode, req.CustomerName, req.CreditCode,
		req.ContactPerson, req.ContactPhone, req.Address, req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
	}

	// 审计日志调用
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "CREATE", "CUSTOMER", "customer", id, nil)
	if err != nil {
		return 0, err
	}

	return id, tx.Commit(ctx)
}

// UpdateCustomer 更新客户
func UpdateCustomer(ctx context.Context, id int64, req *request.UpdateCustomerReq, userID int64, username string) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 禁用验证：检查是否有关联的进行中销售订单
	if req.Status == "disabled" {
		var refCount int64
		err = tx.QueryRow(ctx, `SELECT count(*) FROM sales_order WHERE customer_id = $1 AND status NOT IN ('closed', 'cancelled') AND deleted_at IS NULL`, id).Scan(&refCount)
		if err != nil {
			return err
		}
		if refCount > 0 {
			return fmt.Errorf("该客户有 %d 个进行中的销售订单，无法禁用", refCount)
		}
	}

	query := `
		UPDATE customer
		SET customer_name = $1, credit_code = $2, contact_person = $3, contact_phone = $4, 
		    address = $5, remark = $6, status = $7, updated_by = $8, updated_at = NOW()
		WHERE id = $9 AND deleted_at IS NULL
	`
	res, err := tx.Exec(ctx, query, req.CustomerName, req.CreditCode, req.ContactPerson,
		req.ContactPhone, req.Address, req.Remark, req.Status, userID, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("customer not found")
	}

	// 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "UPDATE", "CUSTOMER", "customer", id, nil)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// DeleteCustomer 软删除客户
func DeleteCustomer(ctx context.Context, id int64, userID int64, username string) error {
	// 删除验证：检查是否有销售订单关联
	var refCount int64
	err := database.Pool.QueryRow(ctx, `SELECT count(*) FROM sales_order WHERE customer_id = $1 AND deleted_at IS NULL`, id).Scan(&refCount)
	if err != nil {
		return err
	}
	if refCount > 0 {
		return fmt.Errorf("该客户有 %d 个销售订单，无法删除", refCount)
	}

	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `UPDATE customer SET deleted_at = NOW(), updated_by = $1 WHERE id = $2 AND deleted_at IS NULL`
	res, err := tx.Exec(ctx, query, userID, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("customer not found")
	}

	// 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "DELETE", "CUSTOMER", "customer", id, nil)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
