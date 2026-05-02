/**
 * 功能：响应DTO定义
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package response

type LoginResp struct {
	Token    string   `json:"token"`
	UserID   int64    `json:"user_id"`
	Username string   `json:"username"`
	RealName string   `json:"real_name"`
	Roles    []string `json:"roles"`
}
