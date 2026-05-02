/**
 * 功能：dict.go
 * 创建时间：2026-04-18
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

// ========= Dict Type =========

// ListDictTypes 分页查询字典类型
func ListDictTypes(ctx context.Context, q *request.DictTypeQuery) ([]response.DictTypeResp, int64, error) {
	where := []string{"1=1"}
	var args []interface{}
	argID := 1

	if q.DictName != "" {
		where = append(where, fmt.Sprintf("dict_name ILIKE $%d", argID))
		args = append(args, "%"+q.DictName+"%")
		argID++
	}
	if q.DictType != "" {
		where = append(where, fmt.Sprintf("dict_type ILIKE $%d", argID))
		args = append(args, "%"+q.DictType+"%")
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM sys_dict_type WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, dict_type, dict_name, remark, created_at, updated_at
		FROM sys_dict_type
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

	var result []response.DictTypeResp
	for rows.Next() {
		var item response.DictTypeResp
		if err := rows.Scan(&item.ID, &item.DictType, &item.DictName, &item.Remark, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}
	return result, total, rows.Err()
}

// CreateDictType 创建字典类型
func CreateDictType(ctx context.Context, req *request.CreateDictTypeReq) (int64, error) {
	var id int64
	err := database.Pool.QueryRow(ctx, `
		INSERT INTO sys_dict_type (dict_type, dict_name, remark)
		VALUES ($1, $2, $3)
		RETURNING id
	`, req.DictType, req.DictName, req.Remark).Scan(&id)
	return id, err
}

// UpdateDictType 更新字典类型
func UpdateDictType(ctx context.Context, id int64, req *request.UpdateDictTypeReq) error {
	res, err := database.Pool.Exec(ctx, `
		UPDATE sys_dict_type SET dict_type = $1, dict_name = $2, remark = $3, updated_at = NOW()
		WHERE id = $4
	`, req.DictType, req.DictName, req.Remark, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("dict type not found")
	}
	return nil
}

// DeleteDictType 删除字典类型（同时删除其下所有字典数据）
func DeleteDictType(ctx context.Context, id int64) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 获取 dict_type 编码
	var dictType string
	if err := tx.QueryRow(ctx, "SELECT dict_type FROM sys_dict_type WHERE id = $1", id).Scan(&dictType); err != nil {
		return fmt.Errorf("dict type not found")
	}

	// 先删数据
	_, _ = tx.Exec(ctx, "DELETE FROM sys_dict_data WHERE dict_type = $1", dictType)

	// 再删类型
	res, err := tx.Exec(ctx, "DELETE FROM sys_dict_type WHERE id = $1", id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("dict type not found")
	}

	return tx.Commit(ctx)
}

// ========= Dict Data =========

// ListDictData 获取字典数据列表（按排序号）
func ListDictData(ctx context.Context, dictType string) ([]response.DictDataResp, error) {
	query := `
		SELECT id, dict_type, dict_label, dict_value, sort_order, remark, created_at, updated_at
		FROM sys_dict_data
		WHERE dict_type = $1
		ORDER BY sort_order ASC
	`
	rows, err := database.Pool.Query(ctx, query, dictType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []response.DictDataResp
	for rows.Next() {
		var item response.DictDataResp
		if err := rows.Scan(&item.ID, &item.DictType, &item.DictLabel, &item.DictValue, &item.SortOrder, &item.Remark, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// CreateDictData 创建字典数据
func CreateDictData(ctx context.Context, req *request.CreateDictDataReq) (int64, error) {
	var id int64
	err := database.Pool.QueryRow(ctx, `
		INSERT INTO sys_dict_data (dict_type, dict_label, dict_value, sort_order, remark)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, req.DictType, req.DictLabel, req.DictValue, req.SortOrder, req.Remark).Scan(&id)
	return id, err
}

// UpdateDictData 更新字典数据
func UpdateDictData(ctx context.Context, id int64, req *request.UpdateDictDataReq) error {
	res, err := database.Pool.Exec(ctx, `
		UPDATE sys_dict_data SET dict_label = $1, dict_value = $2, sort_order = $3, remark = $4, updated_at = NOW()
		WHERE id = $5
	`, req.DictLabel, req.DictValue, req.SortOrder, req.Remark, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("dict data not found")
	}
	return nil
}

// DeleteDictData 删除字典数据
func DeleteDictData(ctx context.Context, id int64) error {
	res, err := database.Pool.Exec(ctx, "DELETE FROM sys_dict_data WHERE id = $1", id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("dict data not found")
	}
	return nil
}
