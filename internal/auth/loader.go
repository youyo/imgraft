package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/youyo/imgraft/internal/runtime"
)

// Load は credPath から JSON を読み込み Credentials を返す。
// credPath が空文字の場合は runtime.CredentialsFilePath() を使用する。
// ファイルが存在しない場合はデフォルト値を返す（エラーにしない）。
func Load(credPath string) (*Credentials, error) {
	path, err := resolvePath(credPath)
	if err != nil {
		return nil, fmt.Errorf("credentials path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultCredentials(), nil
		}
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	if creds.Profiles == nil {
		creds.Profiles = make(map[string]ProfileCredentials)
	}
	return &creds, nil
}

// resolvePath は credPath が空なら runtime.CredentialsFilePath() を返す。
func resolvePath(credPath string) (string, error) {
	if credPath != "" {
		return credPath, nil
	}
	return runtime.CredentialsFilePath()
}
