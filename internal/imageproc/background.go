package imageproc

import (
	"image"
	"image/color"
	"math"
)

// cornerSampleSize は四隅からサンプルするピクセル数（N×N）。
const cornerSampleSize = 3

// RemoveBackground は corner_sampling_color_distance アルゴリズムで背景を除去する。
//
// 手順:
//  1. 四隅 N×N ピクセルをサンプリングして背景色を推定する
//  2. 各ピクセルと背景色の RGB 距離を計算する
//  3. 距離に応じた段階的 alpha を生成する（distance < threshold → 0、しきい値付近は線形補間）
//  4. 3×3 カーネルによるエッジ平滑化を適用する
func RemoveBackground(img image.Image, threshold float64) *image.NRGBA {
	bounds := img.Bounds()
	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y

	result := image.NewNRGBA(image.Rect(0, 0, w, h))

	if w == 0 || h == 0 {
		return result
	}

	// Step 1: 四隅サンプリングから背景色を推定する
	bgColor := estimateBackgroundColor(img, bounds)

	// Step 2 & 3: 各ピクセルの alpha を計算する
	alpha := make([]float64, w*h)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// RGBA() は 0-65535 で返るため 8bit に変換
			pr := float64(r >> 8)
			pg := float64(g >> 8)
			pb := float64(b >> 8)

			dist := colorDistance(pr, pg, pb, bgColor[0], bgColor[1], bgColor[2])

			var a float64
			if threshold <= 0 {
				// threshold=0 のとき: distance < 0 は存在しないため全て保持
				a = 255
			} else if dist < threshold {
				// 背景に近い → 透明
				a = 0
			} else {
				// 段階的 alpha: threshold から threshold*1.5 の範囲で線形補間
				fadeZone := threshold * 0.5
				if dist < threshold+fadeZone {
					// 線形補間: 0.0 → 1.0
					t := (dist - threshold) / fadeZone
					a = t * 255
				} else {
					// 前景 → 不透明
					a = 255
				}
			}

			idx := (y-bounds.Min.Y)*w + (x - bounds.Min.X)
			alpha[idx] = a
		}
	}

	// Step 4: 3×3 カーネルによるエッジ平滑化
	smoothed := smoothAlpha(alpha, w, h)

	// 結果画像に書き込む
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			idx := (y-bounds.Min.Y)*w + (x - bounds.Min.X)
			a := uint8(math.Round(smoothed[idx]))

			r, g, b, _ := img.At(x, y).RGBA()
			result.SetNRGBA(x-bounds.Min.X, y-bounds.Min.Y, color.NRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: a,
			})
		}
	}

	return result
}

// estimateBackgroundColor は画像四隅の N×N ピクセルから背景色(RGB)を推定する。
// 四隅ピクセルの平均を背景色とする。
func estimateBackgroundColor(img image.Image, bounds image.Rectangle) [3]float64 {
	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y

	n := cornerSampleSize
	if n > w {
		n = w
	}
	if n > h {
		n = h
	}

	var sumR, sumG, sumB float64
	count := 0

	// 四隅のサンプリング範囲
	regions := []image.Rectangle{
		// 左上
		image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+n, bounds.Min.Y+n),
		// 右上
		image.Rect(bounds.Max.X-n, bounds.Min.Y, bounds.Max.X, bounds.Min.Y+n),
		// 左下
		image.Rect(bounds.Min.X, bounds.Max.Y-n, bounds.Min.X+n, bounds.Max.Y),
		// 右下
		image.Rect(bounds.Max.X-n, bounds.Max.Y-n, bounds.Max.X, bounds.Max.Y),
	}

	for _, region := range regions {
		for y := region.Min.Y; y < region.Max.Y; y++ {
			for x := region.Min.X; x < region.Max.X; x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				sumR += float64(r >> 8)
				sumG += float64(g >> 8)
				sumB += float64(b >> 8)
				count++
			}
		}
	}

	if count == 0 {
		return [3]float64{0, 0, 0}
	}

	return [3]float64{sumR / float64(count), sumG / float64(count), sumB / float64(count)}
}

// colorDistance は2色のユークリッド距離を計算する。
// sqrt((r1-r2)^2 + (g1-g2)^2 + (b1-b2)^2)
func colorDistance(r1, g1, b1, r2, g2, b2 float64) float64 {
	dr := r1 - r2
	dg := g1 - g2
	db := b1 - b2
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// smoothAlpha は alpha スライスに 3×3 平均フィルタを適用する（軽いエッジ補正）。
func smoothAlpha(alpha []float64, w, h int) []float64 {
	result := make([]float64, len(alpha))

	for y := range h {
		for x := range w {
			var sum float64
			count := 0

			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					ny := y + dy
					nx := x + dx
					if ny >= 0 && ny < h && nx >= 0 && nx < w {
						sum += alpha[ny*w+nx]
						count++
					}
				}
			}

			if count > 0 {
				result[y*w+x] = sum / float64(count)
			}
		}
	}

	return result
}
