package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Save は creds を credPath に JSON で書き込む。
// credPath が空文字の場合は runtime.CredentialsFilePath() を使用する。
// パーミッションは 0600（オーナーのみ読み書き）。
// 親ディレクトリが存在しない場合は 0700 で作成する。
//
// credentials.json は API key を平文で格納するため、config.toml (0644/0755) とは
// 異なりファイル 0600 / ディレクトリ 0700 でオーナー以外のアクセスを遮断する。
//
// creds が nil の場合はエラーを返す。
func Save(creds *Credentials, credPath string) error {
	if creds == nil {
		return errors.New("credentials must not be nil")
	}

	path, err := resolvePath(credPath)
	if err != nil {
		return fmt.Errorf("credentials path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create credentials dir: %w", err)
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}
	return nil
}
