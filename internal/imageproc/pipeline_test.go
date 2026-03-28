package imageproc_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/youyo/imgraft/internal/imageproc"
)

// makeGreenBackgroundWithSubject は純緑背景に前景オブジェクトを持つ画像を生成する。
func makeGreenBackgroundWithSubject(w, h int, fg color.RGBA, subjectRect image.Rectangle) image.Image {
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetRGBA(x, y, green)
		}
	}
	for y := subjectRect.Min.Y; y < subjectRect.Max.Y; y++ {
		for x := subjectRect.Min.X; x < subjectRect.Max.X; x++ {
			img.SetRGBA(x, y, fg)
		}
	}
	return img
}

// TestTransparentPipeline_ReturnsTrueWhenApplied は透明パイプラインが成功時に true を返すことを確認する。
func TestTransparentPipeline_ReturnsTrueWhenApplied(t *testing.T) {
	// 純緑背景に赤いオブジェクト
	img := makeGreenBackgroundWithSubject(
		30, 30,
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		image.Rect(10, 10, 20, 20),
	)

	result, applied, err := imageproc.TransparentPipeline(img, 40)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !applied {
		t.Error("expected transparent_applied=true, got false")
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

// TestTransparentPipeline_ReturnsNRGBA は戻り値が *image.NRGBA であることを確認する。
func TestTransparentPipeline_ReturnsNRGBA(t *testing.T) {
	img := makeGreenBackgroundWithSubject(
		20, 20,
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		image.Rect(5, 5, 15, 15),
	)

	result, _, err := imageproc.TransparentPipeline(img, 40)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := interface{}(result).(*image.NRGBA); !ok {
		t.Errorf("expected *image.NRGBA, got %T", result)
	}
}

// TestTransparentPipeline_RemovesGreenBackground は背景除去が機能することを確認する。
func TestTransparentPipeline_RemovesGreenBackground(t *testing.T) {
	img := makeGreenBackgroundWithSubject(
		30, 30,
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		image.Rect(10, 10, 20, 20),
	)

	result, _, err := imageproc.TransparentPipeline(img, 40)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// トリム後の画像サイズは前景領域より小さいこと（背景除去により周囲の緑が除去される）
	// エッジ平滑化により実際のサイズは前景領域より若干大きい可能性があるが、
	// 元の 30x30 よりは大幅に小さくなるはず
	bounds := result.Bounds()
	originalW, originalH := 30, 30
	if bounds.Dx() >= originalW || bounds.Dy() >= originalH {
		t.Errorf("trimmed size should be smaller than original 30x30, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestTransparentPipeline_AllGreenReturnsError は全背景画像でエラーが返ることを確認する。
func TestTransparentPipeline_AllGreenReturnsError(t *testing.T) {
	// 純緑単色（前景なし）
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	img := makeSolidColorImage(10, 10, green)

	_, _, err := imageproc.TransparentPipeline(img, 40)
	if err == nil {
		t.Fatal("expected error for all-background image, got nil")
	}
}

// TestTransparentPipeline_ForegroundPixelsOpaque は前景ピクセルが不透明であることを確認する。
func TestTransparentPipeline_ForegroundPixelsOpaque(t *testing.T) {
	img := makeGreenBackgroundWithSubject(
		40, 40,
		color.RGBA{R: 0, G: 0, B: 255, A: 255}, // 青い前景
		image.Rect(15, 15, 25, 25),
	)

	result, _, err := imageproc.TransparentPipeline(img, 40)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// パイプライン後、全ピクセルは何らかの形で前景（不透明）であること
	bounds := result.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		t.Fatal("trimmed image has zero size")
	}

	// 中央ピクセルは不透明であること
	cx, cy := bounds.Min.X+bounds.Dx()/2, bounds.Min.Y+bounds.Dy()/2
	_, _, _, a := result.At(cx, cy).RGBA()
	if a < 0xffff*200/255 {
		t.Errorf("center pixel alpha too low: got %d, expected >= 200", a>>8)
	}
}
