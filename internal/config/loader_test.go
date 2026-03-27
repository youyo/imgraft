package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoad_FullTOML は完全な config.toml が正しくデコードされることを確認する。
func TestLoad_FullTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
current_profile = "myprofile"
last_used_profile = "myprofile"
last_used_backend = "google_ai_studio"
default_model = "pro"
default_output_dir = "/tmp/output"

[models]
flash = "gemini-custom-flash"
pro = "gemini-custom-pro"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.CurrentProfile != "myprofile" {
		t.Errorf("CurrentProfile = %q, want %q", cfg.CurrentProfile, "myprofile")
	}
	if cfg.LastUsedProfile != "myprofile" {
		t.Errorf("LastUsedProfile = %q, want %q", cfg.LastUsedProfile, "myprofile")
	}
	if cfg.LastUsedBackend != "google_ai_studio" {
		t.Errorf("LastUsedBackend = %q, want %q", cfg.LastUsedBackend, "google_ai_studio")
	}
	if cfg.DefaultModel != "pro" {
		t.Errorf("DefaultModel = %q, want %q", cfg.DefaultModel, "pro")
	}
	if cfg.DefaultOutputDir != "/tmp/output" {
		t.Errorf("DefaultOutputDir = %q, want %q", cfg.DefaultOutputDir, "/tmp/output")
	}
	if cfg.Models["flash"] != "gemini-custom-flash" {
		t.Errorf("Models[flash] = %q, want %q", cfg.Models["flash"], "gemini-custom-flash")
	}
	if cfg.Models["pro"] != "gemini-custom-pro" {
		t.Errorf("Models[pro] = %q, want %q", cfg.Models["pro"], "gemini-custom-pro")
	}
}

// TestLoad_FileNotExist はファイルが存在しない場合にデフォルト値を返すことを確認する。
func TestLoad_FileNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.toml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	want := DefaultConfig()
	if cfg.CurrentProfile != want.CurrentProfile {
		t.Errorf("CurrentProfile = %q, want %q", cfg.CurrentProfile, want.CurrentProfile)
	}
	if cfg.DefaultModel != want.DefaultModel {
		t.Errorf("DefaultModel = %q, want %q", cfg.DefaultModel, want.DefaultModel)
	}
	if cfg.Models["flash"] != want.Models["flash"] {
		t.Errorf("Models[flash] = %q, want %q", cfg.Models["flash"], want.Models["flash"])
	}
}

// TestLoad_EmptyFile は空ファイルでデフォルト値が補完されることを確認する。
func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := DefaultConfig()
	if cfg.CurrentProfile != want.CurrentProfile {
		t.Errorf("CurrentProfile = %q, want %q", cfg.CurrentProfile, want.CurrentProfile)
	}
	if cfg.DefaultModel != want.DefaultModel {
		t.Errorf("DefaultModel = %q, want %q", cfg.DefaultModel, want.DefaultModel)
	}
	if cfg.DefaultOutputDir != want.DefaultOutputDir {
		t.Errorf("DefaultOutputDir = %q, want %q", cfg.DefaultOutputDir, want.DefaultOutputDir)
	}
	if cfg.Models["flash"] != want.Models["flash"] {
		t.Errorf("Models[flash] = %q, want %q", cfg.Models["flash"], want.Models["flash"])
	}
	if cfg.Models["pro"] != want.Models["pro"] {
		t.Errorf("Models[pro] = %q, want %q", cfg.Models["pro"], want.Models["pro"])
	}
}

// TestLoad_PartialTOML は部分的な TOML でデフォルト値が補完されることを確認する。
func TestLoad_PartialTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `current_profile = "custom"`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 指定したフィールドはそのまま
	if cfg.CurrentProfile != "custom" {
		t.Errorf("CurrentProfile = %q, want %q", cfg.CurrentProfile, "custom")
	}
	// 他はデフォルト値
	if cfg.DefaultModel != DefaultModelAlias {
		t.Errorf("DefaultModel = %q, want %q", cfg.DefaultModel, DefaultModelAlias)
	}
	if cfg.DefaultOutputDir != DefaultOutputDir {
		t.Errorf("DefaultOutputDir = %q, want %q", cfg.DefaultOutputDir, DefaultOutputDir)
	}
}

// TestLoad_InvalidTOML は不正な TOML でエラーを返すことを確認する。
func TestLoad_InvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	if err := os.WriteFile(path, []byte("[[invalid toml {{"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("Load() error = nil, want non-nil")
	}
}

// TestLoad_PartialModels は models セクションの部分指定でデフォルト補完されることを確認する。
func TestLoad_PartialModels(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[models]
flash = "custom-flash-model"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// flash は設定値
	if cfg.Models["flash"] != "custom-flash-model" {
		t.Errorf("Models[flash] = %q, want %q", cfg.Models["flash"], "custom-flash-model")
	}
	// pro はデフォルト値
	if cfg.Models["pro"] != BuiltinProModel {
		t.Errorf("Models[pro] = %q, want %q", cfg.Models["pro"], BuiltinProModel)
	}
}

// TestLoad_CustomPath はカスタムパスでの読み込みが動作することを確認する。
func TestLoad_CustomPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom", "myconfig.toml")

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}

	content := `current_profile = "testprofile"`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.CurrentProfile != "testprofile" {
		t.Errorf("CurrentProfile = %q, want %q", cfg.CurrentProfile, "testprofile")
	}
}

// TestLoad_EmptyConfigPath は configPath が空の場合に runtime.ConfigFilePath() を使用することを確認する。
// ファイルが存在しない場合はデフォルト値が返される。
func TestLoad_EmptyConfigPath(t *testing.T) {
	// HOME を一時ディレクトリに向けることでファイル不在を保証
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\") error = %v", err)
	}

	want := DefaultConfig()
	if cfg.CurrentProfile != want.CurrentProfile {
		t.Errorf("CurrentProfile = %q, want %q", cfg.CurrentProfile, want.CurrentProfile)
	}
}
