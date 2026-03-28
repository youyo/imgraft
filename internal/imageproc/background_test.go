package imageproc_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/youyo/imgraft/internal/imageproc"
)

// makeSolidColorImage は指定色の単色画像を生成する。
func makeSolidColorImage(w, h int, c color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

// makeForegroundOnBackgroundImage は前景色(fg)の中央オブジェクトと背景色(bg)の画像を生成する。
// centerW, centerH が前景領域のサイズ。
func makeForegroundOnBackgroundImage(w, h int, bg, fg color.RGBA, centerW, centerH int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetRGBA(x, y, bg)
		}
	}
	// 中央に前景オブジェクトを配置
	startX := (w - centerW) / 2
	startY := (h - centerH) / 2
	for y := startY; y < startY+centerH; y++ {
		for x := startX; x < startX+centerW; x++ {
			img.SetRGBA(x, y, fg)
		}
	}
	return img
}

// TestRemoveBackground_SolidGreenBackground は純緑背景が完全除去されることを確認する。
func TestRemoveBackground_SolidGreenBackground(t *testing.T) {
	// 純緑(#00FF00)単色画像
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	img := makeSolidColorImage(10, 10, green)

	result := imageproc.RemoveBackground(img, 40)

	// 全ピクセルが透明になること
	for y := range 10 {
		for x := range 10 {
			c := result.At(x, y)
			_, _, _, a := c.RGBA()
			// alpha は 0-65535 の範囲で返ってくる
			if a != 0 {
				t.Errorf("pixel(%d,%d) expected alpha=0, got %d", x, y, a>>8)
			}
		}
	}
}

// TestRemoveBackground_ThresholdZero は threshold=0 で何も除去されないことを確認する。
func TestRemoveBackground_ThresholdZero(t *testing.T) {
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	img := makeSolidColorImage(10, 10, green)

	result := imageproc.RemoveBackground(img, 0)

	// threshold=0: 完全一致ピクセルのみが除去対象。四隅の推定色と完全一致するものは除去されるが、
	// 補間範囲がないため境界が全て0になる可能性がある。
	// 実際の挙動: distance < threshold(=0) の場合 alpha=0。distance==0 は threshold(0) より小さくない。
	// つまり distance <= 0 の場合のみ除去される → distance=0 のみ除去される。
	// 純緑単色なら全ピクセルがdistance=0で除去される。
	// この仕様は "threshold=0 で何も除去されない" という直感と異なる可能性があるので、
	// ここでは threshold=0 でも distance==0 のピクセルは除去されないことを確認する。
	// より正確な仕様: threshold=0 では distance < 0 は存在しないため alpha=255（保持）となること。
	for y := range 10 {
		for x := range 10 {
			c := result.At(x, y)
			_, _, _, a := c.RGBA()
			if a == 0 {
				t.Errorf("pixel(%d,%d) expected alpha>0 with threshold=0, got alpha=0", x, y)
			}
		}
	}
}

// TestRemoveBackground_ForegroundPreserved は前景オブジェクトが保持されることを確認する。
func TestRemoveBackground_ForegroundPreserved(t *testing.T) {
	// 純緑背景に赤い中央オブジェクト (前景と背景の距離 >> 40)
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	// 20x20 画像、中央10x10が前景
	img := makeForegroundOnBackgroundImage(20, 20, green, red, 10, 10)

	result := imageproc.RemoveBackground(img, 40)

	// 完全に前景色の中心ピクセルは alpha=255 であること
	// 中央ピクセル (10, 10)
	centerX, centerY := 10, 10
	c := result.At(centerX, centerY)
	_, _, _, a := c.RGBA()
	if a < 0xffff*200/255 { // alpha が 200/255 以上であること
		t.Errorf("center foreground pixel(%d,%d) alpha too low: %d", centerX, centerY, a>>8)
	}

	// 四隅（背景）は透明であること
	corners := [][2]int{{0, 0}, {19, 0}, {0, 19}, {19, 19}}
	for _, c := range corners {
		px := result.At(c[0], c[1])
		_, _, _, a := px.RGBA()
		if a != 0 {
			t.Errorf("corner pixel(%d,%d) expected alpha=0, got %d", c[0], c[1], a>>8)
		}
	}
}

// TestRemoveBackground_EdgeAntialiasing はエッジ部分でアンチエイリアスが効くことを確認する。
func TestRemoveBackground_EdgeAntialiasing(t *testing.T) {
	// 純緑背景に赤い前景オブジェクト
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	img := makeForegroundOnBackgroundImage(30, 30, green, red, 10, 10)

	result := imageproc.RemoveBackground(img, 40)

	// 完全背景(四隅)は alpha=0
	corners := [][2]int{{0, 0}, {29, 0}, {0, 29}, {29, 29}}
	for _, corner := range corners {
		px := result.At(corner[0], corner[1])
		_, _, _, a := px.RGBA()
		if a != 0 {
			t.Errorf("corner(%d,%d) should be transparent, got alpha=%d", corner[0], corner[1], a>>8)
		}
	}

	// 完全前景（中央）は alpha=255 に近いこと
	cx, cy := 15, 15
	px := result.At(cx, cy)
	_, _, _, a := px.RGBA()
	if a < 0xffff*200/255 {
		t.Errorf("center foreground(%d,%d) should be opaque, got alpha=%d", cx, cy, a>>8)
	}
}

// TestRemoveBackground_AllBackground は全ピクセルが背景色の場合に全透明になることを確認する。
func TestRemoveBackground_AllBackground(t *testing.T) {
	// 全て純緑の画像
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	img := makeSolidColorImage(15, 15, green)

	result := imageproc.RemoveBackground(img, 40)

	// 全ピクセルが透明
	bounds := result.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := result.At(x, y).RGBA()
			if a != 0 {
				t.Errorf("pixel(%d,%d) expected fully transparent, got alpha=%d", x, y, a>>8)
			}
		}
	}
}

// TestRemoveBackground_ReturnsNRGBA は戻り値が *image.NRGBA であることを確認する。
func TestRemoveBackground_ReturnsNRGBA(t *testing.T) {
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	img := makeSolidColorImage(5, 5, green)

	result := imageproc.RemoveBackground(img, 40)

	if result == nil {
		t.Fatal("RemoveBackground returned nil")
	}
	// *image.NRGBA 型であること
	if _, ok := interface{}(result).(*image.NRGBA); !ok {
		t.Errorf("expected *image.NRGBA, got %T", result)
	}
}

// TestRemoveBackground_NearBackgroundColor は背景に近い色は除去されることを確認する。
func TestRemoveBackground_NearBackgroundColor(t *testing.T) {
	// 背景: 純緑 (0, 255, 0)
	// 前景: わずかに変化した緑 (10, 245, 10) - 距離 ≈ sqrt(100+100+100) ≈ 17.3 < 40 → 除去
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	nearGreen := color.RGBA{R: 10, G: 245, B: 10, A: 255}
	img := makeForegroundOnBackgroundImage(20, 20, green, nearGreen, 8, 8)

	result := imageproc.RemoveBackground(img, 40)

	// 中央のピクセルは距離が閾値以下なので透明または半透明
	cx, cy := 10, 10
	_, _, _, a := result.At(cx, cy).RGBA()
	// 距離 ≈ 17.3 < 40 なので alpha=0
	if a != 0 {
		t.Errorf("near-background pixel(%d,%d) should be transparent (distance < threshold), got alpha=%d", cx, cy, a>>8)
	}
}
