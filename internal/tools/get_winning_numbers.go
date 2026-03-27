package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/imsehwan/dhlottery-mcp/internal/client"
)

func RegisterGetWinningNumbers(s *server.MCPServer, c *client.Client) {
	tool := mcp.NewTool("get_winning_numbers",
		mcp.WithDescription("특정 회차의 로또 6/45 당첨번호를 조회합니다."),
		mcp.WithNumber("round",
			mcp.Required(),
			mcp.Description("조회할 회차 번호 (예: 1150)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		roundFloat, ok := args["round"].(float64)
		if !ok {
			return mcp.NewToolResultError("round 파라미터가 올바르지 않습니다. 숫자를 입력해주세요"), nil
		}
		round := int(roundFloat)

		if round <= 0 {
			return mcp.NewToolResultError("회차 번호는 1 이상이어야 합니다"), nil
		}

		result, err := c.GetWinningNumbers(round)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		output := map[string]interface{}{
			"round":         result.Round,
			"date":          result.Date,
			"numbers":       result.Numbers,
			"bonus":         result.Bonus,
			"total_sales":   result.TotalSales,
			"first_prize":   result.FirstPrize,
			"first_winners": result.FirstWinners,
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("결과 직렬화 실패: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	})
}
