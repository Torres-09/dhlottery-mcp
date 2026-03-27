package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Balance는 예치금 잔액 정보를 담습니다.
type Balance struct {
	Amount int64 `json:"balance"`
}

type userMndpResponse struct {
	ResultCode    interface{} `json:"resultCode"`
	ResultMessage interface{} `json:"resultMessage"`
	Data          struct {
		UserMndp struct {
			TotalAmt int64 `json:"totalAmt"`
		} `json:"userMndp"`
	} `json:"data"`
}

// GetBalance는 예치금 잔액을 조회합니다.
func (c *Client) GetBalance() (*Balance, error) {
	if err := c.EnsureLoggedIn(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, wwwBaseURL+"/mypage/selectUserMndp.do", nil)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Referer", wwwBaseURL+"/mypage/home")
	req.Header.Set("AJAX", "true")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("예치금 조회 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	// 세션 만료 감지: 로그인 페이지로 리다이렉트된 경우
	if resp.Request.URL.Path == "/login" || resp.Request.URL.Path == "/errorPage" {
		c.InvalidateSession()
		if err := c.Login(); err != nil {
			return nil, err
		}
		return c.fetchBalance()
	}

	return c.parseBalanceJSON(resp)
}

func (c *Client) fetchBalance() (*Balance, error) {
	req, err := http.NewRequest(http.MethodGet, wwwBaseURL+"/mypage/selectUserMndp.do", nil)
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Referer", wwwBaseURL+"/mypage/home")
	req.Header.Set("AJAX", "true")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("예치금 재조회 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	return c.parseBalanceJSON(resp)
}

func (c *Client) parseBalanceJSON(resp *http.Response) (*Balance, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	var apiResp userMndpResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("예치금 응답 파싱 실패: %w", err)
	}

	return &Balance{Amount: apiResp.Data.UserMndp.TotalAmt}, nil
}
