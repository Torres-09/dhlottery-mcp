package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/imsehwan/dhlottery-mcp/internal/client"
	"github.com/imsehwan/dhlottery-mcp/internal/config"
)

func RegisterGetPurchaseHistory(s *server.MCPServer, c *client.Client, cfg *config.Config) {
	tool := mcp.NewTool("get_purchase_history",
		mcp.WithDescription("로또 구매 내역을 조회합니다. 로그인이 필요합니다. 인터넷(PC/웹) 구매 내역만 조회되며, 모바일 앱 구매 내역은 포함되지 않습니다. 최대 조회 기간은 31일입니다."),
		mcp.WithString("start_date",
			mcp.Description("조회 시작일 (YYYY-MM-DD 형식, 기본: 30일 전)"),
		),
		mcp.WithString("end_date",
			mcp.Description("조회 종료일 (YYYY-MM-DD 형식, 기본: 오늘)"),
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
		startDate, _ := args["start_date"].(string)
		endDate, _ := args["end_date"].(string)

		purchases, err := c.GetPurchaseHistory(startDate, endDate)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if purchases == nil {
			purchases = []client.Purchase{}
		}
		output := map[string]interface{}{
			"purchases": purchases,
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("결과 직렬화 실패: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	})
}
