package imageproc_test

import (
	"errors"
	"image"
	"image/color"
	"testing"

	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/imageproc"
)

func TestEncodePNG_OpaqueImage(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := range 10 {
		for x := range 10 {
			rgba.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	data, err := imageproc.EncodePNG(rgba)
	if err != nil {
		t.Fatalf("EncodePNG returned error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("EncodePNG returned empty bytes")
	}

	// ラウンドトリップ: デコードして確認
	img, format, err := imageproc.Decode(data)
	if err != nil {
		t.Fatalf("Decode roundtrip failed: %v", err)
	}
	if format != "png" {
		t.Errorf("format = %q, want %q", format, "png")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("size = %dx%d, want 10x10", bounds.Dx(), bounds.Dy())
	}
}

func TestEncodePNG_TransparentImage(t *testing.T) {
	nrgba := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for y := range 8 {
		for x := range 8 {
			nrgba.SetNRGBA(x, y, color.NRGBA{R: 100, G: 200, B: 50, A: 128})
		}
	}

	data, err := imageproc.EncodePNG(nrgba)
	if err != nil {
		t.Fatalf("EncodePNG returned error: %v", err)
	}

	// ラウンドトリップ: デコードしてalpha保持を確認
	img, _, err := imageproc.Decode(data)
	if err != nil {
		t.Fatalf("Decode roundtrip failed: %v", err)
	}
	// 中央ピクセルのalphaを確認
	_, _, _, a := img.At(4, 4).RGBA()
	// image/png は NRGBA → NRGBA で保持するが、RGBA()は premultiplied で返す
	if a == 0xFFFF {
		t.Error("alpha should not be fully opaque for transparent input")
	}
}

func TestEncodePNG_FullyTransparent(t *testing.T) {
	nrgba := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	nrgba.SetNRGBA(0, 0, color.NRGBA{R: 0, G: 0, B: 0, A: 0})

	data, err := imageproc.EncodePNG(nrgba)
	if err != nil {
		t.Fatalf("EncodePNG returned error: %v", err)
	}

	img, _, err := imageproc.Decode(data)
	if err != nil {
		t.Fatalf("Decode roundtrip failed: %v", err)
	}
	_, _, _, a := img.At(0, 0).RGBA()
	if a != 0 {
		t.Errorf("alpha = %d, want 0", a)
	}
}

func TestEncodePNG_NilImage(t *testing.T) {
	_, err := imageproc.EncodePNG(nil)
	if err == nil {
		t.Fatal("EncodePNG(nil) returned nil error")
	}
	var coded *errs.CodedError
	if !errors.As(err, &coded) || coded.Code != errs.ErrInternal {
		t.Errorf("error code = %v, want %v", errs.CodeOf(err), errs.ErrInternal)
	}
}

func TestEncodePNG_LargeImage(t *testing.T) {
	// 512x512 (4096x4096はテストに時間がかかるため控えめに)
	nrgba := image.NewNRGBA(image.Rect(0, 0, 512, 512))
	for y := range 512 {
		for x := range 512 {
			nrgba.SetNRGBA(x, y, color.NRGBA{R: uint8(x % 256), G: uint8(y % 256), B: 100, A: 255})
		}
	}

	data, err := imageproc.EncodePNG(nrgba)
	if err != nil {
		t.Fatalf("EncodePNG returned error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("EncodePNG returned empty bytes for large image")
	}
}
