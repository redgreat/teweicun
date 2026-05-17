package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

// ListFundPayments 分页查询付款单
func ListFundPayments(ctx context.Context, q *request.FundPaymentQuery) ([]response.FundPaymentResp, int64, error) {
	where := []string{"deleted_at IS NULL"}
	var args []interface{}
	argID := 1

	if q.StatementNo != "" {
		where = append(where, fmt.Sprintf("statement_no ILIKE $%d", argID))
		args = append(args, "%"+q.StatementNo+"%")
		argID++
	}
	if q.SupplierID != 0 {
		where = append(where, fmt.Sprintf("supplier_id = $%d", argID))
		args = append(args, q.SupplierID)
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_fund_payment_list WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, statement_no, supplier_id, supplier_name, supplier_code,
		       payer_id, payer_name, TO_CHAR(statement_date, 'YYYY-MM-DD'),
		       payment_amount, discount_amount, advance_amount,
		       settlement_method, settlement_account, settlement_no,
		       remark, status, created_at, updated_at
		FROM v_fund_payment_list
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

	var result []response.FundPaymentResp
	for rows.Next() {
		var item response.FundPaymentResp
		if err := rows.Scan(
			&item.ID, &item.StatementNo, &item.SupplierID, &item.SupplierName, &item.SupplierCode,
			&item.PayerID, &item.PayerName, &item.StatementDate,
			&item.PaymentAmount, &item.DiscountAmount, &item.AdvanceAmount,
			&item.SettlementMethod, &item.SettlementAccount, &item.SettlementNo,
			&item.Remark, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// GetFundPayment 获取付款单详情
func GetFundPayment(ctx context.Context, id int64) (*response.FundPaymentDetailResp, error) {
	query := `
		SELECT id, statement_no, supplier_id, supplier_name, supplier_code,
		       payer_id, payer_name, TO_CHAR(statement_date, 'YYYY-MM-DD'),
		       payment_amount, discount_amount, advance_amount,
		       settlement_method, settlement_account, settlement_no,
		       remark, status, created_at, updated_at
		FROM v_fund_payment_list
		WHERE id = $1
	`
	var order response.FundPaymentDetailResp
	err := database.Pool.QueryRow(ctx, query, id).Scan(
		&order.ID, &order.StatementNo, &order.SupplierID, &order.SupplierName, &order.SupplierCode,
		&order.PayerID, &order.PayerName, &order.StatementDate,
		&order.PaymentAmount, &order.DiscountAmount, &order.AdvanceAmount,
		&order.SettlementMethod, &order.SettlementAccount, &order.SettlementNo,
		&order.Remark, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	itemsQuery := `
		SELECT id, statement_id, source_order_id, source_order_no, business_type,
		       TO_CHAR(order_date, 'YYYY-MM-DD'), order_amount, verified_amount,
		       unverified_amount, current_verify_amount, custom_tax_amount, remark, created_at
		FROM fund_payment_item
		WHERE statement_id = $1
	`
	rows, err := database.Pool.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item response.FundPaymentItemResp
		if err := rows.Scan(
			&item.ID, &item.StatementID, &item.SourceOrderID, &item.SourceOrderNo, &item.BusinessType,
			&item.OrderDate, &item.OrderAmount, &item.VerifiedAmount,
			&item.UnverifiedAmount, &item.CurrentVerifyAmount, &item.CustomTaxAmount, &item.Remark, &item.CreatedAt); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

// CreateFundPayment 创建付款单
func CreateFundPayment(ctx context.Context, req *request.CreateFundPaymentReq, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var statementNo string
	err = tx.QueryRow(ctx, "SELECT fn_generate_bill_no('FK')").Scan(&statementNo)
	if err != nil {
		return 0, fmt.Errorf("生成单号失败: %w", err)
	}

	var id int64
	query := `
		INSERT INTO fund_payment (
			statement_no, supplier_id, payer_id, statement_date,
			payment_amount, discount_amount, advance_amount,
			settlement_method, settlement_account, settlement_no,
			remark, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`
	err = tx.QueryRow(ctx, query,
		statementNo, req.SupplierID, userID, req.StatementDate,
		req.PaymentAmount, req.DiscountAmount, req.AdvanceAmount,
		req.SettlementMethod, req.SettlementAccount, req.SettlementNo,
		req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
	}

	for _, item := range req.Items {
		itemQuery := `
			INSERT INTO fund_payment_item (
				statement_id, source_order_id, source_order_no, business_type,
				order_date, order_amount, verified_amount,
				unverified_amount, current_verify_amount, custom_tax_amount, remark
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		_, err = tx.Exec(ctx, itemQuery,
			id, item.SourceOrderID, item.SourceOrderNo, item.BusinessType,
			item.OrderDate, item.OrderAmount, item.VerifiedAmount,
			item.UnverifiedAmount, item.CurrentVerifyAmount, item.CustomTaxAmount, item.Remark)
		if err != nil {
			return 0, err
		}
		
		// Update purchase order verified amount
		if item.BusinessType == "采购入库" {
			updatePo := `UPDATE purchase_order SET verified_amount = COALESCE(verified_amount, 0) + $1 WHERE id = $2`
			_, err = tx.Exec(ctx, updatePo, item.CurrentVerifyAmount, item.SourceOrderID)
			if err != nil {
				return 0, err
			}
		}
	}

	// 审批付款单
	approveQuery := `UPDATE fund_payment SET status = 'completed' WHERE id = $1`
	_, err = tx.Exec(ctx, approveQuery, id)
	if err != nil {
	    return 0, err
	}

	return id, tx.Commit(ctx)
}

// ListFundCollections 分页查询收款单
func ListFundCollections(ctx context.Context, q *request.FundCollectionQuery) ([]response.FundCollectionResp, int64, error) {
	where := []string{"deleted_at IS NULL"}
	var args []interface{}
	argID := 1

	if q.StatementNo != "" {
		where = append(where, fmt.Sprintf("statement_no ILIKE $%d", argID))
		args = append(args, "%"+q.StatementNo+"%")
		argID++
	}
	if q.CustomerID != 0 {
		where = append(where, fmt.Sprintf("customer_id = $%d", argID))
		args = append(args, q.CustomerID)
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_fund_collection_list WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, statement_no, customer_id, customer_name, customer_code,
		       payee_id, payee_name, TO_CHAR(statement_date, 'YYYY-MM-DD'),
		       collection_amount, discount_amount, advance_amount,
		       settlement_method, settlement_account, settlement_no,
		       remark, status, created_at, updated_at
		FROM v_fund_collection_list
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

	var result []response.FundCollectionResp
	for rows.Next() {
		var item response.FundCollectionResp
		if err := rows.Scan(
			&item.ID, &item.StatementNo, &item.CustomerID, &item.CustomerName, &item.CustomerCode,
			&item.PayeeID, &item.PayeeName, &item.StatementDate,
			&item.CollectionAmount, &item.DiscountAmount, &item.AdvanceAmount,
			&item.SettlementMethod, &item.SettlementAccount, &item.SettlementNo,
			&item.Remark, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// GetFundCollection 获取收款单详情
func GetFundCollection(ctx context.Context, id int64) (*response.FundCollectionDetailResp, error) {
	query := `
		SELECT id, statement_no, customer_id, customer_name, customer_code,
		       payee_id, payee_name, TO_CHAR(statement_date, 'YYYY-MM-DD'),
		       collection_amount, discount_amount, advance_amount,
		       settlement_method, settlement_account, settlement_no,
		       remark, status, created_at, updated_at
		FROM v_fund_collection_list
		WHERE id = $1
	`
	var order response.FundCollectionDetailResp
	err := database.Pool.QueryRow(ctx, query, id).Scan(
		&order.ID, &order.StatementNo, &order.CustomerID, &order.CustomerName, &order.CustomerCode,
		&order.PayeeID, &order.PayeeName, &order.StatementDate,
		&order.CollectionAmount, &order.DiscountAmount, &order.AdvanceAmount,
		&order.SettlementMethod, &order.SettlementAccount, &order.SettlementNo,
		&order.Remark, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	itemsQuery := `
		SELECT id, statement_id, source_order_id, source_order_no, business_type,
		       TO_CHAR(order_date, 'YYYY-MM-DD'), order_amount, verified_amount,
		       unverified_amount, current_verify_amount, custom_tax_amount, remark, created_at
		FROM fund_collection_item
		WHERE statement_id = $1
	`
	rows, err := database.Pool.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item response.FundCollectionItemResp
		if err := rows.Scan(
			&item.ID, &item.StatementID, &item.SourceOrderID, &item.SourceOrderNo, &item.BusinessType,
			&item.OrderDate, &item.OrderAmount, &item.VerifiedAmount,
			&item.UnverifiedAmount, &item.CurrentVerifyAmount, &item.CustomTaxAmount, &item.Remark, &item.CreatedAt); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

// CreateFundCollection 创建收款单
func CreateFundCollection(ctx context.Context, req *request.CreateFundCollectionReq, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var statementNo string
	err = tx.QueryRow(ctx, "SELECT fn_generate_bill_no('SK')").Scan(&statementNo)
	if err != nil {
		return 0, fmt.Errorf("生成单号失败: %w", err)
	}

	var id int64
	query := `
		INSERT INTO fund_collection (
			statement_no, customer_id, payee_id, statement_date,
			collection_amount, discount_amount, advance_amount,
			settlement_method, settlement_account, settlement_no,
			remark, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`
	err = tx.QueryRow(ctx, query,
		statementNo, req.CustomerID, userID, req.StatementDate,
		req.CollectionAmount, req.DiscountAmount, req.AdvanceAmount,
		req.SettlementMethod, req.SettlementAccount, req.SettlementNo,
		req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
	}

	for _, item := range req.Items {
		itemQuery := `
			INSERT INTO fund_collection_item (
				statement_id, source_order_id, source_order_no, business_type,
				order_date, order_amount, verified_amount,
				unverified_amount, current_verify_amount, custom_tax_amount, remark
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		_, err = tx.Exec(ctx, itemQuery,
			id, item.SourceOrderID, item.SourceOrderNo, item.BusinessType,
			item.OrderDate, item.OrderAmount, item.VerifiedAmount,
			item.UnverifiedAmount, item.CurrentVerifyAmount, item.CustomTaxAmount, item.Remark)
		if err != nil {
			return 0, err
		}
		
		// Update sales order verified amount
		if item.BusinessType == "销售出库" {
			updateSo := `UPDATE sales_order SET verified_amount = COALESCE(verified_amount, 0) + $1 WHERE id = $2`
			_, err = tx.Exec(ctx, updateSo, item.CurrentVerifyAmount, item.SourceOrderID)
			if err != nil {
				return 0, err
			}
		}
	}

	approveQuery := `UPDATE fund_collection SET status = 'completed' WHERE id = $1`
	_, err = tx.Exec(ctx, approveQuery, id)
	if err != nil {
	    return 0, err
	}

	return id, tx.Commit(ctx)
}
