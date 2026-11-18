package azcaptcha_test

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	azcaptcha "github.com/azcaptcha/azcaptcha-go"
)

type taxonomySpec struct {
	EnvelopeCodes map[string]struct {
		Retryable bool `json:"retryable"`
	} `json:"envelope_codes"`
	Errors map[string]struct {
		Retryable bool `json:"retryable"`
	} `json:"errors"`
}

func loadTaxonomy(t *testing.T) taxonomySpec {
	t.Helper()
	raw, err := os.ReadFile("testdata/fixtures/errors.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec taxonomySpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatal(err)
	}
	return spec
}

func TestEveryCodeMapsToTypedError(t *testing.T) {
	spec := loadTaxonomy(t)
	flags := azcaptcha.Taxonomy()
	for code := range spec.Errors {
		err := azcaptcha.ErrFor(code)
		if errors.Is(err, azcaptcha.ErrUnknown) {
			t.Fatalf("%s mapped to ErrUnknown", code)
		}
		if err.Code != code || err.Retryable != spec.Errors[code].Retryable {
			t.Fatalf("%s = %+v", code, err)
		}
		if flags[code] != err.Retryable {
			t.Fatalf("%s taxonomy flag", code)
		}
	}
	for code, meta := range spec.EnvelopeCodes {
		if azcaptcha.ErrFor(code).Retryable != meta.Retryable {
			t.Fatalf("%s retryable", code)
		}
	}
}

func TestForwardCompatibleCode(t *testing.T) {
	err := azcaptcha.ErrFor("ERROR_FROM_THE_FUTURE")
	if !errors.Is(err, azcaptcha.ErrUnknown) || err.Retryable {
		t.Fatalf("err = %+v", err)
	}
}

func TestCategorySentinels(t *testing.T) {
	cases := map[string]error{
		"ERROR_KEY_DOES_NOT_EXIST":    azcaptcha.ErrAuth,
		"ERROR_ZERO_BALANCE":          azcaptcha.ErrBalance,
		"ERROR_BAD_PARAMETERS":        azcaptcha.ErrRequest,
		"ERROR_NO_SLOT_AVAILABLE":     azcaptcha.ErrCapacity,
		"ERROR_CAPTCHA_UNSOLVABLE":    azcaptcha.ErrResult,
		"ERROR_PROXY_BANNED":          azcaptcha.ErrProxy,
		"ERROR_PINGBACK_NOT_VERIFIED": azcaptcha.ErrPingback,
		"ERROR_MODULE_NOT_TRAINED":    azcaptcha.ErrModule,
	}
	for code, sentinel := range cases {
		if !errors.Is(azcaptcha.ErrFor(code), sentinel) {
			t.Fatalf("%s does not match sentinel", code)
		}
	}
}
