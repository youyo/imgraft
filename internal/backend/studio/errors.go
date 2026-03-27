package studio

import (
	"encoding/json"
	"fmt"

	"github.com/youyo/imgraft/internal/errs"
)

// mapHTTPError maps an HTTP status code and response body to a *errs.CodedError.
// It attempts to parse the body as an API error response for a detailed message.
func mapHTTPError(statusCode int, body []byte) *errs.CodedError {
	code := mapStatusToCode(statusCode)

	// Try to extract a detailed message from the API error response.
	var apiErr apiErrorResponse
	msg := fmt.Sprintf("API request failed with status %d", statusCode)
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
		msg = apiErr.Error.Message
	}

	return errs.New(code, msg)
}

// mapStatusToCode maps an HTTP status code to an errs.ErrorCode.
func mapStatusToCode(statusCode int) errs.ErrorCode {
	switch statusCode {
	case 400:
		return errs.ErrInvalidArgument
	case 401:
		return errs.ErrAuthInvalid
	case 403:
		return errs.ErrAuthInvalid // PERMISSION_DENIED
	case 404:
		return errs.ErrBackendUnavailable
	case 429:
		return errs.ErrRateLimitExceeded
	case 503:
		return errs.ErrBackendUnavailable
	default:
		if statusCode >= 500 {
			return errs.ErrInternal
		}
		return errs.ErrInternal
	}
}
