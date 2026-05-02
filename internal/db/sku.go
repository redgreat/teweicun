/**
 * 功能：SKU数据库操作
 * 创建时间：2026-04-18
 * 创建人：CodeArts Agent
 * 修改时间：2026-04-19
 * 修改内容：使用pkg/database包、v_sku_list视图、fn_generate_sku_code函数
 */

package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

// ListSKUs 分页查询SKU列表
func ListSKUs(ctx context.Context, q *request.SKUQuery) ([]response.SKUListItem, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argID := 1

	if q.MaterialID != 0 {
		where = append(where, fmt.Sprintf("material_id = $%d", argID))
		args = append(args, q.MaterialID)
		argID++
	}
	if q.SKUCode != "" {
		where = append(where, fmt.Sprintf("sku_code ILIKE $%d", argID))
		args = append(args, "%"+q.SKUCode+"%")
		argID++
	}
	if q.SKUName != "" {
		where = append(where, fmt.Sprintf("(sku_name ILIKE $%d OR attr_summary ILIKE $%d)", argID, argID))
		args = append(args, "%"+q.SKUName+"%")
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_sku_list WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT v.id, v.material_id, v.material_code, v.material_name,
		       v.unit, v.reference_price, v.category_name,
		       v.sku_code, v.sku_name, v.custom_attributes,
		       v.attr_summary, v.status, v.status_name, v.remark, v.created_at
		FROM v_sku_list v
		WHERE %s
		ORDER BY v.material_code, v.sku_code
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)
	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []response.SKUListItem
	for rows.Next() {
		var item response.SKUListItem
		var customAttrsJSON []byte
		err := rows.Scan(
			&item.ID, &item.MaterialID, &item.MaterialCode, &item.MaterialName,
			&item.Unit, &item.ReferencePrice, &item.CategoryName,
			&item.SKUCode, &item.SKUName, &customAttrsJSON,
			&item.AttrSummary, &item.Status, &item.StatusName, &item.Remark, &item.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if len(customAttrsJSON) > 0 {
			json.Unmarshal(customAttrsJSON, &item.CustomAttributes)
		}
		list = append(list, item)
	}

	return list, total, rows.Err()
}

// GetSKUByID 获取SKU详情
func GetSKUByID(ctx context.Context, id int64) (*response.SKUDetail, error) {
	query := `
		SELECT v.id, v.material_id, v.material_code, v.material_name,
		       v.unit, v.reference_price, v.category_name,
		       v.sku_code, v.sku_name, v.custom_attributes,
		       v.attr_summary, v.status, v.status_name, v.remark, v.created_at, v.updated_at
		FROM v_sku_list v
		WHERE v.id = $1
	`
	var detail response.SKUDetail
	var customAttrsJSON []byte
	err := database.Pool.QueryRow(ctx, query, id).Scan(
		&detail.ID, &detail.MaterialID, &detail.MaterialCode, &detail.MaterialName,
		&detail.Unit, &detail.ReferencePrice, &detail.CategoryName,
		&detail.SKUCode, &detail.SKUName, &customAttrsJSON,
		&detail.AttrSummary, &detail.Status, &detail.StatusName, &detail.Remark,
		&detail.CreatedAt, &detail.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if len(customAttrsJSON) > 0 {
		json.Unmarshal(customAttrsJSON, &detail.CustomAttributes)
	}
	return &detail, nil
}

// CreateSKU 创建SKU（编码由数据库函数自动生成）
func CreateSKU(ctx context.Context, req *request.CreateSKUReq, userID int64) (int64, error) {
	// 检查物料是否启用了SKU管理
	var skuManaged bool
	err := database.Pool.QueryRow(ctx, `
		SELECT COALESCE(sku_managed, false) FROM material WHERE id = $1 AND deleted_at IS NULL
	`, req.MaterialID).Scan(&skuManaged)
	if err != nil {
		return 0, fmt.Errorf("物料不存在: %w", err)
	}
	if !skuManaged {
		return 0, fmt.Errorf("该物料未启用SKU管理，请先在物料设置中开启")
	}

	// 使用数据库函数生成SKU编码
	var skuCode string
	err = database.Pool.QueryRow(ctx, `SELECT fn_generate_sku_code($1)`, req.MaterialID).Scan(&skuCode)
	if err != nil {
		return 0, fmt.Errorf("生成SKU编码失败: %w", err)
	}

	// 自动生成SKU名称（如果未提供）
	skuName := req.SKUName
	if skuName == "" {
		// 用属性值组合生成名称
		var parts []string
		for _, attr := range req.CustomAttributes {
			if attr.AttrValue != "" {
				val := attr.AttrValue
				if attr.AttrUnit != "" {
					val += attr.AttrUnit
				}
				parts = append(parts, val)
			}
		}
		if len(parts) > 0 {
			skuName = strings.Join(parts, " / ")
		}
	}

	customAttrsJSON, _ := json.Marshal(req.CustomAttributes)

	query := `
		INSERT INTO material_sku (material_id, sku_code, sku_name, reference_price, custom_attributes, remark, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	var id int64
	err = database.Pool.QueryRow(ctx, query,
		req.MaterialID, skuCode, skuName, req.ReferencePrice, customAttrsJSON, req.Remark, userID).Scan(&id)
	return id, err
}

// UpdateSKU 更新SKU
func UpdateSKU(ctx context.Context, id int64, req *request.UpdateSKUReq, userID int64) error {
	// 检查是否有关联的库存记录（如果有，不允许修改属性）
	var invCount int64
	err := database.Pool.QueryRow(ctx, `
		SELECT count(*) FROM inventory WHERE sku_id = $1 AND quantity > 0
	`, id).Scan(&invCount)
	if err != nil {
		return err
	}

	customAttrsJSON, _ := json.Marshal(req.CustomAttributes)

	if req.Status == "disabled" {
		// 检查是否有未完成的单据
		if err := checkSKUActiveDocuments(ctx, id); err != nil {
			return err
		}
	}

	if invCount > 0 {
		// 有库存记录，只允许修改名称、备注和状态
		query := `
			UPDATE material_sku
			SET sku_name = $1, reference_price = $2, remark = $3, status = $4,
			    updated_by = $5, updated_at = NOW()
			WHERE id = $6 AND deleted_at IS NULL
		`
		_, err = database.Pool.Exec(ctx, query,
			req.SKUName, req.ReferencePrice, req.Remark, req.Status, userID, id)
	} else {
		// 无库存记录，允许修改所有字段
		query := `
			UPDATE material_sku
			SET sku_name = $1, reference_price = $2, custom_attributes = $3, remark = $4, status = $5,
			    updated_by = $6, updated_at = NOW()
			WHERE id = $7 AND deleted_at IS NULL
		`
		_, err = database.Pool.Exec(ctx, query,
			req.SKUName, req.ReferencePrice, customAttrsJSON, req.Remark, req.Status, userID, id)
	}
	return err
}

// checkSKUActiveDocuments 检查SKU是否有未完成的业务单据
func checkSKUActiveDocuments(ctx context.Context, id int64) error {
	// 1. 检查是否有未完成的采购订单
	var poCount int64
	err := database.Pool.QueryRow(ctx, `
		SELECT count(*) FROM purchase_order_item poi
		JOIN purchase_order po ON po.id = poi.order_id
		WHERE poi.sku_id = $1 AND po.order_status NOT IN ('closed', 'full_received')
		AND po.deleted_at IS NULL
	`, id).Scan(&poCount)
	if err != nil {
		return err
	}
	if poCount > 0 {
		return fmt.Errorf("该SKU正处于 %d 个执行中的采购订单中，无法禁用", poCount)
	}

	// 2. 检查是否有未确认的入库单
	var siCount int64
	err = database.Pool.QueryRow(ctx, `
		SELECT count(*) FROM stock_in_item sii
		JOIN stock_in si ON si.id = sii.stock_in_id
		WHERE sii.sku_id = $1 AND si.stock_in_status = 'pending'
		AND si.deleted_at IS NULL
	`, id).Scan(&siCount)
	if err != nil {
		return err
	}
	if siCount > 0 {
		return fmt.Errorf("该SKU有 %d 个待质检的入库单，无法禁用", siCount)
	}

	// 3. 检查是否有未确认的出库单
	var soCount int64
	err = database.Pool.QueryRow(ctx, `
		SELECT count(*) FROM stock_out_item soi
		JOIN stock_out so ON so.id = soi.stock_out_id
		WHERE soi.sku_id = $1 AND so.status = 'draft'
		AND so.deleted_at IS NULL
	`, id).Scan(&soCount)
	if err != nil {
		return err
	}
	if soCount > 0 {
		return fmt.Errorf("该SKU有 %d 个未确认的出库单，无法禁用", soCount)
	}

	return nil
}

func DeleteSKU(ctx context.Context, id int64, userID int64) error {
	// 1. 检查是否有历史采购记录
	var hasHistory bool
	err := database.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM purchase_order_item WHERE sku_id = $1)
	`, id).Scan(&hasHistory)
	if err != nil {
		return err
	}
	if hasHistory {
		return fmt.Errorf("该SKU已有采购记录，不可删除，建议停用")
	}

	// 2. 检查是否有库存
	var invCount int64
	err = database.Pool.QueryRow(ctx, `
		SELECT count(*) FROM inventory WHERE sku_id = $1 AND (quantity > 0 OR locked_quantity > 0)
	`, id).Scan(&invCount)
	if err != nil {
		return err
	}
	if invCount > 0 {
		return fmt.Errorf("该SKU仍有库存或被锁定，无法删除")
	}

	query := `UPDATE material_sku SET deleted_at = NOW(), updated_by = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err = database.Pool.Exec(ctx, query, userID, id)
	return err
}

// ListSKUsByMaterial 获取指定物料下的所有启用SKU（用于采购/入库时选择）
func ListSKUsByMaterial(ctx context.Context, materialID int64) ([]response.SKUSelectItem, error) {
	query := `
		SELECT id, sku_code, sku_name, reference_price, custom_attributes,
		       (
		           SELECT string_agg(
		               (attr->>'attr_name') || ':' || (attr->>'attr_value') ||
		               CASE WHEN (attr->>'attr_unit') IS NOT NULL AND (attr->>'attr_unit') != ''
		                    THEN (attr->>'attr_unit') ELSE '' END,
		               ' | '
		           )
		           FROM jsonb_array_elements(custom_attributes) AS attr
		       ) AS attr_summary
		FROM material_sku
		WHERE material_id = $1 AND deleted_at IS NULL AND status = 'enabled'
		ORDER BY sku_code
	`
	rows, err := database.Pool.Query(ctx, query, materialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []response.SKUSelectItem
	for rows.Next() {
		var item response.SKUSelectItem
		var customAttrsJSON []byte
		err := rows.Scan(&item.ID, &item.SKUCode, &item.SKUName, &item.ReferencePrice, &customAttrsJSON, &item.AttrSummary)
		if err != nil {
			return nil, err
		}
		if len(customAttrsJSON) > 0 {
			json.Unmarshal(customAttrsJSON, &item.CustomAttributes)
		}
		list = append(list, item)
	}
	return list, rows.Err()
}
