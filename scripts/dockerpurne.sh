#!/usr/bin/env zsh
#
# 功能：清理 Docker 构建缓存与未使用资源，释放磁盘空间（macOS）
# 创建时间：2026-05-08
# 创建人：GPT-5.3-Codex
#

set -euo pipefail

if ! command -v docker >/dev/null 2>&1; then
  echo "未找到 docker 命令，请先安装 Docker Desktop"
  exit 1
fi

echo "1) 清理 builder 缓存..."
docker builder prune -af

echo "2) 清理 buildx 缓存..."
docker buildx prune -af

echo "3) 清理已停止容器..."
docker container prune -f

echo "4) 清理未使用网络..."
docker network prune -f

echo "5) 清理未使用数据卷..."
docker volume prune -f

echo "6) 仅清理悬空镜像（不删除已打 tag 的基础镜像）..."
docker image prune -f

echo "完成。已避免使用 docker system prune -a，基础镜像（如 alpine:latest/golang:1.25-alpine/node:20-slim）会保留。"
