package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"github.com/Torres-09/dhlottery-mcp/internal/client"
	"github.com/Torres-09/dhlottery-mcp/internal/config"
	"github.com/Torres-09/dhlottery-mcp/internal/tools"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("dhlottery-mcp version %s\n", version)
		os.Exit(0)
	}

	cfg := config.Load()

	c, err := client.New(cfg.UserID, cfg.UserPW)
	if err != nil {
		fmt.Fprintf(os.Stderr, "클라이언트 초기화 실패: %v\n", err)
		os.Exit(1)
	}

	s := server.NewMCPServer(
		"dhlottery-mcp",
		version,
		server.WithToolCapabilities(true),
	)

	// MCP Tool 등록
	tools.RegisterGetWinningNumbers(s, c)
	tools.RegisterGetLatestRound(s, c)
	tools.RegisterCheckNumbers(s, c)
	tools.RegisterGetBalance(s, c, cfg)
	tools.RegisterBuyLotto(s, c, cfg)
	tools.RegisterGetPurchaseHistory(s, c, cfg)

	// stdio transport로 서버 시작
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "서버 오류: %v\n", err)
		os.Exit(1)
	}
}
