/**
 * 功能：请求DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package request

type InventoryAlertQuery struct {
	PageQuery
	MaterialName string `form:"material_name"`
	AlertType    string `form:"alert_type"` // 低库存, 超储
}
