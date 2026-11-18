package azcaptcha

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Transport is the minimal HTTP surface the client needs; tests provide a fake.
type Transport interface {
	PostForm(ctx context.Context, path string, form url.Values) (map[string]any, error)
	PostJSON(ctx context.Context, path string, payload any) (map[string]any, error)
	Get(ctx context.Context, path string, query url.Values) (map[string]any, error)
}

// HTTPTransport is the default transport.
type HTTPTransport struct {
	BaseURL string
	HTTP    *http.Client
}

func NewHTTPTransport(baseURL string, timeout time.Duration) *HTTPTransport {
	return &HTTPTransport{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP:    &http.Client{Timeout: timeout},
	}
}

func (t *HTTPTransport) PostForm(ctx context.Context, path string, form url.Values) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.BaseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, networkError(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return t.do(req)
}

func (t *HTTPTransport) PostJSON(ctx context.Context, path string, payload any) (map[string]any, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, ErrFor("ERROR_BAD_PARAMETERS")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, networkError(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return t.do(req)
}

func (t *HTTPTransport) Get(ctx context.Context, path string, query url.Values) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.BaseURL+path+"?"+query.Encode(), nil)
	if err != nil {
		return nil, networkError(err)
	}
	return t.do(req)
}

func (t *HTTPTransport) do(req *http.Request) (map[string]any, error) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "azcaptcha-go/1.0.0")
	res, err := t.HTTP.Do(req)
	if err != nil {
		return nil, networkError(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, networkError(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, &Error{Code: "ERROR_BAD_RESPONSE", Message: string(body), Err: ErrService}
	}
	if res.StatusCode >= 500 {
		return nil, &Error{Code: "ERROR_HTTP_5XX", Message: res.Status, Err: ErrService}
	}
	return payload, nil
}

func networkError(err error) error {
	return &Error{Code: "ERROR_NETWORK", Retryable: true, Message: err.Error(), Err: ErrNetwork}
}
