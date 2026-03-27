package imageproc_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/imageproc"
)

func TestInspectFile_PNG(t *testing.T) {
	data := mustMakePNG(t, 32, 24, false)
	path := mustWriteTempFile(t, data, ".png")

	meta, err := imageproc.InspectFile(path)
	if err != nil {
		t.Fatalf("InspectFile(PNG) returned error: %v", err)
	}
	if meta.Width != 32 || meta.Height != 24 {
		t.Errorf("size = %dx%d, want 32x24", meta.Width, meta.Height)
	}
	if meta.MimeType != "image/png" {
		t.Errorf("mime = %q, want %q", meta.MimeType, "image/png")
	}
}

func TestInspectFile_JPEG(t *testing.T) {
	data := mustMakeJPEG(t, 100, 50)
	path := mustWriteTempFile(t, data, ".jpg")

	meta, err := imageproc.InspectFile(path)
	if err != nil {
		t.Fatalf("InspectFile(JPEG) returned error: %v", err)
	}
	if meta.Width != 100 || meta.Height != 50 {
		t.Errorf("size = %dx%d, want 100x50", meta.Width, meta.Height)
	}
	if meta.MimeType != "image/jpeg" {
		t.Errorf("mime = %q, want %q", meta.MimeType, "image/jpeg")
	}
}

func TestInspectFile_WebP(t *testing.T) {
	path := filepath.Join("testdata", "sample.webp")
	meta, err := imageproc.InspectFile(path)
	if err != nil {
		t.Fatalf("InspectFile(WebP) returned error: %v", err)
	}
	if meta.Width == 0 || meta.Height == 0 {
		t.Error("InspectFile(WebP) returned zero dimensions")
	}
	if meta.MimeType != "image/webp" {
		t.Errorf("mime = %q, want %q", meta.MimeType, "image/webp")
	}
}

func TestInspectFile_PNGWithAlpha(t *testing.T) {
	data := mustMakePNG(t, 64, 48, true)
	path := mustWriteTempFile(t, data, ".png")

	meta, err := imageproc.InspectFile(path)
	if err != nil {
		t.Fatalf("InspectFile(PNG with alpha) returned error: %v", err)
	}
	if meta.Width != 64 || meta.Height != 48 {
		t.Errorf("size = %dx%d, want 64x48", meta.Width, meta.Height)
	}
	if meta.MimeType != "image/png" {
		t.Errorf("mime = %q, want %q", meta.MimeType, "image/png")
	}
}

func TestInspectFile_NotExists(t *testing.T) {
	_, err := imageproc.InspectFile(filepath.Join(t.TempDir(), "nonexistent.png"))
	if err == nil {
		t.Fatal("InspectFile(nonexistent) returned nil error")
	}
	var coded *errs.CodedError
	if !errors.As(err, &coded) || coded.Code != errs.ErrFileNotFound {
		t.Errorf("error code = %v, want %v", errs.CodeOf(err), errs.ErrFileNotFound)
	}
}

func TestInspectFile_Corrupted(t *testing.T) {
	path := mustWriteTempFile(t, []byte{0xDE, 0xAD, 0xBE, 0xEF}, ".png")
	_, err := imageproc.InspectFile(path)
	if err == nil {
		t.Fatal("InspectFile(corrupted) returned nil error")
	}
	code := errs.CodeOf(err)
	if code != errs.ErrInvalidImage {
		t.Errorf("error code = %v, want %v", code, errs.ErrInvalidImage)
	}
}

func TestInspectBytes_PNG(t *testing.T) {
	data := mustMakePNG(t, 20, 15, false)
	meta, err := imageproc.InspectBytes(data)
	if err != nil {
		t.Fatalf("InspectBytes(PNG) returned error: %v", err)
	}
	if meta.Width != 20 || meta.Height != 15 {
		t.Errorf("size = %dx%d, want 20x15", meta.Width, meta.Height)
	}
	if meta.MimeType != "image/png" {
		t.Errorf("mime = %q, want %q", meta.MimeType, "image/png")
	}
}

func TestInspectBytes_Empty(t *testing.T) {
	_, err := imageproc.InspectBytes([]byte{})
	if err == nil {
		t.Fatal("InspectBytes(empty) returned nil error")
	}
	var coded *errs.CodedError
	if !errors.As(err, &coded) || coded.Code != errs.ErrInvalidImage {
		t.Errorf("error code = %v, want %v", errs.CodeOf(err), errs.ErrInvalidImage)
	}
}

func TestInspectBytes_GIF(t *testing.T) {
	data := makeGIFBytes()
	_, err := imageproc.InspectBytes(data)
	if err == nil {
		t.Fatal("InspectBytes(GIF) returned nil error")
	}
	code := errs.CodeOf(err)
	if code != errs.ErrUnsupportedImageFormat && code != errs.ErrInvalidImage {
		t.Errorf("error code = %v, want %v or %v", code, errs.ErrUnsupportedImageFormat, errs.ErrInvalidImage)
	}
}
