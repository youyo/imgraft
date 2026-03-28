// Package auth はログイン・ログアウト・whoami の認証コマンドロジックを提供する。
package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/youyo/imgraft/internal/backend/studio"
	"github.com/youyo/imgraft/internal/config"
)

// APIKeyValidator は API key の疎通確認インターフェース。
// テスト時にモック実装を注入できる。
type APIKeyValidator interface {
	ValidateAPIKey(ctx context.Context) error
}

// LoginOptions は Login 関数の入力パラメータ。
// テスタビリティのため io.Reader と Validator を DI する。
type LoginOptions struct {
	// APIKey はフラグ経由の API key。空なら stdin から読み取る。
	APIKey string

	// Profile はフラグ経由の profile 名。空なら stdin から読み取る。
	Profile string

	// CredPath は credentials.json のパス。空なら既定パスを使用。
	CredPath string

	// ConfigPath は config.toml のパス。空なら既定パスを使用。
	ConfigPath string

	// BaseURL は Studio API の base URL。空なら本番 URL を使用。
	// テスト時は httptest.Server の URL を指定する。
	// Validator が設定されている場合は無視される。
	BaseURL string

	// Validator は API key 疎通確認の実装。nil なら Studio HTTPClient を使用。
	// テスト時にモックを注入する。
	Validator APIKeyValidator

	// Reader は stdin の代替。対話入力をテスト可能にするための DI。
	Reader *bufio.Reader
}

// LoginResult は Login の結果。
type LoginResult struct {
	Profile string
	Backend string
}

// Login は認証ログインフローを実行する。
// 対話フロー:
//  1. profile 名入力（--profile フラグがあればスキップ）
//  2. API key 入力（--api-key フラグがあればスキップ）
//  3. 疎通確認（models.list）
//  4. 成功時のみ credentials.json に保存
//  5. config.toml の current_profile / last_used_profile / last_used_backend を更新
func Login(ctx context.Context, opts LoginOptions) (LoginResult, error) {
	profile := opts.Profile
	apiKey := opts.APIKey

	// profile 名の解決
	if profile == "" {
		var err error
		profile, err = promptInput(opts.Reader, "Profile name [default]: ")
		if err != nil {
			return LoginResult{}, fmt.Errorf("read profile: %w", err)
		}
		if profile == "" {
			profile = config.DefaultProfile
		}
	}

	// API key の解決
	if apiKey == "" {
		var err error
		apiKey, err = promptInput(opts.Reader, "Google AI Studio API key: ")
		if err != nil {
			return LoginResult{}, fmt.Errorf("read api key: %w", err)
		}
	}

	if apiKey == "" {
		return LoginResult{}, errors.New("API key must not be empty")
	}

	// 疎通確認
	validator := opts.Validator
	if validator == nil {
		validator = buildValidator(apiKey, opts.BaseURL)
	}
	if err := validator.ValidateAPIKey(ctx); err != nil {
		return LoginResult{}, fmt.Errorf("API key validation failed: %w", err)
	}

	// credentials 読み込み
	creds, err := Load(opts.CredPath)
	if err != nil {
		return LoginResult{}, fmt.Errorf("load credentials: %w", err)
	}

	// credentials 更新
	pc := creds.Profiles[profile]
	pc.GoogleAIStudio = &GoogleAIStudioCredentials{APIKey: apiKey}
	creds.Profiles[profile] = pc

	// credentials 保存
	if err := Save(creds, opts.CredPath); err != nil {
		return LoginResult{}, fmt.Errorf("save credentials: %w", err)
	}

	// config 更新
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return LoginResult{}, fmt.Errorf("load config: %w", err)
	}
	cfg.CurrentProfile = profile
	cfg.LastUsedProfile = profile
	cfg.LastUsedBackend = config.DefaultBackend
	if err := config.Save(cfg, opts.ConfigPath); err != nil {
		return LoginResult{}, fmt.Errorf("save config: %w", err)
	}

	return LoginResult{
		Profile: profile,
		Backend: config.DefaultBackend,
	}, nil
}

// promptInput は reader から1行読み取り、末尾の改行を除去して返す。
func promptInput(reader *bufio.Reader, _ string) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// buildValidator は apiKey と baseURL から APIKeyValidator を構築する。
// baseURL が空なら本番 URL（studio.BaseURL）を使う。
func buildValidator(apiKey, baseURL string) APIKeyValidator {
	if baseURL == "" {
		return studio.New(apiKey)
	}
	httpClient := &http.Client{Timeout: 30 * time.Second}
	return studio.NewWithBaseURL(apiKey, baseURL, httpClient)
}
