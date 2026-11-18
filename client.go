package azcaptcha

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

const (
	// NotReady is part of the wire contract — spelling included.
	NotReady = "CAPCHA_NOT_READY"
	// SoftID is sent for upstream attribution once assigned; zero omits it.
	SoftID = 0

	pollInterval  = 5 * time.Second
	firstDelay    = time.Second
	noSlotBackoff = 10 * time.Second
)

// Config overrides defaults; the zero value is production-ready.
type Config struct {
	Host        string
	HTTPTimeout time.Duration
	SoftID      int
	Transport   Transport
	Sleep       func(time.Duration)
}

// Client talks to the AZcaptcha API. It is safe for concurrent use.
type Client struct {
	apiKey string
	softID int
	http   Transport
	sleep  func(time.Duration)
}

// New validates the key and wires defaults.
func New(apiKey string, cfg Config) (*Client, error) {
	if apiKey == "" {
		return nil, ErrFor("ERROR_KEY_DOES_NOT_EXIST")
	}
	host := cfg.Host
	if host == "" {
		host = "https://azcaptcha.com"
	}
	timeout := cfg.HTTPTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	transport := cfg.Transport
	if transport == nil {
		transport = NewHTTPTransport(host, timeout)
	}
	sleep := cfg.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	return &Client{apiKey: apiKey, softID: cfg.SoftID, http: transport, sleep: sleep}, nil
}

// Submit creates a legacy task and returns its id.
func (c *Client) Submit(ctx context.Context, method string, params map[string]any) (string, error) {
	form := url.Values{}
	form.Set("key", c.apiKey)
	form.Set("method", method)
	form.Set("json", "1")
	if c.softID > 0 {
		form.Set("soft_id", strconv.Itoa(c.softID))
	}
	if proxy, ok := params["proxy"].(string); ok && proxy != "" {
		normalized, proxyType, err := NormalizeProxy(proxy)
		if err != nil {
			return "", ErrFor("ERROR_BAD_PARAMETERS")
		}
		form.Set("proxy", normalized)
		form.Set("proxytype", proxyType)
	}
	if proxyType, ok := params["proxytype"].(string); ok && proxyType != "" {
		form.Set("proxytype", proxyType)
	}
	for key, value := range params {
		if value == nil || key == "proxy" || key == "proxytype" {
			continue
		}
		form.Set(key, fmt.Sprint(value))
	}
	body, err := c.http.PostForm(ctx, "/in.php", form)
	if err != nil {
		return "", err
	}
	return unwrap(body)
}

// Result fetches a solution once; ErrProcessing means poll again.
func (c *Client) Result(ctx context.Context, taskID string) (string, error) {
	body, err := c.http.Get(ctx, "/res.php", url.Values{
		"key": {c.apiKey}, "action": {"get"}, "id": {taskID}, "json": {"1"},
	})
	if err != nil {
		return "", err
	}
	if request, _ := body["request"].(string); request == NotReady {
		return "", ErrFor(NotReady)
	}
	return unwrap(body)
}

// WaitResult polls with the contract delays until ready or timeout.
func (c *Client) WaitResult(ctx context.Context, taskID string, timeout time.Duration) (string, error) {
	if timeout == 0 {
		timeout = 300 * time.Second
	}
	deadline := time.Now().Add(timeout)
	delay, backoff := firstDelay, noSlotBackoff
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		solution, err := c.Result(ctx, taskID)
		if err == nil {
			return solution, nil
		}
		apiErr, ok := err.(*Error)
		if !ok {
			return "", err
		}
		var pause time.Duration
		switch {
		case apiErr.Code == NotReady:
			pause, delay = delay, pollInterval
		case apiErr.Err == ErrCapacity || (apiErr.Err == ErrLimit && apiErr.Retryable):
			pause, backoff = backoff, time.Duration(float64(backoff)*1.5)
			if backoff > time.Minute {
				backoff = time.Minute
			}
		default:
			return "", err
		}
		if time.Now().Add(pause).After(deadline) {
			return "", ErrFor("ERROR_TIMEOUT")
		}
		c.sleep(pause)
	}
}

// Balance returns the account balance in USD.
func (c *Client) Balance(ctx context.Context) (float64, error) {
	body, err := c.http.Get(ctx, "/res.php", url.Values{
		"key": {c.apiKey}, "action": {"getbalance"}, "json": {"1"},
	})
	if err != nil {
		return 0, err
	}
	value, err := unwrap(body)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(value, 64)
}

// ReportBad flags a solution as incorrect.
func (c *Client) ReportBad(ctx context.Context, taskID string) (bool, error) {
	body, err := c.http.Get(ctx, "/res.php", url.Values{
		"key": {c.apiKey}, "action": {"reportbad"}, "id": {taskID}, "json": {"1"},
	})
	if err != nil {
		return false, err
	}
	value, err := unwrap(body)
	if err != nil {
		return false, err
	}
	return len(value) >= 2 && (value[:2] == "OK" || value[:2] == "ok"), nil
}

func unwrap(body map[string]any) (string, error) {
	request, _ := body["request"].(string)
	if len(request) >= 6 && request[:6] == "ERROR_" {
		return "", ErrFor(request)
	}
	if status, _ := body["status"].(float64); status != 1 {
		if request != "" {
			return "", ErrFor(request)
		}
		return "", ErrFor("ERROR_REQUEST_FAILED")
	}
	value, ok := body["request"]
	if !ok || value == nil {
		return "", nil
	}
	return fmt.Sprint(value), nil
}
