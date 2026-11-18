# AZcaptcha Go SDK

Official Go client for [AZcaptcha](https://azcaptcha.com) — image captchas,
reCAPTCHA v2/v3, hCaptcha, Cloudflare Turnstile, FunCaptcha and GeeTest.

- Idiomatic Go: `context.Context` on every call, typed errors with `errors.Is`
  sentinels, `Config` overrides, no required options.
- 2Captcha-compatible surface (`Solve`, `Result`, `Balance`, `ReportBad`) plus the
  modern JSON API (`CreateTask`, `GetTaskResult`).
- Typed errors for every API code
  ([spec-v1.0.0](https://github.com/AZcaptcha/sdk-spec)).
- Contract polling with `NO_SLOT` backoff and per-type deadlines.
- Zero dependencies beyond the standard library.

## Install

```bash
go get github.com/azcaptcha/azcaptcha-go
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"

	azcaptcha "github.com/azcaptcha/azcaptcha-go"
)

func main() {
	client, err := azcaptcha.New("YOUR_API_KEY", azcaptcha.Config{})
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	token, err := client.SolveRecaptchaV2(ctx,
		"6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-",
		"https://example.com/login", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(token)

	answer, err := client.SolveImageFile(ctx, "captcha.png", nil)
	fmt.Println(answer, err)
}
```

### Low-level (2Captcha style)

```go
taskID, err := client.Submit(ctx, "userrecaptcha", map[string]any{
	"googlekey": "SITEKEY",
	"pageurl":   "https://example.com/",
})
token, err := client.WaitResult(ctx, taskID, 300*time.Second)
```

### Modern JSON API

```go
taskID, err := client.CreateTask(ctx, "RecaptchaV2TaskProxyless", map[string]any{
	"websiteURL": "https://example.com/login",
	"websiteKey": "SITEKEY",
})
solution, err := client.WaitTaskResult(ctx, taskID, 300*time.Second)
fmt.Println(solution["gRecaptchaResponse"])
```

## Proxies

```go
_, err = client.Submit(ctx, "turnstile", map[string]any{
	"sitekey":  sitekey,
	"pageurl":  pageurl,
	"proxy":    "socks5://user:pass@1.2.3.4:1080",
})
```

Any accepted spelling works (`scheme://user:pass@host:port`,
`host:port:user:pass`, `user:pass@host:port`). The scheme is never sent
upstream; `proxytype` is derived (`http`/`https` → `HTTP`, `socks4`, `socks5`).

## Errors

`*azcaptcha.Error` carries `Code` and `Retryable`; category sentinels support
`errors.Is`:

```go
import "errors"

token, err := client.SolveTurnstile(ctx, sitekey, pageurl, nil)
switch {
case errors.Is(err, azcaptcha.ErrBalance):
	// ERROR_ZERO_BALANCE
case errors.Is(err, azcaptcha.ErrCapacity):
	// ERROR_NO_SLOT_AVAILABLE — retry with backoff
case errors.Is(err, azcaptcha.ErrResult):
	// unsolvable / timeout / expired
}
```

`WaitResult` retries only codes marked retryable and backs off 10s ×1.5 up to
60s. Context cancellation aborts polling immediately.

## Timeouts

Defaults match server settings: image 120s, hCaptcha 120s, Turnstile 180s,
reCAPTCHA and others 300s. Override with
`SolveWithTimeout(ctx, method, params, 90*time.Second)`.

## Custom transport

```go
client, err := azcaptcha.New("KEY", azcaptcha.Config{
	Transport: myTransport, // implements azcaptcha.Transport
	Sleep:     mySleep,     // func(time.Duration)
})
```

## Links

- API contract: <https://github.com/AZcaptcha/sdk-spec>
- Docs: <https://azcaptcha.com/docs>
- Go guide: <https://azcaptcha.com/lang/go>
- Changelog: [CHANGELOG.md](CHANGELOG.md)

## License

MIT — see [LICENSE](LICENSE).
