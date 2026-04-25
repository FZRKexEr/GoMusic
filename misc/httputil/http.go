package httputil

import (
	"log/slog"
	"net/http"
)

// not allow redirect client
var client *http.Client

// allow redirect client
var clientNoRedirect *http.Client

func init() {
	client = &http.Client{}
	clientNoRedirect = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // return this error to prevent redirect
		},
	}
}

// GetRedirectLocation ...
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
	req.Header.Set("User-Agent", "Mozilla/5.0")
	return req, nil
}
