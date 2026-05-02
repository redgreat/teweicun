/**
 * 功能：API流程测试用的最小响应类型（避免强依赖内部包）
 * 创建时间：2026-04-28
 * 创建人：GPT-5.2
 */

package testutil

// IDResp 通用返回：{id: xxx}
type IDResp struct {
	ID int64 `json:"id"`
}

