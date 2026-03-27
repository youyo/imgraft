package model

import (
	"testing"

	"github.com/youyo/imgraft/internal/config"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		cfg       *config.Config
		wantModel string
	}{
		// N01: --model 未指定時に default_model を使う
		{
			name:  "N01_empty_alias_flash_default",
			alias: "",
			cfg: &config.Config{
				DefaultModel: "flash",
				Models:       map[string]string{"flash": "gemini-3.1-flash-image-preview"},
			},
			wantModel: "gemini-3.1-flash-image-preview",
		},
		// N02: flash alias 解決
		{
			name:  "N02_flash_alias",
			alias: "flash",
			cfg: &config.Config{
				Models: map[string]string{"flash": "gemini-3.1-flash-image-preview"},
			},
			wantModel: "gemini-3.1-flash-image-preview",
		},
		// N03: pro alias 解決
		{
			name:  "N03_pro_alias",
			alias: "pro",
			cfg: &config.Config{
				Models: map[string]string{"pro": "gemini-3-pro-image-preview"},
			},
			wantModel: "gemini-3-pro-image-preview",
		},
		// N04: フルモデル名直接指定
		{
			name:  "N04_full_model_name",
			alias: "gemini-3.1-flash-image-preview",
			cfg: &config.Config{
				Models: map[string]string{"flash": "some-other-model"},
			},
			wantModel: "gemini-3.1-flash-image-preview",
		},
		// N05: 空 config.Models 時は built-in fallback (flash)
		{
			name:  "N05_empty_models_flash",
			alias: "flash",
			cfg: &config.Config{
				Models: map[string]string{},
			},
			wantModel: "gemini-3.1-flash-image-preview",
		},
		// N06: 空 config.Models 時は built-in fallback (pro)
		{
			name:  "N06_empty_models_pro",
			alias: "pro",
			cfg: &config.Config{
				Models: map[string]string{},
			},
			wantModel: "gemini-3-pro-image-preview",
		},
		// N07: DefaultModel も空なら built-in flash
		{
			name:  "N07_empty_default_model",
			alias: "",
			cfg: &config.Config{
				DefaultModel: "",
				Models:       map[string]string{},
			},
			wantModel: "gemini-3.1-flash-image-preview",
		},
		// C01: case-sensitive、大文字は alias 扱いしない
		{
			name:  "C01_uppercase_flash",
			alias: "FLASH",
			cfg: &config.Config{
				Models: map[string]string{"flash": "gemini-3.1-flash-image-preview"},
			},
			wantModel: "FLASH",
		},
		// C02: trim しない（呼び出し側の責務）
		{
			name:  "C02_whitespace_flash",
			alias: " flash ",
			cfg: &config.Config{
				Models: map[string]string{"flash": "gemini-3.1-flash-image-preview"},
			},
			wantModel: " flash ",
		},
		// C03: default_model が pro の場合
		{
			name:  "C03_default_model_pro",
			alias: "",
			cfg: &config.Config{
				DefaultModel: "pro",
				Models:       map[string]string{"pro": "gemini-3-pro-image-preview"},
			},
			wantModel: "gemini-3-pro-image-preview",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Resolve(tc.alias, tc.cfg)
			if got != tc.wantModel {
				t.Errorf("Resolve(%q) = %q, want %q", tc.alias, got, tc.wantModel)
			}
		})
	}
}

func TestResolveNilModelsMap(t *testing.T) {
	// cfg.Models が nil でも panic しないことを確認
	cfg := &config.Config{
		DefaultModel: "flash",
		Models:       nil,
	}
	got := Resolve("flash", cfg)
	if got != BuiltinFlashModel {
		t.Errorf("Resolve with nil Models = %q, want %q", got, BuiltinFlashModel)
	}
}

func TestBuiltinDefaults(t *testing.T) {
	if got, want := BuiltinFlashModel, "gemini-3.1-flash-image-preview"; got != want {
		t.Errorf("BuiltinFlashModel = %q, want %q", got, want)
	}
	if got, want := BuiltinProModel, "gemini-3-pro-image-preview"; got != want {
		t.Errorf("BuiltinProModel = %q, want %q", got, want)
	}
	if got, want := AliasFlash, "flash"; got != want {
		t.Errorf("AliasFlash = %q, want %q", got, want)
	}
	if got, want := AliasPro, "pro"; got != want {
		t.Errorf("AliasPro = %q, want %q", got, want)
	}
}
