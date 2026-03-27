package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func createTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return path
}

func TestLoad_FileNotExist(t *testing.T) {
	creds, err := Load("/nonexistent/path/credentials.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds == nil {
		t.Fatal("creds is nil")
	}
	if creds.Profiles == nil {
		t.Fatal("Profiles map is nil")
	}
	if len(creds.Profiles) != 0 {
		t.Errorf("Profiles length = %d, want 0", len(creds.Profiles))
	}
}

func TestLoad_ValidJSON(t *testing.T) {
	path := createTempFile(t, `{
		"profiles": {
			"default": {
				"google_ai_studio": {
					"api_key": "test-key-123"
				}
			}
		}
	}`)

	creds, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pc, ok := creds.Profiles["default"]
	if !ok {
		t.Fatal("profile 'default' not found")
	}
	if pc.GoogleAIStudio == nil {
		t.Fatal("GoogleAIStudio is nil")
	}
	if pc.GoogleAIStudio.APIKey != "test-key-123" {
		t.Errorf("APIKey = %q, want %q", pc.GoogleAIStudio.APIKey, "test-key-123")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	path := createTempFile(t, "not-valid-json{{{")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	// 空パスは runtime.CredentialsFilePath() を使う。
	// デフォルトの credentials ファイルが存在しない環境では DefaultCredentials を返す。
	creds, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds == nil {
		t.Fatal("creds is nil")
	}
	if creds.Profiles == nil {
		t.Fatal("Profiles map is nil")
	}
}

func TestLoad_MultipleProfiles(t *testing.T) {
	path := createTempFile(t, `{
		"profiles": {
			"default": {
				"google_ai_studio": {
					"api_key": "key-default"
				}
			},
			"work": {
				"google_ai_studio": {
					"api_key": "key-work"
				}
			}
		}
	}`)

	creds, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(creds.Profiles) != 2 {
		t.Fatalf("Profiles length = %d, want 2", len(creds.Profiles))
	}

	if creds.Profiles["default"].GoogleAIStudio.APIKey != "key-default" {
		t.Errorf("default APIKey = %q, want %q", creds.Profiles["default"].GoogleAIStudio.APIKey, "key-default")
	}
	if creds.Profiles["work"].GoogleAIStudio.APIKey != "key-work" {
		t.Errorf("work APIKey = %q, want %q", creds.Profiles["work"].GoogleAIStudio.APIKey, "key-work")
	}
}

func TestLoad_NilProfilesHandled(t *testing.T) {
	path := createTempFile(t, `{"profiles": null}`)

	creds, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.Profiles == nil {
		t.Fatal("Profiles map is nil, expected empty map")
	}
	if len(creds.Profiles) != 0 {
		t.Errorf("Profiles length = %d, want 0", len(creds.Profiles))
	}
}
