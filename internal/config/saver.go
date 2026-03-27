package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Save は Config を TOML 形式で configPath に保存する。
// configPath が空文字の場合は runtime.ConfigFilePath() を使用する。
// ディレクトリが存在しない場合は自動作成する。
func Save(cfg *Config, configPath string) error {
	path, err := resolvePath(configPath)
	if err != nil {
		return fmt.Errorf("config path: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
