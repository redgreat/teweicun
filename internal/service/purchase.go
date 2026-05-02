/**
 * 功能：采购订单业务服务
 * 创建时间：2026-04-18
 * 创建人：CodeArts Agent
 */

package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
)

func ListPurchaseOrders(ctx context.Context, q *request.PurchaseOrderQuery) ([]response.PurchaseOrderResp, int64, error) {
	return db.ListPurchaseOrders(ctx, q)
}

func GetPurchaseOrder(ctx context.Context, id int64) (*response.PurchaseOrderDetailResp, error) {
	return db.GetPurchaseOrderByID(ctx, id)
}

func CreatePurchaseOrder(ctx context.Context, req *request.CreatePurchaseOrderReq, userID int64) (int64, error) {
	return db.CreatePurchaseOrder(ctx, req, userID)
}

func UpdatePurchaseOrder(ctx context.Context, id int64, req *request.UpdatePurchaseOrderReq) error {
	return db.UpdatePurchaseOrder(ctx, id, req)
}

func DeletePurchaseOrder(ctx context.Context, id int64) error {
	return db.DeletePurchaseOrder(ctx, id)
}

func ConfirmPurchaseOrder(ctx context.Context, orderID, userID int64) error {
	return db.ConfirmPurchaseOrder(ctx, orderID, userID)
}
