package request

import "time"

type ReversalOrderQuery struct {
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

type ReversalOrderCreate struct {
	ProjectNo     string                    `json:"project_no" binding:"required,max=50"`
	ProductName   string                    `json:"product_name" binding:"required,max=200"`
	WarehouseID   int64                     `json:"warehouse_id"`   // 可选；为 0 时按明细 inventory_id 解析单仓
	WarehouseCode string                    `json:"warehouse_code"` // 与 warehouse_id 同时传时须与库存一致
	OrderDate     string                    `json:"order_date" binding:"required"` // 前端传 YYYY-MM-DD
	DesignerID    int64                     `json:"designer_id"`                    // 由后端登录态注入
	DesignerName  string                    `json:"designer_name"`                  // 由后端登录态注入
	Remark        string                    `json:"remark"`
	Items         []ReversalOrderItemCreate `json:"items" binding:"required,min=1,dive"`
}

type ReversalOrderItemCreate struct {
	InventoryID int64   `json:"inventory_id" binding:"required,min=1"`
	MaterialID  int64   `json:"material_id"` // 可选；服务端以库存为准并校验
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	Unit        string  `json:"unit" binding:"max=20"` // 可空，服务端用库存/物料单位
	Remark      string  `json:"remark"`
}

type ReversalOrderUpdate struct {
	ProjectNo     string                     `json:"project_no" binding:"max=50"`
	ProductName   string                     `json:"product_name" binding:"max=200"`
	WarehouseID   int64                      `json:"warehouse_id"`
	WarehouseCode string                     `json:"warehouse_code"`
	OrderDate     string                     `json:"order_date" binding:"max=10"`
	DesignerID    int64                      `json:"designer_id" binding:"omitempty,min=1"`
	DesignerName  string                     `json:"designer_name" binding:"omitempty,max=50"`
	Remark        string                     `json:"remark"`
	Items         []ReversalOrderItemUpdate  `json:"items" binding:"required,min=1,dive"`
}

type ReversalOrderItemUpdate struct {
	InventoryID int64   `json:"inventory_id" binding:"required,min=1"`
	MaterialID  int64   `json:"material_id"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	Unit        string  `json:"unit" binding:"max=20"`
	Remark      string  `json:"remark"`
}

type ReversalOrderConfirm struct {
	OrderID int64 `json:"order_id" binding:"required,min=1"`
}