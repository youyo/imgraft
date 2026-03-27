// Package errs defines error codes for imgraft.
// All error codes correspond to SPEC.md section 16.1.
package errs

// ErrorCode is a string identifier for imgraft error categories.
// These codes are used in the JSON output's error.code field.
type ErrorCode string

const (
	// Input / argument errors
	ErrInvalidArgument ErrorCode = "INVALID_ARGUMENT"

	// Authentication errors
	ErrAuthRequired ErrorCode = "AUTH_REQUIRED"
	ErrAuthInvalid  ErrorCode = "AUTH_INVALID"

	// File / image input errors
	ErrFileNotFound             ErrorCode = "FILE_NOT_FOUND"
	ErrFileReadFailed           ErrorCode = "FILE_READ_FAILED"
	ErrUnsupportedImageFormat   ErrorCode = "UNSUPPORTED_IMAGE_FORMAT"
	ErrImageTooLarge            ErrorCode = "IMAGE_TOO_LARGE"
	ErrInvalidImage             ErrorCode = "INVALID_IMAGE"

	// Reference image (URL) errors
	ErrReferenceFetchFailed              ErrorCode = "REFERENCE_FETCH_FAILED"
	ErrReferenceTimeout                  ErrorCode = "REFERENCE_TIMEOUT"
	ErrReferenceRedirectLimitExceeded    ErrorCode = "REFERENCE_REDIRECT_LIMIT_EXCEEDED"
	ErrReferenceURLForbidden             ErrorCode = "REFERENCE_URL_FORBIDDEN"

	// Output errors
	ErrOutputDirCreateFailed ErrorCode = "OUTPUT_DIR_CREATE_FAILED"
	ErrFileWriteFailed       ErrorCode = "FILE_WRITE_FAILED"
	ErrFileAlreadyExists     ErrorCode = "FILE_ALREADY_EXISTS"
	ErrInvalidOutputPath     ErrorCode = "INVALID_OUTPUT_PATH"

	// Model / backend errors
	ErrModelResolutionFailed ErrorCode = "MODEL_RESOLUTION_FAILED"
	ErrBackendUnavailable    ErrorCode = "BACKEND_UNAVAILABLE"
	ErrRateLimitExceeded     ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Catch-all
	ErrInternal ErrorCode = "INTERNAL_ERROR"
)

// allCodes is the canonical list of all defined error codes.
var allCodes = []ErrorCode{
	ErrInvalidArgument,
	ErrAuthRequired,
	ErrAuthInvalid,
	ErrFileNotFound,
	ErrFileReadFailed,
	ErrUnsupportedImageFormat,
	ErrImageTooLarge,
	ErrInvalidImage,
	ErrReferenceFetchFailed,
	ErrReferenceTimeout,
	ErrReferenceRedirectLimitExceeded,
	ErrReferenceURLForbidden,
	ErrOutputDirCreateFailed,
	ErrFileWriteFailed,
	ErrFileAlreadyExists,
	ErrInvalidOutputPath,
	ErrModelResolutionFailed,
	ErrBackendUnavailable,
	ErrRateLimitExceeded,
	ErrInternal,
}

// AllCodes returns a copy of all defined ErrorCode values.
func AllCodes() []ErrorCode {
	result := make([]ErrorCode, len(allCodes))
	copy(result, allCodes)
	return result
}

// validSet is a lookup set for fast IsValid checks.
var validSet map[ErrorCode]struct{}

func init() {
	validSet = make(map[ErrorCode]struct{}, len(allCodes))
	for _, c := range allCodes {
		validSet[c] = struct{}{}
	}
}

// IsValid reports whether c is a known ErrorCode.
func (c ErrorCode) IsValid() bool {
	_, ok := validSet[c]
	return ok
}
