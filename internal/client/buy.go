package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// BuyGame은 구매할 게임 설정을 담습니다.
type BuyGame struct {
	Mode    string // "auto", "manual", "semi"
	Numbers []int  // 수동/반자동 시 지정 번호
}

// BuyResult는 구매 결과를 담습니다.
type BuyResult struct {
	Success     bool        `json:"success"`
	Round       int         `json:"round"`
	Games       []GameSlot  `json:"games"`
	TotalAmount int         `json:"total_amount"`
	Message     string      `json:"message"`
}

// GameSlot은 구매된 게임 슬롯 정보를 담습니다.
type GameSlot struct {
	Slot    string `json:"slot"`
	Numbers []int  `json:"numbers"`
	Mode    string `json:"mode"`
}

type buyAPIResponse struct {
	Result struct {
		ResultMsg   string `json:"resultMsg"`
		WinNumbers  string `json:"winNumbers"`
		BuyRound    int    `json:"buyRound"`
		NatWinNum   []struct {
			Issuing string `json:"issuing"`
			IssNos  string `json:"issNos"`
		} `json:"natWinNum"`
	} `json:"result"`
}

type buyParam struct {
	GenType          string `json:"genType"`
	ArrGameChoiceNum interface{} `json:"arrGameChoiceNum"`
	Alpabet          string `json:"alpabet"`
}

var slotNames = []string{"A", "B", "C", "D", "E"}

// BuyLotto는 로또를 구매합니다.
func (c *Client) BuyLotto(round int, games []BuyGame) (*BuyResult, error) {
	if err := c.EnsureLoggedIn(); err != nil {
		return nil, err
	}

	if len(games) == 0 || len(games) > 5 {
		return nil, fmt.Errorf("게임 수는 1~5개 사이여야 합니다 (입력: %d개)", len(games))
	}

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

	formData := url.Values{
		"round":       {fmt.Sprintf("%d", round)},
		"direct":      {"172.17.20.52"},
		"nBuyAmount":  {fmt.Sprintf("%d", len(games)*1000)},
		"param":       {string(paramJSON)},
		"gameCnt":     {fmt.Sprintf("%d", len(games))},
	}

	req, err := http.NewRequest(http.MethodPost, buyURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("구매 요청 생성 실패: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Origin", "https://ol.dhlottery.co.kr")
	req.Header.Set("Referer", "https://ol.dhlottery.co.kr/olotto/game/game645.do")

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
		return nil, fmt.Errorf("구매 응답 파싱 실패: %w", err)
	}

	if apiResp.Result.ResultMsg != "SUCCESS" {
		msg := apiResp.Result.ResultMsg
		switch {
		case strings.Contains(msg, "잔액"):
			return nil, fmt.Errorf("예치금이 부족합니다. 동행복권 사이트에서 충전 후 다시 시도해주세요")
		case strings.Contains(msg, "한도"):
			return nil, fmt.Errorf("주당 구매 한도를 초과했습니다. 로또 6/45는 주당 최대 5게임까지 구매 가능합니다")
		default:
			return nil, fmt.Errorf("구매에 실패했습니다: %s", msg)
		}
	}

	// 구매된 게임 번호 파싱
	var gameSlots []GameSlot
	for i, natWin := range apiResp.Result.NatWinNum {
		if i >= len(games) {
			break
		}
		var nums []int
		for _, numStr := range strings.Split(natWin.IssNos, ",") {
			var n int
			fmt.Sscanf(strings.TrimSpace(numStr), "%d", &n)
			if n > 0 {
				nums = append(nums, n)
			}
		}

		mode := "auto"
		if games[i].Mode != "" {
			mode = games[i].Mode
		}

		gameSlots = append(gameSlots, GameSlot{
			Slot:    slotNames[i],
			Numbers: nums,
			Mode:    mode,
		})
	}

	return &BuyResult{
		Success:     true,
		Round:       apiResp.Result.BuyRound,
		Games:       gameSlots,
		TotalAmount: len(games) * 1000,
		Message:     "구매가 완료되었습니다",
	}, nil
}
