package errs

import "errors"

// CodedError pairs an ErrorCode with an underlying error.
// It implements the error interface and supports errors.As / errors.Is.
type CodedError struct {
	Code ErrorCode
	Err  error
}

// Error returns a human-readable string containing the code and the underlying message.
func (e *CodedError) Error() string {
	if e.Err != nil {
		return string(e.Code) + ": " + e.Err.Error()
	}
	return string(e.Code)
}

// Unwrap returns the underlying error so errors.Is / errors.As can traverse the chain.
func (e *CodedError) Unwrap() error {
	return e.Err
}

// New creates a *CodedError with the given code and a new error message.
func New(code ErrorCode, msg string) *CodedError {
	return &CodedError{
		Code: code,
		Err:  errors.New(msg),
	}
}

// Wrap wraps an existing error with an ErrorCode.
// If err is nil, the CodedError is created with a nil Err field.
func Wrap(code ErrorCode, err error) *CodedError {
	return &CodedError{
		Code: code,
		Err:  err,
	}
}

// CodeOf extracts the ErrorCode from err.
// If err contains a *CodedError anywhere in its chain, that code is returned.
// Otherwise ErrInternal is returned.
func CodeOf(err error) ErrorCode {
	if err == nil {
		return ErrInternal
	}
	var coded *CodedError
	if errors.As(err, &coded) {
		return coded.Code
	}
	return ErrInternal
}
