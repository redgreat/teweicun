#!/usr/bin/env bash
# 功能：本地联调启动后端与前端（非 Docker）；依赖 Go、nvm、npm。
# 创建时间：2026-04-25
# 创建人：Cursor Agent

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
if [ ! -s "$NVM_DIR/nvm.sh" ]; then
	echo "未找到 nvm（期望 \$NVM_DIR/nvm.sh）。请先安装 nvm: https://github.com/nvm-sh/nvm" >&2
	exit 1
fi
# shellcheck source=/dev/null
. "$NVM_DIR/nvm.sh"

if [ -f "$ROOT/.nvmrc" ]; then
	(cd "$ROOT" && nvm use)
else
	echo "仓库根目录缺少 .nvmrc" >&2
	exit 1
fi

cleanup() {
	if [ -n "${BACK_PID:-}" ] && kill -0 "$BACK_PID" 2>/dev/null; then
		echo ""
		echo "正在停止后端 (pid $BACK_PID)..."
		kill "$BACK_PID" 2>/dev/null || true
		wait "$BACK_PID" 2>/dev/null || true
	fi
}
trap cleanup EXIT INT TERM

echo "启动后端: go run ./cmd/server/main.go"
go run ./cmd/server/main.go &
BACK_PID=$!

sleep 2

echo "启动前端: npm run dev（Ctrl+C 将结束前端并停止后端）"
cd "$ROOT/frontend"
npm run dev
