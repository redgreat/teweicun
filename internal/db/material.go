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

	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

func formatAttrItems(attrs []response.CustomAttributeItem, pairSep string) []string {
	items := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		k := strings.TrimSpace(attr.AttrName)
		v := strings.TrimSpace(attr.AttrValue)
		if k == "" && v == "" {
			continue
		}
		if k == "" {
			items = append(items, v)
			continue
		}
		if v == "" {
			items = append(items, k)
			continue
		}
		items = append(items, k+pairSep+v)
	}
	return items
}

func buildStoredMaterialName(materialName string, attrs []response.CustomAttributeItem) string {
	base := strings.TrimSpace(materialName)
	if base == "" {
		return ""
	}
	items := formatAttrItems(attrs, "：")
	if len(items) == 0 {
		return base
	}
	return base + "（" + strings.Join(items, "，") + "）"
}

func extractMaterialBaseName(materialName string, attrs []response.CustomAttributeItem) string {
	name := strings.TrimSpace(materialName)
	if name == "" {
		return ""
	}

	// 新规范：全中文标点，如 钢板（材质：Q235B，长度：200mm）
	newItems := formatAttrItems(attrs, "：")
	if len(newItems) > 0 {
		newSuffix := "（" + strings.Join(newItems, "，") + "）"
		if strings.HasSuffix(name, newSuffix) {
			return strings.TrimSpace(strings.TrimSuffix(name, newSuffix))
		}
	}
	return name
}

func mapReqAttrs(attrs []request.CustomAttributeItem) []response.CustomAttributeItem {
	out := make([]response.CustomAttributeItem, 0, len(attrs))
	for _, a := range attrs {
		out = append(out, response.CustomAttributeItem{
			AttrName:  a.AttrName,
			AttrValue: a.AttrValue,
		})
	}
	return out
}

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
		where = append(where, fmt.Sprintf("(material_name ILIKE $%d OR custom_attributes::text ILIKE $%d)", argID, argID))
		args = append(args, "%"+q.MaterialName+"%")
		argID++
	}
	if q.CategoryID != 0 {
		where = append(where, fmt.Sprintf("category_id = $%d", argID))
		args = append(args, q.CategoryID)
		argID++
	}
	if q.Unit != "" {
		where = append(where, fmt.Sprintf("unit = $%d", argID))
		args = append(args, q.Unit)
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
		       is_code, custom_attributes, default_warehouse_id, default_warehouse_name,
		       status, status_name, remark, created_at, updated_at
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
			&item.IsCode, &customAttrsJSON, &item.DefaultWarehouseID, &item.DefaultWarehouseName,
			&item.Status, &item.StatusName, &item.Remark, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}

		if len(customAttrsJSON) > 0 {
			json.Unmarshal(customAttrsJSON, &item.CustomAttributes)
		}
		item.MaterialNameBase = extractMaterialBaseName(item.MaterialName, item.CustomAttributes)
		item.MaterialDisplayName = item.MaterialName

		result = append(result, item)
	}

	return result, total, rows.Err()
}

func GetMaterial(ctx context.Context, id int64) (*response.MaterialResp, error) {
	query := `
		SELECT id, category_id, category_name, material_code, material_name,
		       unit, unit_name, safety_stock, max_stock,
		       is_code, custom_attributes, default_warehouse_id, default_warehouse_name,
		       status, status_name, remark, created_at, updated_at
		FROM v_material_list
		WHERE id = $1
		LIMIT 1
	`
	var item response.MaterialResp
	var customAttrsJSON []byte
	err := database.Pool.QueryRow(ctx, query, id).Scan(
		&item.ID, &item.CategoryID, &item.CategoryName, &item.MaterialCode, &item.MaterialName,
		&item.Unit, &item.UnitName, &item.SafetyStock, &item.MaxStock,
		&item.IsCode, &customAttrsJSON, &item.DefaultWarehouseID, &item.DefaultWarehouseName,
		&item.Status, &item.StatusName, &item.Remark, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if len(customAttrsJSON) > 0 {
		json.Unmarshal(customAttrsJSON, &item.CustomAttributes)
	}
	item.MaterialNameBase = extractMaterialBaseName(item.MaterialName, item.CustomAttributes)
	item.MaterialDisplayName = item.MaterialName
	return &item, nil
}

// CreateMaterial inserts a new material
func CreateMaterial(ctx context.Context, req *request.CreateMaterialReq, userID int64) (int64, error) {
	customAttrsJSON, _ := json.Marshal(req.CustomAttributes)
	storedMaterialName := buildStoredMaterialName(req.MaterialName, mapReqAttrs(req.CustomAttributes))

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
		                      safety_stock, max_stock, is_code, custom_attributes, remark, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	var id int64
	err = tx.QueryRow(ctx, query,
		req.CategoryID, materialCode, storedMaterialName, req.Unit,
		req.SafetyStock, req.MaxStock, req.IsCode, customAttrsJSON, req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
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
	storedMaterialName := buildStoredMaterialName(req.MaterialName, mapReqAttrs(req.CustomAttributes))

	query := `
		UPDATE material
		SET category_id = $1, material_code = $2, material_name = $3, unit = $4,
		    safety_stock = $5, max_stock = $6, is_code = $7,
		    custom_attributes = $8, remark = $9, status = $10,
		    updated_by = $11, updated_at = NOW()
		WHERE id = $12 AND deleted_at IS NULL
	`
	_, err := database.Pool.Exec(ctx, query,
		req.CategoryID, req.MaterialCode, storedMaterialName, req.Unit,
		req.SafetyStock, req.MaxStock, req.IsCode,
		customAttrsJSON, req.Remark, req.Status, userID, id)
	if err != nil {
		return err
	}
	return nil
}
