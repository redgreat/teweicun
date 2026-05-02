#!/bin/bash

# =================================================================
# TeWeiCun (特维存) 本地调试Docker镜像启动脚本
# 功能：不触动数据库镜像 -> 构建后端/前端 -> 精准重启业务容器 -> 自动 SQL 初始化
# =================================================================

COMPOSE_FILE="docker-compose-local.yml"

function full_deploy() {
    echo "================================================"
    echo " 1. 构建业务镜像（前后端一次性构建）"
    echo "================================================"
    docker compose -f $COMPOSE_FILE build backend frontend
    if [ $? -ne 0 ]; then
        echo "❌ 镜像构建失败，退出！"
        exit 1
    fi

    echo "================================================"
    echo " 2. 仅停止并删除业务容器"
    echo "================================================"

    docker compose -f $COMPOSE_FILE rm -f -s backend frontend

    echo "================================================"
    echo " 3. 启动所有容器（不重复构建）"
    echo "================================================"
    docker compose -f $COMPOSE_FILE up -d --no-build
    if [ $? -ne 0 ]; then
        echo "❌ 容器启动失败，退出！"
        exit 1
    fi

    echo "================================================"
    echo " 4. 部署完成"
    echo "================================================"

    echo ""
    echo "✅ 环境就绪！"
    echo "🔥 前端: http://localhost:3000 | 后端: http://localhost:8080"
}

# 指令路由
COMMAND=$1
case $COMMAND in
  "down")
    # 只有显式执行 down 时才会彻底摧毁基座
    docker compose -f $COMPOSE_FILE down
    ;;
  "logs")
    docker compose -f $COMPOSE_FILE logs -f
    ;;
  *)
    full_deploy
    ;;
esac
