// Package cli は imgraft の CLI 定義を kong を使って提供する。
// このファイルは config サブコマンドの Run ロジックを実装する。
package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/youyo/imgraft/internal/auth"
	"github.com/youyo/imgraft/internal/backend/studio"
	"github.com/youyo/imgraft/internal/config"
	"github.com/youyo/imgraft/internal/model"
)

// ModelLister は models.list API を呼び出すインターフェース。
// テスト時にモック実装を注入できる。
type ModelLister interface {
	ListModels(ctx context.Context) ([]studio.RemoteModel, error)
}

// APIKeyValidator は API key の疎通確認インターフェース（auth パッケージと共通）。
type APIKeyValidator interface {
	ValidateAPIKey(ctx context.Context) error
}

// ConfigInitOptions は RunConfigInit 関数の入力パラメータ。
type ConfigInitOptions struct {
	// APIKey はフラグ経由の API key。空なら stdin から読み取る。
	APIKey string

	// Profile はフラグ経由の profile 名。空なら stdin から読み取る。
	Profile string

	// ConfigPath は config.toml のパス。空なら既定パスを使用。
	ConfigPath string

	// CredPath は credentials.json のパス。空なら既定パスを使用。
	CredPath string

	// BaseURL は Studio API の base URL。空なら本番 URL を使用。
	// Validator/Lister が設定されている場合は無視される。
	BaseURL string

	// Validator は API key 疎通確認の実装。nil なら Studio HTTPClient を使用。
	Validator APIKeyValidator

	// Lister はモデル一覧取得の実装。nil なら Studio HTTPClient を使用。
	Lister ModelLister

	// Reader は stdin の代替。対話入力をテスト可能にするための DI。
	Reader *bufio.Reader
}

// ConfigUseOptions は RunConfigUse 関数の入力パラメータ。
type ConfigUseOptions struct {
	// Profile は切り替える profile 名。
	Profile string

	// ConfigPath は config.toml のパス。空なら既定パスを使用。
	ConfigPath string
}

// ConfigRefreshModelsOptions は RunConfigRefreshModels 関数の入力パラメータ。
type ConfigRefreshModelsOptions struct {
	// ConfigPath は config.toml のパス。空なら既定パスを使用。
	ConfigPath string

	// CredPath は credentials.json のパス。空なら既定パスを使用。
	CredPath string

	// BaseURL は Studio API の base URL。空なら本番 URL を使用。
	// Lister が設定されている場合は無視される。
	BaseURL string

	// Lister はモデル一覧取得の実装。nil なら Studio HTTPClient を使用。
	Lister ModelLister

	// Profile は使用する profile。空なら config の CurrentProfile を使用。
	Profile string
}

// RunConfigInit は config init フローを実行する。
//
// フロー:
//  1. profile 名の解決（フラグ → stdin 対話）
//  2. API key の解決（フラグ → stdin 対話）
//  3. API key 疎通確認（models.list）
//  4. models.list を実行して alias マップを生成
//  5. credentials.json に API key を保存
//  6. config.toml を更新（current_profile, models セクション）
func RunConfigInit(ctx context.Context, opts ConfigInitOptions) error {
	profile := opts.Profile
	apiKey := opts.APIKey

	reader := opts.Reader
	if reader == nil {
		reader = bufio.NewReader(os.Stdin)
	}

	// profile 名の解決
	if profile == "" {
		var err error
		profile, err = promptConfigInput(reader, "Profile name [default]: ")
		if err != nil {
			return fmt.Errorf("read profile: %w", err)
		}
		if profile == "" {
			profile = config.DefaultProfile
		}
	}

	// API key の解決
	if apiKey == "" {
		var err error
		apiKey, err = promptConfigInput(reader, "Google AI Studio API key: ")
		if err != nil {
			return fmt.Errorf("read api key: %w", err)
		}
	}

	if apiKey == "" {
		return fmt.Errorf("API key must not be empty")
	}

	// Studio クライアントの構築（DI なければ本物を使う）
	client := buildStudioClient(apiKey, opts.BaseURL)

	validator := opts.Validator
	if validator == nil {
		validator = client
	}

	lister := opts.Lister
	if lister == nil {
		lister = client
	}

	// 疎通確認
	if err := validator.ValidateAPIKey(ctx); err != nil {
		return fmt.Errorf("API key validation failed: %w", err)
	}

	// models.list を実行してエイリアスマップを生成
	remoteModels, err := lister.ListModels(ctx)
	if err != nil {
		// ListModels が失敗しても続行（組み込みデフォルトを使用）
		remoteModels = nil
	}

	// モデル名リストを抽出
	modelNames := make([]string, 0, len(remoteModels))
	for _, m := range remoteModels {
		if m.SupportedGeneration {
			modelNames = append(modelNames, m.Name)
		}
	}
	aliases := model.ResolveAliasesFromModels(modelNames)

	// credentials 読み込み
	creds, err := auth.Load(opts.CredPath)
	if err != nil {
		return fmt.Errorf("load credentials: %w", err)
	}

	// credentials 更新
	pc := creds.Profiles[profile]
	pc.GoogleAIStudio = &auth.GoogleAIStudioCredentials{APIKey: apiKey}
	creds.Profiles[profile] = pc

	// credentials 保存
	if err := auth.Save(creds, opts.CredPath); err != nil {
		return fmt.Errorf("save credentials: %w", err)
	}

	// config 読み込み
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// config 更新
	cfg.CurrentProfile = profile
	cfg.LastUsedProfile = profile
	cfg.LastUsedBackend = config.DefaultBackend

	// aliases をマージ（空でない場合のみ上書き）
	if cfg.Models == nil {
		cfg.Models = make(map[string]string)
	}
	for alias, fullName := range aliases {
		if fullName != "" {
			cfg.Models[alias] = fullName
		}
	}

	// config 保存
	if err := config.Save(cfg, opts.ConfigPath); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Logged in as profile %q (backend: %s)\n", profile, config.DefaultBackend)
	return nil
}

// RunConfigUse は config use フローを実行する。
// current_profile を指定プロファイルに切り替えて config.toml を保存する。
func RunConfigUse(ctx context.Context, opts ConfigUseOptions) error {
	if opts.Profile == "" {
		return fmt.Errorf("profile name must not be empty")
	}

	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	cfg.CurrentProfile = opts.Profile
	cfg.LastUsedProfile = opts.Profile

	if err := config.Save(cfg, opts.ConfigPath); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Switched to profile %q\n", opts.Profile)
	return nil
}

// RunConfigRefreshModels は config refresh-models フローを実行する。
//
// フロー:
//  1. config.toml から現在の profile と credentials を取得
//  2. API key を credentials から取得
//  3. models.list を実行してエイリアスマップを生成
//  4. config.toml の [models] セクションを上書き保存
func RunConfigRefreshModels(ctx context.Context, opts ConfigRefreshModelsOptions) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	profile := opts.Profile
	if profile == "" {
		profile = cfg.CurrentProfile
	}

	// lister が DI されていない場合は credentials から API key を取得して構築
	lister := opts.Lister
	if lister == nil {
		creds, err := auth.Load(opts.CredPath)
		if err != nil {
			return fmt.Errorf("load credentials: %w", err)
		}

		pc, ok := creds.Profiles[profile]
		if !ok || pc.GoogleAIStudio == nil || pc.GoogleAIStudio.APIKey == "" {
			return fmt.Errorf("no API key found for profile %q; please run `imgraft auth login` or `imgraft config init` first", profile)
		}

		lister = buildStudioClient(pc.GoogleAIStudio.APIKey, opts.BaseURL)
	}

	// models.list 実行
	remoteModels, err := lister.ListModels(ctx)
	if err != nil {
		return fmt.Errorf("list models: %w", err)
	}

	// モデル名リストを抽出
	modelNames := make([]string, 0, len(remoteModels))
	for _, m := range remoteModels {
		if m.SupportedGeneration {
			modelNames = append(modelNames, m.Name)
		}
	}

	// エイリアスマップを生成
	aliases := model.ResolveAliasesFromModels(modelNames)

	// config の [models] セクションを更新
	if cfg.Models == nil {
		cfg.Models = make(map[string]string)
	}
	for alias, fullName := range aliases {
		if fullName != "" {
			cfg.Models[alias] = fullName
		}
	}

	if err := config.Save(cfg, opts.ConfigPath); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Models refreshed: flash=%s, pro=%s\n", cfg.Models["flash"], cfg.Models["pro"])
	return nil
}

// --- 内部ヘルパー ---

// promptConfigInput は reader から1行読み取り、末尾の改行を除去して返す。
func promptConfigInput(reader *bufio.Reader, _ string) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// buildStudioClient は apiKey と baseURL から Studio クライアントを構築する。
func buildStudioClient(apiKey, baseURL string) *studio.HTTPClient {
	if baseURL == "" {
		return studio.New(apiKey)
	}
	httpClient := &http.Client{Timeout: 30 * time.Second}
	return studio.NewWithBaseURL(apiKey, baseURL, httpClient)
}

// runConfigInitInteractive は ConfigInitCmd から呼ばれる対話フロー実行関数。
func runConfigInitInteractive(c *ConfigInitCmd) error {
	return RunConfigInit(context.Background(), ConfigInitOptions{
		APIKey:  c.APIKey,
		Profile: c.Profile,
	})
}

// runConfigUseInteractive は ConfigUseCmd から呼ばれる profile 切り替え実行関数。
func runConfigUseInteractive(c *ConfigUseCmd) error {
	return RunConfigUse(context.Background(), ConfigUseOptions{
		Profile: c.Profile,
	})
}

// runConfigRefreshModelsInteractive は ConfigRefreshModelsCmd から呼ばれる実行関数。
func runConfigRefreshModelsInteractive() error {
	return RunConfigRefreshModels(context.Background(), ConfigRefreshModelsOptions{})
}
