package runtime

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDirName       = ".config"
	appName             = "imgraft"
	configFileName      = "config.toml"
	credentialsFileName = "credentials.json"
)

// ConfigDir は ~/.config/imgraft/ の絶対パスを返す。
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	return filepath.Join(home, configDirName, appName), nil
}

// ConfigFilePath は ~/.config/imgraft/config.toml の絶対パスを返す。
func ConfigFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// CredentialsFilePath は ~/.config/imgraft/credentials.json の絶対パスを返す。
func CredentialsFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, credentialsFileName), nil
}
