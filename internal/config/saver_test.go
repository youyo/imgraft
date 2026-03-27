package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSave_RoundTrip は Save → Load で同一の Config が得られることを確認する。
func TestSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	original := DefaultConfig()
	if err := Save(original, path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.CurrentProfile != original.CurrentProfile {
		t.Errorf("CurrentProfile: got %q, want %q", loaded.CurrentProfile, original.CurrentProfile)
	}
	if loaded.LastUsedProfile != original.LastUsedProfile {
		t.Errorf("LastUsedProfile: got %q, want %q", loaded.LastUsedProfile, original.LastUsedProfile)
	}
	if loaded.LastUsedBackend != original.LastUsedBackend {
		t.Errorf("LastUsedBackend: got %q, want %q", loaded.LastUsedBackend, original.LastUsedBackend)
	}
	if loaded.DefaultModel != original.DefaultModel {
		t.Errorf("DefaultModel: got %q, want %q", loaded.DefaultModel, original.DefaultModel)
	}
	if loaded.DefaultOutputDir != original.DefaultOutputDir {
		t.Errorf("DefaultOutputDir: got %q, want %q", loaded.DefaultOutputDir, original.DefaultOutputDir)
	}
	if loaded.Models["flash"] != original.Models["flash"] {
		t.Errorf("Models[flash]: got %q, want %q", loaded.Models["flash"], original.Models["flash"])
	}
	if loaded.Models["pro"] != original.Models["pro"] {
		t.Errorf("Models[pro]: got %q, want %q", loaded.Models["pro"], original.Models["pro"])
	}
}

// TestSave_AutoCreateDir はディレクトリが存在しない場合に自動作成されることを確認する。
func TestSave_AutoCreateDir(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "subdir", "nested", "config.toml")

	if err := Save(DefaultConfig(), path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

// TestSave_FilePermission は保存ファイルのパーミッションが 0644 であることを確認する。
func TestSave_FilePermission(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	if err := Save(DefaultConfig(), path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}

	got := info.Mode() & 0777
	want := os.FileMode(0644)
	if got != want {
		t.Errorf("file permission = %o, want %o", got, want)
	}
}

// TestSave_Overwrite は既存ファイルが上書きされることを確認する。
func TestSave_Overwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	// 1回目の保存
	first := DefaultConfig()
	first.CurrentProfile = "first"
	if err := Save(first, path); err != nil {
		t.Fatalf("Save() 1st error = %v", err)
	}

	// 2回目の保存（上書き）
	second := DefaultConfig()
	second.CurrentProfile = "second"
	if err := Save(second, path); err != nil {
		t.Fatalf("Save() 2nd error = %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.CurrentProfile != "second" {
		t.Errorf("CurrentProfile = %q, want %q", loaded.CurrentProfile, "second")
	}
}

// TestSave_CustomPath はカスタムパスでの保存が動作することを確認する。
func TestSave_CustomPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "myconfig.toml")

	if err := Save(DefaultConfig(), path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not found at custom path: %v", err)
	}
}

// TestSave_ModelsSection は Models マップが [models] セクションとして出力されることを確認する。
func TestSave_ModelsSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	cfg := DefaultConfig()
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "[models]") {
		t.Errorf("TOML output does not contain [models] section:\n%s", content)
	}
}
