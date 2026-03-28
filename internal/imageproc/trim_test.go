package imageproc_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/youyo/imgraft/internal/imageproc"
)

// makeNRGBA はテスト用の *image.NRGBA を生成する。
func makeNRGBA(w, h int, c color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}

// makeNRGBAWithSubject は指定サブ領域のみ不透明、その他は透明の *image.NRGBA を生成する。
func makeNRGBAWithSubject(w, h int, subjectRect image.Rectangle, subjectColor color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	// 全て透明にする
	for y := range h {
		for x := range w {
			img.SetNRGBA(x, y, color.NRGBA{A: 0})
		}
	}
	// サブジェクト領域を不透明にする
	for y := subjectRect.Min.Y; y < subjectRect.Max.Y; y++ {
		for x := subjectRect.Min.X; x < subjectRect.Max.X; x++ {
			img.SetNRGBA(x, y, subjectColor)
		}
	}
	return img
}

// TestTrimTransparent_AllTransparent は全透明画像でエラーが返ることを確認する。
func TestTrimTransparent_AllTransparent(t *testing.T) {
	img := makeNRGBA(10, 10, color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	_, err := imageproc.TrimTransparent(img)
	if err == nil {
		t.Fatal("expected error for fully transparent image, got nil")
	}
}

// TestTrimTransparent_FullyOpaque は全不透明画像でサイズが変わらないことを確認する。
func TestTrimTransparent_FullyOpaque(t *testing.T) {
	img := makeNRGBA(10, 10, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	result, err := imageproc.TrimTransparent(img)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.Bounds()
	if got.Dx() != 10 || got.Dy() != 10 {
		t.Errorf("expected 10x10, got %dx%d", got.Dx(), got.Dy())
	}
}

// TestTrimTransparent_CentersObject は中央オブジェクトが正しくトリミングされることを確認する。
func TestTrimTransparent_CentersObject(t *testing.T) {
	// 20x20 画像、中央 [5,5]〜[14,14] にオブジェクトを配置（10x10 領域）
	subjectRect := image.Rect(5, 5, 15, 15)
	img := makeNRGBAWithSubject(20, 20, subjectRect, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

	result, err := imageproc.TrimTransparent(img)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bounds := result.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("expected trimmed size 10x10, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// トリム後の左上ピクセルは不透明であること
	_, _, _, a := result.At(bounds.Min.X, bounds.Min.Y).RGBA()
	if a == 0 {
		t.Error("top-left pixel of trimmed image should be opaque")
	}
}

// TestTrimTransparent_SinglePixel は単一不透明ピクセルの場合に1x1になることを確認する。
func TestTrimTransparent_SinglePixel(t *testing.T) {
	img := makeNRGBAWithSubject(10, 10, image.Rect(5, 5, 6, 6), color.NRGBA{R: 0, G: 0, B: 255, A: 255})

	result, err := imageproc.TrimTransparent(img)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bounds := result.Bounds()
	if bounds.Dx() != 1 || bounds.Dy() != 1 {
		t.Errorf("expected 1x1, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestTrimTransparent_PreservesColors はトリム後にピクセル値が保持されることを確認する。
func TestTrimTransparent_PreservesColors(t *testing.T) {
	subjectColor := color.NRGBA{R: 100, G: 150, B: 200, A: 255}
	subjectRect := image.Rect(3, 3, 7, 7)
	img := makeNRGBAWithSubject(10, 10, subjectRect, subjectColor)

	result, err := imageproc.TrimTransparent(img)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// トリム後の全ピクセルのカラーが元の subjectColor と一致すること
	bounds := result.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c, ok := result.At(x, y).(color.NRGBA)
			if !ok {
				// NRGBA型で取得できない場合はRGBAで確認
				r, g, b, a := result.At(x, y).RGBA()
				if uint8(r>>8) != subjectColor.R || uint8(g>>8) != subjectColor.G ||
					uint8(b>>8) != subjectColor.B || uint8(a>>8) != subjectColor.A {
					t.Errorf("pixel(%d,%d): expected %v, got rgba(%d,%d,%d,%d)",
						x, y, subjectColor, r>>8, g>>8, b>>8, a>>8)
				}
			} else {
				if c != subjectColor {
					t.Errorf("pixel(%d,%d): expected %v, got %v", x, y, subjectColor, c)
				}
			}
		}
	}
}

// TestTrimTransparent_PartialAlpha は alpha > 0 のピクセルが保持されることを確認する。
func TestTrimTransparent_PartialAlpha(t *testing.T) {
	// 半透明ピクセルを持つ画像
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	// 全て透明
	for y := range 10 {
		for x := range 10 {
			img.SetNRGBA(x, y, color.NRGBA{A: 0})
		}
	}
	// 一部のピクセルを半透明(alpha=128)に設定
	img.SetNRGBA(2, 3, color.NRGBA{R: 255, G: 255, B: 255, A: 128})
	img.SetNRGBA(7, 8, color.NRGBA{R: 255, G: 255, B: 255, A: 1})

	result, err := imageproc.TrimTransparent(img)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// bounding box は (2,3)〜(8,9) = 6x6
	bounds := result.Bounds()
	if bounds.Dx() != 6 || bounds.Dy() != 6 {
		t.Errorf("expected 6x6, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestTrimTransparent_ReturnsNRGBA は戻り値が *image.NRGBA であることを確認する。
func TestTrimTransparent_ReturnsNRGBA(t *testing.T) {
	img := makeNRGBA(5, 5, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	result, err := imageproc.TrimTransparent(img)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := interface{}(result).(*image.NRGBA); !ok {
		t.Errorf("expected *image.NRGBA, got %T", result)
	}
}
