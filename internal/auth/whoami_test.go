package auth_test

import (
	"path/filepath"
	"testing"

	"github.com/youyo/imgraft/internal/auth"
	"github.com/youyo/imgraft/internal/config"
)

// TestWhoami_BasicOutput は whoami の基本出力をテスト。
func TestWhoami_BasicOutput(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	cfgPath := filepath.Join(dir, "config.toml")

	// credentials を保存
	creds := &auth.Credentials{
		Profiles: map[string]auth.ProfileCredentials{
			"default": {
				GoogleAIStudio: &auth.GoogleAIStudioCredentials{APIKey: "testkey1234abcd"},
			},
		},
	}
	if err := auth.Save(creds, credPath); err != nil {
		t.Fatalf("Save credentials failed: %v", err)
	}

	// config を保存
	cfg := config.DefaultConfig()
	cfg.CurrentProfile = "default"
	cfg.LastUsedProfile = "default"
	cfg.LastUsedBackend = "google_ai_studio"
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("Save config failed: %v", err)
	}

	opts := auth.WhoamiOptions{
		CredPath:   credPath,
		ConfigPath: cfgPath,
	}

	result, err := auth.Whoami(opts)
	if err != nil {
		t.Fatalf("Whoami failed: %v", err)
	}

	if result.Profile != "default" {
		t.Errorf("Profile = %q, want %q", result.Profile, "default")
	}
	if result.LastUsedBackend != "google_ai_studio" {
		t.Errorf("LastUsedBackend = %q, want %q", result.LastUsedBackend, "google_ai_studio")
	}
	if len(result.Backends) == 0 {
		t.Error("Backends should not be empty")
	}
}

// TestWhoami_MaskedAPIKey は API key が末尾4文字のみ表示されることをテスト。
func TestWhoami_MaskedAPIKey(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	cfgPath := filepath.Join(dir, "config.toml")

	creds := &auth.Credentials{
		Profiles: map[string]auth.ProfileCredentials{
			"default": {
				GoogleAIStudio: &auth.GoogleAIStudioCredentials{APIKey: "abcdefgh1234"},
			},
		},
	}
	if err := auth.Save(creds, credPath); err != nil {
		t.Fatalf("Save credentials failed: %v", err)
	}

	cfg := config.DefaultConfig()
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("Save config failed: %v", err)
	}

	opts := auth.WhoamiOptions{
		CredPath:   credPath,
		ConfigPath: cfgPath,
	}

	result, err := auth.Whoami(opts)
	if err != nil {
		t.Fatalf("Whoami failed: %v", err)
	}

	for _, b := range result.Backends {
		if b.Name == "google_ai_studio" {
			// 末尾4文字のみ表示: "****1234"
			if b.MaskedKey != "****1234" {
				t.Errorf("MaskedKey = %q, want %q", b.MaskedKey, "****1234")
			}
		}
	}
}

// TestWhoami_NoCredentials は credentials がない場合のテスト。
func TestWhoami_NoCredentials(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "nonexistent.json")
	cfgPath := filepath.Join(dir, "config.toml")

	cfg := config.DefaultConfig()
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("Save config failed: %v", err)
	}

	opts := auth.WhoamiOptions{
		CredPath:   credPath,
		ConfigPath: cfgPath,
	}

	result, err := auth.Whoami(opts)
	if err != nil {
		t.Fatalf("Whoami failed: %v", err)
	}

	// backends は空
	if len(result.Backends) != 0 {
		t.Errorf("Backends = %v, want empty", result.Backends)
	}
}

// TestWhoami_MultipleProfilesShowsCurrent は current_profile のみ表示することをテスト。
func TestWhoami_MultipleProfilesShowsCurrent(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	cfgPath := filepath.Join(dir, "config.toml")

	creds := &auth.Credentials{
		Profiles: map[string]auth.ProfileCredentials{
			"default": {
				GoogleAIStudio: &auth.GoogleAIStudioCredentials{APIKey: "key-default"},
			},
			"work": {
				GoogleAIStudio: &auth.GoogleAIStudioCredentials{APIKey: "key-work"},
			},
		},
	}
	if err := auth.Save(creds, credPath); err != nil {
		t.Fatalf("Save credentials failed: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.CurrentProfile = "work"
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("Save config failed: %v", err)
	}

	opts := auth.WhoamiOptions{
		CredPath:   credPath,
		ConfigPath: cfgPath,
	}

	result, err := auth.Whoami(opts)
	if err != nil {
		t.Fatalf("Whoami failed: %v", err)
	}

	if result.Profile != "work" {
		t.Errorf("Profile = %q, want %q", result.Profile, "work")
	}
	// work profile の API key が表示されるはず
	for _, b := range result.Backends {
		if b.Name == "google_ai_studio" {
			if b.MaskedKey != "****work" {
				t.Errorf("MaskedKey = %q, want %q", b.MaskedKey, "****work")
			}
		}
	}
}
