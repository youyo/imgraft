package auth

import (
	"encoding/json"
	"testing"
)

func TestDefaultCredentials(t *testing.T) {
	creds := DefaultCredentials()
	if creds == nil {
		t.Fatal("DefaultCredentials() returned nil")
	}
	if creds.Profiles == nil {
		t.Fatal("Profiles map is nil, expected empty map")
	}
	if len(creds.Profiles) != 0 {
		t.Errorf("Profiles map length = %d, want 0", len(creds.Profiles))
	}
}

func TestProfileCredentials_GoogleAIStudio(t *testing.T) {
	pc := ProfileCredentials{
		GoogleAIStudio: &GoogleAIStudioCredentials{APIKey: "test-key"},
	}
	if pc.GoogleAIStudio == nil {
		t.Fatal("GoogleAIStudio is nil")
	}
	if pc.GoogleAIStudio.APIKey != "test-key" {
		t.Errorf("APIKey = %q, want %q", pc.GoogleAIStudio.APIKey, "test-key")
	}
}

func TestCredentials_JSONMarshal(t *testing.T) {
	creds := &Credentials{
		Profiles: map[string]ProfileCredentials{
			"default": {
				GoogleAIStudio: &GoogleAIStudioCredentials{APIKey: "my-key"},
			},
		},
	}

	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Credentials
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	pc, ok := decoded.Profiles["default"]
	if !ok {
		t.Fatal("profile 'default' not found after round-trip")
	}
	if pc.GoogleAIStudio == nil {
		t.Fatal("GoogleAIStudio is nil after round-trip")
	}
	if pc.GoogleAIStudio.APIKey != "my-key" {
		t.Errorf("APIKey = %q, want %q", pc.GoogleAIStudio.APIKey, "my-key")
	}
}

func TestCredentials_OmitEmpty(t *testing.T) {
	creds := &Credentials{
		Profiles: map[string]ProfileCredentials{
			"empty": {},
		},
	}

	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// GoogleAIStudio が nil なので JSON に "google_ai_studio" キーが出現しないことを確認
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal to map failed: %v", err)
	}

	profiles := raw["profiles"].(map[string]interface{})
	emptyProfile := profiles["empty"].(map[string]interface{})
	if _, exists := emptyProfile["google_ai_studio"]; exists {
		t.Error("google_ai_studio should be omitted when nil")
	}
}
