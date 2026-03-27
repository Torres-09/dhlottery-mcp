package client

import (
	"net/http"
	"net/http/cookiejar"
	"time"

	"golang.org/x/net/publicsuffix"
)

const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

	baseURL = "https://dhlottery.co.kr"
	buyURL  = "https://ol.dhlottery.co.kr/olotto/game/execBuy.do"
)

// Client는 동행복권 HTTP 클라이언트입니다.
type Client struct {
	http       *http.Client
	loggedIn   bool
	userID     string
	userPW     string
}

// New는 새로운 Client를 생성합니다.
func New(userID, userPW string) (*Client, error) {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	return &Client{
		http:   httpClient,
		userID: userID,
		userPW: userPW,
	}, nil
}

