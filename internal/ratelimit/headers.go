package ratelimit

import (
	"net/http"
	"strconv"
	"time"
)

// httpDateFormats lists the RFC 1123 and legacy HTTP date formats to try when
// parsing the Retry-After header as an HTTP-date.
var httpDateFormats = []string{
	"Mon, 02 Jan 2006 15:04:05 GMT",   // RFC 1123
	"Monday, 02-Jan-06 15:04:05 GMT",  // RFC 850
	"Mon Jan _2 15:04:05 2006",        // ANSI C
}

// ParseHeaders extracts rate-limit information from HTTP response headers.
//
// The provider string (e.g., "google_ai_studio") is stored directly in the
// returned RateLimit.Provider field; an empty provider string results in a nil
// Provider field.
//
// Headers read:
//   - X-RateLimit-Limit-Requests     → RequestsLimit
//   - X-RateLimit-Remaining-Requests → RequestsRemaining
//   - X-RateLimit-Limit-Tokens       → TokensLimit
//   - X-RateLimit-Remaining-Tokens   → TokensRemaining
//   - X-RateLimit-Reset-Requests     → ResetAt (string, passed through as-is)
//   - Retry-After                    → RetryAfterSeconds (integer seconds or HTTP-date)
//
// Any header that is absent or cannot be parsed is left as nil (not set to a zero
// value), in accordance with the imgraft JSON contract (SPEC.md section 15.1).
func ParseHeaders(header http.Header, provider string) RateLimit {
	rl := RateLimit{}

	if provider != "" {
		rl.Provider = &provider
	}

	if v := header.Get("X-RateLimit-Limit-Requests"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rl.RequestsLimit = &n
		}
	}

	if v := header.Get("X-RateLimit-Remaining-Requests"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rl.RequestsRemaining = &n
		}
	}

	if v := header.Get("X-RateLimit-Limit-Tokens"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rl.TokensLimit = &n
		}
	}

	if v := header.Get("X-RateLimit-Remaining-Tokens"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rl.TokensRemaining = &n
		}
	}

	if v := header.Get("X-RateLimit-Reset-Requests"); v != "" {
		rl.ResetAt = &v
	}

	if v := header.Get("Retry-After"); v != "" {
		if seconds := parseRetryAfter(v); seconds != nil {
			rl.RetryAfterSeconds = seconds
		}
	}

	return rl
}

// parseRetryAfter attempts to parse the Retry-After header value as either an
// integer number of seconds or an HTTP-date (RFC 1123).
//
// Returns:
//   - A pointer to the number of seconds to wait. For HTTP-dates in the past,
//     returns a pointer to 0.
//   - nil if the value cannot be parsed as either format.
func parseRetryAfter(value string) *int {
	// Try integer seconds first.
	if n, err := strconv.Atoi(value); err == nil {
		return &n
	}

	// Try HTTP-date formats.
	for _, format := range httpDateFormats {
		t, err := time.Parse(format, value)
		if err != nil {
			continue
		}
		// Calculate seconds from now until the retry time.
		secs := int(time.Until(t).Seconds())
		if secs < 0 {
			secs = 0
		}
		return &secs
	}

	return nil
}
