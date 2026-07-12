/**
 * 功能：生产单/生产退货单 响应DTO定义
 * 创建时间：2026-06-06
 * 创建人：GPT-5.2
 */

package response

import "time"

type ProductionOrderResp struct {
	ID                   int64     `json:"id"`
	ProductionNo         string    `json:"production_no"`
	Status                 string    `json:"status"`
	StatusName             string    `json:"status_name,omitempty"`
	ConsumptionOrderID   int64     `json:"consumption_order_id"`
	ConsumptionOrderNo   string    `json:"consumption_order_no"`
	StockOutID           int64     `json:"stock_out_id"`
	StockOutNo           string    `json:"stock_out_no"`
	StockInID            int64     `json:"stock_in_id"`
	StockInNo            string    `json:"stock_in_no"`
	ProducedMaterialID   int64     `json:"produced_material_id"`
	ProducedMaterialCode string    `json:"produced_material_code"`
	ProducedMaterialName string    `json:"produced_material_name"`
	ProducedWarehouseID  int64     `json:"produced_warehouse_id"`
	ProducedWarehouseCode string   `json:"produced_warehouse_code"`
	ProducedWarehouseName string   `json:"produced_warehouse_name"`
	ProducedQuantity     float64   `json:"produced_quantity"`
	ProducedUnitCost     float64   `json:"produced_unit_cost"`
	CostPrice            float64   `json:"cost_price"`
	Remark               string    `json:"remark"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ProductionOrderListResp struct {
	List  []ProductionOrderResp `json:"list"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Size  int                   `json:"size"`
}

type ProductionReturnOrderResp struct {
	ID                    int64     `json:"id"`
	ReturnNo              string    `json:"return_no"`
	Status                 string    `json:"status"`
	StatusName             string    `json:"status_name,omitempty"`
	ProductionOrderID      int64     `json:"production_order_id"`
	ProductionNo           string    `json:"production_no"`
	ConsumptionOrderID     int64     `json:"consumption_order_id"`
	ConsumptionOrderNo     string    `json:"consumption_order_no"`
	StockOutID             int64     `json:"stock_out_id"`
	StockOutNo             string    `json:"stock_out_no"`
	ProducedMaterialID     int64     `json:"produced_material_id"`
	ProducedMaterialCode   string    `json:"produced_material_code"`
	ProducedMaterialName   string    `json:"produced_material_name"`
	ProducedWarehouseID    int64     `json:"produced_warehouse_id"`
	ProducedWarehouseCode  string   `json:"produced_warehouse_code"`
	ProducedWarehouseName  string   `json:"produced_warehouse_name"`
	ReturnedQuantity       float64   `json:"returned_quantity"`
	CostPrice              float64   `json:"cost_price"`
	Remark                 string    `json:"remark"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type ProductionReturnOrderListResp struct {
	List  []ProductionReturnOrderResp `json:"list"`
	Total int64                       `json:"total"`
	Page  int                         `json:"page"`
	Size  int                         `json:"size"`
}

