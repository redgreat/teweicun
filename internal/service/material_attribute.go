/**
 * 功能：物料属性业务逻辑
 * 创建时间：2026-04-17
 * 创建人：wangcw
 */

package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
)

func ListMaterialAttributeDefs(ctx context.Context, q *request.MaterialAttributeDefQuery) (*response.PageResult, error) {
	items, total, err := db.ListMaterialAttributeDefs(ctx, q)
	if err != nil {
		return nil, err
	}
	return &response.PageResult{
		List:  items,
		Total: total,
	}, nil
}

func CreateMaterialAttributeDef(ctx context.Context, req *request.CreateMaterialAttributeDefReq, userID int64) (int64, error) {
	return db.CreateMaterialAttributeDef(ctx, req, userID)
}

func UpdateMaterialAttributeDef(ctx context.Context, id int64, req *request.UpdateMaterialAttributeDefReq, userID int64) error {
	return db.UpdateMaterialAttributeDef(ctx, id, req, userID)
}

func DeleteMaterialAttributeDef(ctx context.Context, id int64, userID int64) error {
	return db.DeleteMaterialAttributeDef(ctx, id, userID)
}

func GetMaterialAttributes(ctx context.Context, materialID int64) ([]response.MaterialAttributeValueResp, error) {
	return db.GetMaterialAttributes(ctx, materialID)
}

func UpdateMaterialAttributes(ctx context.Context, materialID int64, req *request.UpdateMaterialAttributesReq, userID int64) error {
	return db.UpdateMaterialAttributes(ctx, materialID, req, userID)
}
