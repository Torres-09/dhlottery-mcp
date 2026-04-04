// Package main은 동행복권 API 디버깅용 진단 도구입니다.
// 사용법: DHLOTTERY_USER_ID=<id> DHLOTTERY_USER_PW=<pw> go run ./cmd/diag/
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

const ua = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

func main() {
	userID := os.Getenv("DHLOTTERY_USER_ID")
	userPW := os.Getenv("DHLOTTERY_USER_PW")
	if userID == "" || userPW == "" {
		fmt.Fprintln(os.Stderr, "DHLOTTERY_USER_ID / DHLOTTERY_USER_PW 환경변수를 설정하세요")
		os.Exit(1)
	}

	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	hc := &http.Client{Jar: jar, Timeout: 30 * time.Second}

	// 1. 로그인
	fmt.Println("[1] 로그인 중...")
	if err := login(hc, userID, userPW); err != nil {
		fmt.Fprintf(os.Stderr, "로그인 실패: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("    완료")

	// 2. 잔액 조회
	fmt.Println("[2] 잔액 조회...")
	printJSON(get(hc, "https://www.dhlottery.co.kr/mypage/selectUserMndp.do",
		map[string]string{"X-Requested-With": "XMLHttpRequest", "AJAX": "true",
			"Referer": "https://www.dhlottery.co.kr/mypage/home"}))

	// 3. 구매내역 조회
	fmt.Println("[3] 구매내역 조회 (최근 30일)...")
	end := time.Now().Format("20060102")
	start := time.Now().AddDate(0, 0, -30).Format("20060102")
	params := url.Values{
		"srchStrDt": {start}, "srchEndDt": {end},
		"sort": {""}, "ltGdsCd": {""}, "winResult": {""}, "lramSmam": {""},
		"pageNum": {"1"}, "recordCountPerPage": {"10"},
	}
	printJSON(get(hc, "https://www.dhlottery.co.kr/mypage/selectMyLotteryledger.do?"+params.Encode(),
		map[string]string{
			"X-Requested-With": "XMLHttpRequest", "AJAX": "true",
			"Referer": "https://www.dhlottery.co.kr/mypage/mylotteryledger",
		}))
}

func login(hc *http.Client, userID, userPW string) error {
	r, _ := http.NewRequest(http.MethodGet, "https://www.dhlottery.co.kr/login", nil)
	r.Header.Set("User-Agent", ua)
	resp, err := hc.Do(r)
	if err != nil {
		return err
	}
	loginBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	reMod := regexp.MustCompile(`var\s+rsaModulus\s*=\s*'([0-9a-f]+)'`)
	reExp := regexp.MustCompile(`var\s+publicExponent\s*=\s*'([0-9a-f]+)'`)
	mMod := reMod.FindSubmatch(loginBody)
	mExp := reExp.FindSubmatch(loginBody)
	if len(mMod) < 2 || len(mExp) < 2 {
		return fmt.Errorf("RSA 키 파싱 실패")
	}
	modBytes, _ := hex.DecodeString(string(mMod[1]))
	expHex := string(mExp[1])
	if len(expHex)%2 != 0 {
		expHex = "0" + expHex
	}
	expBytes, _ := hex.DecodeString(expHex)
	pubKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(modBytes),
		E: int(new(big.Int).SetBytes(expBytes).Int64()),
	}
	enc := func(s string) string {
		b, _ := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(s))
		return hex.EncodeToString(b)
	}

	formData := url.Values{
		"inpUserId":    {userID},
		"userId":       {enc(userID)},
		"userPswdEncn": {enc(userPW)},
	}
	r, _ = http.NewRequest(http.MethodPost, "https://www.dhlottery.co.kr/login/securityLoginCheck.do",
		strings.NewReader(formData.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("User-Agent", ua)
	r.Header.Set("Referer", "https://www.dhlottery.co.kr/login")
	resp, err = hc.Do(r)
	if err != nil {
		return err
	}
	resp.Body.Close()
	fmt.Printf("    최종 URL: %s\n", resp.Request.URL)
	return nil
}

func get(hc *http.Client, apiURL string, headers map[string]string) string {
	req, _ := http.NewRequest(http.MethodGet, apiURL, nil)
	req.Header.Set("User-Agent", ua)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := hc.Do(req)
	if err != nil {
		return fmt.Sprintf("오류: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

func printJSON(s string) {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		fmt.Println("  ", s)
		return
	}
	b, _ := json.MarshalIndent(v, "    ", "  ")
	fmt.Println("   ", string(b))
}
