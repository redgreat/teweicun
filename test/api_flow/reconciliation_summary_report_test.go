package api_flow

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redgreat/teweicun/test/testutil"
)

type customerRecSummary struct {
	CustomerID       int64   `json:"customer_id"`
	CustomerCode     string  `json:"customer_code"`
	CustomerName     string  `json:"customer_name"`
	ReceivableAmount float64 `json:"receivable_amount"`
	VerifiedAmount   float64 `json:"verified_amount"`
	BalanceAmount    float64 `json:"balance_amount"`
	ActualAmount     float64 `json:"actual_amount"`
}

type supplierRecSummary struct {
	SupplierID     int64   `json:"supplier_id"`
	SupplierCode   string  `json:"supplier_code"`
	SupplierName   string  `json:"supplier_name"`
	PayableAmount  float64 `json:"payable_amount"`
	VerifiedAmount float64 `json:"verified_amount"`
	BalanceAmount  float64 `json:"balance_amount"`
	ActualAmount   float64 `json:"actual_amount"`
}

type profitReport struct {
	SalesAmount float64 `json:"sales_amount"`
	CostAmount  float64 `json:"cost_amount"`
	Profit      float64 `json:"profit"`
}

func TestReport_ReconciliationSummariesAndProfit(t *testing.T) {
	env := testutil.LoadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	admin := testutil.NewClient(env.BaseURL)
	if _, err := admin.Login(ctx, env.AdminUser, env.AdminPass); err != nil {
		t.Fatalf("admin login failed: %v", err)
	}

	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "10")

	var custList []customerRecSummary
	var custTotal int64
	if err := admin.DoPage(ctx, http.MethodGet, "/api/v1/reports/reconciliation/customers", q, &custList, &custTotal); err != nil {
		t.Fatalf("customer reconciliation summary failed: %v", err)
	}

	var suppList []supplierRecSummary
	var suppTotal int64
	if err := admin.DoPage(ctx, http.MethodGet, "/api/v1/reports/reconciliation/suppliers", q, &suppList, &suppTotal); err != nil {
		t.Fatalf("supplier reconciliation summary failed: %v", err)
	}

	var pr profitReport
	if err := admin.DoJSON(ctx, http.MethodGet, "/api/v1/reports/profit", nil, nil, &pr); err != nil {
		t.Fatalf("profit report failed: %v", err)
	}

	_ = custTotal
	_ = suppTotal
}

func TestReport_DBAggregatesMatchAPI(t *testing.T) {
	env := testutil.LoadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	admin, fixture := mustLoginAndLoadFixture(ctx, t, env)
	pool := openOptionalDBPool(ctx, t)
	if pool == nil {
		t.Skip("TWC_DB_DSN is not set or unavailable; skip DB aggregate comparison")
	}

	materialID := mustCreateMaterial(ctx, t, admin, fixture.CategoryID, uniqueChineseName("报表对账聚合"), false)
	poID := mustCreatePurchaseOrderSingle(ctx, t, admin, fixture.SupplierCode, materialID, 2)
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/purchase/orders/%d/confirm", poID))
	po := waitPurchaseOrderDetailExt(ctx, t, admin, poID, func(p PurchaseOrderDetailExt) bool { return p.TotalAmount > 0 })
	supplierID := mustFindSupplierID(ctx, t, admin, fixture.SupplierCode)
	mustCreateFundPayment(ctx, t, admin, supplierID, po.OrderNo, po.OrderDate, po.TotalAmount, poID)
	stockInID := mustCreateManualStockIn(ctx, t, admin, fixture.MainMaterialWarehouse, materialID, 2, 10, "purchase")
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/stock-in/%d/confirm", stockInID))

	salesID := mustCreateSalesOrderSingle(ctx, t, admin, fixture.CustomerCode, materialID, 1, 88.8)
	mustConfirm(ctx, t, admin, fmt.Sprintf("/api/v1/sales/orders/%d/confirm", salesID))
	var soDetail struct {
		OrderNo     string  `json:"order_no"`
		OrderDate   string  `json:"order_date"`
		TotalAmount float64 `json:"total_amount"`
	}
	if err := admin.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/api/v1/sales/orders/%d", salesID), nil, nil, &soDetail); err != nil {
		t.Fatalf("get sales order detail failed: %v", err)
	}
	customerID := mustFindCustomerID(ctx, t, admin, fixture.CustomerCode)
	mustCreateFundCollection(ctx, t, admin, customerID, soDetail.OrderNo, soDetail.OrderDate, soDetail.TotalAmount, salesID)

	apiSupplier, ok := mustGetSupplierReconciliationSnapshot(ctx, t, admin, fixture.SupplierCode)
	if !ok {
		t.Fatalf("supplier reconciliation API row not found: supplierCode=%s", fixture.SupplierCode)
	}
	dbSupplier := mustQueryDBSupplierAggregate(ctx, t, pool, supplierID)
	assertFloatNear(t, apiSupplier.PayableAmount, dbSupplier.PayableAmount, 0.001, "供应商对账应付金额 API/SQL")
	assertFloatNear(t, apiSupplier.VerifiedAmount, dbSupplier.VerifiedAmount, 0.001, "供应商对账已核销金额 API/SQL")
	assertFloatNear(t, apiSupplier.BalanceAmount, dbSupplier.BalanceAmount, 0.001, "供应商对账余额 API/SQL")
	assertFloatNear(t, apiSupplier.ActualAmount, dbSupplier.ActualAmount, 0.001, "供应商对账实付金额 API/SQL")

	apiCustomer := mustGetCustomerReconciliationSnapshot(ctx, t, admin, fixture.CustomerCode)
	dbCustomer := mustQueryDBCustomerAggregate(ctx, t, pool, customerID)
	assertFloatNear(t, apiCustomer.ReceivableAmount, dbCustomer.ReceivableAmount, 0.001, "客户对账应收金额 API/SQL")
	assertFloatNear(t, apiCustomer.VerifiedAmount, dbCustomer.VerifiedAmount, 0.001, "客户对账已核销金额 API/SQL")
	assertFloatNear(t, apiCustomer.BalanceAmount, dbCustomer.BalanceAmount, 0.001, "客户对账余额 API/SQL")
	assertFloatNear(t, apiCustomer.ActualAmount, dbCustomer.ActualAmount, 0.001, "客户对账实收金额 API/SQL")

	var apiProfit profitReport
	if err := admin.DoJSON(ctx, http.MethodGet, "/api/v1/reports/profit", nil, nil, &apiProfit); err != nil {
		t.Fatalf("profit report failed: %v", err)
	}
	dbProfit := mustQueryDBProfitAggregate(ctx, t, pool)
	assertFloatNear(t, apiProfit.SalesAmount, dbProfit.SalesAmount, 0.001, "利润报表销售额 API/SQL")
	assertFloatNear(t, apiProfit.CostAmount, dbProfit.CostAmount, 0.001, "利润报表成本 API/SQL")
	assertFloatNear(t, apiProfit.Profit, dbProfit.Profit, 0.001, "利润报表利润 API/SQL")
}

func mustGetCustomerReconciliationSnapshot(ctx context.Context, t *testing.T, c *testutil.Client, customerCode string) customerRecSummary {
	t.Helper()
	for page := 1; page <= 5; page++ {
		q := url.Values{}
		q.Set("page", fmt.Sprintf("%d", page))
		q.Set("page_size", "100")

		var customers []customerRecSummary
		if err := c.DoPage(ctx, http.MethodGet, "/api/v1/reports/reconciliation/customers", q, &customers, nil); err != nil {
			t.Fatalf("query customer reconciliation failed: %v", err)
		}
		for _, row := range customers {
			if row.CustomerCode == customerCode {
				return row
			}
		}
		if len(customers) < 100 {
			break
		}
	}
	t.Fatalf("customer reconciliation row not found: customerCode=%s", customerCode)
	return customerRecSummary{}
}

func mustQueryDBSupplierAggregate(ctx context.Context, t *testing.T, pool *pgxpool.Pool, supplierID int64) supplierRecSummary {
	t.Helper()
	var row supplierRecSummary
	if err := pool.QueryRow(ctx, `
		WITH src AS (
			SELECT
				supplier_id,
				COALESCE(supplier_code, '') AS supplier_code,
				COALESCE(supplier_name, '') AS supplier_name,
				COALESCE(SUM(order_amount), 0)::float8 AS payable_amount,
				COALESCE(SUM(verified_amount), 0)::float8 AS verified_amount,
				COALESCE(SUM(unverified_amount), 0)::float8 AS balance_amount
			FROM v_fund_payment_source
			WHERE supplier_id = $1
			GROUP BY supplier_id, supplier_code, supplier_name
		),
		pay AS (
			SELECT
				supplier_id,
				COALESCE(SUM(actual_amount), 0)::float8 AS actual_amount
			FROM fund_payment
			WHERE deleted_at IS NULL AND status = 'completed' AND supplier_id = $1
			GROUP BY supplier_id
		)
		SELECT src.supplier_id, src.supplier_code, src.supplier_name,
		       src.payable_amount, src.verified_amount, src.balance_amount,
		       COALESCE(pay.actual_amount, 0)::float8
		FROM src
		LEFT JOIN pay ON pay.supplier_id = src.supplier_id
	`, supplierID).Scan(&row.SupplierID, &row.SupplierCode, &row.SupplierName, &row.PayableAmount, &row.VerifiedAmount, &row.BalanceAmount, &row.ActualAmount); err != nil {
		t.Fatalf("query db supplier aggregate failed: %v", err)
	}
	return row
}

func mustQueryDBCustomerAggregate(ctx context.Context, t *testing.T, pool *pgxpool.Pool, customerID int64) customerRecSummary {
	t.Helper()
	var row customerRecSummary
	if err := pool.QueryRow(ctx, `
		WITH src AS (
			SELECT
				customer_id,
				COALESCE(customer_code, '') AS customer_code,
				COALESCE(customer_name, '') AS customer_name,
				COALESCE(SUM(order_amount), 0)::float8 AS receivable_amount,
				COALESCE(SUM(verified_amount), 0)::float8 AS verified_amount,
				COALESCE(SUM(unverified_amount), 0)::float8 AS balance_amount
			FROM v_fund_collection_source
			WHERE customer_id = $1
			GROUP BY customer_id, customer_code, customer_name
		),
		pay AS (
			SELECT
				customer_id,
				COALESCE(SUM(actual_amount), 0)::float8 AS actual_amount
			FROM fund_collection
			WHERE deleted_at IS NULL AND status = 'completed' AND customer_id = $1
			GROUP BY customer_id
		)
		SELECT src.customer_id, src.customer_code, src.customer_name,
		       src.receivable_amount, src.verified_amount, src.balance_amount,
		       COALESCE(pay.actual_amount, 0)::float8
		FROM src
		LEFT JOIN pay ON pay.customer_id = src.customer_id
	`, customerID).Scan(&row.CustomerID, &row.CustomerCode, &row.CustomerName, &row.ReceivableAmount, &row.VerifiedAmount, &row.BalanceAmount, &row.ActualAmount); err != nil {
		t.Fatalf("query db customer aggregate failed: %v", err)
	}
	return row
}

func mustQueryDBProfitAggregate(ctx context.Context, t *testing.T, pool *pgxpool.Pool) profitReport {
	t.Helper()
	var row profitReport
	if err := pool.QueryRow(ctx, `
		WITH sales AS (
			SELECT COALESCE(SUM(COALESCE(total_amount, 0)), 0)::float8 AS sales_amount
			FROM sales_order
			WHERE deleted_at IS NULL AND order_status NOT IN ('draft', 'cancelled')
		),
		cost AS (
			SELECT COALESCE(SUM(soi.quantity * COALESCE(inv.unit_cost, 0)), 0)::float8 AS cost_amount
			FROM stock_out so
			INNER JOIN stock_out_item soi ON soi.stock_out_id = so.id
			LEFT JOIN inventory inv ON inv.id = soi.inventory_id
			WHERE so.deleted_at IS NULL
			  AND so.out_type = 'sales'
			  AND so.status = 'confirmed'
		)
		SELECT sales.sales_amount, cost.cost_amount, (sales.sales_amount - cost.cost_amount)::float8
		FROM sales, cost
	`).Scan(&row.SalesAmount, &row.CostAmount, &row.Profit); err != nil {
		t.Fatalf("query db profit aggregate failed: %v", err)
	}
	return row
}
