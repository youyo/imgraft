package auth

import "testing"

func TestMaskAPIKey_EmptyKey(t *testing.T) {
	got := MaskAPIKey("")
	if got != "****" {
		t.Errorf("MaskAPIKey(%q) = %q, want %q", "", got, "****")
	}
}

func TestMaskAPIKey_ShortKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"a", "****"},
		{"ab", "****"},
		{"abc", "****"},
	}
	for _, tt := range tests {
		got := MaskAPIKey(tt.input)
		if got != tt.want {
			t.Errorf("MaskAPIKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMaskAPIKey_ExactlyFour(t *testing.T) {
	got := MaskAPIKey("abcd")
	if got != "****" {
		t.Errorf("MaskAPIKey(%q) = %q, want %q", "abcd", got, "****")
	}
}

func TestMaskAPIKey_FiveChars(t *testing.T) {
	got := MaskAPIKey("abcde")
	want := "****bcde"
	if got != want {
		t.Errorf("MaskAPIKey(%q) = %q, want %q", "abcde", got, want)
	}
}

func TestMaskAPIKey_NormalKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"xxxxxabcd", "****abcd"},
		{"AIzaSyTest1234", "****1234"},
		{"AIzaSyDk0123456789ABCDEF", "****CDEF"},
	}
	for _, tt := range tests {
		got := MaskAPIKey(tt.input)
		if got != tt.want {
			t.Errorf("MaskAPIKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestHasBackend_GoogleAIStudio(t *testing.T) {
	pc := ProfileCredentials{
		GoogleAIStudio: &GoogleAIStudioCredentials{APIKey: "key"},
	}
	if !HasBackend(pc, "google_ai_studio") {
		t.Error("HasBackend should return true for google_ai_studio with valid key")
	}
	if HasBackend(pc, "vertex_ai") {
		t.Error("HasBackend should return false for vertex_ai")
	}
}

func TestHasBackend_EmptyAPIKey(t *testing.T) {
	pc := ProfileCredentials{
		GoogleAIStudio: &GoogleAIStudioCredentials{APIKey: ""},
	}
	if HasBackend(pc, "google_ai_studio") {
		t.Error("HasBackend should return false for empty API key")
	}
}

func TestHasBackend_NilBackend(t *testing.T) {
	pc := ProfileCredentials{}
	if HasBackend(pc, "google_ai_studio") {
		t.Error("HasBackend should return false for nil GoogleAIStudio")
	}
}

func TestHasBackend_UnknownBackend(t *testing.T) {
	pc := ProfileCredentials{
		GoogleAIStudio: &GoogleAIStudioCredentials{APIKey: "key"},
	}
	if HasBackend(pc, "unknown_backend") {
		t.Error("HasBackend should return false for unknown backend")
	}
}

func TestGetAPIKey_GoogleAIStudio(t *testing.T) {
	pc := ProfileCredentials{
		GoogleAIStudio: &GoogleAIStudioCredentials{APIKey: "my-key"},
	}
	key, ok := GetAPIKey(pc, "google_ai_studio")
	if !ok {
		t.Fatal("GetAPIKey should return true")
	}
	if key != "my-key" {
		t.Errorf("key = %q, want %q", key, "my-key")
	}
}

func TestGetAPIKey_NotFound(t *testing.T) {
	pc := ProfileCredentials{}
	_, ok := GetAPIKey(pc, "google_ai_studio")
	if ok {
		t.Error("GetAPIKey should return false for nil GoogleAIStudio")
	}
}

func TestGetAPIKey_EmptyKey(t *testing.T) {
	pc := ProfileCredentials{
		GoogleAIStudio: &GoogleAIStudioCredentials{APIKey: ""},
	}
	_, ok := GetAPIKey(pc, "google_ai_studio")
	if ok {
		t.Error("GetAPIKey should return false for empty API key")
	}
}

func TestGetAPIKey_UnknownBackend(t *testing.T) {
	pc := ProfileCredentials{
		GoogleAIStudio: &GoogleAIStudioCredentials{APIKey: "key"},
	}
	_, ok := GetAPIKey(pc, "unknown")
	if ok {
		t.Error("GetAPIKey should return false for unknown backend")
	}
}

func TestAvailableBackends_WithGoogleAIStudio(t *testing.T) {
	pc := ProfileCredentials{
		GoogleAIStudio: &GoogleAIStudioCredentials{APIKey: "key"},
	}
	backends := AvailableBackends(pc)
	if len(backends) != 1 {
		t.Fatalf("AvailableBackends length = %d, want 1", len(backends))
	}
	if backends[0] != "google_ai_studio" {
		t.Errorf("backends[0] = %q, want %q", backends[0], "google_ai_studio")
	}
}

func TestAvailableBackends_Empty(t *testing.T) {
	pc := ProfileCredentials{}
	backends := AvailableBackends(pc)
	if len(backends) != 0 {
		t.Errorf("AvailableBackends length = %d, want 0", len(backends))
	}
}
