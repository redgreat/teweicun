package response

type DashboardKPI struct {
	PurchaseQty          float64 `json:"purchase_qty"`
	PurchaseAmount       float64 `json:"purchase_amount"`
	ConsumptionQty       float64 `json:"consumption_qty"`
	ConsumptionAmount    float64 `json:"consumption_amount"`
	ReversalAmount       float64 `json:"reversal_amount"`
	NetConsumptionAmount float64 `json:"net_consumption_amount"`
	InventoryAmount      float64 `json:"inventory_amount"`
}

type DashboardTrendPoint struct {
	BizDate           string  `json:"biz_date"`
	PurchaseQty       float64 `json:"purchase_qty"`
	PurchaseAmount    float64 `json:"purchase_amount"`
	ConsumptionQty    float64 `json:"consumption_qty"`
	ConsumptionAmount float64 `json:"consumption_amount"`
}

type DashboardTopItem struct {
	MaterialID   int64   `json:"material_id"`
	MaterialCode string  `json:"material_code"`
	MaterialName string  `json:"material_name"`
	Quantity     float64 `json:"quantity"`
	Amount       float64 `json:"amount"`
}

type DashboardBusinessSummary struct {
	PurchaseMinusConsumptionAmount float64 `json:"purchase_minus_consumption_amount"`
	ActiveMaterialCount            int64   `json:"active_material_count"`
	ActiveSerialCount              int64   `json:"active_serial_count"`
	MaxSingleConsumptionAmount     float64 `json:"max_single_consumption_amount"`
}

type DashboardBigscreenResp struct {
	Range                string                   `json:"range"`
	UpdatedAt            string                   `json:"updated_at"`
	KPI                  DashboardKPI             `json:"kpi"`
	Trend                []DashboardTrendPoint    `json:"trend"`
	TopPurchaseAmount    []DashboardTopItem       `json:"top_purchase_amount"`
	TopConsumptionAmount []DashboardTopItem       `json:"top_consumption_amount"`
	TopConsumptionQty    []DashboardTopItem       `json:"top_consumption_qty"`
	Summary              DashboardBusinessSummary `json:"summary"`
}
