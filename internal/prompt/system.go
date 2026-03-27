// Package prompt はプロンプト構築を担う。
// transparent モード対応のシステムプロンプト定義とユーザープロンプトの合成を行う。
package prompt

// transparentSystemPrompt は SPEC §11.2 に定義された transparent ON 時のシステムプロンプト。
// #00FF00 純緑背景を強制し、アセット生成用途へ誘導する。
const transparentSystemPrompt = `Generate a single isolated subject asset for compositing.

Use a solid pure green background.
Do not use gradients.
Do not use shadows on the background.

Do not include background objects, scenery, environment, text, borders, or frames.

Center the subject.
Keep the full silhouette visible and cleanly separated from the background.

Ensure strong color contrast between subject and background.`

// SystemPrompt は transparent モードに応じたシステムプロンプト文字列を返す。
//
// transparent=true のとき SPEC §11.2 の固定文言を返す。
// transparent=false のとき空文字列を返す（SPEC §11.3: 背景強制を外す）。
func SystemPrompt(transparent bool) string {
	if transparent {
		return transparentSystemPrompt
	}
	return ""
}
