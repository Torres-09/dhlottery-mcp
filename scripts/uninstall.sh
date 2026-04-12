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

# OS 감지 (Claude Desktop 설정 경로에 필요)
RAW_OS="$(uname -s)"
case "$RAW_OS" in
  Darwin) OS="darwin" ;;
  Linux)  OS="linux"  ;;
  *)      OS="unknown" ;;
esac

# ============================================================
# [1/3] 바이너리 제거
# ============================================================
echo -e "${BOLD}[1/3] 바이너리 제거 중...${RESET}"

BINARY_PATH="$INSTALL_DIR/$BINARY"

if [ -f "$BINARY_PATH" ]; then
  rm -f "$BINARY_PATH"
  success "바이너리 제거 완료: $BINARY_PATH"
else
  warn "바이너리를 찾을 수 없습니다: $BINARY_PATH"
  info "이미 제거되었거나 다른 경로에 설치되어 있을 수 있습니다."
fi

echo ""

# ============================================================
# [2/3] Claude Code MCP 해제
# ============================================================
echo -e "${BOLD}[2/3] Claude Code MCP 해제 중...${RESET}"

if command -v claude &>/dev/null; then
  if claude mcp remove "$BINARY" -s user 2>/dev/null; then
    success "Claude Code MCP 해제 완료"
  else
    warn "Claude Code MCP 해제에 실패했습니다. 이미 등록되지 않은 상태일 수 있습니다."
    info "수동으로 해제하려면: claude mcp remove \"$BINARY\" -s user"
  fi
else
  warn "claude CLI를 찾을 수 없습니다. 수동으로 해제해주세요:"
  info "  claude mcp remove \"$BINARY\" -s user"
fi

echo ""

# ============================================================
# [3/3] Claude Desktop 설정 제거
# ============================================================
echo -e "${BOLD}[3/3] Claude Desktop 설정 제거 중...${RESET}"

if [ "$OS" = "darwin" ]; then
  DESKTOP_CONFIG="$HOME/Library/Application Support/Claude/claude_desktop_config.json"
else
  DESKTOP_CONFIG="$HOME/.config/Claude/claude_desktop_config.json"
fi

if [ -f "$DESKTOP_CONFIG" ]; then
  BACKUP_PATH="${DESKTOP_CONFIG}.backup"
  cp "$DESKTOP_CONFIG" "$BACKUP_PATH"
  info "백업 생성: $BACKUP_PATH"

  UPDATED=false

  # jq가 있으면 jq로 편집
  if command -v jq &>/dev/null; then
    if jq -e '.mcpServers.dhlottery' "$DESKTOP_CONFIG" &>/dev/null; then
      NEW_JSON="$(jq 'del(.mcpServers.dhlottery)' "$DESKTOP_CONFIG")"
      echo "$NEW_JSON" > "$DESKTOP_CONFIG"
      UPDATED=true
    else
      warn "Claude Desktop 설정에 dhlottery 항목이 없습니다."
      UPDATED=true
      rm -f "$BACKUP_PATH"
    fi
  fi

  # jq 없으면 python3으로 편집
  if [ "$UPDATED" = false ] && command -v python3 &>/dev/null; then
    python3 - "$DESKTOP_CONFIG" <<'PYEOF'
import sys, json

config_path = sys.argv[1]

with open(config_path, 'r') as f:
    config = json.load(f)

if 'mcpServers' in config and 'dhlottery' in config['mcpServers']:
    del config['mcpServers']['dhlottery']

with open(config_path, 'w') as f:
    json.dump(config, f, indent=2, ensure_ascii=False)
    f.write('\n')
PYEOF
    UPDATED=true
  fi

  if [ "$UPDATED" = true ]; then
    success "Claude Desktop 설정에서 dhlottery 항목 제거 완료"
  else
    warn "jq 또는 python3이 없어 자동 편집에 실패했습니다. 수동으로 설정해주세요."
    info "  설정 파일: $DESKTOP_CONFIG"
    info "  mcpServers.dhlottery 항목을 직접 삭제하세요."
    rm -f "$BACKUP_PATH"
  fi
else
  warn "Claude Desktop 설정 파일을 찾을 수 없습니다."
  info "  이미 제거되었거나 Claude Desktop이 설치되지 않은 상태입니다."
fi

echo ""

# ============================================================
# PATH 정리 안내
# ============================================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}${BOLD}제거가 완료되었습니다!${RESET}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${YELLOW}PATH 정리 (선택사항):${RESET}"
echo "  ~/.local/bin 을 PATH에서 제거하려면"
echo "  ~/.zshrc 또는 ~/.bashrc 에서 관련 라인을 직접 삭제하세요."
echo ""
