package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redgreat/teweicun/pkg/database"
)

func UpdateStockInItemSerialSelections(ctx context.Context, stockInItemID int64, serialIDs []int64, userID int64) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var stockInID int64
	var materialID int64
	if err := tx.QueryRow(ctx, `
		SELECT sii.stock_in_id, sii.material_id
		FROM stock_in_item sii
		WHERE sii.id = $1
		FOR UPDATE
	`, stockInItemID).Scan(&stockInID, &materialID); err != nil {
		return err
	}

	var stockInType string
	var stockInStatus string
	if err := tx.QueryRow(ctx, `
		SELECT stock_in_type, stock_in_status
		FROM stock_in
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, stockInID).Scan(&stockInType, &stockInStatus); err != nil {
		return err
	}
	if stockInType != "reversal" {
		return fmt.Errorf("仅退料入库单支持备货编码")
	}
	if stockInStatus == "passed" {
		return fmt.Errorf("入库单已完成，不能再备货编码")
	}

	// 先清空当前明细已有选择
	if _, err := tx.Exec(ctx, `DELETE FROM stock_in_item_serial_selection WHERE stock_in_item_id = $1`, stockInItemID); err != nil {
		return err
	}

	if len(serialIDs) == 0 {
		return tx.Commit(ctx)
	}

	// 插入新选择：逐个锁定编码并校验可用性（事务内防并发）
	for _, sid := range serialIDs {
		if sid <= 0 {
			continue
		}

		// 1) 锁定编码并校验状态/物料归属
		var status string
		var serialCode string
		if err := tx.QueryRow(ctx, `
			SELECT sc.status, sc.serial_code
			FROM material_serial_code sc
			WHERE sc.id = $1
			  AND sc.material_id = $2
			FOR UPDATE
		`, sid, materialID).Scan(&status, &serialCode); err != nil {
			return fmt.Errorf("编码不存在或物料不匹配 [id=%d]", sid)
		}
		if status != "issued" {
			return fmt.Errorf("编码状态不允许备货（需要已出库/已领用）：%s", serialCode)
		}

		// 2) 必须来自领料出库且出库单已完成
		var ok bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1
				FROM material_serial_trace tr
				INNER JOIN stock_out so ON so.id = tr.ref_doc_id AND so.deleted_at IS NULL
				WHERE tr.serial_code_id = $1
				  AND tr.action = 'stock_out'
				  AND tr.ref_doc_type = 'stock_out'
				  AND so.status = 'confirmed'
				  AND so.ref_doc_type = 'consumption_order'
			)
		`, sid).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("编码不满足退料回库条件：%s", serialCode)
		}

		// 3) 不能被其它待入库/部分入库的退料入库单占用
		if err := tx.QueryRow(ctx, `
			SELECT NOT EXISTS(
				SELECT 1
				FROM stock_in_item_serial_selection other_sel
				INNER JOIN stock_in other_si ON other_si.id = other_sel.stock_in_id
				WHERE other_sel.serial_code_id = $1
				  AND other_sel.stock_in_item_id <> $2
				  AND other_si.deleted_at IS NULL
				  AND other_si.stock_in_status IN ('preparing', 'pending')
			)
		`, sid, stockInItemID).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("编码已被其它退料入库单占用：%s", serialCode)
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO stock_in_item_serial_selection (stock_in_id, stock_in_item_id, serial_code_id, created_by)
			VALUES ($1, $2, $3, $4)
		`, stockInID, stockInItemID, sid, userID)
		if err != nil {
			// 唯一约束冲突，转换为可读错误
			if pgxErr, ok := err.(*pgconn.PgError); ok && pgxErr.Code == "23505" {
				return fmt.Errorf("存在编码已被其它退料入库单占用，请刷新后重试")
			}
			return err
		}
	}

	return tx.Commit(ctx)
}

