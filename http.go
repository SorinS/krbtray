package main

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

// DefaultHTTPTimeout is the default timeout for HTTP requests (30 seconds)
const DefaultHTTPTimeout = 30 * time.Second

// insecureClient is an HTTP client that skips TLS certificate verification
var insecureClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

// HTTPSession maintains cookies across multiple HTTP requests
type HTTPSession struct {
	jar        *cookiejar.Jar
	client     *http.Client
	skipVerify bool
}

// NewHTTPSession creates a new HTTP session with cookie jar support
func NewHTTPSession(skipVerify bool) (*HTTPSession, error) {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, err
	}

	transport := http.DefaultTransport
	if skipVerify {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client := &http.Client{
		Jar:       jar,
		Transport: transport,
	}

	return &HTTPSession{
		jar:        jar,
		client:     client,
		skipVerify: skipVerify,
	}, nil
}

// Get performs an HTTP GET request using the session's cookie jar
func (s *HTTPSession) Get(url string, headers map[string]string, timeout time.Duration) (string, error) {
	if timeout <= 0 {
		timeout = DefaultHTTPTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Post performs an HTTP POST request using the session's cookie jar
func (s *HTTPSession) Post(url string, body string, headers map[string]string, timeout time.Duration) (string, error) {
	if timeout <= 0 {
		timeout = DefaultHTTPTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if err != nil {
		return "", err
	}

	// Default content type
	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}

// httpGet performs an HTTP GET request with optional headers, timeout, and skip_verify
func httpGet(url string, headers map[string]string, timeout time.Duration, skipVerify bool) (string, error) {
	if timeout <= 0 {
		timeout = DefaultHTTPTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := http.DefaultClient
	if skipVerify {
		client = insecureClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// httpPost performs an HTTP POST request with body, optional headers, timeout, and skip_verify
func httpPost(url string, body string, headers map[string]string, timeout time.Duration, skipVerify bool) (string, error) {
	if timeout <= 0 {
		timeout = DefaultHTTPTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if err != nil {
		return "", err
	}

	// Default content type
	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := http.DefaultClient
	if skipVerify {
		client = insecureClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}