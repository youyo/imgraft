// Package cli は imgraft の CLI 定義を kong を使って提供する。
// このファイルは completion コマンドの実装を提供する。
package cli

import (
	"fmt"
	"io"
	"os"
)

// CompletionCmd は completion サブコマンド。
// 指定シェル用の補完スクリプトを stdout に出力する。
type CompletionCmd struct {
	Shell string `arg:"" help:"Shell type (currently only 'zsh' is supported)" enum:"zsh"`
}

// Run は imgraft completion コマンドを実行する。
// 補完スクリプトを stdout に出力する。
func (c *CompletionCmd) Run() error {
	return c.RunWithWriter(os.Stdout)
}

// RunWithWriter は補完スクリプトを指定した io.Writer に出力する。
// テスト時にバッファを渡せるようにする。
func (c *CompletionCmd) RunWithWriter(w io.Writer) error {
	switch c.Shell {
	case "zsh":
		return writeZshCompletion(w)
	default:
		return fmt.Errorf("unsupported shell: %s (currently only 'zsh' is supported)", c.Shell)
	}
}

// writeZshCompletion は zsh 補完スクリプトを出力する。
func writeZshCompletion(w io.Writer) error {
	script := `#compdef imgraft

# imgraft zsh completion script
# Usage: eval "$(imgraft completion zsh)"
# or: imgraft completion zsh > "${fpath[1]}/_imgraft"

_imgraft() {
  local state

  _arguments -C \
    '1: :->command' \
    '*: :->args'

  case $state in
    command)
      local -a commands
      commands=(
        'auth:Authentication commands'
        'config:Configuration commands'
        'version:Show version information'
        'completion:Generate shell completion script'
      )
      _describe 'command' commands
      ;;
    args)
      case $words[2] in
        auth)
          local -a auth_commands
          auth_commands=(
            'login:Login to backend'
            'logout:Logout from backend'
            'whoami:Show current authentication status'
          )
          _describe 'auth command' auth_commands
          ;;
        config)
          local -a config_commands
          config_commands=(
            'init:Initialize configuration'
            'use:Switch to a different profile'
            'refresh-models:Refresh model list from API'
          )
          _describe 'config command' config_commands
          ;;
        completion)
          local -a shells
          shells=('zsh:Zsh shell')
          _describe 'shell' shells
          ;;
        *)
          # メイン生成コマンドのフラグ
          _arguments \
            '--model[Model alias or full name (flash, pro, or full model name)]:model:(flash pro)' \
            '--ref[Reference image path or URL (can be specified multiple times)]:file:_files' \
            '--output[Output file path]:file:_files' \
            '--dir[Output directory]:directory:_directories' \
            '--no-transparent[Disable transparent mode]' \
            '--profile[Profile name to use]:profile:' \
            '--config[Config file path]:file:_files' \
            '--pretty[Pretty-print JSON output]' \
            '--verbose[Enable verbose logging to stderr]' \
            '--debug[Enable debug logging to stderr]' \
            ':prompt:'
          ;;
      esac
      ;;
  esac
}

_imgraft "$@"
`
	_, err := fmt.Fprint(w, script)
	return err
}
