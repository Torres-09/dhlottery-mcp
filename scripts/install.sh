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

# ============================================================
# 인자 파싱
# ============================================================
USER_ID=""
USER_PW=""

# 셸 환경에 따라 따옴표가 제거되지 않고 전달되는 경우 방어 처리
strip_quotes() {
  local v="$1"
  v="${v#\'}" ; v="${v%\'}"   # 앞뒤 ASCII 작은따옴표 제거
  v="${v#\"}" ; v="${v%\"}"   # 앞뒤 큰따옴표 제거
  # 유니코드 곡선 따옴표 제거 (복사·붙여넣기 시 자동 변환 대응)
  v="${v#$'\xe2\x80\x98'}" ; v="${v%$'\xe2\x80\x99'}"  # U+2018/U+2019
  v="${v#$'\xe2\x80\x9c'}" ; v="${v%$'\xe2\x80\x9d'}"  # U+201C/U+201D
  printf '%s' "$v"
}

while [ $# -gt 0 ]; do
  case "$1" in
    --id)   USER_ID="$(strip_quotes "$2")"; shift 2 ;;
    --pw)   USER_PW="$(strip_quotes "$2")"; shift 2 ;;
    --help|-h)
      echo "Usage: install.sh [--id <동행복권_아이디>] [--pw <동행복권_비밀번호>]"
      echo ""
      echo "  --id  동행복권 아이디 (생략 시 나중에 직접 설정)"
      echo "  --pw  동행복권 비밀번호 (생략 시 나중에 직접 설정)"
      exit 0
      ;;
    *) die "알 수 없는 옵션: $1" ;;
  esac
done

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
CHECKSUMS_PATH="$TMPDIR_WORK/checksums.txt"
CHECKSUMS_URL="https://github.com/$OWNER/$REPO/releases/download/$TAG/checksums.txt"

download "$DOWNLOAD_URL" "$ARCHIVE_PATH" \
  || die "GitHub에서 바이너리를 다운로드할 수 없습니다. 네트워크를 확인해주세요."

download "$CHECKSUMS_URL" "$CHECKSUMS_PATH" \
  || die "체크섬 파일을 다운로드할 수 없습니다."

# 체크섬 검증
if command -v sha256sum &>/dev/null; then
  (cd "$TMPDIR_WORK" && grep "$ARCHIVE" checksums.txt | sha256sum -c --status) \
    || die "체크섬 검증 실패: 바이너리가 손상되었거나 변조되었을 수 있습니다."
elif command -v shasum &>/dev/null; then
  (cd "$TMPDIR_WORK" && grep "$ARCHIVE" checksums.txt | shasum -a 256 -c --status) \
    || die "체크섬 검증 실패: 바이너리가 손상되었거나 변조되었을 수 있습니다."
else
  warn "sha256sum/shasum을 찾을 수 없어 체크섬 검증을 건너뜁니다."
fi

tar -xzf "$ARCHIVE_PATH" -C "$TMPDIR_WORK"
success "다운로드 및 무결성 검증 완료"
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
echo ""

# dhlottery-mcp setup 호출 — Go 레이어에서 JSON/TOML 직접 수정
# $, !, ~ 등 특수문자 쉘 확장 문제 없음
SETUP_ARGS=()
[ -n "$USER_ID" ] && SETUP_ARGS+=("--id" "$USER_ID")
[ -n "$USER_PW" ] && SETUP_ARGS+=("--pw" "$USER_PW")

"$BINARY_PATH" setup ${SETUP_ARGS[@]+"${SETUP_ARGS[@]}"}

echo ""
