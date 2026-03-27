package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// WinningNumbers는 당첨번호 정보를 담습니다.
type WinningNumbers struct {
	Round        int    `json:"round"`
	Date         string `json:"date"`
	Numbers      []int  `json:"numbers"`
	Bonus        int    `json:"bonus"`
	TotalSales   int64  `json:"total_sales"`
	FirstPrize   int64  `json:"first_prize"`
	FirstWinners int    `json:"first_winners"`
}

// LatestRound는 최신 회차 정보를 담습니다.
type LatestRound struct {
	LatestDrawnRound int    `json:"latest_drawn_round"`
	NextRound        int    `json:"next_round"`
	LatestDrawDate   string `json:"latest_draw_date"`
}

// 새 API 응답 구조체 (2025년 이후 변경된 dhlottery.co.kr API)
type lt645Item struct {
	LtEpsd              int    `json:"ltEpsd"`
	LtRflYmd            string `json:"ltRflYmd"`
	Tm1WnNo             int    `json:"tm1WnNo"`
	Tm2WnNo             int    `json:"tm2WnNo"`
	Tm3WnNo             int    `json:"tm3WnNo"`
	Tm4WnNo             int    `json:"tm4WnNo"`
	Tm5WnNo             int    `json:"tm5WnNo"`
	Tm6WnNo             int    `json:"tm6WnNo"`
	BnsWnNo             int    `json:"bnsWnNo"`
	RlvtEpsdSumNtslAmt  int64  `json:"rlvtEpsdSumNtslAmt"`
	Rnk1WnAmt           int64  `json:"rnk1WnAmt"`
	Rnk1WnNope          int    `json:"rnk1WnNope"`
}

type lt645APIResponse struct {
	ResultCode    interface{} `json:"resultCode"`
	ResultMessage interface{} `json:"resultMessage"`
	Data          struct {
		List []lt645Item `json:"list"`
	} `json:"data"`
}

// GetWinningNumbers는 특정 회차의 당첨번호를 조회합니다.
func (c *Client) GetWinningNumbers(round int) (*WinningNumbers, error) {
	url := fmt.Sprintf("%s/lt645/selectPstLt645InfoNew.do?srchDir=center&srchLtEpsd=%d", baseURL, round)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", "https://dhlottery.co.kr/lt645/result")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("당첨번호 조회 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	var apiResp lt645APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w", err)
	}

	// 해당 회차 찾기
	for _, item := range apiResp.Data.List {
		if item.LtEpsd == round {
			date := formatDate(item.LtRflYmd)
			return &WinningNumbers{
				Round:        item.LtEpsd,
				Date:         date,
				Numbers:      []int{item.Tm1WnNo, item.Tm2WnNo, item.Tm3WnNo, item.Tm4WnNo, item.Tm5WnNo, item.Tm6WnNo},
				Bonus:        item.BnsWnNo,
				TotalSales:   item.RlvtEpsdSumNtslAmt,
				FirstPrize:   item.Rnk1WnAmt,
				FirstWinners: item.Rnk1WnNope,
			}, nil
		}
	}

	return nil, fmt.Errorf("%d회차 당첨번호를 찾을 수 없습니다. 유효한 회차 번호를 입력해주세요", round)
}

// GetLatestRound는 최신 회차 정보를 조회합니다.
func (c *Client) GetLatestRound() (*LatestRound, error) {
	// 1단계: 결과 페이지 HTML에서 드롭다운 최대 회차 파싱
	latestRound, err := c.fetchLatestRoundFromHTML()
	if err != nil {
		return nil, err
	}

	// 2단계: 해당 회차 상세 정보 조회
	winNums, err := c.GetWinningNumbers(latestRound)
	if err != nil {
		return nil, fmt.Errorf("최신 회차 정보 조회 실패: %w", err)
	}

	return &LatestRound{
		LatestDrawnRound: latestRound,
		NextRound:        latestRound + 1,
		LatestDrawDate:   winNums.Date,
	}, nil
}

// fetchLatestRoundFromHTML는 결과 페이지 HTML의 드롭다운에서 최신 회차를 파싱합니다.
func (c *Client) fetchLatestRoundFromHTML() (int, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+"/lt645/result", nil)
	if err != nil {
		return 0, fmt.Errorf("요청 생성 실패: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("결과 페이지 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("HTML 파싱 실패: %w", err)
	}

	maxRound := 0
	doc.Find(".option-il").Each(func(_ int, s *goquery.Selection) {
		val, exists := s.Attr("data-value")
		if !exists {
			return
		}
		n, err := strconv.Atoi(strings.TrimSpace(val))
		if err == nil && n > maxRound {
			maxRound = n
		}
	})

	if maxRound == 0 {
		return 0, fmt.Errorf("회차 정보를 찾을 수 없습니다. 사이트 구조가 변경되었을 수 있습니다")
	}

	return maxRound, nil
}

// formatDate는 "20241214" 형식을 "2024-12-14" 형식으로 변환합니다.
func formatDate(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 8 {
		y, m, d := s[:4], s[4:6], s[6:]
		if _, err := strconv.Atoi(s); err == nil {
			return y + "-" + m + "-" + d
		}
	}
	return s
}
