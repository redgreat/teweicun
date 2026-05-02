/**
 * 功能：trace.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package db

import (
	"context"

	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

// TraceForward 正向追溯 (输入批次号)
func TraceForward(ctx context.Context, batchNo string) ([]response.TraceForwardResp, error) {
	query := `
		SELECT trace_type, doc_no, doc_date::text, material_code, material_name, 
		       quantity, warehouse
		FROM fn_trace_forward($1)
	`
	rows, err := database.Pool.Query(ctx, query, batchNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []response.TraceForwardResp
	for rows.Next() {
		var item response.TraceForwardResp
		if err := rows.Scan(&item.TraceType, &item.DocNo, &item.DocDate, &item.MaterialCode, &item.MaterialName,
			&item.Quantity, &item.Warehouse); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

// TraceBackward 反向追溯 (输入项目编号/产品名称)
func TraceBackward(ctx context.Context, keyword string) ([]response.TraceBackwardResp, error) {
	query := `
		SELECT trace_type, doc_no, doc_date::text, material_code, material_name, 
		       quantity, COALESCE(supplier_name, ''), COALESCE(cert_no, '')
		FROM fn_trace_backward($1)
	`
	rows, err := database.Pool.Query(ctx, query, keyword)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []response.TraceBackwardResp
	for rows.Next() {
		var item response.TraceBackwardResp
		if err := rows.Scan(&item.TraceType, &item.DocNo, &item.DocDate, &item.MaterialCode, &item.MaterialName,
			&item.Quantity, &item.SupplierName, &item.CertNo); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}
