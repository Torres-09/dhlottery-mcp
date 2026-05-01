package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"golang.org/x/term"
)

func runSetup() {
	fmt.Println()
	fmt.Println("dhlottery-mcp 설정을 시작합니다.")
	fmt.Println("Claude Code, Claude Desktop, Codex에 자동으로 등록합니다.")
	fmt.Println()

	// 현재 바이너리 절대 경로 획득
	exePath, err := os.Executable()
	if err != nil {
		die("바이너리 경로를 확인할 수 없습니다: " + err.Error())
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		die("바이너리 경로 확인 실패: " + err.Error())
	}

	// --id/--pw 플래그 파싱 (install.sh 등 자동화 경로에서 전달)
	userID := ""
	userPW := ""
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--id":
			if i+1 < len(args) {
				userID = args[i+1]
				i++
			}
		case "--pw":
			if i+1 < len(args) {
				userPW = args[i+1]
				i++
			}
		}
	}

	// 플래그로 제공되지 않은 항목만 대화형으로 입력
	reader := bufio.NewReader(os.Stdin)
	if userID == "" || userPW == "" {
		fmt.Println("동행복권 계정 정보를 입력하면 구매·잔액 조회 기능을 사용할 수 있습니다.")
		fmt.Println("(건너뛰려면 Enter를 누르세요)")
		fmt.Println()
	}

	if userID == "" {
		fmt.Print("  아이디: ")
		userID, _ = reader.ReadString('\n')
		userID = strings.TrimSpace(userID)
	}

	if userPW == "" {
		fmt.Print("  비밀번호: ")
		pwBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			// ReadPassword 실패 시 일반 입력으로 fallback
			fmt.Print("")
			pw, _ := reader.ReadString('\n')
			pwBytes = []byte(strings.TrimSpace(pw))
		}
		userPW = string(pwBytes)
		fmt.Println()
		fmt.Println()
	}

	// 각 클라이언트 등록
	ok := false

	if err := setupClaudeCode(exePath, userID, userPW); err != nil {
		printFail("Claude Code", err)
	} else {
		printOK("Claude Code", "~/.claude.json")
		ok = true
	}

	if err := setupClaudeDesktop(exePath, userID, userPW); err != nil {
		printSkip("Claude Desktop", err.Error())
	} else {
		printOK("Claude Desktop", desktopConfigPath())
		ok = true
	}

	if err := setupCodex(exePath, userID, userPW); err != nil {
		printSkip("Codex", err.Error())
	} else {
		printOK("Codex (CLI·Desktop·IDE)", "~/.codex/config.toml")
		ok = true
	}

	fmt.Println()
	if ok {
		fmt.Println("설정이 완료되었습니다.")
		fmt.Println()
		fmt.Println("  Claude Code / Codex : 새 세션을 열면 적용됩니다.")
		fmt.Println("  Claude Desktop       : 완전 종료 후 재시작하면 적용됩니다.")
	} else {
		fmt.Fprintln(os.Stderr, "등록 가능한 클라이언트를 찾지 못했습니다.")
		fmt.Fprintln(os.Stderr, "Claude Code, Claude Desktop, Codex 중 하나가 설치되어 있어야 합니다.")
		os.Exit(1)
	}
	fmt.Println()
}

// ── Claude Code ──────────────────────────────────────────────────────────────

func setupClaudeCode(exePath, userID, userPW string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".claude.json")

	data := map[string]interface{}{}
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &data)
	}

	servers, _ := data["mcpServers"].(map[string]interface{})
	if servers == nil {
		servers = map[string]interface{}{}
		data["mcpServers"] = servers
	}

	servers["dhlottery-mcp"] = map[string]interface{}{
		"type":    "stdio",
		"command": exePath,
		"args":    []string{},
		"env": map[string]string{
			"DHLOTTERY_USER_ID": userID,
			"DHLOTTERY_USER_PW": userPW,
		},
	}

	return writeJSON(path, data)
}

// ── Claude Desktop ────────────────────────────────────────────────────────────

func desktopConfigPath() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "Claude", "claude_desktop_config.json")
	default:
		return filepath.Join(home, ".config", "Claude", "claude_desktop_config.json")
	}
}

func setupClaudeDesktop(exePath, userID, userPW string) error {
	path := desktopConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("설정 파일 없음 (Claude Desktop 미설치)")
	}

	data := map[string]interface{}{}
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &data)
	}

	servers, _ := data["mcpServers"].(map[string]interface{})
	if servers == nil {
		servers = map[string]interface{}{}
		data["mcpServers"] = servers
	}

	servers["dhlottery"] = map[string]interface{}{
		"command": exePath,
		"env": map[string]string{
			"DHLOTTERY_USER_ID": userID,
			"DHLOTTERY_USER_PW": userPW,
		},
	}

	return writeJSON(path, data)
}

// ── Codex ─────────────────────────────────────────────────────────────────────

type codexConfig struct {
	MCPServers map[string]codexMCPServer `toml:"mcp_servers"`
	Rest       map[string]interface{}    `toml:"-"`
}

type codexMCPServer struct {
	Command string            `toml:"command"`
	Args    []string          `toml:"args,omitempty"`
	Env     map[string]string `toml:"env,omitempty"`
}

func setupCodex(exePath, userID, userPW string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".codex")
	path := filepath.Join(dir, "config.toml")

	// 기존 TOML을 raw map으로 읽어 기존 설정 보존
	raw := map[string]interface{}{}
	if b, err := os.ReadFile(path); err == nil {
		if err := toml.Unmarshal(b, &raw); err != nil {
			return fmt.Errorf("config.toml 파싱 실패: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	// mcp_servers 섹션 upsert
	mcpServers, _ := raw["mcp_servers"].(map[string]interface{})
	if mcpServers == nil {
		mcpServers = map[string]interface{}{}
		raw["mcp_servers"] = mcpServers
	}

	entry := map[string]interface{}{
		"command": exePath,
		"env": map[string]interface{}{
			"DHLOTTERY_USER_ID": userID,
			"DHLOTTERY_USER_PW": userPW,
		},
	}
	mcpServers["dhlottery-mcp"] = entry

	// Codex 디렉토리 없으면 생성
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(raw)
}

// ── 공통 유틸 ──────────────────────────────────────────────────────────────────

func writeJSON(path string, data interface{}) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}

func printOK(label, path string) {
	fmt.Printf("  ✓ %-30s (%s)\n", label+" 등록 완료", path)
}

func printFail(label string, err error) {
	fmt.Fprintf(os.Stderr, "  ✗ %s 등록 실패: %v\n", label, err)
}

func printSkip(label, reason string) {
	fmt.Printf("  - %-30s (%s)\n", label+" 건너뜀", reason)
}

func die(msg string) {
	fmt.Fprintln(os.Stderr, "오류: "+msg)
	os.Exit(1)
}

// ── unsetup ───────────────────────────────────────────────────────────────────

func runUnsetup() {
	fmt.Println()
	fmt.Println("dhlottery-mcp 설정을 제거합니다.")
	fmt.Println()

	ok := false

	if err := unsetupClaudeCode(); err != nil {
		printSkip("Claude Code", err.Error())
	} else {
		printOK("Claude Code", "~/.claude.json")
		ok = true
	}

	if err := unsetupClaudeDesktop(); err != nil {
		printSkip("Claude Desktop", err.Error())
	} else {
		printOK("Claude Desktop", desktopConfigPath())
		ok = true
	}

	if err := unsetupCodex(); err != nil {
		printSkip("Codex", err.Error())
	} else {
		printOK("Codex (CLI·Desktop·IDE)", "~/.codex/config.toml")
		ok = true
	}

	fmt.Println()
	if ok {
		fmt.Println("설정 제거가 완료되었습니다.")
	} else {
		fmt.Fprintln(os.Stderr, "제거할 설정을 찾지 못했습니다.")
		os.Exit(1)
	}
	fmt.Println()
}

func unsetupClaudeCode() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".claude.json")

	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	servers, _ := data["mcpServers"].(map[string]interface{})
	if servers == nil {
		return fmt.Errorf("mcpServers 항목 없음")
	}
	if _, ok := servers["dhlottery-mcp"]; !ok {
		return fmt.Errorf("dhlottery-mcp 항목 없음")
	}

	delete(servers, "dhlottery-mcp")
	return writeJSON(path, data)
}

func unsetupClaudeDesktop() error {
	path := desktopConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("설정 파일 없음 (Claude Desktop 미설치)")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	servers, _ := data["mcpServers"].(map[string]interface{})
	if servers == nil {
		return fmt.Errorf("mcpServers 항목 없음")
	}
	if _, ok := servers["dhlottery"]; !ok {
		return fmt.Errorf("dhlottery 항목 없음")
	}

	delete(servers, "dhlottery")
	return writeJSON(path, data)
}

func unsetupCodex() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".codex", "config.toml")

	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config.toml 없음 (Codex 미설치)")
		}
		return err
	}

	raw := map[string]interface{}{}
	if err := toml.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("config.toml 파싱 실패: %w", err)
	}

	mcpServers, _ := raw["mcp_servers"].(map[string]interface{})
	if mcpServers == nil {
		return fmt.Errorf("mcp_servers 항목 없음")
	}
	if _, ok := mcpServers["dhlottery-mcp"]; !ok {
		return fmt.Errorf("dhlottery-mcp 항목 없음")
	}

	delete(mcpServers, "dhlottery-mcp")

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(raw)
}
