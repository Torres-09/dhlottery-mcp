package client

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/publicsuffix"
)

const (
	wwwBaseURL = "https://www.dhlottery.co.kr"
)

// EnsureLoggedIn은 로그인이 필요한 경우 로그인을 수행합니다.
func (c *Client) EnsureLoggedIn() error {
	if c.loggedIn {
		return nil
	}
	return c.Login()
}

// Login은 동행복권에 RSA 암호화 방식으로 로그인합니다.
func (c *Client) Login() error {
	if c.userID == "" || c.userPW == "" {
		return fmt.Errorf("로그인 자격증명이 설정되지 않았습니다. DHLOTTERY_USER_ID와 DHLOTTERY_USER_PW 환경변수를 설정해주세요")
	}

	// 재로그인 시 기존 쿠키를 초기화해 서버가 로그인 폼을 반환하도록 보장
	// 구 세션 쿠키가 남아있으면 /login GET이 홈으로 리다이렉트되어 rsaModulus를 찾지 못함
	freshJar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return fmt.Errorf("쿠키 초기화 실패: %w", err)
	}
	c.http.Jar = freshJar

	// 1단계: www 도메인 로그인 페이지로 세션 쿠키 확보
	// dhlottery.co.kr → www.dhlottery.co.kr 301 리다이렉트를 피하기 위해 www 직접 사용
	req, err := http.NewRequest(http.MethodGet, wwwBaseURL+"/login", nil)
	if err != nil {
		return fmt.Errorf("로그인 페이지 요청 생성 실패: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("로그인 페이지 요청 실패: %w", err)
	}
	loginPageBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("로그인 페이지 읽기 실패: %w", err)
	}

	// 2단계: 로그인 페이지 HTML에서 RSA 공개키 추출 (세션-키 바인딩 보장)
	pubKey, err := parseRSAKeyFromHTML(loginPageBody)
	if err != nil {
		return fmt.Errorf("RSA 공개키 파싱 실패: %w", err)
	}

	// 3단계: userId, password RSA 암호화 (hex 인코딩)
	encUserID, err := rsaEncrypt(pubKey, c.userID)
	if err != nil {
		return fmt.Errorf("아이디 암호화 실패: %w", err)
	}
	encUserPW, err := rsaEncrypt(pubKey, c.userPW)
	if err != nil {
		return fmt.Errorf("비밀번호 암호화 실패: %w", err)
	}

	// 4단계: 로그인 POST (www 도메인 직접)
	formData := url.Values{
		"inpUserId":    {c.userID},
		"userId":       {encUserID},
		"userPswdEncn": {encUserPW},
	}

	req, err = http.NewRequest(http.MethodPost, wwwBaseURL+"/login/securityLoginCheck.do",
		strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("로그인 요청 생성 실패: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", wwwBaseURL)
	req.Header.Set("Referer", wwwBaseURL+"/login")

	resp, err = c.http.Do(req)
	if err != nil {
		return fmt.Errorf("로그인 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("로그인 응답 읽기 실패: %w", err)
	}
	bodyStr := string(body)

	// 로그인 실패 감지: 에러 메시지가 body에 포함되면 실패
	if strings.Contains(bodyStr, "아이디 또는 비밀번호가 일치하지 않습니다") ||
		strings.Contains(bodyStr, "errorMessage = '아이디") {
		return fmt.Errorf("로그인에 실패했습니다. 아이디와 비밀번호를 확인해주세요")
	}

	// 계정 잠금 감지
	if strings.Contains(bodyStr, "잠금") || strings.Contains(bodyStr, "locked") {
		return fmt.Errorf("계정이 잠금 상태입니다. 동행복권 사이트에서 계정 상태를 확인해주세요")
	}

	c.loggedIn = true
	return nil
}

// parseRSAKeyFromHTML은 로그인 페이지 HTML에서 RSA 공개키를 추출합니다.
// 로그인 페이지에 세션-바인딩된 키가 var rsaModulus / var publicExponent 로 포함됩니다.
func parseRSAKeyFromHTML(body []byte) (*rsa.PublicKey, error) {
	reModulus := regexp.MustCompile(`var\s+rsaModulus\s*=\s*'([0-9a-f]+)'`)
	reExp := regexp.MustCompile(`var\s+publicExponent\s*=\s*'([0-9a-f]+)'`)

	mMod := reModulus.FindSubmatch(body)
	if len(mMod) < 2 {
		return nil, fmt.Errorf("로그인 페이지에서 RSA modulus를 찾을 수 없습니다")
	}
	mExp := reExp.FindSubmatch(body)
	if len(mExp) < 2 {
		return nil, fmt.Errorf("로그인 페이지에서 RSA exponent를 찾을 수 없습니다")
	}

	modulusBytes, err := hex.DecodeString(string(mMod[1]))
	if err != nil {
		return nil, fmt.Errorf("RSA modulus 파싱 실패: %w", err)
	}

	expHex := string(mExp[1])
	if len(expHex)%2 != 0 {
		expHex = "0" + expHex
	}
	expBytes, err := hex.DecodeString(expHex)
	if err != nil {
		return nil, fmt.Errorf("RSA exponent 파싱 실패: %w", err)
	}

	n := new(big.Int).SetBytes(modulusBytes)
	e := int(new(big.Int).SetBytes(expBytes).Int64())

	return &rsa.PublicKey{N: n, E: e}, nil
}

// rsaEncrypt는 JSEncrypt 방식(PKCS1v15)으로 문자열을 암호화하여 hex 문자열로 반환합니다.
func rsaEncrypt(pubKey *rsa.PublicKey, plaintext string) (string, error) {
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(plaintext))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(encrypted), nil
}

// InvalidateSession은 세션을 초기화합니다 (재로그인 유도).
func (c *Client) InvalidateSession() {
	c.loggedIn = false
}
