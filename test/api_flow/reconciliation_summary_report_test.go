package api_flow

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

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
