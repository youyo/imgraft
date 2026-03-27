package studio

import (
	"net/http"
	"testing"
)

func TestParseRateLimit_RetryAfter(t *testing.T) {
	h := http.Header{}
	h.Set("Retry-After", "30")

	info := parseRateLimit(h, "google_ai_studio")

	if info.Provider == nil || *info.Provider != "google_ai_studio" {
		t.Errorf("Provider = %v, want google_ai_studio", info.Provider)
	}
	if info.RetryAfterSeconds == nil || *info.RetryAfterSeconds != 30 {
		t.Errorf("RetryAfterSeconds = %v, want 30", info.RetryAfterSeconds)
	}
	// Other fields should be nil.
	if info.RequestsLimit != nil {
		t.Errorf("RequestsLimit should be nil, got %v", info.RequestsLimit)
	}
}

func TestParseRateLimit_NoHeaders(t *testing.T) {
	h := http.Header{}

	info := parseRateLimit(h, "google_ai_studio")

	if info.Provider == nil || *info.Provider != "google_ai_studio" {
		t.Errorf("Provider = %v, want google_ai_studio", info.Provider)
	}
	if info.RequestsLimit != nil {
		t.Errorf("RequestsLimit should be nil")
	}
	if info.RequestsRemaining != nil {
		t.Errorf("RequestsRemaining should be nil")
	}
	if info.ResetAt != nil {
		t.Errorf("ResetAt should be nil")
	}
	if info.RetryAfterSeconds != nil {
		t.Errorf("RetryAfterSeconds should be nil")
	}
	if info.LimitType != nil {
		t.Errorf("LimitType should be nil")
	}
	if info.RequestsUsed != nil {
		t.Errorf("RequestsUsed should be nil")
	}
}

func TestParseRateLimit_RequestsLimit(t *testing.T) {
	h := http.Header{}
	h.Set("x-ratelimit-limit-requests", "60")

	info := parseRateLimit(h, "google_ai_studio")

	if info.RequestsLimit == nil || *info.RequestsLimit != 60 {
		t.Errorf("RequestsLimit = %v, want 60", info.RequestsLimit)
	}
}

func TestParseRateLimit_RequestsRemaining(t *testing.T) {
	h := http.Header{}
	h.Set("x-ratelimit-remaining-requests", "42")

	info := parseRateLimit(h, "google_ai_studio")

	if info.RequestsRemaining == nil || *info.RequestsRemaining != 42 {
		t.Errorf("RequestsRemaining = %v, want 42", info.RequestsRemaining)
	}
}

func TestParseRateLimit_ResetAt(t *testing.T) {
	h := http.Header{}
	h.Set("x-ratelimit-reset-requests", "2026-03-28T12:00:00Z")

	info := parseRateLimit(h, "google_ai_studio")

	if info.ResetAt == nil || *info.ResetAt != "2026-03-28T12:00:00Z" {
		t.Errorf("ResetAt = %v, want 2026-03-28T12:00:00Z", info.ResetAt)
	}
}

func TestParseRateLimit_RetryAfterNonNumeric(t *testing.T) {
	h := http.Header{}
	h.Set("Retry-After", "abc")

	info := parseRateLimit(h, "google_ai_studio")

	if info.RetryAfterSeconds != nil {
		t.Errorf("RetryAfterSeconds should be nil for non-numeric value, got %v", *info.RetryAfterSeconds)
	}
}

func TestParseRateLimit_EmptyProvider(t *testing.T) {
	h := http.Header{}

	info := parseRateLimit(h, "")

	if info.Provider != nil {
		t.Errorf("Provider should be nil for empty provider, got %v", *info.Provider)
	}
}

func TestParseRateLimit_AllHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("x-ratelimit-limit-requests", "100")
	h.Set("x-ratelimit-remaining-requests", "95")
	h.Set("x-ratelimit-reset-requests", "2026-03-28T13:00:00Z")
	h.Set("Retry-After", "10")

	info := parseRateLimit(h, "google_ai_studio")

	if info.RequestsLimit == nil || *info.RequestsLimit != 100 {
		t.Errorf("RequestsLimit = %v, want 100", info.RequestsLimit)
	}
	if info.RequestsRemaining == nil || *info.RequestsRemaining != 95 {
		t.Errorf("RequestsRemaining = %v, want 95", info.RequestsRemaining)
	}
	if info.ResetAt == nil || *info.ResetAt != "2026-03-28T13:00:00Z" {
		t.Errorf("ResetAt = %v, want 2026-03-28T13:00:00Z", info.ResetAt)
	}
	if info.RetryAfterSeconds == nil || *info.RetryAfterSeconds != 10 {
		t.Errorf("RetryAfterSeconds = %v, want 10", info.RetryAfterSeconds)
	}
}
