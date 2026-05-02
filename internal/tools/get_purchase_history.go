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

func RegisterGetPurchaseHistory(s *server.MCPServer, c *client.Client, cfg *config.Config) {
	tool := mcp.NewTool("get_purchase_history",
		mcp.WithDescription("로또 구매/당첨 내역을 조회합니다. 로그인이 필요합니다. 인터넷(PC/웹) 구매 내역만 조회되며, 모바일 앱 구매 내역은 포함되지 않습니다. 최대 조회 기간은 31일입니다. 미추첨(추첨 대기) 내역도 함께 조회됩니다. 날짜 범위 지정 시 주의: 로또는 토요일 추첨이므로 '이번 주 구매'를 조회할 때 월요일 기준으로 잡으면 지난 토/일 구매분이 누락될 수 있습니다. 기간이 불명확하면 start_date를 7일 전으로 설정하거나 생략(기본 30일)하세요."),
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
