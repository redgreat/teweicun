import os

def create_file(path, content):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w') as f:
        f.write(content.strip() + '\n')
    print(f"Created {path}")

# DTO: request
dto_req = """
package request

import "time"

type FundPaymentQuery struct {
	BaseQuery
	StatementNo string `form:"statement_no"`
	SupplierID  int64  `form:"supplier_id"`
	Status      string `form:"status"`
}

type CreateFundPaymentReq struct {
	SupplierID        int64                     `json:"supplier_id" binding:"required"`
	StatementDate     string                    `json:"statement_date" binding:"required"`
	PaymentAmount     float64                   `json:"payment_amount"`
	DiscountAmount    float64                   `json:"discount_amount"`
	AdvanceAmount     float64                   `json:"advance_amount"`
	SettlementMethod  string                    `json:"settlement_method"`
	SettlementAccount string                    `json:"settlement_account"`
	SettlementNo      string                    `json:"settlement_no"`
	Remark            string                    `json:"remark"`
	Items             []FundPaymentItemReq      `json:"items" binding:"required"`
}

type FundPaymentItemReq struct {
	SourceOrderID       int64   `json:"source_order_id" binding:"required"`
	SourceOrderNo       string  `json:"source_order_no" binding:"required"`
	BusinessType        string  `json:"business_type" binding:"required"`
	OrderDate           string  `json:"order_date" binding:"required"`
	OrderAmount         float64 `json:"order_amount"`
	VerifiedAmount      float64 `json:"verified_amount"`
	UnverifiedAmount    float64 `json:"unverified_amount"`
	CurrentVerifyAmount float64 `json:"current_verify_amount" binding:"required"`
	CustomTaxAmount     float64 `json:"custom_tax_amount"`
	Remark              string  `json:"remark"`
}

type FundCollectionQuery struct {
	BaseQuery
	StatementNo string `form:"statement_no"`
	CustomerID  int64  `form:"customer_id"`
	Status      string `form:"status"`
}

type CreateFundCollectionReq struct {
	CustomerID        int64                        `json:"customer_id" binding:"required"`
	StatementDate     string                       `json:"statement_date" binding:"required"`
	CollectionAmount  float64                      `json:"collection_amount"`
	DiscountAmount    float64                      `json:"discount_amount"`
	AdvanceAmount     float64                      `json:"advance_amount"`
	SettlementMethod  string                       `json:"settlement_method"`
	SettlementAccount string                       `json:"settlement_account"`
	SettlementNo      string                       `json:"settlement_no"`
	Remark            string                       `json:"remark"`
	Items             []FundCollectionItemReq      `json:"items" binding:"required"`
}

type FundCollectionItemReq struct {
	SourceOrderID       int64   `json:"source_order_id" binding:"required"`
	SourceOrderNo       string  `json:"source_order_no" binding:"required"`
	BusinessType        string  `json:"business_type" binding:"required"`
	OrderDate           string  `json:"order_date" binding:"required"`
	OrderAmount         float64 `json:"order_amount"`
	VerifiedAmount      float64 `json:"verified_amount"`
	UnverifiedAmount    float64 `json:"unverified_amount"`
	CurrentVerifyAmount float64 `json:"current_verify_amount" binding:"required"`
	CustomTaxAmount     float64 `json:"custom_tax_amount"`
	Remark              string  `json:"remark"`
}
"""

create_file('internal/dto/request/fund.go', dto_req)

# DTO: response
dto_resp = """
package response

import "time"

type FundPaymentResp struct {
	ID                int64     `json:"id"`
	StatementNo       string    `json:"statement_no"`
	SupplierID        int64     `json:"supplier_id"`
	SupplierName      string    `json:"supplier_name"`
	SupplierCode      string    `json:"supplier_code"`
	PayerID           int64     `json:"payer_id"`
	PayerName         string    `json:"payer_name"`
	StatementDate     string    `json:"statement_date"`
	PaymentAmount     float64   `json:"payment_amount"`
	DiscountAmount    float64   `json:"discount_amount"`
	AdvanceAmount     float64   `json:"advance_amount"`
	SettlementMethod  string    `json:"settlement_method"`
	SettlementAccount string    `json:"settlement_account"`
	SettlementNo      string    `json:"settlement_no"`
	Remark            string    `json:"remark"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type FundPaymentDetailResp struct {
	FundPaymentResp
	Items []FundPaymentItemResp `json:"items"`
}

type FundPaymentItemResp struct {
	ID                  int64     `json:"id"`
	StatementID         int64     `json:"statement_id"`
	SourceOrderID       int64     `json:"source_order_id"`
	SourceOrderNo       string    `json:"source_order_no"`
	BusinessType        string    `json:"business_type"`
	OrderDate           string    `json:"order_date"`
	OrderAmount         float64   `json:"order_amount"`
	VerifiedAmount      float64   `json:"verified_amount"`
	UnverifiedAmount    float64   `json:"unverified_amount"`
	CurrentVerifyAmount float64   `json:"current_verify_amount"`
	CustomTaxAmount     float64   `json:"custom_tax_amount"`
	Remark              string    `json:"remark"`
	CreatedAt           time.Time `json:"created_at"`
}

type FundCollectionResp struct {
	ID                int64     `json:"id"`
	StatementNo       string    `json:"statement_no"`
	CustomerID        int64     `json:"customer_id"`
	CustomerName      string    `json:"customer_name"`
	CustomerCode      string    `json:"customer_code"`
	PayeeID           int64     `json:"payee_id"`
	PayeeName         string    `json:"payee_name"`
	StatementDate     string    `json:"statement_date"`
	CollectionAmount  float64   `json:"collection_amount"`
	DiscountAmount    float64   `json:"discount_amount"`
	AdvanceAmount     float64   `json:"advance_amount"`
	SettlementMethod  string    `json:"settlement_method"`
	SettlementAccount string    `json:"settlement_account"`
	SettlementNo      string    `json:"settlement_no"`
	Remark            string    `json:"remark"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type FundCollectionDetailResp struct {
	FundCollectionResp
	Items []FundCollectionItemResp `json:"items"`
}

type FundCollectionItemResp struct {
	ID                  int64     `json:"id"`
	StatementID         int64     `json:"statement_id"`
	SourceOrderID       int64     `json:"source_order_id"`
	SourceOrderNo       string    `json:"source_order_no"`
	BusinessType        string    `json:"business_type"`
	OrderDate           string    `json:"order_date"`
	OrderAmount         float64   `json:"order_amount"`
	VerifiedAmount      float64   `json:"verified_amount"`
	UnverifiedAmount    float64   `json:"unverified_amount"`
	CurrentVerifyAmount float64   `json:"current_verify_amount"`
	CustomTaxAmount     float64   `json:"custom_tax_amount"`
	Remark              string    `json:"remark"`
	CreatedAt           time.Time `json:"created_at"`
}
"""

create_file('internal/dto/response/fund.go', dto_resp)

# DB Layer
db_content = """
package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/pkg/database"
)

// ListFundPayments 分页查询付款单
func ListFundPayments(ctx context.Context, q *request.FundPaymentQuery) ([]response.FundPaymentResp, int64, error) {
	where := []string{"deleted_at IS NULL"}
	var args []interface{}
	argID := 1

	if q.StatementNo != "" {
		where = append(where, fmt.Sprintf("statement_no ILIKE $%d", argID))
		args = append(args, "%"+q.StatementNo+"%")
		argID++
	}
	if q.SupplierID != 0 {
		where = append(where, fmt.Sprintf("supplier_id = $%d", argID))
		args = append(args, q.SupplierID)
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_fund_payment_list WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, statement_no, supplier_id, supplier_name, supplier_code,
		       payer_id, payer_name, TO_CHAR(statement_date, 'YYYY-MM-DD'),
		       payment_amount, discount_amount, advance_amount,
		       settlement_method, settlement_account, settlement_no,
		       remark, status, created_at, updated_at
		FROM v_fund_payment_list
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

	var result []response.FundPaymentResp
	for rows.Next() {
		var item response.FundPaymentResp
		if err := rows.Scan(
			&item.ID, &item.StatementNo, &item.SupplierID, &item.SupplierName, &item.SupplierCode,
			&item.PayerID, &item.PayerName, &item.StatementDate,
			&item.PaymentAmount, &item.DiscountAmount, &item.AdvanceAmount,
			&item.SettlementMethod, &item.SettlementAccount, &item.SettlementNo,
			&item.Remark, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// GetFundPayment 获取付款单详情
func GetFundPayment(ctx context.Context, id int64) (*response.FundPaymentDetailResp, error) {
	query := `
		SELECT id, statement_no, supplier_id, supplier_name, supplier_code,
		       payer_id, payer_name, TO_CHAR(statement_date, 'YYYY-MM-DD'),
		       payment_amount, discount_amount, advance_amount,
		       settlement_method, settlement_account, settlement_no,
		       remark, status, created_at, updated_at
		FROM v_fund_payment_list
		WHERE id = $1
	`
	var order response.FundPaymentDetailResp
	err := database.Pool.QueryRow(ctx, query, id).Scan(
		&order.ID, &order.StatementNo, &order.SupplierID, &order.SupplierName, &order.SupplierCode,
		&order.PayerID, &order.PayerName, &order.StatementDate,
		&order.PaymentAmount, &order.DiscountAmount, &order.AdvanceAmount,
		&order.SettlementMethod, &order.SettlementAccount, &order.SettlementNo,
		&order.Remark, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	itemsQuery := `
		SELECT id, statement_id, source_order_id, source_order_no, business_type,
		       TO_CHAR(order_date, 'YYYY-MM-DD'), order_amount, verified_amount,
		       unverified_amount, current_verify_amount, custom_tax_amount, remark, created_at
		FROM fund_payment_item
		WHERE statement_id = $1
	`
	rows, err := database.Pool.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item response.FundPaymentItemResp
		if err := rows.Scan(
			&item.ID, &item.StatementID, &item.SourceOrderID, &item.SourceOrderNo, &item.BusinessType,
			&item.OrderDate, &item.OrderAmount, &item.VerifiedAmount,
			&item.UnverifiedAmount, &item.CurrentVerifyAmount, &item.CustomTaxAmount, &item.Remark, &item.CreatedAt); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

// CreateFundPayment 创建付款单
func CreateFundPayment(ctx context.Context, req *request.CreateFundPaymentReq, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var statementNo string
	err = tx.QueryRow(ctx, "SELECT fn_generate_bill_no('FK')").Scan(&statementNo)
	if err != nil {
		return 0, fmt.Errorf("生成单号失败: %w", err)
	}

	var id int64
	query := `
		INSERT INTO fund_payment (
			statement_no, supplier_id, payer_id, statement_date,
			payment_amount, discount_amount, advance_amount,
			settlement_method, settlement_account, settlement_no,
			remark, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`
	err = tx.QueryRow(ctx, query,
		statementNo, req.SupplierID, userID, req.StatementDate,
		req.PaymentAmount, req.DiscountAmount, req.AdvanceAmount,
		req.SettlementMethod, req.SettlementAccount, req.SettlementNo,
		req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
	}

	for _, item := range req.Items {
		itemQuery := `
			INSERT INTO fund_payment_item (
				statement_id, source_order_id, source_order_no, business_type,
				order_date, order_amount, verified_amount,
				unverified_amount, current_verify_amount, custom_tax_amount, remark
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		_, err = tx.Exec(ctx, itemQuery,
			id, item.SourceOrderID, item.SourceOrderNo, item.BusinessType,
			item.OrderDate, item.OrderAmount, item.VerifiedAmount,
			item.UnverifiedAmount, item.CurrentVerifyAmount, item.CustomTaxAmount, item.Remark)
		if err != nil {
			return 0, err
		}
		
		// Update purchase order verified amount
		if item.BusinessType == "采购入库" {
			updatePo := `UPDATE purchase_order SET verified_amount = COALESCE(verified_amount, 0) + $1 WHERE id = $2`
			_, err = tx.Exec(ctx, updatePo, item.CurrentVerifyAmount, item.SourceOrderID)
			if err != nil {
				return 0, err
			}
		}
	}

	// 审批付款单
	approveQuery := `UPDATE fund_payment SET status = 'completed' WHERE id = $1`
	_, err = tx.Exec(ctx, approveQuery, id)
	if err != nil {
	    return 0, err
	}

	return id, tx.Commit(ctx)
}

// ListFundCollections 分页查询收款单
func ListFundCollections(ctx context.Context, q *request.FundCollectionQuery) ([]response.FundCollectionResp, int64, error) {
	where := []string{"deleted_at IS NULL"}
	var args []interface{}
	argID := 1

	if q.StatementNo != "" {
		where = append(where, fmt.Sprintf("statement_no ILIKE $%d", argID))
		args = append(args, "%"+q.StatementNo+"%")
		argID++
	}
	if q.CustomerID != 0 {
		where = append(where, fmt.Sprintf("customer_id = $%d", argID))
		args = append(args, q.CustomerID)
		argID++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argID))
		args = append(args, q.Status)
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM v_fund_collection_list WHERE %s", whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, statement_no, customer_id, customer_name, customer_code,
		       payee_id, payee_name, TO_CHAR(statement_date, 'YYYY-MM-DD'),
		       collection_amount, discount_amount, advance_amount,
		       settlement_method, settlement_account, settlement_no,
		       remark, status, created_at, updated_at
		FROM v_fund_collection_list
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

	var result []response.FundCollectionResp
	for rows.Next() {
		var item response.FundCollectionResp
		if err := rows.Scan(
			&item.ID, &item.StatementNo, &item.CustomerID, &item.CustomerName, &item.CustomerCode,
			&item.PayeeID, &item.PayeeName, &item.StatementDate,
			&item.CollectionAmount, &item.DiscountAmount, &item.AdvanceAmount,
			&item.SettlementMethod, &item.SettlementAccount, &item.SettlementNo,
			&item.Remark, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	return result, total, rows.Err()
}

// GetFundCollection 获取收款单详情
func GetFundCollection(ctx context.Context, id int64) (*response.FundCollectionDetailResp, error) {
	query := `
		SELECT id, statement_no, customer_id, customer_name, customer_code,
		       payee_id, payee_name, TO_CHAR(statement_date, 'YYYY-MM-DD'),
		       collection_amount, discount_amount, advance_amount,
		       settlement_method, settlement_account, settlement_no,
		       remark, status, created_at, updated_at
		FROM v_fund_collection_list
		WHERE id = $1
	`
	var order response.FundCollectionDetailResp
	err := database.Pool.QueryRow(ctx, query, id).Scan(
		&order.ID, &order.StatementNo, &order.CustomerID, &order.CustomerName, &order.CustomerCode,
		&order.PayeeID, &order.PayeeName, &order.StatementDate,
		&order.CollectionAmount, &order.DiscountAmount, &order.AdvanceAmount,
		&order.SettlementMethod, &order.SettlementAccount, &order.SettlementNo,
		&order.Remark, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	itemsQuery := `
		SELECT id, statement_id, source_order_id, source_order_no, business_type,
		       TO_CHAR(order_date, 'YYYY-MM-DD'), order_amount, verified_amount,
		       unverified_amount, current_verify_amount, custom_tax_amount, remark, created_at
		FROM fund_collection_item
		WHERE statement_id = $1
	`
	rows, err := database.Pool.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item response.FundCollectionItemResp
		if err := rows.Scan(
			&item.ID, &item.StatementID, &item.SourceOrderID, &item.SourceOrderNo, &item.BusinessType,
			&item.OrderDate, &item.OrderAmount, &item.VerifiedAmount,
			&item.UnverifiedAmount, &item.CurrentVerifyAmount, &item.CustomTaxAmount, &item.Remark, &item.CreatedAt); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

// CreateFundCollection 创建收款单
func CreateFundCollection(ctx context.Context, req *request.CreateFundCollectionReq, userID int64, username string) (int64, error) {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var statementNo string
	err = tx.QueryRow(ctx, "SELECT fn_generate_bill_no('SK')").Scan(&statementNo)
	if err != nil {
		return 0, fmt.Errorf("生成单号失败: %w", err)
	}

	var id int64
	query := `
		INSERT INTO fund_collection (
			statement_no, customer_id, payee_id, statement_date,
			collection_amount, discount_amount, advance_amount,
			settlement_method, settlement_account, settlement_no,
			remark, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`
	err = tx.QueryRow(ctx, query,
		statementNo, req.CustomerID, userID, req.StatementDate,
		req.CollectionAmount, req.DiscountAmount, req.AdvanceAmount,
		req.SettlementMethod, req.SettlementAccount, req.SettlementNo,
		req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
	}

	for _, item := range req.Items {
		itemQuery := `
			INSERT INTO fund_collection_item (
				statement_id, source_order_id, source_order_no, business_type,
				order_date, order_amount, verified_amount,
				unverified_amount, current_verify_amount, custom_tax_amount, remark
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		_, err = tx.Exec(ctx, itemQuery,
			id, item.SourceOrderID, item.SourceOrderNo, item.BusinessType,
			item.OrderDate, item.OrderAmount, item.VerifiedAmount,
			item.UnverifiedAmount, item.CurrentVerifyAmount, item.CustomTaxAmount, item.Remark)
		if err != nil {
			return 0, err
		}
		
		// Update sales order verified amount
		if item.BusinessType == "销售出库" {
			updateSo := `UPDATE sales_order SET verified_amount = COALESCE(verified_amount, 0) + $1 WHERE id = $2`
			_, err = tx.Exec(ctx, updateSo, item.CurrentVerifyAmount, item.SourceOrderID)
			if err != nil {
				return 0, err
			}
		}
	}

	approveQuery := `UPDATE fund_collection SET status = 'completed' WHERE id = $1`
	_, err = tx.Exec(ctx, approveQuery, id)
	if err != nil {
	    return 0, err
	}

	return id, tx.Commit(ctx)
}
"""
create_file('internal/db/fund.go', db_content)

# Service Layer
service_content = """
package service

import (
	"context"

	"github.com/redgreat/teweicun/internal/db"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
)

func ListFundPayments(ctx context.Context, q *request.FundPaymentQuery) ([]response.FundPaymentResp, int64, error) {
	return db.ListFundPayments(ctx, q)
}

func GetFundPayment(ctx context.Context, id int64) (*response.FundPaymentDetailResp, error) {
	return db.GetFundPayment(ctx, id)
}

func CreateFundPayment(ctx context.Context, req *request.CreateFundPaymentReq, userID int64, username string) (int64, error) {
	return db.CreateFundPayment(ctx, req, userID, username)
}

func ListFundCollections(ctx context.Context, q *request.FundCollectionQuery) ([]response.FundCollectionResp, int64, error) {
	return db.ListFundCollections(ctx, q)
}

func GetFundCollection(ctx context.Context, id int64) (*response.FundCollectionDetailResp, error) {
	return db.GetFundCollection(ctx, id)
}

func CreateFundCollection(ctx context.Context, req *request.CreateFundCollectionReq, userID int64, username string) (int64, error) {
	return db.CreateFundCollection(ctx, req, userID, username)
}
"""
create_file('internal/service/fund.go', service_content)

# Handler Layer
handler_content = """
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/request"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/middleware"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/internal/service"
)

func ListFundPayments(c *gin.Context) {
	var q request.FundPaymentQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 10
	}

	list, total, err := service.ListFundPayments(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func GetFundPayment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	order, err := service.GetFundPayment(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	if order == nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.Success(c, order)
}

func CreateFundPayment(c *gin.Context) {
	var req request.CreateFundPaymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	username, _ := c.Get("username")
	id, err := service.CreateFundPayment(c.Request.Context(), &req, userID, username.(string))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func ListFundCollections(c *gin.Context) {
	var q request.FundCollectionQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 10
	}

	list, total, err := service.ListFundCollections(c.Request.Context(), &q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessPage(c, total, list)
}

func GetFundCollection(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Invalid ID", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	order, err := service.GetFundCollection(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	if order == nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}
	response.Success(c, order)
}

func CreateFundCollection(c *gin.Context) {
	var req request.CreateFundCollectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
		return
	}

	userID, _ := middleware.GetUserID(c)
	username, _ := c.Get("username")
	id, err := service.CreateFundCollection(c.Request.Context(), &req, userID, username.(string))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}
"""
create_file('internal/handler/fund.go', handler_content)
