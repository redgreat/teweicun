//go:build schema_manager
// +build schema_manager

/**
 * 功能：schema_manager.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config 配置结构
type Config struct {
	Database struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		User         string `yaml:"user"`
		Password     string `yaml:"password"`
		DBName       string `yaml:"dbname"`
		SSLMode      string `yaml:"sslmode"`
		PoolMaxConns int    `yaml:"pool_max_conns"`
		PoolMinConns int    `yaml:"pool_min_conns"`
	} `yaml:"database"`
}

// TableInfo 表信息
type TableInfo struct {
	Schema     string
	Name       string
	Type       string // TABLE, VIEW, FUNCTION, PROCEDURE, TRIGGER
	Definition string
	Comment    string
	CreatedAt  time.Time
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "export":
		exportSchema()
	case "diff":
		generateDiff()
	case "seed":
		exportSeedData()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("SQL Schema Manager - 数据库结构管理工具")
	fmt.Println("\n用法:")
	fmt.Println("  schema_manager export    - 导出数据库结构到schema.sql")
	fmt.Println("  schema_manager diff      - 生成DDL差异语句")
	fmt.Println("  schema_manager seed      - 导出seed数据")
	fmt.Println("\n示例:")
	fmt.Println("  go run schema_manager.go export")
}

// loadConfig 加载配置文件
func loadConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// connectDB 连接数据库
func connectDB(config *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.DBName,
		config.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	return db, nil
}

// exportSchema 导出数据库结构
func exportSchema() {
	fmt.Println("=== 导出数据库结构 ===")

	// 加载配置
	config, err := loadConfig("configs/config.yaml")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	// 连接数据库
	db, err := connectDB(config)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 导出表结构
	var output strings.Builder
	output.WriteString("-- =====================================================\n")
	output.WriteString("-- 特维存（TeWeiCun）数据库结构导出\n")
	output.WriteString(fmt.Sprintf("-- 导出时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	output.WriteString("-- PostgreSQL 数据库\n")
	output.WriteString("-- =====================================================\n\n")

	// 1. 导出表结构
	tables, err := exportTables(db)
	if err != nil {
		fmt.Printf("导出表结构失败: %v\n", err)
		os.Exit(1)
	}
	output.WriteString(tables)

	// 2. 导出视图
	views, err := exportViews(db)
	if err != nil {
		fmt.Printf("导出视图失败: %v\n", err)
	}
	output.WriteString(views)

	// 3. 导出函数
	functions, err := exportFunctions(db)
	if err != nil {
		fmt.Printf("导出函数失败: %v\n", err)
	}
	output.WriteString(functions)

	// 4. 导出存储过程
	procedures, err := exportProcedures(db)
	if err != nil {
		fmt.Printf("导出存储过程失败: %v\n", err)
	}
	output.WriteString(procedures)

	// 5. 导出触发器
	triggers, err := exportTriggers(db)
	if err != nil {
		fmt.Printf("导出触发器失败: %v\n", err)
	}
	output.WriteString(triggers)

	// 写入文件
	outputPath := "sql/schema.sql"
	if err := ioutil.WriteFile(outputPath, []byte(output.String()), 0644); err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 数据库结构已导出到: %s\n", outputPath)
}

// exportTables 导出表结构
func exportTables(db *gorm.DB) (string, error) {
	var output strings.Builder
	output.WriteString("-- =====================================================\n")
	output.WriteString("-- 表结构定义\n")
	output.WriteString("-- =====================================================\n\n")

	// 获取所有表
	var tables []string
	db.Raw(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		ORDER BY tablename
	`).Scan(&tables)

	for _, table := range tables {
		// 使用pg_dump风格导出表结构
		var tableDef string
		db.Raw(`
			SELECT 'CREATE TABLE ' || quote_ident($1) || ' (' || E'\n' ||
			string_agg(
				'    ' || quote_ident(a.attname) || ' ' || 
				pg_catalog.format_type(a.atttypid, a.atttypmod) ||
				CASE WHEN attnotnull THEN ' NOT NULL' ELSE '' END ||
				CASE WHEN a.atthasdef THEN ' DEFAULT ' || pg_catalog.pg_get_expr(d.adbin, d.adrelid) ELSE '' END,
				E',\n' ORDER BY a.attnum
			) || E'\n);' as definition
			FROM pg_catalog.pg_attribute a
			LEFT JOIN pg_catalog.pg_attrdef d ON (a.attrelid = d.adrelid AND a.attnum = d.adnum)
			JOIN pg_catalog.pg_class c ON a.attrelid = c.oid
			WHERE c.relname = $1 AND a.attnum > 0 AND NOT a.attisdropped
			GROUP BY a.attrelid
		`, table).Scan(&tableDef)

		if tableDef == "" {
			// 如果上面的查询失败，使用备用方法
			db.Raw(fmt.Sprintf(`
				SELECT 'CREATE TABLE %s (' || E'\n' ||
				string_agg(
					'    ' || quote_ident(a.attname) || ' ' || 
					pg_catalog.format_type(a.atttypid, a.atttypmod) ||
					CASE WHEN attnotnull THEN ' NOT NULL' ELSE '' END ||
					CASE WHEN a.atthasdef THEN ' DEFAULT ' || pg_catalog.pg_get_expr(d.adbin, d.adrelid) ELSE '' END,
					E',\n' ORDER BY a.attnum
				) || E'\n);' as definition
				FROM pg_catalog.pg_attribute a
				LEFT JOIN pg_catalog.pg_attrdef d ON (a.attrelid = d.adrelid AND a.attnum = d.adnum)
				JOIN pg_catalog.pg_class c ON a.attrelid = c.oid
				WHERE c.relname = '%s' AND a.attnum > 0 AND NOT a.attisdropped
				GROUP BY a.attrelid
			`, table, table)).Scan(&tableDef)
		}

		output.WriteString(fmt.Sprintf("-- 表: %s\n", table))
		output.WriteString(fmt.Sprintf("-- 创建时间: %s\n", time.Now().Format("2006-01-02")))
		output.WriteString(fmt.Sprintf("%s\n\n", tableDef))

		// 导出表注释
		var tableComment string
		db.Raw(`
			SELECT obj_description((SELECT oid FROM pg_class WHERE relname = $1), 'pg_class')
		`, table).Scan(&tableComment)
		if tableComment != "" {
			output.WriteString(fmt.Sprintf("COMMENT ON TABLE %s IS '%s';\n\n", table, tableComment))
		}

		// 导出列注释
		var columns []struct {
			ColumnName string
			Comment    string
		}
		db.Raw(`
			SELECT a.attname as column_name, col_description(a.attrelid, a.attnum) as comment
			FROM pg_catalog.pg_attribute a
			JOIN pg_catalog.pg_class c ON a.attrelid = c.oid
			WHERE c.relname = $1 AND a.attnum > 0 AND NOT a.attisdropped
			AND col_description(a.attrelid, a.attnum) IS NOT NULL
			ORDER BY a.attnum
		`, table).Scan(&columns)

		for _, col := range columns {
			output.WriteString(fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s';\n", table, col.ColumnName, col.Comment))
		}
		output.WriteString("\n")

		// 导出索引
		var indexes []string
		db.Raw(`
			SELECT indexdef 
			FROM pg_indexes 
			WHERE tablename = $1 AND schemaname = 'public'
		`, table).Scan(&indexes)
		for _, idx := range indexes {
			output.WriteString(fmt.Sprintf("%s;\n", idx))
		}
		output.WriteString("\n")
	}

	return output.String(), nil
}

// exportViews 导出视图
func exportViews(db *gorm.DB) (string, error) {
	var output strings.Builder
	output.WriteString("-- =====================================================\n")
	output.WriteString("-- 视图定义\n")
	output.WriteString("-- =====================================================\n\n")

	var views []string
	db.Raw(`
		SELECT viewname 
		FROM pg_views 
		WHERE schemaname = 'public' 
		ORDER BY viewname
	`).Scan(&views)

	for _, view := range views {
		var viewDef string
		db.Raw(`
			SELECT definition
			FROM pg_views 
			WHERE viewname = $1 AND schemaname = 'public'
		`, view).Scan(&viewDef)

		output.WriteString(fmt.Sprintf("-- 视图: %s\n", view))
		output.WriteString(fmt.Sprintf("DROP VIEW IF EXISTS %s;\n", view))
		output.WriteString(fmt.Sprintf("CREATE VIEW %s AS %s;\n\n", view, trimTrailingSemicolons(viewDef)))
	}

	return output.String(), nil
}

// exportFunctions 导出函数
func exportFunctions(db *gorm.DB) (string, error) {
	var output strings.Builder
	output.WriteString("-- =====================================================\n")
	output.WriteString("-- 函数定义\n")
	output.WriteString("-- =====================================================\n\n")

	var functions []struct {
		Name       string
		Definition string
	}
	db.Raw(`
		SELECT p.proname as name, 
			   pg_catalog.pg_get_functiondef(p.oid) as definition
		FROM pg_catalog.pg_proc p
		JOIN pg_catalog.pg_namespace n ON p.pronamespace = n.oid
		WHERE n.nspname = 'public' 
		AND p.prokind = 'f'
		ORDER BY p.proname
	`).Scan(&functions)

	for _, fn := range functions {
		output.WriteString(fmt.Sprintf("-- 函数: %s\n", fn.Name))
		def := trimTrailingSemicolons(fn.Definition)
		output.WriteString(fmt.Sprintf("DROP FUNCTION IF EXISTS %s CASCADE;\n", fn.Name))
		output.WriteString(fmt.Sprintf("%s;\n\n", def))
	}

	return output.String(), nil
}

// exportProcedures 导出存储过程
func exportProcedures(db *gorm.DB) (string, error) {
	var output strings.Builder
	output.WriteString("-- =====================================================\n")
	output.WriteString("-- 存储过程定义\n")
	output.WriteString("-- =====================================================\n\n")

	var procedures []struct {
		Name       string
		Definition string
	}
	db.Raw(`
		SELECT p.proname as name, 
			   pg_catalog.pg_get_functiondef(p.oid) as definition
		FROM pg_catalog.pg_proc p
		JOIN pg_catalog.pg_namespace n ON p.pronamespace = n.oid
		WHERE n.nspname = 'public' 
		AND p.prokind = 'p'
		ORDER BY p.proname
	`).Scan(&procedures)

	for _, proc := range procedures {
		output.WriteString(fmt.Sprintf("-- 存储过程: %s\n", proc.Name))
		def := trimTrailingSemicolons(proc.Definition)
		output.WriteString(fmt.Sprintf("DROP PROCEDURE IF EXISTS %s CASCADE;\n", proc.Name))
		output.WriteString(fmt.Sprintf("%s;\n\n", def))
	}

	return output.String(), nil
}

// exportTriggers 导出触发器
func exportTriggers(db *gorm.DB) (string, error) {
	var output strings.Builder
	output.WriteString("-- =====================================================\n")
	output.WriteString("-- 触发器定义\n")
	output.WriteString("-- =====================================================\n\n")

	var triggers []struct {
		Name       string
		Definition string
	}
	db.Raw(`
		SELECT t.tgname as name,
			   'CREATE TRIGGER ' || t.tgname || 
			   ' ' || CASE WHEN t.tgtype & 4 = 4 THEN 'BEFORE' ELSE 'AFTER' END ||
			   ' ' || CASE 
			   		WHEN t.tgtype & 2 = 2 THEN 'INSERT'
			   		WHEN t.tgtype & 4 = 4 THEN 'UPDATE'
			   		WHEN t.tgtype & 8 = 8 THEN 'DELETE'
			   		ELSE 'INSERT OR UPDATE OR DELETE'
			   END ||
			   ' ON ' || c.relname ||
			   ' FOR EACH ' || CASE WHEN t.tgtype & 16 = 16 THEN 'ROW' ELSE 'STATEMENT' END ||
			   ' EXECUTE FUNCTION ' || p.proname || '();' as definition
		FROM pg_catalog.pg_trigger t
		JOIN pg_catalog.pg_class c ON t.tgrelid = c.oid
		JOIN pg_catalog.pg_proc p ON t.tgfoid = p.oid
		WHERE NOT t.tgisinternal
		ORDER BY t.tgname
	`).Scan(&triggers)

	for _, trg := range triggers {
		output.WriteString(fmt.Sprintf("-- 触发器: %s\n", trg.Name))
		output.WriteString(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s;\n", trg.Name, extractTableFromTrigger(trg.Definition)))
		output.WriteString(fmt.Sprintf("%s\n\n", trg.Definition))
	}

	return output.String(), nil
}

// generateDiff 生成DDL差异
func generateDiff() {
	fmt.Println("=== 生成DDL差异 ===")

	// 加载配置
	config, err := loadConfig("configs/config.yaml")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	// 连接数据库
	db, err := connectDB(config)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 读取schema.sql文件
	schemaContent, err := ioutil.ReadFile("sql/schema.sql")
	if err != nil {
		fmt.Printf("读取schema.sql失败: %v\n", err)
		os.Exit(1)
	}

	// 获取数据库中的表
	var dbTables []string
	db.Raw(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		ORDER BY tablename
	`).Scan(&dbTables)

	// 解析schema.sql中的表
	fileTables := parseTablesFromSQL(string(schemaContent))

	// 生成差异
	var diffOutput strings.Builder
	diffOutput.WriteString("-- =====================================================\n")
	diffOutput.WriteString("-- DDL差异语句\n")
	diffOutput.WriteString(fmt.Sprintf("-- 生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	diffOutput.WriteString("-- =====================================================\n\n")

	// 找出新增的表
	for _, table := range fileTables {
		if !contains(dbTables, table) {
			diffOutput.WriteString(fmt.Sprintf("-- 新增表: %s\n", table))
			diffOutput.WriteString(fmt.Sprintf("-- TODO: 请手动添加CREATE TABLE语句\n\n"))
		}
	}

	// 找出删除的表
	for _, table := range dbTables {
		if !contains(fileTables, table) {
			diffOutput.WriteString(fmt.Sprintf("-- 删除表: %s\n", table))
			diffOutput.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;\n\n", table))
		}
	}

	// 写入差异文件
	outputPath := "sql/ddl_diff.sql"
	if err := ioutil.WriteFile(outputPath, []byte(diffOutput.String()), 0644); err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ DDL差异已生成到: %s\n", outputPath)
	fmt.Println("\n请手动检查并执行需要的DDL语句")
}

// parseTablesFromSQL 从SQL内容中解析表名
func parseTablesFromSQL(content string) []string {
	var tables []string
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(strings.ToUpper(line), "CREATE TABLE") {
			// 提取表名
			parts := strings.Fields(line)
			for i, part := range parts {
				if strings.ToUpper(part) == "TABLE" && i+1 < len(parts) {
					tableName := strings.Trim(parts[i+1], "(")
					tables = append(tables, tableName)
					break
				}
			}
		}
	}

	return tables
}

// extractTableFromTrigger 从触发器定义中提取表名
func extractTableFromTrigger(definition string) string {
	// 格式: CREATE TRIGGER ... ON table_name ...
	parts := strings.Fields(definition)
	for i, part := range parts {
		if strings.ToUpper(part) == "ON" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// trimTrailingSemicolons 清理SQL末尾多余分号
func trimTrailingSemicolons(sql string) string {
	return strings.TrimRight(strings.TrimSpace(sql), ";\n\r\t ")
}

// exportSeedData 导出seed数据
func exportSeedData() {
	fmt.Println("=== 导出Seed数据 ===")

	// 加载配置
	config, err := loadConfig("configs/config.yaml")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	// 连接数据库
	db, err := connectDB(config)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	var output strings.Builder
	output.WriteString("-- =====================================================\n")
	output.WriteString("-- 特维存（TeWeiCun）初始化数据\n")
	output.WriteString(fmt.Sprintf("-- 导出时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	output.WriteString("-- =====================================================\n\n")

	// 定义需要导出seed数据的表
	seedTables := []string{
		"sys_dict_type",
		"sys_dict_data",
		"sys_user",
		"sys_role",
		"sys_permission",
		"sys_user_role",
		"sys_role_permission",
	}

	for _, table := range seedTables {
		output.WriteString(fmt.Sprintf("-- 表: %s\n", table))
		output.WriteString(fmt.Sprintf("-- =====================================================\n\n"))

		// 检查表是否存在
		var exists bool
		db.Raw(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' AND table_name = $1
			)
		`, table).Scan(&exists)

		if !exists {
			output.WriteString(fmt.Sprintf("-- 表 %s 不存在\n\n", table))
			continue
		}

		// 导出数据
		var rows []map[string]interface{}
		db.Table(table).Find(&rows)

		if len(rows) == 0 {
			output.WriteString(fmt.Sprintf("-- 表 %s 无数据\n\n", table))
			continue
		}

		// 生成INSERT语句
		for _, row := range rows {
			var columns []string
			var values []string
			for col, val := range row {
				columns = append(columns, col)
				if val == nil {
					values = append(values, "NULL")
				} else {
					values = append(values, fmt.Sprintf("'%v'", val))
				}
			}
			output.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);\n",
				table,
				strings.Join(columns, ", "),
				strings.Join(values, ", ")))
		}
		output.WriteString("\n")
	}

	// 写入文件
	outputPath := "sql/seed_data.sql"
	if err := ioutil.WriteFile(outputPath, []byte(output.String()), 0644); err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Seed数据已导出到: %s\n", outputPath)
}
