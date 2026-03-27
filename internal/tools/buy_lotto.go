package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/Torres-09/dhlottery-mcp/internal/client"
	"github.com/Torres-09/dhlottery-mcp/internal/config"
)

func RegisterBuyLotto(s *server.MCPServer, c *client.Client, cfg *config.Config) {
	tool := mcp.NewTool("buy_lotto",
		mcp.WithDescription("로또 6/45를 구매합니다. 실제 금전 거래가 발생하므로 반드시 사용자 확인 후 실행해야 합니다."),
		mcp.WithNumber("count",
			mcp.Required(),
			mcp.Description("구매할 게임 수 (1~5)"),
		),
		mcp.WithString("mode",
			mcp.Description("구매 방식: auto(자동), manual(수동), semi(반자동). 기본값: auto"),
		),
		mcp.WithArray("numbers",
			mcp.Description("수동/반자동 시 번호 지정. 수동: 6개 (예: [[1,2,3,4,5,6]]), 반자동: 일부만 (예: [[1,2,3]])"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if !cfg.HasCredentials() {
			return mcp.NewToolResultError(
				"로그인 자격증명이 설정되지 않았습니다.\n" +
					"다음 환경변수를 설정해주세요:\n" +
					"  DHLOTTERY_USER_ID=<동행복권 아이디>\n" +
					"  DHLOTTERY_USER_PW=<동행복권 비밀번호>",
			), nil
		}

		args := req.GetArguments()
		countFloat, ok := args["count"].(float64)
		if !ok {
			return mcp.NewToolResultError("count 파라미터가 올바르지 않습니다"), nil
		}
		count := int(countFloat)

		if count < 1 || count > 5 {
			return mcp.NewToolResultError("게임 수는 1~5개 사이여야 합니다"), nil
		}

		mode := "auto"
		if modeVal, ok := args["mode"].(string); ok && modeVal != "" {
			mode = modeVal
		}

		// numbers 파싱
		var numberSets [][]int
		if numbersRaw, ok := args["numbers"]; ok && numbersRaw != nil {
			numbersSlice, ok := numbersRaw.([]interface{})
			if !ok {
				return mcp.NewToolResultError("numbers는 배열이어야 합니다"), nil
			}

			for i, setRaw := range numbersSlice {
				setSlice, ok := setRaw.([]interface{})
				if !ok {
					return mcp.NewToolResultError(fmt.Sprintf("%d번째 번호 세트가 올바르지 않습니다", i+1)), nil
				}

				var nums []int
				for _, numRaw := range setSlice {
					numFloat, ok := numRaw.(float64)
					if !ok {
						return mcp.NewToolResultError(fmt.Sprintf("%d번째 번호 세트에 잘못된 값이 있습니다", i+1)), nil
					}
					num := int(numFloat)
					if num < 1 || num > 45 {
						return mcp.NewToolResultError(fmt.Sprintf("번호는 1~45 사이여야 합니다: %d", num)), nil
					}
					nums = append(nums, num)
				}
				numberSets = append(numberSets, nums)
			}
		}

		// 수동 모드 검증
		if mode == "manual" {
			if len(numberSets) == 0 {
				return mcp.NewToolResultError("수동 모드에서는 numbers 파라미터가 필요합니다"), nil
			}
			for i, nums := range numberSets {
				if len(nums) != 6 {
					return mcp.NewToolResultError(fmt.Sprintf("수동 모드에서 %d번째 세트는 정확히 6개 번호가 필요합니다 (현재: %d개)", i+1, len(nums))), nil
				}
			}
		}

		// 현재 구매 가능 회차 조회
		latest, err := c.GetLatestRound()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("회차 정보 조회 실패: %v", err)), nil
		}
		round := latest.NextRound

		// 게임 설정 구성
		games := make([]client.BuyGame, count)
		for i := range games {
			games[i] = client.BuyGame{Mode: mode}
			if i < len(numberSets) {
				games[i].Numbers = numberSets[i]
			}
		}

		result, err := c.BuyLotto(round, games)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		output := map[string]interface{}{
			"success":      result.Success,
			"round":        result.Round,
			"games":        result.Games,
			"total_amount": result.TotalAmount,
			"message":      result.Message,
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("결과 직렬화 실패: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	})
}
