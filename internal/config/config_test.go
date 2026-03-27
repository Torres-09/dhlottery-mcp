package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	os.Setenv("DHLOTTERY_USER_ID", "testuser")
	os.Setenv("DHLOTTERY_USER_PW", "testpass")
	defer func() {
		os.Unsetenv("DHLOTTERY_USER_ID")
		os.Unsetenv("DHLOTTERY_USER_PW")
	}()

	cfg := Load()

	if cfg.UserID != "testuser" {
		t.Errorf("UserID = %q, want %q", cfg.UserID, "testuser")
	}
	if cfg.UserPW != "testpass" {
		t.Errorf("UserPW = %q, want %q", cfg.UserPW, "testpass")
	}
}

func TestHasCredentials(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		userPW string
		want   bool
	}{
		{"둘 다 설정됨", "user", "pass", true},
		{"ID만 설정됨", "user", "", false},
		{"PW만 설정됨", "", "pass", false},
		{"둘 다 미설정", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{UserID: tt.userID, UserPW: tt.userPW}
			if got := cfg.HasCredentials(); got != tt.want {
				t.Errorf("HasCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}
