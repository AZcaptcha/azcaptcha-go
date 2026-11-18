package azcaptcha

import (
	"context"
	"time"
)

// Solve submits and waits, using the per-method timeout from DefaultTimeouts.
func (c *Client) Solve(ctx context.Context, method string, params map[string]any) (string, error) {
	return c.SolveWithTimeout(ctx, method, params, timeoutFor(method))
}

// SolveWithTimeout is Solve with an explicit deadline.
func (c *Client) SolveWithTimeout(ctx context.Context, method string, params map[string]any, timeout time.Duration) (string, error) {
	taskID, err := c.Submit(ctx, method, params)
	if err != nil {
		return "", err
	}
	return c.WaitResult(ctx, taskID, timeout)
}

// SolveImage solves a raw image (PNG/JPEG bytes).
func (c *Client) SolveImage(ctx context.Context, image []byte, params map[string]any) (string, error) {
	return c.Solve(ctx, "base64", mergeParams(params, "body", ImageBody(image)))
}

// SolveImageFile reads a file and solves it.
func (c *Client) SolveImageFile(ctx context.Context, path string, params map[string]any) (string, error) {
	body, err := ImageFile(path)
	if err != nil {
		return "", err
	}
	return c.Solve(ctx, "base64", mergeParams(params, "body", body))
}

// SolveImageBase64 solves an already-encoded image (data: URI accepted).
func (c *Client) SolveImageBase64(ctx context.Context, image string, params map[string]any) (string, error) {
	return c.Solve(ctx, "base64", mergeParams(params, "body", ImageBase64(image)))
}

// RecaptchaV2Options tunes the v2 helper.
type RecaptchaV2Options struct {
	Invisible  bool
	Enterprise bool
	Params     map[string]any
}

func (c *Client) SolveRecaptchaV2(ctx context.Context, sitekey, pageurl string, opts *RecaptchaV2Options) (string, error) {
	params := map[string]any{"googlekey": sitekey, "pageurl": pageurl}
	if opts != nil {
		if opts.Invisible {
			params["invisible"] = 1
		}
		if opts.Enterprise {
			params["enterprise"] = 1
		}
		for key, value := range opts.Params {
			params[key] = value
		}
	}
	return c.Solve(ctx, "userrecaptcha", params)
}

// RecaptchaV3Options tunes the v3 helper.
type RecaptchaV3Options struct {
	MinScore float64
	Params   map[string]any
}

func (c *Client) SolveRecaptchaV3(ctx context.Context, sitekey, pageurl, action string, opts *RecaptchaV3Options) (string, error) {
	minScore := 0.6
	params := map[string]any{
		"googlekey": sitekey, "pageurl": pageurl,
		"version": "v3", "action": action,
	}
	if opts != nil {
		if opts.MinScore > 0 {
			minScore = opts.MinScore
		}
		for key, value := range opts.Params {
			params[key] = value
		}
	}
	params["min_score"] = minScore
	return c.Solve(ctx, "userrecaptcha", params)
}

func (c *Client) SolveHCaptcha(ctx context.Context, sitekey, pageurl string, params map[string]any) (string, error) {
	return c.Solve(ctx, "hcaptcha", mergeParams(params, "sitekey", sitekey, "pageurl", pageurl))
}

func (c *Client) SolveTurnstile(ctx context.Context, sitekey, pageurl string, params map[string]any) (string, error) {
	return c.Solve(ctx, "turnstile", mergeParams(params, "sitekey", sitekey, "pageurl", pageurl))
}

func (c *Client) SolveFunCaptcha(ctx context.Context, publickey, pageurl string, params map[string]any) (string, error) {
	return c.Solve(ctx, "funcaptcha", mergeParams(params, "publickey", publickey, "pageurl", pageurl))
}

func (c *Client) SolveGeeTest(ctx context.Context, gt, challenge, pageurl string, params map[string]any) (string, error) {
	return c.Solve(ctx, "geetest", mergeParams(params, "gt", gt, "challenge", challenge, "pageurl", pageurl))
}

func mergeParams(base map[string]any, pairs ...any) map[string]any {
	out := make(map[string]any, len(base)+len(pairs)/2)
	for key, value := range base {
		out[key] = value
	}
	for i := 0; i+1 < len(pairs); i += 2 {
		out[pairs[i].(string)] = pairs[i+1]
	}
	return out
}
