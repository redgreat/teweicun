/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

type StockInSummaryReport struct {
	ReportMonth   string  `json:"report_month"`
	SupplierName  string  `json:"supplier_name"`
	MaterialCode  string  `json:"material_code"`
	MaterialName  string  `json:"material_name"`
	Unit          string  `json:"unit"`
	TotalQuantity float64 `json:"total_quantity"`
	OrderCount    int     `json:"order_count"`
}

type StockOutSummaryReport struct {
	ReportMonth   string  `json:"report_month"`
	OutType       string  `json:"out_type"`
	MaterialCode  string  `json:"material_code"`
	MaterialName  string  `json:"material_name"`
	Unit          string  `json:"unit"`
	TotalQuantity float64 `json:"total_quantity"`
	OrderCount    int     `json:"order_count"`
}

type InventoryStatusReport struct {
	WarehouseName     string  `json:"warehouse_name"`
	CategoryName      string  `json:"category_name"`
	MaterialCode      string  `json:"material_code"`
	MaterialName      string  `json:"material_name"`
	Unit              string  `json:"unit"`
	CurrentQuantity   float64 `json:"current_quantity"`
	LockedQuantity    float64 `json:"locked_quantity"`
	AvailableQuantity float64 `json:"available_quantity"`
}

type CustomerReconciliationSummaryReport struct {
	CustomerID       int64   `json:"customer_id"`
	CustomerCode     string  `json:"customer_code"`
	CustomerName     string  `json:"customer_name"`
	ReceivableAmount float64 `json:"receivable_amount"`
	VerifiedAmount   float64 `json:"verified_amount"`
	BalanceAmount    float64 `json:"balance_amount"`
	ActualAmount     float64 `json:"actual_amount"`
	InvoiceAmount    float64 `json:"invoice_amount"`
	DiscountAmount   float64 `json:"discount_amount"`
}

type SupplierReconciliationSummaryReport struct {
	SupplierID     int64   `json:"supplier_id"`
	SupplierCode   string  `json:"supplier_code"`
	SupplierName   string  `json:"supplier_name"`
	PayableAmount  float64 `json:"payable_amount"`
	VerifiedAmount float64 `json:"verified_amount"`
	BalanceAmount  float64 `json:"balance_amount"`
	ActualAmount   float64 `json:"actual_amount"`
	InvoiceAmount  float64 `json:"invoice_amount"`
	DiscountAmount float64 `json:"discount_amount"`
}

type ProfitReport struct {
	SalesAmount float64 `json:"sales_amount"`
	CostAmount  float64 `json:"cost_amount"`
	Profit      float64 `json:"profit"`
}
