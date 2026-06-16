# 特维存 AI 规则引用与业务流程测试计划

> 目的：给 Codex、Claude、Trae、VS Code Copilot 等 AI 工具统一项目上下文，并把核心业务流程拆成可自动化的测试用例。

## 1. AI 规则文件引用结论

### 1.1 当前仓库已有入口

| 文件 | 当前用途 | 建议 |
| --- | --- | --- |
| `AGENTS.md` | 通用 Agent 规则主入口，适合 Codex、GitHub Copilot Coding Agent、部分通用 Agent 工具读取 | 保持为最小但完整的“项目规则源” |
| `CLAUDE.md` | Claude Code 项目记忆入口，内容已扩展为完整项目指引 | 可继续保留；若要减少重复，可在文件开头声明以 `AGENTS.md` 为准 |
| `MEMORY.md` | 长期项目记忆和高发问题备忘 | 适合人工和 AI 显式读取；不建议写敏感信息 |
| `.trae/rules/特维存.md` | Trae 项目规则，`alwaysApply: true` | Trae 中会作为项目规则使用；当前内容可用 |

### 1.2 Trae 是否能引用

`.trae/rules/特维存.md` 是 Trae 规则文件，且头部配置了：

```yaml
---
description: TeWeiCun AI 编码规则
alwaysApply: true
---
```

结论：Trae 可以直接使用这个规则文件。文件中已经说明“根目录 `AGENTS.md` 是通用 AI 规则主入口”，但这更像文本提示，不等于所有工具都会自动展开 `AGENTS.md` 内容。为了稳定，建议 `.trae/rules/特维存.md` 保留关键规则摘要，不只写“请看 AGENTS.md”。

### 1.3 VS Code / Copilot 是否能引用

VS Code 的 Copilot Chat / Agent Customizations 当前支持 always-on instructions，包括：

- `.github/copilot-instructions.md`
- `AGENTS.md`
- `CLAUDE.md`

结论：根目录 `AGENTS.md` 和 `CLAUDE.md` 对 VS Code/Copilot 是可用入口；`.trae/rules/` 不是 VS Code 的标准规则目录，通常不会被 VS Code 自动引用。

建议补一个轻量文件：`.github/copilot-instructions.md`，内容只写：

```md
# Copilot Instructions

请遵循仓库根目录 `AGENTS.md`。重点规则：
- DB-First，核心业务逻辑优先在 PostgreSQL 存储过程/函数/视图。
- Go 后端只用 pgx/pgxpool + 手写参数化 SQL，禁止 ORM。
- 数据库变更只新增 `sql/migrations/NNN_xxx.sql`，不要直接改 `sql/schema.sql`。
- 前端使用 SvelteKit 2 + Svelte 5 + TypeScript + Tailwind/DaisyUI。
- 测试优先参考 `test/api_flow/` 的全流程 API 测试。
```

### 1.4 Claude Code 是否能引用

Claude Code 会读取 `CLAUDE.md`。当前 `CLAUDE.md` 已经足够完整，且明确指向 `AGENTS.md` 和 `MEMORY.md`。

建议：不要让 `AGENTS.md`、`CLAUDE.md`、`.trae/rules/特维存.md` 三份长期发散。更稳的维护方式是：

- `AGENTS.md`：通用主规则。
- `CLAUDE.md`：Claude 专用，保留详细项目地图和常用命令。
- `.trae/rules/特维存.md`：Trae 摘要规则，保留关键约束。
- `MEMORY.md`：只存长期经验、坑点、环境记忆。

## 2. 项目业务流程总览

特维存是压力容器/特种设备制造企业的进销存系统。库存变化全部围绕两类数据：

- `inventory`：实时库存台账，按物料、仓库、批次等维度保存数量。
- `inventory_transaction`：库存流水，记录每次入库、出库、锁定、解锁、调整。

核心动作由数据库过程完成：

- `sp_confirm_purchase_order`：采购订单确认。
- `sp_confirm_stock_in`：普通入库确认。
- `sp_confirm_stock_out`：出库确认。
- `sp_confirm_reversal` / 退料相关过程：退料单据闭环。

### 2.1 主业务链路

```text
基础数据
  -> 采购订单
  -> 采购入库
  -> 库存可用量
  -> 领料出库
  -> 生产入库
  -> 销售订单
  -> 销售出库
  -> 销售退货入库
  -> 对账/报表/追溯
```

### 2.2 子流程

| 流程 | 业务动作 | 库存影响 | 自动化关注点 |
| --- | --- | --- | --- |
| 基础数据 | 分类、供应商、客户、仓库、物料、证书 | 无 | 编码唯一、必填校验、分页查询 |
| 采购 | 创建采购订单、确认订单 | 一般不直接改库存 | 状态从 draft 到 ordered，必要时生成入库单 |
| 入库 | 创建/确认入库单 | `quantity` 增加，写入流水 | 编码物料自动生成序列号 |
| 领料 | 创建领料单、确认、生成出库单、确认出库 | `quantity` 减少，已发数量增加 | 编码物料必须选择序列号 |
| 退料 | 创建退料单、生成入库单、确认退料入库 | `quantity` 回增，序列号回库 | 只能退已发物料 |
| 生产 | 领料时可自动生成生产单和生产入库 | 原料减少，成品增加 | 生产单关联领料单/入库单 |
| 销售 | 创建销售订单、确认锁库、发货、出库确认 | 先锁定，再扣减并释放锁定 | `locked_quantity` 和可用量正确 |
| 销售退货 | 创建销售退货单、确认、生成入库单、确认入库 | 成品库存增加 | 退货数量和关联入库正确 |
| 采购退货 | 创建退货单、生成出库单、确认出库 | 库存减少 | 只能退已有库存 |
| 调拨 | 创建调拨、确认调出、确认调入 | 源仓减少、在途、目标仓增加 | `in_transit_quantity` 正确 |
| 盘点 | 创建盘点单、确认差异 | 库存调整 | 调整流水和账面数量一致 |
| 追溯 | 按序列号/批次/材质证明查询 | 无 | 入库、出库、退库轨迹完整 |
| 对账报表 | 付款、收款、客户/供应商对账、利润 | 无 | 应收应付、实收实付、利润口径 |

## 3. 自动化测试分层

### 3.1 推荐优先级

1. **P0 API 流程测试**：使用 Go `testing` + `test/testutil` 直接调用后端 API。优先覆盖真实业务闭环。
2. **P0 数据库过程断言**：通过 API 完成动作后，查询库存接口、台账接口、追溯接口做结果断言；必要时补 DB 层集成测试。
3. **P1 前端 E2E**：用 Playwright 走登录、列表、创建、确认、查询，验证页面交互和分页。
4. **P2 组件/单元测试**：DataGrid 分页、API client 解包、表单校验等。

### 3.2 已有测试参考

| 文件 | 已覆盖内容 |
| --- | --- |
| `test/api_flow/purchase_consumption_reversal_flow_test.go` | 采购 -> 入库 -> 领料 -> 出库 -> 退料 -> 入库 -> 采购退货 -> 出库，含角色、追溯、台账、大屏 |
| `test/api_flow/purchase_production_sales_cycle_flow_test.go` | 编码物料全流程：采购入库、付款、领料生产、退料、销售出库、收款、销售退库、并发出库 |
| `test/api_flow/reconciliation_summary_report_test.go` | 客户/供应商对账汇总与利润报表接口 |

运行入口：

```bash
make test-flow
go test -v -count=1 ./test/... -run TestFlow_PurchaseProductionSalesCycle
go test -v -count=1 ./test/... -run TestReport_ReconciliationSummariesAndProfit
```

## 4. P0 测试用例清单

### TC-P0-001 基础数据准备

**目标**：创建后续流程所需主数据。

**步骤**：
1. 登录管理员。
2. 创建物料分类。
3. 创建供应商。
4. 创建客户。
5. 创建原料仓、成品仓。
6. 创建编码原料物料、非编码原料物料、编码成品物料。

**断言**：
- 每个创建接口返回 `id` 或 `code`。
- 列表接口可按编码查回数据。
- 分页返回结构为 `{ total, list }`。

### TC-P0-002 采购订单确认与采购入库

**目标**：验证采购到库存增加。

**步骤**：
1. 创建采购订单，包含编码原料和非编码原料。
2. 确认采购订单：`POST /api/v1/purchase/orders/:id/confirm`。
3. 获取采购订单详情，读取或创建关联入库单。
4. 设置入库仓库。
5. 确认入库：`POST /api/v1/stock-in/:id/confirm`。
6. 查询库存可用量：`GET /api/v1/inventory/available`。
7. 查询编码：`GET /api/v1/serial-codes/stock-in-item/:id`。

**断言**：
- 采购订单状态已推进。
- 入库单状态为已确认/通过。
- 库存可用量等于合格入库数量。
- 编码物料生成的序列号数量等于入库数量。

### TC-P0-003 销售订单确认、锁库与销售出库

**目标**：验证销售锁库、发货、出库扣减。

**步骤**：
1. 创建销售订单。
2. 确认销售订单：`POST /api/v1/sales/orders/:id/confirm`。
3. 查询库存可用量和锁定量。
4. 发货生成出库单：`POST /api/v1/sales/orders/:id/ship`。
5. 为编码物料自动选择序列号。
6. 确认出库：`POST /api/v1/stock-out/:id/confirm`。
7. 查询库存可用量和出库单序列号。

**断言**：
- 确认销售订单后，`locked_quantity` 增加，可用量减少。
- 出库确认后，`quantity` 减少，`locked_quantity` 释放。
- 出库单序列号数量等于出库数量。
- 销售订单发货数量和状态正确。

### TC-P0-004 领料出库与生产入库

**目标**：验证生产领料扣原料、自动生成生产单和成品入库。

**步骤**：
1. 从原料仓选择可用库存。
2. 创建领料单，并填写生产成品物料、成品仓、产出数量。
3. 确认领料单：`POST /api/v1/consumption/orders/:id/confirm`。
4. 获取领料单详情，读取 `stock_out_id`、`production_order_id`、`production_stock_in_id`。
5. 为领料出库选择序列号。
6. 确认出库。
7. 如存在生产入库单，查询成品编码。
8. 查询原料仓和成品仓库存。

**断言**：
- 原料库存减少领料数量。
- 成品库存增加产出数量。
- 领料单关联生产单。
- 编码成品生成序列号。

### TC-P0-005 退料入库

**目标**：验证已发原料可以退回库存。

**步骤**：
1. 查询已发库存：`GET /api/v1/inventory/issued`。
2. 创建退料单。
3. 确认退料单：`POST /api/v1/reversal/orders/:id/confirm`。
4. 获取退料单详情，读取 `stock_in_id`。
5. 为退料入库选择已发序列号。
6. 确认退料入库：`POST /api/v1/stock-in/:id/confirm-reversal`。
7. 查询库存可用量。

**断言**：
- 退料入库后原料可用量增加。
- 退回的序列号状态从已发回到在库。
- 退料单和入库单状态正确。

### TC-P0-006 销售退货入库

**目标**：验证销售出库后的成品退货回库。

**步骤**：
1. 创建销售退货单：`POST /api/v1/returns`。
2. 确认退货单：`POST /api/v1/returns/:id/confirm`。
3. 获取退货单详情，读取 `stock_in_id`。
4. 确认退货入库：`POST /api/v1/stock-in/:id/confirm`。
5. 查询成品仓库存。

**断言**：
- 成品库存增加退货数量。
- 退货单生成入库单。
- 入库流水存在。

### TC-P0-007 采购付款、销售收款与对账报表

**目标**：验证采购/销售资金闭环和报表接口可用。

**步骤**：
1. 基于采购订单创建付款：`POST /api/v1/fund/payments`。
2. 基于销售订单创建收款：`POST /api/v1/fund/collections`。
3. 查询付款和收款详情。
4. 查询供应商对账：`GET /api/v1/reports/reconciliation/suppliers`。
5. 查询客户对账：`GET /api/v1/reports/reconciliation/customers`。
6. 查询利润报表：`GET /api/v1/reports/profit`。

**断言**：
- 付款/收款状态为 completed。
- 对账汇总包含对应供应商/客户。
- 应付、已核销、余额、实付/实收字段口径正确。
- 利润报表接口返回金额字段且不报错。

### TC-P0-008 追溯查询

**目标**：验证编码物料全链路可追溯。

**步骤**：
1. 取采购入库生成的序列号。
2. 查询序列号追溯：`GET /api/v1/trace/material/serial`。
3. 按正向/反向追溯接口查询。
4. 对经历过入库、出库、退料、销售退货的序列号分别查询。

**断言**：
- 返回 `serial_info`。
- `traces` 至少包含入库动作。
- 出库后包含出库动作。
- 退料/退货后包含回库动作。

### TC-P0-009 并发出库确认

**目标**：验证同一物料多订单并发出库时库存一致。

**步骤**：
1. 创建编码物料并入库 20 个。
2. 创建 5 个销售订单，每单 2 个。
3. 每单确认、发货、选择序列号。
4. 并发调用 5 个出库单确认接口。
5. 查询最终库存。

**断言**：
- 所有确认接口成功。
- 最终可用量为 `20 - 5*2 = 10`。
- 无重复序列号出库。
- 库存不为负。

## 5. P1 测试用例清单

| 用例 | 目标 | 关键断言 |
| --- | --- | --- |
| TC-P1-001 库存不足销售确认 | 销售数量大于可用库存 | 接口返回业务错误，库存不变 |
| TC-P1-002 重复确认入库单 | 已确认入库单再次确认 | 返回错误或幂等拒绝，库存不重复增加 |
| TC-P1-003 重复确认出库单 | 已确认出库单再次确认 | 库存不重复扣减 |
| TC-P1-004 编码物料未选序列号出库 | 编码物料出库前不选择编码 | 确认失败 |
| TC-P1-005 退料数量超过已发数量 | 超量退料 | 创建或确认失败 |
| TC-P1-006 销售退货数量超过已发数量 | 超量退货 | 创建或确认失败 |
| TC-P1-007 调拨完整流程 | 源仓调出、目标仓调入 | 源仓减少、目标仓增加、在途量归零 |
| TC-P1-008 盘点调整 | 盘点差异确认 | 库存调整数量和流水一致 |
| TC-P1-009 列表筛选分页 | 翻页后重置筛选 | 前端页码回到 1，后端 page/page_size 正确 |
| TC-P1-010 权限拦截 | 普通用户访问无权限接口 | 返回 401/403，不产生数据 |

## 6. 给 AI 生成自动化脚本的提示词模板

### 6.1 生成 Go API 流程测试

```text
请基于 docs/ai_rules_and_business_test_plan.md 的 TC-P0-00X，
参考 test/api_flow/purchase_production_sales_cycle_flow_test.go 和 test/testutil/client.go，
新增或扩展 Go API 流程测试。

要求：
1. 使用 testutil.NewClient、Login、DoJSON、DoPage。
2. 测试数据使用 UniquePrefix，避免污染已有数据。
3. 不要直连数据库，优先通过 API 验证；确需查 DB 时说明原因。
4. 断言库存数量、单据状态、序列号数量、对账金额。
5. 不要复述配置中的密码或密钥。
6. 运行命令：go test -v -count=1 ./test/... -run <测试名>。
```

### 6.2 生成 Playwright 前端 E2E

```text
请基于 docs/ai_rules_and_business_test_plan.md 的 TC-P0-00X，
为 frontend 新增 Playwright E2E 测试。

要求：
1. 先登录，再通过页面完成创建/确认/查询动作。
2. 不要绕过页面直接调用 API，除非用于测试数据准备。
3. 重点检查表格分页、Toast、确认弹窗、状态字段、库存数量展示。
4. 前端技术栈是 SvelteKit 2 + Svelte 5，不要引入 React/Vue 写法。
5. 若页面缺少稳定选择器，先提出需要添加 data-testid 的位置。
```

## 7. 建议执行顺序

1. 先稳定 `TC-P0-001` 到 `TC-P0-003`：基础数据、采购入库、销售出库。
2. 再跑通 `TC-P0-004` 到 `TC-P0-006`：生产、退料、销售退货。
3. 接着补 `TC-P0-007` 和 `TC-P0-008`：对账报表、追溯。
4. 最后跑 `TC-P0-009`：并发出库，验证库存一致性。
5. P0 全部稳定后，再逐条补 P1 异常路径。

## 8. 当前缺口

- `make test-flow` 当前只运行 `TestFlow_PurchaseToReversalAndReturn`，没有覆盖 `TestFlow_PurchaseProductionSalesCycle`。建议后续扩展 Makefile 或新增独立目标。
- 报表测试目前只验证接口可调用，金额口径断言较弱。
- 前端 E2E 暂未看到现成 Playwright 测试目录，需要先确认是否要引入 Playwright。
- `.github/copilot-instructions.md` 目前不存在；若团队用 VS Code/Copilot，建议补齐。

## 9. 基础资料新增专项测试计划

### 9.1 范围

- 只测新增：客户、供应商、物料分类、物料。
- 每项新增后立即查询，确认数据已落库。
- 物料测试需填写较完整属性，避免只测最小字段。

### 9.2 执行方法

1. 使用管理员账号登录系统。
2. 按顺序执行：物料分类 -> 客户 -> 供应商 -> 物料。
3. 每次保存成功后，立即进入列表或树查询。
4. 用名称或编码搜索，确认能查到刚新增的数据。
5. 若页面有详情抽屉或编辑页，再核对关键字段是否一致。

### 9.3 测试数据建议

| 类型 | 建议值 |
| --- | --- |
| 客户编码 | `AT_CUS_001` |
| 客户名称 | `测试客户一` |
| 供应商编码 | `AT_SUP_001` |
| 供应商名称 | `测试供应商一` |
| 分类编码 | `AT_CAT_001` |
| 分类名称 | `测试物料分类一` |
| 物料名称 | `Q345R钢板测试料` |

### 9.4 用例清单

| 用例 | 新增内容 | 操作 | 验证 |
| --- | --- | --- | --- |
| TC-BASE-ADD-001 | 客户 | 填编码、名称、联系人、电话、地址、备注后保存 | 客户列表可按名称或编码查到，名称与联系人一致 |
| TC-BASE-ADD-002 | 供应商 | 填编码、名称、类型、联系人、电话、地址、是否合格、资质到期、开户行、账号、备注后保存 | 供应商列表可查到，类型和合格状态正确 |
| TC-BASE-ADD-003 | 物料分类 | 填上级、分类编码、分类名称、排序后保存 | 分类树可看到新节点，名称与编码正确 |
| TC-BASE-ADD-004 | 物料 | 选择分类，填写名称、单位、安全库存、最大库存、是否编码、备注和扩展属性后保存 | 物料列表可查到，分类、编码属性、库存阈值和扩展属性正确 |

### 9.5 物料属性建议

物料新增时，扩展属性至少填写以下内容：

- 材质：`Q345R`
- 规格：`20mm`
- 宽度：`2000mm`
- 长度：`6000mm`
- 执行标准：`GB/T 713`
- 炉批号：`HEAT-001`
- 计量单位：`kg`
- 重量系数：`1.000`
- 生产厂家：`测试钢厂`
- 备注：`基础资料新增测试`

### 9.6 成功标准

- 新增接口返回成功，无报错提示。
- 列表或分类树中能查到新增数据。
- 查询结果中的名称、编码、联系人或属性与录入值一致。
- 物料的扩展属性数量不少于 8 个。
- 重复刷新页面后，新增数据仍能查到。

### 9.7 可落地的 API 执行参考

如需走 API 验证，可按以下接口执行：

- 客户新增：`POST /api/v1/base/customers`
- 客户查询：`GET /api/v1/base/customers`
- 供应商新增：`POST /api/v1/base/suppliers`
- 供应商查询：`GET /api/v1/base/suppliers`
- 物料分类新增：`POST /api/v1/base/categories`
- 物料分类查询：`GET /api/v1/base/categories/tree`
- 物料新增：`POST /api/v1/base/materials`
- 物料查询：`GET /api/v1/base/materials`
