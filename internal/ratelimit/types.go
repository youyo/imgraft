// Package ratelimit provides types and utilities for parsing HTTP rate-limit
// headers from API responses.
//
// Design: All fields use pointer types so that absent values are represented
// as nil (JSON: null) rather than zero values. This is required by the imgraft
// JSON contract (SPEC.md section 15).
package ratelimit

import "github.com/youyo/imgraft/internal/output"

// RateLimit holds parsed rate-limit information extracted from HTTP response headers.
// Fields that are absent or could not be parsed are nil.
// See SPEC.md section 15 and 22.
type RateLimit struct {
	// Provider identifies the API backend (e.g., "google_ai_studio").
	// Set from the caller-supplied provider string, not from headers.
	Provider *string

	// LimitType identifies the type of rate limit (e.g., "rpm", "rpd").
	// Currently nil because Google AI Studio does not expose this via headers.
	LimitType *string

	// RequestsLimit is the total request quota for the current window.
	// Sourced from X-RateLimit-Limit-Requests.
	RequestsLimit *int

	// RequestsRemaining is the number of requests left in the current window.
	// Sourced from X-RateLimit-Remaining-Requests.
	RequestsRemaining *int

	// RequestsUsed is the number of requests consumed. Per SPEC.md section 22.1,
	// this is NOT calculated as limit-remaining; it must come from a header.
	// Currently nil because Google AI Studio does not provide a used-requests header.
	RequestsUsed *int

	// TokensLimit is the total token quota for the current window.
	// Sourced from X-RateLimit-Limit-Tokens.
	TokensLimit *int

	// TokensRemaining is the number of tokens left in the current window.
	// Sourced from X-RateLimit-Remaining-Tokens.
	TokensRemaining *int

	// ResetAt is the ISO-8601 timestamp when the rate limit resets.
	// Sourced from X-RateLimit-Reset-Requests.
	ResetAt *string

	// RetryAfterSeconds is how many seconds to wait before retrying.
	// Sourced from the Retry-After header. Supports both integer-seconds
	// and HTTP-date (RFC 1123) formats. For past dates, the value is 0.
	RetryAfterSeconds *int
}

// ToOutput converts a RateLimit into the output.RateLimit type used in JSON responses.
// The TokensLimit and TokensRemaining fields are currently not part of the output schema;
// they are available in this internal type for future use.
func (r RateLimit) ToOutput() output.RateLimit {
	return output.RateLimit{
		Provider:          r.Provider,
		LimitType:         r.LimitType,
		RequestsLimit:     r.RequestsLimit,
		RequestsRemaining: r.RequestsRemaining,
		RequestsUsed:      r.RequestsUsed,
		ResetAt:           r.ResetAt,
		RetryAfterSeconds: r.RetryAfterSeconds,
	}
}
