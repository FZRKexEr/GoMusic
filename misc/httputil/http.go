package httputil

import (
	"log/slog"
	"net/http"
	"time"
)

const (
	defaultTimeout   = 10 * time.Second
	defaultUserAgent = "Mozilla/5.0"
)

var client *http.Client

var clientNoRedirect *http.Client

func init() {
	client = &http.Client{Timeout: defaultTimeout}
	clientNoRedirect = &http.Client{
		Timeout: defaultTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func GetRedirectLocation(link string) (string, error) {
	req, err := newGetRequest(link)
	if err != nil {
		return "", err
	}
	rsp, err := clientNoRedirect.Do(req)
	if err != nil {
		slog.Error("http Get error", "err", err)
		return "", err
	}
	defer rsp.Body.Close()
	return rsp.Header.Get("Location"), nil
}

func Get(link string) (*http.Response, error) {
	req, err := newGetRequest(link)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func newGetRequest(link string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		slog.Error("http NewRequest error", "err", err)
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	return req, nil
}
