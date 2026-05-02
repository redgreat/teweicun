# 特维存（TeWeiCun）— 代码规范

> **文档版本**：v2.0（DB-First 架构）  
> **适用范围**：项目所有开发人员和 AI 辅助编码工具  
> **核心原则**：业务逻辑下沉数据库，后端使用手写 SQL，不使用 ORM

---

## 一、Go 后端代码规范

### 1.1 项目结构规范

```
遵循标准 Go 项目布局：
- cmd/          应用入口
- internal/     私有代码（不可被外部导入）
- pkg/          可导出的公共库
- sql/          数据库 SQL 脚本（⭐ 核心）
- configs/      配置文件
```

**分层职责（DB-First 简化三层）**：

| 层级 | 包名 | 职责 | 允许依赖 |
|------|------|------|---------|
| Handler | `internal/handler` | 请求解析、参数校验、调用 Service、响应组装 | DTO、Service |
| Service | `internal/service` | 业务编排、参数转换、调用 DB 层 | DTO、DB |
| DB | `internal/db` | 手写 SQL 执行、调用存储过程/函数/视图 | pgxpool |
| DTO | `internal/dto` | 请求/响应数据传输对象 | 无 |
| SQL 脚本 | `sql/` | DDL、存储过程、函数、触发器、视图、种子数据 | — |

> **注意**：没有 `model/` 和 `repository/` 层。数据结构直接在 `db/` 层定义行扫描结构体，或在 `dto/response/` 中定义。
>
> **事务控制在数据库层完成**：核心业务流程（入库/出库/审批/盘点等）通过调用存储过程执行，存储过程内部管理事务，Go 层不需要手动 BEGIN/COMMIT。

### 1.2 命名规范

```go
// ✅ 文件名：小写下划线
material_svc.go
material_db.go           // 数据库操作文件以 _db.go 结尾
purchase_db.go

// ✅ 结构体/接口：大驼峰
type MaterialService struct {}
type MaterialRow struct {} // 数据库行扫描用结构体以 Row 结尾

// ✅ 方法/导出函数：大驼峰
func (s *MaterialService) GetByID(ctx context.Context, id int64) (*model.Material, error) {}

// ✅ 私有函数/变量：小驼峰
func buildQueryConditions(query *dto.MaterialListReq) {}
var defaultPageSize = 20

// ✅ 常量
const MaxPageSize = 100
const StatusEnabled = "enabled"

// ✅ 接口命名：动词/形容词 + er 或者直接用名词
type Reader interface {}          // 标准 Go 风格
type MaterialRepository interface {} // 业务领域风格

// ✅ 错误变量：Err 前缀
var ErrMaterialNotFound = errors.New("material not found")
var ErrInsufficientStock = errors.New("insufficient stock")

// ❌ 不要使用
var Material_Service struct {} // 下划线
func get_by_id() {}           // 下划线
const max_page_size = 100     // 常量小写
```

### 1.3 函数签名规范

```go
// ✅ Handler 方法：接收 *gin.Context
func (h *MaterialHandler) Create(c *gin.Context) {
    var req dto.CreateMaterialReq
    if err := c.ShouldBindJSON(&req); err != nil {
        Error(c, errcode.ErrBadRequest.WithDetail(err.Error()))
        return
    }

    result, err := h.materialSvc.Create(c.Request.Context(), &req)
    if err != nil {
        handleServiceError(c, err)
        return
    }

    Success(c, result)
}

// ✅ Service 方法：第一个参数是 context.Context，返回值最后是 error
func (s *MaterialService) Create(ctx context.Context, req *dto.CreateMaterialReq) (*dto.MaterialResp, error) {
    // ...
}

// ✅ DB 层函数：接收 pgxpool.Pool + context.Context，直接操作数据库
func FindMaterialByID(ctx context.Context, pool *pgxpool.Pool, id int64) (*MaterialRow, error) {
    // ...
}

// ✅ 多返回值：数据在前，error 在后
func FindMaterialsByPage(ctx context.Context, pool *pgxpool.Pool, query *MaterialQuery, page, size int) ([]MaterialRow, int64, error) {
    // ...
}
```

### 1.4 错误处理规范

```go
// ✅ 使用项目定义的错误码系统
func (s *MaterialService) Create(ctx context.Context, req *dto.CreateMaterialReq) (*dto.MaterialResp, error) {
    // 业务逻辑校验
    existing, err := s.repo.FindByCode(ctx, req.MaterialCode)
    if err != nil {
        return nil, fmt.Errorf("查询物料编码失败: %w", err)  // 包装底层错误
    }
    if existing != nil {
        return nil, errcode.ErrConflict.WithMsg("物料编码已存在")  // 使用业务错误码
    }
    
    // ...
}

// ✅ Handler 中统一错误转换
func handleServiceError(c *gin.Context, err error) {
    var appErr *errcode.AppError
    if errors.As(err, &appErr) {
        Error(c, appErr)
        return
    }
    logger.Error("internal error", zap.Error(err))
    Error(c, errcode.ErrInternalServer)
}

// ❌ 不要忽略错误
result, _ := s.repo.FindByID(ctx, id)  // 错误，_ 会隐藏问题

// ❌ 不要 panic
if err != nil {
    panic(err)  // 禁止在业务代码中使用 panic
}
```

### 1.5 数据库操作规范（手写 SQL + pgx）

> ⚠️ 本项目**禁止使用 ORM**（包括 GORM、sqlx 等），所有数据库操作使用 `pgxpool` + 手写 SQL。

```go
// ✅ 查询单条记录
func FindMaterialByID(ctx context.Context, pool *pgxpool.Pool, id int64) (*MaterialRow, error) {
    var m MaterialRow
    err := pool.QueryRow(ctx, `
        SELECT m.id, m.material_code, m.material_name, m.specification,
               m.category_id, mc.category_name, m.material_grade, m.standard_no,
               m.unit, m.is_pressure_part, m.safety_stock, m.max_stock, m.status,
               m.created_at
        FROM material m
        LEFT JOIN material_category mc ON mc.id = m.category_id
        WHERE m.id = $1 AND m.deleted_at IS NULL
    `, id).Scan(
        &m.ID, &m.MaterialCode, &m.MaterialName, &m.Specification,
        &m.CategoryID, &m.CategoryName, &m.MaterialGrade, &m.StandardNo,
        &m.Unit, &m.IsPressurePart, &m.SafetyStock, &m.MaxStock, &m.Status,
        &m.CreatedAt,
    )
    if err == pgx.ErrNoRows {
        return nil, nil
    }
    return &m, err
}

// ✅ 查询列表（动态 WHERE 拼接）
func FindMaterialsByPage(ctx context.Context, pool *pgxpool.Pool, q *MaterialQuery, page, size int) ([]MaterialRow, int64, error) {
    // 构建 WHERE 条件
    where := "m.deleted_at IS NULL"
    args := []interface{}{}
    argIdx := 1

    if q.MaterialName != "" {
        where += fmt.Sprintf(" AND m.material_name LIKE $%d", argIdx)
        args = append(args, "%"+q.MaterialName+"%")
        argIdx++
    }
    if q.CategoryID > 0 {
        where += fmt.Sprintf(" AND m.category_id = $%d", argIdx)
        args = append(args, q.CategoryID)
        argIdx++
    }
    if q.Status != "" {
        where += fmt.Sprintf(" AND m.status = $%d", argIdx)
        args = append(args, q.Status)
        argIdx++
    }

    // 查总数
    var total int64
    countSQL := "SELECT COUNT(*) FROM material m WHERE " + where
    pool.QueryRow(ctx, countSQL, args...).Scan(&total)

    // 查分页数据
    offset := (page - 1) * size
    dataSQL := fmt.Sprintf(`
        SELECT m.id, m.material_code, m.material_name, m.specification,
               mc.category_name, m.unit, m.status, m.created_at
        FROM material m
        LEFT JOIN material_category mc ON mc.id = m.category_id
        WHERE %s
        ORDER BY m.created_at DESC
        LIMIT $%d OFFSET $%d
    `, where, argIdx, argIdx+1)
    args = append(args, size, offset)

    rows, err := pool.Query(ctx, dataSQL, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var materials []MaterialRow
    for rows.Next() {
        var m MaterialRow
        rows.Scan(&m.ID, &m.MaterialCode, &m.MaterialName, &m.Specification,
            &m.CategoryName, &m.Unit, &m.Status, &m.CreatedAt)
        materials = append(materials, m)
    }
    return materials, total, nil
}

// ✅ 插入数据
func CreateMaterial(ctx context.Context, pool *pgxpool.Pool, m *CreateMaterialParams) (int64, error) {
    var id int64
    err := pool.QueryRow(ctx, `
        INSERT INTO material (material_code, material_name, specification, category_id,
                              material_grade, standard_no, unit, is_pressure_part,
                              safety_stock, max_stock, created_by)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING id
    `, m.MaterialCode, m.MaterialName, m.Specification, m.CategoryID,
       m.MaterialGrade, m.StandardNo, m.Unit, m.IsPressurePart,
       m.SafetyStock, m.MaxStock, m.CreatedBy,
    ).Scan(&id)
    return id, err
}

// ✅ 调用存储过程（核心业务流程）
func ConfirmStockIn(ctx context.Context, pool *pgxpool.Pool, stockInID, operatorID int64) error {
    // 事务控制在存储过程内部完成，Go 层只需 CALL
    _, err := pool.Exec(ctx, `CALL sp_confirm_stock_in($1, $2)`, stockInID, operatorID)
    if err != nil {
        // 解析 PG 异常为业务错误
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) {
            return errcode.ErrBusinessLogic.WithMsg(pgErr.Message)
        }
    }
    return err
}

// ✅ 调用数据库函数
func GenerateSerialNo(ctx context.Context, pool *pgxpool.Pool, prefix string) (string, error) {
    var serialNo string
    err := pool.QueryRow(ctx, `SELECT fn_generate_serial_no($1)`, prefix).Scan(&serialNo)
    return serialNo, err
}

// ✅ 查询视图（用于报表）
func GetInventoryAlerts(ctx context.Context, pool *pgxpool.Pool) ([]AlertRow, error) {
    rows, err := pool.Query(ctx, `SELECT * FROM v_inventory_alert`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    // ... 遍历结果
}

// ✅ 调用返回表的函数（追溯查询）
func TraceForward(ctx context.Context, pool *pgxpool.Pool, heatNo string) ([]TraceRow, error) {
    rows, err := pool.Query(ctx, `SELECT * FROM fn_trace_forward($1)`, heatNo)
    // ...
}

// ❌ 禁止使用 ORM
// ❌ 禁止使用 GORM、sqlx、ent 等 ORM 框架
// ❌ 禁止在 Go 层编写复杂事务（应在存储过程中完成）
// ❌ 禁止在循环中执行单条数据库操作（N+1 问题）
```

### 1.6 数据结构定义规范（DB Row 结构体）

```go
// ✅ 数据库行扫描结构体（放在 internal/db/ 包中）
// 不使用 gorm tag，使用 json tag 用于 API 响应
type MaterialRow struct {
    ID             int64   `json:"id"`
    MaterialCode   string  `json:"material_code"`
    MaterialName   string  `json:"material_name"`
    Specification  string  `json:"specification"`
    CategoryID     int64   `json:"category_id"`
    CategoryName   *string `json:"category_name"`    // JOIN 查询带出，可能为 NULL
    MaterialGrade  *string `json:"material_grade"`   // 可选字段用指针
    StandardNo     *string `json:"standard_no"`
    Unit           string  `json:"unit"`
    IsPressurePart bool    `json:"is_pressure_part"`
    SafetyStock    float64 `json:"safety_stock"`     // NUMERIC 用 float64 或 pgtype.Numeric
    MaxStock       float64 `json:"max_stock"`
    Status         string  `json:"status"`
    CreatedAt      time.Time `json:"created_at"`
}

// ✅ 数据库写入参数结构体
type CreateMaterialParams struct {
    MaterialCode   string
    MaterialName   string
    Specification  string
    CategoryID     int64
    MaterialGrade  *string
    StandardNo     *string
    Unit           string
    IsPressurePart bool
    SafetyStock    float64
    MaxStock       float64
    CreatedBy      int64
}

// ✅ 查询条件结构体
type MaterialQuery struct {
    MaterialName string
    MaterialCode string
    CategoryID   int64
    Status       string
    MaterialGrade string
}

// ✅ 可选字段（数据库 NULL）使用指针类型 *string, *int64, *time.Time
// ✅ NUMERIC 字段可根据精度需求使用 float64 或 github.com/shopspring/decimal
// ✅ JSONB 字段使用 json.RawMessage 或 map[string]interface{}
```

### 1.7 DTO 定义规范

```go
// ✅ 请求 DTO：使用 binding tag 做校验
type CreateMaterialReq struct {
    MaterialCode   string `json:"material_code" binding:"required,min=1,max=30"`
    MaterialName   string `json:"material_name" binding:"required,min=1,max=100"`
    Specification  string `json:"specification" binding:"required,min=1,max=200"`
    CategoryID     int64  `json:"category_id" binding:"required,gt=0"`
    MaterialGrade  string `json:"material_grade"`   // 业务校验在 Service 层
    Unit           string `json:"unit" binding:"required"`
    IsPressurePart bool   `json:"is_pressure_part"`
}

// ✅ 列表请求 DTO
type MaterialListReq struct {
    Page         int    `form:"page" binding:"omitempty,min=1"`
    PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
    MaterialName string `form:"material_name"`
    MaterialCode string `form:"material_code"`
    CategoryID   int64  `form:"category_id"`
    Status       string `form:"status"`
    Sort         string `form:"sort" binding:"omitempty,oneof=created_at material_code material_name"`
    Order        string `form:"order" binding:"omitempty,oneof=asc desc"`
}

// 提供默认值
func (r *MaterialListReq) SetDefaults() {
    if r.Page <= 0 {
        r.Page = 1
    }
    if r.PageSize <= 0 {
        r.PageSize = 20
    }
}

// ✅ 响应 DTO：只返回需要的字段
type MaterialResp struct {
    ID             int64  `json:"id"`
    MaterialCode   string `json:"material_code"`
    MaterialName   string `json:"material_name"`
    Specification  string `json:"specification"`
    CategoryName   string `json:"category_name"`     // 关联数据展开
    MaterialGrade  string `json:"material_grade"`
    Unit           string `json:"unit"`
    IsPressurePart bool   `json:"is_pressure_part"`
    Status         string `json:"status"`
    CreatedAt      string `json:"created_at"`         // 格式化后的时间字符串
}
```

### 1.8 注释规范

```go
// ✅ 包注释
// Package service 提供业务编排层的实现。
// 参数校验和简单逻辑在此层处理，核心业务流程委托给数据库存储过程。
package service

// ✅ 结构体注释
// MaterialService 提供物料管理的业务编排。
type MaterialService struct {
    pool *pgxpool.Pool  // 数据库连接池
    cache *redis.Client // Redis 缓存
}

// ✅ 方法注释（导出方法必须有注释）
// Create 创建新物料。
// 业务规则（参数校验在此）：
//   - 物料编码全局唯一（通过 INSERT 返回错误判断）
//   - 受压元件材料必须填写材质牌号和标准号
// 数据写入通过 db.CreateMaterial() 手写 SQL 完成。
func (s *MaterialService) Create(ctx context.Context, req *dto.CreateMaterialReq) (*dto.MaterialResp, error) {
    // ...
}

// ✅ 存储过程调用注释
// ConfirmStockIn 确认入库。
// 委托 sp_confirm_stock_in 存储过程执行，内部包含：
//   - 更新入库单状态
//   - 增加 inventory 库存
//   - 写入 inventory_transaction 流水
//   - 更新采购订单已到货量和状态
//   - 校验受压元件材料的材质证明关联
func (s *InventoryService) ConfirmStockIn(ctx context.Context, stockInID, operatorID int64) error {
    return db.ConfirmStockIn(ctx, s.pool, stockInID, operatorID)
}
```

### 1.9 日志规范

```go
// ✅ 使用结构化日志
logger.Info("物料创建成功",
    zap.String("material_code", material.MaterialCode),
    zap.Int64("user_id", userID),
)

logger.Error("入库失败",
    zap.Int64("stock_in_id", stockInID),
    zap.Error(err),
)

// ✅ 日志级别使用规范
// Debug：开发调试信息（SQL、请求参数详情）
// Info：业务关键操作（创建、修改、删除、审批）
// Warn：非致命但需关注的问题（库存预警、性能降低）
// Error：业务错误或异常（入库失败、权限校验失败）
// Fatal：系统无法继续运行的错误（数据库连接失败）

// ❌ 不要在日志中记录敏感信息
logger.Info("用户登录", zap.String("password", password))  // 禁止记录密码
```

---

## 二、Svelte 前端代码规范

### 2.1 目录结构与命名

- 路由页面使用 SvelteKit 约定：`src/routes/**/+page.svelte`
- 服务器侧加载/提交逻辑放在：`src/routes/**/+page.server.ts`
- 复用组件放在：`src/lib/components/`，文件名使用 PascalCase
- 状态管理放在：`src/lib/store/`，文件名小写或小写下划线
- API 封装放在：`src/lib/api/`，禁止在页面中直接写裸 `fetch`

### 2.2 组件与状态规范

- 统一使用 `lang="ts"`，禁止无类型脚本
- 使用 Svelte 5 runes：`$state`、`$derived`、`$effect`、`$props`
- 跨组件优先 props/events；跨页面共享状态再使用 store
- 页面内复杂业务逻辑抽离到 `src/lib/*`，避免 `+page.svelte` 过重

### 2.3 UI 与交互规范

- 样式优先使用 Tailwind + DaisyUI，避免大段自定义样式
- 所有异步操作提供 loading、成功提示、失败提示
- 表格筛选重置后，分页参数与分页组件必须双向同步
- 禁止直接渲染不可信 HTML；必须渲染时先做白名单过滤

### 2.4 接口与错误处理

- 响应格式统一：
```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```
- 参数错误返回 `400`，业务错误返回 `422`，权限错误返回 `401/403`
- 页面只处理展示语义，错误码到中文文案映射在前端统一层维护

---

## 三、SQL 与数据库脚本规范

### 3.1 脚本管理

- 所有结构与数据变更必须落库到 `sql/migrations/`
- 禁止直接改 `sql/schema.sql`（该文件仅用于结构导出备份）
- 迁移命名建议：`Vxxx__<feature>.sql`
- 迁移必须幂等或可重复执行（使用 `IF EXISTS/IF NOT EXISTS`）

### 3.2 DB-First 规则

- 核心流程（入库、出库、盘点、审批）在存储过程中实现事务
- Go 层调用形式：
  - `CALL sp_xxx($1, $2)`
  - `SELECT fn_xxx($1)`
  - `SELECT * FROM v_xxx`
- 单据编号统一调用 `fn_generate_serial_no`

### 3.3 可空字段与 Scan 规则

- 可空文本字段：
  - SQL 侧用 `COALESCE(col, '')`，或
  - Go 侧用 `*string`/`sql.NullString`/`pgtype.Text`
- 新增可空列或新增 LEFT JOIN 后，必须逐一检查 `Scan` 类型匹配

### 3.4 注释规范

- 每个表、列、函数、存储过程、视图必须有中文 `COMMENT`
- 迁移脚本中创建对象时同步补齐注释，避免后补遗漏

---

## 四、测试与质量门禁

### 4.1 后端测试

- 单元测试文件放 `test/`，命名 `*_test.go`
- Service 层优先测试业务分支（成功、参数错误、状态错误、并发冲突）
- DB 层至少覆盖关键查询和关键存储过程调用

### 4.2 SQL 验证

- 每个核心存储过程至少提供一条“成功路径 + 失败路径”验证 SQL
- 验证重点：事务原子性、状态流转、库存不为负、追溯链完整

### 4.3 基础质量检查

- Go：`go test ./...`、`go vet ./...`
- 前端：`npm run typecheck`、`npm run lint`
- SQL：关键迁移在本地数据库执行通过后再提交

---

## 五、Git 与提交规范

### 5.1 分支规范

- 功能开发：`feature/<name>`
- 缺陷修复：`fix/<name>`
- 紧急修复：`hotfix/<name>`

### 5.2 提交信息规范

```
<type>(<scope>): <subject>
```

常用 `type`：`feat`、`fix`、`docs`、`refactor`、`test`、`chore`

示例：
- `feat(inventory): 增加库存预警筛选条件`
- `fix(stock-out): 修复批次扣减并发冲突`
- `docs(design): 对齐 DB-First 架构说明`

### 5.3 Pull Request 规范

- PR 描述必须包含：变更目的、影响范围、验证步骤、回滚方式
- 涉及数据库改动时必须明确对应迁移文件

---

## 六、安全规范

1. 密码必须使用 `bcrypt` 哈希，禁止明文或可逆加密
2. 所有 SQL 必须参数化，禁止字符串拼接用户输入
3. JWT 必须设置过期时间并支持失效/刷新
4. 日志中禁止输出密码、Token、密钥、完整身份证号/手机号
5. 上传文件必须校验类型与大小，文件名重写后存储
6. 关键业务操作写审计日志（操作者、对象、动作、时间、结果）
