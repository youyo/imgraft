package output

import (
	"encoding/json"
	"testing"
)

// TestOutputSchema_FieldsAlwaysPresent verifies that all top-level fields
// are always present in the JSON output, even when they hold zero/null values.
func TestOutputSchema_FieldsAlwaysPresent(t *testing.T) {
	out := NewSuccessOutput()
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	requiredKeys := []string{
		"success", "model", "backend", "images",
		"rate_limit", "warnings", "error",
	}
	for _, key := range requiredKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("expected key %q to be present in JSON output", key)
		}
	}
}

// TestOutputSchema_NullFields verifies that pointer fields are null in JSON
// when they are nil in Go.
func TestOutputSchema_NullFields(t *testing.T) {
	out := NewSuccessOutput()
	// Model and Backend should be nil by default
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// model should be null (nil in map)
	if m["model"] != nil {
		t.Errorf("expected model to be null, got %v", m["model"])
	}
	// backend should be null
	if m["backend"] != nil {
		t.Errorf("expected backend to be null, got %v", m["backend"])
	}
	// error.code and error.message should be null
	errorMap, ok := m["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected error to be an object, got %T", m["error"])
	}
	if errorMap["code"] != nil {
		t.Errorf("expected error.code to be null, got %v", errorMap["code"])
	}
	if errorMap["message"] != nil {
		t.Errorf("expected error.message to be null, got %v", errorMap["message"])
	}
}

// TestOutputSchema_ErrorFieldsAlwaysPresent verifies that error object fields
// are always present even when null.
func TestOutputSchema_ErrorFieldsAlwaysPresent(t *testing.T) {
	out := NewSuccessOutput()
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	errorMap, ok := m["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected error to be an object, got %T", m["error"])
	}

	for _, key := range []string{"code", "message"} {
		if _, ok := errorMap[key]; !ok {
			t.Errorf("expected error.%s to be present in JSON output", key)
		}
	}
}

// TestRateLimit_NullInitial verifies that NewEmptyRateLimit returns a RateLimit
// where all fields serialize to null.
func TestRateLimit_NullInitial(t *testing.T) {
	rl := NewEmptyRateLimit()
	b, err := json.Marshal(rl)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	requiredKeys := []string{
		"provider", "limit_type", "requests_limit",
		"requests_remaining", "requests_used",
		"reset_at", "retry_after_seconds",
	}
	for _, key := range requiredKeys {
		val, ok := m[key]
		if !ok {
			t.Errorf("expected key %q to be present in RateLimit JSON", key)
			continue
		}
		if val != nil {
			t.Errorf("expected %q to be null, got %v", key, val)
		}
	}
}

// TestRateLimit_AllFieldsPresent verifies that a fully populated RateLimit
// serializes all fields.
func TestRateLimit_AllFieldsPresent(t *testing.T) {
	rl := RateLimit{
		Provider:          StrPtr("google_ai_studio"),
		LimitType:         StrPtr("per_minute"),
		RequestsLimit:     IntPtr(60),
		RequestsRemaining: IntPtr(55),
		RequestsUsed:      IntPtr(5),
		ResetAt:           StrPtr("2026-03-28T12:00:00Z"),
		RetryAfterSeconds: IntPtr(30),
	}
	b, err := json.Marshal(rl)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if m["provider"] != "google_ai_studio" {
		t.Errorf("expected provider to be google_ai_studio, got %v", m["provider"])
	}
	// requests_limit should be 60 (JSON numbers are float64)
	if m["requests_limit"] != float64(60) {
		t.Errorf("expected requests_limit to be 60, got %v", m["requests_limit"])
	}
}

// TestImageItem_AllFields verifies that all ImageItem fields are serialized.
func TestImageItem_AllFields(t *testing.T) {
	item := ImageItem{
		Index:              0,
		Path:               "/abs/path/imgraft-20260324-153012-001.png",
		Filename:           "imgraft-20260324-153012-001.png",
		Width:              1024,
		Height:             1024,
		MimeType:           "image/png",
		SHA256:             "abc123",
		TransparentApplied: true,
	}
	b, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	requiredKeys := []string{
		"index", "path", "filename", "width", "height",
		"mime_type", "sha256", "transparent_applied",
	}
	for _, key := range requiredKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("expected key %q to be present in ImageItem JSON", key)
		}
	}

	if m["transparent_applied"] != true {
		t.Errorf("expected transparent_applied to be true, got %v", m["transparent_applied"])
	}
}
