/**
 * 功能：report.go
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

// ReportStockInSummary 入库汇总报表
func ReportStockInSummary(ctx context.Context, month string) ([]response.StockInSummaryReport, error) {
	query := `
		SELECT report_month, supplier_name, material_code, material_name, unit, total_quantity, order_count
		FROM v_report_stock_in_summary
	`
	var (
		args  []interface{}
		where []string
	)
	if month != "" {
		where = append(where, "report_month = $1")
		args = append(args, month)
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY report_month DESC, supplier_name ASC"

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []response.StockInSummaryReport
	for rows.Next() {
		var item response.StockInSummaryReport
		if err := rows.Scan(&item.ReportMonth, &item.SupplierName, &item.MaterialCode, &item.MaterialName,
			&item.Unit, &item.TotalQuantity, &item.OrderCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

// ReportStockOutSummary 出库汇总报表
func ReportStockOutSummary(ctx context.Context, month string) ([]response.StockOutSummaryReport, error) {
	query := `
		SELECT report_month, out_type, material_code, material_name, unit, total_quantity, order_count
		FROM v_report_stock_out_summary
	`
	var (
		args  []interface{}
		where []string
	)
	if month != "" {
		where = append(where, "report_month = $1")
		args = append(args, month)
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY report_month DESC, out_type ASC"

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []response.StockOutSummaryReport
	for rows.Next() {
		var item response.StockOutSummaryReport
		if err := rows.Scan(&item.ReportMonth, &item.OutType, &item.MaterialCode, &item.MaterialName,
			&item.Unit, &item.TotalQuantity, &item.OrderCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

// ReportInventoryStatus 实时库存分布报表
func ReportInventoryStatus(ctx context.Context) ([]response.InventoryStatusReport, error) {
	query := `
		SELECT warehouse_name, category_name, material_code, material_name, unit, 
		       current_quantity, locked_quantity, available_quantity
		FROM v_report_inventory_status
		ORDER BY warehouse_name, category_name
	`
	rows, err := database.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []response.InventoryStatusReport
	for rows.Next() {
		var item response.InventoryStatusReport
		if err := rows.Scan(&item.WarehouseName, &item.CategoryName, &item.MaterialCode, &item.MaterialName,
			&item.Unit, &item.CurrentQuantity, &item.LockedQuantity, &item.AvailableQuantity); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func ReportCustomerReconciliationSummary(ctx context.Context, q *request.ReconciliationSummaryQuery) ([]response.CustomerReconciliationSummaryReport, int64, error) {
	whereSrc := []string{"1=1"}
	wherePay := []string{"deleted_at IS NULL", "status = 'completed'"}
	args := make([]interface{}, 0)
	argID := 1

	if q.StartDate != "" {
		whereSrc = append(whereSrc, fmt.Sprintf("order_date::date >= $%d::date", argID))
		wherePay = append(wherePay, fmt.Sprintf("statement_date::date >= $%d::date", argID))
		args = append(args, q.StartDate)
		argID++
	}
	if q.EndDate != "" {
		whereSrc = append(whereSrc, fmt.Sprintf("order_date::date <= $%d::date", argID))
		wherePay = append(wherePay, fmt.Sprintf("statement_date::date <= $%d::date", argID))
		args = append(args, q.EndDate)
		argID++
	}
	if q.CustomerID != 0 {
		whereSrc = append(whereSrc, fmt.Sprintf("customer_id = $%d", argID))
		wherePay = append(wherePay, fmt.Sprintf("customer_id = $%d", argID))
		args = append(args, q.CustomerID)
		argID++
	}
	if strings.TrimSpace(q.Keyword) != "" {
		whereSrc = append(whereSrc, fmt.Sprintf("(customer_name ILIKE $%d OR customer_code ILIKE $%d)", argID, argID))
		args = append(args, "%"+strings.TrimSpace(q.Keyword)+"%")
		argID++
	}

	countSQL := fmt.Sprintf(`
		WITH src AS (
			SELECT
				customer_id,
				COALESCE(customer_code, '') AS customer_code,
				COALESCE(customer_name, '') AS customer_name,
				COALESCE(SUM(order_amount), 0)::float8 AS receivable_amount,
				COALESCE(SUM(verified_amount), 0)::float8 AS verified_amount,
				COALESCE(SUM(unverified_amount), 0)::float8 AS balance_amount
			FROM v_fund_collection_source
			WHERE %s
			GROUP BY customer_id, customer_code, customer_name
		)
		SELECT COUNT(*) FROM src
	`, strings.Join(whereSrc, " AND "))

	var total int64
	if err := database.Pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	querySQL := fmt.Sprintf(`
		WITH src AS (
			SELECT
				customer_id,
				COALESCE(customer_code, '') AS customer_code,
				COALESCE(customer_name, '') AS customer_name,
				COALESCE(SUM(order_amount), 0)::float8 AS receivable_amount,
				COALESCE(SUM(verified_amount), 0)::float8 AS verified_amount,
				COALESCE(SUM(unverified_amount), 0)::float8 AS balance_amount
			FROM v_fund_collection_source
			WHERE %s
			GROUP BY customer_id, customer_code, customer_name
		),
		pay AS (
			SELECT
				customer_id,
				COALESCE(SUM(actual_amount), 0)::float8 AS actual_amount,
				COALESCE(SUM(invoice_amount), 0)::float8 AS invoice_amount,
				COALESCE(SUM(discount_amount), 0)::float8 AS discount_amount
			FROM fund_collection
			WHERE %s
			GROUP BY customer_id
		)
		SELECT
			src.customer_id,
			src.customer_code,
			src.customer_name,
			src.receivable_amount,
			src.verified_amount,
			src.balance_amount,
			COALESCE(pay.actual_amount, 0)::float8 AS actual_amount,
			COALESCE(pay.invoice_amount, 0)::float8 AS invoice_amount,
			COALESCE(pay.discount_amount, 0)::float8 AS discount_amount
		FROM src
		LEFT JOIN pay ON pay.customer_id = src.customer_id
		ORDER BY ABS(src.balance_amount) DESC, src.customer_code ASC, src.customer_id DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(whereSrc, " AND "), strings.Join(wherePay, " AND "), argID, argID+1)

	args = append(args, q.PageSize, q.Offset())
	rows, err := database.Pool.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]response.CustomerReconciliationSummaryReport, 0)
	for rows.Next() {
		var it response.CustomerReconciliationSummaryReport
		if err := rows.Scan(
			&it.CustomerID, &it.CustomerCode, &it.CustomerName,
			&it.ReceivableAmount, &it.VerifiedAmount, &it.BalanceAmount,
			&it.ActualAmount, &it.InvoiceAmount, &it.DiscountAmount,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func ReportSupplierReconciliationSummary(ctx context.Context, q *request.ReconciliationSummaryQuery) ([]response.SupplierReconciliationSummaryReport, int64, error) {
	whereSrc := []string{"1=1"}
	wherePay := []string{"deleted_at IS NULL", "status = 'completed'"}
	args := make([]interface{}, 0)
	argID := 1

	if q.StartDate != "" {
		whereSrc = append(whereSrc, fmt.Sprintf("order_date::date >= $%d::date", argID))
		wherePay = append(wherePay, fmt.Sprintf("statement_date::date >= $%d::date", argID))
		args = append(args, q.StartDate)
		argID++
	}
	if q.EndDate != "" {
		whereSrc = append(whereSrc, fmt.Sprintf("order_date::date <= $%d::date", argID))
		wherePay = append(wherePay, fmt.Sprintf("statement_date::date <= $%d::date", argID))
		args = append(args, q.EndDate)
		argID++
	}
	if q.SupplierID != 0 {
		whereSrc = append(whereSrc, fmt.Sprintf("supplier_id = $%d", argID))
		wherePay = append(wherePay, fmt.Sprintf("supplier_id = $%d", argID))
		args = append(args, q.SupplierID)
		argID++
	}
	if strings.TrimSpace(q.Keyword) != "" {
		whereSrc = append(whereSrc, fmt.Sprintf("(supplier_name ILIKE $%d OR supplier_code ILIKE $%d)", argID, argID))
		args = append(args, "%"+strings.TrimSpace(q.Keyword)+"%")
		argID++
	}

	countSQL := fmt.Sprintf(`
		WITH src AS (
			SELECT
				supplier_id,
				COALESCE(supplier_code, '') AS supplier_code,
				COALESCE(supplier_name, '') AS supplier_name,
				COALESCE(SUM(order_amount), 0)::float8 AS payable_amount,
				COALESCE(SUM(verified_amount), 0)::float8 AS verified_amount,
				COALESCE(SUM(unverified_amount), 0)::float8 AS balance_amount
			FROM v_fund_payment_source
			WHERE %s
			GROUP BY supplier_id, supplier_code, supplier_name
		)
		SELECT COUNT(*) FROM src
	`, strings.Join(whereSrc, " AND "))

	var total int64
	if err := database.Pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	querySQL := fmt.Sprintf(`
		WITH src AS (
			SELECT
				supplier_id,
				COALESCE(supplier_code, '') AS supplier_code,
				COALESCE(supplier_name, '') AS supplier_name,
				COALESCE(SUM(order_amount), 0)::float8 AS payable_amount,
				COALESCE(SUM(verified_amount), 0)::float8 AS verified_amount,
				COALESCE(SUM(unverified_amount), 0)::float8 AS balance_amount
			FROM v_fund_payment_source
			WHERE %s
			GROUP BY supplier_id, supplier_code, supplier_name
		),
		pay AS (
			SELECT
				supplier_id,
				COALESCE(SUM(actual_amount), 0)::float8 AS actual_amount,
				COALESCE(SUM(invoice_amount), 0)::float8 AS invoice_amount,
				COALESCE(SUM(discount_amount), 0)::float8 AS discount_amount
			FROM fund_payment
			WHERE %s
			GROUP BY supplier_id
		)
		SELECT
			src.supplier_id,
			src.supplier_code,
			src.supplier_name,
			src.payable_amount,
			src.verified_amount,
			src.balance_amount,
			COALESCE(pay.actual_amount, 0)::float8 AS actual_amount,
			COALESCE(pay.invoice_amount, 0)::float8 AS invoice_amount,
			COALESCE(pay.discount_amount, 0)::float8 AS discount_amount
		FROM src
		LEFT JOIN pay ON pay.supplier_id = src.supplier_id
		ORDER BY ABS(src.balance_amount) DESC, src.supplier_code ASC, src.supplier_id DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(whereSrc, " AND "), strings.Join(wherePay, " AND "), argID, argID+1)

	args = append(args, q.PageSize, q.Offset())
	rows, err := database.Pool.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]response.SupplierReconciliationSummaryReport, 0)
	for rows.Next() {
		var it response.SupplierReconciliationSummaryReport
		if err := rows.Scan(
			&it.SupplierID, &it.SupplierCode, &it.SupplierName,
			&it.PayableAmount, &it.VerifiedAmount, &it.BalanceAmount,
			&it.ActualAmount, &it.InvoiceAmount, &it.DiscountAmount,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func ReportProfit(ctx context.Context, q *request.ProfitReportQuery) (*response.ProfitReport, error) {
	whereSales := []string{"deleted_at IS NULL", "order_status NOT IN ('draft', 'cancelled')"}
	whereCost := []string{"so.deleted_at IS NULL", "so.out_type = 'sales'", "so.status = 'confirmed'"}
	args := make([]interface{}, 0)
	argID := 1

	if q.StartDate != "" {
		whereSales = append(whereSales, fmt.Sprintf("order_date::date >= $%d::date", argID))
		whereCost = append(whereCost, fmt.Sprintf("so.stock_out_date::date >= $%d::date", argID))
		args = append(args, q.StartDate)
		argID++
	}
	if q.EndDate != "" {
		whereSales = append(whereSales, fmt.Sprintf("order_date::date <= $%d::date", argID))
		whereCost = append(whereCost, fmt.Sprintf("so.stock_out_date::date <= $%d::date", argID))
		args = append(args, q.EndDate)
		argID++
	}

	sql := fmt.Sprintf(`
		WITH sales AS (
			SELECT COALESCE(SUM(COALESCE(total_amount, 0)), 0)::float8 AS sales_amount
			FROM sales_order
			WHERE %s
		),
		cost AS (
			SELECT COALESCE(SUM(soi.quantity * COALESCE(inv.unit_cost, 0)), 0)::float8 AS cost_amount
			FROM stock_out so
			INNER JOIN stock_out_item soi ON soi.stock_out_id = so.id
			LEFT JOIN inventory inv ON inv.id = soi.inventory_id
			WHERE %s
		)
		SELECT sales.sales_amount, cost.cost_amount, (sales.sales_amount - cost.cost_amount) AS profit
		FROM sales, cost
	`, strings.Join(whereSales, " AND "), strings.Join(whereCost, " AND "))

	var out response.ProfitReport
	if err := database.Pool.QueryRow(ctx, sql, args...).Scan(&out.SalesAmount, &out.CostAmount, &out.Profit); err != nil {
		return nil, err
	}
	return &out, nil
}
