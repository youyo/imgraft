package config

import (
	"reflect"
	"testing"
)

// TestDefaultConfig_AllFields は DefaultConfig() が全フィールドを正しく返すことを確認する。
func TestDefaultConfig_AllFields(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.CurrentProfile != DefaultProfile {
		t.Errorf("CurrentProfile: got %q, want %q", cfg.CurrentProfile, DefaultProfile)
	}
	if cfg.LastUsedProfile != DefaultProfile {
		t.Errorf("LastUsedProfile: got %q, want %q", cfg.LastUsedProfile, DefaultProfile)
	}
	if cfg.LastUsedBackend != DefaultBackend {
		t.Errorf("LastUsedBackend: got %q, want %q", cfg.LastUsedBackend, DefaultBackend)
	}
	if cfg.DefaultModel != DefaultModelAlias {
		t.Errorf("DefaultModel: got %q, want %q", cfg.DefaultModel, DefaultModelAlias)
	}
	if cfg.DefaultOutputDir != DefaultOutputDir {
		t.Errorf("DefaultOutputDir: got %q, want %q", cfg.DefaultOutputDir, DefaultOutputDir)
	}
	if cfg.Models == nil {
		t.Fatal("Models is nil")
	}
	if cfg.Models["flash"] != BuiltinFlashModel {
		t.Errorf("Models[flash]: got %q, want %q", cfg.Models["flash"], BuiltinFlashModel)
	}
	if cfg.Models["pro"] != BuiltinProModel {
		t.Errorf("Models[pro]: got %q, want %q", cfg.Models["pro"], BuiltinProModel)
	}
}

// TestConfig_TOMLTags は Config の各フィールドに期待する toml タグが存在することを確認する。
func TestConfig_TOMLTags(t *testing.T) {
	expectedTags := map[string]string{
		"CurrentProfile":   "current_profile",
		"LastUsedProfile":  "last_used_profile",
		"LastUsedBackend":  "last_used_backend",
		"DefaultModel":     "default_model",
		"DefaultOutputDir": "default_output_dir",
		"Models":           "models",
	}

	rt := reflect.TypeOf(Config{})
	for fieldName, wantTag := range expectedTags {
		field, ok := rt.FieldByName(fieldName)
		if !ok {
			t.Errorf("field %q not found in Config", fieldName)
			continue
		}
		gotTag := field.Tag.Get("toml")
		if gotTag != wantTag {
			t.Errorf("field %q: toml tag = %q, want %q", fieldName, gotTag, wantTag)
		}
	}
}
