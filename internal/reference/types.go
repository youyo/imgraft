// Package reference は参照画像の読込・検証・統一型変換を提供する。
package reference

// ReferenceImage は参照画像の内部統一型。
// SPEC.md セクション10.9, 20.6 を参照。
type ReferenceImage struct {
	// SourceType は参照元の種別。"file" または "url"。
	SourceType string

	// OriginalInput はユーザーが指定した元のパスまたは URL。
	OriginalInput string

	// LocalCachedPath はローカルキャッシュパス。ローカルファイルの場合はそのパス。
	LocalCachedPath string

	// Filename はファイル名（パスなし）。
	Filename string

	// MimeType は画像の MIME タイプ（例: "image/png"）。
	MimeType string

	// Width は画像の幅（ピクセル）。
	Width int

	// Height は画像の高さ（ピクセル）。
	Height int

	// SizeBytes はファイルサイズ（バイト）。
	SizeBytes int64

	// Data は画像のバイト列。
	Data []byte
}
