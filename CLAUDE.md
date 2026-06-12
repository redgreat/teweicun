# TeWeiCun AI 编码指引

> 供 Claude Code 和其他 AI 辅助编码工具读取。通用规则见 `AGENTS.md`，调试记忆见 `MEMORY.md`。

## 项目概况

特维存（TeWeiCun）是面向特种设备/压力容器制造企业的进销存管理系统，覆盖模块：

| 模块 | 功能 |
|------|------|
| 基础数据 | 物料、分类、供应商、客户、仓库、证书、字典 |
| 采购管理 | 采购订单 → 入库确认 → 付款 |
| 入库管理 | 采购入库、生产入库、退货入库、序列号登记 |
| 出库管理 | 销售出库、领料出库、调拨出库、序列号追踪 |
| 库存管理 | 库存明细/汇总/可用量、台账、盘点、预警 |
| 领料/退料 | 领料订单确认出库、退料订单确认入库 |
| 生产管理 | 生产单、生产退货单，关联领料/退料 |
| 销售管理 | 销售订单 → 确认 → 发货 → 退货 |
| 对账管理 | 客户/供应商对账汇总、收付款管理 |
| 追溯管理 | 正向/反向追溯、序列号追溯 |
| 报表系统 | 入库/出库/库存/台账/对账/利润报表 |
| 系统管理 | 用户、角色、权限（RBAC）、通知 |

**架构：DB-First 三层简化**

```text
前端 (SvelteKit)  →  Go 后端 (Gin)  →  PostgreSQL
     展示+交互       路由+鉴权+编排    业务逻辑+事务+状态机
```

- 数据库负责核心业务、状态流转、事务一致性、编号生成、库存流水、报表视图。
- Go 后端负责 API 路由、参数绑定、鉴权、轻量编排、手写 SQL 调用。
- SvelteKit 前端负责业务页面、交互状态、统一 API 调用和展示。

**模块名：** `github.com/redgreat/teweicun`

## 技术栈

| 层 | 技术 | 版本要求 |
|----|------|---------|
| 后端语言 | Go | 1.25+ (go.mod: 1.25.0) |
| Web 框架 | Gin | v1.12.0 |
| 数据库驱动 | pgx/pgxpool | v5.9.1 |
| 日志 | Zap | v1.27.1 |
| 配置 | Viper | v1.21.0 |
| 缓存 | go-redis | v9.18.0 |
| JWT | golang-jwt | v5.3.1 |
| Excel | excelize | v2.10.1 |
| 前端框架 | SvelteKit 2 + Svelte 5 | ^2.57 / ^5.55 |
| 前端语言 | TypeScript | ^6.0 |
| CSS | Tailwind CSS 4 + DaisyUI 5 | ^4.2 / ^5.5 |
| HTTP 客户端 | axios | ^1.15 |
| 图标 | lucide-svelte | ^1.0 |
| 构建工具 | Vite | ^8.0 |
| 数据库 | PostgreSQL | 16+ |

## 项目目录结构

```text
cmd/server/main.go          # 后端入口：配置加载、路由注册、启动
internal/
  config/config.go           # Viper 配置结构
  handler/<module>.go        # HTTP Handler：参数绑定、响应
  service/<module>.go        # 业务编排：参数转换、调用 db 层
  db/<module>.go             # SQL 执行：pgxpool、手写 SQL、过程调用
  dto/request/<module>.go   # 请求 DTO（含 Query、Create、Update）
  dto/response/<module>.go  # 响应 DTO + 通用响应函数
  middleware/                 # JWT、CORS、日志、RBAC、Recovery
  pkg/errcode/               # 统一错误码（AppError）
  pkg/utils/                 # JWT 解析等工具
pkg/
  database/postgres.go       # pgxpool 初始化（全局 Pool）
  cache/redis.go             # Redis 客户端初始化
  logger/logger.go           # Zap 日志初始化
  storage/local.go           # 本地文件存储
configs/                     # YAML 配置文件（config.yaml、config.dev.yaml）
sql/
  schema.sql                 # 数据库结构导出备份（不作为变更入口）
  migrations/NNN_desc.sql    # 递增迁移脚本（唯一变更入口）
frontend/                    # SvelteKit 前端（非 web/）
  src/routes/                # 文件路由
  src/lib/api/client.ts      # 统一 axios 客户端
  src/lib/components/        # 复用组件（DataGrid、Modal、Sidebar、Toast）
  src/lib/store/             # Svelte stores（auth、toast）
test/
  api_flow/                  # 核心全流程接口测试
  testutil/                  # 测试工具函数
scripts/                     # 运维脚本（migrate.sh、schema_manager.go 等）
docs/                        # 设计文档
```

## 必守原则

1. **禁止 ORM**。数据库访问只用 `pgxpool` / `pgx.Tx` + 手写 SQL。go.mod 中虽有 gorm 依赖但代码中未使用，不要引入。
2. **核心业务下沉数据库**。Go 层主要调用：
   - `CALL sp_xxx($1, $2)` — 存储过程（含事务控制）
   - `SELECT fn_xxx($1)` — 函数（编号生成、计算）
   - `SELECT ... FROM v_xxx` — 视图（报表、汇总）
3. **数据库变更只走迁移**。新增 `sql/migrations/` 下递增三位编号文件，不直接改 `sql/schema.sql`。
4. **迁移头部规范**：需含 `-- MIGRATION_ID: <唯一ID>` 和 `-- MIGRATION_APPLIED: pending`，供 `scripts/migrate.sh` 防重执行。部分历史迁移可能缺头，不要在无关任务中批量重写。
5. **新增 DB 对象补中文 COMMENT**。
6. **可空字段处理**：SQL 侧优先 `COALESCE(col, '')`；确有"有或无"语义时用 `*string`、`sql.NullString` 或 `pgtype.Text`。
7. **敏感信息不扩散**。不在日志/文档中暴露数据库密码、Redis 密码、JWT 密钥。
8. **保留文件头风格**：功能、创建时间、创建人。不要为小改动大面积改文件头。

## 后端分层详解

```text
Handler  → internal/handler   参数绑定、校验、调用 service、组装响应
Service  → internal/service   业务编排、参数转换、调用 db（常为透传）
DB       → internal/db        手写 SQL、Scan 映射、调用过程/函数/视图
DTO      → internal/dto       request（入参）/ response（出参）结构
Infra    → pkg/*              database/logger/cache/storage 基础设施
```

### Handler 模式

```go
// internal/handler/xxx.go
func ListXxx(c *gin.Context) {
    var q request.XxxQuery
    if err := c.ShouldBindQuery(&q); err != nil {
        response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, err.Error(), errcode.ErrInvalidParam.HTTPCode))
        return
    }
    if q.Page == 0 { q.Page = 1 }
    if q.PageSize == 0 { q.PageSize = 10 }

    list, total, err := service.ListXxx(c.Request.Context(), &q)
    if err != nil {
        response.Error(c, err)
        return
    }
    response.SuccessPage(c, total, list)
}
```

### Service 模式

```go
// internal/service/xxx.go — 常为透传，复杂场景在此编排
func ListXxx(ctx context.Context, q *request.XxxQuery) ([]response.XxxResp, int64, error) {
    return db.ListXxx(ctx, q)
}
```

### DB 层模式

```go
// internal/db/xxx.go — 手写 SQL、动态条件拼接、Scan 映射
func ListXxx(ctx context.Context, q *request.XxxQuery) ([]response.XxxResp, int64, error) {
    where := []string{"t.deleted_at IS NULL"}
    args := []interface{}{}
    argID := 1
    if q.Keyword != "" {
        where = append(where, fmt.Sprintf("t.name ILIKE $%d", argID))
        args = append(args, "%"+q.Keyword+"%")
        argID++
    }
    // COUNT + SELECT 分页查询
    // 注意：事务控制在存储过程内部，Go 层不手动 BEGIN/COMMIT
}
```

### 统一响应结构

```go
// 后端统一返回 { code: 0, msg: "success", data: ... }
// code=0 表示成功，非 0 表示错误
response.Success(c, data)           // 单条数据
response.SuccessPage(c, total, list) // 分页列表
response.Error(c, appErr)           // 错误（AppError）
```

### 错误码体系

```go
// internal/pkg/errcode/errcode.go
// 系统级: 1000-1999 (InternalServer, InvalidParam, Unauthorized, Forbidden, NotFound)
// 业务级: 2000-9999 (UserNotFound, InvalidPassword, MaterialNotFound, StockNotEnough)
// 用 errcode.NewAppError(code, msg, httpCode) 创建自定义错误
```

### 中间件链

```text
CustomRecovery → CORS → RequestLogger → AuthJWT → (默认分页参数补全)
```

JWT 认证后 `c.Set("user_id", claims.UserID)` 和 `c.Set("username", claims.Username)`，下游通过 `middleware.GetUserID(c)` 获取。

### 分页兼容

main.go 中全局中间件自动补全分页参数：`page` 默认 1，`page_size` 默认 10，同时兼容旧前端传入的 `limit` 参数。

## 前端约定

### 技术要点

- Svelte 5 runes：`$state`、`$derived`、`$effect`、`$props`、`$bindable`
- 适配器：`@sveltejs/adapter-static`（构建为静态文件，由 Go 后端 serve）
- API 客户端：`frontend/src/lib/api/client.ts`，baseURL 默认 `/api/v1`
- 响应解包：后端返回 `{ code: 0, msg, data }`，拦截器在 code===0 时返回 `body.data`
- 认证：JWT 通过 `Authorization: Bearer <token>` 发送，401 时自动跳转 `/login`

### 路由结构

```text
routes/
  +layout.svelte              # 根布局
  +page.svelte                # 首页重定向
  (auth)/login/+page.svelte   # 登录页
  (dashboard)/                # 主布局（含 Sidebar）
    +layout.svelte
    (home)/                   # 首页仪表盘
    (base-data)/              # 基础数据：物料/分类/供应商/客户/仓库
    (purchase)/               # 采购管理
    (stockio)/                # 入库/出库/调拨
    (consumption)/            # 领料/退料
    (ledger)/                 # 库存台账/盘点/预警
    (sales)/                  # （如有）销售管理
    (reconciliation)/         # 对账管理
    (trace)/                  # 追溯查询
    (misc)/                   # 报表等
    (system)/                 # 系统管理：用户/角色/权限
```

### 核心组件

| 组件 | 用途 | 关键 prop |
|------|------|----------|
| `DataGrid.svelte` | 通用列表页 | `bind:page`, `columns`, `data`, `total`, `loading`, `onPageChange` |
| `Modal.svelte` | 模态框 | — |
| `Sidebar.svelte` | 侧边栏导航 | — |
| `ConfirmDialog.svelte` | 确认弹窗 | — |
| `Toast.svelte` | 消息提示 | — |

### DataGrid 使用要点

- 分页 `page` 必须 `$bindable`，父组件重置筛选时要同步 `page = 1`
- `columns` 定义：`{ key, label, sortable?, class?, width? }`
- `cellRender` 可自定义单元格渲染
- `showDefaultSearch` 控制默认搜索框显隐

## 数据库与迁移

### 迁移规则

- 文件名：`sql/migrations/NNN_description.sql`，NNN 为递增三位编号
- 当前最大编号为 `025`（`025_fix_procedures_use_material_serial_tables.sql`）
- 新增迁移头部：
  ```sql
  -- MIGRATION_ID: <唯一描述>
  -- MIGRATION_APPLIED: pending
  ```
- 所有 DDL 优先可重复执行：`IF EXISTS`、`IF NOT EXISTS`、`CREATE OR REPLACE`
- 新增表/列/函数/过程/视图必须补中文 `COMMENT`

### 核心数据库对象命名

| 类型 | 前缀 | 示例 |
|------|------|------|
| 存储过程 | `sp_` | `sp_confirm_stock_in`、`sp_confirm_purchase_order` |
| 函数 | `fn_` | `fn_generate_bill_no`、`fn_generate_serial_no` |
| 视图 | `v_` | `v_inventory_detail`、`v_inventory_summary` |
| 库存流水表 | — | `inventory_transaction` |
| 序列号表 | — | `material_serial_code`、`material_serial_trace` |

### 库存核心语义

| 字段 | 含义 |
|------|------|
| `quantity` | 当前库存量 |
| `locked_quantity` | 已锁定量（已下单未出库） |
| `in_transit_quantity` | 在途量（调拨中） |
| 可用量 | `quantity - locked_quantity` |

## API 路由全景

所有受保护路由统一前缀 `/api/v1`，经 JWT + 分页中间件。

### 基础数据 `/api/v1/base/`
- `GET/POST /categories`、`PUT/DELETE /categories/:id`、`GET /categories/tree`
- `GET/POST /materials`、`GET/PUT /materials/:id`
- `GET/POST /suppliers`、`PUT/DELETE /suppliers/:id`
- `GET/POST /warehouses`、`PUT/DELETE /warehouses/:id`
- `GET/POST /customers`、`PUT/DELETE /customers/:id`
- `GET /partners/dropdown`
- `GET/POST /certificates`、`PUT/DELETE /certificates/:id`

### 采购 `/api/v1/purchase/`
- `GET/POST /orders`、`GET/PUT/DELETE /orders/:id`、`POST /orders/:id/confirm`

### 入库 `/api/v1/stock-in`
- `GET/POST`、`GET /:id`、`PUT /:id`
- `POST /:id/confirm`、`POST /:id/confirm-reversal`
- `PUT /item/:id/serial-selections`、`GET /:id/confirm-logs`

### 出库 `/api/v1/stock-out`
- `GET/POST`、`GET /:id`
- `POST /:id/confirm`
- `PUT /:id/serial-selections`、`PUT /item/:id/serial-selections`

### 领料 `/api/v1/consumption/orders`
- `GET/POST`、`GET/PUT/DELETE /:id`、`POST /:id/confirm`

### 退料 `/api/v1/reversal/orders`
- `GET/POST`、`GET/PUT/DELETE /:id`、`POST /:id/confirm`

### 生产 `/api/v1/production/`
- `GET /orders`、`GET /orders/dropdown`、`GET/PUT /orders/:id`
- `GET /orders/:id/consumption-orders`、`GET /orders/:id/reversal-orders`
- `GET /returns`、`GET /returns/dropdown`、`GET/PUT /returns/:id`
- `GET /returns/:id/consumption-orders`、`GET /returns/:id/reversal-orders`

### 调拨 `/api/v1/stock-transfers`
- `GET/POST`、`GET /:id`、`POST /:id/confirm-out`、`POST /:id/confirm-in`

### 盘点 `/api/v1/inventory-checks`
- `GET/POST`、`GET /:id`、`POST /:id/confirm`

### 库存 `/api/v1/inventory/`
- `GET /detail`、`/summary`、`/available`、`/issued`
- `GET /material-ledger`、`/material-ledger/serials`、`/material-ledger/export`
- `GET /sku-ledger`、`/sku-ledger/serials`、`/sku-ledger/export`
- `GET /alerts`、`POST /alerts/check`

### 销售 `/api/v1/sales/orders`
- `GET/POST`、`GET/PUT /:id`
- `POST /:id/confirm`、`POST /:id/cancel`、`POST /:id/ship`

### 退货 `/api/v1/returns`
- `GET/POST`、`GET/PUT/DELETE /:id`
- `PUT /:id/sales`、`POST /:id/confirm`

### 追溯 `/api/v1/trace/`
- `GET /forward`、`/backward`、`/material/serial`

### 序列号 `/api/v1/serial-codes/` (兼容别名 `/api/v1/sku-serial/`)
- `GET /stock-in-item/:id`、`/stock-in-item/:id/available-issued`
- `GET /stock-out-item/:id`、`/stock-out-item/:id/available`

### 报表 `/api/v1/reports/`
- `GET /stock-in`、`/stock-out`、`/inventory`、`/balance`、`/turnover`
- `GET /reconciliation/customers`、`/reconciliation/suppliers`
- `GET /profit`

### 财务 `/api/v1/fund/`
- `GET/POST /payments`、`GET /payments/:id`、`GET /payment-sources`
- `GET/POST /collections`、`GET /collections/:id`、`GET /collection-sources`

### 系统 `/api/v1/system/`
- `GET/POST /dict-types`、`PUT/DELETE /dict-types/:id`
- `GET /dict/:type/data`、`POST /dict-data`、`PUT/DELETE /dict-data/:id`
- `GET/POST /users`、`PUT/DELETE /users/:id`、`PUT /users/:id/password`、`POST /users/:id/roles`
- `GET/POST /roles`、`GET/PUT/DELETE /roles/:id`、`POST /roles/:id/permissions`
- `GET /permissions/tree`
- `POST /upload`

### 其他
- `GET /notifications`、`POST /notifications/:id/read`
- `GET /dashboard/bigscreen`
- `GET /api/v1/health`（公开，含 DB/Redis 连通性检查）

## 调 bug 路线

从外到内，逐层排查：

1. `cmd/server/main.go` — 确认路由注册和中间件链。
2. `internal/handler/<module>.go` — 参数绑定方式（`ShouldBindQuery` vs `ShouldBindJSON`）、分页默认值。
3. `internal/service/<module>.go` — 业务编排逻辑、参数转换。
4. `internal/db/<module>.go` — 手写 SQL、Scan 类型映射、动态条件拼接。
5. `sql/migrations/` + `sql/schema.sql` — 真实表结构、过程/函数/视图定义。
6. `frontend/src/routes/**/+page.svelte` + `frontend/src/lib/api/client.ts` — 前端请求和展示。

## 高发问题速查

| 现象 | 根因 | 排查方向 |
|------|------|---------|
| `cannot scan NULL into *string` | SQL 可空列未 COALESCE | 查 SELECT 列和 Go Scan 目标类型 |
| 库存数量异常 | 过程未更新 locked/in_transit/quantity | 查 `inventory` + `inventory_transaction` + `sp_confirm_*` |
| 单据状态不对 | 过程内状态机判断有误或 Go 层重复调用 | 查过程的 status 检查、Go 层 operator 传参 |
| 前端翻页错乱 | page 未同步 | 查 `DataGrid bind:page` + 父组件 page 重置 |
| 接口跨域/地址异常 | 前端写了绝对地址 | 默认应走 `/api/v1`，不写 `localhost:8080` |
| 序列号/追溯查询异常 | 序列号表未关联或路径断了 | 查 `material.is_code`、`material_serial_code`、`material_serial_trace` |
| 领料单确认后库存不扣减 | `sp_confirm_stock_out` 内 locked_quantity 未释放 | 查询出库明细 + 库存流水 |
| 生产单关联领料单查不到 | 视图未 join 正确的关联表 | 查 `022_production_order_link_consumption_reversal.sql` |

## 常用命令

```bash
# 后端
go run cmd/server/main.go                       # 默认配置启动
go run cmd/server/main.go -c configs/config.dev.yaml  # 指定配置

# 测试
make test                                        # 后端单测 + 全流程测试
make test-flow                                   # 仅核心流程测试 (TestFlow_PurchaseToReversalAndReturn)

# 代码质量
make lint                                        # Go lint (golangci-lint)
make lint-web                                    # 前端 lint (ESLint + Prettier)
make check-web                                   # 前端 svelte-check

# 联调启动
make local                                       # lint + test 全部通过后启动 Go + Vite
# Windows：开两个终端分别 go run 和 npm run dev

# 前端
cd frontend && npm run dev                       # Vite 开发服务器
cd frontend && npm run build                     # 生产构建

# 数据库
bash scripts/migrate.sh                          # 运行迁移
go run scripts/schema_manager.go export          # 导出 schema.sql
go run scripts/schema_manager.go seed            # 导出种子数据
go run scripts/init_admin_pass.go                # 初始化管理员密码

# 运维
make redeploy                                    # 本地重新部署
make release                                     # 发布版本标签
make build                                       # 编译后端
```

## 开发环境速查

| 工具 | 版本管理 | 当前版本 | 安装位置 |
|------|---------|---------|---------|
| Node.js | nvm (v0.40.3) | v22.22.3 | `~/.nvm/versions/node/` |
| Go | goenv | 1.26.4 | `~/.goenv/versions/1.26.4` |
| 包管理器 | npm | 10.9.8 | nvm 托管 |

```bash
# nvm 常用
nvm use 22              # 切换 Node 22（项目 .nvmrc 指定）
nvm install 24          # 安装新版本
nvm ls                  # 查看已安装版本

# goenv 常用
goenv versions          # 查看已安装版本
goenv global 1.26.4     # 设置全局默认
goenv install 1.25.0    # 安装特定版本
```

## 文档索引

| 文档 | 内容 |
|------|------|
| `README.md` | 项目概览、发布与命令 |
| `docs/requirements.md` | 业务需求 |
| `docs/design.md` | 系统设计（DB-First 架构） |
| `docs/database_design.md` | 数据库设计、PG 可编程对象 |
| `docs/data_flow.md` | 业务数据流、核心流程图 |
| `docs/coding_standards.md` | 代码规范（Go/前端/SQL） |
| `docs/testing_standards.md` | 测试规范（层级/策略/CI） |
| `docs/customer_process_guide.md` | 客户业务流程指南 |
| `AGENTS.md` | 通用 Agent 规则（Codex/Cursor/Trae 等） |
| `MEMORY.md` | 长期项目记忆与高发缺陷 |
