#!/bin/bash
#
# 功能：执行数据库迁移脚本
# 创建时间：2026-04-18
# 创建人：CodeArts Agent
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=== TeWeiCun 数据库迁移脚本 ==="
echo ""

CONFIG_FILE="${CONFIG_FILE:-${PROJECT_ROOT}/configs/config.yaml}"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 配置文件不存在: $CONFIG_FILE"
    exit 1
fi

DB_HOST=$(grep -A 10 "database:" "$CONFIG_FILE" | grep "host:" | awk '{print $2}')
DB_PORT=$(grep -A 10 "database:" "$CONFIG_FILE" | grep "port:" | awk '{print $2}')
DB_NAME=$(grep -A 10 "database:" "$CONFIG_FILE" | grep "dbname:" | awk '{print $2}')
DB_USER=$(grep -A 10 "database:" "$CONFIG_FILE" | grep "user:" | awk '{print $2}')
DB_PASSWORD=$(grep -A 10 "database:" "$CONFIG_FILE" | grep "password:" | awk '{print $2}')
DB_SSLMODE=$(grep -A 10 "database:" "$CONFIG_FILE" | grep "sslmode:" | awk '{print $2}')

if [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ] || [ -z "$DB_NAME" ] || [ -z "$DB_USER" ]; then
    echo "错误: 无法从配置文件读取数据库连接信息"
    exit 1
fi

echo "数据库连接信息:"
echo "  主机: $DB_HOST"
echo "  端口: $DB_PORT"
echo "  数据库: $DB_NAME"
echo "  用户: $DB_USER"
if [ -n "$DB_SSLMODE" ]; then
    echo "  SSL: $DB_SSLMODE"
fi
echo ""

MIGRATION_DIR="${PROJECT_ROOT}/sql/migrations"
if [ ! -d "$MIGRATION_DIR" ]; then
    echo "错误: 迁移脚本目录不存在: $MIGRATION_DIR"
    exit 1
fi

echo "可用的迁移脚本:"
ls -1 "$MIGRATION_DIR"/*.sql 2>/dev/null | while read file; do
    echo "  - $(basename "$file")"
done
echo ""

AUTO_YES=0
if [ "$1" = "-y" ] || [ "$1" = "--yes" ]; then
    AUTO_YES=1
fi

if [ $AUTO_YES -eq 0 ]; then
    read -p "是否执行所有迁移脚本? (y/n): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "已取消"
        exit 0
    fi
fi

echo ""
echo "开始执行迁移..."
echo ""

for migration_file in "$MIGRATION_DIR"/*.sql; do
    if [ -f "$migration_file" ]; then
        filename=$(basename "$migration_file")
        echo "执行: $filename"
        
        export PGPASSWORD="$DB_PASSWORD"
        if [ -n "$DB_SSLMODE" ]; then
            export PGSSLMODE="$DB_SSLMODE"
        fi
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration_file" 2>&1 | while read line; do
            echo "  $line"
        done
        
        if [ ${PIPESTATUS[0]} -eq 0 ]; then
            echo "  ✓ 成功"
        else
            echo "  ✗ 失败"
            exit 1
        fi
        echo ""
    fi
done

echo "=== 迁移完成 ==="
