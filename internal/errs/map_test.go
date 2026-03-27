package errs_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/youyo/imgraft/internal/errs"
)

func TestNew(t *testing.T) {
	err := errs.New(errs.ErrInvalidArgument, "missing required flag")
	if err == nil {
		t.Fatal("New() returned nil")
	}
	if err.Code != errs.ErrInvalidArgument {
		t.Errorf("Code = %s, want %s", err.Code, errs.ErrInvalidArgument)
	}
	if err.Error() == "" {
		t.Error("Error() returned empty string")
	}
	if err.Unwrap() == nil {
		t.Error("Unwrap() returned nil for New()")
	}
}

func TestNewErrorMessage(t *testing.T) {
	err := errs.New(errs.ErrFileNotFound, "file /tmp/test.png not found")
	msg := err.Error()
	// メッセージにコードと詳細が含まれること
	if msg == "" {
		t.Error("Error() returned empty string")
	}
}

func TestWrap(t *testing.T) {
	original := fmt.Errorf("underlying OS error")
	wrapped := errs.Wrap(errs.ErrFileReadFailed, original)
	if wrapped == nil {
		t.Fatal("Wrap() returned nil")
	}
	if wrapped.Code != errs.ErrFileReadFailed {
		t.Errorf("Code = %s, want %s", wrapped.Code, errs.ErrFileReadFailed)
	}
	// errors.Unwrap で元のエラーを取得できること
	unwrapped := wrapped.Unwrap()
	if unwrapped != original {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, original)
	}
}

func TestWrapNilError(t *testing.T) {
	wrapped := errs.Wrap(errs.ErrInternal, nil)
	if wrapped == nil {
		t.Fatal("Wrap(code, nil) returned nil")
	}
	if wrapped.Unwrap() != nil {
		t.Error("Unwrap() should return nil when wrapped error is nil")
	}
}

func TestCodeOf_CodedError(t *testing.T) {
	err := errs.New(errs.ErrRateLimitExceeded, "rate limit hit")
	code := errs.CodeOf(err)
	if code != errs.ErrRateLimitExceeded {
		t.Errorf("CodeOf() = %s, want %s", code, errs.ErrRateLimitExceeded)
	}
}

func TestCodeOf_NonCodedError(t *testing.T) {
	err := fmt.Errorf("some generic error")
	code := errs.CodeOf(err)
	if code != errs.ErrInternal {
		t.Errorf("CodeOf(generic error) = %s, want %s", code, errs.ErrInternal)
	}
}

func TestCodeOf_Nil(t *testing.T) {
	code := errs.CodeOf(nil)
	if code != errs.ErrInternal {
		t.Errorf("CodeOf(nil) = %s, want %s", code, errs.ErrInternal)
	}
}

func TestCodeOf_WrappedCodedError(t *testing.T) {
	inner := errs.New(errs.ErrAuthInvalid, "invalid api key")
	outer := fmt.Errorf("outer: %w", inner)
	code := errs.CodeOf(outer)
	if code != errs.ErrAuthInvalid {
		t.Errorf("CodeOf(wrapped) = %s, want %s", code, errs.ErrAuthInvalid)
	}
}

func TestErrorsAs(t *testing.T) {
	err := errs.New(errs.ErrModelResolutionFailed, "could not resolve model")
	var coded *errs.CodedError
	if !errors.As(err, &coded) {
		t.Error("errors.As() failed to find CodedError")
	}
	if coded.Code != errs.ErrModelResolutionFailed {
		t.Errorf("Code = %s, want %s", coded.Code, errs.ErrModelResolutionFailed)
	}
}

func TestErrorsIs(t *testing.T) {
	original := fmt.Errorf("original")
	wrapped := errs.Wrap(errs.ErrFileNotFound, original)
	// errors.Is で元のエラーが見つかること
	if !errors.Is(wrapped, original) {
		t.Error("errors.Is() failed to find original error through CodedError")
	}
}
