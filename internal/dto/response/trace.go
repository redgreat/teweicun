/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

type TraceForwardResp struct {
	TraceType    string  `json:"trace_type"`
	DocNo        string  `json:"doc_no"`
	DocDate      string  `json:"doc_date"`
	MaterialCode string  `json:"material_code"`
	MaterialName string  `json:"material_name"`
	Quantity     float64 `json:"quantity"`
	Warehouse    string  `json:"warehouse"`
}

type TraceBackwardResp struct {
	TraceType    string  `json:"trace_type"`
	DocNo        string  `json:"doc_no"`
	DocDate      string  `json:"doc_date"`
	MaterialCode string  `json:"material_code"`
	MaterialName string  `json:"material_name"`
	Quantity     float64 `json:"quantity"`
	SupplierName string  `json:"supplier_name"`
	CertNo       string  `json:"cert_no"`
}
