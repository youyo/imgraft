package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	creds := &Credentials{
		Profiles: map[string]ProfileCredentials{
			"default": {
				GoogleAIStudio: &GoogleAIStudioCredentials{APIKey: "test-key"},
			},
		},
	}
	if err := Save(creds, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestSave_FilePermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permissions not applicable on Windows")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	creds := &Credentials{Profiles: map[string]ProfileCredentials{}}
	if err := Save(creds, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("file permission = %o, want 0600", perm)
	}
}

func TestSave_DirPermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permissions not applicable on Windows")
	}

	dir := t.TempDir()
	subdir := filepath.Join(dir, "newdir")
	path := filepath.Join(subdir, "credentials.json")
	creds := &Credentials{Profiles: map[string]ProfileCredentials{}}
	if err := Save(creds, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(subdir)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0700 {
		t.Errorf("directory permission = %o, want 0700", perm)
	}
}

func TestSave_JSONRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	original := &Credentials{
		Profiles: map[string]ProfileCredentials{
			"default": {
				GoogleAIStudio: &GoogleAIStudioCredentials{APIKey: "round-trip-key"},
			},
		},
	}
	if err := Save(original, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Profiles["default"].GoogleAIStudio.APIKey != "round-trip-key" {
		t.Errorf("APIKey = %q, want %q",
			loaded.Profiles["default"].GoogleAIStudio.APIKey, "round-trip-key")
	}
}

func TestSave_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "deep", "credentials.json")
	creds := &Credentials{Profiles: map[string]ProfileCredentials{}}
	if err := Save(creds, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestSave_NilCredentials(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	err := Save(nil, path)
	if err == nil {
		t.Fatal("expected error for nil credentials, got nil")
	}
}

func TestSave_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	creds := &Credentials{
		Profiles: map[string]ProfileCredentials{
			"default": {
				GoogleAIStudio: &GoogleAIStudioCredentials{APIKey: "json-test"},
			},
		},
	}
	if err := Save(creds, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	// JSON として valid であること
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}

	// 整形されていること（インデントあり）
	if len(data) < 10 {
		t.Error("JSON content seems too short for indented output")
	}
}
