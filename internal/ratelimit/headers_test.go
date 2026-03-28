package ratelimit_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/youyo/imgraft/internal/ratelimit"
)

// TestParseHeaders_AllPresent tests that all rate-limit headers are parsed correctly.
func TestParseHeaders_AllPresent(t *testing.T) {
	h := http.Header{}
	h.Set("X-RateLimit-Limit-Requests", "100")
	h.Set("X-RateLimit-Remaining-Requests", "95")
	h.Set("X-RateLimit-Limit-Tokens", "50000")
	h.Set("X-RateLimit-Remaining-Tokens", "48000")
	h.Set("Retry-After", "30")

	info := ratelimit.ParseHeaders(h, "google_ai_studio")

	if info.Provider == nil || *info.Provider != "google_ai_studio" {
		t.Errorf("Provider = %v, want google_ai_studio", info.Provider)
	}
	if info.RequestsLimit == nil || *info.RequestsLimit != 100 {
		t.Errorf("RequestsLimit = %v, want 100", info.RequestsLimit)
	}
	if info.RequestsRemaining == nil || *info.RequestsRemaining != 95 {
		t.Errorf("RequestsRemaining = %v, want 95", info.RequestsRemaining)
	}
	if info.TokensLimit == nil || *info.TokensLimit != 50000 {
		t.Errorf("TokensLimit = %v, want 50000", info.TokensLimit)
	}
	if info.TokensRemaining == nil || *info.TokensRemaining != 48000 {
		t.Errorf("TokensRemaining = %v, want 48000", info.TokensRemaining)
	}
	if info.RetryAfterSeconds == nil || *info.RetryAfterSeconds != 30 {
		t.Errorf("RetryAfterSeconds = %v, want 30", info.RetryAfterSeconds)
	}
}

// TestParseHeaders_NoHeaders tests that all fields are nil when no headers are present.
func TestParseHeaders_NoHeaders(t *testing.T) {
	h := http.Header{}

	info := ratelimit.ParseHeaders(h, "google_ai_studio")

	// Provider is set from the argument, not headers
	if info.Provider == nil || *info.Provider != "google_ai_studio" {
		t.Errorf("Provider = %v, want google_ai_studio", info.Provider)
	}
	if info.RequestsLimit != nil {
		t.Errorf("RequestsLimit should be nil, got %v", *info.RequestsLimit)
	}
	if info.RequestsRemaining != nil {
		t.Errorf("RequestsRemaining should be nil, got %v", *info.RequestsRemaining)
	}
	if info.TokensLimit != nil {
		t.Errorf("TokensLimit should be nil, got %v", *info.TokensLimit)
	}
	if info.TokensRemaining != nil {
		t.Errorf("TokensRemaining should be nil, got %v", *info.TokensRemaining)
	}
	if info.RetryAfterSeconds != nil {
		t.Errorf("RetryAfterSeconds should be nil, got %v", *info.RetryAfterSeconds)
	}
	if info.LimitType != nil {
		t.Errorf("LimitType should be nil, got %v", *info.LimitType)
	}
	if info.RequestsUsed != nil {
		t.Errorf("RequestsUsed should be nil, got %v", *info.RequestsUsed)
	}
	if info.ResetAt != nil {
		t.Errorf("ResetAt should be nil, got %v", *info.ResetAt)
	}
}

// TestParseHeaders_EmptyProvider tests that Provider is nil when empty string is passed.
func TestParseHeaders_EmptyProvider(t *testing.T) {
	h := http.Header{}

	info := ratelimit.ParseHeaders(h, "")

	if info.Provider != nil {
		t.Errorf("Provider should be nil for empty provider, got %v", *info.Provider)
	}
}

// TestParseHeaders_RetryAfterSeconds tests parsing of Retry-After as integer seconds.
func TestParseHeaders_RetryAfterSeconds(t *testing.T) {
	h := http.Header{}
	h.Set("Retry-After", "120")

	info := ratelimit.ParseHeaders(h, "google_ai_studio")

	if info.RetryAfterSeconds == nil || *info.RetryAfterSeconds != 120 {
		t.Errorf("RetryAfterSeconds = %v, want 120", info.RetryAfterSeconds)
	}
}

// TestParseHeaders_RetryAfterHTTPDate tests parsing of Retry-After as HTTP-date format.
func TestParseHeaders_RetryAfterHTTPDate(t *testing.T) {
	// HTTP-date format: "Mon, 02 Jan 2006 15:04:05 GMT" (RFC 1123)
	// Use a fixed future time
	futureTime := time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)
	httpDate := futureTime.Format("Mon, 02 Jan 2006 15:04:05 GMT")

	h := http.Header{}
	h.Set("Retry-After", httpDate)

	// We provide a fixed "now" to make the test deterministic via a clock
	// Since ParseHeaders uses real time, we check that the result is positive
	// and within a reasonable range if futureTime > now
	// For unit testing, we only verify that the result is parsed (non-nil)
	info := ratelimit.ParseHeaders(h, "google_ai_studio")

	// The parsing should succeed and set RetryAfterSeconds
	if info.RetryAfterSeconds == nil {
		t.Errorf("RetryAfterSeconds should not be nil for valid HTTP-date Retry-After")
	}
}

// TestParseHeaders_RetryAfterHTTPDatePast tests that a past HTTP-date results in 0 seconds.
func TestParseHeaders_RetryAfterHTTPDatePast(t *testing.T) {
	// Use a time far in the past
	pastTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	httpDate := pastTime.Format("Mon, 02 Jan 2006 15:04:05 GMT")

	h := http.Header{}
	h.Set("Retry-After", httpDate)

	info := ratelimit.ParseHeaders(h, "google_ai_studio")

	// Past date: RetryAfterSeconds should be 0
	if info.RetryAfterSeconds == nil {
		t.Errorf("RetryAfterSeconds should not be nil for past HTTP-date Retry-After")
	}
	if *info.RetryAfterSeconds != 0 {
		t.Errorf("RetryAfterSeconds = %v, want 0 for past date", *info.RetryAfterSeconds)
	}
}

// TestParseHeaders_RetryAfterInvalid tests that invalid Retry-After values result in nil.
func TestParseHeaders_RetryAfterInvalid(t *testing.T) {
	h := http.Header{}
	h.Set("Retry-After", "not-a-valid-value")

	info := ratelimit.ParseHeaders(h, "google_ai_studio")

	if info.RetryAfterSeconds != nil {
		t.Errorf("RetryAfterSeconds should be nil for invalid value, got %v", *info.RetryAfterSeconds)
	}
}

// TestParseHeaders_InvalidIntegerValues tests that non-integer header values are ignored.
func TestParseHeaders_InvalidIntegerValues(t *testing.T) {
	h := http.Header{}
	h.Set("X-RateLimit-Limit-Requests", "not-a-number")
	h.Set("X-RateLimit-Remaining-Requests", "also-not-a-number")
	h.Set("X-RateLimit-Limit-Tokens", "abc")
	h.Set("X-RateLimit-Remaining-Tokens", "def")

	info := ratelimit.ParseHeaders(h, "google_ai_studio")

	if info.RequestsLimit != nil {
		t.Errorf("RequestsLimit should be nil for invalid value, got %v", *info.RequestsLimit)
	}
	if info.RequestsRemaining != nil {
		t.Errorf("RequestsRemaining should be nil for invalid value, got %v", *info.RequestsRemaining)
	}
	if info.TokensLimit != nil {
		t.Errorf("TokensLimit should be nil for invalid value, got %v", *info.TokensLimit)
	}
	if info.TokensRemaining != nil {
		t.Errorf("TokensRemaining should be nil for invalid value, got %v", *info.TokensRemaining)
	}
}

// TestParseHeaders_PartialHeaders tests that only present headers are set.
func TestParseHeaders_PartialHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("X-RateLimit-Limit-Requests", "60")
	// Remaining is absent

	info := ratelimit.ParseHeaders(h, "google_ai_studio")

	if info.RequestsLimit == nil || *info.RequestsLimit != 60 {
		t.Errorf("RequestsLimit = %v, want 60", info.RequestsLimit)
	}
	if info.RequestsRemaining != nil {
		t.Errorf("RequestsRemaining should be nil, got %v", *info.RequestsRemaining)
	}
}

// TestRateLimit_ToOutput tests conversion from RateLimit to output.RateLimit.
func TestRateLimit_ToOutput(t *testing.T) {
	h := http.Header{}
	h.Set("X-RateLimit-Limit-Requests", "100")
	h.Set("X-RateLimit-Remaining-Requests", "80")

	rl := ratelimit.ParseHeaders(h, "google_ai_studio")
	out := rl.ToOutput()

	if out.Provider == nil || *out.Provider != "google_ai_studio" {
		t.Errorf("Provider = %v, want google_ai_studio", out.Provider)
	}
	if out.RequestsLimit == nil || *out.RequestsLimit != 100 {
		t.Errorf("RequestsLimit = %v, want 100", out.RequestsLimit)
	}
	if out.RequestsRemaining == nil || *out.RequestsRemaining != 80 {
		t.Errorf("RequestsRemaining = %v, want 80", out.RequestsRemaining)
	}
}
