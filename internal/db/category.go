/**
 * 功能：category.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/pkg/database"
)

// GetCategoryTree fetches the material category tree as raw JSON directly from PostgreSQL
func GetCategoryTree(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := database.Pool.QueryRow(ctx, `SELECT fn_get_category_tree()`).Scan(&result)
	return result, err
}

// CreateCategory inserts a new material category
func CreateCategory(ctx context.Context, req *request.CreateCategoryReq) (int64, error) {
	// 自动生成分类编码
	var categoryCode string
	err := database.Pool.QueryRow(ctx, "SELECT fn_generate_base_code('MC')").Scan(&categoryCode)
	if err != nil {
		return 0, fmt.Errorf("生成分类编码失败: %w", err)
	}

	query := `
		INSERT INTO material_category (parent_id, category_code, category_name, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var id int64
	err = database.Pool.QueryRow(ctx, query,
		req.ParentID, categoryCode, req.CategoryName, req.SortOrder).Scan(&id)
	return id, err
}

// UpdateCategory updates an existing material category
func UpdateCategory(ctx context.Context, id int64, req *request.UpdateCategoryReq) error {
	// 禁用验证：检查是否有物料使用该分类
	if req.Status == "disabled" {
		var refCount int64
		err := database.Pool.QueryRow(ctx, `SELECT count(*) FROM material WHERE category_id = $1 AND deleted_at IS NULL`, id).Scan(&refCount)
		if err != nil {
			return err
		}
		if refCount > 0 {
			return fmt.Errorf("该分类下有 %d 个物料，无法禁用", refCount)
		}
	}

	query := `
		UPDATE material_category
		SET parent_id = $1, category_code = $2, category_name = $3, sort_order = $4, status = $5,
		    updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL
	`
	_, err := database.Pool.Exec(ctx, query,
		req.ParentID, req.CategoryCode, req.CategoryName, req.SortOrder, req.Status, id)
	return err
}

// DeleteCategory soft deletes a material category
func DeleteCategory(ctx context.Context, id int64) error {
	// 删除验证：检查是否有物料使用该分类
	var refCount int64
	err := database.Pool.QueryRow(ctx, `SELECT count(*) FROM material WHERE category_id = $1 AND deleted_at IS NULL`, id).Scan(&refCount)
	if err != nil {
		return err
	}
	if refCount > 0 {
		return fmt.Errorf("该分类下有 %d 个物料，无法删除", refCount)
	}

	query := `
		UPDATE material_category
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err = database.Pool.Exec(ctx, query, id)
	return err
}
