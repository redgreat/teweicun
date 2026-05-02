//go:build standardize_comments
// +build standardize_comments

/**
 * 功能：标准化代码文件注释
 * 创建时间：2026-04-18
 * 创建人：wangcw
 *
 * 该脚本用于统一Go和Svelte文件的注释格式，添加标准文件头
 */

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	goHeaderTemplate = `/**
 * 功能：%s
 * 创建时间：%s
 * 创建人：%s
 */

`

	svelteHeaderTemplate = `<!--
功能：%s
创建时间：%s
创建人：%s
-->

`
)

func main() {
	projectRoot := "."
	if len(os.Args) > 1 {
		projectRoot = os.Args[1]
	}

	fmt.Println("开始标准化文件注释...")
	fmt.Printf("项目根目录: %s\n\n", projectRoot)

	goCount := 0
	svelteCount := 0

	filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if strings.Contains(path, "node_modules") ||
				strings.Contains(path, "vendor") ||
				strings.Contains(path, ".git") ||
				strings.Contains(path, "tmp") {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".go" {
			if processGoFile(path) {
				goCount++
			}
		} else if ext == ".svelte" {
			if processSvelteFile(path) {
				svelteCount++
			}
		}

		return nil
	})

	fmt.Printf("\n处理完成！\n")
	fmt.Printf("Go文件: %d 个\n", goCount)
	fmt.Printf("Svelte文件: %d 个\n", svelteCount)
}

func processGoFile(path string) bool {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("读取文件失败: %s - %v\n", path, err)
		return false
	}

	strContent := string(content)

	if strings.Contains(strContent, "/**\n * 功能：") {
		return false
	}

	funcName := extractGoFunction(strContent, path)
	if funcName == "" {
		funcName = filepath.Base(path)
	}

	header := fmt.Sprintf(goHeaderTemplate, funcName, time.Now().Format("2006-01-02"), "wangcw")

	newContent := header + strContent

	if err := ioutil.WriteFile(path, []byte(newContent), 0644); err != nil {
		fmt.Printf("写入文件失败: %s - %v\n", path, err)
		return false
	}

	fmt.Printf("✓ Go: %s\n", path)
	return true
}

func extractGoFunction(content, path string) string {
	baseName := filepath.Base(path)
	
	switch {
	case strings.Contains(baseName, "handler"):
		return extractHandlerFunction(content)
	case strings.Contains(baseName, "service"):
		return extractServiceFunction(content)
	case strings.Contains(baseName, "db"):
		return extractDBFunction(content)
	case strings.Contains(path, "dto/request"):
		return "请求DTO定义"
	case strings.Contains(path, "dto/response"):
		return "响应DTO定义"
	case strings.Contains(baseName, "main.go"):
		return "主程序入口"
	case strings.Contains(baseName, "router.go"):
		return "路由配置"
	case strings.Contains(baseName, "middleware"):
		return "中间件"
	default:
		return baseName
	}
}

func extractHandlerFunction(content string) string {
	re := regexp.MustCompile(`func\s+(\w+)\s*\(`)
	matches := re.FindAllStringSubmatch(content, -1)
	if len(matches) > 0 {
		funcs := make([]string, 0)
		for _, match := range matches {
			if !strings.HasPrefix(match[1], "main") {
				funcs = append(funcs, match[1])
			}
		}
		if len(funcs) > 0 {
			return fmt.Sprintf("HTTP处理器: %s", strings.Join(funcs[:min(3, len(funcs))], ", "))
		}
	}
	return "HTTP处理器"
}

func extractServiceFunction(content string) string {
	re := regexp.MustCompile(`func\s+(\w+)\s*\(`)
	matches := re.FindAllStringSubmatch(content, -1)
	if len(matches) > 0 {
		funcs := make([]string, 0)
		for _, match := range matches {
			if !strings.HasPrefix(match[1], "main") {
				funcs = append(funcs, match[1])
			}
		}
		if len(funcs) > 0 {
			return fmt.Sprintf("业务逻辑: %s", strings.Join(funcs[:min(3, len(funcs))], ", "))
		}
	}
	return "业务逻辑层"
}

func extractDBFunction(content string) string {
	re := regexp.MustCompile(`func\s+(\w+)\s*\(`)
	matches := re.FindAllStringSubmatch(content, -1)
	if len(matches) > 0 {
		funcs := make([]string, 0)
		for _, match := range matches {
			if !strings.HasPrefix(match[1], "main") {
				funcs = append(funcs, match[1])
			}
		}
		if len(funcs) > 0 {
			return fmt.Sprintf("数据库操作: %s", strings.Join(funcs[:min(3, len(funcs))], ", "))
		}
	}
	return "数据库操作层"
}

func processSvelteFile(path string) bool {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("读取文件失败: %s - %v\n", path, err)
		return false
	}

	strContent := string(content)

	if strings.HasPrefix(strings.TrimSpace(strContent), "<!--\n功能：") {
		return false
	}

	pageName := extractSveltePageName(path, strContent)
	header := fmt.Sprintf(svelteHeaderTemplate, pageName, time.Now().Format("2006-01-02"), "wangcw")

	newContent := header + strContent

	if err := ioutil.WriteFile(path, []byte(newContent), 0644); err != nil {
		fmt.Printf("写入文件失败: %s - %v\n", path, err)
		return false
	}

	fmt.Printf("✓ Svelte: %s\n", path)
	return true
}

func extractSveltePageName(path, content string) string {
	if strings.Contains(path, "+page.svelte") {
		dir := filepath.Dir(path)
		parts := strings.Split(dir, string(filepath.Separator))
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] != "" && parts[i] != "routes" && !strings.HasPrefix(parts[i], "(") {
				return fmt.Sprintf("%s页面", parts[i])
			}
		}
	}
	
	if strings.Contains(path, "+layout.svelte") {
		return "布局组件"
	}

	re := regexp.MustCompile(`<title>([^<]+)</title>`)
	if match := re.FindStringSubmatch(content); len(match) > 1 {
		return match[1]
	}

	return filepath.Base(path)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
