/**
 * 功能：测试环境变量读取（API流程测试）
 * 创建时间：2026-04-28
 * 创建人：GPT-5.2
 */

package testutil

import (
	"os"
	"strings"
)

type Env struct {
	BaseURL   string
	AdminUser string
	AdminPass string
}

func LoadEnv() Env {
	baseURL := strings.TrimSpace(os.Getenv("TWC_BASE_URL"))
	if baseURL == "" {
		baseURL = "http://twc.wongcw.cn:8080"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	adminUser := strings.TrimSpace(os.Getenv("TWC_ADMIN_USER"))
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPass := os.Getenv("TWC_ADMIN_PASS")
	if adminPass == "" {
		adminPass = "admin123"
	}

	return Env{
		BaseURL:   baseURL,
		AdminUser: adminUser,
		AdminPass: adminPass,
	}
}

