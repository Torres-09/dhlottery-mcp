# dhlottery-mcp

[![Go Version](https://img.shields.io/github/go-mod/go-version/Torres-09/dhlottery-mcp?branch=main)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/Torres-09/dhlottery-mcp)](https://github.com/Torres-09/dhlottery-mcp/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

동행복권 로또 6/45 MCP(Model Context Protocol) 서버 — Go 단일 바이너리, Playwright 불필요

Claude에서 자연어로 로또 당첨번호를 조회하고, 예치금을 확인하고, 복권을 구매할 수 있습니다.

## 주요 기능

| Tool | 설명 | 로그인 필요 |
|------|------|:----------:|
| `get_winning_numbers` | 특정 회차 당첨번호 조회 | ❌ |
| `get_latest_round` | 최신/다음 회차 정보 조회 | ❌ |
| `check_numbers` | 내 번호와 당첨번호 비교 | ❌ |
| `get_balance` | 예치금 잔액 조회 | ✅ |
| `buy_lotto` | 로또 6/45 구매 (자동/수동/반자동) | ✅ |
| `get_purchase_history` | 인터넷 구매 내역 조회 | ✅ |

> `get_winning_numbers`, `get_latest_round`, `check_numbers`는 계정 설정 없이 바로 사용 가능합니다.

## 설치

### 🌐 웹 인스톨러 (비개발자 추천)

터미널 없이 브라우저만으로 설치할 수 있습니다.

👉 **[https://torres-09.github.io/dhlottery-mcp](https://torres-09.github.io/dhlottery-mcp)**

### 원클릭 설치

바이너리 설치 후 Claude Code와 Claude Desktop에 자동 등록까지 진행합니다. 설치 중 계정 정보를 대화형으로 입력할 수 있습니다.

**macOS / Linux**

```bash
curl -fsSL https://raw.githubusercontent.com/Torres-09/dhlottery-mcp/main/scripts/install.sh | bash
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/Torres-09/dhlottery-mcp/main/scripts/install.ps1 | iex
```

### Homebrew (macOS / Linux)

```bash
brew install Torres-09/tap/dhlottery-mcp
```

### go install

```bash
go install github.com/Torres-09/dhlottery-mcp/cmd/dhlottery-mcp@latest
claude mcp add --scope user -t stdio dhlottery-mcp -- dhlottery-mcp
```

## 제거

**macOS / Linux**

```bash
curl -fsSL https://raw.githubusercontent.com/Torres-09/dhlottery-mcp/main/scripts/uninstall.sh | bash
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/Torres-09/dhlottery-mcp/main/scripts/uninstall.ps1 | iex
```

**Homebrew**

```bash
brew uninstall Torres-09/tap/dhlottery-mcp
claude mcp remove "dhlottery-mcp" -s user
```

## 사용 예시

Claude에게 자연어로 요청하세요:

```
1216회 당첨번호 알려줘
이번 주 로또 몇 회차야?
내 번호 [3, 13, 20, 29, 37, 42]가 1216회에 몇 등인지 확인해줘
로또 자동으로 1게임 사줘
내 예치금 잔액 얼마야?
최근 구매 내역 보여줘
```

## Claude Desktop 설정

설정 파일 위치:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

아래 내용을 추가하세요:

```json
{
  "mcpServers": {
    "dhlottery": {
      "command": "/path/to/dhlottery-mcp",
      "env": {
        "DHLOTTERY_USER_ID": "동행복권_아이디",
        "DHLOTTERY_USER_PW": "동행복권_비밀번호"
      }
    }
  }
}
```

## Claude Code 설정

```bash
claude mcp add --scope user -t stdio dhlottery-mcp -- /path/to/dhlottery-mcp
```

## 환경변수

| 변수 | 설명 | 필수 |
|------|------|:----:|
| `DHLOTTERY_USER_ID` | 동행복권 아이디 | 로그인 필요 기능만 |
| `DHLOTTERY_USER_PW` | 동행복권 비밀번호 | 로그인 필요 기능만 |

로그인이 필요 없는 도구(`get_winning_numbers`, `get_latest_round`, `check_numbers`)는 환경변수 없이도 바로 사용 가능합니다.

## 구매 관련 안내

| 항목 | 내용                                          |
|------|---------------------------------------------|
| 판매 시간 | 일\~금 06:00\~24:00 / 토(추첨일) 06:00\~20:00 |
| 주당 구매 한도 | **최대 5게임** (모바일·인터넷 합산)                     |
| 구매 방식 | `auto`(자동번호), `manual`(수동번호 6개), `semi`(반자동) |
| 구매 내역 조회 | 인터넷(PC/웹) 구매만 조회됨 (모바일 앱 구매 제외)             |

## 법적 고지

- 이 도구는 동행복권의 공식 API를 사용하지 않으며, 비공식 HTTP 요청 기반으로 동작합니다.
- 사용에 따른 모든 책임은 사용자에게 있습니다.
- 과도한 요청으로 인한 계정 제재 등의 불이익에 대해 책임지지 않습니다.

## 라이선스

MIT
