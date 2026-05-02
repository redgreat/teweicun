# TeWeiCun 项目指引

> 供 Claude Code 中使用 AI 模型辅助编码时自动加载。

## 项目概述

特维存（TeWeiCun）是面向小型压力容器制造企业的进销存管理系统。  
**架构策略**：DB-First（数据库优先），业务逻辑下沉 PostgreSQL。

## 核心架构原则

- ⚠️ **禁止使用 ORM（GORM/sqlx/ent 等）**，所有数据库操作使用 `pgxpool` + 手写 SQL
- ⚠️ **核心业务逻辑在数据库中实现**：存储过程处理事务、函数处理计算、视图处理报表
- ⚠️ **事务控制在存储过程内部**，Go 层只需 `CALL sp_xxx($1, $2)`
- ⚠️ **单据编号由数据库函数生成**：`SELECT fn_generate_serial_no('PO')`
- ⚠️ **所有数据库对象必须有中文 COMMENT**
- ⚠️ **SQL 脚本在 `sql/` 目录管理**，按类型分 functions/procedures/triggers/views/seed

## 技术栈

- **后端**: Go 1.22+ / Gin / pgx (pgxpool) / Zap / Viper
- **前端**: Vue 3 / Vite / Element Plus / Pinia
- **数据库**: PostgreSQL 16+（存储过程 / 函数 / 触发器 / 视图）
- **缓存**: Redis 7+
- **存储**: MinIO

## 后端分层

```
Handler (internal/handler)  → 请求解析、参数校验、路由
    ↓
Service (internal/service)  → 业务编排、参数转换
    ↓
DB (internal/db)            → 手写SQL、调用存储过程/函数/视图
    ↓
PostgreSQL                  → 存储过程执行事务、触发器自动维护、视图报表
```

## SQL 脚本目录

```
sql/
└── migrations/    DDL 迁移脚本（V001, V002...）
```

## 关键文档

- `docs/requirements.md` — 需求规格
- `docs/design.md` — 系统设计
- `docs/coding_standards.md` — 代码规范
- `docs/testing_standards.md` — 测试规范

## 工作流程

1. **阶段〇先行**：先完成所有数据库对象（DDL + 函数 + 存储过程 + 触发器 + 视图）
2. 然后再写后端 Go 代码，Go 代码调用数据库对象
3. 每个 Step 完成后进行验证
4. 遵循 `docs/coding_standards.md` 的代码规范
5. Git 提交遵循 `type(scope): subject` 格式

## 新增可空列与 Go 查询映射

PostgreSQL 中未设 `NOT NULL`、无默认值、或 LEFT JOIN 产生的列，查询结果可能为 **NULL**。`pgx`/`database/sql` 将 **NULL 扫入 `string` 会报错**（如 `cannot scan NULL into *string`）。

约定：

1. **优先在 SQL 侧收口**：列表/详情 `SELECT` 对字符串展示字段使用 `COALESCE(col, '')`（或业务需要的占位值），与 JSON 中「空串」语义一致时最省事。
2. **或改 DTO 类型**：字段语义确为「有或无」时用 `*string`、`sql.NullString`、`pgtype.Text` 等可空类型，再在组装响应时转成前端需要的形态。
3. **新增列、改 JOIN 后**：自检所有 `Scan`/`QueryRow`/`CollectRows` 目标类型与每一列是否可空，避免只测「有数据」路径。

4. 对数据库的改动都总结一个SQL文件放在sql/migrations文件夹里，不要动sql/schema.sql，这个文件是导出表结构备份用的

5.数据库连接在 configs/config.yaml 里面，每次迭代数据库表结构或者数据时，直接执行就可以了

前端列表若使用「重置筛选 + 拉取第一页」，分页组件内部页码须与父组件同步（如 `DataGrid` 的 `bind:page`），否则会出现页脚页码与请求参数不一致。