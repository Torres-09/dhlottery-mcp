package client

import (
	"testing"
)

func TestLogin_NoCredentials(t *testing.T) {
	c, err := New("", "")
	if err != nil {
		t.Fatalf("클라이언트 생성 실패: %v", err)
	}

	err = c.Login()
	if err == nil {
		t.Error("자격증명 없이 로그인 시 에러가 반환되어야 합니다")
	}

	expected := "로그인 자격증명이 설정되지 않았습니다"
	if len(err.Error()) < len(expected) || err.Error()[:len(expected)] != expected {
		t.Errorf("에러 메시지 = %q, want prefix %q", err.Error(), expected)
	}
}

func TestEnsureLoggedIn_AlreadyLoggedIn(t *testing.T) {
	c, err := New("user", "pass")
	if err != nil {
		t.Fatalf("클라이언트 생성 실패: %v", err)
	}

	// 이미 로그인된 상태로 설정
	c.loggedIn = true

	// EnsureLoggedIn은 이미 로그인되어 있으면 아무 작업 없이 nil 반환
	err = c.EnsureLoggedIn()
	if err != nil {
		t.Errorf("이미 로그인된 상태에서 에러가 발생하면 안 됩니다: %v", err)
	}
}

func TestInvalidateSession(t *testing.T) {
	c, err := New("user", "pass")
	if err != nil {
		t.Fatalf("클라이언트 생성 실패: %v", err)
	}

	c.loggedIn = true
	c.InvalidateSession()

	if c.loggedIn {
		t.Error("InvalidateSession 후 loggedIn이 false여야 합니다")
	}
}
