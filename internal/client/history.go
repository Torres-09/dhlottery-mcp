package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// validateDateRange는 날짜 형식과 기간(최대 31일)을 검증합니다.
func validateDateRange(startDate, endDate string) error {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return fmt.Errorf("시작일 형식이 올바르지 않습니다 (YYYY-MM-DD): %s", startDate)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Errorf("종료일 형식이 올바르지 않습니다 (YYYY-MM-DD): %s", endDate)
	}
	if end.Before(start) {
		return fmt.Errorf("종료일이 시작일보다 앞설 수 없습니다")
	}
	if end.Sub(start) > 31*24*time.Hour {
		return fmt.Errorf("조회 기간은 최대 31일입니다")
	}
	return nil
}

// TicketGame은 티켓 내 개별 게임(슬롯) 정보를 담습니다.
type TicketGame struct {
	Slot    string `json:"slot"`    // A, B, C, D, E
	Numbers []int  `json:"numbers"` // 선택 번호
	Mode    string `json:"mode"`    // auto, manual, semi
}

// Purchase는 구매 내역을 담습니다.
type Purchase struct {
	Date      string       `json:"date"`
	DrawDate  string       `json:"draw_date,omitempty"`
	Round     int          `json:"round"`
	GameName  string       `json:"game_name,omitempty"`
	WinResult string       `json:"win_result,omitempty"`
	WinRank   *int         `json:"win_rank,omitempty"`
	WinAmount *int64       `json:"win_amount,omitempty"`
	WinNums   []int        `json:"win_numbers,omitempty"` // 당첨번호 (추첨 후)
	Games     []TicketGame `json:"games,omitempty"`
	OrderNo   string       `json:"order_no,omitempty"`
}

type ledgerItem struct {
	LtGdsCd    string `json:"ltGdsCd"`
	LtGdsNm    string `json:"ltGdsNm"`
	LtEpsd     int    `json:"ltEpsd"`
	LtEpsdView string `json:"ltEpsdView"`
	GmInfo     string `json:"gmInfo"`
	NtslOrdrNo string `json:"ntslOrdrNo"`
	LtWnResult string `json:"ltWnResult"`
	EltOrdrDt  string `json:"eltOrdrDt"`
	EpsdRflDt  string `json:"epsdRflDt"`
	PrchsQty   int    `json:"prchsQty"`
	WnRnk      *int   `json:"wnRnk"`
	LtWnAmt    *int64 `json:"ltWnAmt"`
}

type ledgerAPIResponse struct {
	ResultCode    interface{} `json:"resultCode"`
	ResultMessage interface{} `json:"resultMessage"`
	Data          struct {
		Total int          `json:"total"`
		List  []ledgerItem `json:"list"`
	} `json:"data"`
}

type ticketGameDetail struct {
	Idx  string `json:"idx"`  // A, B, C, D, E
	Num  []int  `json:"num"`
	Rank int    `json:"rank"`
	Amt  int64  `json:"amt"`
	Type int    `json:"type"` // 1=수동, 2=반자동, 3=자동
}

type ticketDetailResponse struct {
	ResultCode interface{} `json:"resultCode"`
	Data       struct {
		Ticket struct {
			GameRound   int                `json:"game_round"`
			SaleDate    string             `json:"sale_date"`
			DrawDate    string             `json:"draw_date"`
			Drawed      bool               `json:"drawed"`
			WinNum      []int              `json:"win_num"`
			WinTotalAmt int64              `json:"win_total_amt"`
			GameDtl     []ticketGameDetail `json:"game_dtl"`
		} `json:"ticket"`
	} `json:"data"`
}

// getTicketDetail은 티켓 상세(번호 포함)를 조회합니다.
func (c *Client) getTicketDetail(ntslOrdrNo, barcd, srchStrDt, srchEndDt string) ([]TicketGame, []int, error) {
	params := url.Values{
		"ntslOrdrNo": {ntslOrdrNo},
		"barcd":      {barcd},
		"srchStrDt":  {srchStrDt},
		"srchEndDt":  {srchEndDt},
	}
	req, err := http.NewRequest(http.MethodGet,
		wwwBaseURL+"/mypage/lotto645TicketDetail.do?"+params.Encode(), nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("AJAX", "true")
	req.Header.Set("Referer", wwwBaseURL+"/mypage/lotto645Ticket")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	var detail ticketDetailResponse
	if err := json.Unmarshal(body, &detail); err != nil {
		return nil, nil, err
	}

	modeMap := map[int]string{1: "manual", 2: "semi", 3: "auto"}
	var games []TicketGame
	for _, g := range detail.Data.Ticket.GameDtl {
		mode := modeMap[g.Type]
		if mode == "" {
			mode = "auto"
		}
		games = append(games, TicketGame{
			Slot:    g.Idx,
			Numbers: g.Num,
			Mode:    mode,
		})
	}

	winNums := detail.Data.Ticket.WinNum
	// 추첨 전이면 당첨번호가 모두 0 → nil로 반환
	allZero := true
	for _, n := range winNums {
		if n != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		winNums = nil
	}

	return games, winNums, nil
}

// GetPurchaseHistory는 구매 내역을 조회합니다.
// 최대 조회 기간은 31일입니다.
func (c *Client) GetPurchaseHistory(startDate, endDate string) ([]Purchase, error) {
	if err := c.EnsureLoggedIn(); err != nil {
		return nil, err
	}

	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	if err := validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	startFormatted := formatDateParam(startDate)
	endFormatted := formatDateParam(endDate)

	params := url.Values{
		"srchStrDt":          {startFormatted},
		"srchEndDt":          {endFormatted},
		"sort":               {""},
		"ltGdsCd":            {""},
		"winResult":          {""},
		"lramSmam":           {""},
		"pageNum":            {"1"},
		"recordCountPerPage": {"50"},
	}

	req, err := http.NewRequest(http.MethodGet,
		wwwBaseURL+"/mypage/selectMyLotteryledger.do?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("AJAX", "true")
	req.Header.Set("Referer", wwwBaseURL+"/mypage/mylotteryledger")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("구매내역 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	// 세션 만료 감지: 로그인 페이지로 리다이렉트된 경우
	if resp.Request.URL.Path == "/login" || resp.Request.URL.Path == "/errorPage" {
		c.InvalidateSession()
		if err := c.Login(); err != nil {
			return nil, err
		}
		return c.GetPurchaseHistory(startDate, endDate)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	var apiResp ledgerAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		snippet := string(body)
		if len(snippet) > 300 {
			snippet = snippet[:300]
		}
		return nil, fmt.Errorf("구매내역 응답 파싱 실패: %w\n응답: %s", err, snippet)
	}

	if apiResp.ResultCode != nil {
		msg := ""
		if apiResp.ResultMessage != nil {
			msg = fmt.Sprintf("%v", apiResp.ResultMessage)
		}
		return nil, fmt.Errorf("구매내역 조회 실패: %s", msg)
	}

	var purchases []Purchase
	for _, item := range apiResp.Data.List {
		p := Purchase{
			Date:      item.EltOrdrDt,
			DrawDate:  item.EpsdRflDt,
			Round:     item.LtEpsd,
			GameName:  item.LtGdsNm,
			WinResult: item.LtWnResult,
			WinRank:   item.WnRnk,
			WinAmount: item.LtWnAmt,
			OrderNo:   item.NtslOrdrNo,
		}

		// 로또6/45 티켓만 번호 조회
		if item.LtGdsCd == "LO40" && item.GmInfo != "" {
			games, winNums, err := c.getTicketDetail(item.NtslOrdrNo, item.GmInfo, startFormatted, endFormatted)
			if err == nil {
				p.Games = games
				p.WinNums = winNums
			}
		}

		purchases = append(purchases, p)
	}

	return purchases, nil
}

// formatDateParam은 "2006-01-02" 형식을 "20060102" 형식으로 변환합니다.
func formatDateParam(s string) string {
	if len(s) == 8 {
		return s
	}
	if len(s) == 10 && s[4] == '-' {
		return s[:4] + s[5:7] + s[8:]
	}
	return s
}
