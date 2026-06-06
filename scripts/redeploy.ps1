#!/usr/bin/env pwsh

<#
.SYNOPSIS
  TeWeiCun (特维存) Docker 重部署脚本（本地 / 远端）
.DESCRIPTION
  本地：不触动数据库镜像 -> 构建业务镜像 -> 精准重启业务容器
  远端：通过 SSH 在服务器上拉取镜像并重启业务容器（可选执行 git pull）
.EXAMPLE
  .\scripts\redeploy.ps1
  .\scripts\redeploy.ps1 down
  .\scripts\redeploy.ps1 logs
  .\scripts\redeploy.ps1 remote -Host 1.2.3.4 -User root -RemotePath /root/twc -RemoteComposeFile docker-compose.yml -GitPull -Pull
#>

param(
  [Parameter(Position=0)]
  [ValidateSet('', 'down', 'logs', 'remote')]
  [string]$Command = '',

  [string]$ComposeFile = 'docker-compose-local.yml',
  [string]$Service = 'twc',

  [switch]$Help,

  [Alias('Host')]
  [string]$RemoteHost = '',
  [string]$User = 'root',
  [int]$Port = 22,
  [string]$RemotePath = '/root/twc',
  [string]$RemoteComposeFile = 'docker-compose.yml',
  [switch]$Pull,
  [switch]$GitPull
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# 设置控制台编码为UTF-8
$OutputEncoding = [System.Text.Encoding]::UTF8
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# 项目根目录
$ProjectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $ProjectRoot

function Write-Info([string]$msg)    { Write-Host "➤ $msg" -ForegroundColor Blue }
function Write-Success([string]$msg) { Write-Host "✔ $msg" -ForegroundColor Green }
function Write-Warn([string]$msg)    { Write-Host "⚠ $msg" -ForegroundColor Yellow }
function Write-ErrorMsg([string]$msg){ Write-Host "✖ $msg" -ForegroundColor Red }

function Show-Usage {
  @"
用法:
  .\scripts\redeploy.ps1 [command] [options]

命令:
  (空)    本地：build + 精准重启容器（默认）
  down    本地：停止并移除容器
  logs    本地：实时日志
  remote  远端：SSH 到服务器执行 pull/up（可选 git pull）

本地参数:
  -ComposeFile <string>   默认: docker-compose-local.yml
  -Service <string>       默认: twc

远端参数:
  -Host <string>          必填：服务器地址（等价于 -RemoteHost）
  -RemoteHost <string>    必填：服务器地址
  -User <string>          默认: root
  -Port <int>             默认: 22
  -RemotePath <string>    默认: /root/twc
  -RemoteComposeFile <string> 默认: docker-compose.yml
  -GitPull                远端先执行 git pull --rebase
  -Pull                   远端先执行 docker compose pull <service>

示例:
  .\scripts\redeploy.ps1
  .\scripts\redeploy.ps1 logs
  .\scripts\redeploy.ps1 remote -Host 1.2.3.4 -User root -RemotePath /root/twc -RemoteComposeFile docker-compose.yml -GitPull -Pull
"@ | Write-Host
}

function Get-ComposeRunner {
  function Add-ToPathIfExists([string]$dir) {
    if ([string]::IsNullOrWhiteSpace($dir)) { return }
    if (-not (Test-Path -LiteralPath $dir)) { return }
    $parts = @($env:PATH -split ';' | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
    if ($parts -contains $dir) { return }
    $env:PATH = ($dir + ';' + $env:PATH)
  }

  function Ensure-DockerOnPath {
    if (Get-Command docker -ErrorAction SilentlyContinue) { return }

    $dirs = @()
    if ($env:ProgramFiles) {
      $dirs += (Join-Path $env:ProgramFiles 'Docker\Docker\resources\bin')
      $dirs += (Join-Path $env:ProgramFiles 'Docker\Docker\resources')
    }
    if ($env:ProgramFiles -and ${env:ProgramFiles(x86)}) {
      $dirs += (Join-Path ${env:ProgramFiles(x86)} 'Docker\Docker\resources\bin')
      $dirs += (Join-Path ${env:ProgramFiles(x86)} 'Docker\Docker\resources')
    }

    foreach ($d in $dirs | Select-Object -Unique) {
      if (Test-Path -LiteralPath (Join-Path $d 'docker.exe')) {
        Add-ToPathIfExists $d
        break
      }
    }
  }

  Ensure-DockerOnPath

  $docker = Get-Command docker -ErrorAction SilentlyContinue
  if ($docker) {
    try {
      & docker compose version | Out-Null
      if ($LASTEXITCODE -eq 0) {
        return @{ Cmd = 'docker'; Prefix = @('compose') }
      }
    } catch {
    }
  }

  $dockerCompose = Get-Command docker-compose -ErrorAction SilentlyContinue
  if ($dockerCompose) {
    return @{ Cmd = 'docker-compose'; Prefix = @() }
  }

  throw @"
未找到 docker compose / docker-compose。
常见原因：
1) Docker Desktop 已安装但当前终端未刷新 PATH（关闭并重新打开终端/IDE 即可）
2) Docker Desktop 未启动（先启动 Docker Desktop，确保右下角 Docker 图标为 Running）
3) docker.exe 不在 PATH（可临时把 C:\Program Files\Docker\Docker\resources\bin 加到 PATH）
"@
}

function Invoke-Compose {
  param(
    [Parameter(Mandatory=$true)]
    [string]$File,
    [Parameter(Mandatory=$true)]
    [string[]]$Args
  )
  $runner = Get-ComposeRunner
  $fullArgs = @()
  $fullArgs += $runner.Prefix
  $fullArgs += @('-f', $File)
  $fullArgs += $Args
  & $runner.Cmd @fullArgs
  if ($LASTEXITCODE -ne 0) {
    throw "Compose 执行失败：$($runner.Cmd) $($fullArgs -join ' ')"
  }
}

function Full-DeployLocal {
  Write-Host ""
  Write-Host "================================================" -ForegroundColor Blue
  Write-Host " 1. 构建业务镜像（单镜像）" -ForegroundColor Blue
  Write-Host "================================================" -ForegroundColor Blue
  Invoke-Compose -File $ComposeFile -Args @('build', $Service)

  Write-Host ""
  Write-Host "================================================" -ForegroundColor Blue
  Write-Host " 2. 仅停止并删除业务容器" -ForegroundColor Blue
  Write-Host "================================================" -ForegroundColor Blue
  try {
    Invoke-Compose -File $ComposeFile -Args @('rm', '-f', '-s', $Service)
  } catch {
    Write-Warn "删除容器失败（可能是首次启动或容器不存在）：$($_.Exception.Message)"
  }

  Write-Host ""
  Write-Host "================================================" -ForegroundColor Blue
  Write-Host " 3. 启动所有容器（不重复构建）" -ForegroundColor Blue
  Write-Host "================================================" -ForegroundColor Blue
  Invoke-Compose -File $ComposeFile -Args @('up', '-d', '--no-build')

  Write-Host ""
  Write-Host "================================================" -ForegroundColor Blue
  Write-Host " 4. 部署完成" -ForegroundColor Blue
  Write-Host "================================================" -ForegroundColor Blue
  Write-Host ""
  Write-Success "环境就绪： http://localhost:8080"
}

function Invoke-RemoteRedeploy {
  if ([string]::IsNullOrWhiteSpace($RemoteHost)) {
    throw 'remote 模式必须提供 -Host / -RemoteHost'
  }

  $ssh = Get-Command ssh -ErrorAction SilentlyContinue
  if (-not $ssh) {
    throw '未找到 ssh 命令。请先安装/启用 Windows OpenSSH Client。'
  }

  $gitPullPart = ''
  if ($GitPull) {
    $gitPullPart = 'git pull --rebase;'
  }

  $remoteScript = @"
set -e
cd "$RemotePath"
$gitPullPart
if docker compose version >/dev/null 2>&1; then
  DC="docker compose -f $RemoteComposeFile"
else
  DC="docker-compose -f $RemoteComposeFile"
fi
if [ "$($Pull.IsPresent)" = "True" ]; then
  \$DC pull "$Service"
fi
\$DC up -d "$Service"
\$DC ps
"@

  $target = "$User@$RemoteHost"
  $sshArgs = @('-p', "$Port", $target, 'bash', '-lc', $remoteScript)
  & ssh @sshArgs
  if ($LASTEXITCODE -ne 0) {
    throw "远端重部署失败：ssh $target"
  }
}

if ($Help) { Show-Usage; exit 0 }

switch ($Command) {
  'down' {
    Invoke-Compose -File $ComposeFile -Args @('down')
  }
  'logs' {
    Invoke-Compose -File $ComposeFile -Args @('logs', '-f')
  }
  'remote' {
    Invoke-RemoteRedeploy
  }
  default {
    Full-DeployLocal
  }
}
