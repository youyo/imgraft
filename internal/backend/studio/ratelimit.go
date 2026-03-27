package studio

import (
	"net/http"
	"strconv"
)

// parseRateLimit extracts rate-limit information from HTTP response headers.
// Fields that are absent or unparseable are left as nil.
// provider is set to the given value (e.g., "google_ai_studio") if non-empty.
func parseRateLimit(header http.Header, provider string) RateLimitInfo {
	info := RateLimitInfo{}

	if provider != "" {
		info.Provider = &provider
	}

	if v := header.Get("x-ratelimit-limit-requests"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			info.RequestsLimit = &n
		}
	}

	if v := header.Get("x-ratelimit-remaining-requests"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			info.RequestsRemaining = &n
		}
	}

	if v := header.Get("x-ratelimit-reset-requests"); v != "" {
		info.ResetAt = &v
	}

	if v := header.Get("Retry-After"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			info.RetryAfterSeconds = &n
		}
		// Non-integer Retry-After values (e.g., HTTP-date) are silently ignored.
	}

	return info
}
