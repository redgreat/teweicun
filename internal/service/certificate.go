/**
 * 功能：certificate.go
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

func ListCertificates(ctx context.Context, q *request.CertificateQuery) ([]response.CertificateResp, int64, error) {
	return db.ListCertificates(ctx, q)
}

func CreateCertificate(ctx context.Context, req *request.CreateCertificateReq, userID int64, username string) (int64, error) {
	return db.CreateCertificate(ctx, req, userID, username)
}

func UpdateCertificate(ctx context.Context, id int64, req *request.UpdateCertificateReq, userID int64, username string) error {
	return db.UpdateCertificate(ctx, id, req, userID, username)
}

func DeleteCertificate(ctx context.Context, id int64, userID int64, username string) error {
	return db.DeleteCertificate(ctx, id, userID, username)
}
