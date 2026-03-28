package imageproc

import "image"

// DefaultThreshold は背景除去のデフォルトしきい値。
// SPEC.md 12.5 に基づき 40 を使用する。
const DefaultThreshold = 40.0

// TransparentPipeline は背景除去 → Trim の透明パイプラインを実行する。
//
// 手順:
//  1. RemoveBackground で背景を除去する
//  2. TrimTransparent で alpha > 0 の bounding box にクロップする
//
// 戻り値:
//   - *image.NRGBA: 処理済み画像
//   - bool: transparent_applied フラグ（成功時は常に true）
//   - error: 全透明になった場合などはエラーを返す
func TransparentPipeline(img image.Image, threshold float64) (*image.NRGBA, bool, error) {
	// Step 1: 背景除去
	removed := RemoveBackground(img, threshold)

	// Step 2: Trim
	trimmed, err := TrimTransparent(removed)
	if err != nil {
		return nil, false, err
	}

	return trimmed, true, nil
}
