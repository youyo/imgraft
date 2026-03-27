package main

import (
	"context"
	"fmt"
	"os"

	"github.com/youyo/imgraft/internal/app"
	"github.com/youyo/imgraft/internal/cli"
	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/output"
)

func main() {
	// panic recovery: パニック時も JSON エラーを stdout に出力する
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("unexpected panic: %v", r)
			out := output.NewErrorOutput(string(errs.ErrInternal), msg)
			_ = output.Encode(os.Stdout, out, false)
			os.Exit(1)
		}
	}()

	// ステップ 1: CLI parse
	c, ctx, err := cli.Parse(os.Args[1:])
	if err != nil {
		// kong の UsageOnError() でエラーメッセージは stderr に出力される
		// JSON エラーを stdout に出力
		out := output.NewErrorOutput(string(errs.ErrInvalidArgument), err.Error())
		_ = output.Encode(os.Stdout, out, false)
		os.Exit(1)
	}

	// サブコマンドが指定された場合は Run() を呼んで処理
	// generate コマンド以外のサブコマンドが指定された場合
	cmd := ctx.Command()
	isGenerateCmd := cmd == "generate" || cmd == "generate <prompt>" || cmd == ""
	if !isGenerateCmd {
		if err := ctx.Run(); err != nil {
			out := output.NewErrorOutput(string(errs.ErrInternal), err.Error())
			_ = output.Encode(os.Stdout, out, false)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// メイン生成パイプラインを実行
	input := app.RunInput{
		Prompt:        c.Generate.Prompt,
		ModelAlias:    c.Generate.Model,
		Refs:          c.Generate.Ref,
		OutputPath:    c.Generate.Output,
		Dir:           c.Generate.Dir,
		NoTransparent: c.Generate.NoTransparent,
		Profile:       c.Generate.Profile,
		ConfigPath:    c.Generate.ConfigPath,
		Pretty:        c.Generate.Pretty,
		Verbose:       c.Generate.Verbose,
		Debug:         c.Generate.Debug,
	}

	result := app.Run(context.Background(), input, app.Dependencies{
		Stderr: os.Stderr,
	})

	// JSON を stdout に出力
	if err := output.Encode(os.Stdout, result.Output, input.Pretty); err != nil {
		// Encode 失敗は致命的エラー
		fmt.Fprintf(os.Stderr, "failed to encode JSON output: %v\n", err)
		os.Exit(1)
	}

	os.Exit(result.ExitCode)
}
