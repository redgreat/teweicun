package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
)

func ListFundPayments(ctx context.Context, q *request.FundPaymentQuery) ([]response.FundPaymentResp, int64, error) {
	return db.ListFundPayments(ctx, q)
}

func GetFundPayment(ctx context.Context, id int64) (*response.FundPaymentDetailResp, error) {
	return db.GetFundPayment(ctx, id)
}

func ListFundPaymentSources(ctx context.Context, q *request.FundPaymentSourceQuery) ([]response.FundPaymentSourceResp, error) {
	return db.ListFundPaymentSources(ctx, q)
}

func CreateFundPayment(ctx context.Context, req *request.CreateFundPaymentReq, userID int64, username string) (int64, error) {
	return db.CreateFundPayment(ctx, req, userID, username)
}

func ListFundCollections(ctx context.Context, q *request.FundCollectionQuery) ([]response.FundCollectionResp, int64, error) {
	return db.ListFundCollections(ctx, q)
}

func GetFundCollection(ctx context.Context, id int64) (*response.FundCollectionDetailResp, error) {
	return db.GetFundCollection(ctx, id)
}

func ListFundCollectionSources(ctx context.Context, q *request.FundCollectionSourceQuery) ([]response.FundCollectionSourceResp, error) {
	return db.ListFundCollectionSources(ctx, q)
}

func CreateFundCollection(ctx context.Context, req *request.CreateFundCollectionReq, userID int64, username string) (int64, error) {
	return db.CreateFundCollection(ctx, req, userID, username)
}
