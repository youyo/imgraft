package runtime_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/youyo/imgraft/internal/runtime"
)

func TestConfigDir(t *testing.T) {
	t.Setenv("HOME", "/tmp/testhome")

	dir, err := runtime.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}

	want := filepath.Join("/tmp/testhome", ".config", "imgraft")
	if dir != want {
		t.Errorf("ConfigDir() = %q, want %q", dir, want)
	}
}

func TestConfigDir_NoTrailingSlash(t *testing.T) {
	t.Setenv("HOME", "/tmp/testhome")

	dir, err := runtime.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}

	if strings.HasSuffix(dir, "/") {
		t.Errorf("ConfigDir() = %q, should not end with /", dir)
	}
}

func TestConfigFilePath(t *testing.T) {
	t.Setenv("HOME", "/tmp/testhome")

	path, err := runtime.ConfigFilePath()
	if err != nil {
		t.Fatalf("ConfigFilePath() error = %v", err)
	}

	want := filepath.Join("/tmp/testhome", ".config", "imgraft", "config.toml")
	if path != want {
		t.Errorf("ConfigFilePath() = %q, want %q", path, want)
	}
}

func TestCredentialsFilePath(t *testing.T) {
	t.Setenv("HOME", "/tmp/testhome")

	path, err := runtime.CredentialsFilePath()
	if err != nil {
		t.Fatalf("CredentialsFilePath() error = %v", err)
	}

	want := filepath.Join("/tmp/testhome", ".config", "imgraft", "credentials.json")
	if path != want {
		t.Errorf("CredentialsFilePath() = %q, want %q", path, want)
	}
}
