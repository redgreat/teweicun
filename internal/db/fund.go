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
		       bill_type, payment_amount, invoice_amount, actual_amount, difference_amount,
		       discount_amount, advance_amount,
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
			&item.BillType, &item.PaymentAmount, &item.InvoiceAmount, &item.ActualAmount,
			&item.DifferenceAmount, &item.DiscountAmount, &item.AdvanceAmount,
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
		       bill_type, payment_amount, invoice_amount, actual_amount, difference_amount,
		       discount_amount, advance_amount,
		       settlement_method, settlement_account, settlement_no,
		       remark, status, created_at, updated_at
		FROM v_fund_payment_list
		WHERE id = $1
	`
	var order response.FundPaymentDetailResp
	err := database.Pool.QueryRow(ctx, query, id).Scan(
		&order.ID, &order.StatementNo, &order.SupplierID, &order.SupplierName, &order.SupplierCode,
		&order.PayerID, &order.PayerName, &order.StatementDate,
		&order.BillType, &order.PaymentAmount, &order.InvoiceAmount, &order.ActualAmount,
		&order.DifferenceAmount, &order.DiscountAmount, &order.AdvanceAmount,
		&order.SettlementMethod, &order.SettlementAccount, &order.SettlementNo,
		&order.Remark, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	itemsQuery := `
		SELECT id, statement_id, source_doc_type, source_order_id, source_order_no, business_type,
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
			&item.ID, &item.StatementID, &item.SourceDocType, &item.SourceOrderID, &item.SourceOrderNo, &item.BusinessType,
			&item.OrderDate, &item.OrderAmount, &item.VerifiedAmount,
			&item.UnverifiedAmount, &item.CurrentVerifyAmount, &item.CustomTaxAmount, &item.Remark, &item.CreatedAt); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

// ListFundPaymentSources 查询供应商可付款/抵充来源单据
func ListFundPaymentSources(ctx context.Context, q *request.FundPaymentSourceQuery) ([]response.FundPaymentSourceResp, error) {
	where := []string{"supplier_id = $1", "ABS(unverified_amount) >= 0.005"}
	args := []interface{}{q.SupplierID}
	argID := 2

	if strings.TrimSpace(q.Keyword) != "" {
		where = append(where, fmt.Sprintf("(source_order_no ILIKE $%d OR business_type ILIKE $%d)", argID, argID))
		args = append(args, "%"+strings.TrimSpace(q.Keyword)+"%")
		argID++
	}

	query := fmt.Sprintf(`
		SELECT source_doc_type, source_order_id, source_order_no, business_type,
		       TO_CHAR(order_date, 'YYYY-MM-DD'), supplier_id, supplier_code, supplier_name,
		       order_amount, verified_amount, unverified_amount, verify_status
		FROM v_fund_payment_source
		WHERE %s
		ORDER BY order_date DESC, source_order_id DESC
		LIMIT 200
	`, strings.Join(where, " AND "))

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []response.FundPaymentSourceResp
	for rows.Next() {
		var item response.FundPaymentSourceResp
		if err := rows.Scan(
			&item.SourceDocType, &item.SourceOrderID, &item.SourceOrderNo, &item.BusinessType,
			&item.OrderDate, &item.SupplierID, &item.SupplierCode, &item.SupplierName,
			&item.OrderAmount, &item.VerifiedAmount, &item.UnverifiedAmount, &item.VerifyStatus,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, rows.Err()
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
			bill_type, payment_amount, invoice_amount, actual_amount,
			discount_amount, advance_amount,
			settlement_method, settlement_account, settlement_no,
			remark, created_by
		) VALUES ($1, $2, $3, $4, COALESCE(NULLIF($5, ''), 'cash'), $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`
	err = tx.QueryRow(ctx, query,
		statementNo, req.SupplierID, userID, req.StatementDate,
		req.BillType, req.PaymentAmount, req.InvoiceAmount, req.ActualAmount,
		req.DiscountAmount, req.AdvanceAmount,
		req.SettlementMethod, req.SettlementAccount, req.SettlementNo,
		req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
	}

	for _, item := range req.Items {
		sourceDocType := normalizePaymentSourceDocType(item.SourceDocType, item.BusinessType)
		itemQuery := `
			INSERT INTO fund_payment_item (
				statement_id, source_doc_type, source_order_id, source_order_no, business_type,
				order_date, order_amount, verified_amount,
				unverified_amount, current_verify_amount, custom_tax_amount, remark
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`
		_, err = tx.Exec(ctx, itemQuery,
			id, sourceDocType, item.SourceOrderID, item.SourceOrderNo, item.BusinessType,
			item.OrderDate, item.OrderAmount, item.VerifiedAmount,
			item.UnverifiedAmount, item.CurrentVerifyAmount, item.CustomTaxAmount, item.Remark)
		if err != nil {
			return 0, err
		}

		if sourceDocType == "purchase_order" {
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
		       bill_type, collection_amount, invoice_amount, actual_amount, difference_amount,
		       discount_amount, advance_amount,
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
			&item.BillType, &item.CollectionAmount, &item.InvoiceAmount, &item.ActualAmount,
			&item.DifferenceAmount, &item.DiscountAmount, &item.AdvanceAmount,
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
		       bill_type, collection_amount, invoice_amount, actual_amount, difference_amount,
		       discount_amount, advance_amount,
		       settlement_method, settlement_account, settlement_no,
		       remark, status, created_at, updated_at
		FROM v_fund_collection_list
		WHERE id = $1
	`
	var order response.FundCollectionDetailResp
	err := database.Pool.QueryRow(ctx, query, id).Scan(
		&order.ID, &order.StatementNo, &order.CustomerID, &order.CustomerName, &order.CustomerCode,
		&order.PayeeID, &order.PayeeName, &order.StatementDate,
		&order.BillType, &order.CollectionAmount, &order.InvoiceAmount, &order.ActualAmount,
		&order.DifferenceAmount, &order.DiscountAmount, &order.AdvanceAmount,
		&order.SettlementMethod, &order.SettlementAccount, &order.SettlementNo,
		&order.Remark, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	itemsQuery := `
		SELECT id, statement_id, source_doc_type, source_order_id, source_order_no, business_type,
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
			&item.ID, &item.StatementID, &item.SourceDocType, &item.SourceOrderID, &item.SourceOrderNo, &item.BusinessType,
			&item.OrderDate, &item.OrderAmount, &item.VerifiedAmount,
			&item.UnverifiedAmount, &item.CurrentVerifyAmount, &item.CustomTaxAmount, &item.Remark, &item.CreatedAt); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

// ListFundCollectionSources 查询客户可收款/抵充来源单据
func ListFundCollectionSources(ctx context.Context, q *request.FundCollectionSourceQuery) ([]response.FundCollectionSourceResp, error) {
	where := []string{"customer_id = $1", "ABS(unverified_amount) >= 0.005"}
	args := []interface{}{q.CustomerID}
	argID := 2

	if strings.TrimSpace(q.Keyword) != "" {
		where = append(where, fmt.Sprintf("(source_order_no ILIKE $%d OR business_type ILIKE $%d)", argID, argID))
		args = append(args, "%"+strings.TrimSpace(q.Keyword)+"%")
		argID++
	}

	query := fmt.Sprintf(`
		SELECT source_doc_type, source_order_id, source_order_no, business_type,
		       TO_CHAR(order_date, 'YYYY-MM-DD'), customer_id, customer_code, customer_name,
		       order_amount, verified_amount, unverified_amount, verify_status
		FROM v_fund_collection_source
		WHERE %s
		ORDER BY order_date DESC, source_order_id DESC
		LIMIT 200
	`, strings.Join(where, " AND "))

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []response.FundCollectionSourceResp
	for rows.Next() {
		var item response.FundCollectionSourceResp
		if err := rows.Scan(
			&item.SourceDocType, &item.SourceOrderID, &item.SourceOrderNo, &item.BusinessType,
			&item.OrderDate, &item.CustomerID, &item.CustomerCode, &item.CustomerName,
			&item.OrderAmount, &item.VerifiedAmount, &item.UnverifiedAmount, &item.VerifyStatus,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, rows.Err()
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
			bill_type, collection_amount, invoice_amount, actual_amount,
			discount_amount, advance_amount,
			settlement_method, settlement_account, settlement_no,
			remark, created_by
		) VALUES ($1, $2, $3, $4, COALESCE(NULLIF($5, ''), 'cash'), $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`
	err = tx.QueryRow(ctx, query,
		statementNo, req.CustomerID, userID, req.StatementDate,
		req.BillType, req.CollectionAmount, req.InvoiceAmount, req.ActualAmount,
		req.DiscountAmount, req.AdvanceAmount,
		req.SettlementMethod, req.SettlementAccount, req.SettlementNo,
		req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
	}

	for _, item := range req.Items {
		sourceDocType := normalizeCollectionSourceDocType(item.SourceDocType, item.BusinessType)
		itemQuery := `
			INSERT INTO fund_collection_item (
				statement_id, source_doc_type, source_order_id, source_order_no, business_type,
				order_date, order_amount, verified_amount,
				unverified_amount, current_verify_amount, custom_tax_amount, remark
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`
		_, err = tx.Exec(ctx, itemQuery,
			id, sourceDocType, item.SourceOrderID, item.SourceOrderNo, item.BusinessType,
			item.OrderDate, item.OrderAmount, item.VerifiedAmount,
			item.UnverifiedAmount, item.CurrentVerifyAmount, item.CustomTaxAmount, item.Remark)
		if err != nil {
			return 0, err
		}

		if sourceDocType == "sales_order" {
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

func normalizePaymentSourceDocType(sourceDocType string, businessType string) string {
	sourceDocType = strings.TrimSpace(sourceDocType)
	if sourceDocType != "" {
		return sourceDocType
	}
	if strings.Contains(businessType, "退货") || strings.Contains(businessType, "抵充") {
		return "purchase_return"
	}
	return "purchase_order"
}

func normalizeCollectionSourceDocType(sourceDocType string, businessType string) string {
	sourceDocType = strings.TrimSpace(sourceDocType)
	if sourceDocType != "" {
		return sourceDocType
	}
	if strings.Contains(businessType, "退货") || strings.Contains(businessType, "抵充") {
		return "sales_return"
	}
	return "sales_order"
}
