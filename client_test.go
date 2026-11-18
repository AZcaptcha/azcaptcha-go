package azcaptcha_test

import (
	"context"
	"errors"
	"testing"

	azcaptcha "github.com/azcaptcha/azcaptcha-go"
)

func TestSubmitImageFixture(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{
		fixtureJSON(t, "submit-image"),
		{"status": float64(1), "request": "wxyz"},
	}}
	client, _ := newClient(t, transport, 0)
	solution, err := client.SolveImage(context.Background(), []byte("PNG"), nil)
	if err != nil || solution != "wxyz" {
		t.Fatalf("solve image = %q, %v", solution, err)
	}
	first := transport.calls[0]
	if first.kind != "form" || first.path != "/in.php" {
		t.Fatalf("call = %+v", first)
	}
	form := first.payload.(map[string]any)
	if form["method"] != "base64" || form["key"] != "test-key" || form["json"] != "1" {
		t.Fatalf("form = %+v", form)
	}
	if form["body"] != azcaptcha.ImageBody([]byte("PNG")) {
		t.Fatalf("body = %v", form["body"])
	}
}

func TestSubmitRecaptchaV2Fixture(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{fixtureJSON(t, "submit-recaptcha-v2")}}
	client, _ := newClient(t, transport, 0)
	id, err := client.Submit(context.Background(), "userrecaptcha", map[string]any{
		"googlekey": "6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-",
		"pageurl":   "https://example.com/login",
	})
	if err != nil || id != "1234567891" {
		t.Fatalf("submit = %q, %v", id, err)
	}
	if _, exists := transport.calls[0].payload.(map[string]any)["soft_id"]; exists {
		t.Fatal("zero soft id must be omitted")
	}
}

func TestSoftIDSent(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{{"status": float64(1), "request": "1"}}}
	client, _ := newClient(t, transport, 4242)
	if _, err := client.Submit(context.Background(), "base64", map[string]any{"body": "AA=="}); err != nil {
		t.Fatal(err)
	}
	if transport.calls[0].payload.(map[string]any)["soft_id"] != "4242" {
		t.Fatalf("payload = %+v", transport.calls[0].payload)
	}
}

func TestProxyNormalised(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{{"status": float64(1), "request": "1"}}}
	client, _ := newClient(t, transport, 0)
	_, err := client.Submit(context.Background(), "turnstile", map[string]any{
		"sitekey": "0x4AAA", "pageurl": "https://e.com",
		"proxy": "socks5://u:p@1.2.3.4:1080",
	})
	if err != nil {
		t.Fatal(err)
	}
	form := transport.calls[0].payload.(map[string]any)
	if form["proxy"] != "u:p@1.2.3.4:1080" || form["proxytype"] != "SOCKS5" {
		t.Fatalf("proxy fields = %+v", form)
	}
}

func TestZeroBalanceFixture(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{fixtureJSON(t, "error-zero-balance")}}
	client, _ := newClient(t, transport, 0)
	_, err := client.Submit(context.Background(), "userrecaptcha", map[string]any{
		"googlekey": "k", "pageurl": "https://e.com",
	})
	if !errors.Is(err, azcaptcha.ErrBalance) {
		t.Fatalf("err = %v", err)
	}
	var apiErr *azcaptcha.Error
	if !errors.As(err, &apiErr) || apiErr.Code != "ERROR_ZERO_BALANCE" || apiErr.Retryable {
		t.Fatalf("typed err = %+v", apiErr)
	}
}

func TestNoSlotFixture(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{fixtureJSON(t, "error-no-slot")}}
	client, _ := newClient(t, transport, 0)
	_, err := client.Submit(context.Background(), "turnstile", map[string]any{
		"sitekey": "k", "pageurl": "https://e.com",
	})
	if !errors.Is(err, azcaptcha.ErrCapacity) {
		t.Fatalf("err = %v", err)
	}
	var apiErr *azcaptcha.Error
	if !errors.As(err, &apiErr) || !apiErr.Retryable {
		t.Fatalf("retryable = %+v", apiErr)
	}
}

func TestResultFixtures(t *testing.T) {
	ready := &fakeTransport{responses: []map[string]any{fixtureJSON(t, "poll-ready-recaptcha")}}
	client, _ := newClient(t, ready, 0)
	if solution, err := client.Result(context.Background(), "1234567891"); err != nil || solution != "03AGdBq26..." {
		t.Fatalf("ready = %q, %v", solution, err)
	}

	processing := &fakeTransport{responses: []map[string]any{fixtureJSON(t, "poll-processing")}}
	client, _ = newClient(t, processing, 0)
	if _, err := client.Result(context.Background(), "1234567891"); !errors.Is(err, azcaptcha.ErrProcessing) {
		t.Fatalf("processing err = %v", err)
	}
}

func TestWaitResultDelays(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{
		fixtureJSON(t, "poll-processing"),
		fixtureJSON(t, "poll-ready-recaptcha"),
	}}
	client, sleeps := newClient(t, transport, 0)
	if solution, err := client.WaitResult(context.Background(), "1234567891", 0); err != nil || solution != "03AGdBq26..." {
		t.Fatalf("wait = %q, %v", solution, err)
	}
	if len(*sleeps) != 1 || (*sleeps)[0] != "1s" {
		t.Fatalf("sleeps = %v", *sleeps)
	}
}

func TestWaitResultBacksOffOnNoSlot(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{
		{"status": float64(0), "request": "ERROR_NO_SLOT_AVAILABLE"},
		{"status": float64(1), "request": "token"},
	}}
	client, sleeps := newClient(t, transport, 0)
	if _, err := client.WaitResult(context.Background(), "1", 0); err != nil {
		t.Fatal(err)
	}
	if len(*sleeps) != 1 || (*sleeps)[0] != "10s" {
		t.Fatalf("sleeps = %v", *sleeps)
	}
}

func TestWaitResultNonRetryable(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{
		{"status": float64(0), "request": "ERROR_CAPTCHA_UNSOLVABLE"},
	}}
	client, _ := newClient(t, transport, 0)
	if _, err := client.WaitResult(context.Background(), "1", 0); !errors.Is(err, azcaptcha.ErrResult) {
		t.Fatalf("err = %v", err)
	}
}

func TestWaitResultTimeout(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{
		{"status": float64(0), "request": "CAPCHA_NOT_READY"},
	}}
	client, _ := newClient(t, transport, 0)
	if _, err := client.WaitResult(context.Background(), "1", 1); !errors.Is(err, azcaptcha.ErrResult) {
		t.Fatalf("err = %v", err)
	}
}

func TestBalanceAndReportBad(t *testing.T) {
	transport := &fakeTransport{responses: []map[string]any{
		{"status": float64(1), "request": 12.5},
		{"status": float64(1), "request": "OK_REPORT_RECORDED"},
	}}
	client, _ := newClient(t, transport, 0)
	balance, err := client.Balance(context.Background())
	if err != nil || balance != 12.5 {
		t.Fatalf("balance = %v, %v", balance, err)
	}
	ok, err := client.ReportBad(context.Background(), "123")
	if err != nil || !ok {
		t.Fatalf("report = %v, %v", ok, err)
	}
}

func TestEmptyKeyRejected(t *testing.T) {
	if _, err := azcaptcha.New("", azcaptcha.Config{}); !errors.Is(err, azcaptcha.ErrAuth) {
		t.Fatalf("err = %v", err)
	}
}
