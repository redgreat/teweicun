/**
 * 功能：customer.go
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

func ListCustomers(ctx context.Context, q *request.CustomerQuery) ([]response.CustomerResp, int64, error) {
	return db.ListCustomers(ctx, q)
}

func CreateCustomer(ctx context.Context, req *request.CreateCustomerReq, userID int64, username string) (int64, error) {
	return db.CreateCustomer(ctx, req, userID, username)
}

func UpdateCustomer(ctx context.Context, id int64, req *request.UpdateCustomerReq, userID int64, username string) error {
	return db.UpdateCustomer(ctx, id, req, userID, username)
}

func DeleteCustomer(ctx context.Context, id int64, userID int64, username string) error {
	return db.DeleteCustomer(ctx, id, userID, username)
}
