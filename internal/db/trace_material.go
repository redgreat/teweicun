package db

import (
	"context"
	"time"

	"github.com/redgreat/teweicun/pkg/database"
)

type SerialCodeInfo struct {
	ID            int64     `json:"id"`
	SerialCode    string    `json:"serial_code"`
	MaterialID    int64     `json:"material_id"`
	MaterialCode  string    `json:"material_code"`
	MaterialName  string    `json:"material_name"`
	Status        string    `json:"status"`
	WarehouseName string    `json:"warehouse_name"`
	CreatedAt     time.Time `json:"created_at"`
}

type SerialTraceRecord struct {
	ID            int64     `json:"id"`
	SerialCode    string    `json:"serial_code"`
	MaterialName  string    `json:"material_name"`
	MaterialCode  string    `json:"material_code"`
	Status        string    `json:"status"`
	Action        string    `json:"action"`
	ActionLabel   string    `json:"action_label"`
	RefDocType    string    `json:"ref_doc_type"`
	RefDocID      int64     `json:"ref_doc_id"`
	RefDocNo      string    `json:"ref_doc_no"`
	OperatorName  string    `json:"operator_name"`
	ActionTime    time.Time `json:"action_time"`
	FromWarehouse string    `json:"from_warehouse"`
	ToWarehouse   string    `json:"to_warehouse"`
	Remark        string    `json:"remark"`
}

type BatchSerialInfo struct {
	SerialCode    string    `json:"serial_code"`
	MaterialName  string    `json:"material_name"`
	MaterialCode  string    `json:"material_code"`
	Status        string    `json:"status"`
	StatusLabel   string    `json:"status_label"`
	WarehouseName string    `json:"warehouse_name"`
	CreatedAt     time.Time `json:"created_at"`
}

func QueryTraceBySerial(ctx context.Context, serialCode string) (*SerialCodeInfo, []SerialTraceRecord, error) {
	infoSQL := `
		SELECT sc.id, sc.serial_code, sc.material_id, sc.material_code,
		       sc.material_name, sc.status,
		       COALESCE(w.warehouse_name, ''),
		       sc.created_at
		FROM sku_serial_code sc
		LEFT JOIN warehouse w ON w.id = sc.warehouse_id
		WHERE sc.serial_code ILIKE '%' || $1 || '%'
		ORDER BY
			CASE
				WHEN sc.serial_code = $1 THEN 0
				WHEN sc.serial_code ILIKE $1 || '%' THEN 1
				ELSE 2
			END,
			sc.created_at DESC
		LIMIT 1
	`
	var info SerialCodeInfo
	err := database.Pool.QueryRow(ctx, infoSQL, serialCode).Scan(
		&info.ID, &info.SerialCode, &info.MaterialID,
		&info.MaterialCode, &info.MaterialName,
		&info.Status, &info.WarehouseName,
		&info.CreatedAt,
	)
	if err != nil {
		return nil, nil, err
	}

	traceSQL := `
		SELECT st.id, st.serial_code, sc.material_name, sc.material_code,
		       sc.status, st.action,
		       CASE st.action
		           WHEN 'stock_in' THEN '采购入库'
		           WHEN 'stock_out' THEN '领料出库'
		           WHEN 'return' THEN '退料退回'
		           WHEN 'transfer' THEN '仓库调拨'
		           WHEN 'scrap' THEN '报废处理'
		           ELSE st.action
		       END,
		       COALESCE(st.ref_doc_type, ''), COALESCE(st.ref_doc_id, 0), COALESCE(st.ref_doc_no, ''),
		       COALESCE(st.operator_name, ''),
		       st.created_at,
		       COALESCE(fw.warehouse_name, ''), COALESCE(tw.warehouse_name, ''),
		       COALESCE(st.remark, '')
		FROM sku_serial_trace st
		JOIN sku_serial_code sc ON sc.id = st.serial_code_id
		LEFT JOIN warehouse fw ON fw.id = st.from_warehouse_id
		LEFT JOIN warehouse tw ON tw.id = st.to_warehouse_id
		WHERE st.serial_code_id = $1
		ORDER BY st.created_at ASC
	`
	rows, err := database.Pool.Query(ctx, traceSQL, info.ID)
	if err != nil {
		return &info, nil, err
	}
	defer rows.Close()

	var traces []SerialTraceRecord
	for rows.Next() {
		var t SerialTraceRecord
		err := rows.Scan(
			&t.ID, &t.SerialCode, &t.MaterialName, &t.MaterialCode,
			&t.Status, &t.Action, &t.ActionLabel,
			&t.RefDocType, &t.RefDocID, &t.RefDocNo, &t.OperatorName,
			&t.ActionTime, &t.FromWarehouse, &t.ToWarehouse, &t.Remark,
		)
		if err != nil {
			return &info, traces, err
		}
		traces = append(traces, t)
	}

	return &info, traces, nil
}
