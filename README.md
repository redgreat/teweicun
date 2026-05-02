# 特维存 (TeWeiCun)

面向特种设备制造企业的进销存管理系统，覆盖基础数据、采购、库存、销售、追溯与系统管理等核心业务。  
项目采用 **DB-First** 架构：核心业务流程下沉 PostgreSQL，后端使用 Go + 手写 SQL 调用数据库对象。

## 项目特性

- DB-First：核心业务通过存储过程/函数/视图实现，Go 层负责编排与接口
- 纯 SQL 数据访问：禁用 ORM（如 GORM/sqlx/ent）
- 前后端分离：后端 Go（Gin），前端 SvelteKit
- 容器化交付：基于 Docker / Compose 运行与部署
- 发布自动化：推送 Git Tag 后由 GitHub Actions 构建并推送镜像

## 技术栈

- 后端：Go、Gin、pgxpool、Zap、Viper
- 前端：SvelteKit、Tailwind CSS、DaisyUI
- 数据库：PostgreSQL 16+
- 缓存：Redis 7+
- 存储：MinIO
- CI/CD：GitHub Actions（Aliyun ACR + GHCR）

## 目录说明

```text
.
├── cmd/                  # 应用入口
├── internal/             # 后端业务代码
├── web/                  # 前端项目
├── sql/                  # 数据库脚本（迁移/函数/过程/视图）
├── scripts/              # 运维与发布脚本
├── configs/              # 配置文件
├── docs/                 # 项目文档
└── .github/workflows/    # CI/CD 工作流
```

## 本地开发

### 环境要求

- Go（建议 1.22+）
- Node.js（建议 20+）
- Docker / Docker Compose
- PostgreSQL（如不使用容器）

### 常用命令

```bash
# 查看可用命令
make help

# 启动后端
make run

# 后端测试
make test

# Go 代码检查
make lint

# 前端 lint / 检查
make lint-web
make check-web
```

### 一键开发环境（可选）

```bash
# 首次执行可能需要授权
chmod +x scripts/*.sh

# 启动开发环境
./scripts/dev.sh up

# 查看日志
./scripts/dev.sh logs

# 停止并清理
./scripts/dev.sh down
```

## 发布流程（当前规范）

当前发布脚本为“纯推 Tag”模式：**不做本地构建、不做本地测试检查**。  
镜像构建与推送由 GitHub Actions 完成。

### 1) 本地创建并推送 Tag

Linux/macOS:

```bash
./scripts/release.sh
# 或指定版本
./scripts/release.sh v1.2.0
```

Windows PowerShell:

```powershell
.\scripts\release.ps1
# 或指定版本
.\scripts\release.ps1 v1.2.0
```

也可统一使用：

```bash
make release
```

### 2) Actions 自动构建与推送

所有发布相关步骤集中在 **一个** 工作流文件：`.github/workflows/release.yml`。

- **触发条件**：
  - 推送 **`v*` 版本标签**（例如 `v1.2.0`），或
  - 在 Actions 中 **手动运行（workflow_dispatch）**，并填写已存在的同名标签（如 `v1.2.0`）
- **普通分支 push / PR**：不触发该工作流（不在 GitHub Actions 里做打包或检查）
- **工作流内容**：
  1. 构建前端产物
  2. 构建 Docker 镜像并同时推送到 **GHCR** 与 **阿里云 ACR**
  3. 创建 **GitHub Release**（含变更说明与镜像拉取命令）

日常代码质量请在本地执行：`make lint`、`make test`、`make lint-web` 等。

## GitHub Secrets 配置

若要让发布工作流正常推送 **阿里云 ACR**，需要在仓库 Secrets 中配置：

- `ALIYUN_USERNAME`：阿里云 ACR 登录用户名
- `ALIYUN_PASSWORD`：阿里云 ACR 登录密码/令牌
- `ALIYUN_NAMESPACE`：ACR 命名空间
- `DOCKER_IMAGE_NAME`：镜像名称（不含 registry 和 namespace）

说明：

- GHCR 推送使用 `GITHUB_TOKEN`，通常无需额外配置
- 建议在组织/仓库层统一管理上述 Secrets，避免重复维护

## 相关脚本

- `scripts/release.sh` / `scripts/release.ps1`：创建并推送 Git Tag
- `scripts/redeploy.sh` / `scripts/redeploy.ps1`：服务端重部署
- `scripts/dev.sh`：本地 Docker 开发环境管理

## 参考文档

- `docs/requirements.md`：需求文档
- `docs/design.md`：系统设计文档
- `docs/database_design.md`：数据库设计文档
- `docs/coding_standards.md`：代码规范
- `docs/testing_standards.md`：测试规范

---

© 2026 RedGreat / TeWeiCun
