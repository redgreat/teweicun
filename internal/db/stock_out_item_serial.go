package db

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redgreat/teweicun/pkg/database"
)

func UpdateStockOutItemSerialSelections(ctx context.Context, stockOutItemID int64, serialIDs []int64, userID int64) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var stockOutID int64
	var inventoryID int64
	var qty float64
	var status string
	if err := tx.QueryRow(ctx, `
		SELECT soi.stock_out_id, soi.inventory_id, soi.quantity, so.status
		FROM stock_out_item soi
		INNER JOIN stock_out so ON so.id = soi.stock_out_id
		WHERE soi.id = $1
		  AND so.deleted_at IS NULL
		FOR UPDATE
	`, stockOutItemID).Scan(&stockOutID, &inventoryID, &qty, &status); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("出库明细不存在")
		}
		return err
	}
	if status == "confirmed" {
		return fmt.Errorf("出库单已完成，不能再调整编码")
	}

	need := int(math.Floor(qty + 1e-9))

	// 与退料入库备货一致：先清空当前明细，再支持「保存空列表」= 仅删除备货（清空已选）
	if _, err := tx.Exec(ctx, `DELETE FROM stock_out_item_serial_selection WHERE stock_out_item_id = $1`, stockOutItemID); err != nil {
		return err
	}
	if len(serialIDs) == 0 {
		return tx.Commit(ctx)
	}

	if need <= 0 {
		return fmt.Errorf("该明细不需要备货")
	}
	if len(serialIDs) != need {
		return fmt.Errorf("编码物料需要选择 %d 个编码", need)
	}

	// 逐个锁定并校验
	for _, sid := range serialIDs {
		if sid <= 0 {
			continue
		}
		var serialCode string
		if err := tx.QueryRow(ctx, `
			SELECT serial_code
			FROM material_serial_code
			WHERE id = $1
			  AND inventory_id = $2
			  AND status = 'in_stock'
			FOR UPDATE
		`, sid, inventoryID).Scan(&serialCode); err != nil {
			return fmt.Errorf("编码无效或不在当前库存中 [id=%d]", sid)
		}

		var occupied bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1
				FROM stock_out_item_serial_selection other_sel
				INNER JOIN stock_out_item other_soi ON other_soi.id = other_sel.stock_out_item_id
				INNER JOIN stock_out other_so ON other_so.id = other_soi.stock_out_id
				WHERE other_sel.serial_code_id = $1
				  AND other_so.id <> $2
				  AND other_so.deleted_at IS NULL
				  AND other_so.status IN ('draft', 'pending')
			)
		`, sid, stockOutID).Scan(&occupied); err != nil {
			return err
		}
		if occupied {
			return fmt.Errorf("编码已被其它待出库单占用：%s", serialCode)
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO stock_out_item_serial_selection (stock_out_item_id, serial_code_id, created_by)
			VALUES ($1, $2, $3)
		`, stockOutItemID, sid, userID); err != nil {
			if pgxErr, ok := err.(*pgconn.PgError); ok && pgxErr.Code == "23505" {
				return fmt.Errorf("编码重复或已占用，请刷新后重试")
			}
			return err
		}
	}

	return tx.Commit(ctx)
}

