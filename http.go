package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultHTTPTimeout is the default timeout for HTTP requests (30 seconds)
const DefaultHTTPTimeout = 30 * time.Second

// httpGet performs an HTTP GET request with optional headers and timeout
func httpGet(url string, headers map[string]string, timeout time.Duration) (string, error) {
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

	resp, err := http.DefaultClient.Do(req)
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

// httpPost performs an HTTP POST request with body, optional headers and timeout
func httpPost(url string, body string, headers map[string]string, timeout time.Duration) (string, error) {
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

	resp, err := http.DefaultClient.Do(req)
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