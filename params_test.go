package azcaptcha_test

import (
	"os"
	"path/filepath"
	"testing"

	azcaptcha "github.com/azcaptcha/azcaptcha-go"
)

func TestNormalizeProxy(t *testing.T) {
	cases := []struct {
		in, proxy, proxyType string
	}{
		{"1.2.3.4:8080", "1.2.3.4:8080", "HTTP"},
		{"user:pass@1.2.3.4:8080", "user:pass@1.2.3.4:8080", "HTTP"},
		{"1.2.3.4:8080:user:pass", "user:pass@1.2.3.4:8080", "HTTP"},
		{"http://u:p@1.2.3.4:8080", "u:p@1.2.3.4:8080", "HTTP"},
		{"https://u:p@1.2.3.4:8080", "u:p@1.2.3.4:8080", "HTTP"},
		{"socks4://u:p@1.2.3.4:1080", "u:p@1.2.3.4:1080", "SOCKS4"},
		{"socks5://u:p@1.2.3.4:1080", "u:p@1.2.3.4:1080", "SOCKS5"},
	}
	for _, c := range cases {
		proxy, proxyType, err := azcaptcha.NormalizeProxy(c.in)
		if err != nil || proxy != c.proxy || proxyType != c.proxyType {
			t.Fatalf("%s = %q, %q, %v", c.in, proxy, proxyType, err)
		}
	}
	if _, _, err := azcaptcha.NormalizeProxy("1.2.3.4"); err == nil {
		t.Fatal("proxy without port must fail")
	}
}

func TestImageEncodings(t *testing.T) {
	if got := azcaptcha.ImageBody([]byte("PNG")); got != "UE5H" {
		t.Fatalf("ImageBody = %q", got)
	}
	if got := azcaptcha.ImageBase64("data:image/png;base64,AAAA"); got != "AAAA" {
		t.Fatalf("data URI = %q", got)
	}
	if got := azcaptcha.ImageBase64("iVBORw0KGgo=\n"); got != "iVBORw0KGgo=" {
		t.Fatalf("base64 = %q", got)
	}

	path := filepath.Join(t.TempDir(), "captcha.png")
	if err := os.WriteFile(path, []byte("PNG"), 0o600); err != nil {
		t.Fatal(err)
	}
	encoded, err := azcaptcha.ImageFile(path)
	if err != nil || encoded != "UE5H" {
		t.Fatalf("ImageFile = %q, %v", encoded, err)
	}
	if _, err := azcaptcha.ImageFile(filepath.Join(t.TempDir(), "missing.png")); err == nil {
		t.Fatal("missing file must fail")
	}
}
