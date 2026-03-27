// Package app は imgraft のメイン生成パイプラインを実装する。
// CLI parse → Config load → Auth → Model → Prompt → API → Save → JSON の
// 13ステップパイプラインを統合する。
package app

import (
	"context"
	"io"

	"github.com/youyo/imgraft/internal/auth"
	"github.com/youyo/imgraft/internal/backend/studio"
	"github.com/youyo/imgraft/internal/config"
	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/model"
	"github.com/youyo/imgraft/internal/output"
	"github.com/youyo/imgraft/internal/prompt"
	"github.com/youyo/imgraft/internal/runtime"
)

// RunInput はパイプラインへの入力パラメータ。
type RunInput struct {
	Prompt        string
	ModelAlias    string
	Refs          []string // M12以降で使用
	OutputPath    string
	Dir           string
	NoTransparent bool
	Profile       string
	ConfigPath    string
	Pretty        bool
	Verbose       bool
	Debug         bool
}

// RunOutput はパイプラインの実行結果。
type RunOutput struct {
	Output   output.Output
	ExitCode int
}

// Dependencies はパイプラインの外部依存性（テスト時に DI 可能）。
type Dependencies struct {
	// StudioClientFactory は API キーから StudioClient を生成するファクトリ関数。
	// nil の場合は studio.New を使用する。
	StudioClientFactory func(apiKey string) studio.StudioClient

	// Clock は時刻プロバイダー。nil の場合は runtime.SystemClock を使用する。
	Clock runtime.Clock

	// CredPath は credentials.json へのパス。空文字の場合はデフォルトパスを使用する。
	CredPath string

	// Stderr はログ出力先。nil の場合は io.Discard を使用する。
	Stderr io.Writer
}

// Run は 13 ステップパイプラインを実行し RunOutput を返す。
// パイプラインのどのステップでエラーが発生しても必ず RunOutput を返す。
func Run(ctx context.Context, input RunInput, deps Dependencies) RunOutput {
	// 依存性のデフォルト値を設定
	if deps.StudioClientFactory == nil {
		deps.StudioClientFactory = func(apiKey string) studio.StudioClient {
			return studio.New(apiKey)
		}
	}
	if deps.Clock == nil {
		deps.Clock = runtime.SystemClock{}
	}
	if deps.Stderr == nil {
		deps.Stderr = io.Discard
	}

	// backend は v1 では常に google_ai_studio
	const backend = "google_ai_studio"

	// ステップ 2: Config load
	cfg, err := config.Load(input.ConfigPath)
	if err != nil {
		return errorOutput(nil, nil, err)
	}

	// ステップ 3: Profile resolve
	profile := resolveProfile(input.Profile, cfg)

	// ステップ 4: Auth resolve
	creds, err := auth.Load(deps.CredPath)
	if err != nil {
		return errorOutput(nil, nil, err)
	}

	// ステップ 5: API key extract
	apiKey, err := extractAPIKey(creds, profile, backend)
	if err != nil {
		return errorOutput(nil, nil, err)
	}

	// ステップ 6: Model resolve
	modelName := model.Resolve(input.ModelAlias, cfg)

	// ステップ 7: Reference load/validate（M12 以降）
	// 現在は未実装のためスキップ

	// ステップ 8: Prompt build
	transparent := !input.NoTransparent
	parts, err := prompt.Build(ctx, input.Prompt, transparent)
	if err != nil {
		return errorOutput(nil, nil, err)
	}

	// prompt.Parts から SystemPrompt と UserPrompt を抽出
	systemPromptText, userPromptText := extractPromptParts(parts)

	// ステップ 9: API generate
	client := deps.StudioClientFactory(apiKey)
	genReq := studio.GenerateRequest{
		Model:        modelName,
		Prompt:       userPromptText,
		SystemPrompt: systemPromptText,
	}
	genResp, _, err := client.Generate(ctx, genReq)
	if err != nil {
		return errorOutput(&modelName, strPtr(backend), err)
	}

	// ステップ 10: Transparent pipeline（M14/M15 以降）
	// 現在は未実装のため、生成された画像データをそのまま使用
	imageData := genResp.ImageData

	// ステップ 11: Save
	saveOpts := output.SaveOptions{
		OutputPath:         input.OutputPath,
		Dir:                resolveDir(input.Dir, cfg),
		Clock:              deps.Clock,
		Index:              0,
		TransparentApplied: false, // M14/M15 実装後に更新
	}
	item, err := output.SavePNG(imageData, saveOpts)
	if err != nil {
		return errorOutput(&modelName, strPtr(backend), err)
	}

	// ステップ 12: Inspect/hash は SavePNG 内部で実行済み

	// ステップ 13: JSON emit
	out := output.NewSuccessOutput()
	out.Model = &modelName
	backendStr := backend
	out.Backend = &backendStr
	out.Images = []output.ImageItem{item}

	return RunOutput{
		Output:   out,
		ExitCode: 0,
	}
}

// resolveProfile はプロファイル名を解決する。
// 優先順位: input.Profile > config.CurrentProfile > "default"
func resolveProfile(inputProfile string, cfg *config.Config) string {
	if inputProfile != "" {
		return inputProfile
	}
	if cfg.CurrentProfile != "" {
		return cfg.CurrentProfile
	}
	return config.DefaultProfile
}

// extractAPIKey は指定プロファイルから API キーを抽出する。
// 見つからない場合は AUTH_REQUIRED エラーを返す。
func extractAPIKey(creds *auth.Credentials, profile, backend string) (string, error) {
	if creds == nil || len(creds.Profiles) == 0 {
		return "", errs.New(errs.ErrAuthRequired, "no credentials found; run 'imgraft auth login' to authenticate")
	}

	pc, ok := creds.Profiles[profile]
	if !ok {
		return "", errs.New(errs.ErrAuthRequired, "profile not found: "+profile+"; run 'imgraft auth login' to authenticate")
	}

	apiKey, ok := auth.GetAPIKey(pc, backend)
	if !ok || apiKey == "" {
		return "", errs.New(errs.ErrAuthRequired, "no API key found for profile "+profile+"; run 'imgraft auth login' to authenticate")
	}

	return apiKey, nil
}

// resolveDir は出力ディレクトリを解決する。
// 優先順位: input.Dir > config.DefaultOutputDir > "."
func resolveDir(inputDir string, cfg *config.Config) string {
	if inputDir != "" {
		return inputDir
	}
	if cfg.DefaultOutputDir != "" {
		return cfg.DefaultOutputDir
	}
	return "."
}

// extractPromptParts は prompt.Parts から SystemPrompt と UserPrompt を抽出する。
func extractPromptParts(parts []prompt.Part) (systemPrompt, userPrompt string) {
	for _, p := range parts {
		switch p.Role {
		case prompt.RoleSystem:
			systemPrompt = p.Text
		case prompt.RoleUser:
			userPrompt = p.Text
		}
	}
	return
}

// errorOutput はエラー時の RunOutput を生成する。
func errorOutput(modelName *string, backend *string, err error) RunOutput {
	code := string(errs.CodeOf(err))
	msg := err.Error()

	out := output.NewErrorOutput(code, msg)
	out.Model = modelName
	out.Backend = backend

	return RunOutput{
		Output:   out,
		ExitCode: 1,
	}
}

// strPtr は文字列ポインタを返すヘルパー。
func strPtr(s string) *string {
	return &s
}
