# dhlottery-mcp

동행복권 로또 6/45 MCP(Model Context Protocol) 서버 — Go 구현

Claude Desktop에서 자연어로 로또 당첨번호를 조회하고, 예치금을 확인하고, 복권을 구매할 수 있습니다.

## 기능

| Tool | 설명 | 로그인 필요 |
|------|------|:----------:|
| `get_winning_numbers` | 특정 회차 당첨번호 조회 | ❌ |
| `get_latest_round` | 최신/다음 회차 정보 조회 | ❌ |
| `check_numbers` | 내 번호와 당첨번호 비교 | ❌ |
| `get_balance` | 예치금 잔액 조회 | ✅ |
| `buy_lotto` | 로또 6/45 구매 (자동/수동/반자동) | ✅ |
| `get_purchase_history` | 인터넷 구매 내역 조회 | ✅ |

## 설치

### 바이너리 다운로드 (권장)

[GitHub Releases](https://github.com/imsehwan/dhlottery-mcp/releases)에서 운영체제에 맞는 파일을 다운로드하세요.

```bash
# macOS (Apple Silicon)
curl -L https://github.com/imsehwan/dhlottery-mcp/releases/latest/download/dhlottery-mcp_darwin_arm64.tar.gz | tar xz
chmod +x dhlottery-mcp
```

### 소스에서 빌드

```bash
git clone https://github.com/imsehwan/dhlottery-mcp
cd dhlottery-mcp
go build -o dhlottery-mcp ./cmd/dhlottery-mcp
```

### Docker

```bash
docker run --rm -i \
  -e DHLOTTERY_USER_ID=<아이디> \
  -e DHLOTTERY_USER_PW=<비밀번호> \
  ghcr.io/imsehwan/dhlottery-mcp
```

## Claude Desktop 설정

`~/Library/Application Support/Claude/claude_desktop_config.json`에 추가:

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

> 로그인이 필요 없는 도구(`get_winning_numbers`, `get_latest_round`, `check_numbers`)는 `env` 없이도 동작합니다.

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

## 구매 관련 안내

| 항목 | 내용 |
|------|------|
| 판매 시간 | 일~금 06:00~24:00 / 토(추첨일) 06:00~20:00 |
| 주당 구매 한도 | **최대 5게임** (모바일·인터넷 합산) |
| 구매 방식 | `auto`(자동번호), `manual`(수동번호 6개), `semi`(반자동) |
| 구매 내역 조회 | 인터넷(PC/웹) 구매만 조회됨 (모바일 앱 구매 제외) |

## 법적 고지

- 이 도구는 동행복권의 공식 API를 사용하지 않으며, 비공식 HTTP 요청 기반으로 동작합니다.
- 사용에 따른 모든 책임은 사용자에게 있습니다.
- 과도한 요청으로 인한 계정 제재 등의 불이익에 대해 책임지지 않습니다.

## 라이선스

MIT
