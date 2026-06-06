# TeWeiCun Agent Guide

本文件是仓库级 AI 编码规则，供 Codex、Claude、Cursor、Trae 等工具使用。长期项目记忆见 `MEMORY.md`，Claude 专用入口见 `CLAUDE.md`。

## 先读事实

- 项目：特维存（TeWeiCun），特种设备/压力容器制造企业进销存系统。
- 架构：DB-First。核心业务逻辑在 PostgreSQL，Go 后端只做 API、鉴权、编排和手写 SQL。
- 后端：Go + Gin + pgx/pgxpool + Zap + Viper。
- 前端：`frontend/`，SvelteKit 2 + Svelte 5 + TypeScript + Tailwind CSS 4 + DaisyUI 5。
- 数据库：PostgreSQL 16+，迁移目录 `sql/migrations/`。
- 禁止：ORM、无迁移直接改库、直接改 `sql/schema.sql`、在文档/日志里扩散密钥。

## 工作方式

1. 先看 `git status --short`，不要覆盖用户已有改动。
2. 调 bug 时按链路查：前端页面 -> API client -> 路由 -> handler -> service -> db -> SQL 迁移/过程。
3. 涉及数据库结构、过程、视图、种子数据修复时，新建下一号迁移，例如 `016_fix_xxx.sql`。
4. 新增迁移头部写 `-- MIGRATION_ID: <唯一ID>` 和 `-- MIGRATION_APPLIED: pending`，供 `scripts/migrate.sh` 防重。
5. 每次改 SQL 后检查 Go `Scan` 目标类型，特别是可空字段和 LEFT JOIN 字段。
6. 涉及列表筛选/分页时检查 `page`、`page_size`、前端 `DataGrid bind:page`。
7. 保持改动小而准，不做无关重构。

## 常用命令

```powershell
go run cmd/server/main.go
go run cmd/server/main.go -c configs/config.dev.yaml
make test
make test-flow
make lint
cd frontend; npm run lint
cd frontend; npm run check
```

Windows 下 `make local` 目前只提示不可用；本地联调通常开两个终端跑 Go 和 Vite。

## 后端规则

- 数据访问只用 `pgxpool` / `pgx.Tx` + 参数化 SQL。
- 核心业务流程优先调用数据库对象：
  - `CALL sp_confirm_stock_in($1, $2)`
  - `CALL sp_confirm_stock_out($1, $2)`
  - `CALL sp_confirm_purchase_order($1, $2)`
  - `SELECT fn_generate_serial_no('XX')`
- Go 层可做参数校验、状态预检查、审计调用、响应组装，但不要把复杂多表事务搬出数据库。
- PostgreSQL 错误要转换为业务错误时，优先保留底层信息，便于调试。

## SQL 规则

- 当前迁移命名是三位数字下划线格式：`001_material_only_expand.sql`，新增继续用下一号。
- 新增迁移必须带 `MIGRATION_ID` / `MIGRATION_APPLIED` 头部；部分历史迁移可能缺头，不要在无关任务中批量重写。
- `sql/schema.sql` 是结构导出备份，不作为日常变更入口。
- 新增表、列、函数、过程、视图要补中文 `COMMENT`。
- 可重复执行优先：`IF EXISTS`、`IF NOT EXISTS`、`CREATE OR REPLACE`。
- 查询展示字段优先 `COALESCE`，避免 `cannot scan NULL into *string`。
- 金额/数量逻辑注意 `quantity`、`locked_quantity`、`in_transit_quantity` 的组合语义。

## 前端规则

- 使用 Svelte 5 runes 和现有组件，不要引入 Vue/React 写法。
- API 走 `frontend/src/lib/api/client.ts`，默认 `/api/v1`，响应按 `{ code: 0, msg, data }` 解包。
- UI 组件以 Tailwind/DaisyUI/lucide-svelte 为主。
- `DataGrid` 分页页码支持 `$bindable`，父组件重置筛选时要同步页码。
- 页面文件在 `frontend/src/routes/**/+page.svelte`，复用组件在 `frontend/src/lib/components/`。

## 高发问题备忘

- 新增可空列、LEFT JOIN 展示字段后，最容易触发 Go Scan 空值错误。
- 单据确认类 bug 多发生在数据库过程、状态机和库存锁定/在途量更新。
- 销售、领料、退料、退货会串联库存、出入库单、序列号和流水，查问题时不要只看单表。
- 前端列表“请求第一页但页脚还显示旧页码”通常是没有 `bind:page`。
