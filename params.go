package azcaptcha

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"
)

// DefaultTimeouts mirror the server-side expiry per method.
var DefaultTimeouts = map[string]time.Duration{
	"base64":              120 * time.Second,
	"grid":                120 * time.Second,
	"userrecaptcha":       300 * time.Second,
	"recaptcha_v3":        300 * time.Second,
	"hcaptcha":            120 * time.Second,
	"hcaptcha_enterprise": 120 * time.Second,
	"turnstile":           180 * time.Second,
	"funcaptcha":          300 * time.Second,
	"geetest":             300 * time.Second,
	"geetest_v4":          300 * time.Second,
	"mtcaptcha":           300 * time.Second,
	"friendlycaptcha":     300 * time.Second,
	"altcha":              300 * time.Second,
	"cf_clearance":        300 * time.Second,
	"aws_waf":             300 * time.Second,
	"filecrypt":           300 * time.Second,
}

var proxySchemes = map[string]string{
	"http": "HTTP", "https": "HTTP", "socks4": "SOCKS4", "socks5": "SOCKS5",
}

// NormalizeProxy accepts scheme://user:pass@host:port, host:port:user:pass and
// user:pass@host:port. The scheme is never sent in the proxy field.
func NormalizeProxy(proxy string) (string, string, error) {
	raw := strings.TrimSpace(proxy)
	proxyType := "HTTP"
	if idx := strings.Index(raw, "://"); idx >= 0 {
		scheme := strings.ToLower(raw[:idx])
		proxyType = proxySchemes[scheme]
		if proxyType == "" {
			proxyType = "HTTP"
		}
		raw = raw[idx+3:]
	}
	parts := strings.Split(raw, ":")
	if len(parts) == 4 && isDigits(parts[1]) {
		raw = parts[2] + ":" + parts[3] + "@" + parts[0] + ":" + parts[1]
	} else if !strings.Contains(raw, "@") {
		idx := strings.LastIndex(raw, ":")
		if idx < 0 || !isDigits(raw[idx+1:]) {
			return "", "", fmt.Errorf("proxy must include a numeric port")
		}
	}
	return raw, proxyType, nil
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// ImageBody encodes raw bytes as bare base64.
func ImageBody(image []byte) string {
	return base64.StdEncoding.EncodeToString(image)
}

// ImageFile reads a captcha image from disk and encodes it.
func ImageFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ErrFor("ERROR_IMAGE_NOT_FOUND")
	}
	return ImageBody(data), nil
}

// ImageBase64 normalises a data: URI or base64 string to bare base64.
func ImageBase64(image string) string {
	if strings.HasPrefix(image, "data:") {
		if idx := strings.Index(image, ","); idx >= 0 {
			return image[idx+1:]
		}
	}
	return strings.Join(strings.Fields(image), "")
}

func timeoutFor(method string) time.Duration {
	if d, ok := DefaultTimeouts[method]; ok {
		return d
	}
	return 300 * time.Second
}
