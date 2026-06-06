/**
 * 功能：report.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
)

func ReportStockInSummary(ctx context.Context, month string) ([]response.StockInSummaryReport, error) {
	return db.ReportStockInSummary(ctx, month)
}

func ReportStockOutSummary(ctx context.Context, month string) ([]response.StockOutSummaryReport, error) {
	return db.ReportStockOutSummary(ctx, month)
}

func ReportInventoryStatus(ctx context.Context) ([]response.InventoryStatusReport, error) {
	return db.ReportInventoryStatus(ctx)
}

func ReportCustomerReconciliationSummary(ctx context.Context, q *request.ReconciliationSummaryQuery) ([]response.CustomerReconciliationSummaryReport, int64, error) {
	return db.ReportCustomerReconciliationSummary(ctx, q)
}

func ReportSupplierReconciliationSummary(ctx context.Context, q *request.ReconciliationSummaryQuery) ([]response.SupplierReconciliationSummaryReport, int64, error) {
	return db.ReportSupplierReconciliationSummary(ctx, q)
}

func ReportProfit(ctx context.Context, q *request.ProfitReportQuery) (*response.ProfitReport, error) {
	return db.ReportProfit(ctx, q)
}
