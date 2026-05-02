/**
 * 功能：物料属性数据库操作
 * 创建时间：2026-04-17
 * 创建人：wangcw
 */

package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

func ListMaterialAttributeDefs(ctx context.Context, q *request.MaterialAttributeDefQuery) ([]response.MaterialAttributeDefResp, int64, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if q.AttrCode != "" {
		where = append(where, fmt.Sprintf("attr_code ILIKE $%d", argID))
		args = append(args, "%"+q.AttrCode+"%")
		argID++
	}
	if q.AttrName != "" {
		where = append(where, fmt.Sprintf("attr_name ILIKE $%d", argID))
		args = append(args, "%"+q.AttrName+"%")
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM material_attribute_def WHERE %s AND deleted_at IS NULL", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, attr_code, attr_name, attr_type, attr_unit, select_options, is_required, sort_order, remark, status, created_at, updated_at
		FROM material_attribute_def
		WHERE %s AND deleted_at IS NULL
		ORDER BY sort_order, id
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.MaterialAttributeDefResp
	for rows.Next() {
		var item response.MaterialAttributeDefResp
		if err := rows.Scan(&item.ID, &item.AttrCode, &item.AttrName, &item.AttrType, &item.AttrUnit, &item.SelectOptions,
			&item.IsRequired, &item.SortOrder, &item.Remark, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

func CreateMaterialAttributeDef(ctx context.Context, req *request.CreateMaterialAttributeDefReq, userID int64) (int64, error) {
	// 自动生成属性编码
	var attrCode string
	err := database.Pool.QueryRow(ctx, "SELECT fn_generate_base_code('AT')").Scan(&attrCode)
	if err != nil {
		return 0, fmt.Errorf("生成属性编码失败: %w", err)
	}

	query := `
		INSERT INTO material_attribute_def (attr_code, attr_name, attr_type, attr_unit, select_options, is_required, sort_order, remark, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	var id int64
	err = database.Pool.QueryRow(ctx, query,
		attrCode, req.AttrName, req.AttrType, req.AttrUnit, req.SelectOptions, req.IsRequired, req.SortOrder, req.Remark, userID).Scan(&id)
	return id, err
}

func UpdateMaterialAttributeDef(ctx context.Context, id int64, req *request.UpdateMaterialAttributeDefReq, userID int64) error {
	query := `
		UPDATE material_attribute_def
		SET attr_name = $1, attr_type = $2, attr_unit = $3, select_options = $4, is_required = $5, sort_order = $6, remark = $7, status = $8,
		    updated_by = $9, updated_at = NOW()
		WHERE id = $10 AND deleted_at IS NULL
	`
	_, err := database.Pool.Exec(ctx, query,
		req.AttrName, req.AttrType, req.AttrUnit, req.SelectOptions, req.IsRequired, req.SortOrder, req.Remark, req.Status, userID, id)
	return err
}

func DeleteMaterialAttributeDef(ctx context.Context, id int64, userID int64) error {
	var count int64
	err := database.Pool.QueryRow(ctx, `SELECT count(*) FROM material_attribute_value WHERE attr_id = $1 AND deleted_at IS NULL`, id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该属性已被 %d 个物料使用，无法删除", count)
	}

	query := `UPDATE material_attribute_def SET deleted_at = NOW(), updated_by = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err = database.Pool.Exec(ctx, query, userID, id)
	return err
}

func GetMaterialAttributes(ctx context.Context, materialID int64) ([]response.MaterialAttributeValueResp, error) {
	query := `
		SELECT mav.id, mav.material_id, mav.attr_id, mad.attr_code, mad.attr_name, mad.attr_type, mad.attr_unit,
		       mav.attr_value, mav.created_at, mav.updated_at
		FROM material_attribute_value mav
		JOIN material_attribute_def mad ON mav.attr_id = mad.id
		WHERE mav.material_id = $1 AND mav.deleted_at IS NULL AND mad.deleted_at IS NULL
		ORDER BY mad.sort_order, mad.id
	`

	rows, err := database.Pool.Query(ctx, query, materialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []response.MaterialAttributeValueResp
	for rows.Next() {
		var item response.MaterialAttributeValueResp
		if err := rows.Scan(&item.ID, &item.MaterialID, &item.AttrID, &item.AttrCode, &item.AttrName,
			&item.AttrType, &item.AttrUnit, &item.AttrValue, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, rows.Err()
}

func UpdateMaterialAttributes(ctx context.Context, materialID int64, req *request.UpdateMaterialAttributesReq, userID int64) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE material_attribute_value SET deleted_at = NOW(), updated_by = $1 WHERE material_id = $2`, userID, materialID)
	if err != nil {
		return err
	}

	for _, attr := range req.Attributes {
		query := `
			INSERT INTO material_attribute_value (material_id, attr_id, attr_value, created_by)
			VALUES ($1, $2, $3, $4)
		`
		_, err = tx.Exec(ctx, query, materialID, attr.AttrID, attr.AttrValue, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
