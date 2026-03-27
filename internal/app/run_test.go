package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/youyo/imgraft/internal/app"
	"github.com/youyo/imgraft/internal/backend/studio"
	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/runtime"
)

// minimalPNG は 1x1 の最小 PNG バイト列。
// transparent パイプライン（M14/M15）の前に SavePNG のテストに使う。
var minimalPNG = func() []byte {
	// 1x1 黒ピクセル PNG (正確な PNG バイナリ)
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
		0x00, 0x00, 0x02, 0x00, 0x01, 0xE2, 0x21, 0xBC,
		0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
}()

// mockStudioClient は studio.StudioClient のモック実装。
type mockStudioClient struct {
	generateFn func(ctx context.Context, req studio.GenerateRequest) (studio.GenerateResponse, http.Header, error)
}

func (m *mockStudioClient) Generate(ctx context.Context, req studio.GenerateRequest) (studio.GenerateResponse, http.Header, error) {
	if m.generateFn != nil {
		return m.generateFn(ctx, req)
	}
	return studio.GenerateResponse{
		ImageData: minimalPNG,
		MimeType:  "image/png",
	}, http.Header{}, nil
}

func (m *mockStudioClient) ListModels(ctx context.Context) ([]studio.RemoteModel, error) {
	return nil, nil
}

func (m *mockStudioClient) ValidateAPIKey(ctx context.Context) error {
	return nil
}

// mockFactory は StudioClientFactory のモック。
func mockFactory(client studio.StudioClient) func(string) studio.StudioClient {
	return func(apiKey string) studio.StudioClient {
		return client
	}
}

// makeTempCredentials は一時的な credentials.json を作成し、パスを返す。
func makeTempCredentials(t *testing.T, apiKey string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	content := `{"profiles":{"default":{"google_ai_studio":{"api_key":"` + apiKey + `"}}}}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write credentials: %v", err)
	}
	return path
}

// makeFixedDeps は固定時刻の Dependencies を作成する。
func makeFixedDeps(t *testing.T, client studio.StudioClient, credPath string) app.Dependencies {
	t.Helper()
	fixedTime := time.Date(2026, 3, 28, 15, 0, 0, 0, time.UTC)
	return app.Dependencies{
		StudioClientFactory: mockFactory(client),
		Clock:               runtime.NewFixedClock(fixedTime),
		CredPath:            credPath,
	}
}

func TestRun_Success(t *testing.T) {
	credPath := makeTempCredentials(t, "test-api-key")
	outDir := t.TempDir()

	mock := &mockStudioClient{}
	deps := makeFixedDeps(t, mock, credPath)

	input := app.RunInput{
		Prompt: "blue robot mascot",
		Dir:    outDir,
	}

	result := app.Run(context.Background(), input, deps)
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; error: %v", result.ExitCode, result.Output.Error)
	}
	if !result.Output.Success {
		t.Errorf("expected success=true")
	}
	if len(result.Output.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(result.Output.Images))
	}
	if result.Output.Images[0].Path == "" {
		t.Error("expected non-empty image path")
	}
	if result.Output.Model == nil || *result.Output.Model == "" {
		t.Error("expected non-empty model")
	}
}

func TestRun_AuthRequired(t *testing.T) {
	// credentials.json がない = 空 credentials
	dir := t.TempDir()
	emptyCredPath := filepath.Join(dir, "credentials.json")
	// 空の credentials（profiles なし）
	os.WriteFile(emptyCredPath, []byte(`{"profiles":{}}`), 0o600)

	mock := &mockStudioClient{}
	fixedTime := time.Date(2026, 3, 28, 15, 0, 0, 0, time.UTC)
	deps := app.Dependencies{
		StudioClientFactory: mockFactory(mock),
		Clock:               runtime.NewFixedClock(fixedTime),
		CredPath:            emptyCredPath,
	}

	input := app.RunInput{
		Prompt: "blue robot mascot",
		Dir:    dir,
	}

	result := app.Run(context.Background(), input, deps)
	if result.ExitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", result.ExitCode)
	}
	if result.Output.Success {
		t.Error("expected success=false")
	}
	if result.Output.Error.Code == nil {
		t.Fatal("expected non-nil error.code")
	}
	if *result.Output.Error.Code != string(errs.ErrAuthRequired) {
		t.Errorf("expected error code %q, got %q", errs.ErrAuthRequired, *result.Output.Error.Code)
	}
}

func TestRun_GenerateError(t *testing.T) {
	credPath := makeTempCredentials(t, "test-api-key")
	outDir := t.TempDir()

	mock := &mockStudioClient{
		generateFn: func(ctx context.Context, req studio.GenerateRequest) (studio.GenerateResponse, http.Header, error) {
			return studio.GenerateResponse{}, nil, errs.New(errs.ErrBackendUnavailable, "API unavailable")
		},
	}
	deps := makeFixedDeps(t, mock, credPath)

	input := app.RunInput{
		Prompt: "blue robot mascot",
		Dir:    outDir,
	}

	result := app.Run(context.Background(), input, deps)
	if result.ExitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", result.ExitCode)
	}
	if result.Output.Success {
		t.Error("expected success=false")
	}
	if result.Output.Error.Code == nil {
		t.Fatal("expected non-nil error.code")
	}
	if *result.Output.Error.Code != string(errs.ErrBackendUnavailable) {
		t.Errorf("expected error code %q, got %q", errs.ErrBackendUnavailable, *result.Output.Error.Code)
	}
}

func TestRun_NoTransparent(t *testing.T) {
	credPath := makeTempCredentials(t, "test-api-key")
	outDir := t.TempDir()

	var capturedReq studio.GenerateRequest
	mock := &mockStudioClient{
		generateFn: func(ctx context.Context, req studio.GenerateRequest) (studio.GenerateResponse, http.Header, error) {
			capturedReq = req
			return studio.GenerateResponse{
				ImageData: minimalPNG,
				MimeType:  "image/png",
			}, http.Header{}, nil
		},
	}
	deps := makeFixedDeps(t, mock, credPath)

	input := app.RunInput{
		Prompt:        "blue robot mascot",
		NoTransparent: true,
		Dir:           outDir,
	}

	result := app.Run(context.Background(), input, deps)
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; error: %v", result.ExitCode, result.Output.Error)
	}
	// transparent=false の場合、SystemPrompt は空文字列
	if capturedReq.SystemPrompt != "" {
		t.Errorf("expected empty system prompt for --no-transparent, got %q", capturedReq.SystemPrompt)
	}
	// 出力画像の transparent_applied は false
	if len(result.Output.Images) > 0 && result.Output.Images[0].TransparentApplied {
		t.Error("expected transparent_applied=false for --no-transparent")
	}
}

func TestRun_CustomModel(t *testing.T) {
	credPath := makeTempCredentials(t, "test-api-key")
	outDir := t.TempDir()

	var capturedModel string
	mock := &mockStudioClient{
		generateFn: func(ctx context.Context, req studio.GenerateRequest) (studio.GenerateResponse, http.Header, error) {
			capturedModel = req.Model
			return studio.GenerateResponse{
				ImageData: minimalPNG,
				MimeType:  "image/png",
			}, http.Header{}, nil
		},
	}
	deps := makeFixedDeps(t, mock, credPath)

	input := app.RunInput{
		Prompt:     "blue robot mascot",
		ModelAlias: "pro",
		Dir:        outDir,
	}

	result := app.Run(context.Background(), input, deps)
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; error: %v", result.ExitCode, result.Output.Error)
	}
	// pro alias は built-in pro モデルに解決される
	if capturedModel != "gemini-3-pro-image-preview" {
		t.Errorf("expected model %q, got %q", "gemini-3-pro-image-preview", capturedModel)
	}
}

func TestRun_OutputPath(t *testing.T) {
	credPath := makeTempCredentials(t, "test-api-key")
	outDir := t.TempDir()
	outputPath := filepath.Join(outDir, "custom-output.png")

	mock := &mockStudioClient{}
	deps := makeFixedDeps(t, mock, credPath)

	input := app.RunInput{
		Prompt:     "blue robot mascot",
		OutputPath: outputPath,
	}

	result := app.Run(context.Background(), input, deps)
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; error: %v", result.ExitCode, result.Output.Error)
	}
	if len(result.Output.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(result.Output.Images))
	}
	// 指定したパスにファイルが保存される
	absExpected, _ := filepath.Abs(outputPath)
	if result.Output.Images[0].Path != absExpected {
		t.Errorf("expected path %q, got %q", absExpected, result.Output.Images[0].Path)
	}
}

func TestRun_OutputDir(t *testing.T) {
	credPath := makeTempCredentials(t, "test-api-key")
	outDir := t.TempDir()

	mock := &mockStudioClient{}
	deps := makeFixedDeps(t, mock, credPath)

	input := app.RunInput{
		Prompt: "blue robot mascot",
		Dir:    outDir,
	}

	result := app.Run(context.Background(), input, deps)
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; error: %v", result.ExitCode, result.Output.Error)
	}
	if len(result.Output.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(result.Output.Images))
	}
	// ファイルが outDir に保存される
	if !strings.HasPrefix(result.Output.Images[0].Path, outDir) {
		t.Errorf("expected image path to start with %q, got %q", outDir, result.Output.Images[0].Path)
	}
}

func TestRun_EmptyPromptError(t *testing.T) {
	credPath := makeTempCredentials(t, "test-api-key")
	outDir := t.TempDir()

	mock := &mockStudioClient{}
	deps := makeFixedDeps(t, mock, credPath)

	input := app.RunInput{
		Prompt: "",
		Dir:    outDir,
	}

	result := app.Run(context.Background(), input, deps)
	if result.ExitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", result.ExitCode)
	}
	if result.Output.Success {
		t.Error("expected success=false")
	}
}

func TestRun_OutputIsValidJSON(t *testing.T) {
	credPath := makeTempCredentials(t, "test-api-key")
	outDir := t.TempDir()

	mock := &mockStudioClient{}
	deps := makeFixedDeps(t, mock, credPath)

	input := app.RunInput{
		Prompt: "blue robot mascot",
		Dir:    outDir,
	}

	result := app.Run(context.Background(), input, deps)

	// output が JSON としてシリアライズできることを確認
	data, err := json.Marshal(result.Output)
	if err != nil {
		t.Fatalf("failed to marshal output: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON output")
	}
}
