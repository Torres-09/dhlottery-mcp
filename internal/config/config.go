package config

import "os"

// Config는 동행복권 MCP 서버 설정을 담습니다.
type Config struct {
	UserID string
	UserPW string
}

// Load는 환경변수에서 설정을 로드합니다.
func Load() *Config {
	return &Config{
		UserID: os.Getenv("DHLOTTERY_USER_ID"),
		UserPW: os.Getenv("DHLOTTERY_USER_PW"),
	}
}

// HasCredentials는 로그인 자격증명이 설정되어 있는지 확인합니다.
func (c *Config) HasCredentials() bool {
	return c.UserID != "" && c.UserPW != ""
}
