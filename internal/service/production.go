/**
 * 功能：生产单/生产退货单 业务层
 * 创建时间：2026-06-06
 * 创建人：GPT-5.2
 */

package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
)

func ListProductionOrders(ctx context.Context, q request.ProductionOrderQuery) (*response.ProductionOrderListResp, error) {
	return db.ListProductionOrders(ctx, q)
}

func GetProductionOrderDetail(ctx context.Context, id int64) (*response.ProductionOrderResp, error) {
	return db.GetProductionOrderDetail(ctx, id)
}

func UpdateProductionOrder(ctx context.Context, id int64, req request.ProductionOrderUpdate, userID int64) error {
	return db.UpdateProductionOrder(ctx, id, req, userID)
}

func ListConsumptionOrdersByProduction(ctx context.Context, productionOrderID int64) ([]response.ConsumptionOrderResp, error) {
	return db.ListConsumptionOrdersByProduction(ctx, productionOrderID)
}

func ListReversalOrdersByProduction(ctx context.Context, productionOrderID int64) ([]response.ReversalOrderResp, error) {
	return db.ListReversalOrdersByProduction(ctx, productionOrderID)
}

func ListProductionReturnOrders(ctx context.Context, q request.ProductionReturnOrderQuery) (*response.ProductionReturnOrderListResp, error) {
	return db.ListProductionReturnOrders(ctx, q)
}

func GetProductionReturnOrderDetail(ctx context.Context, id int64) (*response.ProductionReturnOrderResp, error) {
	return db.GetProductionReturnOrderDetail(ctx, id)
}

func UpdateProductionReturnOrder(ctx context.Context, id int64, req request.ProductionReturnOrderUpdate, userID int64) error {
	return db.UpdateProductionReturnOrder(ctx, id, req, userID)
}

func ListConsumptionOrdersByProductionReturn(ctx context.Context, returnOrderID int64) ([]response.ConsumptionOrderResp, error) {
	return db.ListConsumptionOrdersByProductionReturn(ctx, returnOrderID)
}

func ListReversalOrdersByProductionReturn(ctx context.Context, returnOrderID int64) ([]response.ReversalOrderResp, error) {
	return db.ListReversalOrdersByProductionReturn(ctx, returnOrderID)
}

func ListProductionOrdersForDropdown(ctx context.Context, keyword string) ([]map[string]interface{}, error) {
	return db.ListProductionOrdersForDropdown(ctx, keyword)
}

func ListProductionReturnOrdersForDropdown(ctx context.Context, keyword string) ([]map[string]interface{}, error) {
	return db.ListProductionReturnOrdersForDropdown(ctx, keyword)
}

func CreateProductionReturnOrder(ctx context.Context, req request.CreateProductionReturnOrderReq, userID int64, username string) (int64, error) {
	return db.CreateProductionReturnOrder(ctx, req, userID, username)
}
