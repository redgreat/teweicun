/**
 * 功能：inventory.go
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

func ListInventoryDetail(ctx context.Context, q *request.InventoryQuery) ([]response.InventoryDetailResp, int64, error) {
	return db.ListInventoryDetail(ctx, q)
}

func ListInventorySummary(ctx context.Context, q *request.InventoryQuery) ([]response.InventorySummaryResp, int64, error) {
	return db.ListInventorySummary(ctx, q)
}

func ListInventoryAvailable(ctx context.Context, q *request.InventoryAvailableQuery) ([]response.InventoryAvailableResp, int64, error) {
	return db.ListInventoryAvailable(ctx, q)
}

func ListInventoryIssued(ctx context.Context, q *request.InventoryIssuedQuery) ([]response.InventoryIssuedResp, int64, error) {
	return db.ListInventoryIssued(ctx, q)
}

func ListInventoryMaterialLedger(ctx context.Context, q *request.InventoryMaterialLedgerQuery) ([]response.InventoryMaterialLedgerResp, int64, *response.InventoryMaterialLedgerStatsResp, error) {
	return db.ListInventoryMaterialLedger(ctx, q)
}

func ListInventoryMaterialLedgerSerials(ctx context.Context, q *request.InventoryMaterialLedgerSerialQuery) ([]response.InventoryMaterialLedgerSerialResp, error) {
	return db.ListInventoryMaterialLedgerSerials(ctx, q)
}

func ExportInventoryMaterialLedger(ctx context.Context, q *request.InventoryMaterialLedgerQuery) ([]response.InventoryMaterialLedgerResp, *response.InventoryMaterialLedgerStatsResp, error) {
	return db.ExportInventoryMaterialLedger(ctx, q)
}
