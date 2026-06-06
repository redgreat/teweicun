package request

import "time"

type ConsumptionOrderQuery struct {
	Page          int       `form:"page" binding:"required,min=1"`
	PageSize      int       `form:"page_size" binding:"required,min=1,max=100"`
	OrderNo       string    `form:"order_no"`
	ProjectNo     string    `form:"project_no"`
	ProductName   string    `form:"product_name"`
	WarehouseCode string    `form:"warehouse_code"`
	DesignerID    int64     `form:"designer_id"`
	Status        string    `form:"status"`
	StartDate     time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate       time.Time `form:"end_date" time_format:"2006-01-02"`
}

type ConsumptionOrderCreate struct {
	ProjectNo   string `json:"project_no" binding:"required,max=50"`
	ProductName string `json:"product_name" binding:"required,max=200"`
	// 注意：JSON 反序列化到 time.Time 会按 RFC3339 解析，前端传 YYYY-MM-DD 会失败；
	// 这里保留字符串并在 Handler 中按 2006-01-02 解析。
	OrderDate     string                       `json:"order_date" binding:"required,max=10"`
	OrderDateTime time.Time                    `json:"-"`                                        // Handler 解析后写入
	DesignerID    int64                        `json:"designer_id" binding:"omitempty,min=1"`    // 由 Handler 注入
	DesignerName  string                       `json:"designer_name" binding:"omitempty,max=50"` // 由 Handler 注入
	Remark        string                       `json:"remark"`
	ProducedMaterialID        int64   `json:"produced_material_id" binding:"omitempty,min=1"`
	ProducedQuantity          float64 `json:"produced_quantity" binding:"omitempty,gt=0"`
	ProducedWarehouseID       int64   `json:"produced_warehouse_id" binding:"omitempty,min=1"`
	ProductionOrderID         int64   `json:"production_order_id"`         // 关联已有生产单（可选）
	ProductionReturnOrderID   int64   `json:"production_return_order_id"`  // 关联已有生产退货单（可选）
	Items         []ConsumptionOrderItemCreate `json:"items" binding:"required,min=1,dive"`
}

type ConsumptionOrderItemCreate struct {
	MaterialID  int64   `json:"material_id" binding:"required,min=1"`
	InventoryID int64   `json:"inventory_id" binding:"required,min=1"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	Unit        string  `json:"unit" binding:"required,max=20"`
	Remark      string  `json:"remark" binding:"max=500"`
}

type ConsumptionOrderUpdate struct {
	ProjectNo   string                       `json:"project_no" binding:"max=50"`
	ProductName string                       `json:"product_name" binding:"max=200"`
	OrderDate   string                       `json:"order_date" binding:"max=10"`
	DesignerID  int64                        `json:"designer_id" binding:"omitempty,min=1"`
	DesignerName string                      `json:"designer_name" binding:"omitempty,max=50"`
	Remark      string                       `json:"remark"`
	ProducedMaterialID  int64   `json:"produced_material_id" binding:"omitempty,min=1"`
	ProducedQuantity    float64 `json:"produced_quantity" binding:"omitempty,gt=0"`
	ProducedWarehouseID int64   `json:"produced_warehouse_id" binding:"omitempty,min=1"`
	Items       []ConsumptionOrderItemUpdate `json:"items" binding:"required,min=1,dive"`
}

type ConsumptionOrderItemUpdate struct {
	MaterialID  int64   `json:"material_id" binding:"required,min=1"`
	InventoryID int64   `json:"inventory_id" binding:"required,min=1"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	Unit        string  `json:"unit" binding:"required,max=20"`
	Remark      string  `json:"remark" binding:"max=500"`
}

type ConsumptionOrderConfirm struct {
	OrderID int64 `json:"order_id" binding:"required,min=1"`
}
