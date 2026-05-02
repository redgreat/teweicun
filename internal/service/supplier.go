/**
 * 功能：supplier.go
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

func ListSuppliers(ctx context.Context, q *request.SupplierQuery) ([]response.SupplierResp, int64, error) {
	return db.ListSuppliers(ctx, q)
}

func CreateSupplier(ctx context.Context, req *request.CreateSupplierReq, userID int64, username string) (int64, error) {
	return db.CreateSupplier(ctx, req, userID, username)
}

func UpdateSupplier(ctx context.Context, id int64, req *request.UpdateSupplierReq, userID int64, username string) error {
	return db.UpdateSupplier(ctx, id, req, userID, username)
}

func DeleteSupplier(ctx context.Context, id int64, userID int64, username string) error {
	return db.DeleteSupplier(ctx, id, userID, username)
}
