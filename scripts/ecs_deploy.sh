#!/bin/bash
#
# 腾讯云 ECS 一键部署脚本
# 功能：初始化 Docker 环境、配置 GHCR 登录、启动 Watchtower 自动更新
# 使用方式：在 ECS 上执行 bash scripts/ecs_deploy.sh
#
# 前置条件：
#   1. ECS 已安装 Docker 和 Docker Compose
#   2. 已创建 GitHub Personal Access Token (read:packages 权限)
#   3. 已将 configs/config.yaml 配置为 ECS 环境
#

set -euo pipefail

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'

echo "========================================"
echo "  特维存 ECS 生产部署脚本"
echo "========================================"

# ── 1. 检查 Docker ──
if ! command -v docker &> /dev/null; then
    echo "请先安装 Docker: https://docs.docker.com/engine/install/"
    exit 1
fi
echo -e "${GREEN}✓${NC} Docker 已安装: $(docker --version)"

# ── 2. 检查 docker compose ──
if ! docker compose version &> /dev/null; then
    echo "请确保 Docker Compose 可用"
    exit 1
fi
echo -e "${GREEN}✓${NC} Docker Compose 可用"

# ── 3. 登录 GHCR ──
if [ -z "${GITHUB_TOKEN:-}" ]; then
    echo -e "${YELLOW}请设置 GITHUB_TOKEN 环境变量${NC}"
    echo ""
    echo "创建步骤："
    echo "  1. 打开 https://github.com/settings/tokens"
    echo "  2. 点击 'Generate new token (classic)'"
    echo "  3. 勾选 'read:packages' 权限"
    echo "  4. 复制生成的 Token"
    echo ""
    echo "然后执行:"
    echo "  export GITHUB_TOKEN=ghp_xxxxxxxxxxxx"
    echo "  bash scripts/ecs_deploy.sh"
    exit 1
fi

echo "$GITHUB_TOKEN" | docker login ghcr.io -u redgreat --password-stdin
echo -e "${GREEN}✓${NC} GHCR 登录成功"

# ── 4. 创建必要目录 ──
mkdir -p configs uploads logs
echo -e "${GREEN}✓${NC} 目录就绪"

# ── 5. 拉取最新镜像 ──
echo "拉取最新镜像..."
docker compose -f docker-compose-prod.yml pull twc

# ── 6. 启动服务 ──
echo "启动服务..."
docker compose -f docker-compose-prod.yml up -d

# ── 7. 等待健康检查 ──
echo "等待服务就绪..."
for i in $(seq 1 15); do
    if curl -sf http://localhost:8080/api/v1/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} 服务启动成功！"
        echo ""
        echo "========================================"
        echo "  部署完成"
        echo "  访问: http://$(curl -s ifconfig.me 2>/dev/null || echo 'YOUR_IP'):8080"
        echo ""
        echo "  自动更新: Watchtower 每5分钟检查新镜像"
        echo "  查看日志: docker compose -f docker-compose-prod.yml logs -f"
        echo "  手动更新: docker compose -f docker-compose-prod.yml pull && docker compose -f docker-compose-prod.yml up -d"
        echo "========================================"
        exit 0
    fi
    sleep 2
done

echo -e "${YELLOW}服务启动中，请稍后访问${NC}"
docker compose -f docker-compose-prod.yml logs --tail=20
