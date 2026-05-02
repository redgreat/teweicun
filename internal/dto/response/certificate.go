/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

import "time"

type CertificateResp struct {
	ID              int64                  `json:"id"`
	CertificateNo   string                 `json:"certificate_no"`
	MaterialID      int64                  `json:"material_id"`
	MaterialCode    string                 `json:"material_code"`
	MaterialName    string                 `json:"material_name"`
	StandardCode    string                 `json:"standard_code"`
	MaterialGrade   string                 `json:"material_grade"`
	ChemicalContent map[string]interface{} `json:"chemical_content"`
	PhysicalProps   map[string]interface{} `json:"physical_props"`
	FileID          string                 `json:"file_id"`
	FileURL         string                 `json:"file_url"`
	Remark          string                 `json:"remark"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}
