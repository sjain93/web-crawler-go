package util

import (
	"errors"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultCxnTime = 60 * time.Second
)

var ErrUtilInvalidHost = errors.New("hostname is invalid, must be of http or https scheme")

func NewDefaultHTTPClient() *http.Client {
	return NewHTTPClient(defaultCxnTime)
}

func NewHTTPClient(customDuration time.Duration) *http.Client {
	return &http.Client{
		Timeout: customDuration,
	}
}

func GetHost(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if u.Hostname() == "" {
		return "", ErrUtilInvalidHost
	}

	return u.Hostname(), nil
}

func IsHTTPScheme(urlStr string) bool {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	return (parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https")
}

func GetAbsoluteURL(urlStr, rootUrl string) (string, error) {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	if parsedUrl.IsAbs() {
		return parsedUrl.String(), nil
	}

	root, err := url.Parse(rootUrl)
	if err != nil {
		return "", err
	}

	// ensure a referential URL is transformed into an absolute URL
	return root.ResolveReference(parsedUrl).String(), nil
}

func IsSameDomain(link, rootURL string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}

	if !u.IsAbs() {
		// there is an assumption that reference links are passed in from the
		// preceding page
		return true
	}

	domain, err := GetHost(rootURL)
	if err != nil {
		return false
	}

	return u.Hostname() == domain
}
