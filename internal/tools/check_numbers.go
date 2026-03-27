package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/imsehwan/dhlottery-mcp/internal/client"
)

// CheckResult는 번호 확인 결과를 담습니다.
type CheckResult struct {
	Numbers      []int `json:"numbers"`
	Matched      []int `json:"matched"`
	BonusMatched bool  `json:"bonus_matched"`
	Rank         int   `json:"rank"` // 0=미당첨, 1~5등
}

func RegisterCheckNumbers(s *server.MCPServer, c *client.Client) {
	tool := mcp.NewTool("check_numbers",
		mcp.WithDescription("입력한 로또 번호가 특정 회차에서 몇 등에 해당하는지 확인합니다."),
		mcp.WithArray("numbers",
			mcp.Required(),
			mcp.Description("확인할 번호 세트 배열 (예: [[1,2,3,4,5,6],[7,8,9,10,11,12]])"),
		),
		mcp.WithNumber("round",
			mcp.Description("확인할 회차 번호 (미입력 시 최신 회차)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		// round 파라미터 처리
		var round int
		if roundVal, ok := args["round"]; ok && roundVal != nil {
			if roundFloat, ok := roundVal.(float64); ok {
				round = int(roundFloat)
			}
		}

		// numbers 파라미터 처리
		numbersRaw, ok := args["numbers"]
		if !ok {
			return mcp.NewToolResultError("numbers 파라미터가 필요합니다"), nil
		}

		numbersSlice, ok := numbersRaw.([]interface{})
		if !ok {
			return mcp.NewToolResultError("numbers는 배열이어야 합니다. 예: [[1,2,3,4,5,6]]"), nil
		}

		// 각 번호 세트 파싱
		var numberSets [][]int
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
					return mcp.NewToolResultError(fmt.Sprintf("번호는 1~45 사이여야 합니다. 입력값: %d", num)), nil
				}
				nums = append(nums, num)
			}

			if len(nums) != 6 {
				return mcp.NewToolResultError(fmt.Sprintf("%d번째 번호 세트는 정확히 6개 번호가 필요합니다 (현재: %d개)", i+1, len(nums))), nil
			}

			numberSets = append(numberSets, nums)
		}

		if len(numberSets) == 0 {
			return mcp.NewToolResultError("확인할 번호 세트가 없습니다"), nil
		}

		// 회차가 지정되지 않으면 최신 회차 사용
		var winNums *client.WinningNumbers
		var err error

		if round == 0 {
			latest, err := c.GetLatestRound()
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("최신 회차 조회 실패: %v", err)), nil
			}
			round = latest.LatestDrawnRound
		}

		winNums, err = c.GetWinningNumbers(round)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// 당첨번호를 map으로 변환 (빠른 조회)
		winSet := make(map[int]bool)
		for _, n := range winNums.Numbers {
			winSet[n] = true
		}

		// 각 번호 세트 확인
		var results []CheckResult
		for _, nums := range numberSets {
			matched := []int{}
			bonusMatched := false

			for _, n := range nums {
				if winSet[n] {
					matched = append(matched, n)
				}
				if n == winNums.Bonus {
					bonusMatched = true
				}
			}

			rank := calcRank(len(matched), bonusMatched)
			results = append(results, CheckResult{
				Numbers:      nums,
				Matched:      matched,
				BonusMatched: bonusMatched,
				Rank:         rank,
			})
		}

		output := map[string]interface{}{
			"round":   round,
			"winning": map[string]interface{}{"numbers": winNums.Numbers, "bonus": winNums.Bonus},
			"results": results,
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("결과 직렬화 실패: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	})
}

// calcRank는 매칭된 번호 수와 보너스 매칭 여부로 등수를 계산합니다.
// 0=미당첨, 1~5등
func calcRank(matchCount int, bonusMatched bool) int {
	switch matchCount {
	case 6:
		return 1
	case 5:
		if bonusMatched {
			return 2
		}
		return 3
	case 4:
		return 4
	case 3:
		return 5
	default:
		return 0
	}
}
