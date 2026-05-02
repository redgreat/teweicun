/**
 * 功能：certificate.go
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
	"github.com/redgreat/teweicun/pkg/storage"
)

// ListCertificates 分页查询材质证明
func ListCertificates(ctx context.Context, q *request.CertificateQuery) ([]response.CertificateResp, int64, error) {
	where := []string{"c.deleted_at IS NULL"}
	var args []interface{}
	argID := 1

	if q.CertificateNo != "" {
		where = append(where, fmt.Sprintf("c.certificate_no ILIKE $%d", argID))
		args = append(args, "%"+q.CertificateNo+"%")
		argID++
	}
	if q.MaterialCode != "" {
		where = append(where, fmt.Sprintf("m.material_code ILIKE $%d", argID))
		args = append(args, "%"+q.MaterialCode+"%")
		argID++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`
		SELECT count(*) 
		FROM material_certificate c
		INNER JOIN material m ON m.id = c.material_id
		WHERE %s`, whereClause)
	if err := database.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.certificate_no, c.material_id, m.material_code, m.material_name,
		       c.standard_code, c.material_grade, c.chemical_content, c.physical_props,
		       COALESCE(c.file_id, ''), COALESCE(c.remark, ''), c.created_at, c.updated_at
		FROM material_certificate c
		INNER JOIN material m ON m.id = c.material_id
		WHERE %s
		ORDER BY c.id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argID, argID+1)

	args = append(args, q.PageSize, q.Offset())

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []response.CertificateResp
	for rows.Next() {
		var item response.CertificateResp
		var chemBytes, physBytes []byte
		if err := rows.Scan(&item.ID, &item.CertificateNo, &item.MaterialID, &item.MaterialCode, &item.MaterialName,
			&item.StandardCode, &item.MaterialGrade, &chemBytes, &physBytes,
			&item.FileID, &item.Remark, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		
		if len(chemBytes) > 0 {
			json.Unmarshal(chemBytes, &item.ChemicalContent)
		}
		if len(physBytes) > 0 {
			json.Unmarshal(physBytes, &item.PhysicalProps)
		}
		
		if item.FileID != "" && storage.GlobalStorage != nil {
			item.FileURL, _ = storage.GlobalStorage.GetURL(item.FileID)
		}

		result = append(result, item)
	}

	return result, total, rows.Err()
}

// CreateCertificate 创建材质证明书
func CreateCertificate(ctx context.Context, req *request.CreateCertificateReq, userID int64, username string) (int64, error) {
	chemBytes, _ := json.Marshal(req.ChemicalContent)
	physBytes, _ := json.Marshal(req.PhysicalProps)

	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var id int64
	query := `
		INSERT INTO material_certificate (certificate_no, material_id, standard_code, 
		                                 material_grade, chemical_content, physical_props, file_id, remark, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	err = tx.QueryRow(ctx, query, req.CertificateNo, req.MaterialID, req.StandardCode,
		req.MaterialGrade, chemBytes, physBytes, req.FileID, req.Remark, userID).Scan(&id)
	if err != nil {
		return 0, err
	}

	// 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "CREATE", "CERTIFICATE", "material_certificate", id, nil)
	if err != nil {
		return 0, err
	}

	return id, tx.Commit(ctx)
}

// UpdateCertificate 更新材质证明书
func UpdateCertificate(ctx context.Context, id int64, req *request.UpdateCertificateReq, userID int64, username string) error {
	chemBytes, _ := json.Marshal(req.ChemicalContent)
	physBytes, _ := json.Marshal(req.PhysicalProps)

	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE material_certificate
		SET standard_code = $1, material_grade = $2, chemical_content = $3, 
		    physical_props = $4, file_id = $5, remark = $6, updated_by = $7, updated_at = NOW()
		WHERE id = $8 AND deleted_at IS NULL
	`
	res, err := tx.Exec(ctx, query, req.StandardCode, req.MaterialGrade, chemBytes,
		physBytes, req.FileID, req.Remark, userID, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("certificate not found")
	}

	// 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "UPDATE", "CERTIFICATE", "material_certificate", id, nil)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// DeleteCertificate 软删除材质证明书
func DeleteCertificate(ctx context.Context, id int64, userID int64, username string) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `UPDATE material_certificate SET deleted_at = NOW(), updated_by = $1 WHERE id = $2 AND deleted_at IS NULL`
	res, err := tx.Exec(ctx, query, userID, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("certificate not found")
	}

	// 审计日志
	auditQuery := `CALL sp_write_audit_log($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, auditQuery, userID, username, "DELETE", "CERTIFICATE", "material_certificate", id, nil)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
