package model

import "strings"

// ResolveAliasesFromModels は ListModels の結果から flash/pro の alias マップを生成する。
//
// ヒューリスティック:
//   - モデル名に "flash" を含む → "flash" alias として登録
//   - モデル名に "pro" を含む → "pro" alias として登録
//   - 複数マッチする場合はリストの最後のモデルを採用（より新しい可能性が高いため）
//
// 引数 models はフルモデル名のスライス（例: "models/gemini-2.0-flash-exp"）。
// 戻り値は alias → full model name のマップ。エントリがない alias はマップに含まれない。
func ResolveAliasesFromModels(models []string) map[string]string {
	result := make(map[string]string)

	for _, name := range models {
		// モデル名の最後のセグメントのみをチェック（プレフィックス部分を除く）
		segment := name
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			segment = name[idx+1:]
		}
		lowerSegment := strings.ToLower(segment)

		if strings.Contains(lowerSegment, AliasFlash) {
			result[AliasFlash] = name
		}
		if strings.Contains(lowerSegment, AliasPro) {
			result[AliasPro] = name
		}
	}

	return result
}
