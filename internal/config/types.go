package config

const (
	// DefaultProfile は初期 profile 名。
	DefaultProfile = "default"

	// DefaultBackend は v1 のデフォルト backend。
	DefaultBackend = "google_ai_studio"

	// DefaultModelAlias は未指定時のデフォルトモデル alias。
	DefaultModelAlias = "flash"

	// DefaultOutputDir は未指定時の出力ディレクトリ。
	DefaultOutputDir = "."

	// BuiltinFlashModel は built-in fallback の flash モデル名。
	BuiltinFlashModel = "gemini-3.1-flash-image-preview"

	// BuiltinProModel は built-in fallback の pro モデル名。
	BuiltinProModel = "gemini-3-pro-image-preview"
)

// Config は ~/.config/imgraft/config.toml の構造体表現。
type Config struct {
	CurrentProfile   string            `toml:"current_profile"`
	LastUsedProfile  string            `toml:"last_used_profile"`
	LastUsedBackend  string            `toml:"last_used_backend"`
	DefaultModel     string            `toml:"default_model"`
	DefaultOutputDir string            `toml:"default_output_dir"`
	Models           map[string]string `toml:"models"`
}

// DefaultConfig はすべてのフィールドが既定値で埋まった Config を返す。
func DefaultConfig() *Config {
	return &Config{
		CurrentProfile:   DefaultProfile,
		LastUsedProfile:  DefaultProfile,
		LastUsedBackend:  DefaultBackend,
		DefaultModel:     DefaultModelAlias,
		DefaultOutputDir: DefaultOutputDir,
		Models: map[string]string{
			"flash": BuiltinFlashModel,
			"pro":   BuiltinProModel,
		},
	}
}
