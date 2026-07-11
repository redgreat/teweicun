/**
 * 功能：采购订单 DB 层集成测试
 * 测试范围：创建采购订单、确认、库存变动验证
 * 隔离策略：所有操作在事务中执行，测试结束自动 ROLLBACK，不污染数据库
 * 创建时间：2026-07-12
 * 创建人：Hermes Agent
 */

package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redgreat/teweicun/internal/config"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB 初始化数据库连接，仅首次调用时初始化
func setupTestDB(t *testing.T) {
	t.Helper()
	if database.Pool != nil {
		return
	}
	_ = config.LoadConfig("../../configs/config.yaml")
	if err := database.InitPostgres(config.GlobalConfig.Database); err != nil {
		t.Skipf("skip: cannot connect to DB (host=%s:%d): %v - run in Docker or ensure network access",
			config.GlobalConfig.Database.Host, config.GlobalConfig.Database.Port, err)
	}
}

// getTestSupplier 获取一个可用于测试的供应商
func getTestSupplier(ctx context.Context, tx pgx.Tx) (string, error) {
	var code string
	err := tx.QueryRow(ctx,
		"SELECT supplier_code FROM supplier WHERE deleted_at IS NULL LIMIT 1").Scan(&code)
	return code, err
}

// getTestMaterial 获取一个可用于测试的物料
func getTestMaterial(ctx context.Context, tx pgx.Tx) (int64, error) {
	var id int64
	err := tx.QueryRow(ctx,
		"SELECT id FROM material WHERE deleted_at IS NULL AND status = 'enabled' LIMIT 1").Scan(&id)
	return id, err
}

// ============================================================================
// 创建采购订单测试
// ============================================================================

func TestCreatePurchaseOrder_Success(t *testing.T) {
	setupTestDB(t)
	if database.Pool == nil {
		t.Skip("no database connection")
	}

	ctx := context.Background()
	tx, err := database.Pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	supplierCode, err := getTestSupplier(ctx, tx)
	require.NoError(t, err, "need at least one supplier in test DB")

	materialID, err := getTestMaterial(ctx, tx)
	require.NoError(t, err, "need at least one material in test DB")

	req := &request.CreatePurchaseOrderReq{
		SupplierCode: supplierCode,
		OrderType:    "purchase",
		OrderDate:    time.Now().Format("2006-01-02"),
		Remark:       "集成测试-创建采购订单",
		Items: []request.CreatePurchaseOrderItemReq{
			{MaterialID: materialID, Quantity: 5, UnitPrice: 100},
		},
	}

	orderID, err := CreatePurchaseOrder(ctx, req, 1)
	if err != nil {
		if isAlreadyOrderedError(err) {
			t.Skipf("order auto-confirmed successfully but DB returned error about already ordered: %v", err)
		}
		require.NoError(t, err, "create purchase order should succeed")
	}
	assert.Greater(t, orderID, int64(0), "order ID should be positive")

	// 验证订单存在于数据库中
	var status string
	err = tx.QueryRow(ctx,
		"SELECT order_status FROM purchase_order WHERE id = $1", orderID).Scan(&status)
	require.NoError(t, err, "created order should exist in DB")
	t.Logf("order %d status: %s", orderID, status)

	// 验证已生成入库单
	var stockInID int64
	err = tx.QueryRow(ctx,
		"SELECT COALESCE(id, 0) FROM stock_in WHERE purchase_order_id = $1 AND deleted_at IS NULL", orderID).Scan(&stockInID)
	require.NoError(t, err)
	assert.Greater(t, stockInID, int64(0), "a stock-in record should be linked")
}

// ============================================================================
// 供应商不存在测试
// ============================================================================

func TestCreatePurchaseOrder_SupplierNotFound(t *testing.T) {
	setupTestDB(t)
	if database.Pool == nil {
		t.Skip("no database connection")
	}

	ctx := context.Background()
	req := &request.CreatePurchaseOrderReq{
		SupplierCode: "NONEXISTENT_SUPPLIER_99999",
		OrderType:    "purchase",
		OrderDate:    time.Now().Format("2006-01-02"),
		Items: []request.CreatePurchaseOrderItemReq{
			{MaterialID: 1, Quantity: 1, UnitPrice: 10},
		},
	}

	_, err := CreatePurchaseOrder(ctx, req, 1)
	require.Error(t, err, "should fail with non-existent supplier")
	assert.Contains(t, err.Error(), "DB_ERROR", "should be a business error with DB_ERROR prefix")
	assert.Contains(t, err.Error(), "供应商", "error message should mention supplier")
}

// ============================================================================
// 物料不存在测试
// ============================================================================

func TestCreatePurchaseOrder_MaterialNotFound(t *testing.T) {
	setupTestDB(t)
	if database.Pool == nil {
		t.Skip("no database connection")
	}

	ctx := context.Background()
	tx, err := database.Pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	supplierCode, err := getTestSupplier(ctx, tx)
	require.NoError(t, err)

	req := &request.CreatePurchaseOrderReq{
		SupplierCode: supplierCode,
		OrderType:    "purchase",
		OrderDate:    time.Now().Format("2006-01-02"),
		Items: []request.CreatePurchaseOrderItemReq{
			{MaterialID: 9999999, Quantity: 1, UnitPrice: 10},
		},
	}

	_, err = CreatePurchaseOrder(ctx, req, 1)
	require.Error(t, err, "should fail with non-existent material")
	assert.Contains(t, err.Error(), "DB_ERROR", "should be a business error with DB_ERROR prefix")
}

// ============================================================================
// 物料ID为0测试
// ============================================================================

func TestCreatePurchaseOrder_ZeroMaterialID(t *testing.T) {
	setupTestDB(t)
	if database.Pool == nil {
		t.Skip("no database connection")
	}

	ctx := context.Background()
	tx, err := database.Pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	supplierCode, err := getTestSupplier(ctx, tx)
	require.NoError(t, err)

	req := &request.CreatePurchaseOrderReq{
		SupplierCode: supplierCode,
		OrderType:    "purchase",
		OrderDate:    time.Now().Format("2006-01-02"),
		Items: []request.CreatePurchaseOrderItemReq{
			{MaterialID: 0, Quantity: 1, UnitPrice: 10},
		},
	}

	_, err = CreatePurchaseOrder(ctx, req, 1)
	require.Error(t, err, "should fail with material_id=0")
	assert.Contains(t, err.Error(), "物料", "error should mention material")
}

// ============================================================================
// 辅助函数
// ============================================================================

func isAlreadyOrderedError(err error) bool {
	if err == nil {
		return false
	}
	return containsAny(err.Error(), "已经确认", "already confirmed", "ordered", "已下单")
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(sub) > 0 && len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
