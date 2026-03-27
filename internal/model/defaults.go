// Package model はモデル alias の解決を担う。
// flash/pro の alias を実モデル名に変換する。
package model

const (
	// AliasFlash は flash alias の文字列。
	AliasFlash = "flash"
	// AliasPro は pro alias の文字列。
	AliasPro = "pro"

	// BuiltinFlashModel は config がない場合の flash フォールバック。
	// config/types.go にも同値の定数があるが、model パッケージの独立性のため再定義する。
	BuiltinFlashModel = "gemini-3.1-flash-image-preview"
	// BuiltinProModel は config がない場合の pro フォールバック。
	// config/types.go にも同値の定数があるが、model パッケージの独立性のため再定義する。
	BuiltinProModel = "gemini-3-pro-image-preview"
)
