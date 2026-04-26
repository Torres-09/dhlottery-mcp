# ============================================================
# dhlottery-mcp Windows 설치 스크립트
# GitHub: https://github.com/Torres-09/dhlottery-mcp
# ============================================================
#
# 사용법:
#   irm https://raw.githubusercontent.com/Torres-09/dhlottery-mcp/main/scripts/install.ps1 | iex
#
# 계정 포함:
#   & ([scriptblock]::Create((irm https://raw.githubusercontent.com/Torres-09/dhlottery-mcp/main/scripts/install.ps1))) -Id <아이디> -Pw <비밀번호>

param(
    [string]$Id = "",
    [string]$Pw = ""
)

$ErrorActionPreference = "Stop"

$OWNER   = "Torres-09"
$REPO    = "dhlottery-mcp"
$BINARY  = "dhlottery-mcp"
$INSTALL_DIR = "$env:LOCALAPPDATA\Programs\dhlottery-mcp"

function Write-Info    { param($msg) Write-Host "  $msg" }
function Write-Success { param($msg) Write-Host "  [OK] $msg" -ForegroundColor Green }
function Write-Warn    { param($msg) Write-Host "  [!] $msg"  -ForegroundColor Yellow }
function Write-Err     { param($msg) Write-Host "  [X] $msg"  -ForegroundColor Red }
function Exit-Error    { param($msg) Write-Err $msg; exit 1 }

Write-Host ""
Write-Host "dhlottery-mcp 설치 스크립트" -ForegroundColor White
Write-Host "=========================="
Write-Host ""

# ============================================================
# [1/4] 환경 감지
# ============================================================
Write-Host "[1/4] 환경 감지 중..." -ForegroundColor White

$rawArch = $env:PROCESSOR_ARCHITECTURE
switch ($rawArch) {
    "AMD64" { $ARCH = "amd64" }
    "ARM64" { $ARCH = "arm64" }
    default { Exit-Error "지원하지 않는 아키텍처입니다: $rawArch" }
}

Write-Info "OS: windows, ARCH: $ARCH"
Write-Host ""

# ============================================================
# [2/4] 최신 버전 다운로드
# ============================================================
Write-Host "[2/4] 최신 버전 다운로드 중..." -ForegroundColor White

$apiUrl = "https://api.github.com/repos/$OWNER/$REPO/releases/latest"
try {
    $release = Invoke-RestMethod -Uri $apiUrl -Headers @{ "User-Agent" = "dhlottery-mcp-installer" }
} catch {
    Exit-Error "GitHub에서 최신 버전 정보를 가져올 수 없습니다. 네트워크를 확인해주세요."
}

$TAG     = $release.tag_name
if (-not $TAG) { Exit-Error "GitHub에서 버전 정보를 찾을 수 없습니다." }

$ARCHIVE      = "${BINARY}_windows_${ARCH}.zip"
$DOWNLOAD_URL = "https://github.com/$OWNER/$REPO/releases/download/$TAG/$ARCHIVE"
$CHECKSUMS_URL = "https://github.com/$OWNER/$REPO/releases/download/$TAG/checksums.txt"

Write-Info "버전: $TAG"
Write-Info "다운로드: $ARCHIVE"

$tmpDir       = Join-Path $env:TEMP "dhlottery-mcp-install-$([System.IO.Path]::GetRandomFileName())"
New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

$archivePath   = Join-Path $tmpDir $ARCHIVE
$checksumsPath = Join-Path $tmpDir "checksums.txt"

try {
    Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $archivePath -UseBasicParsing
} catch {
    Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    Exit-Error "GitHub에서 바이너리를 다운로드할 수 없습니다. 네트워크를 확인해주세요."
}

try {
    Invoke-WebRequest -Uri $CHECKSUMS_URL -OutFile $checksumsPath -UseBasicParsing
} catch {
    Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    Exit-Error "체크섬 파일을 다운로드할 수 없습니다."
}

# SHA256 검증
$expected = (Get-Content $checksumsPath | Where-Object { $_ -match [regex]::Escape($ARCHIVE) }) -replace "\s+.*", ""
if ($expected) {
    $actual = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLower()
    $expected = $expected.Trim().ToLower()
    if ($actual -ne $expected) {
        Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        Exit-Error "체크섬 검증 실패: 바이너리가 손상되었거나 변조되었을 수 있습니다."
    }
} else {
    Write-Warn "체크섬 항목을 찾을 수 없어 검증을 건너뜁니다."
}

Expand-Archive -Path $archivePath -DestinationPath $tmpDir -Force
Write-Success "다운로드 및 무결성 검증 완료"
Write-Host ""

# ============================================================
# [3/4] 바이너리 설치
# ============================================================
Write-Host "[3/4] 바이너리 설치 중..." -ForegroundColor White

New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null

$srcExe  = Join-Path $tmpDir "${BINARY}.exe"
$destExe = Join-Path $INSTALL_DIR "${BINARY}.exe"
Copy-Item -Path $srcExe -Destination $destExe -Force

Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue

Write-Info "설치 경로: $destExe"
Write-Success "설치 완료"

# PATH 등록 (사용자 환경변수)
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notlike "*$INSTALL_DIR*") {
    [Environment]::SetEnvironmentVariable("PATH", "$INSTALL_DIR;$userPath", "User")
    $env:PATH = "$INSTALL_DIR;$env:PATH"
    Write-Success "PATH에 등록 완료: $INSTALL_DIR"
    Write-Warn "새 터미널을 열거나 로그아웃 후 재로그인하면 PATH가 적용됩니다."
} else {
    Write-Info "PATH에 이미 등록되어 있습니다."
}
Write-Host ""

# ============================================================
# [4/4] MCP 클라이언트 등록
# ============================================================
Write-Host "[4/4] MCP 클라이언트 등록 중..." -ForegroundColor White

# -Id/-Pw 파라미터가 없으면 대화형으로 입력받음
if (-not $Id) {
    Write-Host ""
    Write-Host "동행복권 계정 정보를 입력하면 구매·잔액 조회 기능을 바로 사용할 수 있습니다."
    Write-Host "(건너뛰려면 Enter를 누르세요)"
    Write-Host ""
    $Id = Read-Host "  아이디"
    $securePw = Read-Host "  비밀번호" -AsSecureString
    $Pw = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
        [Runtime.InteropServices.Marshal]::SecureStringToBSTR($securePw))
    Write-Host ""
}

$REGISTERED_ANY = $false

# Claude Code 등록
$claudeExe = Get-Command "claude" -ErrorAction SilentlyContinue
if ($claudeExe) {
    # 기존 항목 제거 후 재등록 (업데이트 보장)
    & claude mcp remove $BINARY -s user 2>$null
    try {
        & claude mcp add --scope user -t stdio $BINARY -- $destExe 2>$null

        # 환경변수 주입: PowerShell JSON으로 .claude.json 직접 편집
        if ($Id -or $Pw) {
            $claudeJson = "$env:USERPROFILE\.claude.json"
            if (Test-Path $claudeJson) {
                try {
                    $cdata = Get-Content $claudeJson -Raw -Encoding UTF8 | ConvertFrom-Json
                    if ($cdata.mcpServers -and $cdata.mcpServers.PSObject.Properties[$BINARY]) {
                        $srvEntry = $cdata.mcpServers.$BINARY
                        if (-not $srvEntry.PSObject.Properties["env"]) {
                            $srvEntry | Add-Member -MemberType NoteProperty -Name "env" -Value ([PSCustomObject]@{})
                        }
                        $srvEntry.env | Add-Member -MemberType NoteProperty -Name "DHLOTTERY_USER_ID" -Value $Id -Force
                        $srvEntry.env | Add-Member -MemberType NoteProperty -Name "DHLOTTERY_USER_PW" -Value $Pw -Force
                        $cdata | ConvertTo-Json -Depth 10 | Set-Content $claudeJson -Encoding UTF8
                    }
                } catch {
                    Write-Warn "Claude Code 계정 정보 주입에 실패했습니다. 수동으로 설정해주세요."
                }
            }
        }

        Write-Success "Claude Code에 등록 완료"
        $REGISTERED_ANY = $true
    } catch {
        Write-Warn "Claude Code 등록에 실패했습니다. 수동으로 등록해주세요:"
        Write-Info "  claude mcp add --scope user -t stdio $BINARY -- $destExe"
    }
} else {
    Write-Warn "claude CLI를 찾을 수 없습니다. Claude Code 등록을 건너뜁니다."
}

# Claude Desktop 설정 파일
$DESKTOP_CONFIG = "$env:APPDATA\Claude\claude_desktop_config.json"

if (Test-Path $DESKTOP_CONFIG) {
    $backupPath = "${DESKTOP_CONFIG}.backup"
    Copy-Item $DESKTOP_CONFIG $backupPath -Force

    try {
        $config = Get-Content $DESKTOP_CONFIG -Raw -Encoding UTF8 | ConvertFrom-Json

        if (-not $config.mcpServers) {
            $config | Add-Member -MemberType NoteProperty -Name "mcpServers" -Value ([PSCustomObject]@{})
        }

        $serverEntry = [PSCustomObject]@{
            command = $destExe
            env     = [PSCustomObject]@{
                DHLOTTERY_USER_ID = $Id
                DHLOTTERY_USER_PW = $Pw
            }
        }

        if ($config.mcpServers.PSObject.Properties["dhlottery"]) {
            if ($Id -or $Pw) {
                $config.mcpServers.dhlottery.env.DHLOTTERY_USER_ID = $Id
                $config.mcpServers.dhlottery.env.DHLOTTERY_USER_PW = $Pw
            } else {
                Write-Warn "Claude Desktop에 이미 dhlottery-mcp가 등록되어 있습니다."
            }
        } else {
            $config.mcpServers | Add-Member -MemberType NoteProperty -Name "dhlottery" -Value $serverEntry
        }

        $config | ConvertTo-Json -Depth 10 | Set-Content $DESKTOP_CONFIG -Encoding UTF8
        Write-Success "Claude Desktop에 등록 완료"

        if (-not $Id) {
            Write-Info "  → 계정 정보 설정: $DESKTOP_CONFIG"
            Write-Info "     DHLOTTERY_USER_ID와 DHLOTTERY_USER_PW에 동행복권 계정 정보를 입력하세요."
        } else {
            Write-Info "  → 계정 정보가 설정되었습니다."
        }
        $REGISTERED_ANY = $true
    } catch {
        Write-Warn "Claude Desktop JSON 편집에 실패했습니다. 수동으로 설정해주세요."
        Write-Info "  설정 파일: $DESKTOP_CONFIG"
        Remove-Item $backupPath -Force -ErrorAction SilentlyContinue
    }
} else {
    Write-Warn "Claude Desktop 설정 파일을 찾을 수 없습니다. 수동으로 설정해주세요."
    Write-Info "  설정 방법은 README를 참고하세요: https://github.com/$OWNER/$REPO"
}

Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host "설치가 완료되었습니다!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host ""
Write-Host "당첨번호 조회 등 로그인 불필요 기능은 바로 사용 가능합니다."
if (-not $Id) {
    Write-Host "구매 기능을 사용하려면 동행복권 계정 정보를 설정하세요."
    if (Test-Path $DESKTOP_CONFIG) {
        Write-Host ""
        Write-Host "  → $DESKTOP_CONFIG" -ForegroundColor Yellow
    }
}
Write-Host ""
