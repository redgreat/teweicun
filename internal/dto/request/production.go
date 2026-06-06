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

