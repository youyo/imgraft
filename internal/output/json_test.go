package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestEncode_SuccessMinimal verifies that a minimal success output encodes
// to the expected JSON schema.
func TestEncode_SuccessMinimal(t *testing.T) {
	out := NewSuccessOutput()
	var buf bytes.Buffer
	if err := Encode(&buf, out, false); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if m["success"] != true {
		t.Errorf("expected success to be true, got %v", m["success"])
	}

	images, ok := m["images"].([]interface{})
	if !ok {
		t.Fatalf("expected images to be an array, got %T", m["images"])
	}
	if len(images) != 0 {
		t.Errorf("expected images to be empty, got %d items", len(images))
	}

	warnings, ok := m["warnings"].([]interface{})
	if !ok {
		t.Fatalf("expected warnings to be an array, got %T", m["warnings"])
	}
	if len(warnings) != 0 {
		t.Errorf("expected warnings to be empty, got %d items", len(warnings))
	}
}

// TestEncode_ErrorOutput verifies that NewErrorOutput produces a valid
// error JSON with success=false.
func TestEncode_ErrorOutput(t *testing.T) {
	out := NewErrorOutput("INVALID_ARGUMENT", "prompt is required")
	var buf bytes.Buffer
	if err := Encode(&buf, out, false); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if m["success"] != false {
		t.Errorf("expected success to be false, got %v", m["success"])
	}

	images, ok := m["images"].([]interface{})
	if !ok {
		t.Fatalf("expected images to be an array, got %T", m["images"])
	}
	if len(images) != 0 {
		t.Errorf("expected images to be empty on error, got %d items", len(images))
	}

	errorMap, ok := m["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected error to be an object, got %T", m["error"])
	}
	if errorMap["code"] != "INVALID_ARGUMENT" {
		t.Errorf("expected error.code to be INVALID_ARGUMENT, got %v", errorMap["code"])
	}
	if errorMap["message"] != "prompt is required" {
		t.Errorf("expected error.message to be 'prompt is required', got %v", errorMap["message"])
	}
}

// TestEncode_PrettyPrint verifies that pretty=true produces indented JSON.
func TestEncode_PrettyPrint(t *testing.T) {
	out := NewSuccessOutput()
	var buf bytes.Buffer
	if err := Encode(&buf, out, true); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	result := buf.String()
	if !strings.Contains(result, "\n  ") {
		t.Errorf("expected indented JSON output, got: %s", result)
	}
}

// TestEncode_CompactOutput verifies that pretty=false produces compact JSON.
func TestEncode_CompactOutput(t *testing.T) {
	out := NewSuccessOutput()
	var buf bytes.Buffer
	if err := Encode(&buf, out, false); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	result := strings.TrimSpace(buf.String())
	// Compact JSON should be a single line
	lines := strings.Split(result, "\n")
	if len(lines) != 1 {
		t.Errorf("expected compact JSON to be single line, got %d lines", len(lines))
	}
}

// TestEncode_ImagesNilToEmptyArray verifies that even if Images is nil,
// the JSON output contains "images": [] not "images": null.
func TestEncode_ImagesNilToEmptyArray(t *testing.T) {
	out := Output{
		Success:   false,
		Images:    nil, // deliberately nil
		RateLimit: NewEmptyRateLimit(),
		Warnings:  []string{},
		Error:     OutputError{},
	}
	var buf bytes.Buffer
	if err := Encode(&buf, out, false); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// images must be an array, not null
	images, ok := m["images"].([]interface{})
	if !ok {
		t.Fatalf("expected images to be an array (not null), got %T (value: %v)", m["images"], m["images"])
	}
	if len(images) != 0 {
		t.Errorf("expected images to be empty, got %d items", len(images))
	}
}

// TestEncode_WarningsNilToEmptyArray verifies that even if Warnings is nil,
// the JSON output contains "warnings": [] not "warnings": null.
func TestEncode_WarningsNilToEmptyArray(t *testing.T) {
	out := Output{
		Success:   false,
		Images:    []ImageItem{},
		RateLimit: NewEmptyRateLimit(),
		Warnings:  nil, // deliberately nil
		Error:     OutputError{},
	}
	var buf bytes.Buffer
	if err := Encode(&buf, out, false); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// warnings must be an array, not null
	warnings, ok := m["warnings"].([]interface{})
	if !ok {
		t.Fatalf("expected warnings to be an array (not null), got %T (value: %v)", m["warnings"], m["warnings"])
	}
	if len(warnings) != 0 {
		t.Errorf("expected warnings to be empty, got %d items", len(warnings))
	}
}

// TestEncode_FullSuccessOutput verifies a complete success output with all
// fields populated.
func TestEncode_FullSuccessOutput(t *testing.T) {
	out := NewSuccessOutput()
	out.Model = StrPtr("gemini-3.1-flash-image-preview")
	out.Backend = StrPtr("google_ai_studio")
	out.Images = []ImageItem{
		{
			Index:              0,
			Path:               "/abs/path/imgraft-20260324-153012-001.png",
			Filename:           "imgraft-20260324-153012-001.png",
			Width:              1024,
			Height:             1024,
			MimeType:           "image/png",
			SHA256:             "e3b0c44298fc1c149afbf4c8996fb924",
			TransparentApplied: true,
		},
	}
	out.Warnings = []string{"fallback from pro to flash"}
	out.RateLimit = RateLimit{
		Provider: StrPtr("google_ai_studio"),
	}

	var buf bytes.Buffer
	if err := Encode(&buf, out, false); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if m["success"] != true {
		t.Errorf("expected success true, got %v", m["success"])
	}
	if m["model"] != "gemini-3.1-flash-image-preview" {
		t.Errorf("expected model, got %v", m["model"])
	}
	if m["backend"] != "google_ai_studio" {
		t.Errorf("expected backend, got %v", m["backend"])
	}

	images := m["images"].([]interface{})
	if len(images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(images))
	}

	img := images[0].(map[string]interface{})
	if img["index"] != float64(0) {
		t.Errorf("expected index 0, got %v", img["index"])
	}
	if img["width"] != float64(1024) {
		t.Errorf("expected width 1024, got %v", img["width"])
	}
	if img["transparent_applied"] != true {
		t.Errorf("expected transparent_applied true, got %v", img["transparent_applied"])
	}

	warnings := m["warnings"].([]interface{})
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0] != "fallback from pro to flash" {
		t.Errorf("expected warning message, got %v", warnings[0])
	}
}

// TestNewSuccessOutput verifies the default state of NewSuccessOutput.
func TestNewSuccessOutput(t *testing.T) {
	out := NewSuccessOutput()

	if out.Success != true {
		t.Errorf("expected Success to be true")
	}
	if out.Model != nil {
		t.Errorf("expected Model to be nil")
	}
	if out.Backend != nil {
		t.Errorf("expected Backend to be nil")
	}
	if out.Images == nil {
		t.Errorf("expected Images to be non-nil empty slice")
	}
	if len(out.Images) != 0 {
		t.Errorf("expected Images to be empty")
	}
	if out.Warnings == nil {
		t.Errorf("expected Warnings to be non-nil empty slice")
	}
	if len(out.Warnings) != 0 {
		t.Errorf("expected Warnings to be empty")
	}
	if out.Error.Code != nil {
		t.Errorf("expected Error.Code to be nil")
	}
	if out.Error.Message != nil {
		t.Errorf("expected Error.Message to be nil")
	}
}

// TestNewErrorOutput verifies the default state of NewErrorOutput.
func TestNewErrorOutput(t *testing.T) {
	out := NewErrorOutput("AUTH_REQUIRED", "api key not found")

	if out.Success != false {
		t.Errorf("expected Success to be false")
	}
	if out.Images == nil {
		t.Errorf("expected Images to be non-nil empty slice")
	}
	if len(out.Images) != 0 {
		t.Errorf("expected Images to be empty")
	}
	if out.Error.Code == nil {
		t.Fatalf("expected Error.Code to be non-nil")
	}
	if *out.Error.Code != "AUTH_REQUIRED" {
		t.Errorf("expected Error.Code to be AUTH_REQUIRED, got %s", *out.Error.Code)
	}
	if out.Error.Message == nil {
		t.Fatalf("expected Error.Message to be non-nil")
	}
	if *out.Error.Message != "api key not found" {
		t.Errorf("expected Error.Message, got %s", *out.Error.Message)
	}
}

// TestHelpers_StrPtr verifies StrPtr returns a pointer to the given string.
func TestHelpers_StrPtr(t *testing.T) {
	p := StrPtr("hello")
	if p == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *p != "hello" {
		t.Errorf("expected hello, got %s", *p)
	}
}

// TestHelpers_IntPtr verifies IntPtr returns a pointer to the given int.
func TestHelpers_IntPtr(t *testing.T) {
	p := IntPtr(42)
	if p == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *p != 42 {
		t.Errorf("expected 42, got %d", *p)
	}
}

// TestEncode_BufferedWrite verifies that Encode does not produce partial
// writes on encoding error (using a valid output to confirm normal behavior).
func TestEncode_BufferedWrite(t *testing.T) {
	out := NewErrorOutput("INTERNAL_ERROR", "something went wrong")
	var buf bytes.Buffer
	if err := Encode(&buf, out, false); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Verify it's valid JSON
	result := buf.Bytes()
	if !json.Valid(result) {
		t.Errorf("expected valid JSON, got: %s", string(result))
	}
}

// TestEncode_TrailingNewline verifies that Encode output ends with a newline.
func TestEncode_TrailingNewline(t *testing.T) {
	out := NewSuccessOutput()
	var buf bytes.Buffer
	if err := Encode(&buf, out, false); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	result := buf.String()
	if !strings.HasSuffix(result, "\n") {
		t.Errorf("expected trailing newline in output")
	}
}
