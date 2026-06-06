/**
 * 功能：supplier.go
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

// ListSuppliers 分页查询供应商
func ListSuppliers(ctx context.Context, q *request.SupplierQuery) ([]response.SupplierResp, int64, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if q.SupplierName != "" {
		where = append(where, fmt.Sprintf("supplier_name ILIKE $%d", argID))
		args = append(args, "%"+q.SupplierName+"%")
		argID++
	}
	if q.SupplierCode != "" {
		where = append(where, fmt.Sprintf("supplier_code ILIKE $%d", argID))
		args = append(args, "%"+q.SupplierCode+"%")
		argID++
	}
	if q.SupplierType != "" {
		where = append(where, fmt.Sprintf("supplier_type = $%d", argID))
		args = append(args, q.SupplierType)
		argID++
	}
	if q.IsQualified != nil {
		where = append(where, fmt.Sprintf("is_qualified = $%d", argID))
		args = append(args, *q.IsQualified)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_supplier_list WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, supplier_code, supplier_name, credit_code, supplier_type, supplier_type_name,
		       contact_person, contact_phone, address, is_qualified, qualification_expire,
		       supplier_rating, supplier_rating_name, bank_name, bank_account, remark,
		       status, status_name, created_at, updated_at
		FROM v_supplier_list
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

	var result []response.SupplierResp
	for rows.Next() {
		var item response.SupplierResp
		if err := rows.Scan(&item.ID, &item.SupplierCode, &item.SupplierName, &item.CreditCode, &item.SupplierType, &item.SupplierTypeName,
			&item.ContactPerson, &item.ContactPhone, &item.Address, &item.IsQualified, &item.QualificationExpire,
			&item.SupplierRating, &item.SupplierRatingName, &item.BankName, &item.BankAccount, &item.Remark,
			&item.Status, &item.StatusName, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// CreateSupplier 创建供应商
func CreateSupplier(ctx context.Context, req *request.CreateSupplierReq, userID int64, username string) (int64, string, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, "", err
	}
	defer tx.Rollback(ctx)

	// 自动生成供应商编码
	var supplierCode string
	err = tx.QueryRow(ctx, "SELECT fn_generate_base_code('S')").Scan(&supplierCode)
	if err != nil {
		return 0, "", fmt.Errorf("生成供应商编码失败: %w", err)
	}

	var id int64
	query := `
		INSERT INTO supplier (supplier_code, supplier_name, credit_code, supplier_type, contact_person, 
		                      contact_phone, address, is_qualified, qualification_expire, bank_name, 
		                      bank_account, remark, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, '')::DATE, $10, $11, $12, $13)
		RETURNING id
	`
	err = tx.QueryRow(ctx, query,
		supplierCode, req.SupplierName, req.CreditCode, req.SupplierType, req.ContactPerson,
		req.ContactPhone, req.Address, req.IsQualified, req.QualificationExpire, req.BankName,
		req.BankAccount, req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, "", err
	}

	// 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "CREATE", "SUPPLIER", "supplier", id, nil)
	if err != nil {
		return 0, "", err
	}

	return id, supplierCode, tx.Commit(ctx)
}

// UpdateSupplier 更新供应商
func UpdateSupplier(ctx context.Context, id int64, req *request.UpdateSupplierReq, userID int64, username string) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 禁用验证：检查是否有关联的进行中采购订单
	if req.Status == "disabled" {
		var refCount int64
		err = tx.QueryRow(ctx, `SELECT count(*) FROM purchase_order WHERE supplier_id = $1 AND status NOT IN ('closed', 'cancelled') AND deleted_at IS NULL`, id).Scan(&refCount)
		if err != nil {
			return err
		}
		if refCount > 0 {
			return fmt.Errorf("该供应商有 %d 个进行中的采购订单，无法禁用", refCount)
		}
	}

	query := `
		UPDATE supplier
		SET supplier_name = $1, credit_code = $2, supplier_type = $3, contact_person = $4, 
		    contact_phone = $5, address = $6, is_qualified = $7, qualification_expire = NULLIF($8, '')::DATE, 
		    bank_name = $9, bank_account = $10, remark = $11, status = $12, 
		    updated_by = $13, updated_at = NOW()
		WHERE id = $14 AND deleted_at IS NULL
	`
	res, err := tx.Exec(ctx, query,
		req.SupplierName, req.CreditCode, req.SupplierType, req.ContactPerson,
		req.ContactPhone, req.Address, req.IsQualified, req.QualificationExpire,
		req.BankName, req.BankAccount, req.Remark, req.Status, userID, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("supplier not found")
	}

	// 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "UPDATE", "SUPPLIER", "supplier", id, nil)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// DeleteSupplier 软删除供应商
func DeleteSupplier(ctx context.Context, id int64, userID int64, username string) error {
	// 删除验证：检查是否有采购订单关联
	var refCount int64
	err := database.Pool.QueryRow(ctx, `SELECT count(*) FROM purchase_order WHERE supplier_id = $1 AND deleted_at IS NULL`, id).Scan(&refCount)
	if err != nil {
		return err
	}
	if refCount > 0 {
		return fmt.Errorf("该供应商有 %d 个采购订单，无法删除", refCount)
	}

	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `UPDATE supplier SET deleted_at = NOW(), updated_by = $1 WHERE id = $2 AND deleted_at IS NULL`
	res, err := tx.Exec(ctx, query, userID, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("supplier not found")
	}

	// 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "DELETE", "SUPPLIER", "supplier", id, nil)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
