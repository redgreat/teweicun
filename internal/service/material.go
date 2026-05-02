/**
 * 功能：material.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
)

func ListMaterials(ctx context.Context, q *request.MaterialQuery) ([]response.MaterialResp, int64, error) {
	return db.ListMaterials(ctx, q)
}

func CreateMaterial(ctx context.Context, req *request.CreateMaterialReq, userID int64) (int64, error) {
	return db.CreateMaterial(ctx, req, userID)
}

func UpdateMaterial(ctx context.Context, id int64, req *request.UpdateMaterialReq, userID int64) error {
	return db.UpdateMaterial(ctx, id, req, userID)
}

func DeleteMaterial(ctx context.Context, id int64, userID int64) error {
	return db.DeleteMaterial(ctx, id, userID)
}
