# TeWeiCun Memory

## 长期项目记忆

- 项目是特维存，面向特种设备/压力容器制造企业的进销存系统。
- 架构核心是 DB-First：复杂业务、事务、状态流转、库存一致性尽量在 PostgreSQL 存储过程/函数/视图里完成。
- 后端是 Go/Gin/pgx，禁止 ORM；前端是 SvelteKit/Svelte 5，不是 Vue。
- 前端目录是 `frontend/`；README 或旧设计文档里若出现 `web/`，以实际目录为准。
- 数据库迁移现在使用 `sql/migrations/NNN_description.sql`，下一次新增迁移从现有最大编号后递增。
- 新增迁移应带 `-- MIGRATION_ID: <唯一ID>` 和 `-- MIGRATION_APPLIED: pending`；`scripts/migrate.sh` 依赖这个头部做防重。
- 部分历史迁移可能缺头，非迁移治理任务不要顺手批量重写历史 SQL。
- `sql/schema.sql` 是导出备份，不要作为手写修改入口。
- 配置文件里可能包含真实连接信息，回答用户或写文档时不要复述密钥、密码、Token。

## 调 bug 常用路线

- API 路由在 `cmd/server/main.go`。
- 后端链路通常是 `internal/handler` -> `internal/service` -> `internal/db` -> PostgreSQL。
- 前端请求统一经 `frontend/src/lib/api/client.ts`，默认 `/api/v1`，成功响应按 `code === 0` 解包。
- 列表组件重点看 `frontend/src/lib/components/DataGrid.svelte`，分页页码用 `$bindable`。
- 全流程接口测试重点看 `test/api_flow/purchase_consumption_reversal_flow_test.go`。

## 高发缺陷

- Go `Scan` 报 NULL：SQL 可空列未 `COALESCE`，或 DTO/Row 字段不是可空类型。
- 库存异常：查 `inventory.quantity`、`locked_quantity`、`in_transit_quantity`、`inventory_transaction` 和对应 `sp_confirm_*`。
- 确认/取消/发货/退货状态异常：先查数据库过程，再查 Go 是否重复调用或漏传 user/operator。
- 前端分页错位：重置筛选时父组件 page 未同步给 `DataGrid`。
- 线上接口地址异常：检查是否误设 `VITE_API_BASE_URL`，默认应走同源 `/api/v1`。

## 验证偏好

- 小的 Go 改动：优先 `go test -v ./internal/... ./pkg/... ./cmd/...`。
- 核心业务流：跑 `make test-flow`。
- 前端改动：跑 `cd frontend; npm run check`，必要时再 `npm run lint`。
- SQL 改动：先本地执行迁移，再跑相关接口/流程测试。
