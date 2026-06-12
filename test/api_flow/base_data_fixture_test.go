package api_flow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/redgreat/teweicun/test/testutil"
)

var chineseDigits = []rune("甲乙丙丁戊己庚辛壬癸")

const (
	baseSupplierCode = "E2E-SUP"
	baseSupplierName = "华钢供应"
	baseCustomerCode = "E2E-CUS"
	baseCustomerName = "中容客户"
	baseCategoryCode = "E2E-CAT"
	baseCategoryName = "测试板材"
)

type BaseDataFixture struct {
	SupplierCode            string `json:"supplier_code"`
	SupplierName            string `json:"supplier_name"`
	CustomerCode            string `json:"customer_code"`
	CustomerName            string `json:"customer_name"`
	CategoryID              int64  `json:"category_id"`
	CategoryCode            string `json:"category_code"`
	CategoryName            string `json:"category_name"`
	MainMaterialWarehouseID int64  `json:"main_material_warehouse_id"`
	MainMaterialWarehouse   string `json:"main_material_warehouse_code"`
	FinishedWarehouseID     int64  `json:"finished_warehouse_id"`
	FinishedWarehouse       string `json:"finished_warehouse_code"`
	WeldingWarehouseID      int64  `json:"welding_warehouse_id"`
	WeldingWarehouse        string `json:"welding_warehouse_code"`
	AuxiliaryWarehouseID    int64  `json:"auxiliary_warehouse_id"`
	AuxiliaryWarehouse      string `json:"auxiliary_warehouse_code"`
}

func TestSetup_BaseData(t *testing.T) {
	env := testutil.LoadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	admin := testutil.NewClient(env.BaseURL)
	login, err := admin.Login(ctx, env.AdminUser, env.AdminPass)
	if err != nil {
		t.Fatalf("admin login failed: %v", err)
	}

	fixture := ensureBaseDataFixture(ctx, t, admin, login.UserID)
	writeBaseDataFixture(t, fixture)
	t.Logf("基础数据准备完成: supplier=%s customer=%s category=%d main=%s finished=%s welding=%s auxiliary=%s",
		fixture.SupplierCode, fixture.CustomerCode, fixture.CategoryID,
		fixture.MainMaterialWarehouse, fixture.FinishedWarehouse, fixture.WeldingWarehouse, fixture.AuxiliaryWarehouse)
}

func mustLoadBaseDataFixture(ctx context.Context, t *testing.T, c *testutil.Client, managerID int64) BaseDataFixture {
	t.Helper()
	if fixture, ok := readBaseDataFixture(t); ok {
		if fixture.SupplierCode != "" && fixture.CustomerCode != "" && fixture.CategoryID > 0 &&
			fixture.MainMaterialWarehouse != "" && fixture.FinishedWarehouse != "" &&
			fixture.WeldingWarehouse != "" && fixture.AuxiliaryWarehouse != "" {
			return fixture
		}
	}
	fixture := ensureBaseDataFixture(ctx, t, c, managerID)
	writeBaseDataFixture(t, fixture)
	return fixture
}

func ensureBaseDataFixture(ctx context.Context, t *testing.T, c *testutil.Client, managerID int64) BaseDataFixture {
	t.Helper()
	fixture := BaseDataFixture{
		SupplierCode: baseSupplierCode,
		SupplierName: baseSupplierName,
		CustomerCode: baseCustomerCode,
		CustomerName: baseCustomerName,
		CategoryCode: baseCategoryCode,
		CategoryName: baseCategoryName,
	}
	fixture.SupplierCode = ensureBaseSupplier(ctx, t, c)
	fixture.CustomerCode = ensureBaseCustomer(ctx, t, c)
	fixture.CategoryID = ensureBaseCategory(ctx, t, c)
	fixture.MainMaterialWarehouseID, fixture.MainMaterialWarehouse = ensureBaseWarehouse(ctx, t, c, managerID, "WH001", "主材料库", "main_material")
	fixture.FinishedWarehouseID, fixture.FinishedWarehouse = ensureBaseWarehouse(ctx, t, c, managerID, "W0001", "成品库", "finished")
	fixture.WeldingWarehouseID, fixture.WeldingWarehouse = ensureBaseWarehouse(ctx, t, c, managerID, "W0002", "焊材库", "welding")
	fixture.AuxiliaryWarehouseID, fixture.AuxiliaryWarehouse = ensureBaseWarehouse(ctx, t, c, managerID, "W0003", "辅材库", "auxiliary")
	return fixture
}

func ensureBaseSupplier(ctx context.Context, t *testing.T, c *testutil.Client) string {
	t.Helper()
	if code := findSupplierCode(ctx, t, c, baseSupplierCode, baseSupplierName); code != "" {
		return code
	}
	req := map[string]any{
		"supplier_code":        baseSupplierCode,
		"supplier_name":        baseSupplierName,
		"supplier_type":        "manufacturer",
		"contact_person":       "王工",
		"contact_phone":        "13800000000",
		"address":              "浙江",
		"is_qualified":         true,
		"qualification_expire": time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
		"bank_name":            "工行",
		"bank_account":         "6222000000000000",
		"remark":               "自动化基础数据",
	}
	var out struct {
		SupplierCode string `json:"supplier_code"`
	}
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/suppliers", nil, req, &out); err != nil {
		t.Fatalf("创建基础供应商失败: %v", err)
	}
	if out.SupplierCode != "" {
		return out.SupplierCode
	}
	return baseSupplierCode
}

func ensureBaseCustomer(ctx context.Context, t *testing.T, c *testutil.Client) string {
	t.Helper()
	if code := findCustomerCode(ctx, t, c, baseCustomerCode, baseCustomerName); code != "" {
		return code
	}
	req := map[string]any{
		"customer_code":  baseCustomerCode,
		"customer_name":  baseCustomerName,
		"contact_person": "李工",
		"contact_phone":  "13900000000",
		"address":        "江苏",
		"remark":         "自动化基础数据",
	}
	var out struct {
		CustomerCode string `json:"customer_code"`
	}
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/customers", nil, req, &out); err != nil {
		t.Fatalf("创建基础客户失败: %v", err)
	}
	if out.CustomerCode != "" {
		return out.CustomerCode
	}
	return baseCustomerCode
}

func ensureBaseCategory(ctx context.Context, t *testing.T, c *testutil.Client) int64 {
	t.Helper()
	if id := findCategoryID(ctx, t, c, baseCategoryCode, baseCategoryName); id > 0 {
		return id
	}
	req := map[string]any{
		"parent_id":      0,
		"category_code":  baseCategoryCode,
		"category_name":  baseCategoryName,
		"sort_order":     1,
		"category_type":  "material",
		"category_level": 1,
	}
	var out testutil.IDResp
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/categories", nil, req, &out); err != nil {
		t.Fatalf("创建基础物料分类失败: %v", err)
	}
	return out.ID
}

func ensureBaseWarehouse(ctx context.Context, t *testing.T, c *testutil.Client, managerID int64, code, name, warehouseType string) (int64, string) {
	t.Helper()
	if row, ok := findWarehouseByType(ctx, t, c, warehouseType); ok {
		if row.WarehouseName == name && row.WarehouseCode == code {
			return row.ID, row.WarehouseCode
		}
		t.Fatalf("仓库类型 %s 已存在但不是期望基础数据: got code=%s name=%s want code=%s name=%s",
			warehouseType, row.WarehouseCode, row.WarehouseName, code, name)
	}
	req := map[string]any{
		"warehouse_code": code,
		"warehouse_name": name,
		"warehouse_type": warehouseType,
		"manager_id":     managerID,
	}
	var out struct {
		ID            int64  `json:"id"`
		WarehouseCode string `json:"warehouse_code"`
	}
	if err := c.DoJSON(ctx, http.MethodPost, "/api/v1/base/warehouses", nil, req, &out); err != nil {
		t.Fatalf("创建基础仓库失败 type=%s name=%s err=%v", warehouseType, name, err)
	}
	if out.WarehouseCode != "" {
		return out.ID, out.WarehouseCode
	}
	return out.ID, code
}

func findSupplierCode(ctx context.Context, t *testing.T, c *testutil.Client, code, name string) string {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	q.Set("supplier_name", name)
	var list []Supplier
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/base/suppliers", q, &list, nil); err != nil {
		t.Fatalf("查询基础供应商失败: %v", err)
	}
	for _, row := range list {
		if row.SupplierCode == code || row.SupplierName == name {
			return row.SupplierCode
		}
	}
	return ""
}

func findCustomerCode(ctx context.Context, t *testing.T, c *testutil.Client, code, name string) string {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	q.Set("customer_name", name)
	var list []struct {
		CustomerCode string `json:"customer_code"`
		CustomerName string `json:"customer_name"`
	}
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/base/customers", q, &list, nil); err != nil {
		t.Fatalf("查询基础客户失败: %v", err)
	}
	for _, row := range list {
		if row.CustomerCode == code || row.CustomerName == name {
			return row.CustomerCode
		}
	}
	return ""
}

func findCategoryID(ctx context.Context, t *testing.T, c *testutil.Client, code, name string) int64 {
	t.Helper()
	var data []map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/api/v1/base/categories/tree", nil, nil, &data); err != nil {
		t.Fatalf("查询基础分类失败: %v", err)
	}
	var walk func([]map[string]any) int64
	walk = func(nodes []map[string]any) int64 {
		for _, node := range nodes {
			if fmt.Sprint(node["category_code"]) == code || fmt.Sprint(node["category_name"]) == name {
				if id, ok := asInt64(node["id"]); ok {
					return id
				}
			}
			if children, ok := node["children"].([]any); ok {
				nested := make([]map[string]any, 0, len(children))
				for _, child := range children {
					if m, ok := child.(map[string]any); ok {
						nested = append(nested, m)
					}
				}
				if id := walk(nested); id > 0 {
					return id
				}
			}
		}
		return 0
	}
	return walk(data)
}

func findWarehouseByType(ctx context.Context, t *testing.T, c *testutil.Client, warehouseType string) (Warehouse, bool) {
	t.Helper()
	q := url.Values{}
	q.Set("page", "1")
	q.Set("page_size", "100")
	var list []Warehouse
	if err := c.DoPage(ctx, http.MethodGet, "/api/v1/base/warehouses", q, &list, nil); err != nil {
		t.Fatalf("查询基础仓库失败: %v", err)
	}
	var found Warehouse
	for _, row := range list {
		if row.WarehouseType == warehouseType {
			if found.WarehouseCode != "" {
				t.Fatalf("仓库类型重复: type=%s first=%s second=%s", warehouseType, found.WarehouseCode, row.WarehouseCode)
			}
			found = row
		}
	}
	if found.WarehouseCode == "" {
		return Warehouse{}, false
	}
	return found, true
}

func baseDataFixturePath() string {
	if p := os.Getenv("TWC_BASE_DATA_FILE"); p != "" {
		return p
	}
	dir, err := os.Getwd()
	if err != nil {
		return filepath.Join("test", ".runtime", "base_data.json")
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "test", ".runtime", "base_data.json")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return filepath.Join("test", ".runtime", "base_data.json")
		}
		dir = parent
	}
}

func readBaseDataFixture(t *testing.T) (BaseDataFixture, bool) {
	t.Helper()
	raw, err := os.ReadFile(baseDataFixturePath())
	if err != nil {
		return BaseDataFixture{}, false
	}
	var fixture BaseDataFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("读取基础数据 fixture 失败: %v", err)
	}
	return fixture, true
}

func writeBaseDataFixture(t *testing.T, fixture BaseDataFixture) {
	t.Helper()
	path := baseDataFixturePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("创建基础数据 fixture 目录失败: %v", err)
	}
	raw, err := json.MarshalIndent(fixture, "", "  ")
	if err != nil {
		t.Fatalf("序列化基础数据 fixture 失败: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("写入基础数据 fixture 失败: %v", err)
	}
}

func uniqueChineseName(base string) string {
	n := time.Now().UnixNano()
	runes := []rune(base)
	for i := 0; i < 4; i++ {
		runes = append(runes, chineseDigits[n%int64(len(chineseDigits))])
		n /= int64(len(chineseDigits))
	}
	if len(runes) > 10 {
		runes = runes[:10]
	}
	return string(runes)
}
