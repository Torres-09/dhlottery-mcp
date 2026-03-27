package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/imsehwan/dhlottery-mcp/internal/client"
)

func RegisterGetLatestRound(s *server.MCPServer, c *client.Client) {
	tool := mcp.NewTool("get_latest_round",
		mcp.WithDescription("가장 최근 추첨된 회차 번호와 다음 구매 가능 회차를 조회합니다."),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := c.GetLatestRound()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		output := map[string]interface{}{
			"latest_drawn_round": result.LatestDrawnRound,
			"next_round":         result.NextRound,
			"latest_draw_date":   result.LatestDrawDate,
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("결과 직렬화 실패: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	})
}
