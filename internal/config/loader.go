package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/youyo/imgraft/internal/runtime"
)

// Load は configPath から TOML を読み込み Config を返す。
// configPath が空文字の場合は runtime.ConfigFilePath() を使用する。
// ファイルが存在しない場合はデフォルト値を返す（エラーにしない）。
func Load(configPath string) (*Config, error) {
	path, err := resolvePath(configPath)
	if err != nil {
		return nil, fmt.Errorf("config path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

// resolvePath は configPath が空なら runtime.ConfigFilePath() を返す。
func resolvePath(configPath string) (string, error) {
	if configPath != "" {
		return configPath, nil
	}
	return runtime.ConfigFilePath()
}

// applyDefaults は Config の欠損フィールドにデフォルト値を補完する。
func applyDefaults(cfg *Config) {
	if cfg.CurrentProfile == "" {
		cfg.CurrentProfile = DefaultProfile
	}
	if cfg.LastUsedProfile == "" {
		cfg.LastUsedProfile = DefaultProfile
	}
	if cfg.LastUsedBackend == "" {
		cfg.LastUsedBackend = DefaultBackend
	}
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = DefaultModelAlias
	}
	if cfg.DefaultOutputDir == "" {
		cfg.DefaultOutputDir = DefaultOutputDir
	}
	if cfg.Models == nil {
		cfg.Models = make(map[string]string)
	}
	if cfg.Models["flash"] == "" {
		cfg.Models["flash"] = BuiltinFlashModel
	}
	if cfg.Models["pro"] == "" {
		cfg.Models["pro"] = BuiltinProModel
	}
}
