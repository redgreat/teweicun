# 自动化测试规范 (Testing Standards)

本文档定义特维存（TeWeiCun）项目在 DB-First 架构下的自动化测试规范和流程。测试是 CI/CD 流程中的重要环节，必须严格遵守以下规范。

## 1. 测试层级与职责分配

鉴于本项目采用 DB-First 架构，业务逻辑大量依赖 PostgreSQL 的存储过程、函数和触发器，测试重点有所偏移：

| 测试层级 | 主要测试目标 | 推荐工具库 | 覆盖率要求 |
| --- | --- | --- | --- |
| **数据库层测试** | 存储过程业务边界、触发器连带更新、函数逻辑正确性 | `pgTAP` / 后端集成测试脚本 | 核心业务过程 100% |
| **后端单元测试** | `db` 层的 SQL 组装（尤其是动态查询）、`service` 层的参数校验、工具函数 | `testing`, `testify(assert/require)` | > 60% |
| **后端 API 测试** | `handler` 层 HTTP 接口规范、鉴权拦截、输入输出格式 | `httptest`, `testify` | > 80% (核心接口 100%) |
| **前端组件测试** | 可复用 UI 组件、自研工具函数、Svelte store | `Vitest`, `@testing-library/svelte`, `MSW`| 重点组件 > 70% |
| **E2E 自动化测试** | 核心业务链路（如：采购申请 -> 订单 -> 质检入库 -> 发货） | `Playwright` / `Cypress` | 覆盖 P0 级核心流程 |

## 2. 数据库逻辑测试 (DB-First 核心)

存储过程和触发器是逻辑核心，必须对其进行严格测试。

### 2.1 测试原则

*   **事务隔离**：数据库测试必须在一个 `BEGIN;` 和 `ROLLBACK;` 之间运行，确保无论测试成功与否，都不会污染数据库环境。
*   **状态验证**：存储过程测试不仅要验证成功执行，必须验证连带状态变更（如调用 `sp_confirm_stock_in` 后，库存确有增加，流水确有记录）。

### 2.2 使用 Go 进行数据库集成测试 (推荐方案)

使用 Go 来编写针对 DB 的测试代码，通过事务来管控环境：

```go
func TestSPConfirmStockIn_Success(t *testing.T) {
    dbPool := setupTestDB()
    ctx := context.Background()
    
    // 开启事务，测试结束后回滚
    tx, err := dbPool.Begin(ctx)
    require.NoError(t, err)
    defer tx.Rollback(ctx)

    // 1. 准备测试数据 (创建相关的 Material, Warehouse, PO 等)
    // 推荐使用一个帮助函数：setupStockInTestData(ctx, tx)返回对应 ID

    // 2. 调用存储过程
    _, err = tx.Exec(ctx, "CALL sp_confirm_stock_in($1, $2)", testStockInID, adminUserID)
    require.NoError(t, err, "存储过程应当执行成功")

    // 3. 断言数据库状态 (查询库存表看是否增加)
    var quantity float64
    err = tx.QueryRow(ctx, "SELECT quantity FROM inventory WHERE material_id=$1", materialID).Scan(&quantity)
    require.NoError(t, err)
    assert.Equal(t, 100.0, quantity, "入库后库存数量不匹配")
}
```

## 3. 后端 Go API 自动化测试

API 测试主要验证外层边界情况、权限验证及路由正确性。

### 3.1 测试目录结构

遵循 Go 的习惯，测试文件与被测文件在同一目录下，以 `_test.go` 结尾。

```text
internal/
  handler/
    purchase_handler.go
    purchase_handler_test.go   // Handler的单元/API测试
  db/
    purchase_db.go
    purchase_db_test.go        // 针对DB操作层的测试
```

### 3.2 Gin 路由测试规范

使用 `httptest` 进行 API 级别测试：

```go
func TestGetPurchaseRequest_Handler(t *testing.T) {
    // 1. Setup Engine & Mocks
    r := gin.Default()
    r.GET("/api/v1/purchase/requests/:id", GetPurchaseRequestHandler)

    // 2. Create Request
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/v1/purchase/requests/1", nil)
    
    // 3. Serve
    r.ServeHTTP(w, req)
    
    // 4. Assert
    assert.Equal(t, http.StatusOK, w.Code)
    
    // 解析 JSON
    var response ResponseBase
    err := json.Unmarshal(w.Body.Bytes(), &response)
    require.NoError(t, err)
    assert.Equal(t, "success", response.Message)
}
```

### 3.3 Mock使用规范

对于外部依赖（甚至底层 DB 操作层），推荐使用接口抽象并利用 `uber-go/mock` (原 `gomock`) 或 `testify/mock` 进行 Mock，以达到并行快速测试的目标。
但对 **核心业务逻辑**（包含复杂SQL和存储过程），不可简单Mock，应进行使用真实的测试数据库的**集成测试**。

## 4. 前端 SvelteKit 测试规范

### 4.1 单元测试 (Vitest)

主要测试：
1. 复杂的派生状态（`$derived`）和工具函数
2. 公共工具库（Utils）
3. Svelte store 的状态逻辑
4. 基础业务组件（不依赖庞大上下文）

```typescript
import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/svelte'
import CustomButton from '../components/CustomButton.svelte'

describe('CustomButton Component', () => {
  it('正确渲染文本内容', () => {
    render(CustomButton, { props: { label: 'Click Me' } })
    expect(screen.getByText('Click Me')).toBeTruthy()
  })
})
```

### 4.2 接口 Mock (MSW)

前端开发和测试中如果需要脱离后端独立运行，规定使用 **MSW (Mock Service Worker)**。所有基础接口（如字典、基础物料全量查询）必须维护一份 MSW Handler 配置。

## 5. CI / CD 集成 (GitHub Actions)

发布与镜像构建已合并为单一工作流：`.github/workflows/release.yml`。

### 5.1 触发方式

- **仅推送 `v*` 版本标签**（如 `v1.0.0`）时触发：构建前端、构建并推送 Docker 镜像（GHCR + 阿里云 ACR）、创建 GitHub Release。
- **普通分支 push / PR**：当前不在 Actions 中自动执行 lint、单测或镜像构建；提交前请在本地执行 `make lint`、`make test`、`make lint-web` 等。

### 5.2 若后续恢复「PR / push 检查」

可在 `.github/workflows/` 下另增独立工作流（例如 `checks.yml`），仅负责 lint/test，与发布工作流分离，避免每次 push 都打包镜像。

### 5.3 测试纪律（本地与将来 CI 通用）

- 严禁通过 `@ts-ignore` 或 `//nolint` 掩盖真实的逻辑错误。
- 严禁通过 `-skip` 参数绕过 GitHub Actions 的测试环节。
- 本地在提 PR 前，必须使用 `make test` 和 `make lint` 确认环境无误。

## 6. 测试数据构造 (Fixtures / Seed)

*   **隔离性**：每个测试用例不能依赖其他用例预留的数据。
*   **Fixture**：建立 `testutil` 包，提供一系列快捷方法，例如 `testutil.CreateMockUser()`, `testutil.CreateMockMaterial()` 以缩减数据搭建篇幅。
*   **清理机制**：用例结束通过 `defer` 或事务回滚清理现场，坚决不允许测试数据残留主库（即使是测试库）。
