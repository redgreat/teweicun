/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type CertificateQuery struct {
	PageQuery
	CertificateNo string `form:"certificate_no"`
	MaterialCode  string `form:"material_code"`
}

type CreateCertificateReq struct {
	CertificateNo   string                 `json:"certificate_no" binding:"required"`
	MaterialID      int64                  `json:"material_id" binding:"required"`
	StandardCode    string                 `json:"standard_code"`
	MaterialGrade   string                 `json:"material_grade"`
	ChemicalContent map[string]interface{} `json:"chemical_content"`
	PhysicalProps   map[string]interface{} `json:"physical_props"`
	FileID          string                 `json:"file_id"` // Path from upload
	Remark          string                 `json:"remark"`
}

type UpdateCertificateReq struct {
	StandardCode    string                 `json:"standard_code"`
	MaterialGrade   string                 `json:"material_grade"`
	ChemicalContent map[string]interface{} `json:"chemical_content"`
	PhysicalProps   map[string]interface{} `json:"physical_props"`
	FileID          string                 `json:"file_id"`
	Remark          string                 `json:"remark"`
}
