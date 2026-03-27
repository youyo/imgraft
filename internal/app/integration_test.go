package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/youyo/imgraft/internal/app"
	"github.com/youyo/imgraft/internal/backend/studio"
	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/output"
	"github.com/youyo/imgraft/internal/runtime"
)

func TestIntegration_FullPipeline(t *testing.T) {
	credPath := makeTempCredentials(t, "test-api-key")
	outDir := t.TempDir()

	fixedTime := time.Date(2026, 3, 28, 15, 0, 0, 0, time.UTC)
	deps := app.Dependencies{
		StudioClientFactory: mockFactory(&mockStudioClient{}),
		Clock:               runtime.NewFixedClock(fixedTime),
		CredPath:            credPath,
	}

	input := app.RunInput{
		Prompt: "blue robot mascot",
		Dir:    outDir,
	}

	result := app.Run(context.Background(), input, deps)
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; error: %+v", result.ExitCode, result.Output.Error)
	}
	if !result.Output.Success {
		t.Error("expected success=true")
	}
	if len(result.Output.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(result.Output.Images))
	}
	img := result.Output.Images[0]
	if img.Path == "" {
		t.Error("expected non-empty path")
	}
	if img.SHA256 == "" {
		t.Error("expected non-empty sha256")
	}
	if img.Width <= 0 || img.Height <= 0 {
		t.Errorf("expected positive dimensions, got %dx%d", img.Width, img.Height)
	}
	// model, backend が設定されている
	if result.Output.Model == nil || *result.Output.Model == "" {
		t.Error("expected non-empty model")
	}
	if result.Output.Backend == nil || *result.Output.Backend == "" {
		t.Error("expected non-empty backend")
	}
}

func TestIntegration_ErrorJSON_FixedSchema(t *testing.T) {
	// API エラーの場合でも固定スキーマ JSON が出ることを確認
	credPath := makeTempCredentials(t, "test-api-key")
	outDir := t.TempDir()

	apiErrMock := &mockStudioClient{
		generateFn: func(ctx context.Context, req studio.GenerateRequest) (studio.GenerateResponse, http.Header, error) {
			return studio.GenerateResponse{}, nil, errs.New(errs.ErrInternal, "server error")
		},
	}

	fixedTime := time.Date(2026, 3, 28, 15, 0, 0, 0, time.UTC)
	deps := app.Dependencies{
		StudioClientFactory: mockFactory(apiErrMock),
		Clock:               runtime.NewFixedClock(fixedTime),
		CredPath:            credPath,
	}

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

	// JSON にシリアライズして固定スキーマを確認
	data, err := json.Marshal(result.Output)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	requiredFields := []string{"success", "model", "backend", "images", "rate_limit", "warnings", "error"}
	for _, field := range requiredFields {
		if _, ok := schema[field]; !ok {
			t.Errorf("missing required field: %s", field)
		}
	}
	// images は空配列（null ではない）
	if imgs, ok := schema["images"].([]interface{}); !ok || imgs == nil {
		t.Error("expected images to be an empty array")
	}
}

func TestIntegration_PrettyOutput(t *testing.T) {
	credPath := makeTempCredentials(t, "test-api-key")
	outDir := t.TempDir()

	fixedTime := time.Date(2026, 3, 28, 15, 0, 0, 0, time.UTC)
	deps := app.Dependencies{
		StudioClientFactory: mockFactory(&mockStudioClient{}),
		Clock:               runtime.NewFixedClock(fixedTime),
		CredPath:            credPath,
	}

	input := app.RunInput{
		Prompt: "blue robot mascot",
		Dir:    outDir,
		Pretty: true,
	}

	result := app.Run(context.Background(), input, deps)
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; error: %+v", result.ExitCode, result.Output.Error)
	}

	// pretty フラグの場合の Encode テスト
	var buf bytes.Buffer
	if err := output.Encode(&buf, result.Output, input.Pretty); err != nil {
		t.Fatalf("failed to encode: %v", err)
	}
	// pretty=true の場合、改行やインデントが含まれる
	jsonStr := buf.String()
	if len(jsonStr) == 0 {
		t.Error("expected non-empty JSON")
	}
	// インデント付きの場合は "  " が含まれる
	if !bytes.Contains([]byte(jsonStr), []byte("  ")) {
		t.Error("expected indented JSON with spaces")
	}
}

func TestIntegration_FixedSchemaOnAuthError(t *testing.T) {
	// 認証なしでエラーになる場合の JSON スキーマを確認
	dir := t.TempDir()
	emptyCredPath := filepath.Join(dir, "credentials.json")
	os.WriteFile(emptyCredPath, []byte(`{"profiles":{}}`), 0o600)

	fixedTime := time.Date(2026, 3, 28, 15, 0, 0, 0, time.UTC)
	deps := app.Dependencies{
		StudioClientFactory: mockFactory(&mockStudioClient{}),
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

	// output を JSON にシリアライズして固定スキーマを確認
	data, err := json.Marshal(result.Output)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var o output.Output
	if err := json.Unmarshal(data, &o); err != nil {
		t.Fatalf("failed to unmarshal into Output: %v", err)
	}

	if o.Success {
		t.Error("expected success=false")
	}
	if o.Images == nil {
		t.Error("expected non-nil images (empty array)")
	}
	if len(o.Images) != 0 {
		t.Errorf("expected 0 images, got %d", len(o.Images))
	}
	if o.Warnings == nil {
		t.Error("expected non-nil warnings (empty array)")
	}
	if o.Error.Code == nil {
		t.Fatal("expected non-nil error.code")
	}
	if *o.Error.Code != string(errs.ErrAuthRequired) {
		t.Errorf("expected error code %q, got %q", errs.ErrAuthRequired, *o.Error.Code)
	}
}
