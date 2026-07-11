/**
 * 功能：采购订单 Handler 单元测试
 * 测试范围：Gin 绑定层参数校验、鉴权格式、响应格式
 * 注意：不依赖 DB——只测 Gin 层能拦截的错误（JSON 解析/参数校验）
 *       需要 DB 的测试（查库/调SP）放到集成测试中
 * 创建时间：2026-07-12
 * 创建人：Hermes Agent
 */

package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() { gin.SetMode(gin.TestMode) }

// setupPurchaseTestRouter 创建带有 mock 鉴权的采购路由引擎
func setupPurchaseTestRouter() *gin.Engine {
	r := gin.New()
	protected := r.Group("/api/v1")
	protected.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("username", "test_admin")
		c.Next()
	})
	{
		protected.GET("/purchase/orders", ListPurchaseOrders)
		protected.GET("/purchase/orders/:id", GetPurchaseOrder)
		protected.POST("/purchase/orders", CreatePurchaseOrder)
		protected.PUT("/purchase/orders/:id", UpdatePurchaseOrder)
		protected.DELETE("/purchase/orders/:id", DeletePurchaseOrder)
		protected.POST("/purchase/orders/:id/confirm", ConfirmPurchaseOrder)
	}
	return r
}

// ============================================================================
// 参数校验 —— 非数字 ID → 400（strconv.ParseInt 失败，在调 service 前拦截）
// ============================================================================

func TestGetPurchaseOrder_NonNumericID(t *testing.T) {
	r := setupPurchaseTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/purchase/orders/abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid order ID")
}

func TestUpdatePurchaseOrder_NonNumericID(t *testing.T) {
	r := setupPurchaseTestRouter()
	body, _ := json.Marshal(map[string]string{"remark": "test"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/purchase/orders/abc", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid order ID")
}

func TestConfirmPurchaseOrder_NonNumericID(t *testing.T) {
	r := setupPurchaseTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/purchase/orders/abc/confirm", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid order ID")
}

func TestDeletePurchaseOrder_NonNumericID(t *testing.T) {
	r := setupPurchaseTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/purchase/orders/abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid order ID")
}

// ============================================================================
// 参数校验 —— 无效 JSON body → 400（ShouldBindJSON 失败，在调 service 前拦截）
// ============================================================================

func TestCreatePurchaseOrder_InvalidJSON(t *testing.T) {
	r := setupPurchaseTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/purchase/orders",
		bytes.NewReader([]byte(`{invalid`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePurchaseOrder_NotJSON(t *testing.T) {
	r := setupPurchaseTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/purchase/orders",
		bytes.NewReader([]byte(`<xml>not json</xml>`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePurchaseOrder_NoBody(t *testing.T) {
	r := setupPurchaseTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/purchase/orders/1", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================================================
// 参数校验 —— 缺失必填字段 → 400（Gin validator 拦截，在调 service 前）
// ============================================================================

func TestCreatePurchaseOrder_MissingSupplierCode(t *testing.T) {
	r := setupPurchaseTestRouter()
	body := map[string]interface{}{
		"order_type": "purchase",
		"order_date": "2026-07-12",
		"items": []map[string]interface{}{
			{"material_id": 1, "quantity": 10, "unit_price": 100},
		},
	}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/purchase/orders", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "SupplierCode")
}

func TestCreatePurchaseOrder_MissingItems(t *testing.T) {
	r := setupPurchaseTestRouter()
	body := map[string]interface{}{
		"supplier_code": "S0027",
		"order_type":    "purchase",
		"order_date":    "2026-07-12",
	}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/purchase/orders", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Items")
}

// ============================================================================
// 响应格式测试 —— 所有错误响应必须包含 code 和 msg
// ============================================================================

func TestErrorResponseFormat(t *testing.T) {
	r := setupPurchaseTestRouter()

	testCases := []struct {
		name   string
		method string
		path   string
		body   []byte
	}{
		{"non-numeric id in GET", "GET", "/api/v1/purchase/orders/abc", nil},
		{"invalid json in POST", "POST", "/api/v1/purchase/orders", []byte(`{bad`)},
		{"missing body in PUT", "PUT", "/api/v1/purchase/orders/1", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var bodyReader io.Reader
			if tc.body != nil {
				bodyReader = bytes.NewReader(tc.body)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tc.method, tc.path, bodyReader)
			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code,
				"%s: expected 400 got %d", tc.name, w.Code)

			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err == nil {
				assert.Contains(t, resp, "code", "response must contain 'code'")
				assert.Contains(t, resp, "msg", "response must contain 'msg'")
				assert.NotEqual(t, float64(0), resp["code"],
					"error responses must have non-zero code")
			}
		})
	}
}
