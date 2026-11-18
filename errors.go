package azcaptcha

import "errors"

// Category sentinels — use with errors.Is to branch without matching codes.
var (
	ErrProcessing = errors.New("azcaptcha: processing")
	ErrAuth       = errors.New("azcaptcha: auth")
	ErrBalance    = errors.New("azcaptcha: balance")
	ErrRequest    = errors.New("azcaptcha: request")
	ErrLimit      = errors.New("azcaptcha: limit")
	ErrCapacity   = errors.New("azcaptcha: capacity")
	ErrService    = errors.New("azcaptcha: service")
	ErrResult     = errors.New("azcaptcha: result")
	ErrProxy      = errors.New("azcaptcha: proxy")
	ErrPingback   = errors.New("azcaptcha: pingback")
	ErrModule     = errors.New("azcaptcha: module")
	ErrNetwork    = errors.New("azcaptcha: network")
	ErrUnknown    = errors.New("azcaptcha: unknown")
)

// Error carries the API code and the taxonomy retryable flag.
type Error struct {
	Code      string
	Retryable bool
	Message   string
	Err       error
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

func (e *Error) Unwrap() error { return e.Err }

type errMeta struct {
	sentinel  error
	retryable bool
}

var taxonomyTable = map[string]errMeta{
	"CAPCHA_NOT_READY":                {ErrProcessing, true},
	"ERROR_KEY_DOES_NOT_EXIST":        {ErrAuth, false},
	"ERROR_ACCOUNT_SUSPENDED":         {ErrAuth, false},
	"ERROR_ZERO_BALANCE":              {ErrBalance, false},
	"ERROR_BAD_PARAMETERS":            {ErrRequest, false},
	"ERROR_WRONG_ID_FORMAT":           {ErrRequest, false},
	"ERROR_NO_SUCH_CAPCHA_ID":         {ErrRequest, false},
	"ERROR_TASK_NOT_FOUND":            {ErrRequest, false},
	"ERROR_TASK_NOT_SUPPORTED":        {ErrRequest, false},
	"ERROR_WRONG_GOOGLEKEY":           {ErrRequest, false},
	"ERROR_RECAPTCHA_INVALID_SITEKEY": {ErrRequest, false},
	"ERROR_WRONG_PAGEURL":             {ErrRequest, false},
	"ERROR_IMAGE_NOT_FOUND":           {ErrRequest, false},
	"ERROR_IMAGE_TYPE_NOT_SUPPORTED":  {ErrRequest, false},
	"ERROR_ZERO_CAPTCHA_FILESIZE":     {ErrRequest, false},
	"ERROR_TOO_BIG_CAPTCHA_FILESIZE":  {ErrRequest, false},
	"ERROR_DAILY_LIMIT_EXCEEDED":      {ErrLimit, false},
	"ERROR_TOO_MANY_THREADS":          {ErrLimit, true},
	"ERROR_NO_SLOT_AVAILABLE":         {ErrCapacity, true},
	"ERROR_MAINTENANCE":               {ErrService, false},
	"ERROR_TIMEOUT":                   {ErrResult, false},
	"ERROR_TTL":                       {ErrResult, false},
	"ERROR_TOKEN_EXPIRED":             {ErrResult, false},
	"ERROR_CAPTCHA_UNSOLVABLE":        {ErrResult, false},
	"ERROR_PROXY_REQUIRED":            {ErrProxy, false},
	"ERROR_PROXY_NOT_AUTHORIZED":      {ErrProxy, false},
	"ERROR_PROXY_BANNED":              {ErrProxy, false},
	"ERROR_PINGBACK_NOT_VERIFIED":     {ErrPingback, false},
	"ERROR_MODULE_ACCESS_DENIED":      {ErrModule, false},
	"ERROR_MODULE_NOT_TRAINED":        {ErrModule, false},
}

// ErrFor builds the typed error for an API code.
func ErrFor(code string) *Error {
	meta, ok := taxonomyTable[code]
	if !ok {
		meta = errMeta{ErrUnknown, false}
	}
	return &Error{Code: code, Retryable: meta.retryable, Message: code, Err: meta.sentinel}
}

// Taxonomy exposes code -> retryable for tests and callers.
func Taxonomy() map[string]bool {
	out := make(map[string]bool, len(taxonomyTable))
	for code, meta := range taxonomyTable {
		out[code] = meta.retryable
	}
	return out
}
