package auth

import (
	"fmt"

	"github.com/youyo/imgraft/internal/config"
)

// LogoutOptions は Logout 関数の入力パラメータ。
type LogoutOptions struct {
	// Profile はログアウト対象の profile 名。空なら current_profile を使用。
	Profile string

	// Backend は削除する backend 名。空なら "google_ai_studio" を使用。
	Backend string

	// CredPath は credentials.json のパス。空なら既定パスを使用。
	CredPath string

	// ConfigPath は config.toml のパス（profile 解決に使用）。空なら既定パスを使用。
	ConfigPath string
}

// Logout は指定 profile の backend 認証情報を削除する。
// profile 自体は削除しない。
// credentials.json が存在しない場合はエラーにしない（ログアウト済みとみなす）。
func Logout(opts LogoutOptions) error {
	profile := opts.Profile
	backend := opts.Backend

	// backend のデフォルト
	if backend == "" {
		backend = config.DefaultBackend
	}

	// profile の解決
	if profile == "" {
		cfg, err := config.Load(opts.ConfigPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		profile = cfg.CurrentProfile
	}

	// credentials の読み込み
	creds, err := Load(opts.CredPath)
	if err != nil {
		return fmt.Errorf("load credentials: %w", err)
	}

	// profile が存在しない場合はそのまま終了
	pc, ok := creds.Profiles[profile]
	if !ok {
		return nil
	}

	// backend の認証情報を削除
	switch backend {
	case "google_ai_studio":
		pc.GoogleAIStudio = nil
	}

	creds.Profiles[profile] = pc

	// 保存
	if err := Save(creds, opts.CredPath); err != nil {
		return fmt.Errorf("save credentials: %w", err)
	}

	return nil
}
