package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// PurchaseGame은 구매한 게임 정보를 담습니다.
type PurchaseGame struct {
	Numbers []int  `json:"numbers"`
	Mode    string `json:"mode"`
}

// Purchase는 구매 내역을 담습니다.
type Purchase struct {
	Date     string         `json:"date"`
	Round    int            `json:"round"`
	Games    []PurchaseGame `json:"games"`
	Amount   int64          `json:"amount"`
	WinResult string        `json:"win_result,omitempty"`
	GameInfo string         `json:"game_info,omitempty"`
}

type ledgerItem struct {
	LtGdsCd    string `json:"ltGdsCd"`
	LtEpsd     int    `json:"ltEpsd"`
	GmInfo     string `json:"gmInfo"`
	NtslOrdrNo string `json:"ntslOrdrNo"`
	LtWnResult string `json:"ltWnResult"`
	NtslDt     string `json:"ntslDt"`
	NtslAmt    int64  `json:"ntslAmt"`
}

type ledgerAPIResponse struct {
	ResultCode    interface{} `json:"resultCode"`
	ResultMessage interface{} `json:"resultMessage"`
	Data          struct {
		Total int          `json:"total"`
		List  []ledgerItem `json:"list"`
	} `json:"data"`
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

	// YYYYMMDD 형식으로 변환
	startFormatted := formatDateParam(startDate)
	endFormatted := formatDateParam(endDate)

	params := url.Values{
		"srchStrDt":          {startFormatted},
		"srchEndDt":          {endFormatted},
		"sort":               {"1"},
		"ltGdsCd":            {"LO40"}, // 로또 6/45
		"winResult":          {""},
		"pageNum":            {"1"},
		"recordCountPerPage": {"50"},
	}

	apiURL := wwwBaseURL + "/mypage/selectMyLotteryledger.do?" + params.Encode()

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("AJAX", "true")
	req.Header.Set("Referer", wwwBaseURL+"/mypage/mylotteryledger")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("구매내역 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	var apiResp ledgerAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("구매내역 응답 파싱 실패: %w", err)
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
		date := formatDate(item.NtslDt)
		purchases = append(purchases, Purchase{
			Date:      date,
			Round:     item.LtEpsd,
			Games:     []PurchaseGame{},
			Amount:    item.NtslAmt,
			WinResult: item.LtWnResult,
			GameInfo:  item.GmInfo,
		})
	}

	return purchases, nil
}

// formatDateParam은 "2006-01-02" 형식을 "20060102" 형식으로 변환합니다.
func formatDateParam(s string) string {
	// 이미 YYYYMMDD 형식이면 그대로 반환
	if len(s) == 8 {
		return s
	}
	// YYYY-MM-DD 형식이면 하이픈 제거
	if len(s) == 10 && s[4] == '-' {
		return s[:4] + s[5:7] + s[8:]
	}
	return s
}
