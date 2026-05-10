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
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
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
	categoryCode, sortOrder, err := generateCategoryMeta(ctx, req.ParentID)
	if err != nil {
		return 0, err
	}

	query := `
		INSERT INTO material_category (parent_id, category_code, category_name, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var id int64
	err = database.Pool.QueryRow(ctx, query,
		req.ParentID, categoryCode, req.CategoryName, sortOrder).Scan(&id)
	return id, err
}

func generateCategoryMeta(ctx context.Context, parentID int64) (string, int, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return "", 0, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// 同一父级下串行生成编码，避免并发创建时冲突
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", 100000+parentID); err != nil {
		return "", 0, fmt.Errorf("锁定分类编码序列失败: %w", err)
	}

	var nextCode string
	var nextSortOrder int
	if parentID == 0 {
		var lastTopCode string
		err = tx.QueryRow(ctx, `
			SELECT category_code
			FROM material_category
			WHERE parent_id = 0
			  AND deleted_at IS NULL
			  AND category_code ~ '^[A-Z]+$'
			ORDER BY LENGTH(category_code) DESC, category_code DESC
			LIMIT 1
		`).Scan(&lastTopCode)
		if err != nil && err != pgx.ErrNoRows {
			return "", 0, fmt.Errorf("查询顶级分类编码失败: %w", err)
		}
		if err == pgx.ErrNoRows {
			lastTopCode = ""
		}
		nextCode, err = nextAlphaCode(lastTopCode)
		if err != nil {
			return "", 0, err
		}
	} else {
		var parentCode string
		var parentLevel int
		err = tx.QueryRow(ctx, `
			SELECT
				c.category_code,
				(
					WITH RECURSIVE chain AS (
						SELECT id, parent_id
						FROM material_category
						WHERE id = $1 AND deleted_at IS NULL
						UNION ALL
						SELECT p.id, p.parent_id
						FROM material_category p
						JOIN chain ON chain.parent_id = p.id
						WHERE p.deleted_at IS NULL
					)
					SELECT count(*) FROM chain
				) AS level
			FROM material_category c
			WHERE c.id = $1 AND c.deleted_at IS NULL
		`, parentID).Scan(&parentCode, &parentLevel)
		if err != nil {
			if err == pgx.ErrNoRows {
				return "", 0, errcode.NewAppError(errcode.ErrInvalidParam.Code, "上级分类不存在", errcode.ErrInvalidParam.HTTPCode)
			}
			return "", 0, fmt.Errorf("查询上级分类失败: %w", err)
		}
		if parentLevel >= 2 {
			return "", 0, errcode.NewAppError(errcode.ErrInvalidParam.Code, "分类最多支持二级，不能继续新增子分类", errcode.ErrInvalidParam.HTTPCode)
		}

		var maxSuffix int
		err = tx.QueryRow(ctx, `
			SELECT COALESCE(MAX(RIGHT(category_code, 2)::int), 0)
			FROM material_category
			WHERE parent_id = $1
			  AND deleted_at IS NULL
			  AND category_code ~ ('^' || $2 || '\d{2}$')
		`, parentID, parentCode).Scan(&maxSuffix)
		if err != nil {
			return "", 0, fmt.Errorf("查询子级分类编码失败: %w", err)
		}
		if maxSuffix >= 99 {
			return "", 0, errcode.NewAppError(errcode.ErrInvalidParam.Code, "当前分类下编码已用尽（最多99个子分类）", errcode.ErrInvalidParam.HTTPCode)
		}
		nextCode = fmt.Sprintf("%s%02d", parentCode, maxSuffix+1)
	}

	err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(sort_order), 0) + 1
		FROM material_category
		WHERE parent_id = $1
		  AND deleted_at IS NULL
	`, parentID).Scan(&nextSortOrder)
	if err != nil {
		return "", 0, fmt.Errorf("计算分类排序号失败: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return "", 0, err
	}
	return nextCode, nextSortOrder, nil
}

func nextAlphaCode(lastCode string) (string, error) {
	if lastCode == "" {
		return "A", nil
	}

	code := strings.ToUpper(lastCode)
	var n int
	for _, ch := range code {
		if ch < 'A' || ch > 'Z' {
			return "", fmt.Errorf("无效的顶级分类编码: %s", lastCode)
		}
		n = n*26 + int(ch-'A'+1)
	}
	n++

	buf := make([]byte, 0, len(code)+1)
	for n > 0 {
		n--
		buf = append(buf, byte('A'+(n%26)))
		n /= 26
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf), nil
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
			return errcode.NewAppError(errcode.ErrInvalidParam.Code, fmt.Sprintf("该分类下有 %d 个物料，无法禁用", refCount), errcode.ErrInvalidParam.HTTPCode)
		}
	}

	var dupCount int64
	err := database.Pool.QueryRow(ctx, `
		SELECT count(*)
		FROM material_category
		WHERE parent_id = $1
		  AND sort_order = $2
		  AND id <> $3
		  AND deleted_at IS NULL
	`, req.ParentID, req.SortOrder, id).Scan(&dupCount)
	if err != nil {
		return err
	}
	if dupCount > 0 {
		return errcode.NewAppError(errcode.ErrInvalidParam.Code, "同级分类排序号不能重复", errcode.ErrInvalidParam.HTTPCode)
	}

	query := `
		UPDATE material_category
		SET parent_id = $1, category_code = $2, category_name = $3, sort_order = $4, status = $5,
		    updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL
	`
	_, err = database.Pool.Exec(ctx, query,
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
		return errcode.NewAppError(errcode.ErrInvalidParam.Code, fmt.Sprintf("该分类下有 %d 个物料，无法删除", refCount), errcode.ErrInvalidParam.HTTPCode)
	}

	query := `
		UPDATE material_category
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err = database.Pool.Exec(ctx, query, id)
	return err
}
