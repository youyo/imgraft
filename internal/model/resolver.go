package model

import "github.com/youyo/imgraft/internal/config"

// Resolve は alias と Config から実モデル名を返す。
//
// 解決順（SPEC.md §9.3）:
//  1. alias が空なら cfg.DefaultModel を使う
//  2. 値が "flash" なら cfg.Models["flash"]
//  3. 値が "pro" なら cfg.Models["pro"]
//  4. それ以外はフルモデル名としてそのまま返す
//
// cfg.Models が nil でも安全に動作する（Go の nil map 読み取りはゼロ値を返す）。
// Models にエントリが無い場合は built-in デフォルトにフォールバックする。
func Resolve(alias string, cfg *config.Config) string {
	effective := alias
	if effective == "" {
		effective = cfg.DefaultModel
		if effective == "" {
			effective = AliasFlash
		}
	}

	switch effective {
	case AliasFlash:
		if v := cfg.Models[AliasFlash]; v != "" {
			return v
		}
		return BuiltinFlashModel
	case AliasPro:
		if v := cfg.Models[AliasPro]; v != "" {
			return v
		}
		return BuiltinProModel
	default:
		return effective
	}
}
