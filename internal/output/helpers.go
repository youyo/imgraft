// Package output defines the JSON output contract for imgraft.
// All stdout output follows a fixed schema (SPEC.md section 14).
package output

// StrPtr converts a string value to a *string.
// Used to set nullable JSON fields to non-null values.
func StrPtr(s string) *string { return &s }

// IntPtr converts an int value to a *int.
// Used to set nullable JSON fields to non-null values.
func IntPtr(i int) *int { return &i }
