/**
 * 功能：生产单/生产退货单 请求DTO定义
 * 创建时间：2026-06-06
 * 创建人：GPT-5.2
 */

package request

type ProductionOrderQuery struct {
	PageQuery
	ProductionNo string `form:"production_no"`
	Status       string `form:"status"`
}

type ProductionReturnOrderQuery struct {
	PageQuery
	ReturnNo string `form:"return_no"`
	Status   string `form:"status"`
}

// ProductionOrderUpdate 更新生产单（可编辑成本价格和备注）
type ProductionOrderUpdate struct {
	CostPrice float64 `json:"cost_price" binding:"omitempty,min=0"`
	Remark    string  `json:"remark" binding:"omitempty,max=500"`
}

// ProductionReturnOrderUpdate 更新生产退货单
type ProductionReturnOrderUpdate struct {
	CostPrice float64 `json:"cost_price" binding:"omitempty,min=0"`
	Remark    string  `json:"remark" binding:"omitempty,max=500"`
}

// CreateProductionReturnOrderReq 创建生产退货单
type CreateProductionReturnOrderReq struct {
	ProductionOrderID int64   `json:"production_order_id" binding:"required"`
	ReturnedQuantity  float64 `json:"returned_quantity" binding:"required,gt=0"`
	Remark            string  `json:"remark" binding:"omitempty,max=500"`
}

// CreateProductionOrderReq 手动创建生产单
type CreateProductionOrderReq struct {
	ProducedMaterialID   int64   `json:"produced_material_id" binding:"required"`
	ProducedWarehouseID  int64   `json:"produced_warehouse_id" binding:"required"`
	ProducedQuantity     float64 `json:"produced_quantity" binding:"required,gt=0"`
	CostPrice            float64 `json:"cost_price" binding:"omitempty,min=0"`
	ConsumptionOrderID   int64   `json:"consumption_order_id"`    // 单个领料单（兼容旧版）
	ConsumptionOrderIDs  []int64 `json:"consumption_order_ids"`   // 多个领料单
	Remark               string  `json:"remark" binding:"omitempty,max=500"`
}

