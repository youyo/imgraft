package auth_test

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/youyo/imgraft/internal/auth"
	"github.com/youyo/imgraft/internal/config"
)

// mockValidator は APIKeyValidator のモック実装。
type mockValidator struct {
	err error
}

func (m *mockValidator) ValidateAPIKey(_ context.Context) error {
	return m.err
}

// TestLogin_Success は正常なログインフローをテスト。
func TestLogin_Success(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	cfgPath := filepath.Join(dir, "config.toml")

	// stdin からの入力をシミュレート: profile名="default", API key="test-key-1234"
	input := "default\ntest-key-1234\n"
	reader := bufio.NewReader(strings.NewReader(input))

	opts := auth.LoginOptions{
		APIKey:     "",
		Profile:    "",
		CredPath:   credPath,
		ConfigPath: cfgPath,
		Validator:  &mockValidator{err: nil},
		Reader:     reader,
	}

	result, err := auth.Login(context.Background(), opts)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if result.Profile != "default" {
		t.Errorf("Profile = %q, want %q", result.Profile, "default")
	}
	if result.Backend != "google_ai_studio" {
		t.Errorf("Backend = %q, want %q", result.Backend, "google_ai_studio")
	}

	// credentials.json が保存されているか確認
	creds, err := auth.Load(credPath)
	if err != nil {
		t.Fatalf("Load credentials failed: %v", err)
	}
	pc := creds.Profiles["default"]
	if pc.GoogleAIStudio == nil || pc.GoogleAIStudio.APIKey != "test-key-1234" {
		t.Errorf("API key not saved correctly")
	}

	// config.toml が更新されているか確認
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Load config failed: %v", err)
	}
	if cfg.CurrentProfile != "default" {
		t.Errorf("CurrentProfile = %q, want %q", cfg.CurrentProfile, "default")
	}
	if cfg.LastUsedBackend != "google_ai_studio" {
		t.Errorf("LastUsedBackend = %q, want %q", cfg.LastUsedBackend, "google_ai_studio")
	}
}

// TestLogin_WithFlags はフラグで profile/api_key を指定した非対話ログインをテスト。
func TestLogin_WithFlags(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	cfgPath := filepath.Join(dir, "config.toml")

	opts := auth.LoginOptions{
		APIKey:     "flag-key-5678",
		Profile:    "myprofile",
		CredPath:   credPath,
		ConfigPath: cfgPath,
		Validator:  &mockValidator{err: nil},
		Reader:     bufio.NewReader(strings.NewReader("")),
	}

	result, err := auth.Login(context.Background(), opts)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if result.Profile != "myprofile" {
		t.Errorf("Profile = %q, want %q", result.Profile, "myprofile")
	}

	creds, err := auth.Load(credPath)
	if err != nil {
		t.Fatalf("Load credentials failed: %v", err)
	}
	if creds.Profiles["myprofile"].GoogleAIStudio.APIKey != "flag-key-5678" {
		t.Errorf("API key not saved correctly")
	}
}

// TestLogin_InvalidKey は疎通確認が失敗した場合のテスト。
func TestLogin_InvalidKey(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	cfgPath := filepath.Join(dir, "config.toml")

	opts := auth.LoginOptions{
		APIKey:     "invalid-key",
		Profile:    "default",
		CredPath:   credPath,
		ConfigPath: cfgPath,
		Validator:  &mockValidator{err: errors.New("401: API key not valid")},
		Reader:     bufio.NewReader(strings.NewReader("")),
	}

	_, err := auth.Login(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for invalid key, got nil")
	}

	// credentials.json は保存されていないはず
	if _, statErr := os.Stat(credPath); !os.IsNotExist(statErr) {
		t.Error("credentials.json should not be saved on failure")
	}
}

// TestLogin_EmptyAPIKey は空の API key が拒否されることをテスト。
func TestLogin_EmptyAPIKey(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	cfgPath := filepath.Join(dir, "config.toml")

	opts := auth.LoginOptions{
		APIKey:     "",
		Profile:    "default",
		CredPath:   credPath,
		ConfigPath: cfgPath,
		Validator:  &mockValidator{err: nil},
		Reader:     bufio.NewReader(strings.NewReader("\n")), // 空行のみ
	}

	_, err := auth.Login(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
}

// TestLogin_DefaultProfileFromInput は対話でプロファイル未入力時に "default" が使われることをテスト。
func TestLogin_DefaultProfileFromInput(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	cfgPath := filepath.Join(dir, "config.toml")

	// profile 名を空（エンター）で入力し、デフォルトに戻ることをテスト
	input := "\nmy-api-key\n"
	reader := bufio.NewReader(strings.NewReader(input))

	opts := auth.LoginOptions{
		APIKey:     "",
		Profile:    "",
		CredPath:   credPath,
		ConfigPath: cfgPath,
		Validator:  &mockValidator{err: nil},
		Reader:     reader,
	}

	result, err := auth.Login(context.Background(), opts)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if result.Profile != "default" {
		t.Errorf("Profile = %q, want %q", result.Profile, "default")
	}
}
