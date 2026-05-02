package response

import "time"

type ConsumptionOrderResp struct {
	ID            int64                      `json:"id"`
	OrderNo       string                     `json:"order_no"`
	ProjectNo     string                     `json:"project_no"`
	ProductName   string                     `json:"product_name"`
	WarehouseID   int64                      `json:"warehouse_id"`
	WarehouseCode string                     `json:"warehouse_code"`
	WarehouseName string                     `json:"warehouse_name"`
	OrderDate     time.Time                  `json:"order_date"`
	DesignerID    int64                      `json:"designer_id"`
	DesignerName  string                     `json:"designer_name"`
	Status        string                     `json:"status"`
	StatusName    string                     `json:"status_name"`
	StockOutID    int64                      `json:"stock_out_id"`
	StockOutNo    string                     `json:"stock_out_no"`
	Remark        string                     `json:"remark"`
	CreatedAt     time.Time                  `json:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at"`
	Items         []ConsumptionOrderItemResp `json:"items"`
	ItemCount     int                        `json:"item_count"`
	TotalQuantity float64                    `json:"total_quantity"`
	TotalAmount   float64                    `json:"total_amount"`
}

type ConsumptionOrderItemResp struct {
	ID           int64   `json:"id"`
	OrderID      int64   `json:"order_id"`
	MaterialID   int64   `json:"material_id"`
	MaterialCode string  `json:"material_code"`
	MaterialName string  `json:"material_name"`
	InventoryID  int64   `json:"inventory_id"`
	SKUID        int64   `json:"sku_id"`
	SKUCode      string  `json:"sku_code"`
	SKUName      string  `json:"sku_name"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	UnitCost     float64 `json:"unit_cost,omitempty"`
	Remark       string  `json:"remark"`
}

type ConsumptionOrderListResp struct {
	List  []ConsumptionOrderResp `json:"list"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Size  int                    `json:"size"`
}
