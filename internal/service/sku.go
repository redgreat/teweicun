/**
 * 功能：SKU 服务层
 * 创建时间：2026-04-19
 * 创建人：wangcw
 */

package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
)

func ListSKUs(ctx context.Context, q *request.SKUQuery) ([]response.SKUListItem, int64, error) {
	return db.ListSKUs(ctx, q)
}

func GetSKU(ctx context.Context, id int64) (*response.SKUDetail, error) {
	return db.GetSKUByID(ctx, id)
}

func CreateSKU(ctx context.Context, req *request.CreateSKUReq, userID int64) (int64, error) {
	return db.CreateSKU(ctx, req, userID)
}

func UpdateSKU(ctx context.Context, id int64, req *request.UpdateSKUReq, userID int64) error {
	return db.UpdateSKU(ctx, id, req, userID)
}

func DeleteSKU(ctx context.Context, id int64, userID int64) error {
	return db.DeleteSKU(ctx, id, userID)
}

func ListSKUsByMaterial(ctx context.Context, materialID int64) ([]response.SKUSelectItem, error) {
	return db.ListSKUsByMaterial(ctx, materialID)
}
