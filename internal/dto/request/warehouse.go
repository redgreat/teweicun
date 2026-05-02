/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type WarehouseQuery struct {
	PageQuery
	WarehouseName string `form:"warehouse_name"`
	WarehouseCode string `form:"warehouse_code"`
	Status        string `form:"status"`
}

type CreateWarehouseReq struct {
	WarehouseCode  string `json:"warehouse_code"`
	WarehouseName  string `json:"warehouse_name" binding:"required"`
	WarehouseType  string `json:"warehouse_type" binding:"required"`
	ManagerID      int64  `json:"manager_id" binding:"required"`
}

type UpdateWarehouseReq struct {
	WarehouseName  string `json:"warehouse_name" binding:"required"`
	WarehouseType  string `json:"warehouse_type" binding:"required"`
	ManagerID      int64  `json:"manager_id" binding:"required"`
	Status         string `json:"status" binding:"oneof=enabled disabled"`
}
