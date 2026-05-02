#!/bin/bash

# =================================================================
# TeWeiCun (特维存) Git 标签发布脚本
# 功能：自动计算 Git Tag，创建并推送标签（不做本地构建和测试检查）
# =================================================================

Set-StrictMode() {
    set -e
}
Set-StrictMode

# 颜色定义
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log_info()    { echo -e "${BLUE}➤ $1${NC}"; }
log_success() { echo -e "${GREEN}✔ $1${NC}"; }
log_warn()    { echo -e "${YELLOW}⚠ $1${NC}"; }
log_error()   { echo -e "${RED}✖ $1${NC}"; exit 1; }

# 2. 版本计算逻辑
VERSION=$1
if [ -z "$VERSION" ]; then
    log_info "未指定版本号，正在自动获取最新标签..."
    LATEST_TAG=$(git tag --list 'v*' --sort=-version:refname | head -n 1)
    
    if [ -z "$LATEST_TAG" ]; then
        log_warn "未发现任何标签，使用默认 v0.0.1"
        VERSION="v0.0.1"
    else
        log_info "发现最新标签: $LATEST_TAG"
        # 简单的自动增量逻辑: v1.0.1 -> v1.0.2
        VERSION=$(echo $LATEST_TAG | awk -F. '{$NF = $NF + 1;} 1' OFS=.)
        log_info "自动计算下一个版本: $VERSION"
    fi
fi

echo -e "${BLUE}=============================================="
echo -e "   发布流程启动: $VERSION"
echo -e "==============================================${NC}"

# 3. 创建并推送 Git Tag（不做本地构建/测试）
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    log_warn "标签 $VERSION 已存在，跳过创建。"
else
    log_info "创建 Git 标签 $VERSION ..."
    git tag "$VERSION"
    log_success "Git 标签 $VERSION 创建成功。"
fi

log_info "推送标签到远程 origin ..."
git push origin "$VERSION"
log_success "标签 $VERSION 推送完成。"
