# TeWeiCun AI 编码指引

> 供 Claude Code 和其他 AI 辅助编码工具读取。更通用的项目级规则见 `AGENTS.md`，长期调试记忆见 `MEMORY.md`。

## 项目概况

特维存（TeWeiCun）是面向特种设备/压力容器制造企业的进销存管理系统，覆盖基础数据、采购、入库、库存、销售、退货、领料、退料、追溯、报表和系统管理。

当前架构是 **DB-First**：

- 数据库负责核心业务流程、状态流转、事务一致性、编号生成、库存流水、报表视图。
- Go 后端负责 API 路由、参数绑定、鉴权、轻量业务编排和手写 SQL 调用。
- SvelteKit 前端负责业务页面、交互状态、统一 API 调用和展示。

## 技术栈与实际目录

- 后端：Go 1.22+、Gin、pgx/pgxpool、Zap、Viper、Redis、local storage。
- 前端：SvelteKit 2、Svelte 5、TypeScript、Vite、Tailwind CSS 4、DaisyUI 5、axios、lucide-svelte。
- 数据库：PostgreSQL 16+，大量使用存储过程、函数、视图、触发器。
- 后端入口：`cmd/server/main.go`。
- 前端目录：`frontend/`，不是 `web/`。
- 数据库脚本：`sql/`，迁移在 `sql/migrations/`。
- 接口测试：`test/`，核心全流程测试在 `test/api_flow/`。

## 必守原则

1. 禁止引入 ORM（GORM/sqlx/ent 等）。数据库访问使用 `pgxpool` 或 `pgx.Tx` + 手写 SQL。
2. 核心业务流程优先落在 PostgreSQL 存储过程/函数/视图中，Go 层通常只调用：
   - `CALL sp_xxx($1, $2)`
   - `SELECT fn_xxx($1)`
   - `SELECT ... FROM v_xxx`
3. 数据库结构或数据修复变更必须新增 `sql/migrations/*.sql`，不要直接改 `sql/schema.sql`。
4. 当前迁移命名使用递增三位编号，例如 `016_fix_xxx.sql`。
5. 新增迁移头部必须包含 `-- MIGRATION_ID: <唯一ID>` 和 `-- MIGRATION_APPLIED: pending`，供 `scripts/migrate.sh` 防重执行。
6. 所有数据库对象和新增列应补中文 `COMMENT`。
7. 可空字符串或 LEFT JOIN 字段扫入 Go `string` 前，SQL 侧优先 `COALESCE(col, '')`；确有“有或无”语义时使用 `*string`、`sql.NullString` 或 `pgtype.Text`。
8. 不要在日志或文档中扩散配置文件里的真实数据库、Redis、JWT 密钥等敏感信息。
9. 源码文件保留项目文件头风格：功能、创建时间、创建人；不要为了小改动大面积改文件头。

## 后端分层

```text
Handler  -> internal/handler   请求解析、参数校验、响应
Service  -> internal/service   业务编排、参数转换
DB       -> internal/db        手写 SQL、调用过程/函数/视图
DTO      -> internal/dto       request/response 结构
Infra    -> pkg/*              database/logger/cache/storage
```

调 bug 时优先沿着这个链路查：

1. `cmd/server/main.go` 找路由。
2. `internal/handler/<module>.go` 看参数绑定和响应。
3. `internal/service/<module>.go` 看业务编排。
4. `internal/db/<module>.go` 看 SQL、Scan 类型、事务、过程调用。
5. `sql/migrations/` 和 `sql/schema.sql` 查真实表结构/过程定义线索。
6. `frontend/src/routes/**/+page.svelte` 和 `frontend/src/lib/api/client.ts` 查前端请求/展示。

注意：部分历史迁移可能缺少 `MIGRATION_ID` 头部，不要在无关任务中批量重写历史迁移；新增迁移必须按当前脚本要求补齐。

## 前端约定

- 使用 Svelte 5 runes：`$state`、`$derived`、`$effect`、`$props`、`$bindable`。
- API 统一走 `frontend/src/lib/api/client.ts`，默认 baseURL 为 `/api/v1`。
- 后端统一响应当前由前端按 `{ code: 0, msg: "success", data }` 解包；不要误写成只接受 HTTP 200 + `{ code: 200 }`。
- 列表页常用 `DataGrid`。重置筛选并回第一页时，必须让父组件页码和 `DataGrid` 内部页码同步，推荐 `bind:page`。
- 按项目现有风格使用 Tailwind/DaisyUI/lucide-svelte，不要改成 Vue/Element Plus 写法。

## 常用命令

```powershell
# 后端启动
go run cmd/server/main.go

# 指定开发配置启动
go run cmd/server/main.go -c configs/config.dev.yaml

# 后端测试 + 全流程接口测试
make test

# 仅跑核心流程测试
make test-flow

# Go lint
make lint

# 前端检查
cd frontend
npm run lint
npm run check
```

Windows 下 `make local` 当前不适用；可开两个终端分别运行后端和 `cd frontend; npm run dev`。

## 调试重点清单

- `cannot scan NULL into *string`：检查 SQL `SELECT` 中的可空列和 `Scan` 目标类型。
- 库存数量异常：同时查库存表、库存流水、对应 `sp_confirm_*` 过程、是否存在锁定量/在途量。
- 单据状态不对：优先查数据库过程中的状态机判断，再查 Go 层是否重复调用或漏传 operator/user。
- 前端列表翻页错乱：查 page/page_size、`limit` 兼容逻辑、`DataGrid bind:page`。
- 线上跨域或接口地址异常：前端默认走相对 `/api/v1`，不要回退到 `localhost:8080`。
- 涉及受压元件/序列号/批次追溯时，同时检查 `material.is_code`、`material_serial` 相关表、出入库明细和追溯查询。

## 文档索引

- `README.md`：项目概览、发布与常用命令。
- `docs/requirements.md`：业务需求。
- `docs/design.md`：系统设计。
- `docs/database_design.md`：数据库设计。
- `docs/data_flow.md`：业务数据流。
- `docs/coding_standards.md`：代码规范。
- `docs/testing_standards.md`：测试规范。
