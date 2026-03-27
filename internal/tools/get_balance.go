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

func RegisterGetBalance(s *server.MCPServer, c *client.Client, cfg *config.Config) {
	tool := mcp.NewTool("get_balance",
		mcp.WithDescription("동행복권 계정의 예치금 잔액을 조회합니다. 로그인이 필요합니다."),
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

		result, err := c.GetBalance()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		output := map[string]interface{}{
			"balance": result.Amount,
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("결과 직렬화 실패: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	})
}
