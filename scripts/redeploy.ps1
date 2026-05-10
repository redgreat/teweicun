#!/usr/bin/env pwsh

<#
.SYNOPSIS
  TeWeiCun (特维存) 本地调试Docker镜像启动脚本
.DESCRIPTION
  功能：不触动数据库镜像 -> 构建单镜像业务容器 -> 精准重启业务容器
.EXAMPLE
  .\scripts\redepoly.ps1          # 默认执行完整部署
  .\scripts\redepoly.ps1 down     # 彻底摧毁所有容器
  .\scripts\redepoly.ps1 logs     # 查看实时日志
#>

param(
  [Parameter(Position=0)]
  [ValidateSet('down', 'logs', '')]
  [string]$Command = ''
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# 设置控制台编码为UTF-8
$OutputEncoding = [System.Text.Encoding]::UTF8
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# 项目根目录
$ProjectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $ProjectRoot

$COMPOSE_FILE = 'docker-compose-local.yml'

# 彩色日志函数
function Write-Info($msg)    { Write-Host "[INFO] $msg" -ForegroundColor Blue }
function Write-Success($msg) { Write-Host "[OK] $msg" -ForegroundColor Green }
function Write-Warn($msg)    { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-ErrorMsg($msg){ Write-Host "[ERR] $msg" -ForegroundColor Red }

function Full-Deploy {
  # 1. Build app images once
  Write-Host ""
  Write-Host "================================================" -ForegroundColor Blue
  Write-Host " 1. Build app image (single image)" -ForegroundColor Blue
  Write-Host "================================================" -ForegroundColor Blue
  docker-compose -f $COMPOSE_FILE build app
  if ($LASTEXITCODE -ne 0) {
    Write-ErrorMsg 'Image build failed. Exit.'
    exit 1
  }

  # 2. Stop and remove app containers
  Write-Host ""
  Write-Host "================================================" -ForegroundColor Blue
  Write-Host " 2. Stop and remove app containers" -ForegroundColor Blue
  Write-Host "================================================" -ForegroundColor Blue
  # Remove only app containers, keep base services
  docker-compose -f $COMPOSE_FILE rm -f -s twc

  # 3. Start containers without rebuild
  Write-Host ""
  Write-Host "================================================" -ForegroundColor Blue
  Write-Host " 3. Start containers (no rebuild)" -ForegroundColor Blue
  Write-Host "================================================" -ForegroundColor Blue
  docker-compose -f $COMPOSE_FILE up -d --no-build
  if ($LASTEXITCODE -ne 0) {
    Write-ErrorMsg 'Container startup failed.'
    exit 1
  }

  # 4. Deploy done
  Write-Host ""
  Write-Host "================================================" -ForegroundColor Blue
  Write-Host " 4. Deploy done" -ForegroundColor Blue
  Write-Host "================================================" -ForegroundColor Blue
}

# 指令路由
switch ($Command) {
  'down' {
    docker-compose -f $COMPOSE_FILE down
  }
  'logs' {
    docker-compose -f $COMPOSE_FILE logs -f
  }
  default {
    Full-Deploy
  }
}
