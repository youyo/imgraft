package cli_test

import (
	"context"
	"errors"
	"testing"

	"github.com/youyo/imgraft/internal/backend/studio"
	"github.com/youyo/imgraft/internal/cli"
)

// TestCLI_ParseConfigInit は config init サブコマンドのパースをテスト。
func TestCLI_ParseConfigInit(t *testing.T) {
	_, ctx, err := cli.Parse([]string{"config", "init"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Command() != "config init" {
		t.Errorf("Command() = %q, want %q", ctx.Command(), "config init")
	}
}

// TestCLI_ParseConfigInitWithFlags は config init のフラグ付きパースをテスト。
func TestCLI_ParseConfigInitWithFlags(t *testing.T) {
	c, ctx, err := cli.Parse([]string{"config", "init", "--api-key", "mykey", "--profile", "myprofile"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Command() != "config init" {
		t.Errorf("Command() = %q, want %q", ctx.Command(), "config init")
	}
	if c.Config.Init.APIKey != "mykey" {
		t.Errorf("APIKey = %q, want %q", c.Config.Init.APIKey, "mykey")
	}
	if c.Config.Init.Profile != "myprofile" {
		t.Errorf("Profile = %q, want %q", c.Config.Init.Profile, "myprofile")
	}
}

// TestCLI_ParseConfigUse は config use サブコマンドのパースをテスト。
func TestCLI_ParseConfigUse(t *testing.T) {
	c, ctx, err := cli.Parse([]string{"config", "use", "myprofile"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Command() != "config use <profile>" {
		t.Errorf("Command() = %q, want %q", ctx.Command(), "config use <profile>")
	}
	if c.Config.Use.Profile != "myprofile" {
		t.Errorf("Profile = %q, want %q", c.Config.Use.Profile, "myprofile")
	}
}

// TestCLI_ParseConfigRefreshModels は config refresh-models サブコマンドのパースをテスト。
func TestCLI_ParseConfigRefreshModels(t *testing.T) {
	_, ctx, err := cli.Parse([]string{"config", "refresh-models"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Command() != "config refresh-models" {
		t.Errorf("Command() = %q, want %q", ctx.Command(), "config refresh-models")
	}
}

// TestConfigInitRun_WithMockValidator は ConfigInitCmd.Run のモック疎通テスト。
func TestConfigInitRun_WithMockValidator(t *testing.T) {
	validator := &mockInitValidator{valid: true, models: []string{"models/gemini-2.0-flash-exp", "models/gemini-pro"}}

	opts := cli.ConfigInitOptions{
		APIKey:     "testkey",
		Profile:    "testprofile",
		ConfigPath: t.TempDir() + "/config.toml",
		CredPath:   t.TempDir() + "/credentials.json",
		Validator:  validator,
		Lister:     validator,
	}

	err := cli.RunConfigInit(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !validator.validateCalled {
		t.Error("ValidateAPIKey should have been called")
	}
	if !validator.listCalled {
		t.Error("ListModels should have been called")
	}
}

// TestConfigInitRun_ValidationFailure は API key が無効な場合のエラーテスト。
func TestConfigInitRun_ValidationFailure(t *testing.T) {
	validator := &mockInitValidator{valid: false, validateErr: errMockValidation}

	opts := cli.ConfigInitOptions{
		APIKey:     "badkey",
		Profile:    "testprofile",
		ConfigPath: t.TempDir() + "/config.toml",
		CredPath:   t.TempDir() + "/credentials.json",
		Validator:  validator,
		Lister:     validator,
	}

	err := cli.RunConfigInit(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for invalid API key")
	}
}

// TestConfigUseRun は ConfigUseCmd.Run の profile 切り替えテスト。
func TestConfigUseRun_SwitchProfile(t *testing.T) {
	configPath := t.TempDir() + "/config.toml"

	err := cli.RunConfigUse(context.Background(), cli.ConfigUseOptions{
		Profile:    "newprofile",
		ConfigPath: configPath,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestConfigRefreshModelsRun は ConfigRefreshModelsCmd.Run のモック疎通テスト。
func TestConfigRefreshModelsRun_WithMockLister(t *testing.T) {
	lister := &mockInitValidator{models: []string{"models/gemini-2.0-flash-exp", "models/gemini-pro"}}

	opts := cli.ConfigRefreshModelsOptions{
		ConfigPath: t.TempDir() + "/config.toml",
		CredPath:   t.TempDir() + "/credentials.json",
		Lister:     lister,
	}

	err := cli.RunConfigRefreshModels(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !lister.listCalled {
		t.Error("ListModels should have been called")
	}
}

// --- テスト用モック ---

var errMockValidation = errors.New("mock validation error")

type mockInitValidator struct {
	valid          bool
	validateErr    error
	models         []string
	validateCalled bool
	listCalled     bool
}

func (m *mockInitValidator) ValidateAPIKey(ctx context.Context) error {
	m.validateCalled = true
	if m.validateErr != nil {
		return m.validateErr
	}
	return nil
}

func (m *mockInitValidator) ListModels(ctx context.Context) ([]studio.RemoteModel, error) {
	m.listCalled = true
	result := make([]studio.RemoteModel, 0, len(m.models))
	for _, name := range m.models {
		result = append(result, studio.RemoteModel{Name: name, SupportedGeneration: true})
	}
	return result, nil
}
