/**
 * 功能：dict.go
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

func ListDictTypes(ctx context.Context, q *request.DictTypeQuery) ([]response.DictTypeResp, int64, error) {
	return db.ListDictTypes(ctx, q)
}

func ListDictDataByDictType(ctx context.Context, dictType string) ([]response.DictDataResp, error) {
	return db.ListDictData(ctx, dictType)
}

func CreateDictType(ctx context.Context, req *request.CreateDictTypeReq) (int64, error) {
	return db.CreateDictType(ctx, req)
}

func UpdateDictType(ctx context.Context, id int64, req *request.UpdateDictTypeReq) error {
	return db.UpdateDictType(ctx, id, req)
}

func DeleteDictType(ctx context.Context, id int64) error {
	return db.DeleteDictType(ctx, id)
}

func CreateDictData(ctx context.Context, req *request.CreateDictDataReq) (int64, error) {
	return db.CreateDictData(ctx, req)
}

func UpdateDictData(ctx context.Context, id int64, req *request.UpdateDictDataReq) error {
	return db.UpdateDictData(ctx, id, req)
}

func DeleteDictData(ctx context.Context, id int64) error {
	return db.DeleteDictData(ctx, id)
}
