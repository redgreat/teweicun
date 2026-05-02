SHELL := /bin/bash
# 未全局安装 golangci-lint 时，用 go run 拉取该版本（首跑会下载依赖，较慢）
GOLANGCI_VERSION ?= v2.7.2
GOPATH_BIN := $(shell go env GOPATH)/bin

.PHONY: help run test test-flow build lint lint-web check-web local export seed init-pass redeploy release web-dev

help:
	@echo "可用命令列表:"
	@echo "  export - 导出数据库结构到schema.sql"
	@echo "  seed   - 导出seed数据"
	@echo "  init-pass     - 初始化管理员密码(默认admin123)"
	@echo "  redeploy      - 本地重新部署"
	@echo "  release       - 发布新版本（仅创建并推送 Git Tag）"
	@echo "  run           - 启动后端服务"
	@echo "  test          - 运行后端测试 + 全流程接口测试（purchase->stock->consumption->reversal->return）"
	@echo "  test-flow     - 仅运行全流程接口测试（TestFlow_PurchaseToReversalAndReturn）"
	@echo "  build         - 编译后端可执行文件"
	@echo "  lint          - Go 代码检查 (golangci-lint；未安装则用 go run $(GOLANGCI_VERSION))"
	@echo "  lint-web      - 前端 ESLint + Prettier（需 nvm，见仓库根 .nvmrc）"
	@echo "  check-web     - 前端 svelte-check（需 nvm）"
	@echo "  local         - 依次执行 lint、test、lint-web、check-web 通过后，本地启动 Go + Vite（不用 Docker）"
	@echo "  web-dev       - 仅启动前端开发服务器（需自行 nvm use）"

run:
	go run cmd/server/main.go

# 不用 ./...：会扫到 frontend/node_modules 下第三方 .go；不用默认 -cover：部分环境缺少 go tool covdata
TEST_PKGS := ./internal/... ./pkg/... ./cmd/...
FLOW_TEST_FLAGS ?= -count=1

test:
	go test -v $(TEST_PKGS)
	$(MAKE) test-flow

test-flow:
	go test -v $(FLOW_TEST_FLAGS) ./test/... -run TestFlow_PurchaseToReversalAndReturn

build:
	go build -o tmp/api-server cmd/server/main.go

lint:
	@PATH="$(GOPATH_BIN):$$PATH"; \
	if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 不在 PATH 中，使用 go run $(GOLANGCI_VERSION)（首次较慢）…"; \
		go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION) run; \
	fi

lint-web:
	@export NVM_DIR="$${NVM_DIR:-$$HOME/.nvm}"; \
	[ -s "$$NVM_DIR/nvm.sh" ] && . "$$NVM_DIR/nvm.sh"; \
	[ -f .nvmrc ] && nvm use; \
	cd frontend && npm run lint

check-web:
	@export NVM_DIR="$${NVM_DIR:-$$HOME/.nvm}"; \
	[ -s "$$NVM_DIR/nvm.sh" ] && . "$$NVM_DIR/nvm.sh"; \
	[ -f .nvmrc ] && nvm use; \
	cd frontend && npm run check

ifeq ($(OS),Windows_NT)
local:
	@echo "make local 当前仅支持 macOS / Linux（bash + nvm）。Windows 可在 Git Bash 中执行 bash scripts/local_dev.sh，或开两个终端分别 go run / npm run dev。"
	@exit 1
else
local: lint test lint-web check-web
	@bash scripts/local_dev.sh
endif

web-dev:
	cd frontend && npm run dev

export:
	@echo "导出数据库结构..."
	go run scripts/schema_manager.go export

seed:
	@echo "导出seed数据..."
	go run scripts/schema_manager.go seed

init-pass:
	@echo "初始化管理员密码..."
	go run scripts/init_admin_pass.go

redeploy:
ifeq ($(OS),Windows_NT)
	@pwsh -NoProfile -ExecutionPolicy Bypass -Command "[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; Write-Host 'Local redeploy...'"
	@pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/redeploy.ps1
else
	@echo "Local redeploy..."
	@bash scripts/redeploy.sh
endif

release:
ifeq ($(OS),Windows_NT)
	@pwsh -NoProfile -ExecutionPolicy Bypass -Command "[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; Write-Host 'Release tag push...'"
	@pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/release.ps1
else
	@echo "Release tag push..."
	@bash scripts/release.sh
endif
