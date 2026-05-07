package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

func resolveDashboardRange(rangeKey string, now time.Time) (time.Time, string) {
	switch rangeKey {
	case "7d":
		return now.AddDate(0, 0, -6), "7d"
	case "mtd":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), "mtd"
	default:
		return now.AddDate(0, 0, -29), "30d"
	}
}

func QueryDashboardBigscreen(ctx context.Context, rangeKey string) (*response.DashboardBigscreenResp, error) {
	now := time.Now()
	startDate, normalizedRange := resolveDashboardRange(rangeKey, now)

	out := &response.DashboardBigscreenResp{
		Range:     normalizedRange,
		UpdatedAt: now.Format(time.RFC3339),
	}

	kpiSQL := `
		WITH p AS (
			SELECT
				COALESCE(SUM(poi.quantity), 0)::float8 AS qty,
				COALESCE(SUM(COALESCE(poi.amount, poi.quantity * poi.unit_price)), 0)::float8 AS amount
			FROM purchase_order po
			INNER JOIN purchase_order_item poi ON poi.order_id = po.id
			WHERE po.deleted_at IS NULL
			  AND po.order_status <> 'draft'
			  AND po.order_date::date >= $1::date
		),
		c AS (
			SELECT
				COALESCE(SUM(coi.quantity), 0)::float8 AS qty,
				COALESCE(SUM(coi.quantity * COALESCE(inv.unit_cost, 0)), 0)::float8 AS amount
			FROM consumption_order co
			INNER JOIN consumption_order_item coi ON coi.order_id = co.id
			LEFT JOIN inventory inv ON inv.id = coi.inventory_id
			WHERE co.deleted_at IS NULL
			  AND co.status IN ('confirmed', 'completed')
			  AND co.order_date::date >= $1::date
		),
		r AS (
			SELECT
				COALESCE(SUM(roi.quantity * COALESCE(inv.unit_cost, 0)), 0)::float8 AS amount
			FROM reversal_order ro
			INNER JOIN reversal_order_item roi ON roi.order_id = ro.id
			LEFT JOIN inventory inv ON inv.id = roi.inventory_id
			WHERE ro.deleted_at IS NULL
			  AND ro.status IN ('confirmed', 'completed')
			  AND ro.order_date::date >= $1::date
		),
		i AS (
			SELECT COALESCE(SUM(quantity * COALESCE(unit_cost, 0)), 0)::float8 AS amount
			FROM inventory
		)
		SELECT p.qty, p.amount, c.qty, c.amount, r.amount,
		       (c.amount - r.amount) AS net_consumption_amount,
		       i.amount
		FROM p, c, r, i
	`
	if err := database.Pool.QueryRow(ctx, kpiSQL, startDate).Scan(
		&out.KPI.PurchaseQty,
		&out.KPI.PurchaseAmount,
		&out.KPI.ConsumptionQty,
		&out.KPI.ConsumptionAmount,
		&out.KPI.ReversalAmount,
		&out.KPI.NetConsumptionAmount,
		&out.KPI.InventoryAmount,
	); err != nil {
		return nil, err
	}

	trendSQL := `
		WITH days AS (
			SELECT generate_series($1::date, $2::date, interval '1 day')::date AS d
		),
		p AS (
			SELECT
				po.order_date::date AS d,
				COALESCE(SUM(poi.quantity), 0)::float8 AS qty,
				COALESCE(SUM(COALESCE(poi.amount, poi.quantity * poi.unit_price)), 0)::float8 AS amount
			FROM purchase_order po
			INNER JOIN purchase_order_item poi ON poi.order_id = po.id
			WHERE po.deleted_at IS NULL
			  AND po.order_status <> 'draft'
			  AND po.order_date::date BETWEEN $1::date AND $2::date
			GROUP BY po.order_date::date
		),
		c AS (
			SELECT
				co.order_date::date AS d,
				COALESCE(SUM(coi.quantity), 0)::float8 AS qty,
				COALESCE(SUM(coi.quantity * COALESCE(inv.unit_cost, 0)), 0)::float8 AS amount
			FROM consumption_order co
			INNER JOIN consumption_order_item coi ON coi.order_id = co.id
			LEFT JOIN inventory inv ON inv.id = coi.inventory_id
			WHERE co.deleted_at IS NULL
			  AND co.status IN ('confirmed', 'completed')
			  AND co.order_date::date BETWEEN $1::date AND $2::date
			GROUP BY co.order_date::date
		)
		SELECT to_char(days.d, 'YYYY-MM-DD'),
		       COALESCE(p.qty, 0), COALESCE(p.amount, 0),
		       COALESCE(c.qty, 0), COALESCE(c.amount, 0)
		FROM days
		LEFT JOIN p ON p.d = days.d
		LEFT JOIN c ON c.d = days.d
		ORDER BY days.d
	`
	trendRows, err := database.Pool.Query(ctx, trendSQL, startDate, now)
	if err != nil {
		return nil, err
	}
	defer trendRows.Close()
	for trendRows.Next() {
		var p response.DashboardTrendPoint
		if err := trendRows.Scan(&p.BizDate, &p.PurchaseQty, &p.PurchaseAmount, &p.ConsumptionQty, &p.ConsumptionAmount); err != nil {
			return nil, err
		}
		out.Trend = append(out.Trend, p)
	}
	if err := trendRows.Err(); err != nil {
		return nil, err
	}

	topPurchaseSQL := `
		SELECT
			poi.material_id,
			COALESCE(m.material_code, ''),
			COALESCE(m.material_name, ''),
			0::bigint AS sku_id,
			''::varchar AS sku_code,
			''::varchar AS sku_name,
			COALESCE(SUM(poi.quantity), 0)::float8,
			COALESCE(SUM(COALESCE(poi.amount, poi.quantity * poi.unit_price)), 0)::float8
		FROM purchase_order po
		INNER JOIN purchase_order_item poi ON poi.order_id = po.id
		LEFT JOIN material m ON m.id = poi.material_id
		WHERE po.deleted_at IS NULL
		  AND po.order_status <> 'draft'
		  AND po.order_date::date >= $1::date
		GROUP BY poi.material_id, m.material_code, m.material_name
		ORDER BY SUM(COALESCE(poi.amount, poi.quantity * poi.unit_price)) DESC
		LIMIT 8
	`
	if out.TopPurchaseAmount, err = queryDashboardTopItems(ctx, topPurchaseSQL, startDate); err != nil {
		return nil, err
	}

	topConsumptionAmountSQL := `
		SELECT
			coi.material_id,
			COALESCE(m.material_code, ''),
			COALESCE(m.material_name, ''),
			0::bigint AS sku_id,
			''::varchar AS sku_code,
			''::varchar AS sku_name,
			COALESCE(SUM(coi.quantity), 0)::float8,
			COALESCE(SUM(coi.quantity * COALESCE(inv.unit_cost, 0)), 0)::float8
		FROM consumption_order co
		INNER JOIN consumption_order_item coi ON coi.order_id = co.id
		LEFT JOIN inventory inv ON inv.id = coi.inventory_id
		LEFT JOIN material m ON m.id = coi.material_id
		WHERE co.deleted_at IS NULL
		  AND co.status IN ('confirmed', 'completed')
		  AND co.order_date::date >= $1::date
		GROUP BY coi.material_id, m.material_code, m.material_name
		ORDER BY SUM(coi.quantity * COALESCE(inv.unit_cost, 0)) DESC
		LIMIT 8
	`
	if out.TopConsumptionAmount, err = queryDashboardTopItems(ctx, topConsumptionAmountSQL, startDate); err != nil {
		return nil, err
	}

	topConsumptionQtySQL := `
		SELECT
			coi.material_id,
			COALESCE(m.material_code, ''),
			COALESCE(m.material_name, ''),
			0::bigint AS sku_id,
			''::varchar AS sku_code,
			''::varchar AS sku_name,
			COALESCE(SUM(coi.quantity), 0)::float8,
			COALESCE(SUM(coi.quantity * COALESCE(inv.unit_cost, 0)), 0)::float8
		FROM consumption_order co
		INNER JOIN consumption_order_item coi ON coi.order_id = co.id
		LEFT JOIN inventory inv ON inv.id = coi.inventory_id
		LEFT JOIN material m ON m.id = coi.material_id
		WHERE co.deleted_at IS NULL
		  AND co.status IN ('confirmed', 'completed')
		  AND co.order_date::date >= $1::date
		GROUP BY coi.material_id, m.material_code, m.material_name
		ORDER BY SUM(coi.quantity) DESC
		LIMIT 8
	`
	if out.TopConsumptionQty, err = queryDashboardTopItems(ctx, topConsumptionQtySQL, startDate); err != nil {
		return nil, err
	}

	summarySQL := `
		WITH p AS (
			SELECT COALESCE(SUM(COALESCE(poi.amount, poi.quantity * poi.unit_price)), 0)::float8 AS amount
			FROM purchase_order po
			INNER JOIN purchase_order_item poi ON poi.order_id = po.id
			WHERE po.deleted_at IS NULL
			  AND po.order_status <> 'draft'
			  AND po.order_date::date >= $1::date
		),
		c AS (
			SELECT COALESCE(SUM(coi.quantity * COALESCE(inv.unit_cost, 0)), 0)::float8 AS amount
			FROM consumption_order co
			INNER JOIN consumption_order_item coi ON coi.order_id = co.id
			LEFT JOIN inventory inv ON inv.id = coi.inventory_id
			WHERE co.deleted_at IS NULL
			  AND co.status IN ('confirmed', 'completed')
			  AND co.order_date::date >= $1::date
		),
		active_material AS (
			SELECT COALESCE(COUNT(DISTINCT coi.material_id), 0)::int8 AS cnt
			FROM consumption_order co
			INNER JOIN consumption_order_item coi ON coi.order_id = co.id
			WHERE co.deleted_at IS NULL
			  AND co.status IN ('confirmed', 'completed')
			  AND co.order_date::date >= $1::date
		),
		active_sku AS (
			SELECT 0::int8 AS cnt
		),
		max_single AS (
			SELECT COALESCE(MAX(t.amount), 0)::float8 AS amount
			FROM (
				SELECT SUM(coi.quantity * COALESCE(inv.unit_cost, 0)) AS amount
				FROM consumption_order co
				INNER JOIN consumption_order_item coi ON coi.order_id = co.id
				LEFT JOIN inventory inv ON inv.id = coi.inventory_id
				WHERE co.deleted_at IS NULL
				  AND co.status IN ('confirmed', 'completed')
				  AND co.order_date::date >= $1::date
				GROUP BY co.id
			) t
		)
		SELECT
			(p.amount - c.amount) AS purchase_minus_consumption_amount,
			active_material.cnt,
			active_sku.cnt,
			max_single.amount
		FROM p, c, active_material, active_sku, max_single
	`
	if err := database.Pool.QueryRow(ctx, summarySQL, startDate).Scan(
		&out.Summary.PurchaseMinusConsumptionAmount,
		&out.Summary.ActiveMaterialCount,
		&out.Summary.ActiveSKUCount,
		&out.Summary.MaxSingleConsumptionAmount,
	); err != nil {
		return nil, err
	}

	return out, nil
}

func queryDashboardTopItems(ctx context.Context, query string, startDate time.Time) ([]response.DashboardTopItem, error) {
	rows, err := database.Pool.Query(ctx, query, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]response.DashboardTopItem, 0)
	for rows.Next() {
		var item response.DashboardTopItem
		if err := rows.Scan(
			&item.MaterialID, &item.MaterialCode, &item.MaterialName,
			&item.SKUID, &item.SKUCode, &item.SKUName,
			&item.Quantity, &item.Amount,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func ValidateDashboardRange(rangeKey string) error {
	switch rangeKey {
	case "", "7d", "30d", "mtd":
		return nil
	default:
		return fmt.Errorf("invalid range")
	}
}
