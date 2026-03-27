package studio

import (
	"testing"

	"github.com/youyo/imgraft/internal/errs"
)

func TestMapStatusToCode(t *testing.T) {
	cases := []struct {
		name string
		code int
		want errs.ErrorCode
	}{
		{"400 -> INVALID_ARGUMENT", 400, errs.ErrInvalidArgument},
		{"401 -> AUTH_INVALID", 401, errs.ErrAuthInvalid},
		{"403 -> AUTH_INVALID (PERMISSION_DENIED)", 403, errs.ErrAuthInvalid},
		{"404 -> BACKEND_UNAVAILABLE", 404, errs.ErrBackendUnavailable},
		{"429 -> RATE_LIMIT_EXCEEDED", 429, errs.ErrRateLimitExceeded},
		{"500 -> INTERNAL_ERROR", 500, errs.ErrInternal},
		{"503 -> BACKEND_UNAVAILABLE", 503, errs.ErrBackendUnavailable},
		{"502 -> INTERNAL_ERROR (other 5xx)", 502, errs.ErrInternal},
		{"418 -> INTERNAL_ERROR (unknown)", 418, errs.ErrInternal},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapStatusToCode(tc.code)
			if got != tc.want {
				t.Errorf("mapStatusToCode(%d) = %q, want %q", tc.code, got, tc.want)
			}
		})
	}
}

func TestMapHTTPError_WithAPIErrorBody(t *testing.T) {
	body := []byte(`{"error":{"code":429,"message":"Resource exhausted","status":"RESOURCE_EXHAUSTED"}}`)
	coded := mapHTTPError(429, body)

	if coded.Code != errs.ErrRateLimitExceeded {
		t.Errorf("code = %q, want %q", coded.Code, errs.ErrRateLimitExceeded)
	}
	if coded.Error() == "" {
		t.Error("error message should not be empty")
	}
	// The message should contain the API error detail.
	if got := coded.Err.Error(); got != "Resource exhausted" {
		t.Errorf("inner message = %q, want %q", got, "Resource exhausted")
	}
}

func TestMapHTTPError_WithInvalidBody(t *testing.T) {
	body := []byte(`not json`)
	coded := mapHTTPError(500, body)

	if coded.Code != errs.ErrInternal {
		t.Errorf("code = %q, want %q", coded.Code, errs.ErrInternal)
	}
	// Should fall back to a generic message.
	if got := coded.Err.Error(); got != "API request failed with status 500" {
		t.Errorf("inner message = %q, want generic message", got)
	}
}

func TestMapHTTPError_WithEmptyBody(t *testing.T) {
	body := []byte(``)
	coded := mapHTTPError(401, body)

	if coded.Code != errs.ErrAuthInvalid {
		t.Errorf("code = %q, want %q", coded.Code, errs.ErrAuthInvalid)
	}
}
