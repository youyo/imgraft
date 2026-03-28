package app_test

import (
	"context"
	"os"
	"testing"

	"github.com/youyo/imgraft/internal/app"
)

// TestE2E_GenerateImage は実際の Google AI Studio API を使って画像を生成するテスト。
// RUN_E2E=1 環境変数が設定されている場合のみ実行される。
// 環境変数 IMGRAFT_API_KEY から API キーを取得する。
func TestE2E_GenerateImage(t *testing.T) {
	if os.Getenv("RUN_E2E") != "1" {
		t.Skip("skipping E2E test; set RUN_E2E=1 to run")
	}

	apiKey := os.Getenv("IMGRAFT_API_KEY")
	if apiKey == "" {
		t.Fatal("IMGRAFT_API_KEY environment variable is required for E2E tests")
	}

	// 一時的な credentials を作成
	credPath := makeTempCredentials(t, apiKey)
	outDir := t.TempDir()

	deps := app.Dependencies{
		CredPath: credPath,
		Stderr:   os.Stderr,
	}

	input := app.RunInput{
		Prompt: "simple blue circle icon",
		Dir:    outDir,
	}

	result := app.Run(context.Background(), input, deps)
	if result.ExitCode != 0 {
		code, msg := "<nil>", "<nil>"
		if result.Output.Error.Code != nil {
			code = *result.Output.Error.Code
		}
		if result.Output.Error.Message != nil {
			msg = *result.Output.Error.Message
		}
		t.Fatalf("expected exit code 0, got %d; error code=%s, message=%s", result.ExitCode, code, msg)
	}
	if !result.Output.Success {
		t.Error("expected success=true")
	}
	if len(result.Output.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(result.Output.Images))
	}

	img := result.Output.Images[0]
	t.Logf("Generated image: %s", img.Path)
	t.Logf("SHA256: %s", img.SHA256)
	t.Logf("Dimensions: %dx%d", img.Width, img.Height)

	// ファイルが実際に存在することを確認
	if _, err := os.Stat(img.Path); err != nil {
		t.Fatalf("generated file not found: %v", err)
	}
}
