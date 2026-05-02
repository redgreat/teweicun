/**
 * 功能：category.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package service

import (
	"context"
	"encoding/json"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
)

// GetCategoryTree returns the JSON encoded material category tree
func GetCategoryTree(ctx context.Context) (json.RawMessage, error) {
	return db.GetCategoryTree(ctx)
}

// CreateCategory creates a new material category
func CreateCategory(ctx context.Context, req *request.CreateCategoryReq) (int64, error) {
	return db.CreateCategory(ctx, req)
}

// UpdateCategory updates an existing material category
func UpdateCategory(ctx context.Context, id int64, req *request.UpdateCategoryReq) error {
	return db.UpdateCategory(ctx, id, req)
}

// DeleteCategory soft deletes a material category
func DeleteCategory(ctx context.Context, id int64) error {
	return db.DeleteCategory(ctx, id)
}
