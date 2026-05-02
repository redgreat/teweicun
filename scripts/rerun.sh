#!/bin/bash

# 进入指定的绝对路径目录
cd /root/twc || { echo "进入目录失败，中止执行"; exit 1; }

echo "=========================================="
echo "    开始重新部署 twc 容器环境"
echo "=========================================="

echo "➤ 1. 停止并移除旧容器: twc ..."
docker-compose stop twc
docker-compose rm -f twc

echo "➤ 2. 删除本地所有的 twc 镜像记录..."
# 注意：这会查找带有 twc 名称的镜像，并按 ID 强制删除
docker image ls -q --filter "reference=*twc*" | xargs -r docker rmi -f

echo "➤ 3. 清理所有旧的日志数据..."
# 删除根目录下 ./logs/ 内的所有文件及文件夹
rm -rf ./logs/*
echo "日志清理完成。"

echo "➤ 4. 重新拉取最新镜像并后台启动..."
docker-compose pull twc
docker-compose up -d twc

echo "等待 2 秒检查容器状态..."
sleep 2

echo "➤ 5. 当前 twc 容器运行状态:"
docker ps -a --filter "name=twc"

echo "=========================================="
echo "    容器已更新并重启完成！"
echo "=========================================="
