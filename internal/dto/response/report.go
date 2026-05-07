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
