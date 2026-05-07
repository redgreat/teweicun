/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

type InventoryAlertResp struct {
	MaterialID    int64   `json:"material_id"`
	MaterialCode  string  `json:"material_code"`
	MaterialName  string  `json:"material_name"`
	Unit          string  `json:"unit"`
	TotalQuantity float64 `json:"total_quantity"`
	SafetyStock   float64 `json:"safety_stock"`
	MaxStock      float64 `json:"max_stock"`
	AlertType     string  `json:"alert_type"`
	AlertQuantity float64 `json:"alert_quantity"`
}
