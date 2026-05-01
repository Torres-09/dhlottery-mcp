#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# dhlottery-mcp 제거 스크립트
# GitHub: https://github.com/Torres-09/dhlottery-mcp
# ============================================================

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

echo ""
echo -e "${BOLD}dhlottery-mcp 제거 스크립트${RESET}"
echo "=========================="
echo ""

BINARY_PATH="$INSTALL_DIR/$BINARY"

# ============================================================
# [1/2] MCP 클라이언트 설정 제거
# ============================================================
echo -e "${BOLD}[1/2] MCP 클라이언트 설정 제거 중...${RESET}"
echo ""

# 바이너리가 PATH에 있으면 unsetup으로 일괄 처리
if command -v "$BINARY" &>/dev/null; then
  "$BINARY" unsetup || true
else
  warn "dhlottery-mcp 바이너리를 찾을 수 없습니다. 설정 정리를 건너뜁니다."
  info "수동으로 제거하려면:"
  info "  claude mcp remove \"$BINARY\" -s user"
fi

echo ""

# ============================================================
# [2/2] 바이너리 제거
# ============================================================
echo -e "${BOLD}[2/2] 바이너리 제거 중...${RESET}"

if [ -f "$BINARY_PATH" ]; then
  rm -f "$BINARY_PATH"
  success "바이너리 제거 완료: $BINARY_PATH"
else
  warn "바이너리를 찾을 수 없습니다: $BINARY_PATH"
  info "이미 제거되었거나 다른 경로에 설치되어 있을 수 있습니다."
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}${BOLD}제거가 완료되었습니다!${RESET}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${YELLOW}PATH 정리 (선택사항):${RESET}"
echo "  ~/.local/bin 을 PATH에서 제거하려면"
echo "  ~/.zshrc 또는 ~/.bashrc 에서 관련 라인을 직접 삭제하세요."
echo ""
