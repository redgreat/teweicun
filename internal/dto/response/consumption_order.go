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
	ProducedMaterialID  int64   `json:"produced_material_id"`
	ProducedWarehouseID int64   `json:"produced_warehouse_id"`
	ProducedQuantity    float64 `json:"produced_quantity"`
	ProducedMaterialCode string  `json:"produced_material_code"`
	ProducedMaterialName string  `json:"produced_material_name"`
	ProducedWarehouseCode string `json:"produced_warehouse_code"`
	ProducedWarehouseName string `json:"produced_warehouse_name"`
	ProductionOrderID    int64  `json:"production_order_id"`
	ProductionNo         string `json:"production_no"`
	ProductionStockInID  int64  `json:"production_stock_in_id"`
	ProductionStockInNo  string `json:"production_stock_in_no"`
	ProductionReturnOrderID   int64  `json:"production_return_order_id"`
	ProductionReturnNo        string `json:"production_return_no"`
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
