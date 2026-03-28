package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/youyo/imgraft/internal/auth"
)

// Run は auth login コマンドを実行する。
// --api-key / --profile フラグが指定されれば非対話実行。
// 省略された場合は stdin から対話的に読み取る。
func (c *AuthLoginCmd) Run() error {
	opts := auth.LoginOptions{
		APIKey:  c.APIKey,
		Profile: c.Profile,
		Reader:  bufio.NewReader(os.Stdin),
	}

	result, err := auth.Login(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("auth login: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Logged in as profile %q (backend: %s)\n", result.Profile, result.Backend)
	return nil
}

// Run は auth logout コマンドを実行する。
func (c *AuthLogoutCmd) Run() error {
	opts := auth.LogoutOptions{
		Profile: c.Profile,
	}

	if err := auth.Logout(opts); err != nil {
		return fmt.Errorf("auth logout: %w", err)
	}

	profile := c.Profile
	if profile == "" {
		profile = "(current profile)"
	}
	fmt.Fprintf(os.Stderr, "Logged out from profile %s\n", profile)
	return nil
}

// Run は auth whoami コマンドを実行する。
// 結果は stderr に出力する（stdout は常に JSON のため）。
func (c *AuthWhoamiCmd) Run() error {
	opts := auth.WhoamiOptions{}

	result, err := auth.Whoami(opts)
	if err != nil {
		return fmt.Errorf("auth whoami: %w", err)
	}

	fmt.Fprint(os.Stderr, result.String())
	return nil
}
