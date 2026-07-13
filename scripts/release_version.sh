#!/bin/bash
# 自动发布脚本
# 用法: bash scripts/release.sh [版本号]
# 示例: bash scripts/release.sh v1.0.4
#       bash scripts/release.sh patch  (自动递增补丁版本)
set -euo pipefail

cd "$(dirname "$0")/.."

VERSION_FILE="frontend/src/lib/version.ts"
PACKAGE_JSON="frontend/package.json"

# 获取当前版本
current=$(grep "APP_VERSION" "$VERSION_FILE" | sed "s/.*'\(.*\)'.*/\1/" | sed 's/^v//')
echo "当前版本: v$current"

# 确定新版本
if [ -n "${1:-}" ]; then
	if [ "$1" = "patch" ]; then
		IFS='.' read -r major minor patch <<< "$current"
		patch=$((patch + 1))
		new="v${major}.${minor}.${patch}"
	elif [ "$1" = "minor" ]; then
		IFS='.' read -r major minor patch <<< "$current"
		minor=$((minor + 1))
		new="v${major}.${minor}.0"
	else
		new="$1"
	fi
else
	echo "用法: bash scripts/release.sh [patch|minor|v1.0.x]"
	exit 1
fi
echo "新版本: $new"

# 更新 version.ts
sed -i '' "s/APP_VERSION = 'v[^']*'/APP_VERSION = '$new'/" "$VERSION_FILE"
echo "  ✓ $VERSION_FILE"

# 更新 package.json
sed -i '' "s/\"version\": \"[^\"]*\"/\"version\": \"${new#v}\"/" "$PACKAGE_JSON" 2>/dev/null || true
echo "  ✓ $PACKAGE_JSON"

# Git 提交
git add "$VERSION_FILE" "$PACKAGE_JSON"
git commit -m "release: $new"
git tag "$new"
git push origin main --tags

echo ""
echo "✅ 发布完成: $new"
echo "   GitHub Actions 构建中 → ECS 5分钟自动更新"
