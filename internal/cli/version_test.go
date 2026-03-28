package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/youyo/imgraft/internal/cli"
)

func TestVersionCmd_Run_DefaultValues(t *testing.T) {
	// デフォルト値（ldflags未設定）でのバージョン出力テスト
	// パッケージ変数をリセット
	origVersion := cli.Version
	origCommit := cli.Commit
	origDate := cli.Date
	defer func() {
		cli.Version = origVersion
		cli.Commit = origCommit
		cli.Date = origDate
	}()

	cli.Version = "dev"
	cli.Commit = "none"
	cli.Date = "unknown"

	var buf bytes.Buffer
	cmd := &cli.VersionCmd{}
	err := cmd.RunWithWriter(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	expected := "imgraft version dev (commit: none, built: unknown)\n"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestVersionCmd_Run_CustomValues(t *testing.T) {
	origVersion := cli.Version
	origCommit := cli.Commit
	origDate := cli.Date
	defer func() {
		cli.Version = origVersion
		cli.Commit = origCommit
		cli.Date = origDate
	}()

	cli.Version = "1.2.3"
	cli.Commit = "abc1234"
	cli.Date = "2026-03-28"

	var buf bytes.Buffer
	cmd := &cli.VersionCmd{}
	err := cmd.RunWithWriter(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	expected := "imgraft version 1.2.3 (commit: abc1234, built: 2026-03-28)\n"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestVersionCmd_Run_ContainsVersion(t *testing.T) {
	origVersion := cli.Version
	defer func() { cli.Version = origVersion }()

	cli.Version = "0.1.0"

	var buf bytes.Buffer
	cmd := &cli.VersionCmd{}
	err := cmd.RunWithWriter(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "0.1.0") {
		t.Errorf("expected output to contain version %q, got %q", "0.1.0", got)
	}
	if !strings.Contains(got, "imgraft version") {
		t.Errorf("expected output to contain 'imgraft version', got %q", got)
	}
}

func TestVersionCmd_ParseSubcommand(t *testing.T) {
	_, _, err := cli.Parse([]string{"version"})
	if err != nil {
		t.Fatalf("unexpected error parsing version subcommand: %v", err)
	}
}

func TestCompletionCmd_ParseSubcommand(t *testing.T) {
	_, _, err := cli.Parse([]string{"completion", "zsh"})
	if err != nil {
		t.Fatalf("unexpected error parsing completion zsh subcommand: %v", err)
	}
}

func TestCompletionCmd_Zsh_Output(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cli.CompletionCmd{
		Shell: "zsh",
	}
	err := cmd.RunWithWriter(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	// zsh completion スクリプトの基本要素を確認
	if !strings.Contains(got, "#compdef") {
		t.Errorf("expected zsh completion to contain '#compdef', got %q", got[:min(len(got), 200)])
	}
	if !strings.Contains(got, "imgraft") {
		t.Errorf("expected zsh completion to contain 'imgraft', got %q", got[:min(len(got), 200)])
	}
}

func TestCompletionCmd_UnknownShell_Error(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cli.CompletionCmd{
		Shell: "fish",
	}
	err := cmd.RunWithWriter(&buf)
	if err == nil {
		t.Fatal("expected error for unsupported shell, got nil")
	}
	if !strings.Contains(err.Error(), "fish") {
		t.Errorf("expected error message to contain shell name, got %q", err.Error())
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
