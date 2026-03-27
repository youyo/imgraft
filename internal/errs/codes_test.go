package errs_test

import (
	"testing"

	"github.com/youyo/imgraft/internal/errs"
)

func TestErrorCodeValues(t *testing.T) {
	tests := []struct {
		code     errs.ErrorCode
		expected string
	}{
		{errs.ErrInvalidArgument, "INVALID_ARGUMENT"},
		{errs.ErrAuthRequired, "AUTH_REQUIRED"},
		{errs.ErrAuthInvalid, "AUTH_INVALID"},
		{errs.ErrFileNotFound, "FILE_NOT_FOUND"},
		{errs.ErrFileReadFailed, "FILE_READ_FAILED"},
		{errs.ErrUnsupportedImageFormat, "UNSUPPORTED_IMAGE_FORMAT"},
		{errs.ErrImageTooLarge, "IMAGE_TOO_LARGE"},
		{errs.ErrInvalidImage, "INVALID_IMAGE"},
		{errs.ErrReferenceFetchFailed, "REFERENCE_FETCH_FAILED"},
		{errs.ErrReferenceTimeout, "REFERENCE_TIMEOUT"},
		{errs.ErrReferenceRedirectLimitExceeded, "REFERENCE_REDIRECT_LIMIT_EXCEEDED"},
		{errs.ErrReferenceURLForbidden, "REFERENCE_URL_FORBIDDEN"},
		{errs.ErrOutputDirCreateFailed, "OUTPUT_DIR_CREATE_FAILED"},
		{errs.ErrFileWriteFailed, "FILE_WRITE_FAILED"},
		{errs.ErrFileAlreadyExists, "FILE_ALREADY_EXISTS"},
		{errs.ErrInvalidOutputPath, "INVALID_OUTPUT_PATH"},
		{errs.ErrModelResolutionFailed, "MODEL_RESOLUTION_FAILED"},
		{errs.ErrBackendUnavailable, "BACKEND_UNAVAILABLE"},
		{errs.ErrRateLimitExceeded, "RATE_LIMIT_EXCEEDED"},
		{errs.ErrInternal, "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			if string(tt.code) != tt.expected {
				t.Errorf("got %q, want %q", string(tt.code), tt.expected)
			}
		})
	}
}

func TestAllCodes(t *testing.T) {
	all := errs.AllCodes()
	if len(all) != 20 {
		t.Errorf("AllCodes() returned %d codes, want 20", len(all))
	}

	seen := make(map[errs.ErrorCode]bool)
	for _, c := range all {
		if seen[c] {
			t.Errorf("duplicate code: %s", c)
		}
		seen[c] = true
	}
}

func TestIsValidCode(t *testing.T) {
	validCodes := []errs.ErrorCode{
		errs.ErrInvalidArgument,
		errs.ErrAuthRequired,
		errs.ErrInternal,
	}
	for _, c := range validCodes {
		if !c.IsValid() {
			t.Errorf("expected %s to be valid", c)
		}
	}

	invalidCodes := []errs.ErrorCode{
		errs.ErrorCode("UNKNOWN_CODE"),
		errs.ErrorCode(""),
		errs.ErrorCode("invalid_argument"),
	}
	for _, c := range invalidCodes {
		if c.IsValid() {
			t.Errorf("expected %s to be invalid", c)
		}
	}
}
