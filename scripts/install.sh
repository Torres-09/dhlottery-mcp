#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# dhlottery-mcp 설치 스크립트
# GitHub: https://github.com/Torres-09/dhlottery-mcp
# ============================================================

OWNER="Torres-09"
REPO="dhlottery-mcp"
BINARY="dhlottery-mcp"
INSTALL_DIR="${DHLOTTERY_INSTALL_DIR:-$HOME/.local/bin}"

# 색상 출력 (터미널이 아니면 비활성화)
if [ -t 1 ]; then
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[1;33m'
  BOLD='\033[1m'
  RESET='\033[0m'
else
  RED='' GREEN='' YELLOW='' BOLD='' RESET=''
fi

info()    { echo -e "  $*"; }
success() { echo -e "  ${GREEN}✓${RESET} $*"; }
warn()    { echo -e "  ${YELLOW}!${RESET} $*"; }
error()   { echo -e "  ${RED}✗${RESET} $*" >&2; }
die()     { error "$*"; exit 1; }

# 임시 디렉토리 생성 및 종료 시 자동 정리
TMPDIR_WORK="$(mktemp -d)"
trap 'rm -rf "$TMPDIR_WORK"' EXIT

# HTTP 다운로드 함수
download() {
  local url="$1" dest="$2"
  if command -v curl &>/dev/null; then
    curl -sSL --fail "$url" -o "$dest" || return 1
  elif command -v wget &>/dev/null; then
    wget -qO "$dest" "$url" || return 1
  else
    die "curl 또는 wget이 필요합니다."
  fi
}

# HTTP 응답 텍스트 출력 함수
fetch() {
  local url="$1"
  if command -v curl &>/dev/null; then
    curl -sSL --fail "$url" || return 1
  elif command -v wget &>/dev/null; then
    wget -qO- "$url" || return 1
  else
    die "curl 또는 wget이 필요합니다."
  fi
}

echo ""
echo -e "${BOLD}dhlottery-mcp 설치 스크립트${RESET}"
echo "=========================="
echo ""

# ============================================================
# [1/4] 환경 감지
# ============================================================
echo -e "${BOLD}[1/4] 환경 감지 중...${RESET}"

RAW_OS="$(uname -s)"
RAW_ARCH="$(uname -m)"

case "$RAW_OS" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)      die "지원하지 않는 OS입니다: $RAW_OS" ;;
esac

case "$RAW_ARCH" in
  x86_64)          ARCH="amd64" ;;
  aarch64|arm64)   ARCH="arm64" ;;
  *)               die "지원하지 않는 아키텍처입니다: $RAW_ARCH" ;;
esac

info "OS: $OS, ARCH: $ARCH"
echo ""

# ============================================================
# [2/4] 최신 버전 다운로드
# ============================================================
echo -e "${BOLD}[2/4] 최신 버전 다운로드 중...${RESET}"

API_URL="https://api.github.com/repos/$OWNER/$REPO/releases/latest"
TAG="$(fetch "$API_URL" 2>/dev/null \
  | grep '"tag_name"' \
  | head -1 \
  | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')" \
  || die "GitHub에서 최신 버전 정보를 가져올 수 없습니다. 네트워크를 확인해주세요."

[ -z "$TAG" ] && die "GitHub에서 버전 정보를 찾을 수 없습니다."

VERSION="${TAG#v}"
ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$OWNER/$REPO/releases/download/$TAG/$ARCHIVE"

info "버전: $TAG"
info "다운로드: $ARCHIVE"

ARCHIVE_PATH="$TMPDIR_WORK/$ARCHIVE"
download "$DOWNLOAD_URL" "$ARCHIVE_PATH" \
  || die "GitHub에서 바이너리를 다운로드할 수 없습니다. 네트워크를 확인해주세요."

tar -xzf "$ARCHIVE_PATH" -C "$TMPDIR_WORK"
success "다운로드 완료"
echo ""

# ============================================================
# [3/4] 바이너리 설치
# ============================================================
echo -e "${BOLD}[3/4] 바이너리 설치 중...${RESET}"

mkdir -p "$INSTALL_DIR"
install -m 755 "$TMPDIR_WORK/$BINARY" "$INSTALL_DIR/$BINARY"
BINARY_PATH="$INSTALL_DIR/$BINARY"

info "설치 경로: $BINARY_PATH"
success "설치 완료"

# PATH 안내
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    warn "$INSTALL_DIR 이 PATH에 포함되어 있지 않습니다."
    info "다음 명령어로 추가할 수 있습니다:"
    if [ "$OS" = "darwin" ] && [ -f "$HOME/.zshrc" ]; then
      info "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
    else
      info "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc && source ~/.bashrc"
    fi
    ;;
esac
echo ""

# ============================================================
# [4/4] MCP 클라이언트 등록
# ============================================================
echo -e "${BOLD}[4/4] MCP 클라이언트 등록 중...${RESET}"

REGISTERED_ANY=false

# Claude Code 등록
if command -v claude &>/dev/null; then
  if claude mcp add --scope user -t stdio "$BINARY" -- "$BINARY_PATH" 2>/dev/null; then
    success "Claude Code에 등록 완료"
    REGISTERED_ANY=true
  else
    warn "Claude Code 등록에 실패했습니다. 수동으로 등록해주세요:"
    info "  claude mcp add --scope user -t stdio $BINARY -- $BINARY_PATH"
  fi
else
  warn "claude CLI를 찾을 수 없습니다. Claude Code 등록을 건너뜁니다."
fi

# Claude Desktop 설정 파일 위치
if [ "$OS" = "darwin" ]; then
  DESKTOP_CONFIG="$HOME/Library/Application Support/Claude/claude_desktop_config.json"
else
  DESKTOP_CONFIG="$HOME/.config/Claude/claude_desktop_config.json"
fi

# Claude Desktop 등록
if [ -f "$DESKTOP_CONFIG" ]; then
  BACKUP_PATH="${DESKTOP_CONFIG}.backup"
  cp "$DESKTOP_CONFIG" "$BACKUP_PATH"

  UPDATED=false

  # jq가 있으면 jq로 편집
  if command -v jq &>/dev/null; then
    # 이미 등록되어 있으면 스킵
    if jq -e '.mcpServers.dhlottery' "$DESKTOP_CONFIG" &>/dev/null; then
      warn "Claude Desktop에 이미 dhlottery-mcp가 등록되어 있습니다."
      UPDATED=true
    else
      NEW_JSON="$(jq \
        --arg path "$BINARY_PATH" \
        '.mcpServers.dhlottery = {"command": $path, "env": {"DHLOTTERY_USER_ID": "", "DHLOTTERY_USER_PW": ""}}' \
        "$DESKTOP_CONFIG")" \
        && echo "$NEW_JSON" > "$DESKTOP_CONFIG" \
        && UPDATED=true \
        || true
    fi
  fi

  # jq 없으면 python3으로 편집
  if [ "$UPDATED" = false ] && command -v python3 &>/dev/null; then
    python3 - "$DESKTOP_CONFIG" "$BINARY_PATH" <<'PYEOF'
import sys, json

config_path = sys.argv[1]
binary_path = sys.argv[2]

with open(config_path, 'r') as f:
    config = json.load(f)

if 'mcpServers' not in config:
    config['mcpServers'] = {}

if 'dhlottery' not in config['mcpServers']:
    config['mcpServers']['dhlottery'] = {
        'command': binary_path,
        'env': {
            'DHLOTTERY_USER_ID': '',
            'DHLOTTERY_USER_PW': ''
        }
    }

with open(config_path, 'w') as f:
    json.dump(config, f, indent=2, ensure_ascii=False)
    f.write('\n')
PYEOF
    UPDATED=true || true
  fi

  if [ "$UPDATED" = true ]; then
    success "Claude Desktop에 등록 완료"
    info "  → 계정 정보 설정: $DESKTOP_CONFIG"
    info "     DHLOTTERY_USER_ID와 DHLOTTERY_USER_PW에 동행복권 계정 정보를 입력하세요."
    REGISTERED_ANY=true
  else
    warn "Claude Desktop JSON 편집에 실패했습니다. 수동으로 설정해주세요."
    info "  설정 파일: $DESKTOP_CONFIG"
    rm -f "$BACKUP_PATH"
  fi
else
  warn "Claude Desktop 설정 파일을 찾을 수 없습니다. 수동으로 설정해주세요."
  info "  설정 방법은 README를 참고하세요: https://github.com/$OWNER/$REPO"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}${BOLD}설치가 완료되었습니다!${RESET}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "당첨번호 조회 등 로그인 불필요 기능은 바로 사용 가능합니다."
echo "구매 기능을 사용하려면 동행복권 계정 정보를 설정하세요."
if [ -f "$DESKTOP_CONFIG" ]; then
  echo ""
  echo -e "  ${YELLOW}→${RESET} $DESKTOP_CONFIG"
fi
echo ""
