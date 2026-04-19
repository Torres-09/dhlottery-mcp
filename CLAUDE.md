# Design Guide — dhlottery-mcp

동행복권(dhlottery.co.kr) 사이트의 디자인 언어를 기준으로 UI를 생성한다.
아래 규칙을 모든 HTML/CSS 산출물에 적용한다.

---

## 1. Color Palette

### Primary
| 역할 | Hex | 용도 |
|------|-----|------|
| Brand Red | `#E8001D` | 주요 CTA 버튼, 헤더 포인트, 로또 볼 |
| Brand Red Dark | `#C0001A` | hover 상태 |
| Brand Red Light | `#FFF0F2` | 빨간 계열 배경, 뱃지 fill |

### Neutral
| 역할 | Hex | 용도 |
|------|-----|------|
| Page BG | `#F7F7F7` | 전체 배경 (순백 금지) |
| Surface White | `#FFFFFF` | 카드, 패널 배경 |
| Border | `#E0E0E0` | 구분선, 테두리 |
| Border Dark | `#CCCCCC` | 강조 구분선 |
| Text Primary | `#222222` | 본문 텍스트 |
| Text Secondary | `#666666` | 부제목, 설명 텍스트 |
| Text Muted | `#999999` | placeholder, 비활성 |

### Accent
| 역할 | Hex | 용도 |
|------|-----|------|
| Lotto Yellow | `#FFD700` | 로또 볼 1-10번대, 강조 |
| Lotto Blue | `#1565C0` | 로또 볼 번호 색상 |
| Lotto Gray | `#9E9E9E` | 비활성 볼 |
| Link Blue | `#0066CC` | 텍스트 링크 |
| Success Green | `#2E7D32` | 당첨, 완료 상태 |
| Warning Orange | `#F57C00` | 주의 안내 |

### 로또 볼 번호별 색상 (필수 준수)
```
1–10  : #F9CD04 (노랑), text #333
11–20 : #69C8F2 (하늘), text #fff
21–30 : #FF7272 (빨강), text #fff
31–40 : #AAAAAA (회색), text #fff
41–45 : #B0D840 (초록), text #333
```

---

## 2. Typography

```css
font-family: "Malgun Gothic", "맑은 고딕", "Apple SD Gothic Neo",
             "Noto Sans KR", sans-serif;
```

- **절대 사용 금지**: Inter, Roboto, Poppins, DM Sans, Geist
- 한국어 UI이므로 KR 웹폰트를 우선한다
- 외부 폰트 필요 시: `Noto Sans KR` (Google Fonts)

### Scale
| 요소 | Size | Weight |
|------|------|--------|
| 페이지 제목 | 22px | 700 |
| 섹션 헤더 | 18px | 700 |
| 카드 제목 | 15px | 700 |
| 본문 | 14px | 400 |
| 부가 설명 | 13px | 400 |
| 캡션 | 12px | 400 |

---

## 3. Layout & Spacing

- **기본 컨테이너**: `max-width: 1200px; margin: 0 auto; padding: 0 16px`
- **섹션 간격**: `margin-bottom: 32px`
- **카드 내부 padding**: `20px 24px`
- **그리드**: CSS Grid 또는 Flexbox 사용, float 금지

### 카드 컴포넌트
```css
background: #FFFFFF;
border: 1px solid #E0E0E0;
border-radius: 4px;           /* 동행복권은 각진 편, 6px 초과 금지 */
padding: 20px 24px;
```

---

## 4. Component Rules

### 버튼
```css
/* Primary CTA */
background: #E8001D;
color: #FFFFFF;
border: none;
border-radius: 4px;
padding: 10px 24px;
font-size: 14px;
font-weight: 700;
cursor: pointer;

/* Primary hover */
background: #C0001A;

/* Secondary (outline) */
background: #FFFFFF;
color: #E8001D;
border: 1px solid #E8001D;
border-radius: 4px;
padding: 9px 24px;
```

### 테이블
- 헤더: `background: #F2F2F2`, `color: #444444`, `font-weight: 700`
- 행 구분: `border-bottom: 1px solid #E8E8E8`
- 홀짝 교차색 (선택): `#FAFAFA`
- border-radius 없음 (테이블은 각짐 유지)

### 뱃지 / 상태 태그
```css
/* 당첨 */
background: #E8F5E9; color: #2E7D32; border: 1px solid #A5D6A7;

/* 미당첨 */
background: #F5F5F5; color: #757575; border: 1px solid #BDBDBD;

/* 대기 */
background: #FFF3E0; color: #F57C00; border: 1px solid #FFCC80;
```
- `border-radius: 2px`, `font-size: 12px`, `padding: 2px 8px`

### 로또 번호 볼
```css
.ball {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  font-size: 15px;
  font-weight: 700;
}
```
번호 범위별 색상은 Section 1 참조.

### 입력 폼
```css
border: 1px solid #CCCCCC;
border-radius: 4px;
padding: 8px 12px;
font-size: 14px;
color: #222222;
background: #FFFFFF;

/* focus */
border-color: #E8001D;
outline: none;
```

---

## 5. Header / Navigation 패턴

- 상단 GNB: 흰색 배경(`#FFFFFF`) + 하단 `border-bottom: 2px solid #E8001D`
- 로고 영역 좌측, 메뉴 우측 정렬
- 활성 메뉴: `color: #E8001D`, `font-weight: 700`
- 비활성 메뉴: `color: #444444`

---

## 6. 금지 사항 (Avoid Generic AI Aesthetics)

```
❌ 파란/보라 계열 그라디언트 배경
❌ border-radius > 6px (로또 볼 제외)
❌ box-shadow: 0 4px 20px rgba(0,0,0,0.15) 같은 과도한 그림자
❌ Inter, Roboto, system-ui 폰트
❌ glassmorphism, neumorphism 효과
❌ 순백(#FFFFFF) 전체 배경
❌ 파란 primary 버튼 (#3b82f6 계열)
❌ 둥근 pill 버튼 (border-radius: 999px)
```

---

## 7. 참조 사이트

- 메인: https://www.dhlottery.co.kr/
- 로또 결과: https://www.dhlottery.co.kr/lt645/result
- 구매 화면: https://www.dhlottery.co.kr/gameInfo.do?method=buyLotto

위 페이지의 레이아웃, 색상, 컴포넌트 배치를 디자인 기준으로 삼는다.
