package cli_test

import (
	"testing"

	"github.com/youyo/imgraft/internal/cli"
)

// TestCLI_ParseAuthLogin は auth login サブコマンドのパースをテスト。
func TestCLI_ParseAuthLogin(t *testing.T) {
	_, ctx, err := cli.Parse([]string{"auth", "login"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Command() != "auth login" {
		t.Errorf("Command() = %q, want %q", ctx.Command(), "auth login")
	}
}

// TestCLI_ParseAuthLoginWithFlags は auth login のフラグ付きパースをテスト。
func TestCLI_ParseAuthLoginWithFlags(t *testing.T) {
	c, ctx, err := cli.Parse([]string{"auth", "login", "--api-key", "mykey", "--profile", "myprofile"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Command() != "auth login" {
		t.Errorf("Command() = %q, want %q", ctx.Command(), "auth login")
	}
	if c.Auth.Login.APIKey != "mykey" {
		t.Errorf("APIKey = %q, want %q", c.Auth.Login.APIKey, "mykey")
	}
	if c.Auth.Login.Profile != "myprofile" {
		t.Errorf("Profile = %q, want %q", c.Auth.Login.Profile, "myprofile")
	}
}

// TestCLI_ParseAuthLogout は auth logout サブコマンドのパースをテスト。
func TestCLI_ParseAuthLogout(t *testing.T) {
	_, ctx, err := cli.Parse([]string{"auth", "logout"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Command() != "auth logout" {
		t.Errorf("Command() = %q, want %q", ctx.Command(), "auth logout")
	}
}

// TestCLI_ParseAuthLogoutWithProfile は auth logout --profile フラグのパースをテスト。
func TestCLI_ParseAuthLogoutWithProfile(t *testing.T) {
	c, _, err := cli.Parse([]string{"auth", "logout", "--profile", "myprofile"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Auth.Logout.Profile != "myprofile" {
		t.Errorf("Profile = %q, want %q", c.Auth.Logout.Profile, "myprofile")
	}
}

// TestCLI_ParseAuthWhoami は auth whoami サブコマンドのパースをテスト。
func TestCLI_ParseAuthWhoami(t *testing.T) {
	_, ctx, err := cli.Parse([]string{"auth", "whoami"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Command() != "auth whoami" {
		t.Errorf("Command() = %q, want %q", ctx.Command(), "auth whoami")
	}
}
