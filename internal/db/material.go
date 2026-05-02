/**
 * 功能：material.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

// ListMaterials fetches materials with pagination and filters
func ListMaterials(ctx context.Context, q *request.MaterialQuery) ([]response.MaterialResp, int64, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if q.MaterialCode != "" {
		where = append(where, fmt.Sprintf("material_code ILIKE $%d", argID))
		args = append(args, "%"+q.MaterialCode+"%")
		argID++
	}
	if q.MaterialName != "" {
		where = append(where, fmt.Sprintf("material_name ILIKE $%d", argID))
		args = append(args, "%"+q.MaterialName+"%")
		argID++
	}
	if q.CategoryID != 0 {
		where = append(where, fmt.Sprintf("category_id = $%d", argID))
		args = append(args, q.CategoryID)
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_material_list WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, category_id, category_name, material_code, material_name,
		       unit, unit_name, safety_stock, max_stock,
		       is_code, sku_managed, custom_attributes, default_warehouse_id, default_warehouse_name,
		       status, status_name, remark, created_at, updated_at, sku_count
		FROM v_material_list
		WHERE %s
		ORDER BY id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.MaterialResp
	for rows.Next() {
		var item response.MaterialResp
		var customAttrsJSON []byte
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.CategoryName, &item.MaterialCode, &item.MaterialName,
			&item.Unit, &item.UnitName, &item.SafetyStock, &item.MaxStock,
			&item.IsCode, &item.SkuManaged, &customAttrsJSON, &item.DefaultWarehouseID, &item.DefaultWarehouseName,
			&item.Status, &item.StatusName, &item.Remark, &item.CreatedAt, &item.UpdatedAt, &item.SkuCount); err != nil {
			return nil, 0, err
		}

		if len(customAttrsJSON) > 0 {
			json.Unmarshal(customAttrsJSON, &item.CustomAttributes)
		}

		result = append(result, item)
	}

	return result, total, rows.Err()
}

// CreateMaterial inserts a new material
func CreateMaterial(ctx context.Context, req *request.CreateMaterialReq, userID int64) (int64, error) {
	customAttrsJSON, _ := json.Marshal(req.CustomAttributes)

	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var materialCode string
	err = tx.QueryRow(ctx, "SELECT fn_generate_material_code()").Scan(&materialCode)
	if err != nil {
		return 0, fmt.Errorf("生成物料编码失败: %w", err)
	}

	query := `
		INSERT INTO material (category_id, material_code, material_name, unit,
		                      safety_stock, max_stock, is_code, sku_managed, custom_attributes, remark, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	var id int64
	err = tx.QueryRow(ctx, query,
		req.CategoryID, materialCode, req.MaterialName, req.Unit,
		req.SafetyStock, req.MaxStock, req.IsCode, true, customAttrsJSON, req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
	}

	if len(req.CustomAttributes) == 0 {
		var skuCode string
		err = tx.QueryRow(ctx, `SELECT fn_generate_sku_code($1)`, id).Scan(&skuCode)
		if err != nil {
			return 0, fmt.Errorf("生成默认SKU编码失败: %w", err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO material_sku (material_id, sku_code, sku_name, custom_attributes, created_by)
			VALUES ($1, $2, $3, '[]'::jsonb, $4)
		`, id, skuCode, req.MaterialName, userID)
		if err != nil {
			return 0, fmt.Errorf("生成默认SKU失败: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateMaterial updates an existing material
func UpdateMaterial(ctx context.Context, id int64, req *request.UpdateMaterialReq, userID int64) error {
	if req.Status == "disabled" {
		var invCount int64
		err := database.Pool.QueryRow(ctx, `SELECT count(*) FROM inventory WHERE material_id = $1 AND quantity > 0`, id).Scan(&invCount)
		if err != nil {
			return err
		}
		if invCount > 0 {
			return fmt.Errorf("该物料有 %d 条库存记录，无法禁用", invCount)
		}
		var poCount int64
		err = database.Pool.QueryRow(ctx, `
			SELECT count(*) FROM purchase_order_item poi
			JOIN purchase_order po ON po.id = poi.order_id
			WHERE poi.material_id = $1 AND po.status NOT IN ('closed', 'cancelled') AND po.deleted_at IS NULL
		`, id).Scan(&poCount)
		if err != nil {
			return err
		}
		if poCount > 0 {
			return fmt.Errorf("该物料有 %d 个进行中的采购订单，无法禁用", poCount)
		}
	}

	customAttrsJSON, _ := json.Marshal(req.CustomAttributes)

	query := `
		UPDATE material
		SET category_id = $1, material_code = $2, material_name = $3, unit = $4,
		    safety_stock = $5, max_stock = $6, is_code = $7, sku_managed = $8,
		    custom_attributes = $9, remark = $10, status = $11,
		    updated_by = $12, updated_at = NOW()
		WHERE id = $13 AND deleted_at IS NULL
	`
	_, err := database.Pool.Exec(ctx, query,
		req.CategoryID, req.MaterialCode, req.MaterialName, req.Unit,
		req.SafetyStock, req.MaxStock, req.IsCode, true,
		customAttrsJSON, req.Remark, req.Status, userID, id)
	if err != nil {
		return err
	}

	if len(req.CustomAttributes) == 0 {
		var cnt int64
		err := database.Pool.QueryRow(ctx, `
			SELECT count(*) FROM material_sku WHERE material_id = $1 AND deleted_at IS NULL
		`, id).Scan(&cnt)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil
			}
			return err
		}
		if cnt == 0 {
			var skuCode string
			err := database.Pool.QueryRow(ctx, `SELECT fn_generate_sku_code($1)`, id).Scan(&skuCode)
			if err != nil {
				return fmt.Errorf("生成默认SKU编码失败: %w", err)
			}
			_, err = database.Pool.Exec(ctx, `
				INSERT INTO material_sku (material_id, sku_code, sku_name, custom_attributes, created_by)
				VALUES ($1, $2, $3, '[]'::jsonb, $4)
			`, id, skuCode, req.MaterialName, userID)
			if err != nil {
				return fmt.Errorf("生成默认SKU失败: %w", err)
			}
		}
	}
	return nil
}

// DeleteMaterial soft deletes a material
func DeleteMaterial(ctx context.Context, id int64, userID int64) error {
	var poCount int64
	err := database.Pool.QueryRow(ctx, `SELECT count(*) FROM purchase_order_item WHERE material_id = $1`, id).Scan(&poCount)
	if err != nil {
		return err
	}
	if poCount > 0 {
		return fmt.Errorf("该物料已被 %d 个采购订单引用，无法删除", poCount)
	}
	var soCount int64
	err = database.Pool.QueryRow(ctx, `SELECT count(*) FROM sales_order_item WHERE material_id = $1`, id).Scan(&soCount)
	if err != nil {
		return err
	}
	if soCount > 0 {
		return fmt.Errorf("该物料已被 %d 个销售订单引用，无法删除", soCount)
	}
	var siCount int64
	err = database.Pool.QueryRow(ctx, `SELECT count(*) FROM stock_in_item WHERE material_id = $1`, id).Scan(&siCount)
	if err != nil {
		return err
	}
	if siCount > 0 {
		return fmt.Errorf("该物料已被 %d 个入库单引用，无法删除", siCount)
	}

	query := `
		UPDATE material
		SET deleted_at = NOW(), updated_by = $1
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err = database.Pool.Exec(ctx, query, userID, id)
	return err
}
