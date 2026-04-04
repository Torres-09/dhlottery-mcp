package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// BuyGame은 구매할 게임 설정을 담습니다.
type BuyGame struct {
	Mode    string // "auto", "manual", "semi"
	Numbers []int  // 수동/반자동 시 지정 번호
}

// BuyResult는 구매 결과를 담습니다.
type BuyResult struct {
	Success     bool       `json:"success"`
	Round       int        `json:"round"`
	Games       []GameSlot `json:"games"`
	TotalAmount int        `json:"total_amount"`
	Message     string     `json:"message"`
}

// GameSlot은 구매된 게임 슬롯 정보를 담습니다.
type GameSlot struct {
	Slot    string `json:"slot"`
	Numbers []int  `json:"numbers"`
	Mode    string `json:"mode"`
}

type buyAPIResponse struct {
	LoginYn    string `json:"loginYn"`
	UsePopupYn string `json:"usePopupYn"`
	Result     struct {
		ResultCode       string   `json:"resultCode"`
		ResultMsg        string   `json:"resultMsg"`
		BuyRound         string   `json:"buyRound"`
		ArrGameChoiceNum []string `json:"arrGameChoiceNum"` // "A|08|11|12|17|35|38"
		NBuyAmount       int      `json:"nBuyAmount"`
		DrawDate         string   `json:"drawDate"`
	} `json:"result"`
}

type readySocketResponse struct {
	ReadyIP  string `json:"ready_ip"`
	ReadyCnt string `json:"ready_cnt"`
}

type buyParam struct {
	GenType          string      `json:"genType"`
	ArrGameChoiceNum interface{} `json:"arrGameChoiceNum"`
	Alpabet          string      `json:"alpabet"`
}

var slotNames = []string{"A", "B", "C", "D", "E"}

const (
	gamePageURL    = "https://ol.dhlottery.co.kr/olotto/game/game645.do"
	readySocketURL = "https://ol.dhlottery.co.kr/olotto/game/egovUserReadySocket.json"
)

// gamePageInfo는 game645.do 페이지에서 추출한 정보입니다.
type gamePageInfo struct {
	CurRound         string
	RoundDrawDate    string
	WamtPayTlmtEndDt string
}

// visitGamePage는 구매 전 ol.dhlottery.co.kr 세션을 설정하고 페이지 정보를 추출합니다.
func (c *Client) visitGamePage() (*gamePageInfo, error) {
	req, err := http.NewRequest(http.MethodGet, gamePageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", wwwBaseURL+"/")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)
	info := &gamePageInfo{
		CurRound:         extractElementText(html, "curRound"),
		RoundDrawDate:    extractInputValue(html, "ROUND_DRAW_DATE"),
		WamtPayTlmtEndDt: extractInputValue(html, "WAMT_PAY_TLMT_END_DT"),
	}

	if info.CurRound == "" {
		return nil, fmt.Errorf("게임 페이지에서 회차 정보를 찾을 수 없습니다")
	}

	return info, nil
}

// getReadyToken은 접속자 대기 시스템에서 구매 토큰(ready_ip)을 획득합니다.
func (c *Client) getReadyToken() (string, error) {
	req, err := http.NewRequest(http.MethodPost, readySocketURL, strings.NewReader("{}"))
	if err != nil {
		return "", fmt.Errorf("대기 요청 생성 실패: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Origin", "https://ol.dhlottery.co.kr")
	req.Header.Set("Referer", gamePageURL)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("대기 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("대기 응답 읽기 실패: %w", err)
	}

	var readyResp readySocketResponse
	if err := json.Unmarshal(body, &readyResp); err != nil {
		return "", fmt.Errorf("대기 응답 파싱 실패: %w", err)
	}

	return readyResp.ReadyIP, nil
}

// BuyLotto는 로또를 구매합니다.
func (c *Client) BuyLotto(round int, games []BuyGame) (*BuyResult, error) {
	if err := c.EnsureLoggedIn(); err != nil {
		return nil, err
	}

	if len(games) == 0 || len(games) > 5 {
		return nil, fmt.Errorf("게임 수는 1~5개 사이여야 합니다 (입력: %d개)", len(games))
	}

	// 1단계: 게임 페이지 방문 — ol 세션 설정 + 페이지 정보 추출
	pageInfo, err := c.visitGamePage()
	if err != nil {
		return nil, fmt.Errorf("게임 페이지 접근 실패: %w", err)
	}

	// round가 0이면 페이지에서 추출한 회차 사용
	if round == 0 {
		r, err := strconv.Atoi(pageInfo.CurRound)
		if err != nil {
			return nil, fmt.Errorf("회차 변환 실패: %w", err)
		}
		round = r
	}

	// 2단계: 접속자 대기 토큰 획득
	direct, err := c.getReadyToken()
	if err != nil {
		return nil, fmt.Errorf("구매 토큰 획득 실패: %w", err)
	}

	// 3단계: 게임 파라미터 구성
	params := make([]buyParam, len(games))
	for i, game := range games {
		p := buyParam{
			Alpabet: slotNames[i],
		}

		switch game.Mode {
		case "auto", "":
			p.GenType = "0"
			p.ArrGameChoiceNum = nil
		case "manual":
			p.GenType = "1"
			nums := make([]string, len(game.Numbers))
			for j, n := range game.Numbers {
				nums[j] = fmt.Sprintf("%d", n)
			}
			p.ArrGameChoiceNum = strings.Join(nums, ",")
		case "semi":
			p.GenType = "2"
			nums := make([]string, len(game.Numbers))
			for j, n := range game.Numbers {
				nums[j] = fmt.Sprintf("%d", n)
			}
			p.ArrGameChoiceNum = strings.Join(nums, ",")
		default:
			return nil, fmt.Errorf("알 수 없는 구매 모드입니다: %s (auto/manual/semi 중 선택)", game.Mode)
		}

		params[i] = p
	}

	paramJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("파라미터 직렬화 실패: %w", err)
	}

	// 4단계: execBuy.do 요청 (plain 파라미터)
	formData := url.Values{
		"round":                {fmt.Sprintf("%d", round)},
		"direct":               {direct},
		"nBuyAmount":           {fmt.Sprintf("%d", len(games)*1000)},
		"param":                {string(paramJSON)},
		"ROUND_DRAW_DATE":      {pageInfo.RoundDrawDate},
		"WAMT_PAY_TLMT_END_DT": {pageInfo.WamtPayTlmtEndDt},
		"gameCnt":              {fmt.Sprintf("%d", len(games))},
		"saleMdaDcd":           {"10"},
	}

	req, err := http.NewRequest(http.MethodPost, buyURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("구매 요청 생성 실패: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Origin", "https://ol.dhlottery.co.kr")
	req.Header.Set("Referer", gamePageURL)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("구매 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("구매 응답 읽기 실패: %w", err)
	}

	var apiResp buyAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		bodyStr := string(body)
		if strings.Contains(bodyStr, "서비스 중단") || strings.Contains(bodyStr, "토요일") {
			return nil, fmt.Errorf("현재 구매 서비스가 중단되어 있습니다. 토요일 20:00 이후 ~ 일요일 오후는 구매가 불가능합니다")
		}
		snippet := bodyStr
		if len(snippet) > 300 {
			snippet = snippet[:300]
		}
		return nil, fmt.Errorf("구매 응답 파싱 실패: %w\n응답 내용: %s", err, snippet)
	}

	// 로그인 세션 만료 확인
	if apiResp.LoginYn == "N" {
		c.loggedIn = false
		return nil, fmt.Errorf("세션이 만료되었습니다. 다시 시도해주세요")
	}

	result := apiResp.Result
	if result.ResultMsg != "SUCCESS" {
		msg := result.ResultMsg
		code := result.ResultCode
		switch {
		case strings.Contains(msg, "잔액") || strings.Contains(msg, "예치금"):
			return nil, fmt.Errorf("예치금이 부족합니다. 동행복권 사이트에서 충전 후 다시 시도해주세요")
		case strings.Contains(msg, "한도"):
			return nil, fmt.Errorf("주당 구매 한도를 초과했습니다")
		case code == "E98":
			return nil, fmt.Errorf("비정상 접속으로 차단되었습니다. 잠시 후 다시 시도해주세요")
		default:
			return nil, fmt.Errorf("구매에 실패했습니다: [%s] %s", code, msg)
		}
	}

	// 구매된 게임 번호 파싱 — "A|08|11|12|17|35|38" 형식
	var gameSlots []GameSlot
	for i, choiceStr := range result.ArrGameChoiceNum {
		if i >= len(games) {
			break
		}
		parts := strings.Split(choiceStr, "|")
		slot := ""
		var nums []int
		for j, p := range parts {
			if j == 0 {
				slot = p
				continue
			}
			n, err := strconv.Atoi(p)
			if err == nil && n > 0 {
				nums = append(nums, n)
			}
		}

		mode := "auto"
		if games[i].Mode != "" {
			mode = games[i].Mode
		}

		gameSlots = append(gameSlots, GameSlot{
			Slot:    slot,
			Numbers: nums,
			Mode:    mode,
		})
	}

	buyRound, _ := strconv.Atoi(result.BuyRound)

	return &BuyResult{
		Success:     true,
		Round:       buyRound,
		Games:       gameSlots,
		TotalAmount: result.NBuyAmount,
		Message:     "구매가 완료되었습니다",
	}, nil
}

func extractElementText(html, id string) string {
	re := regexp.MustCompile(`id="` + id + `"[^>]*>([^<]*)`)
	m := re.FindStringSubmatch(html)
	if m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func extractInputValue(html, id string) string {
	re := regexp.MustCompile(`<input[^>]*(?:id|name)="` + id + `"[^>]*value="([^"]*)"`)
	m := re.FindStringSubmatch(html)
	if m != nil {
		return m[1]
	}
	return ""
}
