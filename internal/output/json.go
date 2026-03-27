package output

import (
	"bytes"
	"encoding/json"
	"io"
)

// NewEmptyRateLimit returns a RateLimit with all fields set to null.
// Use this as the default when no rate limit information is available.
func NewEmptyRateLimit() RateLimit {
	return RateLimit{}
}

// NewSuccessOutput returns an Output initialized for a successful result.
// The caller should set Model, Backend, Images, RateLimit, and Warnings
// as needed before encoding.
func NewSuccessOutput() Output {
	return Output{
		Success:   true,
		Images:    []ImageItem{},
		RateLimit: NewEmptyRateLimit(),
		Warnings:  []string{},
		Error:     OutputError{},
	}
}

// NewErrorOutput returns an Output initialized for a failure result.
// code should be a string representation of an errs.ErrorCode.
// message should be a human-readable error description.
func NewErrorOutput(code, message string) Output {
	return Output{
		Success:   false,
		Images:    []ImageItem{},
		RateLimit: NewEmptyRateLimit(),
		Warnings:  []string{},
		Error: OutputError{
			Code:    StrPtr(code),
			Message: StrPtr(message),
		},
	}
}

// Encode serializes out as JSON and writes it to w.
// When pretty is true, the output is indented with 2 spaces.
// Nil slices (Images, Warnings) are normalized to empty arrays
// to guarantee the fixed schema contract.
// Writing is buffered to prevent partial output on encoding errors.
func Encode(w io.Writer, out Output, pretty bool) error {
	// Normalize nil slices to empty arrays for schema guarantee.
	if out.Images == nil {
		out.Images = []ImageItem{}
	}
	if out.Warnings == nil {
		out.Warnings = []string{}
	}

	// Buffer the output to prevent partial writes on error.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(out); err != nil {
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}
