/**
 * 功能：report.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package db

import (
	"context"
	"strings"

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
		args []interface{}
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
		args []interface{}
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
