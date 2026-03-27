package imageproc_test

import (
	"errors"
	"testing"

	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/imageproc"
)

func TestDecode_PNG(t *testing.T) {
	data := mustMakePNG(t, 32, 24, false)
	img, format, err := imageproc.Decode(data)
	if err != nil {
		t.Fatalf("Decode(PNG) returned error: %v", err)
	}
	if format != "png" {
		t.Errorf("format = %q, want %q", format, "png")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 24 {
		t.Errorf("size = %dx%d, want 32x24", bounds.Dx(), bounds.Dy())
	}
}

func TestDecode_JPEG(t *testing.T) {
	data := mustMakeJPEG(t, 100, 50)
	img, format, err := imageproc.Decode(data)
	if err != nil {
		t.Fatalf("Decode(JPEG) returned error: %v", err)
	}
	if format != "jpeg" {
		t.Errorf("format = %q, want %q", format, "jpeg")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 50 {
		t.Errorf("size = %dx%d, want 100x50", bounds.Dx(), bounds.Dy())
	}
}

func TestDecode_WebP(t *testing.T) {
	data := mustReadWebP(t)
	img, format, err := imageproc.Decode(data)
	if err != nil {
		t.Fatalf("Decode(WebP) returned error: %v", err)
	}
	if format != "webp" {
		t.Errorf("format = %q, want %q", format, "webp")
	}
	bounds := img.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		t.Error("decoded WebP has zero dimensions")
	}
}

func TestDecode_PNGWithAlpha(t *testing.T) {
	data := mustMakePNG(t, 16, 16, true)
	img, _, err := imageproc.Decode(data)
	if err != nil {
		t.Fatalf("Decode(PNG with alpha) returned error: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 16 || bounds.Dy() != 16 {
		t.Errorf("size = %dx%d, want 16x16", bounds.Dx(), bounds.Dy())
	}
}

func TestDecode_1x1PNG(t *testing.T) {
	data := mustMakePNG(t, 1, 1, false)
	img, _, err := imageproc.Decode(data)
	if err != nil {
		t.Fatalf("Decode(1x1 PNG) returned error: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 1 || bounds.Dy() != 1 {
		t.Errorf("size = %dx%d, want 1x1", bounds.Dx(), bounds.Dy())
	}
}

func TestDecode_NilBytes(t *testing.T) {
	_, _, err := imageproc.Decode(nil)
	if err == nil {
		t.Fatal("Decode(nil) returned nil error")
	}
	var coded *errs.CodedError
	if !errors.As(err, &coded) || coded.Code != errs.ErrInvalidImage {
		t.Errorf("error code = %v, want %v", errs.CodeOf(err), errs.ErrInvalidImage)
	}
}

func TestDecode_EmptyBytes(t *testing.T) {
	_, _, err := imageproc.Decode([]byte{})
	if err == nil {
		t.Fatal("Decode(empty) returned nil error")
	}
	var coded *errs.CodedError
	if !errors.As(err, &coded) || coded.Code != errs.ErrInvalidImage {
		t.Errorf("error code = %v, want %v", errs.CodeOf(err), errs.ErrInvalidImage)
	}
}

func TestDecode_CorruptedData(t *testing.T) {
	_, _, err := imageproc.Decode([]byte{0xDE, 0xAD, 0xBE, 0xEF})
	if err == nil {
		t.Fatal("Decode(corrupted) returned nil error")
	}
	code := errs.CodeOf(err)
	if code != errs.ErrInvalidImage {
		t.Errorf("error code = %v, want %v", code, errs.ErrInvalidImage)
	}
}

func TestDecode_GIF_Unsupported(t *testing.T) {
	data := makeGIFBytes()
	_, _, err := imageproc.Decode(data)
	if err == nil {
		t.Fatal("Decode(GIF) returned nil error")
	}
	code := errs.CodeOf(err)
	// GIF はimage.Decodeで登録されていないため、ErrInvalidImage になる。
	// DetectFormatで "gif" を検出した場合は ErrUnsupportedImageFormat。
	if code != errs.ErrUnsupportedImageFormat && code != errs.ErrInvalidImage {
		t.Errorf("error code = %v, want %v or %v", code, errs.ErrUnsupportedImageFormat, errs.ErrInvalidImage)
	}
}

func TestDecode_TruncatedPNG(t *testing.T) {
	data := mustMakePNG(t, 32, 32, false)
	// PNGヘッダだけ残す
	truncated := data[:8]
	_, _, err := imageproc.Decode(truncated)
	if err == nil {
		t.Fatal("Decode(truncated PNG) returned nil error")
	}
	code := errs.CodeOf(err)
	if code != errs.ErrInvalidImage {
		t.Errorf("error code = %v, want %v", code, errs.ErrInvalidImage)
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"empty", nil, ""},
		{"png", mustMakePNG(t, 2, 2, false), "png"},
		{"jpeg", mustMakeJPEG(t, 2, 2), "jpeg"},
		{"webp", mustReadWebP(t), "webp"},
		{"gif", makeGIFBytes(), "gif"},
		{"random", []byte{0x01, 0x02, 0x03}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := imageproc.DetectFormat(tt.data)
			if got != tt.want {
				t.Errorf("DetectFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}
