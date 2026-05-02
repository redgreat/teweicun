# 特维存（TeWeiCun）— 系统设计文档

> **文档版本**：v2.0（DB-First 架构）  
> **创建日期**：2026-04-11  
> **最后更新**：2026-04-11  
> **文档状态**：草稿  
> **架构策略**：数据库优先，业务逻辑下沉 PostgreSQL

---

## 目录

- [1. 设计概述](#1-设计概述)
- [2. 系统架构](#2-系统架构)
- [3. 技术选型](#3-技术选型)
- [4. 数据库设计](#4-数据库设计)
- [5. API 设计](#5-api-设计)
- [6. 安全设计](#6-安全设计)
- [7. 前端设计](#7-前端设计)
- [8. 部署架构](#8-部署架构)
- [9. 开发规范](#9-开发规范)
- [10. 开发计划](#10-开发计划)

---

## 1. 设计概述

### 1.1 设计原则

| 原则 | 说明 |
|------|------|
| **简洁实用** | 面向小型企业，避免过度设计，以实用为导向 |
| **数据库优先** | 核心业务逻辑下沉 PostgreSQL，充分利用存储过程/函数/触发器/视图 |
| **数据为先** | 确保数据一致性和追溯性，这是特种设备行业的核心要求 |
| **渐进增强** | 一期聚焦核心进销存，预留二期合同/财务的扩展接口 |
| **前后端分离** | 前后端独立开发部署，通过 RESTful API 通信 |

### 1.2 系统边界

```
                    ┌──────────────────────────────────┐
                    │        特维存 系统边界             │
                    │                                  │
  用户（浏览器）───→│  前端 SPA ←─REST API─→ 后端服务   │
                    │                          │       │
                    │                     ┌────┴────┐  │
                    │                     │PostgreSQL│  │
                    │                     └────┬────┘  │
                    │                          │       │
                    │                     ┌────┴────┐  │
                    │                     │  文件存储 │  │
                    │                     └─────────┘  │
                    └──────────────────────────────────┘
```

---

## 2. 系统架构

采用 **前后端分离单体架构 + DB-First 策略**，核心业务逻辑下沉数据库。

```
┌─────────────────────────────────────────────────────┐
│                      前端层                          │
│              SvelteKit + DaisyUI                    │
│     ┌──────┬──────┬──────┬──────┬──────┐            │
│     │ 基础 │ 采购 │ 库存 │ 销售 │ 系统 │            │
│     │ 数据 │ 管理 │ 管理 │ 管理 │ 管理 │            │
│     └──┬───┴──┬───┴──┬───┴──┬───┴──┬───┘            │
│        └──────┴──────┴──────┴──────┘                 │
│                      │ HTTP/HTTPS                    │
├──────────────────────┼──────────────────────────────┤
│                      ▼                               │
│                   后端服务层                           │
│              Go (Gin Framework)                      │
│  ┌─────────────────────────────────────────┐         │
│  │              中间件层                     │         │
│  │  JWT认证 │ RBAC鉴权 │ 日志 │ 限流 │ CORS  │         │
│  ├─────────────────────────────────────────┤         │
│  │         业务编排层 (Service)              │         │
│  │  参数校验 │ 业务判断 │ 调用数据库层        │         │
│  ├─────────────────────────────────────────┤         │
│  │         数据访问层 (DB)                   │         │
│  │  pgxpool + 手写 SQL（禁止 ORM）            │         │
│  │  CALL sp_xxx() │ SELECT fn_xxx()        │         │
│  │  SELECT * FROM v_xxx                    │         │
│  └─────────────────────────────────────────┘         │
│                      │                               │
├──────────────────────┼──────────────────────────────┤
│                      ▼                               │
│              PostgreSQL 数据库层 ⭐                     │
│  ┌─────────────────────────────────────────┐         │
│  │  存储过程：业务流程 + 事务控制                │         │
│  │    sp_confirm_stock_in / sp_confirm_stock_out   │         │
│  │    sp_confirm_sales_order / sp_approve_xxx      │         │
│  │  函数：编号生成 + 库存计算 + 追溯查询            │         │
│  │    fn_generate_serial_no / fn_trace_forward     │         │
│  │  触发器：审计日志 + 时间戳 + 金额汇总              │         │
│  │  视图：报表汇总 + 仪表盘 + 库存预警               │         │
│  └─────────────────────────────────────────┘         │
│  ┌──────────────┬──────────────┐                      │
│  │    Redis     │  本地文件/MinIO│                      │
│  │   (缓存/会话) │  (附件存储)   │                      │
│  └──────────────┴──────────────┘                      │
└─────────────────────────────────────────────────────┘
```

### 2.2 后端分层架构

```
项目结构：
TeWeiCun/
├── cmd/
│   └── server/
│       └── main.go              # 应用入口
├── internal/
│   ├── config/                  # 配置管理
│   │   └── config.go
│   ├── middleware/               # 中间件
│   │   ├── auth.go              # JWT 认证
│   │   ├── rbac.go              # 权限控制
│   │   ├── logger.go            # 请求日志
│   │   ├── recovery.go          # 异常恢复
│   │   └── cors.go              # 跨域处理
│   ├── handler/                 # 请求处理层
│   │   ├── auth.go              # 登录/登出
│   │   ├── material.go          # 物料管理
│   │   ├── supplier.go          # 供应商管理
│   │   ├── customer.go          # 客户管理
│   │   ├── warehouse.go         # 仓库管理
│   │   ├── purchase.go          # 采购管理
│   │   ├── inventory.go         # 库存管理
│   │   ├── sales.go             # 销售管理
│   │   ├── certificate.go       # 材质证明管理
│   │   ├── report.go            # 报表统计
│   │   ├── user.go              # 用户管理
│   │   ├── system.go            # 系统配置
│   │   └── response.go          # 响应封装
│   ├── service/                 # 业务编排层
│   │   ├── material_svc.go
│   │   ├── supplier_svc.go
│   │   ├── customer_svc.go
│   │   ├── warehouse_svc.go
│   │   ├── purchase_svc.go
│   │   ├── inventory_svc.go
│   │   ├── sales_svc.go
│   │   ├── certificate_svc.go
│   │   ├── report_svc.go
│   │   ├── user_svc.go
│   │   └── auth_svc.go
│   ├── db/                      # 数据库访问层（手写 SQL）
│   │   ├── user_db.go           # 用户 SQL
│   │   ├── role_db.go           # 角色/权限 SQL
│   │   ├── audit_db.go          # 审计日志 SQL
│   │   ├── dict_db.go           # 数据字典 SQL
│   │   ├── material_db.go       # 物料 SQL
│   │   ├── supplier_db.go       # 供应商 SQL
│   │   ├── customer_db.go       # 客户 SQL
│   │   ├── warehouse_db.go      # 仓库 SQL
│   │   ├── purchase_db.go       # 采购 SQL
│   │   ├── inventory_db.go      # 库存 SQL
│   │   ├── sales_db.go          # 销售 SQL
│   │   ├── certificate_db.go    # 材质证明 SQL
│   │   └── notification_db.go   # 通知 SQL
│   ├── dto/                     # 数据传输对象
│   │   ├── request/             # 请求 DTO
│   │   └── response/            # 响应 DTO
│   └── pkg/                     # 内部工具包
│       ├── errcode/             # 错误码定义
│       └── utils/               # 通用工具
├── pkg/                         # 可导出的公共包
│   ├── logger/                  # 日志库
│   ├── database/                # 数据库连接池（pgxpool）
│   ├── cache/                   # Redis 缓存封装
│   └── storage/                 # 文件存储
├── sql/                         # ⭐ 数据库 SQL 脚本
│   ├── migrations/              # DDL 迁移脚本
│   ├── functions/               # 函数（fn_xxx）
│   ├── procedures/              # 存储过程（sp_xxx）
│   ├── triggers/                # 触发器（trg_xxx）
│   ├── views/                   # 视图（v_xxx）
│   ├── seed/                    # 初始数据
│   ├── changes/                 # 迭代变更脚本
│   └── tools/                   # 工具脚本
├── api/                         # API 文档
│   └── swagger.yaml
├── configs/                     # 配置文件
│   ├── config.yaml              # 默认配置
│   ├── config.dev.yaml          # 开发环境配置
│   └── config.prod.yaml         # 生产环境配置
├── scripts/                     # 部署/工具脚本
├── docs/                        # 项目文档
│   ├── requirements.md          # 需求文档
│   ├── design.md                # 设计文档（本文件）
│   └── database_design.md       # 数据库设计文档（⭐ 核心参考）
├── web/                         # 前端项目
│   └── (SvelteKit 项目)
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── docker-compose.yml
└── README.md
```

### 2.3 关键设计决策

| 决策项 | 选择 | 理由 |
|-------|------|------|
| 架构模式 | 前后端分离单体 + DB-First | 小团队高效开发，充分利用 PG 能力 |
| 认证方式 | JWT Token | 无状态、前后端分离友好 |
| 权限模型 | RBAC（基于角色） | 满足多角色权限控制需求，实现简洁 |
| 数据库驱动 | pgx (pgxpool) | Go 原生 PG 驱动，性能最佳，支持存储过程 |
| 数据访问 | 手写 SQL（禁止 ORM） | DBA 友好，充分控制 SQL，充分利用 PG 特性 |
| 业务逻辑 | PostgreSQL 存储过程/函数 | 事务安全、可维护、DBA 可直接维护 |
| 报表查询 | PostgreSQL 视图 | 复杂查询封装在视图中，应用层简单调用 |
| 数据库迁移 | SQL 脚本（版本化） | `sql/migrations/` 目录按版本号顺序执行 |
| 配置管理 | Viper | 支持多格式、多环境配置 |
| 日志 | Zap | 高性能结构化日志 |
| API 文档 | Swagger/OpenAPI | 前后端协作标准化 |

---

## 3. 技术选型

### 3.1 后端技术栈

| 类别 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 编程语言 | Go | 1.22+ | 高性能、编译型、部署简单 |
| Web 框架 | Gin | v1.10+ | 高性能 HTTP 框架 |
| 数据库驱动 | pgx (pgxpool) | v5 | Go 原生 PG 驱动，性能最佳 |
| 数据库 | PostgreSQL | 16+ | 存储过程/函数/触发器/视图/JSONB |
| 缓存 | Redis | 7+ | 会话缓存、权限缓存 |
| 认证 | JWT (golang-jwt) | v5 | Token 认证 |
| 配置 | Viper | v1.18+ | 配置文件管理 |
| 日志 | Zap | v1.27+ | 结构化日志 |
| 验证 | validator | v10 | 请求参数校验 |
| API 文档 | swag | v1.16+ | Swagger 文档自动生成 |
| 文件存储 | MinIO SDK | — | 对象存储（材质证明PDF 等）|
| 密码 | bcrypt (golang.org/x/crypto) | — | 密码加密 |

### 3.2 前端技术栈

| 类别 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 框架 | SvelteKit（Svelte 5） | 最新稳定版 | 全栈路由、SSR/Server Load 支持 |
| 构建工具 | Vite | 5+ | 快速构建 |
| UI 组件库 | DaisyUI + Tailwind CSS | 最新稳定版 | 轻量、可主题化、与 Svelte 生态贴合 |
| 状态管理 | Svelte runes + Store | 最新稳定版 | 页面内 runes，跨页面共享用 store |
| 路由 | SvelteKit 文件路由 | 内置 | 约定式路由与服务端数据加载 |
| HTTP 客户端 | 统一 API Client（fetch 封装） | 项目内实现 | 统一鉴权、错误处理、重试策略 |
| 图表 | ECharts | 5+ | 报表图表 |
| 富文本/报表 | — | — | 后续按需引入 |
| CSS 预处理 | SCSS | — | 样式开发 |
| 代码规范 | ESLint + Prettier | — | 代码格式化 |

### 3.3 基础设施

| 类别 | 技术 | 说明 |
|------|------|------|
| 数据库设计文档 | `docs/database_design.md` | ER图、数据流图、存储过程/函数/视图设计 |
| 反向代理 | Nginx | 静态资源服务 + API 反向代理 |
| 容器化 | Docker + Docker Compose | 一键部署 |
| CI/CD | GitHub Actions | 自动构建/测试/部署 |
| 版本控制 | Git + GitHub | 代码管理 |

### 3.4 技术选型理由

**为什么选 Go 而不是 Java/Node/Python？**

| 维度 | Go | Java (Spring) | Node (NestJS) | Python (Django) |
|------|-----|---------------|---------------|-----------------|
| 部署复杂度 | ⭐ 单二进制 | ⚠️ JVM + JAR | ⚠️ Node 运行时 | ⚠️ Python 环境 |
| 性能 | ⭐ 高 | ⭐ 高 | 🔹 中 | 🔸 较低 |
| 并发处理 | ⭐ Goroutine | 🔹 线程池 | 🔹 异步 | 🔸 GIL 限制 |
| 学习曲线 | 🔹 中等 | ⚠️ 较陡峭 | 🔹 中等 | ⭐ 平坦 |
| 内存占用 | ⭐ 低 | ⚠️ 高 | 🔹 中 | 🔹 中 |
| 小团队友效 | ⭐ 高 | 🔸 中（重框架）| 🔹 高 | 🔹 高 |

> Go 的核心优势：**单二进制部署**极简、性能优秀、并发模型强大，非常适合小团队开发中小规模的管理系统。

**为什么选 SvelteKit + DaisyUI/Tailwind？**

- SvelteKit 提供约定式路由和服务端加载能力，适合中后台页面的数据预取
- Tailwind + DaisyUI 组件组合灵活，能快速完成业务页面并保持样式一致性
- 通过统一 API Client 可沉淀鉴权和错误处理，降低页面重复代码
- 与 Go 后端接口边界清晰，便于前后端并行开发

---

## 4. 数据库设计

### 4.1 设计规范

- 所有表使用 `id` 作为主键（BIGSERIAL）
- 通用审计字段：`created_by`, `created_at`, `updated_by`, `updated_at`, `deleted_at`（软删除）
- 外键使用 `xxx_id` 命名，**不使用数据库级外键约束**，由应用层控制
- 金额类字段使用 `DECIMAL(18,2)`
- 数量类字段使用 `DECIMAL(18,3)`（支持重量等小数量）
- 枚举值使用 `VARCHAR` 存储（便于扩展），不使用数据库枚举类型
- 表名使用小写下划线命名，如 `purchase_order`
- 索引命名规则：`idx_表名_字段名`

### 4.2 ER 图（核心实体关系）

```
                    ┌──────────────┐
                    │   supplier   │ ← 合格供方评定
                    │   (供应商)    │
                    └──────┬───────┘
                           │ 1:N
                    ┌──────┴───────┐        ┌──────────────┐
                    │purchase_order│        │   customer   │
                    │  (采购订单)   │        │   (客户)     │
                    └──────┬───────┘        └──────┬───────┘
                           │ 1:N                   │ 1:N
                    ┌──────┴───────┐        ┌──────┴───────┐
                    │purchase_order│        │ sales_order  │
                    │   _item      │        │  (销售订单)   │
                    │ (采购订单明细)│        └──────┬───────┘
                    └──────┬───────┘               │ 1:N
                           │                ┌──────┴───────┐
                    ┌──────┴───────┐        │ sales_order  │
                    │  stock_in    │        │   _item      │
                    │  (入库单)    │        │ (销售订单明细)│
                    └──────┬───────┘        └──────────────┘
                           │ 1:N                   │
                    ┌──────┴───────┐               │
                    │  stock_in    │               │
                    │   _item      │               │
                    │ (入库明细)    │               ▼
                    └──────┬───────┘        ┌──────────────┐
                           │               │  stock_out   │
        ┌──────────────┐   │               │  (出库单)    │
        │material_cert │   │               └──────┬───────┘
        │(材质证明书)   │◄──┤                      │ 1:N
        └──────────────┘   │               ┌──────┴───────┐
                           ▼               │  stock_out   │
                    ┌──────────────┐        │   _item      │
                    │  inventory   │        │ (出库明细)    │
                    │  (库存台账)   │◄───────┘              │
                    └──────┬───────┘        └──────────────┘
                           │
                    ┌──────┴───────┐
                    │   material   │
                    │   (物料)     │
                    └──────────────┘
```

### 4.3 数据表定义

#### 4.3.1 用户与权限

```sql
-- =====================================================
-- 用户表
-- =====================================================
CREATE TABLE sys_user (
    id              BIGSERIAL PRIMARY KEY,
    username        VARCHAR(50)     NOT NULL UNIQUE,           -- 登录用户名
    password_hash   VARCHAR(255)    NOT NULL,                  -- 密码哈希(bcrypt)
    real_name       VARCHAR(50)     NOT NULL,                  -- 真实姓名
    phone           VARCHAR(20),                               -- 手机号
    email           VARCHAR(100),                              -- 邮箱
    department      VARCHAR(100),                              -- 部门
    status          VARCHAR(20)     NOT NULL DEFAULT 'enabled', -- enabled/disabled
    last_login_at   TIMESTAMPTZ,                               -- 最后登录时间
    last_login_ip   VARCHAR(50),                               -- 最后登录IP
    created_by      BIGINT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by      BIGINT,
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ                                -- 软删除
);

CREATE INDEX idx_sys_user_username ON sys_user(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_sys_user_status ON sys_user(status) WHERE deleted_at IS NULL;

-- =====================================================
-- 角色表
-- =====================================================
CREATE TABLE sys_role (
    id              BIGSERIAL PRIMARY KEY,
    role_code       VARCHAR(50)     NOT NULL UNIQUE,           -- 角色编码
    role_name       VARCHAR(100)    NOT NULL,                  -- 角色名称
    description     TEXT,                                      -- 角色描述
    status          VARCHAR(20)     NOT NULL DEFAULT 'enabled',
    created_by      BIGINT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by      BIGINT,
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- =====================================================
-- 用户角色关联表
-- =====================================================
CREATE TABLE sys_user_role (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT          NOT NULL,
    role_id         BIGINT          NOT NULL,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_sys_user_role_user ON sys_user_role(user_id);
CREATE INDEX idx_sys_user_role_role ON sys_user_role(role_id);

-- =====================================================
-- 权限表
-- =====================================================
CREATE TABLE sys_permission (
    id              BIGSERIAL PRIMARY KEY,
    parent_id       BIGINT          DEFAULT 0,                -- 父权限 ID（菜单层级）
    perm_code       VARCHAR(100)    NOT NULL UNIQUE,           -- 权限编码，如 material:create
    perm_name       VARCHAR(100)    NOT NULL,                  -- 权限名称
    perm_type       VARCHAR(20)     NOT NULL,                  -- menu/button/api
    path            VARCHAR(200),                              -- 前端路由路径
    icon            VARCHAR(100),                              -- 菜单图标
    sort_order      INT             DEFAULT 0,                 -- 排序
    status          VARCHAR(20)     NOT NULL DEFAULT 'enabled',
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- =====================================================
-- 角色权限关联表
-- =====================================================
CREATE TABLE sys_role_permission (
    id              BIGSERIAL PRIMARY KEY,
    role_id         BIGINT          NOT NULL,
    permission_id   BIGINT          NOT NULL,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_sys_role_perm_role ON sys_role_permission(role_id);

-- =====================================================
-- 操作日志表
-- =====================================================
CREATE TABLE sys_audit_log (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT          NOT NULL,
    username        VARCHAR(50)     NOT NULL,
    action          VARCHAR(50)     NOT NULL,                  -- CREATE/UPDATE/DELETE/LOGIN/LOGOUT
    module          VARCHAR(50)     NOT NULL,                  -- 模块名
    target_type     VARCHAR(50),                               -- 操作目标类型
    target_id       BIGINT,                                    -- 操作目标 ID
    detail          JSONB,                                     -- 操作详情（变更前后对比）
    ip_address      VARCHAR(50),
    user_agent      VARCHAR(500),
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_user ON sys_audit_log(user_id);
CREATE INDEX idx_audit_log_module ON sys_audit_log(module);
CREATE INDEX idx_audit_log_created ON sys_audit_log(created_at);
```

#### 4.3.2 基础数据

```sql
-- =====================================================
-- 物料分类表
-- =====================================================
CREATE TABLE material_category (
    id              BIGSERIAL PRIMARY KEY,
    parent_id       BIGINT          DEFAULT 0,                -- 父分类 ID，0 为顶级
    category_code   VARCHAR(10)     NOT NULL UNIQUE,          -- 分类编码
    category_name   VARCHAR(100)    NOT NULL,                 -- 分类名称
    sort_order      INT             DEFAULT 0,
    status          VARCHAR(20)     NOT NULL DEFAULT 'enabled',
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- =====================================================
-- 物料表
-- =====================================================
CREATE TABLE material (
    id                  BIGSERIAL PRIMARY KEY,
    material_code       VARCHAR(30)     NOT NULL UNIQUE,       -- 物料编码
    material_name       VARCHAR(100)    NOT NULL,               -- 物料名称
    specification       VARCHAR(200)    NOT NULL,               -- 规格型号
    category_id         BIGINT          NOT NULL,               -- 物料分类
    material_grade      VARCHAR(50),                            -- 材质/材料牌号
    standard_no         VARCHAR(50),                            -- 材料标准号
    unit                VARCHAR(20)     NOT NULL,               -- 主计量单位
    aux_unit            VARCHAR(20),                            -- 辅助计量单位
    conversion_factor   DECIMAL(18,6),                          -- 换算系数
    is_pressure_part    BOOLEAN         NOT NULL DEFAULT FALSE, -- 是否受压元件材料
    need_recheck        BOOLEAN         NOT NULL DEFAULT FALSE, -- 是否需复验
    safety_stock        DECIMAL(18,3)   DEFAULT 0,              -- 安全库存
    max_stock           DECIMAL(18,3)   DEFAULT 0,              -- 最高库存
    default_warehouse_id BIGINT,                                -- 默认仓库
    remark              TEXT,
    status              VARCHAR(20)     NOT NULL DEFAULT 'enabled',
    created_by          BIGINT,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by          BIGINT,
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_material_code ON material(material_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_material_category ON material(category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_material_grade ON material(material_grade) WHERE deleted_at IS NULL;
CREATE INDEX idx_material_name ON material(material_name) WHERE deleted_at IS NULL;

-- =====================================================
-- 供应商表
-- =====================================================
CREATE TABLE supplier (
    id                      BIGSERIAL PRIMARY KEY,
    supplier_code           VARCHAR(20)     NOT NULL UNIQUE,
    supplier_name           VARCHAR(200)    NOT NULL,
    credit_code             VARCHAR(18),                        -- 统一社会信用代码
    supplier_type           VARCHAR(50)     NOT NULL,           -- raw_material/welding/standard_part/purchased/other
    contact_person          VARCHAR(50)     NOT NULL,
    contact_phone           VARCHAR(20)     NOT NULL,
    address                 VARCHAR(500),
    is_qualified            BOOLEAN         NOT NULL DEFAULT FALSE, -- 是否合格供方
    qualification_expire    DATE,                               -- 合格供方有效期
    supplier_rating         VARCHAR(5),                         -- A/B/C/D
    bank_name               VARCHAR(100),
    bank_account            VARCHAR(30),
    remark                  TEXT,
    status                  VARCHAR(20)     NOT NULL DEFAULT 'enabled', -- enabled/disabled/blacklisted
    created_by              BIGINT,
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by              BIGINT,
    updated_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ
);

CREATE INDEX idx_supplier_code ON supplier(supplier_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_supplier_name ON supplier(supplier_name) WHERE deleted_at IS NULL;
CREATE INDEX idx_supplier_qualified ON supplier(is_qualified, qualification_expire) WHERE deleted_at IS NULL;

-- =====================================================
-- 供应商资质证书表
-- =====================================================
CREATE TABLE supplier_certificate (
    id              BIGSERIAL PRIMARY KEY,
    supplier_id     BIGINT          NOT NULL,
    cert_type       VARCHAR(50)     NOT NULL,                  -- 资质类型
    cert_name       VARCHAR(200)    NOT NULL,                  -- 资质名称
    cert_no         VARCHAR(100),                              -- 证书编号
    issue_date      DATE,                                      -- 发证日期
    expire_date     DATE,                                      -- 到期日期
    file_path       VARCHAR(500),                              -- 证书附件路径
    remark          TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_supplier_cert_supplier ON supplier_certificate(supplier_id);
CREATE INDEX idx_supplier_cert_expire ON supplier_certificate(expire_date);

-- =====================================================
-- 客户表
-- =====================================================
CREATE TABLE customer (
    id              BIGSERIAL PRIMARY KEY,
    customer_code   VARCHAR(20)     NOT NULL UNIQUE,
    customer_name   VARCHAR(200)    NOT NULL,
    credit_code     VARCHAR(18),
    customer_type   VARCHAR(50),                               -- direct_user/trader/engineering/other
    contact_person  VARCHAR(50)     NOT NULL,
    contact_phone   VARCHAR(20)     NOT NULL,
    address         VARCHAR(500),
    sales_person_id BIGINT,                                    -- 负责销售员
    remark          TEXT,
    status          VARCHAR(20)     NOT NULL DEFAULT 'enabled',
    created_by      BIGINT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by      BIGINT,
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_customer_code ON customer(customer_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_customer_name ON customer(customer_name) WHERE deleted_at IS NULL;

-- =====================================================
-- 仓库表
-- =====================================================
CREATE TABLE warehouse (
    id              BIGSERIAL PRIMARY KEY,
    warehouse_code  VARCHAR(10)     NOT NULL UNIQUE,
    warehouse_name  VARCHAR(50)     NOT NULL,
    warehouse_type  VARCHAR(30)     NOT NULL,                  -- raw_material/welding/semi_finished/finished/scrap
    manager_id      BIGINT          NOT NULL,                  -- 仓库负责人
    location        VARCHAR(200),
    enable_location BOOLEAN         NOT NULL DEFAULT FALSE,    -- 是否启用库位管理
    status          VARCHAR(20)     NOT NULL DEFAULT 'enabled',
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- =====================================================
-- 库位表
-- =====================================================
CREATE TABLE warehouse_location (
    id              BIGSERIAL PRIMARY KEY,
    warehouse_id    BIGINT          NOT NULL,
    location_code   VARCHAR(20)     NOT NULL,                  -- 库位编码
    capacity_desc   VARCHAR(100),                              -- 容量描述
    status          VARCHAR(20)     NOT NULL DEFAULT 'idle',   -- idle/occupied/locked
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_wh_location_code ON warehouse_location(warehouse_id, location_code);
```

#### 4.3.3 采购管理

```sql
-- =====================================================
-- 采购申请单
-- =====================================================
CREATE TABLE purchase_request (
    id              BIGSERIAL PRIMARY KEY,
    request_no      VARCHAR(20)     NOT NULL UNIQUE,           -- 申请单号
    request_date    DATE            NOT NULL,
    requester_id    BIGINT          NOT NULL,                  -- 申请人
    department      VARCHAR(100),                              -- 申请部门
    required_date   DATE,                                      -- 需求日期
    request_reason  VARCHAR(50)     NOT NULL,                  -- production/safety_stock/project/other
    project_no      VARCHAR(50),                               -- 关联项目号
    approval_status VARCHAR(20)     NOT NULL DEFAULT 'draft',  -- draft/pending/approved/rejected/closed
    approved_by     BIGINT,
    approved_at     TIMESTAMPTZ,
    approval_remark TEXT,
    remark          TEXT,
    created_by      BIGINT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by      BIGINT,
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_pr_no ON purchase_request(request_no) WHERE deleted_at IS NULL;
CREATE INDEX idx_pr_status ON purchase_request(approval_status) WHERE deleted_at IS NULL;
CREATE INDEX idx_pr_requester ON purchase_request(requester_id) WHERE deleted_at IS NULL;

-- =====================================================
-- 采购申请明细
-- =====================================================
CREATE TABLE purchase_request_item (
    id              BIGSERIAL PRIMARY KEY,
    request_id      BIGINT          NOT NULL,                  -- 关联申请单
    material_id     BIGINT          NOT NULL,                  -- 物料
    quantity        DECIMAL(18,3)   NOT NULL,                  -- 申请数量
    unit            VARCHAR(20)     NOT NULL,                  -- 单位
    current_stock   DECIMAL(18,3),                             -- 当前库存（快照）
    remark          TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pri_request ON purchase_request_item(request_id);

-- =====================================================
-- 采购订单
-- =====================================================
CREATE TABLE purchase_order (
    id              BIGSERIAL PRIMARY KEY,
    order_no        VARCHAR(20)     NOT NULL UNIQUE,           -- 订单编号
    request_id      BIGINT,                                    -- 关联采购申请（可选）
    supplier_id     BIGINT          NOT NULL,                  -- 供应商
    buyer_id        BIGINT          NOT NULL,                  -- 采购员
    order_date      DATE            NOT NULL,
    expected_date   DATE,                                      -- 预计到货日期
    payment_method  VARCHAR(50),                               -- prepay/on_delivery/monthly/other
    order_status    VARCHAR(20)     NOT NULL DEFAULT 'draft',  -- draft/ordered/partial_received/full_received/closed
    total_amount    DECIMAL(18,2)   DEFAULT 0,                 -- 总金额
    remark          TEXT,
    created_by      BIGINT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by      BIGINT,
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_po_no ON purchase_order(order_no) WHERE deleted_at IS NULL;
CREATE INDEX idx_po_supplier ON purchase_order(supplier_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_po_status ON purchase_order(order_status) WHERE deleted_at IS NULL;
CREATE INDEX idx_po_date ON purchase_order(order_date) WHERE deleted_at IS NULL;

-- =====================================================
-- 采购订单明细
-- =====================================================
CREATE TABLE purchase_order_item (
    id                  BIGSERIAL PRIMARY KEY,
    order_id            BIGINT          NOT NULL,
    material_id         BIGINT          NOT NULL,
    quantity            DECIMAL(18,3)   NOT NULL,              -- 采购数量
    unit_price          DECIMAL(18,2),                         -- 单价
    amount              DECIMAL(18,2),                         -- 金额
    received_quantity   DECIMAL(18,3)   DEFAULT 0,             -- 已到货数量
    delivery_date       DATE,                                  -- 交货日期
    tech_requirements   TEXT,                                  -- 技术要求
    remark              TEXT,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_poi_order ON purchase_order_item(order_id);
CREATE INDEX idx_poi_material ON purchase_order_item(material_id);
```

#### 4.3.4 库存管理

```sql
-- =====================================================
-- 入库单
-- =====================================================
CREATE TABLE stock_in (
    id              BIGSERIAL PRIMARY KEY,
    stock_in_no     VARCHAR(20)     NOT NULL UNIQUE,           -- 入库单号
    purchase_order_id BIGINT,                                  -- 关联采购订单
    supplier_id     BIGINT,                                    -- 供应商（自动带出）
    warehouse_id    BIGINT          NOT NULL,                  -- 入库仓库
    stock_in_date   DATE            NOT NULL,
    delivery_note_no VARCHAR(50),                              -- 送货单号
    inspector_id    BIGINT,                                    -- 验收人
    stock_in_status       VARCHAR(20)     NOT NULL DEFAULT 'pending',-- pending/passed/failed/concession
    remark          TEXT,
    created_by      BIGINT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by      BIGINT,
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_si_no ON stock_in(stock_in_no) WHERE deleted_at IS NULL;
CREATE INDEX idx_si_po ON stock_in(purchase_order_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_si_date ON stock_in(stock_in_date) WHERE deleted_at IS NULL;

-- =====================================================
-- 入库明细
-- =====================================================
CREATE TABLE stock_in_item (
    id              BIGSERIAL PRIMARY KEY,
    stock_in_id     BIGINT          NOT NULL,                  -- 关联入库单
    material_id     BIGINT          NOT NULL,                  -- 物料
    arrived_quantity DECIMAL(18,3)   NOT NULL,                 -- 到货数量
    accepted_quantity DECIMAL(18,3)  NOT NULL,                 -- 合格数量
    unit            VARCHAR(20)     NOT NULL,
    batch_no        VARCHAR(30)     NOT NULL,                  -- 入库批次号
    location_id     BIGINT,                                    -- 库位
    cert_id         BIGINT,                                    -- 关联材质证明
    remark          TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sii_stock_in ON stock_in_item(stock_in_id);
CREATE INDEX idx_sii_material ON stock_in_item(material_id);
CREATE INDEX idx_sii_batch ON stock_in_item(batch_no);

-- =====================================================
-- 出库单
-- =====================================================
CREATE TABLE stock_out (
    id              BIGSERIAL PRIMARY KEY,
    stock_out_no    VARCHAR(20)     NOT NULL UNIQUE,           -- 出库单号
    out_type        VARCHAR(20)     NOT NULL,                  -- sales/production/transfer/other
    ref_doc_type    VARCHAR(50),                               -- 关联单据类型
    ref_doc_id      BIGINT,                                    -- 关联单据 ID
    warehouse_id    BIGINT          NOT NULL,
    stock_out_date  DATE            NOT NULL,
    receiver        VARCHAR(100),                              -- 领料人/收货方
    remark          TEXT,
    created_by      BIGINT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by      BIGINT,
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_so_no ON stock_out(stock_out_no) WHERE deleted_at IS NULL;
CREATE INDEX idx_so_type ON stock_out(out_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_so_date ON stock_out(stock_out_date) WHERE deleted_at IS NULL;

-- =====================================================
-- 出库明细
-- =====================================================
CREATE TABLE stock_out_item (
    id              BIGSERIAL PRIMARY KEY,
    stock_out_id    BIGINT          NOT NULL,
    material_id     BIGINT          NOT NULL,
    quantity        DECIMAL(18,3)   NOT NULL,                  -- 出库数量
    unit            VARCHAR(20)     NOT NULL,
    batch_no        VARCHAR(30)     NOT NULL,                  -- 出库批次
    location_id     BIGINT,                                    -- 库位
    usage_desc      VARCHAR(200),                              -- 用途说明（产品编号/工程号）
    remark          TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_soi_stock_out ON stock_out_item(stock_out_id);
CREATE INDEX idx_soi_material ON stock_out_item(material_id);
CREATE INDEX idx_soi_batch ON stock_out_item(batch_no);

-- =====================================================
-- 库存台账（实时库存表）
-- =====================================================
CREATE TABLE inventory (
    id              BIGSERIAL PRIMARY KEY,
    material_id     BIGINT          NOT NULL,
    warehouse_id    BIGINT          NOT NULL,
    batch_no        VARCHAR(30)     NOT NULL,                  -- 批次号
    location_id     BIGINT,                                    -- 库位
    quantity        DECIMAL(18,3)   NOT NULL DEFAULT 0,        -- 当前库存量
    locked_quantity DECIMAL(18,3)   NOT NULL DEFAULT 0,        -- 锁定数量
    unit            VARCHAR(20)     NOT NULL,
    stock_in_date   DATE            NOT NULL,                  -- 入库日期（用于超期判断）
    expire_date     DATE,                                      -- 有效期（焊材等）
    unit_cost       DECIMAL(18,2),                             -- 单位成本
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_inventory_qty CHECK (quantity >= 0),
    CONSTRAINT chk_inventory_locked CHECK (locked_quantity >= 0 AND locked_quantity <= quantity)
);

CREATE UNIQUE INDEX idx_inventory_unique ON inventory(material_id, warehouse_id, batch_no, COALESCE(location_id, 0));
CREATE INDEX idx_inventory_material ON inventory(material_id);
CREATE INDEX idx_inventory_warehouse ON inventory(warehouse_id);
CREATE INDEX idx_inventory_batch ON inventory(batch_no);
CREATE INDEX idx_inventory_expire ON inventory(expire_date) WHERE expire_date IS NOT NULL;

-- =====================================================
-- 库存流水（进出记录）
-- =====================================================
CREATE TABLE inventory_transaction (
    id              BIGSERIAL PRIMARY KEY,
    material_id     BIGINT          NOT NULL,
    warehouse_id    BIGINT          NOT NULL,
    batch_no        VARCHAR(30)     NOT NULL,
    trans_type      VARCHAR(20)     NOT NULL,                  -- in/out/adjust/lock/unlock
    quantity        DECIMAL(18,3)   NOT NULL,                  -- 变动数量（正入负出）
    balance         DECIMAL(18,3)   NOT NULL,                  -- 变动后余额
    ref_doc_type    VARCHAR(50),                               -- 关联单据类型
    ref_doc_no      VARCHAR(20),                               -- 关联单据号
    ref_doc_id      BIGINT,
    operator_id     BIGINT          NOT NULL,
    remark          TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inv_trans_material ON inventory_transaction(material_id);
CREATE INDEX idx_inv_trans_batch ON inventory_transaction(batch_no);
CREATE INDEX idx_inv_trans_date ON inventory_transaction(created_at);
CREATE INDEX idx_inv_trans_ref ON inventory_transaction(ref_doc_type, ref_doc_no);

-- =====================================================
-- 库存调拨单
-- =====================================================
CREATE TABLE stock_transfer (
    id              BIGSERIAL PRIMARY KEY,
    transfer_no     VARCHAR(20)     NOT NULL UNIQUE,
    from_warehouse_id BIGINT        NOT NULL,
    to_warehouse_id BIGINT          NOT NULL,
    transfer_date   DATE            NOT NULL,
    transfer_status VARCHAR(20)     NOT NULL DEFAULT 'pending', -- pending/in_transit/completed
    reason          TEXT,
    remark          TEXT,
    created_by      BIGINT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by      BIGINT,
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- =====================================================
-- 调拨明细
-- =====================================================
CREATE TABLE stock_transfer_item (
    id              BIGSERIAL PRIMARY KEY,
    transfer_id     BIGINT          NOT NULL,
    material_id     BIGINT          NOT NULL,
    quantity        DECIMAL(18,3)   NOT NULL,
    batch_no        VARCHAR(30)     NOT NULL,
    unit            VARCHAR(20)     NOT NULL,
    remark          TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- =====================================================
-- 盘点单
-- =====================================================
CREATE TABLE inventory_check (
    id              BIGSERIAL PRIMARY KEY,
    check_no        VARCHAR(20)     NOT NULL UNIQUE,
    check_type      VARCHAR(30)     NOT NULL,                  -- full/category/specific
    warehouse_id    BIGINT          NOT NULL,
    check_date      DATE            NOT NULL,
    checker_id      BIGINT          NOT NULL,
    reviewer_id     BIGINT,
    check_status    VARCHAR(20)     NOT NULL DEFAULT 'draft',  -- draft/checking/pending_review/completed
    reviewed_at     TIMESTAMPTZ,
    remark          TEXT,
    created_by      BIGINT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by      BIGINT,
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- =====================================================
-- 盘点明细
-- =====================================================
CREATE TABLE inventory_check_item (
    id              BIGSERIAL PRIMARY KEY,
    check_id        BIGINT          NOT NULL,
    material_id     BIGINT          NOT NULL,
    batch_no        VARCHAR(30)     NOT NULL,
    book_quantity   DECIMAL(18,3)   NOT NULL,                  -- 账面数量
    actual_quantity DECIMAL(18,3),                              -- 实际数量
    diff_quantity   DECIMAL(18,3),                              -- 差异数量
    diff_reason     VARCHAR(50),                               -- surplus/shortage/loss/error
    remark          TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ici_check ON inventory_check_item(check_id);
```

#### 4.3.5 销售管理

```sql
-- =====================================================
-- 销售订单
-- =====================================================
CREATE TABLE sales_order (
    id              BIGSERIAL PRIMARY KEY,
    order_no        VARCHAR(20)     NOT NULL UNIQUE,
    customer_id     BIGINT          NOT NULL,
    sales_person_id BIGINT          NOT NULL,
    contract_no     VARCHAR(50),                               -- 合同编号（二期关联）
    order_date      DATE            NOT NULL,
    delivery_date   DATE,                                      -- 预计交货日期
    payment_method  VARCHAR(50),
    order_status    VARCHAR(20)     NOT NULL DEFAULT 'draft',  -- draft/confirmed/preparing/shipped/received/closed
    total_amount    DECIMAL(18,2)   DEFAULT 0,
    remark          TEXT,
    created_by      BIGINT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by      BIGINT,
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_sales_no ON sales_order(order_no) WHERE deleted_at IS NULL;
CREATE INDEX idx_sales_customer ON sales_order(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_sales_status ON sales_order(order_status) WHERE deleted_at IS NULL;
CREATE INDEX idx_sales_date ON sales_order(order_date) WHERE deleted_at IS NULL;

-- =====================================================
-- 销售订单明细
-- =====================================================
CREATE TABLE sales_order_item (
    id                  BIGSERIAL PRIMARY KEY,
    order_id            BIGINT          NOT NULL,
    material_id         BIGINT          NOT NULL,
    quantity            DECIMAL(18,3)   NOT NULL,
    unit_price          DECIMAL(18,2),
    amount              DECIMAL(18,2),
    delivery_date       DATE,
    shipped_quantity    DECIMAL(18,3)   DEFAULT 0,             -- 已发货数量
    tech_requirements   TEXT,
    remark              TEXT,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_soi_order ON sales_order_item(order_id);
CREATE INDEX idx_soi2_material ON sales_order_item(material_id);

-- =====================================================
-- 退货单（采购退货 + 销售退货共用）
-- =====================================================
CREATE TABLE return_order (
    id              BIGSERIAL PRIMARY KEY,
    return_no       VARCHAR(20)     NOT NULL UNIQUE,
    return_type     VARCHAR(20)     NOT NULL,                  -- purchase_return/sales_return
    ref_doc_type    VARCHAR(50)     NOT NULL,                  -- stock_in/sales_order
    ref_doc_id      BIGINT          NOT NULL,
    return_reason   VARCHAR(50)     NOT NULL,                  -- quality/spec_mismatch/excess/cancel/other
    return_status   VARCHAR(20)     NOT NULL DEFAULT 'pending',-- pending/returned/completed
    need_restock    BOOLEAN         DEFAULT FALSE,             -- 是否需要退回入库（销售退货时）
    need_recheck    BOOLEAN         DEFAULT FALSE,             -- 是否需要重新质检
    remark          TEXT,
    created_by      BIGINT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_by      BIGINT,
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- =====================================================
-- 退货明细
-- =====================================================
CREATE TABLE return_order_item (
    id              BIGSERIAL PRIMARY KEY,
    return_id       BIGINT          NOT NULL,
    material_id     BIGINT          NOT NULL,
    quantity        DECIMAL(18,3)   NOT NULL,
    batch_no        VARCHAR(30),
    unit            VARCHAR(20)     NOT NULL,
    remark          TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);
```

#### 4.3.6 特种设备专项

```sql
-- =====================================================
-- 材质证明书
-- =====================================================
CREATE TABLE material_certificate (
    id              BIGSERIAL PRIMARY KEY,
    cert_no         VARCHAR(50)     NOT NULL UNIQUE,           -- 证书编号（系统生成）
    material_id     BIGINT          NOT NULL,                  -- 关联物料
    stock_in_item_id BIGINT,                                   -- 关联入库明细
    batch_no        VARCHAR(30)     NOT NULL,                  -- 关联批次号
    material_grade  VARCHAR(50)     NOT NULL,                  -- 材料牌号
    standard_no     VARCHAR(50)     NOT NULL,                  -- 材料标准
    specification   VARCHAR(200),                              -- 规格尺寸
    manufacturer    VARCHAR(200),                              -- 生产厂家
    chemical_composition JSONB,                                -- 化学成分 {"C":0.16,"Si":0.28,...}
    mechanical_properties JSONB,                               -- 力学性能 {"Rm":520,"ReL":350,"A":22,...}
    process_properties JSONB,                                  -- 工艺性能
    nde_result      JSONB,                                     -- 无损检测结果
    file_path       VARCHAR(500),                              -- 扫描件文件路径
    file_name       VARCHAR(200),                              -- 原始文件名
    entered_by      BIGINT          NOT NULL,                  -- 录入人
    entered_at      TIMESTAMPTZ     NOT NULL,                  -- 录入时间
    confirmed_by    BIGINT,                                    -- 确认人（质量员）
    confirmed_at    TIMESTAMPTZ,                               -- 确认时间
    confirm_status  VARCHAR(20)     NOT NULL DEFAULT 'pending',-- pending/confirmed/rejected
    remark          TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_mc_cert_no ON material_certificate(cert_no) WHERE deleted_at IS NULL;
CREATE INDEX idx_mc_material ON material_certificate(material_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_mc_batch ON material_certificate(batch_no) WHERE deleted_at IS NULL;
CREATE INDEX idx_mc_grade ON material_certificate(material_grade) WHERE deleted_at IS NULL;

-- =====================================================
-- 数据字典表
-- =====================================================
CREATE TABLE sys_dict_type (
    id              BIGSERIAL PRIMARY KEY,
    dict_type       VARCHAR(50)     NOT NULL UNIQUE,           -- 字典类型编码
    dict_name       VARCHAR(100)    NOT NULL,                  -- 字典类型名称
    remark          TEXT,
    status          VARCHAR(20)     NOT NULL DEFAULT 'enabled',
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE TABLE sys_dict_data (
    id              BIGSERIAL PRIMARY KEY,
    dict_type       VARCHAR(50)     NOT NULL,                  -- 字典类型
    dict_value      VARCHAR(50)     NOT NULL,                  -- 字典值
    dict_label      VARCHAR(100)    NOT NULL,                  -- 显示标签
    sort_order      INT             DEFAULT 0,
    is_default      BOOLEAN         DEFAULT FALSE,
    status          VARCHAR(20)     NOT NULL DEFAULT 'enabled',
    remark          TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    UNIQUE(dict_type, dict_value)
);

CREATE INDEX idx_dict_data_type ON sys_dict_data(dict_type);

-- =====================================================
-- 系统通知表
-- =====================================================
CREATE TABLE sys_notification (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT          NOT NULL,                  -- 接收用户
    title           VARCHAR(200)    NOT NULL,                  -- 通知标题
    content         TEXT,                                      -- 通知内容
    notify_type     VARCHAR(50)     NOT NULL,                  -- alert/approval/info
    ref_type        VARCHAR(50),                               -- 关联业务类型
    ref_id          BIGINT,                                    -- 关联业务 ID
    is_read         BOOLEAN         NOT NULL DEFAULT FALSE,    -- 是否已读
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notify_user ON sys_notification(user_id, is_read);
CREATE INDEX idx_notify_created ON sys_notification(created_at);

-- =====================================================
-- 单据编号序列表
-- =====================================================
CREATE TABLE sys_serial_number (
    id              BIGSERIAL PRIMARY KEY,
    prefix          VARCHAR(10)     NOT NULL,                  -- 单据前缀
    date_str        VARCHAR(8)      NOT NULL,                  -- 日期字符串 YYYYMMDD
    current_seq     INT             NOT NULL DEFAULT 0,        -- 当前序号
    UNIQUE(prefix, date_str)
);
```

### 4.4 数据库索引策略

| 索引类型 | 应用场景 |
|---------|---------|
| 唯一索引 | 编码字段（物料编码、订单编号等）|
| 普通索引 | 常用查询条件（状态、日期范围、外键关联）|
| 部分索引 | 使用 `WHERE deleted_at IS NULL` 过滤已删除数据 |
| 联合索引 | 库存表的(material_id, warehouse_id, batch_no) |
| JSONB 索引 | 材质证明书的化学成分查询（按需创建 GIN 索引）|

---

## 5. API 设计

### 5.1 API 规范

**基础约定**：
- Base URL: `/api/v1`
- 数据格式: JSON
- 认证方式: Bearer Token (JWT)
- 字符编码: UTF-8
- 时间格式: ISO 8601 (2026-04-11T10:00:00+08:00)

**通用响应格式**：

```json
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

**分页请求**：
```
GET /api/v1/materials?page=1&page_size=20&sort=created_at&order=desc
```

**分页响应**：
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "list": [...],
        "total": 100,
        "page": 1,
        "page_size": 20
    }
}
```

**错误响应**：
```json
{
    "code": 400,
    "message": "参数校验失败",
    "errors": [
        {"field": "material_code", "message": "物料编码不能为空"}
    ]
}
```

### 5.2 错误码定义

| 错误码 | 说明 |
|-------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证（Token 失效）|
| 403 | 无权限 |
| 404 | 资源不存在 |
| 409 | 数据冲突（如编码重复）|
| 422 | 业务逻辑错误（如库存不足）|
| 500 | 服务器内部错误 |

### 5.3 API 清单

#### 认证相关

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/logout` | 用户登出 |
| POST | `/api/v1/auth/refresh` | 刷新 Token |
| GET  | `/api/v1/auth/profile` | 获取当前用户信息 |
| PUT  | `/api/v1/auth/password` | 修改密码 |

#### 基础数据

| 方法 | 路径 | 说明 |
|------|------|------|
| GET    | `/api/v1/materials` | 物料列表（分页/搜索）|
| POST   | `/api/v1/materials` | 创建物料 |
| GET    | `/api/v1/materials/:id` | 物料详情 |
| PUT    | `/api/v1/materials/:id` | 更新物料 |
| DELETE | `/api/v1/materials/:id` | 删除物料（软删除）|
| GET    | `/api/v1/materials/categories` | 物料分类树 |
| POST   | `/api/v1/materials/categories` | 创建分类 |
| PUT    | `/api/v1/materials/categories/:id` | 更新分类 |
| GET    | `/api/v1/suppliers` | 供应商列表 |
| POST   | `/api/v1/suppliers` | 创建供应商 |
| GET    | `/api/v1/suppliers/:id` | 供应商详情 |
| PUT    | `/api/v1/suppliers/:id` | 更新供应商 |
| DELETE | `/api/v1/suppliers/:id` | 删除供应商 |
| GET    | `/api/v1/customers` | 客户列表 |
| POST   | `/api/v1/customers` | 创建客户 |
| GET    | `/api/v1/customers/:id` | 客户详情 |
| PUT    | `/api/v1/customers/:id` | 更新客户 |
| DELETE | `/api/v1/customers/:id` | 删除客户 |
| GET    | `/api/v1/warehouses` | 仓库列表 |
| POST   | `/api/v1/warehouses` | 创建仓库 |
| PUT    | `/api/v1/warehouses/:id` | 更新仓库 |

#### 采购管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET    | `/api/v1/purchase/requests` | 采购申请列表 |
| POST   | `/api/v1/purchase/requests` | 创建采购申请 |
| GET    | `/api/v1/purchase/requests/:id` | 申请详情 |
| PUT    | `/api/v1/purchase/requests/:id` | 更新申请 |
| POST   | `/api/v1/purchase/requests/:id/submit` | 提交审批 |
| POST   | `/api/v1/purchase/requests/:id/approve` | 审批通过 |
| POST   | `/api/v1/purchase/requests/:id/reject` | 审批驳回 |
| GET    | `/api/v1/purchase/orders` | 采购订单列表 |
| POST   | `/api/v1/purchase/orders` | 创建采购订单 |
| GET    | `/api/v1/purchase/orders/:id` | 订单详情 |
| PUT    | `/api/v1/purchase/orders/:id` | 更新订单 |
| POST   | `/api/v1/purchase/orders/:id/confirm` | 确认下单 |
| POST   | `/api/v1/purchase/orders/:id/close` | 关闭订单 |

#### 库存管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET    | `/api/v1/inventory` | 库存台账（支持多维度查询）|
| GET    | `/api/v1/inventory/summary` | 库存汇总 |
| GET    | `/api/v1/inventory/alerts` | 库存预警列表 |
| GET    | `/api/v1/inventory/transactions` | 库存流水 |
| POST   | `/api/v1/stock-in` | 创建入库单 |
| GET    | `/api/v1/stock-in` | 入库单列表 |
| GET    | `/api/v1/stock-in/:id` | 入库单详情 |
| POST   | `/api/v1/stock-in/:id/confirm` | 确认入库 |
| POST   | `/api/v1/stock-out` | 创建出库单 |
| GET    | `/api/v1/stock-out` | 出库单列表 |
| GET    | `/api/v1/stock-out/:id` | 出库单详情 |
| POST   | `/api/v1/stock-out/:id/confirm` | 确认出库 |
| POST   | `/api/v1/stock-transfer` | 创建调拨单 |
| GET    | `/api/v1/stock-transfer` | 调拨单列表 |
| POST   | `/api/v1/inventory-check` | 创建盘点单 |
| GET    | `/api/v1/inventory-check` | 盘点单列表 |
| PUT    | `/api/v1/inventory-check/:id` | 更新盘点结果 |
| POST   | `/api/v1/inventory-check/:id/complete` | 完成盘点 |

#### 销售管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET    | `/api/v1/sales/orders` | 销售订单列表 |
| POST   | `/api/v1/sales/orders` | 创建销售订单 |
| GET    | `/api/v1/sales/orders/:id` | 订单详情 |
| PUT    | `/api/v1/sales/orders/:id` | 更新订单 |
| POST   | `/api/v1/sales/orders/:id/confirm` | 确认订单 |
| POST   | `/api/v1/sales/orders/:id/ship` | 发货 |
| POST   | `/api/v1/sales/orders/:id/close` | 关闭订单 |

#### 特种设备专项

| 方法 | 路径 | 说明 |
|------|------|------|
| GET    | `/api/v1/certificates` | 材质证明书列表 |
| POST   | `/api/v1/certificates` | 录入材质证明书 |
| GET    | `/api/v1/certificates/:id` | 证明书详情 |
| POST   | `/api/v1/certificates/:id/confirm` | 质量员确认 |
| POST   | `/api/v1/certificates/:id/reject` | 质量员驳回 |
| GET    | `/api/v1/trace/backward/:product_no` | 反向追溯（产品→材料）|
| POST   | `/api/v1/upload/certificate` | 上传材质证明扫描件 |

#### 报表统计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/reports/inventory/summary` | 库存汇总报表 |
| GET | `/api/v1/reports/inventory/detail` | 库存明细报表 |
| GET | `/api/v1/reports/inventory/in-out` | 进出存报表 |
| GET | `/api/v1/reports/inventory/turnover` | 库存周转率 |
| GET | `/api/v1/reports/purchase/summary` | 采购汇总报表 |
| GET | `/api/v1/reports/purchase/tracking` | 采购订单跟踪 |
| GET | `/api/v1/reports/sales/summary` | 销售汇总报表 |
| GET | `/api/v1/reports/sales/tracking` | 销售订单跟踪 |
| GET | `/api/v1/reports/dashboard` | 首页仪表盘数据 |

#### 系统管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET    | `/api/v1/system/users` | 用户列表 |
| POST   | `/api/v1/system/users` | 创建用户 |
| PUT    | `/api/v1/system/users/:id` | 更新用户 |
| DELETE | `/api/v1/system/users/:id` | 删除用户 |
| GET    | `/api/v1/system/roles` | 角色列表 |
| POST   | `/api/v1/system/roles` | 创建角色 |
| PUT    | `/api/v1/system/roles/:id` | 更新角色 |
| POST   | `/api/v1/system/roles/:id/permissions` | 设置角色权限 |
| GET    | `/api/v1/system/dict/:type` | 获取字典数据 |
| POST   | `/api/v1/system/dict` | 创建字典 |
| GET    | `/api/v1/system/notifications` | 通知列表 |
| PUT    | `/api/v1/system/notifications/:id/read` | 标记已读 |
| GET    | `/api/v1/system/audit-logs` | 审计日志 |

---

## 6. 安全设计

### 6.1 认证流程

```
                      ┌─────────┐
                      │  登录页  │
                      └────┬────┘
                           │ POST /auth/login
                           │ {username, password}
                           ▼
                      ┌─────────┐
                      │ 后端校验 │
                      │ bcrypt  │
                      │ compare │
                      └────┬────┘
                           │ 返回
                           │ access_token (2h)
                           │ refresh_token (7d)
                           ▼
                      ┌─────────┐
                      │  前端   │
                      │ 存储    │
                      │ Token  │
                      └────┬────┘
                           │ 后续请求
                           │ Header: Authorization: Bearer <token>
                           ▼
                      ┌─────────┐
                      │ JWT中间件│ ←─ Token 过期 → 用 refresh_token 刷新
                      │ 解析校验  │
                      └─────────┘
```

### 6.2 JWT Token 设计

```json
// Access Token Payload
{
    "user_id": 1,
    "username": "admin",
    "real_name": "管理员",
    "roles": ["admin"],
    "exp": 1744340000,
    "iat": 1744332800
}
```

- Access Token 有效期：2 小时
- Refresh Token 有效期：7 天
- Refresh Token 存储在 Redis，支持主动失效

### 6.3 权限校验流程

```go
// 中间件链：
// 请求 → CORS → Logger → Recovery → JWT认证 → RBAC鉴权 → Handler

// RBAC 鉴权逻辑：
// 1. 从 JWT 中获取用户角色列表
// 2. 查询角色对应的权限列表（缓存在 Redis）
// 3. 比对当前请求的 API 路径 + 方法是否在权限列表中
// 4. 通过则继续，否则返回 403
```

### 6.4 敏感数据处理

| 数据 | 处理方式 |
|------|---------|
| 用户密码 | bcrypt 哈希存储，不可逆 |
| 采购单价 | API 层按角色过滤，无权限不返回价格字段 |
| 会话 Token | Redis 存储，支持强制下线 |
| 审计日志 | 只读不可修改，保留关键操作的变更前后数据 |
| 材质证明附件 | 文件路径不直接暴露，通过 API 代理访问 |

---

## 7. 前端设计

### 7.1 页面结构

```
┌─────────────────────────────────────────────────────┐
│                      顶部导航栏                       │
│  Logo  │               │ 通知 │ 用户头像 ▾          │
├────────┼───────────────────────────────────────────── ┤
│        │                                             │
│ 侧边   │                 主内容区                     │
│ 导航   │                                             │
│ 菜单   │  ┌─────────────────────────────────────┐    │
│        │  │         面包屑导航                    │    │
│ 首页   │  ├─────────────────────────────────────┤    │
│ 基础   │  │                                     │    │
│ 数据 ▸ │  │         页面内容                     │    │
│ 采购   │  │                                     │    │
│ 管理 ▸ │  │  ┌───────────────────────────────┐  │    │
│ 库存   │  │  │    搜索/筛选区                  │  │    │
│ 管理 ▸ │  │  ├───────────────────────────────┤  │    │
│ 销售   │  │  │    操作按钮栏                   │  │    │
│ 管理 ▸ │  │  ├───────────────────────────────┤  │    │
│ 追溯   │  │  │    数据表格                     │  │    │
│ 管理 ▸ │  │  │    ···                         │  │    │
│ 报表   │  │  ├───────────────────────────────┤  │    │
│ 统计 ▸ │  │  │    分页组件                     │  │    │
│ 系统   │  │  └───────────────────────────────┘  │    │
│ 管理 ▸ │  │                                     │    │
│        │  └─────────────────────────────────────┘    │
│        │                                             │
└────────┴─────────────────────────────────────────────┘
```

### 7.2 前端菜单结构

```
首页（仪表盘）
│
├── 基础数据
│   ├── 物料管理
│   ├── 物料分类
│   ├── 供应商管理
│   ├── 客户管理
│   └── 仓库管理
│
├── 采购管理
│   ├── 采购申请
│   ├── 采购订单
│   ├── 到货入库
│   └── 采购退货
│
├── 库存管理
│   ├── 库存台账
│   ├── 出库管理
│   ├── 库存调拨
│   ├── 库存盘点
│   └── 库存预警
│
├── 销售管理
│   ├── 销售订单
│   ├── 发货管理
│   └── 销售退货
│
├── 追溯管理
│   ├── 材质证明书
│   ├── 正向追溯
│   └── 反向追溯
│
├── 报表统计
│   ├── 库存报表
│   ├── 采购报表
│   └── 销售报表
│
└── 系统管理
    ├── 用户管理
    ├── 角色管理
    ├── 权限管理
    ├── 数据字典
    ├── 操作日志
    └── 系统参数
```

### 7.3 前端项目结构

```
web/
├── public/
│   └── favicon.ico
├── src/
│   ├── api/                    # API 请求模块
│   │   ├── auth.js
│   │   ├── material.js
│   │   ├── supplier.js
│   │   ├── customer.js
│   │   ├── warehouse.js
│   │   ├── purchase.js
│   │   ├── inventory.js
│   │   ├── sales.js
│   │   ├── certificate.js
│   │   ├── report.js
│   │   └── system.js
│   ├── assets/                 # 静态资源
│   │   ├── images/
│   │   └── styles/
│   │       ├── variables.scss  # 全局变量
│   │       ├── mixins.scss     # 混入
│   │       └── global.scss     # 全局样式
│   ├── components/             # 通用组件
│   │   ├── layout/             # 布局组件
│   │   │   ├── AppHeader.svelte
│   │   │   ├── AppSidebar.svelte
│   │   │   └── AppLayout.svelte
│   │   ├── common/             # 公共组件
│   │   │   ├── SearchForm.svelte  # 通用搜索表单
│   │   │   ├── DataTable.svelte   # 通用数据表格
│   │   │   ├── FileUpload.svelte  # 文件上传
│   │   │   └── BatchSelector.svelte # 批次选择器
│   │   └── business/           # 业务组件
│   │       ├── MaterialSelect.svelte  # 物料选择器
│   │       ├── SupplierSelect.svelte  # 供应商选择器
│   │       └── WarehouseSelect.svelte # 仓库选择器
│   ├── composables/            # 组合式函数
│   │   ├── useAuth.js
│   │   ├── usePagination.js
│   │   └── usePermission.js
│   ├── router/                 # 路由配置
│   │   └── index.js
│   ├── stores/                 # Svelte Stores 状态管理
│   │   ├── auth.js
│   │   ├── app.js
│   │   └── permission.js
│   ├── utils/                  # 工具函数
│   │   ├── client.ts           # API Client 封装
│   │   ├── auth.js             # Token 管理
│   │   ├── format.js           # 格式化工具
│   │   └── validate.js         # 校验工具
│   ├── views/                  # 页面视图
│   │   ├── dashboard/
│   │   │   └── DashboardView.svelte
│   │   ├── base/               # 基础数据
│   │   │   ├── MaterialList.svelte
│   │   │   ├── MaterialForm.svelte
│   │   │   ├── SupplierList.svelte
│   │   │   ├── CustomerList.svelte
│   │   │   └── WarehouseList.svelte
│   │   ├── purchase/           # 采购管理
│   │   │   ├── RequestList.svelte
│   │   │   ├── RequestForm.svelte
│   │   │   ├── OrderList.svelte
│   │   │   ├── OrderForm.svelte
│   │   │   └── StockInForm.svelte
│   │   ├── inventory/          # 库存管理
│   │   │   ├── StockList.svelte
│   │   │   ├── StockOutForm.svelte
│   │   │   ├── TransferForm.svelte
│   │   │   ├── CheckList.svelte
│   │   │   └── AlertList.svelte
│   │   ├── sales/              # 销售管理
│   │   │   ├── OrderList.svelte
│   │   │   ├── OrderForm.svelte
│   │   │   └── ShipmentForm.svelte
│   │   ├── trace/              # 追溯管理
│   │   │   ├── CertList.svelte
│   │   │   ├── CertForm.svelte
│   │   │   ├── ForwardTrace.svelte
│   │   │   └── BackwardTrace.svelte
│   │   ├── report/             # 报表
│   │   │   ├── InventoryReport.svelte
│   │   │   ├── PurchaseReport.svelte
│   │   │   └── SalesReport.svelte
│   │   ├── system/             # 系统管理
│   │   │   ├── UserList.svelte
│   │   │   ├── RoleList.svelte
│   │   │   ├── PermissionList.svelte
│   │   │   ├── DictList.svelte
│   │   │   └── AuditLogList.svelte
│   │   └── login/
│   │       └── LoginView.svelte
│   ├── App.svelte
│   └── main.js
├── index.html
├── vite.config.js
├── package.json
└── .env.development
```

### 7.4 UI 设计要点

| 要点 | 说明 |
|------|------|
| 色调 | 主色蓝色系（#409EFF），辅以工业灰，表达专业稳重 |
| 数据表格 | 紧凑型行高，支持列固定、排序、筛选 |
| 表单设计 | 主从表单模式，主表 + 明细行（可增删行）|
| 状态展示 | 使用 Tag 组件，不同状态对应不同颜色和图标 |
| 操作反馈 | 所有操作均有 Loading 状态和成功/失败提示 |
| 打印支持 | 入库单、出库单、盘点单支持打印 |

---

## 8. 部署架构

### 8.1 开发环境

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: teweicun
      POSTGRES_USER: twc
      POSTGRES_PASSWORD: twc_dev_2026
    ports:
      - "5432:5432"
    volumes:
      - pgdata_dev:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio_dev:/data

volumes:
  pgdata_dev:
  minio_dev:
```

### 8.2 生产环境

```
┌────────────────────────────────────┐
│            服务器/云主机             │
│                                    │
│  ┌──────────────────────────────┐  │
│  │           Nginx              │  │
│  │  ┌──────────┬───────────┐   │  │
│  │  │ 静态资源  │  API 代理  │   │  │
│  │  │ /dist/*  │  /api/* → │   │  │
│  │  └──────────┴────┬──────┘   │  │
│  └──────────────────┼──────────┘  │
│                     │              │
│  ┌──────────────────┼──────────┐  │
│  │    Go 后端服务     │         │  │
│  │    :8080          ◄         │  │
│  └───────┬──────────┬──────────┘  │
│          │          │              │
│  ┌───────▼───┐ ┌────▼──────┐      │
│  │ PostgreSQL│ │   Redis   │      │
│  │   :5432   │ │   :6379   │      │
│  └───────────┘ └───────────┘      │
│                                    │
│  ┌───────────┐                    │
│  │   MinIO   │                    │
│  │   :9000   │                    │
│  └───────────┘                    │
└────────────────────────────────────┘
```

### 8.3 Docker 部署

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/conf.d:/etc/nginx/conf.d
      - ./nginx/ssl:/etc/nginx/ssl
      - ./web/dist:/usr/share/nginx/html
    depends_on:
      - backend

  backend:
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      - TWC_DB_HOST=postgres
      - TWC_DB_PORT=5432
      - TWC_DB_NAME=teweicun
      - TWC_DB_USER=twc
      - TWC_DB_PASSWORD=${DB_PASSWORD}
      - TWC_REDIS_HOST=redis
      - TWC_JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres
      - redis
    restart: always

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: teweicun
      POSTGRES_USER: twc
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./backups:/backups
    restart: always

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    restart: always

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_USER}
      MINIO_ROOT_PASSWORD: ${MINIO_PASSWORD}
    volumes:
      - minio_data:/data
    restart: always

volumes:
  pgdata:
  minio_data:
```

### 8.4 Dockerfile

```dockerfile
# 多阶段构建
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o teweicun ./cmd/server

# 运行阶段
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/teweicun .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./teweicun"]
```

---

## 9. 开发规范

### 9.1 Git 工作流

```
main          ────●────────────●────────────●──── (稳定版本)
                  │            ▲            ▲
develop       ────●──●──●──●──●│──●──●──●──●│──── (开发版本)
                     │        ││           ││
feature/xxx   ───────●──●──●──┘│           ││
                               │           ││
feature/yyy   ─────────────────●──●──●──●──┘│
                                            │
hotfix/zzz    ──────────────────────────────●┘
```

- `main` 分支：稳定发布版本
- `develop` 分支：开发集成分支
- `feature/*` 分支：功能开发分支
- `hotfix/*` 分支：紧急修复分支

### 9.2 代码规范

**Go 后端**：

```go
// 命名规范
// - 文件名：小写下划线 (material_svc.go)
// - 结构体/接口：大驼峰 (MaterialService)
// - 方法/函数：大驼峰 (GetByID)
// - 变量：小驼峰 (materialCode)
// - 常量：大驼峰或全大写 (MaxPageSize, STATUS_ENABLED)

// Handler 方法签名规范
func (h *MaterialHandler) List(c *gin.Context) {
    // 1. 参数绑定与校验
    // 2. 调用 Service
    // 3. 返回响应
}

// Service 方法签名规范
func (s *MaterialService) List(ctx context.Context, req *dto.MaterialListReq) (*dto.PageResult, error) {
    // 1. 业务逻辑处理
    // 2. 调用 DB 层
    // 3. 组装返回数据
}

// DB 层方法签名规范
func FindMaterialByPage(ctx context.Context, pool *pgxpool.Pool, cond *MaterialQuery, page, size int) ([]MaterialRow, int64, error) {
    // 纯数据库操作
}
```

**Svelte 前端**：

```
命名规范：
- 组件文件：大驼峰 (MaterialList.svelte)
- 组合式函数：use开头 (useAuth.js)
- API 模块：小驼峰 (material.js)
- CSS 类名：BEM 命名（如需要）

组件规范：
- 统一使用 `lang="ts"`
- 使用 `$props` / `$state` / `$derived` / `$effect`
- 页面级数据优先在 `+page.server.ts` 处理
```

### 9.3 提交规范

```
<type>(<scope>): <subject>

type:
  feat     - 新功能
  fix      - 修复 Bug
  docs     - 文档变更
  style    - 代码格式（不影响逻辑）
  refactor - 重构
  perf     - 性能优化
  test     - 测试相关
  chore    - 构建/工具变更

示例：
  feat(material): 添加物料分类管理
  fix(inventory): 修复出库数量校验逻辑
  docs: 更新 API 文档
```

---

## 10. 开发计划

### 10.1 阶段划分

```
┌────────────────────────────────────────────────────────────────┐
│                    一期开发计划（约 12-16 周）                    │
├────────────┬───────────────────────────────────────────────────┤
│  阶段       │  内容                                            │
├────────────┼───────────────────────────────────────────────────┤
│ 第 1-2 周  │ 项目搭建                                          │
│            │ • Go 后端项目初始化（Gin + pgxpool + 手写 SQL）      │
│            │ • SvelteKit 前端项目初始化（Vite + Tailwind + DaisyUI）│
│            │ • 数据库初始化（创建表结构）                         │
│            │ • Docker Compose 开发环境配置                       │
│            │ • 认证模块（登录/登出/JWT）                         │
│            │ • 用户管理、角色权限管理                             │
├────────────┼───────────────────────────────────────────────────┤
│ 第 3-5 周  │ 基础数据模块                                       │
│            │ • 物料管理（CRUD + 分类树）                         │
│            │ • 供应商管理（CRUD + 合格供方）                      │
│            │ • 客户管理（CRUD）                                  │
│            │ • 仓库管理（CRUD + 库位可选）                       │
│            │ • 数据字典管理                                      │
├────────────┼───────────────────────────────────────────────────┤
│ 第 6-8 周  │ 采购与入库模块                                      │
│            │ • 采购申请（创建/审批流程）                          │
│            │ • 采购订单（创建/跟踪/关闭）                        │
│            │ • 到货入库（入库单/质检/批次管理）                   │
│            │ • 材质证明书管理（录入/确认/文件上传）               │
│            │ • 采购退货                                          │
├────────────┼───────────────────────────────────────────────────┤
│ 第 9-11 周 │ 库存与销售模块                                      │
│            │ • 库存台账（多维度查询）                             │
│            │ • 出库管理（销售出库/生产领料）                      │
│            │ • 库存调拨                                          │
│            │ • 库存盘点                                          │
│            │ • 库存预警                                          │
│            │ • 销售订单（创建/跟踪/发货）                        │
│            │ • 销售退货                                          │
├────────────┼───────────────────────────────────────────────────┤
│ 第 12-14 周│ 追溯/报表/优化                                      │
│            │ • 批次正向/反向追溯                                  │
│            │ • 库存报表（汇总/明细/进出存）                       │
│            │ • 采购/销售报表                                     │
│            │ • 首页仪表盘                                        │
│            │ • 通知中心                                          │
│            │ • 系统参数配置                                      │
├────────────┼───────────────────────────────────────────────────┤
│ 第 15-16 周│ 测试与上线                                          │
│            │ • 集成测试                                          │
│            │ • 部署文档                                          │
│            │ • 用户验收测试 (UAT)                                │
│            │ • 生产环境部署                                      │
│            │ • 初始数据导入（物料基础数据）                       │
└────────────┴───────────────────────────────────────────────────┘
```

### 10.2 里程碑

| 里程碑 | 时间节点 | 交付内容 |
|-------|---------|---------|
| M1 - 基础框架 | 第 2 周末 | 可登录的系统框架，用户/权限管理可用 |
| M2 - 基础数据 | 第 5 周末 | 物料/供应商/客户/仓库 CRUD 完成 |
| M3 - 采购入库 | 第 8 周末 | 完整的采购→入库流程可跑通 |
| M4 - 库存销售 | 第 11 周末 | 出入库管理 + 销售流程完成 |
| M5 - 功能完整 | 第 14 周末 | 追溯/报表/仪表盘完成 |
| M6 - 正式上线 | 第 16 周末 | 通过 UAT，生产环境可用 |

### 10.3 风险与应对

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|---------|
| 业务理解偏差 | 返工 | 中 | 尽早出原型，每次迭代邀请用户验收 |
| 特种设备法规变更 | 功能调整 | 低 | 关注 TSG 法规更新，数据字典可配置化 |
| 性能瓶颈 | 用户体验差 | 低 | 预留索引优化空间，报表考虑异步 |
| 数据迁移风险 | 历史数据丢失 | 中 | 提前制定数据迁移方案，充分测试 |
| 人力不足 | 延期 | 中 | 优先核心功能，非核心功能推迟到二期 |

---

## 附录

### A. 配置文件示例

```yaml
# configs/config.yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: release          # debug/release

database:
  host: localhost
  port: 5432
  name: teweicun
  user: twc
  password: ""           # 生产环境从环境变量读取
  max_open_conns: 20
  max_idle_conns: 5
  conn_max_lifetime: 30m

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: ""             # 生产环境从环境变量读取
  access_expire: 2h
  refresh_expire: 168h   # 7天

storage:
  type: minio            # local/minio
  minio:
    endpoint: localhost:9000
    access_key: minioadmin
    secret_key: minioadmin
    bucket: teweicun
    use_ssl: false
  local:
    upload_dir: ./uploads

log:
  level: info            # debug/info/warn/error
  format: json           # json/console
  output: stdout         # stdout/file
  file_path: ./logs/app.log
  max_size: 100          # MB
  max_backups: 10
  max_age: 30            # 天
```

### B. Nginx 配置示例

```nginx
# nginx/conf.d/teweicun.conf
server {
    listen 80;
    server_name your-domain.com;

    # 前端静态资源
    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;

        # 缓存策略
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff2?)$ {
            expires 30d;
            add_header Cache-Control "public, immutable";
        }
    }

    # API 反向代理
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 文件上传大小限制
        client_max_body_size 50M;
    }
}
```

### C. Makefile 示例

```makefile
.PHONY: help build run test migrate docker-up docker-down

# 默认目标
help:
	@echo "特维存 - 可用命令:"
	@echo "  make build       - 编译后端"
	@echo "  make run         - 运行后端（开发）"
	@echo "  make test        - 运行测试"
	@echo "  make migrate-up  - 执行数据库迁移"
	@echo "  make migrate-down- 回滚数据库迁移"
	@echo "  make docker-up   - 启动 Docker 开发环境"
	@echo "  make docker-down - 停止 Docker 开发环境"
	@echo "  make lint        - 代码检查"
	@echo "  make swagger     - 生成 API 文档"
	@echo "  make web-dev     - 启动前端开发服务"
	@echo "  make web-build   - 构建前端"

# 后端
build:
	CGO_ENABLED=0 go build -o bin/teweicun ./cmd/server

run:
	go run ./cmd/server -c configs/config.dev.yaml

test:
	go test ./... -v -cover

lint:
	golangci-lint run ./...

swagger:
	swag init -g cmd/server/main.go -o api

# 数据库迁移
migrate-up:
	migrate -path migrations -database "postgresql://twc:twc_dev_2026@localhost:5432/teweicun?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://twc:twc_dev_2026@localhost:5432/teweicun?sslmode=disable" down 1

# Docker
docker-up:
	docker compose -f docker-compose.dev.yml up -d

docker-down:
	docker compose -f docker-compose.dev.yml down

# 前端
web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build
```
