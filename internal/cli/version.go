// Package cli は imgraft の CLI 定義を kong を使って提供する。
// このファイルは version コマンドの実装を提供する。
package cli

import (
	"fmt"
	"io"
	"os"
)

// バージョン情報変数。ldflags で埋め込む。
//
// ビルド例:
//
//	go build -ldflags "-X github.com/youyo/imgraft/internal/cli.Version=1.0.0 \
//	  -X github.com/youyo/imgraft/internal/cli.Commit=abc1234 \
//	  -X github.com/youyo/imgraft/internal/cli.Date=2026-03-28" ./cmd/imgraft/
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Run は imgraft version コマンドを実行する。
// バージョン情報を stdout に出力する。
func (c *VersionCmd) Run() error {
	return c.RunWithWriter(os.Stdout)
}

// RunWithWriter はバージョン情報を指定した io.Writer に出力する。
// テスト時にバッファを渡せるようにする。
func (c *VersionCmd) RunWithWriter(w io.Writer) error {
	_, err := fmt.Fprintf(w, "imgraft version %s (commit: %s, built: %s)\n", Version, Commit, Date)
	return err
}
