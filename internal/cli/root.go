// Package cli は imgraft の CLI 定義を kong を使って提供する。
// メイン生成コマンドとサブコマンドの構造を定義する。
package cli

import (
	"github.com/alecthomas/kong"
)

// CLI は imgraft のルート CLI 構造体。
// kong の struct tags でフラグとサブコマンドを定義する。
// Generate は default:"withargs" パターンで引数なし実行のデフォルトコマンドとなる。
type CLI struct {
	// メイン生成コマンド（デフォルトコマンド）
	Generate GenerateCmd `cmd:"" default:"withargs" help:"Generate a transparent image asset (default command)"`

	// サブコマンド（M18-M20 で実装予定）
	Auth    AuthCmd    `cmd:"" help:"Authentication commands"`
	Config  ConfigCmd  `cmd:"" help:"Configuration commands"`
	Version VersionCmd `cmd:"" help:"Show version information"`
}

// GenerateCmd はメイン画像生成コマンドの引数とフラグ。
// default:"withargs" により、サブコマンドなし実行時のデフォルトとなる。
type GenerateCmd struct {
	Prompt        string   `arg:"" optional:"" help:"Image generation prompt"`
	Model         string   `name:"model" help:"Model alias or full name (flash, pro, or full model name)" default:""`
	Ref           []string `name:"ref" help:"Reference image path or URL (can be specified multiple times)" type:"string"`
	Output        string   `name:"output" help:"Output file path" default:""`
	Dir           string   `name:"dir" help:"Output directory" default:""`
	NoTransparent bool     `name:"no-transparent" help:"Disable transparent mode (default: transparent ON)"`
	Profile       string   `name:"profile" help:"Profile name to use" default:""`
	ConfigPath    string   `name:"config" help:"Config file path" default:""`
	Pretty        bool     `name:"pretty" help:"Pretty-print JSON output"`
	Verbose       bool     `name:"verbose" help:"Enable verbose logging to stderr"`
	Debug         bool     `name:"debug" help:"Enable debug logging to stderr"`
}

// AuthCmd は auth サブコマンドグループ（M18 で実装予定）。
type AuthCmd struct {
	Login  AuthLoginCmd  `cmd:"" help:"Login to backend"`
	Logout AuthLogoutCmd `cmd:"" help:"Logout from backend"`
	Whoami AuthWhoamiCmd `cmd:"" help:"Show current authentication status"`
}

// AuthLoginCmd は auth login サブコマンド（M18 で実装予定）。
type AuthLoginCmd struct{}

// Run は M18 で実装される。
func (c *AuthLoginCmd) Run() error {
	return errNotImplemented("auth login")
}

// AuthLogoutCmd は auth logout サブコマンド（M18 で実装予定）。
type AuthLogoutCmd struct{}

// Run は M18 で実装される。
func (c *AuthLogoutCmd) Run() error {
	return errNotImplemented("auth logout")
}

// AuthWhoamiCmd は auth whoami サブコマンド（M18 で実装予定）。
type AuthWhoamiCmd struct{}

// Run は M18 で実装される。
func (c *AuthWhoamiCmd) Run() error {
	return errNotImplemented("auth whoami")
}

// ConfigCmd は config サブコマンドグループ（M18/M19 で実装予定）。
type ConfigCmd struct {
	Init          ConfigInitCmd          `cmd:"" help:"Initialize configuration"`
	Use           ConfigUseCmd           `cmd:"" help:"Switch to a different profile"`
	RefreshModels ConfigRefreshModelsCmd `cmd:"" help:"Refresh model list from API"`
}

// ConfigInitCmd は config init サブコマンド（M19 で実装予定）。
type ConfigInitCmd struct{}

// Run は M19 で実装される。
func (c *ConfigInitCmd) Run() error {
	return errNotImplemented("config init")
}

// ConfigUseCmd は config use サブコマンド（M19 で実装予定）。
type ConfigUseCmd struct {
	Profile string `arg:"" help:"Profile name to switch to"`
}

// Run は M19 で実装される。
func (c *ConfigUseCmd) Run() error {
	return errNotImplemented("config use")
}

// ConfigRefreshModelsCmd は config refresh-models サブコマンド（M19 で実装予定）。
type ConfigRefreshModelsCmd struct{}

// Run は M19 で実装される。
func (c *ConfigRefreshModelsCmd) Run() error {
	return errNotImplemented("config refresh-models")
}

// VersionCmd は version サブコマンド（M20 で実装予定）。
type VersionCmd struct{}

// Run は M20 で実装される。
func (c *VersionCmd) Run() error {
	return errNotImplemented("version")
}

// errNotImplemented は未実装コマンドのエラーを返す。
func errNotImplemented(cmd string) error {
	return &NotImplementedError{Command: cmd}
}

// NotImplementedError は未実装コマンドのエラー型。
type NotImplementedError struct {
	Command string
}

func (e *NotImplementedError) Error() string {
	return e.Command + ": not implemented yet"
}

// Parse は args をパースして CLI 構造体と kong.Context を返す。
// エラー時は (nil, nil, error) を返す。
// サブコマンドが指定されていない（メイン生成コマンド）場合、
// ctx.Command() は "generate" または "generate <prompt>" になる。
func Parse(args []string) (*CLI, *kong.Context, error) {
	var c CLI
	parser, err := kong.New(&c,
		kong.Name("imgraft"),
		kong.Description("Transparent image asset generator for automation pipelines"),
		kong.UsageOnError(),
	)
	if err != nil {
		return nil, nil, err
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		return nil, nil, err
	}

	return &c, ctx, nil
}
