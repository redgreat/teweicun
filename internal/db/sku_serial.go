package db

import (
	"context"
	"time"

	"github.com/redgreat/teweicun/pkg/database"
)

type SkuSerialCodeItem struct {
	ID            int64     `json:"id"`
	SerialCode    string    `json:"serial_code"`
	MaterialID    int64     `json:"material_id"`
	MaterialCode  string    `json:"material_code"`
	MaterialName  string    `json:"material_name"`
	Status        string    `json:"status"`
	StatusLabel   string    `json:"status_label"`
	WarehouseName string    `json:"warehouse_name"`
	Selected      bool      `json:"selected"`
	CreatedAt     time.Time `json:"created_at"`
}

func QuerySerialCodesByStockInItem(ctx context.Context, stockInItemID int64) ([]SkuSerialCodeItem, error) {
	sql := `
		SELECT sc.id, sc.serial_code, sc.material_id, sc.material_code,
		       sc.material_name, sc.status,
		       CASE sc.status
		           WHEN 'in_stock' THEN '在库'
		           WHEN 'issued' THEN '已领用'
		           WHEN 'returned' THEN '已退回'
		           WHEN 'scrapped' THEN '已报废'
		           ELSE sc.status
		       END,
		       COALESCE(w.warehouse_name, ''), false AS selected,
		       sc.created_at
		FROM sku_serial_code sc
		LEFT JOIN warehouse w ON w.id = sc.warehouse_id
		WHERE sc.stock_in_item_id = $1
		ORDER BY sc.serial_code ASC
	`
	rows, err := database.Pool.Query(ctx, sql, stockInItemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]SkuSerialCodeItem, 0)
	for rows.Next() {
		var item SkuSerialCodeItem
		err := rows.Scan(
			&item.ID, &item.SerialCode, &item.MaterialID, &item.MaterialCode,
			&item.MaterialName, &item.Status, &item.StatusLabel,
			&item.WarehouseName, &item.Selected, &item.CreatedAt,
		)
		if err != nil {
			return items, err
		}
		items = append(items, item)
	}
	return items, nil
}

func QuerySerialCodesByStockOutItem(ctx context.Context, stockOutItemID int64) ([]SkuSerialCodeItem, error) {
	sql := `
		WITH target AS (
			SELECT soi.stock_out_id, soi.inventory_id, soi.material_id, so.status AS stock_out_status
			FROM stock_out_item soi
			INNER JOIN stock_out so ON so.id = soi.stock_out_id
			WHERE soi.id = $1
		)
		SELECT sc.id, sc.serial_code, sc.material_id, sc.material_code,
		       sc.material_name, sc.status,
		       CASE sc.status
		           WHEN 'in_stock' THEN '在库'
		           WHEN 'issued' THEN '已领用'
		           WHEN 'returned' THEN '已退回'
		           WHEN 'scrapped' THEN '已报废'
		           ELSE sc.status
		       END,
		       COALESCE(w.warehouse_name, ''),
		       true AS selected,
		       sc.created_at
		FROM target t
		INNER JOIN stock_out_item_serial_selection sel
		        ON sel.stock_out_item_id = $1
		INNER JOIN sku_serial_code sc ON sc.id = sel.serial_code_id
		LEFT JOIN warehouse w ON w.id = sc.warehouse_id
		WHERE t.stock_out_status <> 'confirmed'

		UNION ALL

		SELECT sc.id, sc.serial_code, sc.material_id, sc.material_code,
		       sc.material_name, sc.status,
		       CASE sc.status
		           WHEN 'in_stock' THEN '在库'
		           WHEN 'issued' THEN '已领用'
		           WHEN 'returned' THEN '已退回'
		           WHEN 'scrapped' THEN '已报废'
		           ELSE sc.status
		       END,
		       COALESCE(w.warehouse_name, ''),
		       true AS selected,
		       tr.created_at
		FROM target t
		INNER JOIN sku_serial_trace tr
		        ON tr.ref_doc_type = 'stock_out'
		       AND tr.ref_doc_id = t.stock_out_id
		       AND tr.action = 'stock_out'
		INNER JOIN sku_serial_code sc ON sc.id = tr.serial_code_id
		LEFT JOIN warehouse w ON w.id = sc.warehouse_id
		WHERE t.stock_out_status = 'confirmed'
		  AND sc.material_id = t.material_id
		  AND sc.inventory_id = t.inventory_id
		ORDER BY serial_code ASC
	`
	rows, err := database.Pool.Query(ctx, sql, stockOutItemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]SkuSerialCodeItem, 0)
	for rows.Next() {
		var item SkuSerialCodeItem
		err := rows.Scan(
			&item.ID, &item.SerialCode, &item.MaterialID, &item.MaterialCode,
			&item.MaterialName, &item.Status, &item.StatusLabel,
			&item.WarehouseName, &item.Selected, &item.CreatedAt,
		)
		if err != nil {
			return items, err
		}
		items = append(items, item)
	}
	return items, nil
}

func QueryAvailableSerialCodesByStockOutItem(ctx context.Context, stockOutItemID int64) ([]SkuSerialCodeItem, error) {
	sql := `
		SELECT sc.id, sc.serial_code, sc.material_id, sc.material_code,
		       sc.material_name, sc.status,
		       CASE sc.status
		           WHEN 'in_stock' THEN '在库'
		           WHEN 'issued' THEN '已领用'
		           WHEN 'returned' THEN '已退回'
		           WHEN 'scrapped' THEN '已报废'
		           ELSE sc.status
		       END,
		       COALESCE(w.warehouse_name, ''),
		       EXISTS (
		           SELECT 1
		           FROM stock_out_item_serial_selection self
		           WHERE self.stock_out_item_id = $1
		             AND self.serial_code_id = sc.id
		       ) AS selected,
		       sc.created_at
		FROM sku_serial_code sc
		INNER JOIN stock_out_item soi ON soi.id = $1 AND soi.inventory_id = sc.inventory_id
		INNER JOIN stock_out so ON so.id = soi.stock_out_id AND so.deleted_at IS NULL
		LEFT JOIN warehouse w ON w.id = sc.warehouse_id
		WHERE sc.status = 'in_stock'
		  AND (
		      EXISTS (
		          SELECT 1
		          FROM stock_out_item_serial_selection self
		          WHERE self.stock_out_item_id = $1
		            AND self.serial_code_id = sc.id
		      )
		      OR NOT EXISTS (
		          SELECT 1
		          FROM stock_out_item_serial_selection other_sel
		          INNER JOIN stock_out_item other_soi ON other_soi.id = other_sel.stock_out_item_id
		          INNER JOIN stock_out other_so ON other_so.id = other_soi.stock_out_id
		          WHERE other_sel.serial_code_id = sc.id
		            AND other_so.id <> so.id
		            AND other_so.deleted_at IS NULL
		            AND other_so.status IN ('draft', 'pending')
		      )
		  )
		ORDER BY sc.serial_code ASC
	`

	rows, err := database.Pool.Query(ctx, sql, stockOutItemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]SkuSerialCodeItem, 0)
	for rows.Next() {
		var item SkuSerialCodeItem
		err := rows.Scan(
			&item.ID, &item.SerialCode, &item.MaterialID, &item.MaterialCode,
			&item.MaterialName, &item.Status, &item.StatusLabel,
			&item.WarehouseName, &item.Selected, &item.CreatedAt,
		)
		if err != nil {
			return items, err
		}
		items = append(items, item)
	}
	return items, nil
}

// QueryAvailableIssuedSerialCodesByStockInItem 退料入库备货：查询可用“已出库(issued)”编码
// 规则：
// - sku_serial_code.status = 'issued'
// - 编码必须来自“领料出库(consumption_order)”且出库单已完成
// - 编码不能被其它“待入库/部分入库”的退料入库单占用（但允许当前 stock_in_item 已选中的编码返回）
func QueryAvailableIssuedSerialCodesByStockInItem(ctx context.Context, stockInItemID int64) ([]SkuSerialCodeItem, error) {
	sql := `
		WITH target AS (
			SELECT sii.id AS stock_in_item_id, sii.stock_in_id, sii.material_id
			FROM stock_in_item sii
			WHERE sii.id = $1
		),
		cur_si AS (
			SELECT si.id, si.stock_in_status
			FROM stock_in si
			INNER JOIN target t ON t.stock_in_id = si.id
			WHERE si.deleted_at IS NULL
		),
		issue_time AS (
			SELECT tr.serial_code_id, MAX(tr.created_at) AS issued_at
			FROM sku_serial_trace tr
			INNER JOIN stock_out so ON so.id = tr.ref_doc_id AND so.deleted_at IS NULL
			WHERE tr.action = 'stock_out'
			  AND tr.ref_doc_type = 'stock_out'
			  AND so.status = 'confirmed'
			  AND so.ref_doc_type = 'consumption_order'
			GROUP BY tr.serial_code_id
		)
		SELECT sc.id, sc.serial_code, sc.material_id, sc.material_code,
		       sc.material_name, sc.status,
		       CASE sc.status
		           WHEN 'in_stock' THEN '在库'
		           WHEN 'issued' THEN '已领用'
		           WHEN 'returned' THEN '已退回'
		           WHEN 'scrapped' THEN '已报废'
		           ELSE sc.status
		       END,
		       COALESCE(w.warehouse_name, ''),
		       EXISTS (
		           SELECT 1
		           FROM stock_in_item_serial_selection self
		           WHERE self.stock_in_item_id = $1
		             AND self.serial_code_id = sc.id
		       ) AS selected,
		       sc.created_at
		FROM target t
		INNER JOIN sku_serial_code sc ON sc.material_id = t.material_id
		INNER JOIN issue_time it ON it.serial_code_id = sc.id
		LEFT JOIN warehouse w ON w.id = sc.warehouse_id
		WHERE sc.status = 'issued'
		  AND EXISTS (
			SELECT 1
			FROM sku_serial_trace tr
			INNER JOIN stock_out so ON so.id = tr.ref_doc_id AND so.deleted_at IS NULL
			WHERE tr.serial_code_id = sc.id
			  AND tr.action = 'stock_out'
			  AND tr.ref_doc_type = 'stock_out'
			  AND so.status = 'confirmed'
			  AND so.ref_doc_type = 'consumption_order'
		  )
		  AND (
		      EXISTS (
		          SELECT 1
		          FROM stock_in_item_serial_selection self
		          WHERE self.stock_in_item_id = $1
		            AND self.serial_code_id = sc.id
		      )
		      OR NOT EXISTS (
		          SELECT 1
		          FROM stock_in_item_serial_selection other_sel
		          INNER JOIN stock_in other_si ON other_si.id = other_sel.stock_in_id
		          WHERE other_sel.serial_code_id = sc.id
		            AND other_sel.stock_in_item_id <> $1
		            AND other_si.deleted_at IS NULL
		            AND other_si.stock_in_status IN ('preparing', 'pending')
		      )
		  )
		ORDER BY it.issued_at ASC, sc.serial_code ASC
	`
	rows, err := database.Pool.Query(ctx, sql, stockInItemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]SkuSerialCodeItem, 0)
	for rows.Next() {
		var item SkuSerialCodeItem
		if err := rows.Scan(
			&item.ID, &item.SerialCode, &item.MaterialID, &item.MaterialCode,
			&item.MaterialName, &item.Status, &item.StatusLabel,
			&item.WarehouseName, &item.Selected, &item.CreatedAt,
		); err != nil {
			return items, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
