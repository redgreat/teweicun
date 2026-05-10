#!/bin/bash
#
# 功能：执行数据库迁移脚本
# 创建时间：2026-04-18
# 创建人：CodeArts Agent
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=== TeWeiCun 数据库迁移脚本 ==="
echo ""

CONFIG_FILE="${CONFIG_FILE:-${PROJECT_ROOT}/configs/config.yaml}"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 配置文件不存在: $CONFIG_FILE"
    exit 1
fi

get_db_field() {
    local field="$1"
    awk -F': *' -v key="$field" '
        $1 == "database" { in_db=1; next }
        in_db && $0 ~ /^[^[:space:]]/ { in_db=0 }
        in_db && $1 ~ ("^[[:space:]]*" key "$") {
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", $2);
            print $2;
            exit;
        }
    ' "$CONFIG_FILE"
}

DB_HOST=$(get_db_field "host")
DB_PORT=$(get_db_field "port")
DB_NAME=$(get_db_field "dbname")
DB_USER=$(get_db_field "user")
DB_PASSWORD=$(get_db_field "password")
DB_SSLMODE=$(get_db_field "sslmode")

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

MIGRATION_FILES=()
while IFS= read -r file; do
    MIGRATION_FILES+=("$file")
done < <(ls -1 "$MIGRATION_DIR"/*.sql 2>/dev/null | sort)
if [ ${#MIGRATION_FILES[@]} -eq 0 ]; then
    echo "错误: 未找到迁移脚本"
    exit 1
fi

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

export PGPASSWORD="$DB_PASSWORD"
if [ -n "$DB_SSLMODE" ]; then
    export PGSSLMODE="$DB_SSLMODE"
fi

PSQL_BASE_CMD=(psql -v ON_ERROR_STOP=1 -X -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME")

"${PSQL_BASE_CMD[@]}" -c "
CREATE TABLE IF NOT EXISTS schema_migration_history (
    migration_id varchar(255) PRIMARY KEY,
    filename varchar(255) NOT NULL,
    file_checksum varchar(128) NOT NULL,
    applied_at timestamp with time zone NOT NULL DEFAULT NOW()
);
"

echo "可用的迁移脚本:"
for migration_file in "${MIGRATION_FILES[@]}"; do
    filename=$(basename "$migration_file")
    migration_id=$(grep -E '^--[[:space:]]*MIGRATION_ID:' "$migration_file" | head -n 1 | sed -E 's/^--[[:space:]]*MIGRATION_ID:[[:space:]]*//')
    if [ -z "$migration_id" ]; then
        echo "错误: $filename 缺少头部标识 '-- MIGRATION_ID:'"
        exit 1
    fi
    echo "  - $filename  ($migration_id)"
done
echo ""

for migration_file in "${MIGRATION_FILES[@]}"; do
    filename=$(basename "$migration_file")
    migration_id=$(grep -E '^--[[:space:]]*MIGRATION_ID:' "$migration_file" | head -n 1 | sed -E 's/^--[[:space:]]*MIGRATION_ID:[[:space:]]*//')
    escaped_id=$(printf "%s" "$migration_id" | sed "s/'/''/g")
    applied=$("${PSQL_BASE_CMD[@]}" -t -A -c "SELECT 1 FROM schema_migration_history WHERE migration_id = '$escaped_id' LIMIT 1;")

    if [ "$applied" = "1" ]; then
        echo "跳过: $filename ($migration_id) 已执行"
        echo ""
        continue
    fi

    echo "执行: $filename ($migration_id)"
    output_file="$(mktemp)"
    if "${PSQL_BASE_CMD[@]}" -f "$migration_file" >"$output_file" 2>&1; then
        while IFS= read -r line; do
            echo "  $line"
        done <"$output_file"

        checksum=$(shasum -a 256 "$migration_file" | awk '{print $1}')
        escaped_filename=$(printf "%s" "$filename" | sed "s/'/''/g")
        escaped_checksum=$(printf "%s" "$checksum" | sed "s/'/''/g")
        "${PSQL_BASE_CMD[@]}" -c "
            INSERT INTO schema_migration_history (migration_id, filename, file_checksum)
            VALUES ('$escaped_id', '$escaped_filename', '$escaped_checksum')
            ON CONFLICT (migration_id) DO NOTHING;
        " >/dev/null

        if grep -q -E '^--[[:space:]]*MIGRATION_APPLIED:' "$migration_file"; then
            applied_tag="applied_$(date '+%Y%m%dT%H%M%S')"
            sed -i '' -E "s/^--[[:space:]]*MIGRATION_APPLIED:.*/-- MIGRATION_APPLIED: ${applied_tag}/" "$migration_file"
        fi

        echo "  ✓ 成功"
        echo ""
    else
        while IFS= read -r line; do
            echo "  $line"
        done <"$output_file"
        echo "  ✗ 失败"
        rm -f "$output_file"
        exit 1
    fi
    rm -f "$output_file"
done

echo "=== 迁移完成 ==="
