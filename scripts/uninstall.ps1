# ============================================================
# dhlottery-mcp Windows 제거 스크립트
# GitHub: https://github.com/Torres-09/dhlottery-mcp
# ============================================================
#
# 사용법:
#   irm https://raw.githubusercontent.com/Torres-09/dhlottery-mcp/main/scripts/uninstall.ps1 | iex

$ErrorActionPreference = "Stop"

$BINARY      = "dhlottery-mcp"
$INSTALL_DIR = "$env:LOCALAPPDATA\Programs\dhlottery-mcp"

function Write-Info    { param($msg) Write-Host "  $msg" }
function Write-Success { param($msg) Write-Host "  [OK] $msg" -ForegroundColor Green }
function Write-Warn    { param($msg) Write-Host "  [!] $msg"  -ForegroundColor Yellow }
function Write-Err     { param($msg) Write-Host "  [X] $msg"  -ForegroundColor Red }

Write-Host ""
Write-Host "dhlottery-mcp 제거 스크립트" -ForegroundColor White
Write-Host "=========================="
Write-Host ""

# ============================================================
# [1/3] 바이너리 및 디렉토리 제거
# ============================================================
Write-Host "[1/3] 바이너리 제거 중..." -ForegroundColor White

if (Test-Path $INSTALL_DIR) {
    Remove-Item -Path $INSTALL_DIR -Recurse -Force
    Write-Success "바이너리 제거 완료: $INSTALL_DIR"
} else {
    Write-Warn "설치 디렉토리를 찾을 수 없습니다: $INSTALL_DIR"
    Write-Info "이미 제거되었거나 다른 경로에 설치되어 있을 수 있습니다."
}

Write-Host ""

# ============================================================
# [2/3] PATH에서 제거
# ============================================================
Write-Host "[2/3] PATH에서 제거 중..." -ForegroundColor White

$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -like "*$INSTALL_DIR*") {
    $newPath = ($userPath -split ";" | Where-Object { $_ -ne $INSTALL_DIR }) -join ";"
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    $env:PATH = ($env:PATH -split ";" | Where-Object { $_ -ne $INSTALL_DIR }) -join ";"
    Write-Success "PATH에서 제거 완료: $INSTALL_DIR"
} else {
    Write-Info "PATH에 등록되어 있지 않습니다."
}

Write-Host ""

# ============================================================
# [3/3] Claude Code MCP 해제 + Claude Desktop 설정 제거
# ============================================================
Write-Host "[3/3] MCP 클라이언트 설정 제거 중..." -ForegroundColor White

# Claude Code MCP 해제
$claudeExe = Get-Command "claude" -ErrorAction SilentlyContinue
if ($claudeExe) {
    try {
        & claude mcp remove $BINARY -s user 2>$null
        Write-Success "Claude Code MCP 해제 완료"
    } catch {
        Write-Warn "Claude Code MCP 해제에 실패했습니다. 이미 등록되지 않은 상태일 수 있습니다."
        Write-Info "  수동으로 해제하려면: claude mcp remove `"$BINARY`" -s user"
    }
} else {
    Write-Warn "claude CLI를 찾을 수 없습니다. 수동으로 해제해주세요:"
    Write-Info "  claude mcp remove `"$BINARY`" -s user"
}

# Claude Desktop 설정 파일
$DESKTOP_CONFIG = "$env:APPDATA\Claude\claude_desktop_config.json"

if (Test-Path $DESKTOP_CONFIG) {
    $backupPath = "${DESKTOP_CONFIG}.backup"
    Copy-Item $DESKTOP_CONFIG $backupPath -Force
    Write-Info "백업 생성: $backupPath"

    try {
        $config = Get-Content $DESKTOP_CONFIG -Raw -Encoding UTF8 | ConvertFrom-Json

        if ($config.mcpServers -and $config.mcpServers.PSObject.Properties["dhlottery"]) {
            $config.mcpServers.PSObject.Properties.Remove("dhlottery")
            $config | ConvertTo-Json -Depth 10 | Set-Content $DESKTOP_CONFIG -Encoding UTF8
            Write-Success "Claude Desktop 설정에서 dhlottery 항목 제거 완료"
        } else {
            Write-Info "Claude Desktop 설정에 dhlottery 항목이 없습니다."
            Remove-Item $backupPath -Force -ErrorAction SilentlyContinue
        }
    } catch {
        Write-Warn "Claude Desktop JSON 편집에 실패했습니다. 수동으로 설정해주세요."
        Write-Info "  설정 파일: $DESKTOP_CONFIG"
        Write-Info "  mcpServers.dhlottery 항목을 직접 삭제하세요."
        Remove-Item $backupPath -Force -ErrorAction SilentlyContinue
    }
} else {
    Write-Warn "Claude Desktop 설정 파일을 찾을 수 없습니다."
    Write-Info "  이미 제거되었거나 Claude Desktop이 설치되지 않은 상태입니다."
}

Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host "제거가 완료되었습니다!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host ""
